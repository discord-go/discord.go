package snowflake

import (
	"testing"
	"time"
)

func TestID_Time(t *testing.T) {
	tests := []struct {
		name string
		id   ID
		want time.Time
	}{
		{
			name: "Discord Epoch",
			id:   ID(0), // 0 timestamp + Discord Epoch
			want: time.Unix(DiscordEpoch/1000, 0).UTC(),
		},
		{
			name: "1 Second after Epoch",
			id:   ID(uint64(1000) << 22),
			want: time.Unix((DiscordEpoch+1000)/1000, 0).UTC(),
		},
		{
			name: "Known Snowflake",
			id:   ID(175928847299117063),
			want: time.Date(2016, time.April, 30, 11, 18, 25, 796000000, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.Time(); !got.Equal(tt.want) {
				t.Errorf("ID.Time() = %v, want %v", got, tt.want)
			}
		})
	}
}
