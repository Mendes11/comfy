package comfy

import (
	"fmt"
	"log/slog"
	"reflect"
	"time"
)

type fileChangeSubscription[T comparable] struct {
	id       string
	callback func(string, T, T)
}

// FileField loads a configuration field from a filepath
type FileField[T comparable] struct {
	field         Field
	path          string
	subscriptions map[string]fileChangeSubscription[T]
	listening     bool
	currentValue  T
}

func NewFileField[T comparable](field Field, path string) *FileField[T] {
	return &FileField[T]{
		field:         field,
		path:          path,
		subscriptions: make(map[string]fileChangeSubscription[T]),
	}
}

func (f *FileField[T]) Setup(field Field, path string) {
	f.field = field
	f.path = path
	f.subscriptions = make(map[string]fileChangeSubscription[T])
}

func (f *FileField[T]) notifyChange(oldValue, newValue T) {
	for id, sub := range f.subscriptions {
		sub.callback(id, oldValue, newValue)
	}
}

func (f *FileField[T]) listenToChanges() {
	t := time.NewTicker(time.Second)
	defer t.Stop()

	for range t.C {
		v, err := f.Value()
		if err != nil {
			slog.Error("[Comfy] Listener failed to obtain value", "field", f.field.structField.Name, "error", err.Error())
			continue
		}
		if v != f.currentValue {
			f.notifyChange(f.currentValue, v)
		}
		f.currentValue = v
	}
}

func (f *FileField[T]) Value() (T, error) {
	v := new(T)
	stringVal, err := loadFromPath(f.path)
	if err != nil {
		return *v, fmt.Errorf("loadFromPath failed: %w", err)
	}
	vValue := reflect.ValueOf(v)
	err = parse(f.field.structField.Name, stringVal, vValue.Elem())
	if err != nil {
		return *v, fmt.Errorf("parse failed: %w", err)
	}

	return *v, nil
}

func (f *FileField[T]) SubscribeToChange(id string, sub func(id string, oldValue, newValue T)) error {
	v, err := f.Value()
	if err != nil {
		return fmt.Errorf("comfy.SubscribeToChange failed to load initial value: %w", err)
	}
	f.currentValue = v
	f.subscriptions[id] = fileChangeSubscription[T]{
		id:       id,
		callback: sub,
	}
	if !f.listening {
		f.listening = true
		go f.listenToChanges()
	}
	return nil
}
