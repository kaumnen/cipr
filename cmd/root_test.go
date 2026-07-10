package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveIPType(t *testing.T) {
	tests := []struct {
		name       string
		ipv4, ipv6 bool
		want       string
	}{
		{name: "neither", want: "both"},
		{name: "ipv4", ipv4: true, want: "ipv4"},
		{name: "ipv6", ipv6: true, want: "ipv6"},
		{name: "both", ipv4: true, ipv6: true, want: "both"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, resolveIPType(tt.ipv4, tt.ipv6))
		})
	}
}

func TestResolveVerbosityFromConfig(t *testing.T) {
	old := viper.Get("verbose_mode")
	viper.Set("verbose_mode", "mini")
	t.Cleanup(func() { viper.Set("verbose_mode", old) })

	got, err := resolveVerbosity(&cobra.Command{})
	require.NoError(t, err)
	assert.Equal(t, "mini", got)
}

func TestResolveVerbosityRejectsInvalidValue(t *testing.T) {
	old := viper.Get("verbose_mode")
	viper.Set("verbose_mode", "loud")
	t.Cleanup(func() { viper.Set("verbose_mode", old) })

	_, err := resolveVerbosity(&cobra.Command{})
	require.Error(t, err)
}

func TestEnsureConfigFileCreatesDefaultsForConfigure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "cipr.toml")
	require.NoError(t, ensureConfigFile(path, true))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	text := string(data)
	assert.Contains(t, text, `proxy = ""`)
	assert.Contains(t, text, "debug = false")
	assert.Contains(t, text, "cloudflare_ipv4_endpoint")
}

func TestEnsureConfigFileLeavesMissingCustomConfigForNormalCommands(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.toml")
	require.NoError(t, ensureConfigFile(path, false))
	_, err := os.Stat(path)
	assert.True(t, os.IsNotExist(err))
}

func TestUsesConfiguredSources(t *testing.T) {
	assert.True(t, usesConfiguredSources("config"))
	assert.True(t, usesConfiguredSources("hosted"))
	assert.False(t, usesConfiguredSources("https://example.test/ranges"))
	assert.False(t, usesConfiguredSources("ranges.txt"))
}
