package msgservice

// ErrorString return empty string on nil error,
// otherwise return Error() string
func ErrorString(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}
