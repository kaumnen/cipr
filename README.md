# cipr - Cloud IP Range Retrieval Tool

![GitHub release (latest by date)](https://img.shields.io/github/v/release/kaumnen/cipr)
![GitHub License](https://img.shields.io/github/license/kaumnen/cipr)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/kaumnen/cipr/releaser.yml)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/kaumnen/cipr)

cipr is a command-line interface (CLI) tool designed to simplify the process of retrieving IP ranges from AWS, Azure, Cloudflare, DigitalOcean, GitHub, Google Cloud, and iCloud Private Relay. It provides a quick and efficient way to access up-to-date IP ranges, which can be particularly useful for network administrators, security professionals, and developers working with cloud infrastructure.

## Installation

To install cipr, you can use Homebrew:

```bash title='CLI command'
brew install kaumnen/tap/cipr
```

> [!NOTE]  
> Currently available for Linux and MacOS systems.
>
> Documentation available at: [cipr.kaumnen.com](https://cipr.kaumnen.com/docs/intro)

## Configuration and diagnostics

cipr creates `$HOME/.config/cipr/cipr.toml` on first use. Inspect all managed
settings with `cipr configure`, or target one source key:

```bash
cipr configure aws --endpoint https://example.com/aws-ranges.json
cipr configure azure --local-file /path/to/azure-ranges.json --cache-ttl 0s
cipr configure cloudflare_ipv4
cipr configure gcp
```

`--proxy` supplies an HTTP(S) proxy for all network requests. When it is not
set, cipr uses the standard `HTTP_PROXY`, `HTTPS_PROXY`, and `NO_PROXY`
environment variables. `--debug` writes source, cache, proxy, and HTTP
diagnostics to stderr without changing IP-range output.

## Google Cloud

The `gcp` provider reads Google's [`cloud.json`](https://www.gstatic.com/ipranges/cloud.json)
feed of [global and regional external IP ranges](https://cloud.google.com/vpc/docs/access-apis-external-ip)
available to Google Cloud customers. These are not project-specific allocations
or the broader ranges used by Google APIs and services.

```bash
cipr gcp --ipv4 --filter-scope us-central1
cipr gcp --ipv6 --filter 'global,Google Cloud' --verbose-mode mini
cipr gcp --list scopes
```

Scopes are Google Cloud region names or `global`. The feed's `service` field is
currently `Google Cloud`; it does not identify individual products or APIs.
Google updates the feed frequently. Use `--no-cache` when a fresh fetch is
required instead of the default 24-hour cached copy.

## Contributing

Pull requests are welcome. For major changes, please open an issue first
to discuss what you would like to change.

Please make sure to update tests as appropriate.

More info available here - [CONTRIBUTING.md](https://github.com/kaumnen/cipr/blob/main/CONTRIBUTING.md)

## License

[MIT](https://choosealicense.com/licenses/mit/)
