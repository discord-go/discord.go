package permissions

import (
	"testing"
)

func TestBuild(t *testing.T) {
	p := Build(SendMessages, KickMembers)
	expected := SendMessages | KickMembers
	if p != expected {
		t.Errorf("Expected %d, got %d", expected, p)
	}

	p2 := Build()
	if p2 != 0 {
		t.Errorf("Expected 0, got %d", p2)
	}
}

func TestBuilder(t *testing.T) {
	b := NewBuilder(SendMessages).Add(KickMembers, BanMembers).Remove(BanMembers)
	p := b.Build()

	expected := SendMessages | KickMembers
	if p != expected {
		t.Errorf("Expected %d, got %d", expected, p)
	}

	b2 := NewBuilder()
	if b2.Build() != 0 {
		t.Errorf("Expected 0, got %d", b2.Build())
	}
}
