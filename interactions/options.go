package interactions

// ApplicationCommandInteractionDataOption represents an option in the application command interaction data.
type ApplicationCommandInteractionDataOption struct {
	Name    string                                    `json:"name"`
	Type    ApplicationCommandOptionType              `json:"type"`
	Value   interface{}                               `json:"value,omitempty"`
	Options []ApplicationCommandInteractionDataOption `json:"options,omitempty"`
	Focused bool                                      `json:"focused,omitempty"`
}

// ApplicationCommandInteractionData represents the data payload for an application command interaction.
type ApplicationCommandInteractionData struct {
	ID       string                                    `json:"id"`
	Name     string                                    `json:"name"`
	Type     int                                       `json:"type"`
	Options  []ApplicationCommandInteractionDataOption `json:"options,omitempty"`
	GuildID  string                                    `json:"guild_id,omitempty"`
	TargetID string                                    `json:"target_id,omitempty"`
}
