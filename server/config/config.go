package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
)

type Config struct {
	FormatVersion int             `json:"formatVersion"`
	Server        ServerConfig    `json:"server"`
	Client        ClientConfig    `json:"client"`
	Game          GameConfig      `json:"game"`
	Streaming     StreamingConfig `json:"streaming"`
}

type ServerConfig struct {
	Address        string `json:"address"`
	TickIntervalMs int    `json:"tickIntervalMs"`
}

type ClientConfig struct {
	Folder string `json:"folder"`
}

type GameConfig struct {
	Folder string `json:"folder"`
}

type StreamingConfig struct {
	ChunkRadius int `json:"chunkRadius"`
}

func LoadFromFile(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %q: %w", path, err)
	}
	return load(data)
}

func LoadFromFS(configFS fs.FS, path string) (Config, error) {
	data, err := fs.ReadFile(configFS, path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %q: %w", path, err)
	}
	return load(data)
}

func load(data []byte) (Config, error) {
	result := Config{Server: ServerConfig{TickIntervalMs: 500}}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	if err := ensureEndOfJSON(decoder); err != nil {
		return Config{}, err
	}
	if err := result.Validate(); err != nil {
		return Config{}, err
	}
	return result, nil
}

func ensureEndOfJSON(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("parse config: multiple JSON values are not allowed")
		}
		return fmt.Errorf("parse config: %w", err)
	}
	return nil
}

func (c Config) Validate() error {
	if c.FormatVersion != 1 {
		return fmt.Errorf("unsupported config format version %d", c.FormatVersion)
	}
	if c.Server.TickIntervalMs < 1 || c.Server.TickIntervalMs > 60000 {
		return errors.New("config server.tickIntervalMs must be between 1 and 60000")
	}
	if c.Server.Address == "" {
		return errors.New("config server.address is required")
	}
	if c.Client.Folder == "" {
		return errors.New("config client.folder is required")
	}
	if c.Game.Folder == "" {
		return errors.New("config game.folder is required")
	}
	if c.Streaming.ChunkRadius < 1 {
		return errors.New("config streaming.chunkRadius must be at least 1")
	}
	return nil
}
