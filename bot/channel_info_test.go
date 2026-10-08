package bot

import (
	"testing"
	"time"
)

// TestChannelInfoDispatch verifies the mixed-case "Channel Info" dispatch
// reaches both its typed handler and a generic subscription, which is stored
// under the uppercased event name.
func TestChannelInfoDispatch(t *testing.T) {
	b := New("token")

	typed := make(chan string, 1)
	b.OnChannelInfo(func(ctx *ChannelInfoContext) {
		if len(ctx.Channels) == 0 || ctx.Channels[0].Status == nil {
			typed <- "<nil>"
			return
		}
		typed <- *ctx.Channels[0].Status
	})

	generic := make(chan string, 1)
	b.On("Channel Info", func(ctx *EventContext) { generic <- ctx.Name })

	feedDispatch(t, b, "Channel Info", `{"guild_id":"300","channels":[{"id":"100","status":"standup"}]}`)

	select {
	case got := <-typed:
		if got != "standup" {
			t.Errorf("status = %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("typed Channel Info handler did not run")
	}
	select {
	case got := <-generic:
		if got != "Channel Info" {
			t.Errorf("event name = %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("generic Channel Info handler did not run")
	}
}

func TestVoiceChannelStatusUpdateDispatch(t *testing.T) {
	b := New("token")

	typed := make(chan string, 1)
	b.OnVoiceChannelStatusUpdate(func(ctx *VoiceChannelStatusUpdateContext) {
		if ctx.Status == nil {
			typed <- "<nil>"
			return
		}
		typed <- *ctx.Status
	})

	feedDispatch(t, b, "VOICE_CHANNEL_STATUS_UPDATE", `{"id":"100","guild_id":"300","status":"brb"}`)
	select {
	case got := <-typed:
		if got != "brb" {
			t.Errorf("status = %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("VOICE_CHANNEL_STATUS_UPDATE handler did not run")
	}
}

func TestVoiceChannelStartTimeUpdateDispatch(t *testing.T) {
	b := New("token")

	typed := make(chan int, 1)
	b.OnVoiceChannelStartTimeUpdate(func(ctx *VoiceChannelStartTimeUpdateContext) {
		if ctx.VoiceStartTime == nil {
			typed <- -1
			return
		}
		typed <- *ctx.VoiceStartTime
	})

	feedDispatch(t, b, "VOICE_CHANNEL_START_TIME_UPDATE", `{"id":"100","guild_id":"300","voice_start_time":1756000000}`)
	select {
	case got := <-typed:
		if got != 1756000000 {
			t.Errorf("voice_start_time = %d", got)
		}
	case <-time.After(time.Second):
		t.Fatal("VOICE_CHANNEL_START_TIME_UPDATE handler did not run")
	}
}
