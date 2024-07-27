## cipr: A CLI tool for retrieving IP ranges

This tool retrieves IP ranges from various providers, currently supporting AWS and Cloudflare.

> [!WARNING]
> This tool is still WIP.

## Installation

```bash
curl https://raw.githubusercontent.com/kaumnen/cipr/main/install.sh | bash
```

## Uninstallation

```bash
curl https://raw.githubusercontent.com/kaumnen/cipr/main/uninstall.sh | bash
```

## Usage

```
cipr [command]
```

### Available Commands

- **aws:** Get AWS IP ranges.
- **cloudflare:** Get Cloudflare IP ranges.

## AWS Command

```
cipr aws [flags]
```

### Flags

- **-\-ipv4:** Get only IPv4 ranges.
- **-\-ipv6:** Get only IPv6 ranges.
- **-\-filter:** Filter results using a single string. Syntax: `aws-region-az,SERVICE,network-border-group`.
- **-\-filter-region:** Filter results by AWS region.
- **-\-filter-service:** Filter results by AWS service.
- **-\-filter-network-border-group:** Filter results by AWS network border group.
- **-\-verbose:** Set verbosity level. Options: `none`, `mini`, `full`. Default is `none`.

### Examples

- Get all AWS IPv4 ranges:

```
cipr aws --ipv4
```

- Output:

```
...
35.71.111.0/24
52.94.13.0/24
52.94.7.0/24
...
```

---

- Get AWS IPv6 ranges for the `eu-west-1` region:

```
cipr aws --ipv6 --filter-region eu-west-1
```

- Output:

```
...
2600:f00e::/39
2a05:d030:8000::/40
2a05:d018::/35
2a05:d000:8000::/40
2a05:d070:8000::/40
...
```

---

- Get all AWS IP ranges for the EC2 service in the `us-east-1` with full verbosity:

```
cipr aws --filter us-east-1,ec2 --verbose full
```

- Output:

```
...
IP Prefix: 3.3.2.0/24, Region: us-east-1, Service: EC2, Network Border Group: us-east-1
IP Prefix: 96.0.48.0/21, Region: us-east-1, Service: EC2, Network Border Group: us-east-1-scl-1
IP Prefix: 2600:f0f0:2::/48, Region: us-east-1, Service: EC2, Network Border Group: us-east-1
IP Prefix: 2605:9cc0:1ff0:500::/56, Region: us-east-1, Service: EC2, Network Border Group: us-east-1
IP Prefix: 2600:1f19:8000::/36, Region: us-east-1, Service: EC2, Network Border Group: us-east-1
...
```

## Cloudflare Command

```
cipr cloudflare [flags]
```

### Flags

- **-\-ipv4:** Get only IPv4 ranges.
- **-\-ipv6:** Get only IPv6 ranges.

### Examples

- Get all Cloudflare IPv4 ranges:

```
cipr cloudflare --ipv4
```

- Output:

```
...
188.114.96.0/20
197.234.240.0/22
198.41.128.0/17
...
```

---

- Get all Cloudflare IPv6 ranges:

```
cipr cloudflare --ipv6
```

- Output:

```
...
2803:f800::/32
2405:b500::/32
2405:8100::/32
...
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.
