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
#   -metrics-addr string
#     	network interface to expose for serving prometheus metrics (default "0.0.0.0:9090")
#   -metrics-path string
#     	http path where prometheus metrics are exported (default "/metrics")
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

## Installing

### From source
The tool is built with `Go 1.15`, but any version down to `Go 1.11` should work just fine to build. To install to go bin path, just run `go install` in this repo and make sure your go bin directory is set in your path.

```zsh
$ go install
```

## Dependencies
This project is not without dependencies, below we detail the intent of our dependencies and provide links to their source.

- We use the [official prometheus go client](https://github.com/prometheus/client_golang) to provide endpoints for prometheus to scrape metrics from the web server. The dependency is also licensed under `Apache 2.0`.

## Primary Maintainers
- [Zee Philip Vieira](https://github.com/zeeraw)

## Changelog
### v0.1.1
- Fix: We were still exposing 9090 in the docker image using ONBUILD.
### v0.1.0
- Do not enable metrics by default within the docker image. This is a breaking change for people relying on the docker image to expose metrics as is.