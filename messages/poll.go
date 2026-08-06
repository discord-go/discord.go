package messages

// Poll represents a poll in a message.
type Poll struct {
	Question         PollMedia    `json:"question"`
	Answers          []PollAnswer `json:"answers"`
	Expiry           string       `json:"expiry,omitempty"`
	AllowMultiselect bool         `json:"allow_multiselect"`
	LayoutType       int          `json:"layout_type"`
	Results          *PollResults `json:"results,omitempty"`
}

// PollMedia represents the media for a poll or a poll answer.
type PollMedia struct {
	Text  string `json:"text,omitempty"`
	Emoji *Emoji `json:"emoji,omitempty"`
}

// Emoji represents a simplified emoji for Polls.
type Emoji struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Animated bool   `json:"animated,omitempty"`
}

// PollAnswer represents an answer in a poll.
type PollAnswer struct {
	AnswerID  int       `json:"answer_id"`
	PollMedia PollMedia `json:"poll_media"`
}

// PollResults represents the results of a poll.
type PollResults struct {
	IsFinalized  bool          `json:"is_finalized"`
	AnswerCounts []AnswerCount `json:"answer_counts"`
}

// AnswerCount represents the vote count for a specific poll answer.
type AnswerCount struct {
	ID      int  `json:"id"`
	Count   int  `json:"count"`
	MeVoted bool `json:"me_voted"`
}
