package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var ViaCEPClient = &http.Client{}
var ViaCEPBaseURL = "https://viacep.com.br"

type CEPInfo struct {
	Localidade string `json:"localidade"`
	Erro       string `json:"erro,omitempty"`
}

func GetCEPInfo(ctx context.Context, cep string) (*CEPInfo, error) {
	tr := otel.Tracer("go-weather-api")
	ctx, span := tr.Start(ctx, "get-cep-info")
	defer span.End()

	span.SetAttributes(attribute.String("cep", cep))

	url := fmt.Sprintf("%s/ws/%s/json/", ViaCEPBaseURL, cep)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	resp, err := ViaCEPClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("could not find cep: %s", cep)
		span.RecordError(err)
		return nil, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	fmt.Println("ViaCEP complete response: ", string(bodyBytes))

	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var cepInfo CEPInfo
	if err := json.NewDecoder(resp.Body).Decode(&cepInfo); err != nil {
		span.RecordError(err)
		return nil, err
	}

	if strings.ToLower(cepInfo.Erro) == "true" {
		err = fmt.Errorf("cep not found")
		span.RecordError(err)
		return nil, err
	}

	return &cepInfo, nil
}
