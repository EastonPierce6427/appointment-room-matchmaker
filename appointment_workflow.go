package appointmentroom

import (
	"context"
	"fmt"
	"strings"
)

type AppointmentRequest struct {
	AppointmentID    string `json:"appointment_id"`
	PatientID        string `json:"patient_id"`
	Specialty        string `json:"specialty"`
	PatientConsented bool   `json:"patient_consented"`
}

type Clinician struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Specialty   string `json:"specialty"`
	Available   bool   `json:"available"`
}

type MatchResult struct {
	AppointmentID string `json:"appointment_id"`
	State         string `json:"state"`
	Room          string `json:"room,omitempty"`
	ClinicianID   string `json:"clinician_id,omitempty"`
}

type RoomGateway interface {
	PrepareRoom(ctx context.Context, roomName, appointmentID, patientID, clinicianID string) error
}

type Matcher struct {
	Gateway RoomGateway
}

func (m Matcher) Match(ctx context.Context, request AppointmentRequest, clinicians []Clinician) (MatchResult, error) {
	result := MatchResult{AppointmentID: request.AppointmentID, State: "waiting"}
	if request.AppointmentID == "" || request.PatientID == "" || request.Specialty == "" {
		return result, fmt.Errorf("appointment_id, patient_id, and specialty are required")
	}
	if !request.PatientConsented {
		return result, nil
	}

	for _, clinician := range clinicians {
		if clinician.Available && strings.EqualFold(clinician.Specialty, request.Specialty) {
			room := "appointment-" + request.AppointmentID
			if err := m.Gateway.PrepareRoom(ctx, room, request.AppointmentID, request.PatientID, clinician.ID); err != nil {
				return result, err
			}
			result.State = "room_ready"
			result.Room = room
			result.ClinicianID = clinician.ID
			return result, nil
		}
	}

	return result, nil
}
