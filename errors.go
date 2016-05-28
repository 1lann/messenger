package messenger

// ParseError is returned if an error due to parsing occurs.
type ParseError struct {
	message string
}

// Error returns the detailed parse error message.
func (p ParseError) Error() string {
	return "messenger: " + p.message
}
