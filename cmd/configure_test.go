package cmd

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateConfigValuesPreservesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cipr.toml")
	original := `# keep this comment
proxy = "http://old.example" # keep inline comment
aws_endpoint = "https://old.example/ranges"
custom_key = "untouched"

[custom]
debug = "section value"
`
	require.NoError(t, os.WriteFile(path, []byte(original), 0o640))

	err := updateConfigValues(path, map[string]any{
		"aws_endpoint": "https://new.example/ranges#fragment",
		"debug":        true,
		"proxy":        "http://proxy.example:8080",
	})
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	text := string(data)
	assert.Contains(t, text, "# keep this comment")
	assert.Contains(t, text, `proxy = 'http://proxy.example:8080' # keep inline comment`)
	assert.Contains(t, text, `aws_endpoint = 'https://new.example/ranges#fragment'`)
	assert.Contains(t, text, `custom_key = "untouched"`)
	assert.Contains(t, text, "debug = true\n\n[custom]")
	assert.Contains(t, text, `debug = "section value"`)

	var decoded map[string]any
	require.NoError(t, toml.Unmarshal(data, &decoded))
	assert.Equal(t, true, decoded["debug"])
	assert.Equal(t, "untouched", decoded["custom_key"])

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o640), info.Mode().Perm())
}

func TestUpdateConfigValuesValidationFailureDoesNotChangeFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cipr.toml")
	original := []byte("proxy = \"\"\n")
	require.NoError(t, os.WriteFile(path, original, 0o600))

	err := updateConfigValues(path, map[string]any{"proxy": make(chan int)})
	require.Error(t, err)

	got, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	assert.Equal(t, original, got)
}

func TestInlineAssignmentCommentIgnoresHashesInsideStrings(t *testing.T) {
	assert.Empty(t, inlineAssignmentComment(`endpoint = "https://example.test/#fragment"`))
	assert.Equal(t, "# comment", inlineAssignmentComment(`endpoint = "https://example.test/#fragment" # comment`))
}

func TestConfiguredSourceKeysAreSorted(t *testing.T) {
	keys := configuredSourceKeys()
	assert.True(t, strings.Contains(strings.Join(keys, ","), "cloudflare_ipv4"))
	assert.Contains(t, keys, "gcp")
	assert.True(t, sort.StringsAreSorted(keys))
}
