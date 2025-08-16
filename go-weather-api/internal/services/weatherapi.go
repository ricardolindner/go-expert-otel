package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var WeatherAPIClient = &http.Client{}
var WeatherAPIBaseURL = "http://api.weatherapi.com/v1"

type Weather struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

func GetWeather(ctx context.Context, city string) (*Weather, error) {
	tr := otel.Tracer("go-weather-api")
	ctx, span := tr.Start(ctx, "get-weather")
	defer span.End()

	span.SetAttributes(attribute.String("city", city))

	apiKey := os.Getenv("WEATHER_API_KEY")
	encodedCity := url.QueryEscape(city)

	url := fmt.Sprintf("%s/current.json?key=%s&q=%s", WeatherAPIBaseURL, apiKey, encodedCity)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	resp, err := WeatherAPIClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Println("WeatherAPI error response: ", string(bodyBytes))
		err = fmt.Errorf("could not find weather for city: %s", city)
		span.RecordError(err)
		return nil, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	fmt.Println("WeatherAPI complete response: ", string(bodyBytes))

	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var weather Weather
	if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil {
		span.RecordError(err)
		return nil, err
	}

	return &weather, nil
}
