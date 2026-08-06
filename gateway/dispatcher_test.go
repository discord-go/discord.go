package gateway

import (
	"testing"
)

func TestDispatcher_AddHandler(t *testing.T) {
	d := NewDispatcher()

	called := false
	handler := func(e []byte) {
		called = true
		if string(e) != "test" {
			t.Errorf("expected test, got %s", string(e))
		}
	}

	d.AddHandler(handler)

	d.Dispatch([]byte("test"))
	if !called {
		t.Errorf("handler was not called")
	}

	// Test dispatch to non-registered type
	d.Dispatch(123)
}

func TestDispatcher_InvalidHandler_NotFunc(t *testing.T) {
	d := NewDispatcher()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for invalid handler")
		}
	}()
	d.AddHandler("not a function")
}

func TestDispatcher_InvalidHandler_Signature(t *testing.T) {
	d := NewDispatcher()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for invalid handler signature")
		}
	}()
	d.AddHandler(func(a, b int) {})
}
