[![CI](https://github.com/gojek/nsxt_exporter/workflows/Master%20Deployment/badge.svg)][ci]
[![Go Report Card](https://goreportcard.com/badge/github.com/gojek/nsxt_exporter)][goreportcard]
[![Docker Pulls](https://img.shields.io/docker/pulls/cloudnativeid/nsxt-exporter-linux-amd64.svg?maxAge=604800)][hub]
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)][license]

[ci]: https://github.com/gojek/nsxt_exporter/actions?query=workflow%3A%22Master+Deployment%22+branch%3Amaster
[goreportcard]: https://goreportcard.com/report/github.com/gojek/nsxt_exporter
[hub]: https://hub.docker.com/r/cloudnativeid/nsxt-exporter-linux-amd64/
[license]: https://opensource.org/licenses/Apache-2.0

# NSX-T Exporter for Prometheus
Simple server that scrapes NSX-T stats and exports them via HTTP for Prometheus consumption.

## Getting Started

To run it:

```bash
./nsxt_exporter [flags]
```

Help on flags:

```bash
./nsxt_exporter --help
```

For more information check the [source code documentation][gdocs].

[gdocs]: http://godoc.org/github.com/gojek/nsxt_exporter

## Usage

Specify host URI for the NSX API using the `--nsxt.host` flag. 
Add the credentials as well by using `--nsxt.username` and `--nsxt.password` flags:
```bash
./nsxt_exporter --nsxt.host localhost --nsxt.username user --nsxt.password password
```

Certificate validation is disabled by default, but
you can enable it using the `--nsxt.insecure=false` flag:
```bash
./nsxt_exporter --nsxt.host localhost --nsxt.username user --nsxt.password password --nsxt.insecure=false
```

### Docker

To run the nsx-t exporter as a Docker container, run:

```bash
docker run -p 9744:9744 cloudnativeid/nsxt-exporter-linux-amd64:latest --nsxt.host localhost --nsxt.username user --nsxt.password password
```

### Building

```bash
make build
```

### Testing

```bash
make test
```

## License

Apache License 2.0, see [LICENSE](https://github.com/gojek/nsxt_exporter/blob/master/LICENSE).
