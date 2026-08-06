package permissions_test

import (
	"testing"

	"github.com/discord-go/discord.go/permissions"
	"github.com/discord-go/discord.go/snowflake"
)

func TestCalculate(t *testing.T) {
	memberID := snowflake.ID(100)
	guildID := snowflake.ID(200)
	guildOwnerID := snowflake.ID(300)

	// max permission
	var maxPerm permissions.Permission = ^permissions.Permission(0)

	t.Run("Guild Owner gets all permissions", func(t *testing.T) {
		res := permissions.Calculate(guildOwnerID, guildID, guildOwnerID, 0, nil, nil, nil)
		if res != maxPerm {
			t.Errorf("expected max perm, got %d", res)
		}
	})

	t.Run("Administrator gets all permissions", func(t *testing.T) {
		res := permissions.Calculate(memberID, guildID, guildOwnerID, permissions.Administrator, nil, nil, nil)
		if res != maxPerm {
			t.Errorf("expected max perm, got %d", res)
		}
	})

	t.Run("Role Administrator gets all permissions", func(t *testing.T) {
		memberRoleIDs := []snowflake.ID{400}
		memberRolePerms := []permissions.Permission{permissions.Administrator}
		res := permissions.Calculate(memberID, guildID, guildOwnerID, 0, memberRoleIDs, memberRolePerms, nil)
		if res != maxPerm {
			t.Errorf("expected max perm, got %d", res)
		}
	})

	t.Run("Overwrites application logic", func(t *testing.T) {
		// Base has ViewChannel
		basePerms := permissions.ViewChannel
		memberRoleIDs := []snowflake.ID{401, 402}
		memberRolePerms := []permissions.Permission{permissions.SendMessages, permissions.AddReactions}

		// Initial base = ViewChannel | SendMessages | AddReactions

		overwrites := []permissions.Overwrite{
			// @everyone overwrite: Deny ViewChannel, Allow ReadMessageHistory
			{
				ID:    guildID,
				Type:  0,
				Deny:  permissions.ViewChannel,
				Allow: permissions.ReadMessageHistory,
			},
			// Role 401 overwrite: Deny SendMessages, Allow EmbedLinks
			{
				ID:    401,
				Type:  0,
				Deny:  permissions.SendMessages,
				Allow: permissions.EmbedLinks,
			},
			// Member overwrite: Deny AddReactions, Allow AttachFiles
			{
				ID:    memberID,
				Type:  1,
				Deny:  permissions.AddReactions,
				Allow: permissions.AttachFiles,
			},
		}

		// Resolution order:
		// Base: ViewChannel | SendMessages | AddReactions
		// @everyone: Deny ViewChannel, Allow ReadMessageHistory -> Base = SendMessages | AddReactions | ReadMessageHistory
		// Roles: Deny SendMessages, Allow EmbedLinks -> Base = AddReactions | ReadMessageHistory | EmbedLinks
		// User: Deny AddReactions, Allow AttachFiles -> Base = ReadMessageHistory | EmbedLinks | AttachFiles

		expected := permissions.ReadMessageHistory | permissions.EmbedLinks | permissions.AttachFiles

		res := permissions.Calculate(memberID, guildID, guildOwnerID, basePerms, memberRoleIDs, memberRolePerms, overwrites)
		if res != expected {
			t.Errorf("expected %d, got %d", expected, res)
		}
	})

	t.Run("Role overwrites deny all then allow all", func(t *testing.T) {
		memberRoleIDs := []snowflake.ID{401, 402}
		memberRolePerms := []permissions.Permission{0, 0}

		overwrites := []permissions.Overwrite{
			{
				ID:    401,
				Type:  0,
				Deny:  permissions.SendMessages,
				Allow: permissions.ViewChannel,
			},
			{
				ID:    402,
				Type:  0,
				Deny:  permissions.ViewChannel,
				Allow: permissions.SendMessages,
			},
		}

		expected := permissions.ViewChannel | permissions.SendMessages

		res := permissions.Calculate(memberID, guildID, guildOwnerID, 0, memberRoleIDs, memberRolePerms, overwrites)
		if res != expected {
			t.Errorf("expected %d, got %d", expected, res)
		}
	})
}
