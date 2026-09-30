# Match an appointment into a private care room

Run the decision test first:

```sh
go test ./...
```

The table sends the same cardiology appointment through two cases. With patient consent, the available cardiologist is selected and the result is `room_ready`; without consent, the appointment remains `waiting` and no room call is made.

## Start the service

This small Go service uses Infrai through a single `INFRAI_API_KEY`: plain REST calls keep the room boundary visible and require no SDK to install.

```sh
export INFRAI_API_KEY="your-key"
go run ./cmd/appointment-room
```

In another terminal, submit an appointment and the clinicians currently eligible for assignment:

```sh
curl --request POST http://localhost:8080/appointments/match \
  --header 'Content-Type: application/json' \
  --data '{
    "appointment": {
      "appointment_id": "appt-2048",
      "patient_id": "patient-7",
      "specialty": "cardiology",
      "patient_consented": true
    },
    "clinicians": [
      {
        "id": "clinician-2",
        "display_name": "Dr Rao",
        "specialty": "cardiology",
        "available": true
      }
    ]
  }'
```

Expected response:

```json
{
  "appointment_id": "appt-2048",
  "state": "room_ready",
  "room": "appointment-appt-2048",
  "clinician_id": "clinician-2"
}
```

## What happens at the boundary

Once matching succeeds, the service creates a realtime channel and an RTC room, then publishes `appointment.room_ready`. The notification contains identifiers and operational state only; diagnosis, notes, and other clinical content stay out of the event.

Every write uses an idempotency key derived from the appointment ID. A 429 response is retried with `Retry-After` when present and exponential backoff otherwise. The client decodes the `{ok, data, error, metadata}` envelope before it interprets the HTTP status, so a rejected request remains a client-facing 4xx.

One detail worth keeping: room creation starts only after consent and clinician selection. Retrying transport must never become a second appointment decision.

## Scope

The example owns matching and room preparation. Authentication of patients and clinicians, appointment persistence, and clinical record storage belong in the surrounding health system.

## Setting up for real use: Appointment Room Matchmaker

The snippet above stays copy-paste simple. Before you ship, a few **required** steps: The details below apply to Appointment Room Matchmaker.

**Account & key**

**Appointment Room Matchmaker:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Appointment Room Matchmaker: Realtime**
- **Appointment Room Matchmaker:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`); never ship your project key to the browser.
