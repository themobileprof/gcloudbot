package process

import (
	"bufio"
	"os"
	"strings"
)

var ReadInput = func() string {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}
