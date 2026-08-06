package intents

import "testing"

func TestIntent_Add(t *testing.T) {
	var i Intent
	i.Add(Guilds)
	if i != Guilds {
		t.Errorf("expected %d, got %d", Guilds, i)
	}
	i.Add(GuildMembers)
	if i != Guilds|GuildMembers {
		t.Errorf("expected %d, got %d", Guilds|GuildMembers, i)
	}
}

func TestIntent_Remove(t *testing.T) {
	i := Guilds | GuildMembers
	i.Remove(GuildMembers)
	if i != Guilds {
		t.Errorf("expected %d, got %d", Guilds, i)
	}
	i.Remove(Guilds)
	if i != 0 {
		t.Errorf("expected %d, got %d", 0, i)
	}
}

func TestIntent_Has(t *testing.T) {
	i := Guilds | GuildMembers
	if !i.Has(Guilds) {
		t.Errorf("expected to have Guilds")
	}
	if !i.Has(GuildMembers) {
		t.Errorf("expected to have GuildMembers")
	}
	if i.Has(GuildBans) {
		t.Errorf("expected not to have GuildBans")
	}

	i2 := Guilds | GuildMembers | DirectMessages
	if !i2.Has(Guilds | GuildMembers) {
		t.Errorf("expected to have Guilds and GuildMembers")
	}
}
