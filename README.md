# cipr - Cloud IP Range Retrieval Tool

![GitHub release (latest by date)](https://img.shields.io/github/v/release/kaumnen/cipr)
![GitHub License](https://img.shields.io/github/license/kaumnen/cipr)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/kaumnen/cipr/releaser.yml)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/kaumnen/cipr)

cipr is a command-line interface (CLI) tool designed to simplify the process of retrieving IP ranges from various cloud providers and services. It provides a quick and efficient way to access up-to-date IP ranges, which can be particularly useful for network administrators, security professionals, and developers working with cloud infrastructure.

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
```

`--proxy` supplies an HTTP(S) proxy for all network requests. When it is not
set, cipr uses the standard `HTTP_PROXY`, `HTTPS_PROXY`, and `NO_PROXY`
environment variables. `--debug` writes source, cache, proxy, and HTTP
diagnostics to stderr without changing IP-range output.

## Contributing

Pull requests are welcome. For major changes, please open an issue first
to discuss what you would like to change.

Please make sure to update tests as appropriate.

More info available here - [CONTRIBUTING.md](https://github.com/kaumnen/cipr/blob/main/CONTRIBUTING.md)

## License

[MIT](https://choosealicense.com/licenses/mit/)
