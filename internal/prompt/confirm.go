package prompt

import (
	"strings"
)

func Confirm(question string) bool {
	answer := Ask(question + " (Y/n)?")
	answer = strings.ToLower(strings.TrimSpace(answer))

	// Empty answer defaults to yes
	if answer == "" || answer == "y" || answer == "yes" {
		return true
	}
	return false
}