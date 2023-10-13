package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/sirupsen/logrus"
)

const (
	cacheExpiry      = 60 * 60 // 1 hour in seconds
	defaultCacheKey  = "proxyCache:"
	chunkSizeBytes   = 900 * 1024 // 900KB
	encryptionEnvVar = "TRANSIT_ENCRYPTION"
)

func setChunks(mc *memcache.Client, baseKey string, data []byte) error {
	numChunks := (len(data) + chunkSizeBytes - 1) / chunkSizeBytes
	for i := 0; i < numChunks; i++ {
		start := i * chunkSizeBytes
		end := start + chunkSizeBytes
		if end > len(data) {
			end = len(data)
		}

		chunkKey := fmt.Sprintf("%s_chunk_%d", baseKey, i)
		err := mc.Set(&memcache.Item{
			Key:        chunkKey,
			Value:      data[start:end],
			Expiration: int32(time.Now().Add(time.Second * time.Duration(cacheExpiry)).Unix()),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func getChunks(mc *memcache.Client, baseKey string) ([]byte, error) {
	var data []byte
	i := 0
	for {
		chunkKey := fmt.Sprintf("%s_chunk_%d", baseKey, i)
		item, err := mc.Get(chunkKey)
		if errors.Is(err, memcache.ErrCacheMiss) {
			break
		} else if err != nil {
			return nil, err
		}
		data = append(data, item.Value...)
		i++
	}

	if len(data) == 0 {
		return nil, memcache.ErrCacheMiss
	}

	return data, nil
}

func main() {
	mc := memcache.New(os.Getenv("MEMCACHED_HOST") + ":" + os.Getenv("MEMCACHED_PORT"))
	if os.Getenv(encryptionEnvVar) == "true" {
		// enable TLS
		mc.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
			tlsDialer := tls.Dialer{
				Config: &tls.Config{InsecureSkipVerify: true},
			}
			return tlsDialer.DialContext(ctx, network, address)
		}
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		targetURL := r.URL.Query().Get("url")
		if targetURL == "" {
			http.Error(w, "URL parameter missing", http.StatusBadRequest)
			return
		}

		// Validate the URL
		if _, err := url.ParseRequestURI(targetURL); err != nil {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		cacheKey := defaultCacheKey + targetURL

		// Try to get from cache
		if value, err := getChunks(mc, cacheKey); err == nil {
			_, err = w.Write(value)
			logrus.WithError(err).Error("failed to write response")
			return
		}

		// If not in cache, forward request
		resp, err := http.Get(targetURL)
		if err != nil {
			logrus.WithError(err).Error("failed to fetch target URL")
			http.Error(w, "Error fetching from target URL", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Error reading response body", http.StatusInternalServerError)
			return
		}

		// Store in cache
		err = setChunks(mc, cacheKey, body)
		if err != nil {
			logrus.WithError(err).Error("failed to write to cache")
			http.Error(w, "Failed to write to cache", http.StatusInternalServerError)
			return
		}

		_, err = w.Write(body)
		if err != nil {
			logrus.WithError(err).Error("failed to write response")
		}
	})

	logrus.WithError(http.ListenAndServe(":8080", nil)).Fatal("failed to boot server")
}
