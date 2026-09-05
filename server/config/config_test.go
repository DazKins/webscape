package config

import (
	"fmt"
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

func TestTickIntervalDefaultsAndValidation(t *testing.T) {
	base := `{"formatVersion":1,"server":{"address":":8080"%s},"client":{"folder":"client/dist"},"game":{"folder":"game-project"},"streaming":{"chunkRadius":1}}`
	for _, test := range []struct {
		setting string
		want    int
		valid   bool
	}{
		{"", 500, true}, {`,"tickIntervalMs":250`, 250, true}, {`,"tickIntervalMs":0`, 0, false}, {`,"tickIntervalMs":-1`, 0, false}, {`,"tickIntervalMs":60001`, 0, false}, {`,"tickIntervalMs":1.5`, 0, false},
	} {
		c, err := load([]byte(fmt.Sprintf(base, test.setting)))
		if (err == nil) != test.valid {
			t.Fatalf("%s: error=%v", test.setting, err)
		}
		if test.valid && c.Server.TickIntervalMs != test.want {
			t.Fatalf("interval=%d want=%d", c.Server.TickIntervalMs, test.want)
		}
	}
}
