package comfy

import (
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"time"

	"github.com/Mendes11/comfy/parsers"
)

type fileChangeSubscription[T comparable] struct {
	id string
	ch chan FileValueChangedEvent[T]
}

// FileField loads a configuration field from a filepath
type FileField[T comparable] struct {
	name          string
	fc            FieldConfig
	parserOpts    parsers.ParserOpts
	path          string
	subscriptions map[string]fileChangeSubscription[T]
	listening     bool
	currentValue  T
}

func (f *FileField[T]) ComfySetup(fieldName string, fc FieldConfig) error {
	f.name = fieldName
	f.fc = fc
	// Disable so comfy doesn't load the contents from the path living on _FILE.
	f.fc.AutoFileLookupEnabled = false
	f.subscriptions = make(map[string]fileChangeSubscription[T])
	return nil
}

func (f *FileField[T]) ComfyParse(s string, opts parsers.ParserOpts) error {
	if _, err := os.Stat(s); err != nil {
		return fmt.Errorf("comfy.FileField#ComfyParse failed to parse: %w", err)
	}
	f.path = s
	f.parserOpts = opts
	return nil
}

func (f *FileField[T]) notifyChange(oldValue, newValue T) {
	for id, sub := range f.subscriptions {
		sub.ch <- FileValueChangedEvent[T]{id, oldValue, newValue}
	}
}

func (f *FileField[T]) listenToChanges() {
	t := time.NewTicker(time.Second)
	defer func() {
		for _, sub := range f.subscriptions {
			close(sub.ch)
		}
	}()
	defer t.Stop()

	for range t.C {
		v, err := f.Value()
		if err != nil {
			slog.Error("[Comfy] Listener failed to obtain value", "field", f.name, "error", err.Error())
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
	err = parse(f.name, stringVal, vValue.Elem(), f.fc)
	if err != nil {
		return *v, fmt.Errorf("parse failed: %w", err)
	}

	return *v, nil
}

type FileValueChangedEvent[T comparable] struct {
	SubscriptionId string
	OldValue       T
	NewValue       T
}

func (f *FileField[T]) SubscribeToChange(id string) (<-chan FileValueChangedEvent[T], error) {
	v, err := f.Value()
	if err != nil {
		return nil, fmt.Errorf("comfy.SubscribeToChange failed to load initial value: %w", err)
	}
	f.currentValue = v
	ch := make(chan FileValueChangedEvent[T], 1)
	f.subscriptions[id] = fileChangeSubscription[T]{
		id: id,
		ch: ch,
	}
	if !f.listening {
		f.listening = true
		go f.listenToChanges()
	}
	return ch, nil
}
