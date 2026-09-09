package component

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"webscape/server/snapshot"
)

// SavedComponent is independent of client serialization and the storage backend.
type SavedComponent = snapshot.Component

type PersistentComponent interface {
	Component
	Save() (SavedComponent, error)
}
type transientComponent interface{ transientSave() }
type idleComponent interface{ idleForRestore(uint64) Component }

type restoreKey struct {
	id      ComponentId
	version int
}

var componentRestorers = map[restoreKey]func(SavedComponent) (Component, error){}

// Each component registers its own versioned codec alongside its implementation.
func registerComponentRestore(id ComponentId, version int, restore func(SavedComponent) (Component, error)) {
	key := restoreKey{id, version}
	if _, exists := componentRestorers[key]; exists {
		panic("duplicate component restore codec: " + string(id))
	}
	componentRestorers[key] = restore
}

func SaveComponent(value Component) (SavedComponent, error) {
	if _, ok := value.(transientComponent); ok {
		return SavedComponent{}, nil
	}
	if c, ok := value.(PersistentComponent); ok {
		return c.Save()
	}
	return SavedComponent{}, fmt.Errorf("component %q has no save codec", value.GetId())
}
func RestoreComponent(id ComponentId, saved SavedComponent) (Component, error) {
	restore, ok := componentRestorers[restoreKey{id, saved.Version}]
	if !ok {
		return nil, fmt.Errorf("unsupported component %q version %d", id, saved.Version)
	}
	return restore(saved)
}
func marshalSaved(value any) (SavedComponent, error) {
	data, err := json.Marshal(value)
	return SavedComponent{Version: 1, Data: data}, err
}

func IdleComponents(components []Component, tick uint64) []Component {
	result := make([]Component, 0, len(components))
	for _, c := range components {
		if _, ok := c.(transientComponent); ok {
			continue
		}
		if idle, ok := c.(idleComponent); ok {
			c = idle.idleForRestore(tick)
		}
		result = append(result, c)
	}
	return result
}

func decodeSaved(data []byte, target any) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return fmt.Errorf("null component payload")
	}
	// Every field in a v1 DTO is required (nullable fields still need a key).
	// This prevents truncated objects from silently becoming valid zero values.
	shape := reflect.TypeOf(target).Elem()
	if shape.Kind() == reflect.Struct {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(data, &fields); err != nil {
			return err
		}
		for i := 0; i < shape.NumField(); i++ {
			field := shape.Field(i)
			name := strings.Split(field.Tag.Get("json"), ",")[0]
			if name == "" {
				name = field.Name
			}
			if _, ok := fields[name]; !ok {
				return fmt.Errorf("missing saved field %s", name)
			}
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("trailing component data")
	}
	return nil
}
