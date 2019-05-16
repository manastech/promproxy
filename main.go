package main

import (
	"fmt"
	"log"
	"net/http"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func main() {
	log.Print("Starting promproxy")
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
			}

			encoder.Encode(&sample)
		}
	}
}
