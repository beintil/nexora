package twilio

import (
	"net/http"
	"telephony/internal/runner"
)

type Handler interface {
	handleTwilioVoiceStatus(w http.ResponseWriter, r *http.Request)

	runner.Runner
}
