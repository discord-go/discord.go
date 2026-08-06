package guilds

import (
	"testing"
)

func TestFeature(t *testing.T) {
	if FeatureAnimatedBanner != "ANIMATED_BANNER" {
		t.Errorf("expected ANIMATED_BANNER, got %s", FeatureAnimatedBanner)
	}
	if FeatureRoleIcons != "ROLE_ICONS" {
		t.Errorf("expected ROLE_ICONS, got %s", FeatureRoleIcons)
	}
}
