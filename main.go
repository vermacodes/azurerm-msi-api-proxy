package main

import (
	"io"
	"log"
	"maps"
	"net/http"
	"net/url"
	"os"
)

func main() {
	// Health check endpoint
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	log.Println("[INFO] Starting Azure Identity Proxy...")

	endpoint := os.Getenv("IDENTITY_ENDPOINT")
	if endpoint == "" {
		log.Fatal("[FATAL] IDENTITY_ENDPOINT environment variable is required")
	}
	log.Printf("[INFO] Using IDENTITY_ENDPOINT: %s", endpoint)

	identityHeader := os.Getenv("IDENTITY_HEADER")
	if identityHeader == "" {
		log.Fatal("[FATAL] IDENTITY_HEADER environment variable is required")
	}
	log.Printf("[INFO] IDENTITY_HEADER is set (value hidden for security)")

	armApiVersion := os.Getenv("ARM_MSI_API_VERSION")
	if armApiVersion == "" {
		armApiVersion = "2019-08-01"
		log.Printf("[WARN] ARM_MSI_API_VERSION not set, defaulting to %s", armApiVersion)
	}
	log.Printf("[INFO] Using ARM_MSI_API_VERSION: %s", armApiVersion)

	proxyPort := os.Getenv("ARM_MSI_API_PROXY_PORT")
	if proxyPort == "" {
		proxyPort = "42300"
		log.Printf("[WARN] ARM_MSI_API_PROXY_PORT not set, defaulting to %s", proxyPort)
	}
	log.Printf("[INFO] Proxy will listen on port: %s", proxyPort)

	target, err := url.Parse(endpoint)
	if err != nil {
		log.Fatalf("[FATAL] Failed to parse IDENTITY_ENDPOINT '%s': %v", endpoint, err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[INFO] Incoming request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("[ERROR] Panic recovered: %v", rec)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()

		r.URL.Host = target.Host
		r.URL.Scheme = target.Scheme
		r.Header.Set("x-identity-header", identityHeader)

		// Update api-version parameter if it exists
		query := r.URL.Query()
		if query.Has("api-version") {
			log.Printf("[DEBUG] Overriding api-version to %s", armApiVersion)
			query.Set("api-version", armApiVersion)
			r.URL.RawQuery = query.Encode()
		}

		r.RequestURI = ""

		log.Printf("[INFO] Forwarding %s %s to %s", r.Method, r.URL.Path, r.URL.String())

		resp, err := http.DefaultClient.Do(r)
		if err != nil {
			log.Printf("[ERROR] Failed to forward the request: %v", err)
			http.Error(w, "Failed to forward the request: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer func() {
			cerr := resp.Body.Close()
			if cerr != nil {
				log.Printf("[WARN] Failed to close response body: %v", cerr)
			}
		}()

		maps.Copy(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		bytesWritten, err := io.Copy(w, resp.Body)
		if err != nil {
			log.Printf("[ERROR] Failed to copy response body: %v", err)
		}
		log.Printf("[INFO] Responded with status %d, %d bytes", resp.StatusCode, bytesWritten)
	})

	addr := ":" + proxyPort
	log.Printf("[INFO] Listening and serving HTTP on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("[FATAL] Failed to start HTTP server: %v", err)
	}
}
