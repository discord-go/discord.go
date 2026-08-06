package permissions

import (
	"testing"
)

func TestPermission_Add(t *testing.T) {
	var p Permission

	p.Add(SendMessages)
	if p != SendMessages {
		t.Errorf("Expected %d, got %d", SendMessages, p)
	}

	p.Add(ManageChannels | KickMembers)
	expected := SendMessages | ManageChannels | KickMembers
	if p != expected {
		t.Errorf("Expected %d, got %d", expected, p)
	}
}

func TestPermission_Remove(t *testing.T) {
	p := SendMessages | ManageChannels | KickMembers

	p.Remove(ManageChannels)
	expected := SendMessages | KickMembers
	if p != expected {
		t.Errorf("Expected %d, got %d", expected, p)
	}

	p.Remove(SendMessages | KickMembers)
	if p != 0 {
		t.Errorf("Expected 0, got %d", p)
	}
}

func TestPermission_Has(t *testing.T) {
	p := SendMessages | KickMembers

	if !p.Has(SendMessages) {
		t.Error("Expected Has(SendMessages) to be true")
	}
	if !p.Has(KickMembers) {
		t.Error("Expected Has(KickMembers) to be true")
	}
	if !p.Has(SendMessages | BanMembers) {
		t.Error("Expected Has(SendMessages | BanMembers) to be true")
	}
	if p.Has(BanMembers) {
		t.Error("Expected Has(BanMembers) to be false")
	}
	if p.Has(0) {
		t.Error("Expected Has(0) to be false")
	}
}

func TestPermission_HasAll(t *testing.T) {
	p := SendMessages | KickMembers

	if !p.HasAll(SendMessages) {
		t.Error("Expected HasAll(SendMessages) to be true")
	}
	if !p.HasAll(SendMessages | KickMembers) {
		t.Error("Expected HasAll(SendMessages | KickMembers) to be true")
	}
	if p.HasAll(SendMessages | BanMembers) {
		t.Error("Expected HasAll(SendMessages | BanMembers) to be false")
	}
	if p.HasAll(BanMembers) {
		t.Error("Expected HasAll(BanMembers) to be false")
	}
	if !p.HasAll(0) {
		t.Error("Expected HasAll(0) to be true")
	}
}
