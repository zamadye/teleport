package utils

import (
	"slices"
	"strings"
)

// ArgValueRedactor redacts a sensitive CLI flag value.
type ArgValueRedactor func(value string) string

// RedactFlagArgs returns a copy of args with values redacted for any flag key
// present in redactors.
//
// Supported formats:
//   - --flag=value
//   - --flag value
func RedactFlagArgs(args []string, redactors map[string]ArgValueRedactor) []string {
	redactedArgs := slices.Clone(args)
	for i := 0; i < len(redactedArgs); i++ {
		arg := redactedArgs[i]

		flag, value, hasInlineValue := strings.Cut(arg, "=")
		if redactor, ok := redactors[flag]; ok && hasInlineValue {
			redactedArgs[i] = flag + "=" + redactor(value)
			continue
		}

		if redactor, ok := redactors[arg]; ok && i+1 < len(redactedArgs) {
			redactedArgs[i+1] = redactor(redactedArgs[i+1])
			i++
		}
	}

	return redactedArgs
}
