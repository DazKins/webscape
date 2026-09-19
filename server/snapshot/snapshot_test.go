package snapshot

import (
	"encoding/json"
	"testing"
)

func TestDecodePlayerSaveContracts(t *testing.T) {
	const id = "11111111-1111-4111-8111-111111111111"
	for _, data := range []string{
		`{"version":99,"players":{}}`,
		`{"version":1,"entities":{},"tick":123,"contentHash":"old"}`,
		`{"version":2,"players":{},"tick":123}`,
		`{"version":2,"players":null}`,
		`{"version":2,"players":{},"entities":{}}`,
		`{"version":2,"players":{"invalid":{"player":{"version":1,"data":{}}}}}`,
		`{"version":2,"players":{"` + id + `":{"inventory":{"version":1,"data":{}}}}}`,
		`{"version":2,"players":{"` + id + `":{"player":{"version":1,"data":null}}}}`,
		`{"version":2,"players":{}} {}`,
	} {
		if _, err := Decode([]byte(data)); err == nil {
			t.Fatalf("accepted invalid save: %s", data)
		}
	}
	for _, data := range []string{`{"version":2,"players":{}}`} {
		state, err := Decode([]byte(data))
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(state)
		if err != nil {
			t.Fatal(err)
		}
		if string(encoded) != `{"version":2,"players":{}}` {
			t.Fatalf("world metadata retained: %s", encoded)
		}
	}
}
