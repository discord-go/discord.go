package gateway

import (
	"reflect"
	"sync"
)

// Dispatcher manages event handlers and dispatches events.
type Dispatcher struct {
	mu       sync.RWMutex
	handlers map[reflect.Type][]interface{}
}

// NewDispatcher creates a new event dispatcher.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[reflect.Type][]interface{}),
	}
}

// AddHandler adds an event handler. The handler must be a function taking one argument (the event type).
func (d *Dispatcher) AddHandler(handler interface{}) {
	d.mu.Lock()
	defer d.mu.Unlock()

	typ := reflect.TypeOf(handler)
	if typ.Kind() != reflect.Func || typ.NumIn() != 1 {
		panic("handler must be a function with one argument")
	}

	argType := typ.In(0)
	d.handlers[argType] = append(d.handlers[argType], handler)
}

// Dispatch dispatches an event to all registered handlers for its type.
func (d *Dispatcher) Dispatch(event interface{}) {
	d.mu.RLock()
	eventType := reflect.TypeOf(event)
	handlers := append([]interface{}(nil), d.handlers[eventType]...)
	d.mu.RUnlock()

	for _, handler := range handlers {
		func() {
			defer func() { _ = recover() }()
			reflect.ValueOf(handler).Call([]reflect.Value{reflect.ValueOf(event)})
		}()
	}
}
