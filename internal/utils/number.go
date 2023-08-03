package utils

import (
	"regexp"
	"strconv"
)

func ParseStringToInt(inputStr string) (int, error) {
	// Use a regular expression to remove all non-numeric characters
	regex := regexp.MustCompile(`[^\d]`)
	numericPart := regex.ReplaceAllString(inputStr, "")

	// Convert the numeric part to an integer
	resultInt, err := strconv.Atoi(numericPart)
	if err != nil {
		return 0, err
	}

	return resultInt, nil
}
