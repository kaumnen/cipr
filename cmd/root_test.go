package cmd

import (
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
