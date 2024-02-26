package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/config"
)

func main() {
	token := os.Getenv("token")
	tlsCert := os.Getenv("tls.crt")
	tlsKey := os.Getenv("tls.key")

	if tlsCert == "" || tlsKey == "" {
		fmt.Println("TLS_CERT_PATH or TLS_KEY_PATH environment variable not set")
		return
	}

	cert, err := tls.X509KeyPair([]byte(tlsCert), []byte(tlsKey))
	if err != nil {
		fmt.Println("Failed to load TLS certificate and key:", err)
		return
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	test(context.Background(), tr)

	{
		fmt.Println("Using CA certificate from environment variable:")
		caCert := os.Getenv("ca.crt")
		caCertPool := x509.NewCertPool()
		if ok := caCertPool.AppendCertsFromPEM([]byte(caCert)); !ok {
			fmt.Println("Failed to append CA certificate to pool")
			return
		}
		rt, err := config.NewRoundTripperFromConfig(config.HTTPClientConfig{
			TLSConfig: config.TLSConfig{
				CA:  caCert,
				Key: config.Secret(tlsKey),
				// InsecureSkipVerify: true,
			},
			BearerToken: config.Secret(token),
		}, "test")

		if err != nil {
			fmt.Printf("Error creating round tripper: %v\n", err)
		}
		client, err := api.NewClient(api.Config{
			Address:      "https://thanos-querier.openshift-monitoring.svc.cluster.local:9091",
			RoundTripper: rt,
		})
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		v1api := v1.NewAPI(client)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		r := v1.Range{
			Start: time.Now().Add(-time.Hour),
			End:   time.Now(),
			Step:  time.Minute,
		}
		result, warnings, err := v1api.QueryRange(ctx, "rate(prometheus_tsdb_head_samples_appended_total[5m])", r)
		if err != nil {
			fmt.Printf("Error querying Prometheus w/ CA Cert: %v\n", err)

		}
		if len(warnings) > 0 {
			fmt.Printf("Warnings: %v\n", warnings)
		}
		fmt.Printf("Result:\n%v\n", result)
	}

	{
		rt, err := config.NewRoundTripperFromConfig(config.HTTPClientConfig{
			TLSConfig: config.TLSConfig{
				Cert:               tlsCert,
				Key:                config.Secret(tlsKey),
				InsecureSkipVerify: true,
			},
			// BearerToken: config.Secret(token),
		}, "test")

		if err != nil {
			fmt.Printf("Error creating round tripper: %v\n", err)
		}
		client, err := api.NewClient(api.Config{
			Address:      "https://thanos-querier.openshift-monitoring.svc.cluster.local:9091",
			RoundTripper: rt,
		})
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		v1api := v1.NewAPI(client)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		r := v1.Range{
			Start: time.Now().Add(-time.Hour),
			End:   time.Now(),
			Step:  time.Minute,
		}
		result, warnings, err := v1api.QueryRange(ctx, "rate(prometheus_tsdb_head_samples_appended_total[5m])", r)
		if err != nil {
			fmt.Printf("Error querying Prometheus w/o RT: %v\n", err)
			os.Exit(1)
		}
		if len(warnings) > 0 {
			fmt.Printf("Warnings: %v\n", warnings)
		}
		fmt.Printf("Result:\n%v\n", result)
	}

	client, err := api.NewClient(api.Config{
		Address:      "https://thanos-querier.openshift-monitoring.svc.cluster.local:9091",
		RoundTripper: config.NewAuthorizationCredentialsRoundTripper("Bearer", config.Secret(token), tr),
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}

	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	r := v1.Range{
		Start: time.Now().Add(-time.Hour),
		End:   time.Now(),
		Step:  time.Minute,
	}
	result, warnings, err := v1api.QueryRange(ctx, "rate(prometheus_tsdb_head_samples_appended_total[5m])", r)
	if err != nil {
		fmt.Printf("Error querying Prometheus: %v\n", err)
		os.Exit(1)
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}
	fmt.Printf("Result:\n%v\n", result)
}

func test(ctx context.Context, tr *http.Transport) {
	client, err := api.NewClient(api.Config{
		Address:      "https://thanos-querier.openshift-monitoring.svc.cluster.local:9091",
		RoundTripper: tr,
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		return
	}

	r := v1.Range{
		Start: time.Now().Add(-time.Hour),
		End:   time.Now(),
		Step:  time.Minute,
	}
	v1api := v1.NewAPI(client)
	result, warnings, err := v1api.QueryRange(ctx, "rate(prometheus_tsdb_head_samples_appended_total[5m])", r)
	if err != nil {
		fmt.Printf("Error querying Prometheus w/o bearer: %v\n", err)
		return
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}
	fmt.Printf("Result:\n%v\n", result)
}
