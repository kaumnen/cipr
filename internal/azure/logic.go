package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/kaumnen/cipr/internal/utils"
	"github.com/spf13/viper"
)

type Prefix struct {
	Address string
	Region  string
	Service string
}

type Config struct {
	Source    string
	IPType    string
	Filter    string
	List      string
	Verbosity string
}

type rawData struct {
	Cloud        string  `json:"cloud"`
	ChangeNumber int     `json:"changeNumber"`
	Values       []value `json:"values"`
}

type value struct {
	Name       string     `json:"name"`
	ID         string     `json:"id"`
	Properties properties `json:"properties"`
}

type properties struct {
	Region          string   `json:"region"`
	Platform        string   `json:"platform"`
	SystemService   string   `json:"systemService"`
	AddressPrefixes []string `json:"addressPrefixes"`
}

const browserUA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 " +
	"(KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

var jsonURLRegex = regexp.MustCompile(`https://download\.microsoft\.com/[^"' ]+ServiceTags_Public_\d+\.json`)

// antiBotStubThreshold is the body size below which Microsoft's
// download page is almost certainly the lightweight anti-bot stub
// rather than the full ~120 KB rendered page.
const antiBotStubThreshold = 10 * 1024

const recoveryHint = "Workarounds: download the JSON from the page in a browser " +
	"and pass `--source <path>`, or set `azure_local_file` in cipr.toml. " +
	"You can also pass a direct `--source <json-url>` if you know one."

func scrapeMissError(pageURL string, body []byte) error {
	if len(body) < antiBotStubThreshold {
		return fmt.Errorf("received anti-bot stub from %s (%d bytes); "+
			"the User-Agent likely needs updating. %s",
			pageURL, len(body), recoveryHint)
	}
	return fmt.Errorf("could not locate ServiceTags JSON URL on %s; "+
		"page layout may have changed. %s", pageURL, recoveryHint)
}

func GetIPRanges(ctx context.Context, config Config) error {
	raw, err := fetchRawData(ctx, config.Source)
	if err != nil {
		return err
	}

	filterSlice := separateFilters(config.Filter)
	prefixes, err := filtrateIPRanges(raw, config.IPType, filterSlice)
	if err != nil {
		return err
	}
	if config.List != "" {
		return printListedValues(prefixes, config.List)
	}
	printIPRanges(prefixes, config.Verbosity)
	return nil
}

func printListedValues(prefixes []Prefix, dim string) error {
	var get func(Prefix) string
	switch dim {
	case "regions":
		get = func(p Prefix) string { return p.Region }
	case "services":
		get = func(p Prefix) string { return p.Service }
	default:
		return fmt.Errorf("unknown list dimension %q (valid: regions, services)", dim)
	}

	values := make([]string, 0, len(prefixes))
	for _, p := range prefixes {
		values = append(values, get(p))
	}
	values = utils.DedupeSorted(values)
	if len(values) == 0 {
		fmt.Println("No values to display.")
		return nil
	}
	for _, v := range values {
		fmt.Println(v)
	}
	return nil
}

// fetchRawData mirrors utils.GetRawData's source dispatch but adds a
// browser-UA scrape step when the resolved endpoint is the Microsoft
// details.aspx page (Microsoft serves an anti-bot stub to non-browser UAs).
// Hosted runs go through utils.GetCached so the resolved JSON gets cached
// under the "azure" key — without that wrapper the scrape+download URL
// path would refetch on every invocation (raw URLs aren't cached).
func fetchRawData(ctx context.Context, source string) (string, error) {
	var endpointURL, cacheKey string
	switch {
	case strings.HasPrefix(source, "https://") || strings.HasPrefix(source, "http://"):
		endpointURL = source
	case strings.Contains(source, "/"):
		return utils.GetRawData(ctx, source)
	default:
		if lf := viper.GetString(source + "_local_file"); lf != "" {
			return utils.GetRawData(ctx, source)
		}
		endpointURL = viper.GetString(source + "_endpoint")
		if endpointURL == "" {
			endpointURL = utils.DefaultEndpoints[source]
		}
		cacheKey = source
	}

	if endpointURL == "" {
		return utils.GetRawData(ctx, source)
	}

	return utils.GetCached(ctx, cacheKey, func(ctx context.Context) (string, error) {
		if strings.HasSuffix(strings.ToLower(endpointURL), ".json") {
			return utils.GetRawData(ctx, endpointURL)
		}
		jsonURL, err := scrapeJSONURL(ctx, endpointURL)
		if err != nil {
			return "", err
		}
		return utils.GetRawData(ctx, jsonURL)
	})
}

func scrapeJSONURL(ctx context.Context, pageURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", fmt.Errorf("build request for %s: %w", pageURL, err)
	}
	req.Header.Set("User-Agent", browserUA)

	fmt.Fprintln(os.Stderr, "Resolving Azure ServiceTags JSON URL from:", pageURL)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch %s: %w", pageURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("unexpected status %d from %s", resp.StatusCode, pageURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body from %s: %w", pageURL, err)
	}

	match := jsonURLRegex.FindString(string(body))
	if match == "" {
		return "", scrapeMissError(pageURL, body)
	}
	return match, nil
}

func separateFilters(filterFlagValues string) []string {
	stripped := strings.ReplaceAll(filterFlagValues, " ", "")
	parts := strings.Split(stripped, ",")

	var filterSlice []string
	for _, p := range parts {
		if p == "" {
			filterSlice = append(filterSlice, "*")
		} else {
			filterSlice = append(filterSlice, p)
		}
	}
	for len(filterSlice) < 2 {
		filterSlice = append(filterSlice, "*")
	}
	if len(filterSlice) > 2 {
		filterSlice = filterSlice[:2]
	}
	return filterSlice
}

func filtrateIPRanges(raw, ipType string, filterSlice []string) ([]Prefix, error) {
	var data rawData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil, fmt.Errorf("parse azure service-tags json: %w", err)
	}

	var result []Prefix
	for _, v := range data.Values {
		if !matchesFilter(v.Properties, filterSlice) {
			continue
		}
		for _, addr := range v.Properties.AddressPrefixes {
			if !ipVersionMatches(addr, ipType) {
				continue
			}
			result = append(result, Prefix{
				Address: addr,
				Region:  v.Properties.Region,
				Service: v.Properties.SystemService,
			})
		}
	}
	return result, nil
}

func matchesFilter(p properties, filterSlice []string) bool {
	return (filterSlice[0] == "*" || strings.EqualFold(p.Region, filterSlice[0])) &&
		(filterSlice[1] == "*" || strings.EqualFold(p.SystemService, filterSlice[1]))
}

func ipVersionMatches(addr, ipType string) bool {
	switch ipType {
	case "ipv4":
		return utils.IsIPv4(addr)
	case "ipv6":
		return utils.IsIPv6(addr)
	}
	return true
}

func printIPRanges(prefixes []Prefix, verbosity string) {
	if len(prefixes) == 0 {
		fmt.Println("No IP ranges to display.")
		return
	}

	var printFunc func(Prefix)
	switch verbosity {
	case "mini":
		printFunc = func(p Prefix) {
			fmt.Printf("%s,%s,%s\n", p.Address, p.Region, p.Service)
		}
	case "full":
		printFunc = func(p Prefix) {
			fmt.Printf("IP Prefix: %s, Region: %s, Service: %s\n", p.Address, p.Region, p.Service)
		}
	default:
		printFunc = func(p Prefix) { fmt.Println(p.Address) }
	}

	for _, p := range prefixes {
		printFunc(p)
	}
}
