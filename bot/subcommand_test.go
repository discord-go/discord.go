package bot

import (
	"testing"

	"github.com/discord-go/discord.go/interactions"
)

func TestInteractionContext_SubcommandExtraction(t *testing.T) {
	ic := newInteractionContext(BaseContext{}, &interactions.Interaction{
		Type: interactions.InteractionTypeApplicationCommand,
		Data: mustMarshal(t, interactions.ApplicationCommandInteractionData{
			Name: "cmd",
			Options: []interactions.ApplicationCommandInteractionDataOption{
				{
					Name: "sub",
					Type: interactions.ApplicationCommandOptionTypeSubCommand,
					Options: []interactions.ApplicationCommandInteractionDataOption{
						{Name: "arg1", Type: interactions.ApplicationCommandOptionTypeString, Value: "x"},
						{Name: "arg2", Type: interactions.ApplicationCommandOptionTypeInteger, Value: 42},
					},
				},
			},
		}),
	})

	if got := ic.Subcommand(); got != "sub" {
		t.Errorf("Subcommand = %q, want sub", got)
	}
	if got := ic.SubcommandGroup(); got != "" {
		t.Errorf("SubcommandGroup = %q, want empty", got)
	}
	opt := ic.SubcommandOption()
	if opt == nil || opt.Name != "sub" {
		t.Fatalf("SubcommandOption = %+v, want sub", opt)
	}
	args := ic.SubcommandOptions()
	if len(args) != 2 {
		t.Fatalf("SubcommandOptions len = %d, want 2", len(args))
	}
	if args[0].Name != "arg1" || args[0].String() != "x" {
		t.Errorf("arg1 = %+v, want x", args[0])
	}
	if args[1].Name != "arg2" || args[1].Int() != 42 {
		t.Errorf("arg2 = %+v, want 42", args[1])
	}
}

func TestInteractionContext_SubcommandGroupExtraction(t *testing.T) {
	ic := newInteractionContext(BaseContext{}, &interactions.Interaction{
		Type: interactions.InteractionTypeApplicationCommand,
		Data: mustMarshal(t, interactions.ApplicationCommandInteractionData{
			Name: "cmd",
			Options: []interactions.ApplicationCommandInteractionDataOption{
				{
					Name: "grp",
					Type: interactions.ApplicationCommandOptionTypeSubCommandGroup,
					Options: []interactions.ApplicationCommandInteractionDataOption{
						{
							Name: "sub",
							Type: interactions.ApplicationCommandOptionTypeSubCommand,
							Options: []interactions.ApplicationCommandInteractionDataOption{
								{Name: "arg1", Type: interactions.ApplicationCommandOptionTypeString, Value: "y"},
							},
						},
					},
				},
			},
		}),
	})

	if got := ic.SubcommandGroup(); got != "grp" {
		t.Errorf("SubcommandGroup = %q, want grp", got)
	}
	if got := ic.Subcommand(); got != "sub" {
		t.Errorf("Subcommand = %q, want sub", got)
	}
	args := ic.SubcommandOptions()
	if len(args) != 1 || args[0].Name != "arg1" || args[0].String() != "y" {
		t.Errorf("SubcommandOptions = %+v, want [arg1=y]", args)
	}
}

func TestInteractionContext_NoSubcommand(t *testing.T) {
	ic := newInteractionContext(BaseContext{}, &interactions.Interaction{
		Type: interactions.InteractionTypeApplicationCommand,
		Data: mustMarshal(t, interactions.ApplicationCommandInteractionData{
			Name: "cmd",
			Options: []interactions.ApplicationCommandInteractionDataOption{
				{Name: "plain", Type: interactions.ApplicationCommandOptionTypeString, Value: "z"},
			},
		}),
	})
	if got := ic.Subcommand(); got != "" {
		t.Errorf("Subcommand = %q, want empty", got)
	}
	if got := ic.SubcommandOption(); got != nil {
		t.Errorf("SubcommandOption = %+v, want nil", got)
	}
	if got := ic.SubcommandOptions(); got != nil {
		t.Errorf("SubcommandOptions = %+v, want nil", got)
	}
}
