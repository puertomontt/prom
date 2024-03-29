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
	"k8s.io/client-go/transport"
)

const (
	prometheusUrl = "https://thanos-querier.openshift-monitoring.svc.cluster.local:9091"
)

var (
	token         = os.Getenv("token")
	tlsCert       = os.Getenv("tls.crt")
	tlsKey        = os.Getenv("tls.key")
	caCert        = os.Getenv("ca.crt")
	serviceCaCert = os.Getenv("service-ca.crt")
	skipInsecure  = os.Getenv("SKIP_INSECURE_VERIFY")
)

func main() {
	// withTLSCert()
	okNowImSerious()
	real()
	openshift()
	openshiftCA()
	withSvcIP()
	withCACertAndTLSConfig()
	withCACert()
	withTokenAndTLS()
	withServiceCACert()
	time.Sleep(5 * time.Minute)
}

func real() {
	fmt.Println("real")
	transport := &http.Transport{}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM([]byte(serviceCaCert)) {
		fmt.Println("failed to append prometheus ca cert to pool")
		return
	}
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}
	transport = &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	rt := config.NewAuthorizationCredentialsRoundTripper("Bearer", config.Secret(token), transport)
	client, err := api.NewClient(api.Config{
		Address:      prometheusUrl,
		RoundTripper: rt,
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		return
	}
	query(context.Background(), client)
}

func withTLSCert() {
	fmt.Println("withTLSCert")
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
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: skipInsecure != "",
	}

	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	client, err := api.NewClient(api.Config{
		Address:      prometheusUrl,
		RoundTripper: tr,
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		return
	}

	query(context.Background(), client)
}

func withCACertAndTLSConfig() {
	fmt.Println("withCACertAndTLSConfig")
	fmt.Println("Using CA certificate from environment variable and transport tls config:")
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM([]byte(caCert)); !ok {
		fmt.Println("Failed to append CA certificate to pool")
		return
	}
	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		InsecureSkipVerify: skipInsecure != "",
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig}

	client, err := api.NewClient(api.Config{
		Address:      prometheusUrl,
		RoundTripper: config.NewAuthorizationCredentialsRoundTripper("Bearer", config.Secret(token), transport),
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		return
	}

	query(context.Background(), client)
}

func withCACert() {
	fmt.Println("withCACert")
	fmt.Println("Using CA certificate from environment variable:")
	caCert := os.Getenv("ca.crt")
	rt, err := config.NewRoundTripperFromConfig(config.HTTPClientConfig{
		TLSConfig: config.TLSConfig{
			CA:                 caCert,
			InsecureSkipVerify: skipInsecure != "",
		},
		BearerToken: config.Secret(token),
	}, "test")

	if err != nil {
		fmt.Printf("Error creating round tripper: %v\n", err)
	}
	client, err := api.NewClient(api.Config{
		Address:      prometheusUrl,
		RoundTripper: rt,
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		return
	}
	query(context.Background(), client)
}

func withSvcIP() {
	fmt.Println("withSvcIP")
	rt, err := config.NewRoundTripperFromConfig(config.HTTPClientConfig{
		TLSConfig: config.TLSConfig{
			CA:                 serviceCaCert,
			InsecureSkipVerify: skipInsecure != "",
		},
		BearerToken: config.Secret(token),
	}, "test")

	if err != nil {
		fmt.Printf("Error creating round tripper: %v\n", err)
	}
	client, err := api.NewClient(api.Config{
		Address:      "https://192.168.221.167:9091",
		RoundTripper: rt,
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		return
	}
	query(context.Background(), client)
}

func withServiceCACert() {
	fmt.Println("withServiceCACert")
	fmt.Println("Using Service CA certificate from environment variable:")
	rt, err := config.NewRoundTripperFromConfig(config.HTTPClientConfig{
		TLSConfig: config.TLSConfig{
			CA:                 serviceCaCert,
			InsecureSkipVerify: skipInsecure != "",
		},
		BearerToken: config.Secret(token),
	}, "test")

	if err != nil {
		fmt.Printf("Error creating round tripper: %v\n", err)
	}
	client, err := api.NewClient(api.Config{
		Address:      prometheusUrl,
		RoundTripper: rt,
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		return
	}
	query(context.Background(), client)
}

func withTokenAndTLS() {
	fmt.Println("withTokenAndTLS")
	rt, err := config.NewRoundTripperFromConfig(config.HTTPClientConfig{
		TLSConfig: config.TLSConfig{
			Cert:               tlsCert,
			Key:                config.Secret(tlsKey),
			InsecureSkipVerify: skipInsecure != "",
		},
		BearerToken: config.Secret(token),
	}, "test")

	if err != nil {
		fmt.Printf("Error creating round tripper: %v\n", err)
	}
	client, err := api.NewClient(api.Config{
		Address:      prometheusUrl,
		RoundTripper: rt,
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		return
	}

	query(context.Background(), client)
}

func openshiftCA() {
	fmt.Println("openshiftCA")
	content, err := os.ReadFile("/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		fmt.Println("Unable to obtain access token from pod or ENV ", err)
		return
	}

	// Convert []byte to string and print to screen
	token := string(content)
	caCertPath := "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	caCert, err := os.ReadFile(caCertPath)

	if err != nil {
		fmt.Println("Error reading service account CA certificate: ", err)
		return
	}

	caCertPool := x509.NewCertPool()
	if parseOk := caCertPool.AppendCertsFromPEM(caCert); !parseOk {
		fmt.Println("Error parsing service account CA certificate")
		return
	}
	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		InsecureSkipVerify: skipInsecure != "",
	}
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	client, err := api.NewClient(api.Config{
		Address:      prometheusUrl,
		RoundTripper: config.NewAuthorizationCredentialsRoundTripper("Bearer", config.Secret(token), transport),
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		return
	}

	query(context.Background(), client)
}

func openshift() {
	fmt.Println("openshift-proxy")
	content, err := os.ReadFile("/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		fmt.Println("Unable to obtain access token from pod or ENV ", err)
		return
	}

	// Convert []byte to string and print to screen
	token := string(content)
	caCertPath := "/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt"
	caCert, err := os.ReadFile(caCertPath)

	if err != nil {
		fmt.Println("Error reading service account CA certificate: ", err)
		return
	}

	caCertPool := x509.NewCertPool()
	if parseOk := caCertPool.AppendCertsFromPEM(caCert); !parseOk {
		fmt.Println("Error parsing service account CA certificate")
		return
	}
	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		InsecureSkipVerify: skipInsecure != "",
	}
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	client, err := api.NewClient(api.Config{
		Address:      prometheusUrl,
		RoundTripper: config.NewAuthorizationCredentialsRoundTripper("Bearer", config.Secret(token), transport),
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		return
	}

	query(context.Background(), client)
}

func query(ctx context.Context, client api.Client) {
	r := v1.Range{
		Start: time.Now().Add(-time.Hour),
		End:   time.Now(),
		Step:  time.Minute,
	}
	v1api := v1.NewAPI(client)
	result, warnings, err := v1api.QueryRange(ctx, "rate(prometheus_tsdb_head_samples_appended_total[5m])", r)
	if err != nil {
		fmt.Printf("Error querying Prometheus: %v\n", err)
		return
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}
	fmt.Printf("Result:\n%v\n", result)
}

func okNowImSerious() {
	fmt.Println("okNowImSerious")
	client, err := newPrometheusClient()
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		return
	}
	query(context.Background(), client)
}

func newPrometheusClient() (api.Client, error) {
	transportConfig := &transport.Config{}

	transportConfig.TLS.CAFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"

	if skipInsecure != "" {
		transportConfig.TLS.Insecure = true
		transportConfig.TLS.CAData = nil
		transportConfig.TLS.CAFile = ""
	}

	transportConfig.BearerTokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"

	roundTripper, err := transport.New(transportConfig)
	if err != nil {
		return nil, err
	}

	return api.NewClient(api.Config{
		Address:      prometheusUrl,
		RoundTripper: roundTripper,
	})
}
