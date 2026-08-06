package components

import (
	"encoding/json"
	"fmt"
)

type ActionRow struct {
	Components []Component `json:"components"`
}

func (r ActionRow) Type() ComponentType { return ComponentTypeActionRow }

type actionRowJSON struct {
	Type       ComponentType     `json:"type"`
	Components []json.RawMessage `json:"components"`
}

func (r *ActionRow) UnmarshalJSON(data []byte) error {
	var raw actionRowJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.Type != ComponentTypeActionRow {
		return fmt.Errorf("expected type %d (ActionRow), got %d", ComponentTypeActionRow, raw.Type)
	}
	components, err := decodeNested(raw.Components)
	if err != nil {
		return err
	}
	r.Components = components
	return nil
}

func (r ActionRow) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type       ComponentType `json:"type"`
		Components []Component   `json:"components"`
	}{
		Type:       r.Type(),
		Components: r.Components,
	})
}
