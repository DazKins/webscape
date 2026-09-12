package server

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"sync"
	"time"
	"webscape/server/auth"
	"webscape/server/command"
	"webscape/server/config"
	"webscape/server/game"
	"webscape/server/game/world"
	"webscape/server/persistence"
)

// Start owns runtime coordination; core game code never opens a database.
func Start(ctx context.Context, distFS fs.FS, gameWorld *world.World, address string, chunkRadius int, tickInterval time.Duration, devMode bool, storageConfig config.PersistenceConfig, authConfig config.AuthConfig) error {
	authManager, err := auth.New(ctx, authConfig, devMode)
	if err != nil {
		return err
	}
	defer authManager.Close()
	mux := http.NewServeMux()
	authManager.RegisterRoutes(mux)
	mux.Handle("/", frontendHandler(distFS, devMode))
	if devMode {
		log.Print("Development mode: frontend caching disabled")
	}
	g := game.NewGameWithWorldAndChunkRadius(gameWorld, chunkRadius)
	timeout := time.Duration(storageConfig.TimeoutSeconds) * time.Second
	var coordinator *persistence.Coordinator
	if storageConfig.Driver == "postgres" {
		connectionString, err := storageConfig.Postgres.ConnectionString()
		if err != nil {
			return err
		}
		openCtx, cancel := context.WithTimeout(ctx, timeout)
		store, err := persistence.OpenPostgres(openCtx, connectionString, storageConfig.WorldKey)
		cancel()
		if err != nil {
			return err
		}
		defer func() {
			closeCtx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			store.Close(closeCtx)
		}()
		loadCtx, cancel := context.WithTimeout(ctx, timeout)
		data, err := store.Load(loadCtx)
		cancel()
		if err != nil {
			return err
		}
		g.RetainOfflinePlayers()
		if data != nil {
			if err := g.RestoreSnapshot(data); err != nil {
				return err
			}
			log.Print("Restored PostgreSQL game snapshot")
		}
		coordinator = persistence.NewCoordinator(g, store)
		saveCtx, cancel := context.WithTimeout(ctx, timeout)
		err = coordinator.Save(saveCtx)
		cancel()
		if err != nil {
			return err
		}
	} else if storageConfig.Driver == "none" {
		log.Print("Persistence disabled: game state is ephemeral")
	} else {
		return fmt.Errorf("unsupported persistence driver %q", storageConfig.Driver)
	}

	failures := make(chan error, 1)
	save := func() error {
		if coordinator == nil {
			return nil
		}
		saveCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		return coordinator.Save(saveCtx)
	}
	checkpoint := func() {
		if err := save(); err != nil {
			select {
			case failures <- fmt.Errorf("game checkpoint failed: %w", err):
			default:
			}
		}
	}
	if coordinator != nil {
		g.SetAfterTick(checkpoint)
	}
	ws := NewWsServer()
	handler := NewClientCommandHandler(g, ws.PlayerID)
	// Drain in-flight commands before the final save. WebSockets are hijacked
	// connections and must be closed explicitly during HTTP shutdown.
	var lifecycle sync.RWMutex
	stopping := false
	ws.SetConnectHandler(func(id string) {
		lifecycle.RLock()
		defer lifecycle.RUnlock()
		if !stopping {
			g.HandleConnect(id)
		}
	})
	ws.SetIncomingMessageHandler(func(clientID string, raw string) {
		lifecycle.RLock()
		defer lifecycle.RUnlock()
		if stopping {
			return
		}
		cmd, err := command.Unmarshal(raw)
		if err != nil {
			log.Printf("error unmarshalling command: %v", err)
			return
		}
		handler.HandleCommand(clientID, cmd)
		// Persistence follows the fixed tick cadence, not peer traffic. Rejected,
		// unknown and transient commands cannot trigger snapshot work or DB I/O.
	})
	ws.SetDisconnectHandler(func(clientID string) {
		lifecycle.RLock()
		defer lifecycle.RUnlock()
		if stopping {
			return
		}
		if g.HandleLeave(clientID) {
			checkpoint()
		}
	})
	g.RegisterBroadcaster(ws.Broadcast)
	g.RegisterSender(ws.SendToClient)
	g.StartUpdateLoop(tickInterval)
	mux.Handle("/ws", authManager.RequireSocket(ws.HandleWebSocket))
	httpServer := &http.Server{Addr: address, Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	httpErrors := make(chan error, 1)
	go func() { httpErrors <- httpServer.ListenAndServe() }()
	log.Printf("Starting server on %s", address)
	var result error
	select {
	case <-ctx.Done():
	case result = <-failures:
	case result = <-httpErrors:
		if errors.Is(result, http.ErrServerClosed) {
			result = nil
		}
	}
	lifecycle.Lock()
	stopping = true
	lifecycle.Unlock()
	g.Stop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	shutdownErr := httpServer.Shutdown(shutdownCtx)
	ws.Close()
	saveErr := save()
	return errors.Join(result, shutdownErr, saveErr)
}

func frontendHandler(distFS fs.FS, devMode bool) http.Handler {
	files := http.FileServer(http.FS(distFS))
	if !devMode {
		return files
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		files.ServeHTTP(w, r)
	})
}
