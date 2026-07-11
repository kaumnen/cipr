package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kaumnen/cipr/internal/utils"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const defaultConfiguredCacheTTL = "24h"

var configureCmd = &cobra.Command{
	Use:   "configure [source-key]",
	Short: "Show or update cipr configuration",
	Long: `Show or update cipr's managed configuration values.

Source-specific settings use the keys written to cipr.toml, including aws,
azure, cloudflare_ipv4, cloudflare_ipv6, digitalocean, gcp, github, and icloud.
With no update flags, the effective managed configuration is displayed.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runConfigure,
}

func init() {
	rootCmd.AddCommand(configureCmd)
	configureCmd.Flags().String("endpoint", "", "HTTP(S) endpoint for the selected source (empty resets to the default)")
	configureCmd.Flags().String("local-file", "", "Local data file for the selected source (empty clears the override)")
	configureCmd.Flags().String("cache-ttl", "", "Cache duration for the selected source (empty resets to 24h; 0s disables caching)")
}

func runConfigure(cmd *cobra.Command, args []string) error {
	var source string
	if len(args) == 1 {
		source = args[0]
		if _, ok := utils.DefaultEndpoints[source]; !ok {
			return fmt.Errorf("unknown source key %q (valid: %s)", source, strings.Join(configuredSourceKeys(), ", "))
		}
	}

	endpointChanged := cmd.Flags().Changed("endpoint")
	localFileChanged := cmd.Flags().Changed("local-file")
	cacheTTLChanged := cmd.Flags().Changed("cache-ttl")
	proxyChanged := cmd.Flags().Changed("proxy")
	debugChanged := cmd.Flags().Changed("debug")

	if (endpointChanged || localFileChanged || cacheTTLChanged) && source == "" {
		return fmt.Errorf("a source key is required with --endpoint, --local-file, or --cache-ttl")
	}
	if endpointChanged && localFileChanged {
		return fmt.Errorf("--endpoint and --local-file cannot be changed together")
	}

	updates := make(map[string]any)
	if endpointChanged {
		endpoint, err := cmd.Flags().GetString("endpoint")
		if err != nil {
			return err
		}
		if endpoint == "" {
			endpoint = utils.DefaultEndpoints[source]
		}
		if err := utils.ValidateHTTPURL(endpoint); err != nil {
			return err
		}
		updates[source+"_endpoint"] = endpoint
		updates[source+"_local_file"] = ""
	}
	if localFileChanged {
		localFile, err := cmd.Flags().GetString("local-file")
		if err != nil {
			return err
		}
		updates[source+"_local_file"] = localFile
	}
	if cacheTTLChanged {
		cacheTTL, err := cmd.Flags().GetString("cache-ttl")
		if err != nil {
			return err
		}
		if cacheTTL == "" {
			cacheTTL = defaultConfiguredCacheTTL
		}
		duration, err := time.ParseDuration(cacheTTL)
		if err != nil || duration < 0 {
			return fmt.Errorf("invalid cache TTL %q (use a non-negative Go duration such as 24h or 0s)", cacheTTL)
		}
		updates[source+"_cache_ttl"] = cacheTTL
	}
	if proxyChanged {
		proxy, err := cmd.Flags().GetString("proxy")
		if err != nil {
			return err
		}
		if err := utils.ValidateProxyURL(proxy); err != nil {
			return err
		}
		updates["proxy"] = proxy
	}
	if debugChanged {
		debug, err := cmd.Flags().GetBool("debug")
		if err != nil {
			return err
		}
		updates["debug"] = debug
	}

	if len(updates) > 0 {
		configPath := viper.ConfigFileUsed()
		if configPath == "" {
			return fmt.Errorf("cannot determine active config file")
		}
		if err := updateConfigValues(configPath, updates); err != nil {
			return err
		}
		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("reload config file: %w", err)
		}
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), "Updated config file:", configPath); err != nil {
			return fmt.Errorf("write configuration output: %w", err)
		}
	}

	return showEffectiveConfiguration(cmd.OutOrStdout(), source)
}

func configuredSourceKeys() []string {
	keys := make([]string, 0, len(utils.DefaultEndpoints))
	for key := range utils.DefaultEndpoints {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func showEffectiveConfiguration(w io.Writer, selectedSource string) error {
	if _, err := fmt.Fprintf(w, "config_file = %q\n", viper.ConfigFileUsed()); err != nil {
		return fmt.Errorf("write configuration output: %w", err)
	}
	proxy := viper.GetString("proxy")
	if proxy != "" {
		proxy = utils.SanitizeURL(proxy)
	} else if hasProxyEnvironment() {
		proxy = "<environment>"
	}
	if _, err := fmt.Fprintf(w, "proxy = %q\n", proxy); err != nil {
		return fmt.Errorf("write configuration output: %w", err)
	}
	if _, err := fmt.Fprintf(w, "debug = %t\n", viper.GetBool("debug")); err != nil {
		return fmt.Errorf("write configuration output: %w", err)
	}

	keys := configuredSourceKeys()
	if selectedSource != "" {
		keys = []string{selectedSource}
	}
	for _, source := range keys {
		endpoint := viper.GetString(source + "_endpoint")
		if endpoint == "" {
			endpoint = utils.DefaultEndpoints[source]
		}
		localFile := viper.GetString(source + "_local_file")
		cacheTTL := viper.GetString(source + "_cache_ttl")
		if parsed, err := time.ParseDuration(cacheTTL); cacheTTL == "" || err != nil || parsed < 0 {
			cacheTTL = defaultConfiguredCacheTTL
		}
		active := "endpoint"
		if localFile != "" {
			active = "local-file"
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return fmt.Errorf("write configuration output: %w", err)
		}
		if _, err := fmt.Fprintf(w, "%s_active_source = %q\n", source, active); err != nil {
			return fmt.Errorf("write configuration output: %w", err)
		}
		if _, err := fmt.Fprintf(w, "%s_endpoint = %q\n", source, endpoint); err != nil {
			return fmt.Errorf("write configuration output: %w", err)
		}
		if _, err := fmt.Fprintf(w, "%s_local_file = %q\n", source, localFile); err != nil {
			return fmt.Errorf("write configuration output: %w", err)
		}
		if _, err := fmt.Fprintf(w, "%s_cache_ttl = %q\n", source, cacheTTL); err != nil {
			return fmt.Errorf("write configuration output: %w", err)
		}
	}
	return nil
}

func hasProxyEnvironment() bool {
	for _, key := range []string{"HTTPS_PROXY", "https_proxy", "HTTP_PROXY", "http_proxy"} {
		if os.Getenv(key) != "" {
			return true
		}
	}
	return false
}

func updateConfigValues(configPath string, updates map[string]any) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("read config file: %w", err)
	}
	info, err := os.Stat(configPath)
	if err != nil {
		return fmt.Errorf("inspect config file: %w", err)
	}

	assignments := make(map[string]string, len(updates))
	for key, value := range updates {
		encoded, err := toml.Marshal(map[string]any{key: value})
		if err != nil {
			return fmt.Errorf("encode config value %s: %w", key, err)
		}
		assignments[key] = strings.TrimSpace(string(encoded))
	}

	lines := strings.Split(string(data), "\n")
	missing := make(map[string]struct{}, len(assignments))
	for key := range assignments {
		missing[key] = struct{}{}
	}
	firstTable := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") {
			firstTable = i
			break
		}
		key, ok := assignmentKey(line)
		if !ok {
			continue
		}
		assignment, managed := assignments[key]
		if !managed {
			continue
		}
		indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
		if comment := inlineAssignmentComment(line); comment != "" {
			assignment += " " + comment
		}
		lines[i] = indent + assignment
		delete(missing, key)
	}

	if len(missing) > 0 {
		keys := make([]string, 0, len(missing))
		for key := range missing {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		block := make([]string, 0, len(keys)+2)
		for _, key := range keys {
			block = append(block, assignments[key])
		}

		insertAt := firstTable
		if insertAt < 0 {
			insertAt = len(lines)
			if insertAt > 0 && lines[insertAt-1] == "" {
				insertAt--
			}
		}
		if insertAt > 0 && strings.TrimSpace(lines[insertAt-1]) != "" {
			block = append([]string{""}, block...)
		}
		if insertAt < len(lines) && strings.TrimSpace(lines[insertAt]) != "" {
			block = append(block, "")
		}
		lines = append(lines[:insertAt], append(block, lines[insertAt:]...)...)
	}

	tmp, err := os.CreateTemp(filepath.Dir(configPath), filepath.Base(configPath)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create config temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}()
	if err := tmp.Chmod(info.Mode().Perm()); err != nil {
		return fmt.Errorf("preserve config permissions: %w", err)
	}
	if _, err := tmp.WriteString(strings.Join(lines, "\n")); err != nil {
		return fmt.Errorf("write config temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("sync config temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close config temp file: %w", err)
	}
	if err := os.Rename(tmpPath, configPath); err != nil {
		return fmt.Errorf("replace config file: %w", err)
	}
	return nil
}

func assignmentKey(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return "", false
	}
	equals := strings.IndexByte(trimmed, '=')
	if equals < 1 {
		return "", false
	}
	key := strings.TrimSpace(trimmed[:equals])
	for _, r := range key {
		if r != '_' && r != '-' && (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
			return "", false
		}
	}
	return key, true
}

func inlineAssignmentComment(line string) string {
	equals := strings.IndexByte(line, '=')
	if equals < 0 {
		return ""
	}
	inBasic, inLiteral, escaped := false, false, false
	for i := equals + 1; i < len(line); i++ {
		char := line[i]
		if escaped {
			escaped = false
			continue
		}
		if inBasic && char == '\\' {
			escaped = true
			continue
		}
		switch char {
		case '"':
			if !inLiteral {
				inBasic = !inBasic
			}
		case '\'':
			if !inBasic {
				inLiteral = !inLiteral
			}
		case '#':
			if !inBasic && !inLiteral {
				return strings.TrimSpace(line[i:])
			}
		}
	}
	return ""
}
