package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	appointmentroom "example.com/appointment-room-service"
)

type matchRequest struct {
	Appointment appointmentroom.AppointmentRequest `json:"appointment"`
	Clinicians  []appointmentroom.Clinician        `json:"clinicians"`
}

func main() {
	apiKey := os.Getenv("INFRAI_API_KEY")
	if apiKey == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}

	matcher := appointmentroom.Matcher{Gateway: appointmentroom.NewInfraiClient(apiKey)}
	handler := http.NewServeMux()
	handler.HandleFunc("POST /appointments/match", func(w http.ResponseWriter, r *http.Request) {
		var input matchRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON request"})
			return
		}
		result, err := matcher.Match(r.Context(), input.Appointment, input.Clinicians)
		if err != nil {
			status := http.StatusBadGateway
			var apiErr *appointmentroom.InfraiError
			if errors.As(err, &apiErr) && apiErr.HTTPStatus >= 400 && apiErr.HTTPStatus < 500 {
				status = apiErr.HTTPStatus
			}
			writeJSON(w, status, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	log.Println("appointment room service listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
