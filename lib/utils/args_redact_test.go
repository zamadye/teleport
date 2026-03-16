package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRedactFlagArgs(t *testing.T) {
	t.Parallel()

	mask := func(v string) string {
		return strings.Repeat("*", len(v))
	}

	original := []string{
		"node",
		"configure",
		"--token=secret-token",
		"--password",
		"super-secret",
		"--proxy=example.teleport.sh:443",
		"--unrelated",
	}

	got := RedactFlagArgs(original, map[string]ArgValueRedactor{
		"--token":    mask,
		"--password": mask,
	})

	require.Equal(t, []string{
		"node",
		"configure",
		"--token=************",
		"--password",
		"************",
		"--proxy=example.teleport.sh:443",
		"--unrelated",
	}, got)
	require.Equal(t, []string{
		"node",
		"configure",
		"--token=secret-token",
		"--password",
		"super-secret",
		"--proxy=example.teleport.sh:443",
		"--unrelated",
	}, original)
}

func TestRedactFlagArgsMissingFlagValue(t *testing.T) {
	t.Parallel()

	original := []string{"node", "configure", "--token"}

	got := RedactFlagArgs(original, map[string]ArgValueRedactor{
		"--token": func(v string) string { return "redacted:" + v },
	})

	require.Equal(t, original, got)
}
