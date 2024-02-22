package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/config"
)

func main() {
	token := "eyJhbGciOiJSUzI1NiIsImtpZCI6IlNrM1ZJUmI2Qkh2UWhZS1FzaUtJNlVldVRkNEhndjU4aFc1QVc3d3pIR3cifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJvcGVuc2hpZnQtdXNlci13b3JrbG9hZC1tb25pdG9yaW5nIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6InByb21ldGhldXMtdXNlci13b3JrbG9hZC10b2tlbi1jYnh6ZCIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50Lm5hbWUiOiJwcm9tZXRoZXVzLXVzZXItd29ya2xvYWQiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC51aWQiOiIxNjY0MmRhNy1hZDFhLTQxMzUtOWZlOC1lNTIyNDU2MzRmYjgiLCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6b3BlbnNoaWZ0LXVzZXItd29ya2xvYWQtbW9uaXRvcmluZzpwcm9tZXRoZXVzLXVzZXItd29ya2xvYWQifQ.ADVG81kOzuzQ3Y7SG1Ji4uC2iIMjEok0okIQOf-znFW16g_xSCqJjwLOKlOgQHQ1THopWeNZaynF2jXvc5LeSxbSMC8H-o-FfhHd18SlKSBHe0n9dGv4CePj6hArqNmj28FEu3_SbRQ_wWC4aMF2TgoT5qN9dTWnOrUNqMkFR84j6dJgnaHDvedzGNt4vyj9rUOlB9KyEsQNtB8hiEKPV36_pIyiy7eNELi8RtiskFopZrJHGfF6Z8-Bg1w2gO9OBErz4K1be-p-7k5ES6WbPdlC0bJbkQvgh6yTD3T_M8elC610M_j1NOPEOe8heGKFrInv4ZUhwxGFm6jDyHnLyLRsJV6zRHupt8iXq7fc6xqt-u_B5RtBHdA1X94EPGuE97RGaTqnnyiNyFSvKK8p5nFrYLKRtkboVn7IrnuL9RzgOv_QP5IpvQxOEFikdPo5RP9qY1mDcXwGW7k0G0lN4eHkJcwL1RnFER6dGR6hfCI7y9yH5dUmiZTwHWylTWwQhMBzRmIn-S0nybzhLF2xeyDfZ1UDxt54CXzu8FEhDDLQmasCb3xsiLJs0168qg1kO6UAv0LfLkRH_eMahRvD695i7zHMQHdR_Edp1YP1hhUnwIEHxSzuaLFbs-MPJw-wi-LN730ElssvIXRG25eYiDv1DheJQCWtgy95XNgD-FY"
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
