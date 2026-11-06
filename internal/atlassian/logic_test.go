package atlassian

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fixturePath = filepath.Join("..", "testdata", "atlassian_sample.json")

func loadFixture(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(fixturePath)
	require.NoError(t, err)
	return string(data)
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = oldStdout })

	fn()

	require.NoError(t, w.Close())
	os.Stdout = oldStdout
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	require.NoError(t, r.Close())
	return buf.String()
}

func cidrs(items []Item) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.CIDR)
	}
	return out
}

func TestFiltrateIPRanges(t *testing.T) {
	rawData := loadFixture(t)
	tests := []struct {
		name    string
		ipType  string
		filters Filters
		want    []string
	}{
		{
			name:   "both families preserve source order",
			ipType: "both",
			want: []string{
				"52.82.172.0/22", "104.192.136.0/21", "131.162.90.20/30", "2401:1d80:3200::/64",
				"44.220.40.160/28", "104.192.138.0/24", "2401:1d80:32fd::/48",
			},
		},
		{
			name:   "IPv4",
			ipType: "ipv4",
			want:   []string{"52.82.172.0/22", "104.192.136.0/21", "131.162.90.20/30", "44.220.40.160/28", "104.192.138.0/24"},
		},
		{
			name:   "IPv6",
			ipType: "ipv6",
			want:   []string{"2401:1d80:3200::/64", "2401:1d80:32fd::/48"},
		},
		{
			name:    "product matches any array value case-insensitively",
			ipType:  "both",
			filters: Filters{Product: []string{"LOOM"}},
			want:    []string{"104.192.138.0/24", "2401:1d80:32fd::/48"},
		},
		{
			name:    "multiple products use OR matching",
			ipType:  "both",
			filters: Filters{Product: []string{"email", "rovo-crawler"}},
			want:    []string{"52.82.172.0/22", "131.162.90.20/30"},
		},
		{
			name:    "ingress includes items with both directions",
			ipType:  "ipv4",
			filters: Filters{Direction: []string{"ingress"}},
			want:    []string{"104.192.136.0/21", "104.192.138.0/24"},
		},
		{
			name:    "region",
			ipType:  "both",
			filters: Filters{Region: []string{"us-east-1"}},
			want:    []string{"104.192.136.0/21", "44.220.40.160/28"},
		},
		{
			name:    "perimeter",
			ipType:  "both",
			filters: Filters{Perimeter: []string{"FedRAMP-Moderate"}},
			want:    []string{"44.220.40.160/28"},
		},
		{
			name:    "separate filters use AND matching",
			ipType:  "both",
			filters: Filters{Product: []string{"jira"}, Direction: []string{"egress"}, Region: []string{"us-east-1"}, Perimeter: []string{"commercial"}},
			want:    []string{"104.192.136.0/21"},
		},
		{
			name:    "no matches",
			ipType:  "both",
			filters: Filters{Region: []string{"moon"}},
			want:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := filtrateIPRanges(rawData, tt.ipType, tt.filters)
			require.NoError(t, err)
			assert.Equal(t, tt.want, cidrs(got))
		})
	}
}

func TestFiltrateIPRangesKeepsItemFields(t *testing.T) {
	got, err := filtrateIPRanges(loadFixture(t), "both", Filters{Product: []string{"email"}})
	require.NoError(t, err)
	assert.Equal(t, []Item{{
		CIDR:      "52.82.172.0/22",
		Region:    []string{"global"},
		Product:   []string{"email"},
		Direction: []string{"egress"},
		Perimeter: "commercial",
	}}, got)
}

func TestFiltrateIPRangesValidation(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "malformed JSON", raw: `{`, want: "parse atlassian ip-ranges json"},
		{name: "empty payload", raw: `{}`, want: "no IP ranges found"},
		{name: "empty items", raw: `{"items":[]}`, want: "no IP ranges found"},
		{name: "missing CIDR", raw: `{"items":[{"product":["jira"]}]}`, want: `item 1: "" is not a valid CIDR`},
		{name: "invalid CIDR", raw: `{"items":[{"cidr":"192.0.2.0/24"},{"cidr":"not-a-cidr"}]}`, want: `item 2: "not-a-cidr"`},
		{name: "string product", raw: `{"items":[{"cidr":"192.0.2.0/24","product":"jira"}]}`, want: "parse atlassian ip-ranges json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := filtrateIPRanges(tt.raw, "both", Filters{})
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.want)
		})
	}
}

func TestFiltrateIPRangesValidatesExcludedEntries(t *testing.T) {
	rawData := `{"items":[
		{"cidr":"192.0.2.0/24","product":["jira"]},
		{"cidr":"invalid","product":["trello"]}
	]}`
	_, err := filtrateIPRanges(rawData, "ipv4", Filters{Product: []string{"jira"}})
	require.Error(t, err)
	assert.ErrorContains(t, err, "validate atlassian item 2")
}

func TestPrintIPRanges(t *testing.T) {
	item := Item{
		CIDR:      "104.192.136.0/21",
		Region:    []string{"global", "us-east-1"},
		Product:   []string{"bitbucket", "jira"},
		Direction: []string{"egress", "ingress"},
		Perimeter: "commercial",
	}
	tests := []struct {
		name      string
		items     []Item
		verbosity string
		want      string
	}{
		{name: "empty", verbosity: "none", want: ""},
		{
			name:      "none prints each CIDR once",
			items:     []Item{item, {CIDR: "2401:1d80:32fd::/48"}, item},
			verbosity: "none",
			want:      "104.192.136.0/21\n2401:1d80:32fd::/48\n",
		},
		{
			name:      "mini joins array fields with semicolons",
			items:     []Item{item},
			verbosity: "mini",
			want:      "104.192.136.0/21,global;us-east-1,bitbucket;jira,egress;ingress,commercial\n",
		},
		{
			name:      "full",
			items:     []Item{item},
			verbosity: "full",
			want:      "IP Prefix: 104.192.136.0/21, Region: global;us-east-1, Product: bitbucket;jira, Direction: egress;ingress, Perimeter: commercial\n",
		},
		{
			name:      "unknown verbosity falls back to CIDR only",
			items:     []Item{item},
			verbosity: "unknown",
			want:      "104.192.136.0/21\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := captureStdout(t, func() { printIPRanges(tt.items, tt.verbosity) })
			assert.Equal(t, tt.want, out)
		})
	}
}

func TestPrintListedValues(t *testing.T) {
	items, err := filtrateIPRanges(loadFixture(t), "both", Filters{})
	require.NoError(t, err)

	tests := []struct {
		dimension string
		want      string
	}{
		{dimension: "directions", want: "egress\ningress\n"},
		{dimension: "perimeters", want: "commercial\nfedramp-moderate\n"},
		{dimension: "regions", want: "ap-northeast-1\nap-southeast-1\nap-southeast-2\nca-central-1\ngcp-us-central1\nglobal\nus-east-1\nus-west-1\nus-west-2\n"},
		{dimension: "products", want: "bitbucket\nconfluence\nemail\nforge\ngithub-for-jira\nhalp\njira\nloom\nopsgenie\nrovo-crawler\nstatuspage\ntrello\n"},
	}
	for _, tt := range tests {
		t.Run(tt.dimension, func(t *testing.T) {
			out := captureStdout(t, func() {
				require.NoError(t, printListedValues(items, tt.dimension))
			})
			assert.Equal(t, tt.want, out)
		})
	}

	t.Run("empty input", func(t *testing.T) {
		out := captureStdout(t, func() {
			require.NoError(t, printListedValues(nil, "products"))
		})
		assert.Empty(t, out, "placeholder belongs on stderr")
	})
	t.Run("unknown dimension", func(t *testing.T) {
		err := printListedValues(items, "services")
		require.Error(t, err)
		assert.ErrorContains(t, err, "valid: products, directions, regions, perimeters")
	})
}

func TestGetIPRanges(t *testing.T) {
	config := Config{
		Source:    fixturePath,
		IPType:    "ipv4",
		Filters:   Filters{Perimeter: []string{"fedramp-moderate"}},
		Verbosity: "mini",
	}
	out := captureStdout(t, func() {
		require.NoError(t, GetIPRanges(context.Background(), config))
	})
	assert.Equal(t, "44.220.40.160/28,us-east-1,confluence;jira;opsgenie,egress,fedramp-moderate\n", out)
}

func TestGetIPRangesListComposesWithFilterAndIgnoresIPType(t *testing.T) {
	config := Config{
		Source:  fixturePath,
		IPType:  "ipv6",
		Filters: Filters{Direction: []string{"ingress"}},
		List:    "regions",
	}
	out := captureStdout(t, func() {
		require.NoError(t, GetIPRanges(context.Background(), config))
	})
	assert.Equal(t, "ap-southeast-1\nap-southeast-2\nca-central-1\nglobal\nus-east-1\nus-west-1\nus-west-2\n", out)
}

func TestGetIPRangesPropagatesSourceError(t *testing.T) {
	err := GetIPRanges(context.Background(), Config{Source: filepath.Join(t.TempDir(), "missing.json")})
	require.Error(t, err)
	assert.ErrorContains(t, err, "read")
}
