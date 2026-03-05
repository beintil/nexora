package mango

import (
	"net/http"
	"telephony/internal/runner"
)

// Handler описывает HTTP-хендлер вебхуков Mango.
type Handler interface {
	handleMangoVoiceStatus(w http.ResponseWriter, r *http.Request)

	runner.Runner
}
