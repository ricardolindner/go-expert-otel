package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/ricardolindner/go-expert-otel/go-weather-input/internal/util"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var serviceBURL = os.Getenv("SERVICE_B_URL")

type cepInput struct {
	CEP string `json:"cep"`
}

func GetWeather(w http.ResponseWriter, r *http.Request) {
	tr := otel.Tracer("go-weather-input")
	ctx, span := tr.Start(r.Context(), "GetWeather-handler")
	defer span.End()

	var input cepInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		sendJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if !util.IsValidCEP(input.CEP) {
		sendJSONError(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	prop := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	newReq, _ := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s?cep=%s", serviceBURL, input.CEP), nil)
	prop.Inject(ctx, propagation.HeaderCarrier(newReq.Header))

	_, serviceBSpan := tr.Start(ctx, "call-service-B")
	defer serviceBSpan.End()

	resp, err := http.DefaultClient.Do(newReq)
	if err != nil {
		span.RecordError(err)
		sendJSONError(w, fmt.Sprintf("failed to get weather from Service B: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
