package process

import (
	"fmt"
	"strings"
)

var PromptYesNo = func(prompt string) bool {
	fmt.Printf("%s (y/n): ", prompt)
	input := strings.ToLower(ReadInput())
	return input == "y" || input == "yes"
}
