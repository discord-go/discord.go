package bot

import (
	"testing"

	"github.com/discord-go/discord.go/gateway"
)

func TestWithGatewayCapabilities(t *testing.T) {
	b := New("token", WithGatewayCapabilities(gateway.CapabilityChannelObfuscation))
	if b.gatewayCaps != gateway.CapabilityChannelObfuscation {
		t.Fatalf("gatewayCaps = %d, want %d", b.gatewayCaps, gateway.CapabilityChannelObfuscation)
	}

	b = New("token", WithGatewayCapabilities(1<<3), WithGatewayCapabilities(1<<4))
	if b.gatewayCaps != (1<<3 | 1<<4) {
		t.Errorf("repeated WithGatewayCapabilities calls should union, got %d", b.gatewayCaps)
	}

	if New("token").gatewayCaps != 0 {
		t.Error("default gateway capabilities should be zero")
	}
}

func TestNewFromConfigGatewayCapabilities(t *testing.T) {
	b := NewFromConfig(Config{GatewayCapabilities: int(gateway.CapabilityChannelObfuscation)})
	if b.gatewayCaps != gateway.CapabilityChannelObfuscation {
		t.Errorf("gatewayCaps = %d, want %d", b.gatewayCaps, gateway.CapabilityChannelObfuscation)
	}
}

func TestConfigFromEnvGatewayCapabilities(t *testing.T) {
	t.Setenv("BOT_GATEWAY_CAPABILITIES", "32768")
	config := ConfigFromEnv()
	if config.GatewayCapabilities != 32768 {
		t.Errorf("GatewayCapabilities = %d, want 32768", config.GatewayCapabilities)
	}
}
