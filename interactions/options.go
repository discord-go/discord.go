package interactions

import (
	"encoding/json"
	"strconv"

	"github.com/discord-go/discord.go/snowflake"
)

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

// OptionValue is a convenience accessor for the raw option value.
func (o *ApplicationCommandInteractionDataOption) OptionValue() interface{} {
	if o == nil {
		return nil
	}
	return o.Value
}

// String returns the option value coerced to a string. It returns an empty
// string when the option is nil or the value is not a string.
func (o *ApplicationCommandInteractionDataOption) String() string {
	if o == nil || o.Value == nil {
		return ""
	}
	if value, ok := o.Value.(string); ok {
		return value
	}
	return ""
}

// Int returns the option value coerced to an int64. Fractional numeric
// values truncate toward zero (3.7 becomes 3), matching float-to-int
// conversion elsewhere in the library. It returns zero when the option is
// nil or the value is not numeric.
func (o *ApplicationCommandInteractionDataOption) Int() int64 {
	if o == nil || o.Value == nil {
		return 0
	}
	switch value := o.Value.(type) {
	case json.Number:
		if result, err := value.Int64(); err == nil {
			return result
		}
		// Fractional number: truncate toward zero via Float64.
		f, _ := value.Float64()
		return int64(f)
	case int:
		return int64(value)
	case int64:
		return value
	case float64:
		return int64(value)
	case string:
		result, _ := strconv.ParseInt(value, 10, 64)
		return result
	default:
		return 0
	}
}

// Float returns the option value coerced to a float64. It returns zero when
// the option is nil or the value is not numeric.
func (o *ApplicationCommandInteractionDataOption) Float() float64 {
	if o == nil || o.Value == nil {
		return 0
	}
	switch value := o.Value.(type) {
	case json.Number:
		result, _ := value.Float64()
		return result
	case float64:
		return value
	case string:
		result, _ := strconv.ParseFloat(value, 64)
		return result
	default:
		return 0
	}
}

// Bool returns the option value coerced to a bool. It returns false when the
// option is nil or the value is not boolean.
func (o *ApplicationCommandInteractionDataOption) Bool() bool {
	if o == nil || o.Value == nil {
		return false
	}
	if value, ok := o.Value.(bool); ok {
		return value
	}
	value, _ := strconv.ParseBool(toString(o.Value))
	return value
}

// Snowflake returns the option value coerced to a snowflake.ID. It returns
// zero when the option is nil or the value is not a valid snowflake.
func (o *ApplicationCommandInteractionDataOption) Snowflake() snowflake.ID {
	if o == nil || o.Value == nil {
		return 0
	}
	var value string
	switch raw := o.Value.(type) {
	case string:
		value = raw
	case json.Number:
		value = raw.String()
	case float64:
		value = strconv.FormatUint(uint64(raw), 10)
	default:
		value = toString(raw)
	}
	id, _ := snowflake.Parse(value)
	return id
}

// Subcommand returns the option name when this option is a subcommand.
func (o *ApplicationCommandInteractionDataOption) Subcommand() string {
	if o != nil && o.Type == ApplicationCommandOptionTypeSubCommand {
		return o.Name
	}
	return ""
}

// SubcommandGroup returns the option name when this option is a subcommand group.
func (o *ApplicationCommandInteractionDataOption) SubcommandGroup() string {
	if o != nil && o.Type == ApplicationCommandOptionTypeSubCommandGroup {
		return o.Name
	}
	return ""
}

// NestedOptions returns the nested options of the option.
func (o *ApplicationCommandInteractionDataOption) NestedOptions() []ApplicationCommandInteractionDataOption {
	if o == nil {
		return nil
	}
	return o.Options
}

func toString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case json.Number:
		return t.String()
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	default:
		return ""
	}
}
