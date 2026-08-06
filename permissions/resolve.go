package permissions

import (
	"github.com/discord-go/discord.go/snowflake"
)

// Overwrite represents a channel permission overwrite used for calculating permissions.
type Overwrite struct {
	ID    snowflake.ID
	Type  int // 0 for role, 1 for member
	Allow Permission
	Deny  Permission
}

// Calculate determines the permissions for a member in a specific channel.
// It follows Discord's exact precedence rules.
func Calculate(
	memberID snowflake.ID,
	guildID snowflake.ID,
	guildOwnerID snowflake.ID,
	baseRolePermissions Permission,
	memberRoleIDs []snowflake.ID,
	memberRolePermissions []Permission,
	overwrites []Overwrite,
) Permission {
	// 1. If memberID == guildOwnerID, return all permissions.
	if memberID == guildOwnerID {
		return ^Permission(0)
	}

	// 2. Base = baseRolePermissions | bitwise OR of all memberRolePermissions.
	base := baseRolePermissions
	for _, perm := range memberRolePermissions {
		base.Add(perm)
	}

	// 3. If Base has Administrator, return max uint64.
	if base.Has(Administrator) {
		return ^Permission(0)
	}

	// 4. Apply overwrites for the @everyone role (deny then allow).
	for _, ow := range overwrites {
		if ow.ID == guildID {
			base.Remove(ow.Deny)
			base.Add(ow.Allow)
			break
		}
	}

	// 5. Apply overwrites for all of the user's specific roles (deny all, then allow all).
	var roleAllow, roleDeny Permission
	for _, ow := range overwrites {
		// 0 represents Role overwrite type
		if ow.Type != 0 || ow.ID == guildID {
			continue
		}

		for _, roleID := range memberRoleIDs {
			if ow.ID == roleID {
				roleDeny.Add(ow.Deny)
				roleAllow.Add(ow.Allow)
				break
			}
		}
	}
	base.Remove(roleDeny)
	base.Add(roleAllow)

	// 6. Apply overwrites for the specific user (deny then allow).
	for _, ow := range overwrites {
		// 1 represents Member overwrite type
		if ow.ID == memberID && ow.Type == 1 {
			base.Remove(ow.Deny)
			base.Add(ow.Allow)
			break
		}
	}

	return base
}
