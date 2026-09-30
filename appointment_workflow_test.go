package appointmentroom

import (
	"context"
	"testing"
)

type recordingGateway struct {
	calls int
	room  string
}

func (g *recordingGateway) PrepareRoom(_ context.Context, roomName, _, _, _ string) error {
	g.calls++
	g.room = roomName
	return nil
}

func TestMatcherDecision(t *testing.T) {
	clinicians := []Clinician{
		{ID: "clinician-1", DisplayName: "Dr Chen", Specialty: "cardiology", Available: false},
		{ID: "clinician-2", DisplayName: "Dr Rao", Specialty: "cardiology", Available: true},
	}

	tests := []struct {
		name          string
		consented     bool
		wantState     string
		wantCalls     int
		wantClinician string
	}{
		{name: "consented patient gets available specialist", consented: true, wantState: "room_ready", wantCalls: 1, wantClinician: "clinician-2"},
		{name: "consent is required before room creation", consented: false, wantState: "waiting", wantCalls: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gateway := &recordingGateway{}
			matcher := Matcher{Gateway: gateway}
			result, err := matcher.Match(context.Background(), AppointmentRequest{
				AppointmentID: "appt-2048", PatientID: "patient-7", Specialty: "cardiology", PatientConsented: test.consented,
			}, clinicians)
			if err != nil {
				t.Fatal(err)
			}
			if result.State != test.wantState || result.ClinicianID != test.wantClinician {
				t.Fatalf("got state=%q clinician=%q", result.State, result.ClinicianID)
			}
			if gateway.calls != test.wantCalls {
				t.Fatalf("gateway calls = %d, want %d", gateway.calls, test.wantCalls)
			}
		})
	}
}
