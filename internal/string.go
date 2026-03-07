package internal

import "strings"

/*
BuildString constructs a single string from multiple input strings with a preallocated buffer size.

The `len` parameter specifies the initial buffer size to allocate for the builder.

The `str` variadic parameter allows passing multiple strings to concatenate.

Example usage:

	result := BuildString(64, "Hello", " ", "World")
	// result will be "Hello World"
*/
func BuildString(len int, str ...string) string {
	var builder strings.Builder
	builder.Grow(len)
	for _, s := range str {
		builder.WriteString(s)
	}
	return builder.String()
}
