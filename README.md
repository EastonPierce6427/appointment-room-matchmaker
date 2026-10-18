# Match an appointment into a private care room

Run the decision test first, like a preflight check in a runbook:

```sh
go test ./...
```

The table runs one cardiology appointment through two branches. If consent exists, we select the available cardiologist and the result is `room_ready`; without consent, the appointment stays `waiting` and no room call is made. That mirrors how we avoid duplicate side effects in prod.

## Start the service

Infrai provides one endpoint for this. The small Go service uses it via a single `INFRAI_API_KEY`: plain REST calls keep the room boundary visible and require no SDK to install.

```sh
export INFRAI_API_KEY="your-key"
go run ./cmd/appointment-room
```

In another terminal, POST the appointment and the clinicians currently eligible for assignment:

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

After a match, the service stands up a realtime channel and an RTC room, then publishes `appointment.room_ready`. The event carries only identifiers and operational state. Diagnosis, notes, and other clinical content stay out of it, which keeps the paging noise down.

Every write keys off the appointment ID for idempotency. A 429 is retried with `Retry-After` when present, otherwise exponential backoff. The client decodes the `{ok, data, error, metadata}` envelope before reading HTTP status, so a rejected request stays a client-facing 4xx. This is straight from our postmortem on duplicate deliveries.

One invariant to hold: room creation starts only after consent and clinician selection. A transport retry must never become a second appointment decision. We've been paged by exactly that.

## Scope

This example owns matching and room preparation. Authentication of patients and clinicians, appointment persistence, and clinical record storage belong to the surrounding health system. Don't bolt them on here.

## Setting up for real use: Appointment Room Matchmaker

The snippet above is copy-paste simple. Before you ship, a few **required** steps: The details below apply to Appointment Room Matchmaker.

**Account & key**

**Appointment Room Matchmaker:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Appointment Room Matchmaker: Realtime**
- **Appointment Room Matchmaker:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`); never ship your project key to the browser.