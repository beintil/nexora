package trimmed

import (
	"runtime"
	"strings"
)

// GetTrimmedStack returns the trimmed stack trace
func GetTrimmedStack(skip int) string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	stack := string(buf[:n])

	lines := strings.Split(stack, "\n")
	var trimmed []string
	for i := skip + 1; i < len(lines); i += 2 {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "runtime.") || strings.HasPrefix(line, "testing.") {
			continue
		}
		if i+1 < len(lines) {
			line += " " + strings.TrimSpace(lines[i+1])
		}
		trimmed = append(trimmed, line)
		if len(trimmed) == 5 {
			break
		}
	}
	return strings.Join(trimmed, "\n")
}
