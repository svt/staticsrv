package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	_ "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ref string // git hash injected on compilation
	tag string // git tag injected on compilation
)

const (
	// DefaultAddr defines the network interface to be used by default.
	// It allows any IP addresses on the port 8080.
	DefaultAddr = "0.0.0.0:8080"

	// DefaultMetricsAddr defines the network interace to be used for scraping metrics by default.
	// It allows any IP addresses on the port 9090.
	DefaultMetricsAddr = "0.0.0.0:9090"

	// DefaultMetricsPath defines the path where metrics should be collected by prometheus.
	DefaultMetricsPath = "/metrics"

	// DefaultReadinessPath is used in conjunction with Kubernetes readiness health checks.
	DefaultReadinessPath = "/readyz"

	// DefaultLivelinessPath is used in conjunction with Kubernetes liveness health checks.
	DefaultLivelinessPath = "/livez"

	// DefaultConfigPath is used as the path for the dynamically served configuration json.
	DefaultConfigPath = "/config.json"

	// DefaultIndexFile is used to determine the name of the index files to be looked up.
	DefaultIndexFile = "index.html"
)

func main() {
	var (
		addr              string       // host interface
		dir               string       // the source directory for the site to host
		configVars        string       // the environment variables to expose to /config.json
		metricsAddr       string       // the network interface for prometheus metrics
		metricsPath       string       // the http path where prometheus metrics are exported
		disableChecks     bool         // used to determine if the web server should disable health checks
		disableConfigVars bool         // used to determine if the web server should provide a /config.json endpoint
		enableFallback    bool         // enable fallback
		enableMetrics     bool         // enable prometheus metrics
		printVersion      bool         // use to print the version of the binary
		bin               = os.Args[0] // name of the entrypoint
		root              = "/"        // url path to host the directory under
	)

	// Setup and parse command line arguments.
	flags := flag.NewFlagSet(bin, flag.ExitOnError)
	flags.SetOutput(os.Stdout)
	flags.StringVar(&addr, "addr", DefaultAddr, "network interface to expose for serving the website")
	flags.BoolVar(&disableChecks, "disable-health-checks", false, "disables the /readyz and /livez endpoints")
	flags.BoolVar(&disableConfigVars, "disable-config-variables", false, "disables the /config.json endpoint")
	flags.BoolVar(&enableFallback, "enable-fallback-to-index", false, fmt.Sprintf("enables serving of fallback file (%s) for any missing file", DefaultIndexFile))
	flags.BoolVar(&enableMetrics, "enable-metrics", false, "enable scraping application metrics")
	flags.BoolVar(&printVersion, "version", false, "print the current version number of staticsrv")
	flags.StringVar(&configVars, "config-variables", "", "comma separated list of environment variables to expose in /config.json")
	flags.StringVar(&metricsAddr, "metrics-addr", DefaultMetricsAddr, "network interface to expose for serving prometheus metrics")
	flags.StringVar(&metricsPath, "metrics-path", DefaultMetricsPath, "http path where prometheus metrics are exported")

	flags.Usage = func() {
		fmt.Fprintf(flags.Output(), "Usage: %s [OPTIONS] [DIR]\nConfiguration Options:\n", bin)
		flags.PrintDefaults()
	}
	if err := flags.Parse(os.Args[1:len(os.Args)]); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if printVersion {
		version := "unversioned build"
		if tag != "" {
			version = tag
		}
		if ref != "" {
			version = fmt.Sprintf("%s (%s)", version, ref)
		}
		fmt.Printf("%s\n", version)
		return
	}

	// Get the current working directory to provide a nice default the dir configuration.
	if flags.NArg() > 0 {
		// Use the remaining argument as the directory host.
		dir = flags.Arg(0)
	} else {
		// Derive the current workdir from the operating system and use as a fallback.
		wd, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		dir = wd
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Fatalf("error: %q cannot be hosted host: directory does not exist\n", dir)
	}

	if err := CheckFile(dir, DefaultIndexFile); err != nil {
		log.Printf("warning: %v: required to display the website\n", err)
	}

	// We use a mux to be able to split traffic between the filesystem and some special case paths.
	mux := http.NewServeMux()

	if !disableChecks {
		// To enable this web service for kubernetes, we provide `livez` and `readyz` endpoints.
		log.Printf("liveliness check available on %q\n", DefaultLivelinessPath)
		mux.HandleFunc(DefaultLivelinessPath, HandleOK)

		log.Printf("readiness check available on %q\n", DefaultReadinessPath)
		mux.HandleFunc(DefaultReadinessPath, HandleOK)

		mux.HandleFunc("/healthz", HandleOK) // Deprecated since kubernetes 1.16
	}

	// Provide a /config.json endpoint unless it's been disabled.
	if !disableConfigVars {
		log.Printf("configuration variables available on %q\n", DefaultConfigPath)
		env := ParseCommaSeparatedVars(configVars)
		mux.HandleFunc(DefaultConfigPath, HandleConfig(env))
	}

	// Host the input directory as a filesystem on the root path.
	if enableFallback {
		log.Printf("requests on missing content will automatically serve %q with status %d\n", DefaultIndexFile, http.StatusOK)
	}
	mux.Handle(root, HandleStaticContent(dir, enableFallback))

	if enableMetrics {
		mux := http.NewServeMux()
		mux.Handle(metricsPath, promhttp.Handler())
		log.Printf("serving prometheus metrics through %s on %q\n", metricsAddr, metricsPath)
		// Spin up the metrics server in a go routine and crash the server if metrics fail.
		go func() {
			log.Fatal(http.ListenAndServe(metricsAddr, mux))
		}()
	}

	log.Printf("serving site from %q through %s on %q\n", dir, addr, path.Join(root))
	log.Fatal(http.ListenAndServe(addr, mux))
}

// HandleStaticContent is the main method for serving content.
func HandleStaticContent(dir string, fallback bool) http.HandlerFunc {
	rootdir := http.Dir(dir)
	fs := http.FileServer(rootdir)

	return func(w http.ResponseWriter, r *http.Request) {
		// If we have fallback is enabled: we serve the index.html file instead
		if fallback {
			p := SanitisePath(r.URL)
			if _, err := rootdir.Open(p); os.IsNotExist(err) {
				f, err := rootdir.Open(DefaultIndexFile)
				if err != nil {
					log.Printf("request error: %v", err)
					return
				}
				buf := bufio.NewReader(f)
				if _, err := buf.WriteTo(w); err != nil {
					log.Printf("request error: %v", err)
					return
				}
				return
			}
		}
		fs.ServeHTTP(w, r)
	}
}

// HandleOK is used to respond well to the health probes.
func HandleOK(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// HandleConfig returns a handler that will will respond with the provided
// environment keys and corresponding values as json.
func HandleConfig(env map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw, err := json.Marshal(env)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(raw)
		return
	}
}

// SanitisePath will derive a sanitised path from a request url and modify the
// url with the sanitised path.
func SanitisePath(u *url.URL) string {
	p := u.Path
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
		u.Path = p
	}
	p = path.Clean(p)
	return p
}

// ParseCommaSeparatedVars will read a string with comma separated environment
// variable names and turn them into a map with the corresponding key/value pair
// of the current environment.
func ParseCommaSeparatedVars(s string) map[string]string {
	var variables = make(map[string]string)
	for _, key := range strings.Split(s, ",") {
		if key = strings.TrimSpace(key); key != "" {
			variables[key] = os.Getenv(key)
		}
	}
	return variables
}

// CheckFile can be used to peek if there's a file in the given directory.
func CheckFile(dir, filePath string) error {
	full := path.Join(dir, filePath)
	if _, err := os.Stat(full); os.IsNotExist(err) {
		return fmt.Errorf("cannot find %q in %s", filePath, dir)
	}
	return nil
}
