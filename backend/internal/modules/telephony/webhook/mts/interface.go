package mts

import (
	"net/http"
	"telephony/internal/runner"
)

// Handler описывает HTTP-хендлер вебхуков МТС.
type Handler interface {
	handleMTSVoiceStatus(w http.ResponseWriter, r *http.Request)

	runner.Runner
}
