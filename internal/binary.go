package internal

import (
	"strconv"
	"strings"
)

/*
IntToBinaryString converts an integer to its binary string representation, padded to a specified length.

Parameters:
  - n:
    The integer to convert.
  - length:
    The length to which the binary string should be padded.

Returns:
- A binary string representation of the integer, padded to the specified length.
*/
func IntToBinaryString(n int, length int) string {
	binaryStr := strconv.FormatInt(int64(n), 2)
	return padString(binaryStr, length)
}

/*
BinaryStringToInt converts a binary string to its integer representation.

Parameters:
  - binaryStr:
    The binary string to convert.

Returns:
- The integer representation of the binary string.
- An error if the conversion fails.
*/
func BinaryStringToInt(binaryStr string) (int64, error) {
	n, err := strconv.ParseInt(binaryStr, 2, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

/*
padString pads a string with leading zeros to a specified length.

Parameters:
  - s:
    The string to pad.
  - length:
    The length to which the string should be padded.

Returns:
- The padded string.
*/
func padString(s string, length int) string {
	if len(s) >= length {
		return s
	}
	return strings.Repeat("0", length-len(s)) + s
}
