package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"promproxy/util"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

var envLabel *dto.LabelPair
var accessToken string

func main() {
	log.Print("Starting promproxy")

	if env, ok := os.LookupEnv("PROMPROXY_ENV_LABEL"); ok {
		envLabel = util.CreateLabelPair("env", env)
		log.Print("Using environment label " + env)
	}

	if token, ok := os.LookupEnv("PROMPROXY_ACCESS_TOKEN"); ok {
		accessToken = "Bearer " + token
		log.Print("Expecting access token " + token)
	}

	http.HandleFunc("/", reqHandler)
	log.Fatal(http.ListenAndServe(":9999", nil))
}

func reqHandler(w http.ResponseWriter, inReq *http.Request) {
	var err error

	request, err := parseRequest(inReq.Context(), inReq.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Print(inReq.URL)

	if accessToken != "" && inReq.Header["Authorization"][0] != accessToken {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	results, err := request.resolver.Resolve(inReq.Context(), request.host)
	if err != nil || len(results) == 0 {
		http.NotFound(w, inReq)
		return
	}

	for _, result := range results {
		url := fmt.Sprintf("http://%s:%d/%s", result.IP, request.port, request.path)
		outReq, _ := http.NewRequest(http.MethodGet, url, nil)

		// Basic auth
		if request.basicAuth != nil {
			outReq.SetBasicAuth(request.basicAuth.username, request.basicAuth.password)
		}

		// Set headers
		for key, values := range request.headers {
			for _, value := range values {
				outReq.Header.Add(key, value)
			}
		}

		outRes, err := http.DefaultClient.Do(outReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if outRes.StatusCode != http.StatusOK {
			http.Error(w, "Upstream error", http.StatusInternalServerError)
			return
		}

		defer outRes.Body.Close()

		var sample dto.MetricFamily
		format := expfmt.ResponseFormat(outRes.Header)
		decoder := expfmt.NewDecoder(outRes.Body, format)
		encoder := expfmt.NewEncoder(w, format)

		for {
			if decoder.Decode(&sample) != nil {
				break
			}

			for _, metric := range sample.Metric {
				metric.Label = append(metric.Label, result.Label)

				if envLabel != nil {
					metric.Label = append(metric.Label, envLabel)
				}
			}

			encoder.Encode(&sample)
		}
	}
}
