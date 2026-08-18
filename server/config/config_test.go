package config

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestLoadFromFS(t *testing.T) {
	loaded, err := LoadFromFS(fstest.MapFS{
		"config.json": {Data: []byte(`{
			"formatVersion":1,
			"server":{"address":":9090"},
			"client":{"folder":"web"},
			"game":{"folder":"content"},
			"streaming":{"chunkRadius":2}
		}`)},
	}, "config.json")
	if err != nil {
		t.Fatalf("LoadFromFS returned error: %v", err)
	}
	if loaded.Server.Address != ":9090" || loaded.Client.Folder != "web" ||
		loaded.Game.Folder != "content" || loaded.Streaming.ChunkRadius != 2 {
		t.Fatalf("loaded config = %#v", loaded)
	}
}

func TestLoadFromFSRejectsInvalidAndUnknownSettings(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		wantError string
	}{
		{
			name:      "radius below one",
			data:      `{"formatVersion":1,"server":{"address":":8080"},"client":{"folder":"client/dist"},"game":{"folder":"game-project"},"streaming":{"chunkRadius":0}}`,
			wantError: "chunkRadius must be at least 1",
		},
		{
			name:      "unknown setting",
			data:      `{"formatVersion":1,"server":{"address":":8080","port":8080},"client":{"folder":"client/dist"},"game":{"folder":"game-project"},"streaming":{"chunkRadius":1}}`,
			wantError: "unknown field",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := LoadFromFS(fstest.MapFS{"config.json": {Data: []byte(test.data)}}, "config.json")
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("error = %v, want %q", err, test.wantError)
			}
		})
	}
}
