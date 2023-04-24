# staticsrv
The command line tool for hosting a directory as a website.

## Quick Start: Docker
The easiest way to get started with this tool is to include the docker image in your repository.

```dockerfile
# The "latest" tag contains the last released version of this binary.
FROM sverigestelevision/staticsrv:latest
# The docker work directory is /srv/www everything will by default get hosted by staticsrv.
COPY ./build .
```

**NOTE:** If you got other compilation steps for your website before you can host the static artifacts, we recommend using [multi stage docker builds](https://docs.docker.com/develop/develop-images/multistage-build/) using *staticsrv* in the last stage.

## Usage
To host your current directory simply type `staticsrv`.

```zsh
$ staticsrv
```

### Serve a specific location
In most circumstances you probably want to host a directory you're currently not in. Simply use the path to the directory.

```zsh
$ staticsrv ./build
```

### More usage information
You can always get the tool's help page by providing the `-h` flag.

```zsh
$ staticsrv -h
# Usage: staticsrv [OPTIONS] [DIR]
# Configuration Options:
#   -addr string
#     	network interface to expose for serving the website (default "0.0.0.0:8080")
#   -config-variables string
#     	comma separated list of environment variables to expose in /config.json
#   -disable-config-variables
#     	disables the /config.json endpoint
#   -disable-health-checks
#     	disables the /readyz and /livez endpoints
#   -enable-fallback-to-index
#     	enables serving of fallback file (index.html) for any missing file
#   -enable-metrics
#     	enable scraping application metrics
#   -enable-access-log
#     	enable access log to see all requests to your server to stderr
#   -metrics-addr string
#     	network interface to expose for serving prometheus metrics (default "0.0.0.0:9090")
#   -metrics-path string
#     	http path where prometheus metrics are exported (default "/metrics")
#   -timeout-idle int
#       the maximum amount of time to wait for the next request (default 60)
#   -timeout-read int
#       the maximum duration for reading the entire request (default 5)
#   -timeout-write int
#       the maximum duration before timing out writes of the response (default 5)
#   -version
#     	print the current version number of staticsrv
```

## Features

### Liveness and readiness endpoints for kubernetes
In a world of container clouds, it's important to have probe endpoints available. Try them out using curl.

```zsh
$ curl -k http://localhost:8080/readyz
# ok
$ curl -k http://localhost:8080/livez
# ok
```

The check endpoints are enabled by default, but can be disabled.

```zsh
$ staticsrv -disable-health-checks
```

### Fallback to index file
If you're running a front-end framework like React, it might be a good idea to make sure all routes except existing static assets all route to the index file.

```zsh
$ staticsrv -enable-fallback-to-index
```

### Serving configuration variables
Depending on your environment, you might want to provision configuration variables to your web front-end javascript.
The webserver, by default, serves a json file on the path `/config.json`. You can configure what environment variables should be exposed in the response, passing a comma separated list to `-config-variables`.

```zsh
$ staticsrv -config-variables="ENV_VAR_ONE,ENV_VAR_TWO"
```

If you hit the configuration endpoint, you should receive a response like this, including your specified variables, wether or not they're set.

```json
{
  "ENV_VAR_ONE": "Foo",
  "ENV_VAR_TWO": "Bar"
}
```

To disable webserver from serving the `/config.json` file.

```zsh
$ staticsrv -disable-config-variables
```

### Export prometheus metrics
Application metrics can be exported for a prometheus collector to scrape over http. This might be good if you want insight in how the staticsrv runtime is impacting your cluster for high volume applications. Go runtime metrics are as the time of writing not well documented, but you can read the [source code for the collector library](https://github.com/prometheus/client_golang/blob/bff02dd5619915e70ec529743f5ebb898a1970c4/prometheus/go_collector.go#L64) for pointers.

```zsh
$ staticsrv -enable-metrics
```

By default prometheus metrics can be scraped through http on `/metrics` on port `9090`, but can be configured.

```zsh
$ staticsrv -enable-metrics -metrics-addr :2112 -metrics-path /stats
```

#### Metrics Exposed
Here's a table with all the metrics that are exposed by the server through the metrics endpoint. We also expose some metrics for the Go process using the [Go prometheus library](https://github.com/prometheus/client_golang/blob/master/prometheus/go_collector.go).

|name|type|labels|description|
|----|----|------|-----------|
|`staticsrv_http_requests_total`|Counter|`method`, `status`|Total amount of requests made to the server process|
|`staticsrv_http_requests_duration_seconds`|Histogram||Request time in seconds (with fractions)|
|`staticsrv_http_requests_size_bytes`|Histogram||Request size in bytes|

### Request access log
You might want to enable access logs if you want to perform analysis on requests to your site. When enabling the access log the log output will post to stderr using the [logfmt](https://brandur.org/logfmt) format.

```zsh
$ staticsrv -enable-access-log
method=GET duration=7.48875ms size=624B size_bytes=624 status=200 path="/" time=1634312038387629000
method=GET duration=129.875µs size=406B size_bytes=406 status=200 path="/index.css" time=1634312038433253000
method=GET duration=117.666µs size=367B size_bytes=367 status=200 path="/index.js" time=1634312038434145000
method=GET duration=153.416µs size=624B size_bytes=624 status=200 path="/" time=1634312038600024000
```

## Installing

### From source
The tool is built with `Go 1.17`, but any version down to `Go 1.11` should work just fine to build. To install to go bin path, just run `go install` in this repo and make sure your go bin directory is set in your path.

```zsh
$ go install
```

## Dependencies
This project is not without dependencies, below we detail the intent of our dependencies and provide links to their source.

- We use the [official prometheus go client](https://github.com/prometheus/client_golang) to provide endpoints for prometheus to scrape metrics from the web server. The dependency is also licensed under `Apache 2.0`.

## Release process
1. Create an annotated tag and push
[Github Actions](.github/workflows) publishes a new image on [Docker Hub](https://hub.docker.com/r/sverigestelevision/staticsrv), once a new tag like "v\*.\*.\*" (e.g. v0.3.0) is created.
```
git tag -s -a v${version} -m "v${version}"
git push origin v${version}
```
2. Create a new release on github, according to the [instruction](https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository). Preferably use Auto-generate release notes for the notes so we don't forget any changes.

## Primary Maintainers
- [Zee Philip Vieira](https://github.com/zeeraw)

## Changelog
### v0.3.0
- Fix: Set timeout values for HTTP requsts.
- Chore: update OS for build-container.
### v0.2.1
- Fix: Handle unhandled errors when writing responses.
### v0.2.0
- Feature: Added request metrics to prometheus export
- Feature: Added request access logs that can be enabled with a flag to the server
### v0.1.1
- Fix: We were still exposing 9090 in the docker image using ONBUILD.
### v0.1.0
- Do not enable metrics by default within the docker image. This is a breaking change for people relying on the docker image to expose metrics as is.
