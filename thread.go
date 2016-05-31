package messenger

// Thread represents a message thread. If IsGroup is false, ThreadID is the
// other user's ID.
type Thread struct {
	ThreadID string
	IsGroup  bool
}

// TODO: Implement helper functions
