package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/ricardolindner/go-expert-otel/goweather-api/internal/services"
)

func mockViaCEPServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/ws/")
		cep := strings.TrimSuffix(path, "/json/")

		switch cep {
		case "89053300":
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"localidade": "Blumenau"}`)
		case "00000000":
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"erro": "true"}`)
		case "89053301":
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"localidade": "UnknownCity"}`)
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"erro": "true"}`)
		}
	}))
}

func mockWeatherAPIServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if query == "Blumenau" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"current": {"temp_c": 17.1}}`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"error": {"code": 1006, "message": "No matching location found."}}`)
	}))
}

func TestGetWeather(t *testing.T) {
	viaCepServer := mockViaCEPServer(t)
	defer viaCepServer.Close()
	weatherAPIServer := mockWeatherAPIServer(t)
	defer weatherAPIServer.Close()

	services.ViaCEPClient = viaCepServer.Client()
	services.ViaCEPBaseURL = viaCepServer.URL
	services.WeatherAPIClient = weatherAPIServer.Client()
	services.WeatherAPIBaseURL = weatherAPIServer.URL

	tests := []struct {
		name       string
		cep        string
		wantStatus int
		wantBody   string
	}{
		{"valid CEP", "89053300", http.StatusOK, `{"temp_C":17.1,"temp_F":62.78,"temp_K":290.1}`},
		{"invalid format CEP", "1234567", http.StatusUnprocessableEntity, `{"error":"invalid zipcode"}`},
		{"non-existent CEP", "00000000", http.StatusNotFound, `{"error":"can not find zipcode"}`},
		{"no weather for location", "89053301", http.StatusNotFound, `{"error":"can not find weather for this location"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", fmt.Sprintf("/weather?cep=%s", tt.cep), nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(GetWeather)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v, want %v", status, tt.wantStatus)
			}

			var got, want map[string]interface{}
			if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
				t.Fatalf("error unmarshaling got JSON: %v", err)
			}

			if err := json.Unmarshal([]byte(tt.wantBody), &want); err != nil {
				t.Fatalf("error unmarshaling want JSON: %v", err)
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("handler returned unexpected JSON: got %v, want %v", got, want)
			}
		})
	}
}
