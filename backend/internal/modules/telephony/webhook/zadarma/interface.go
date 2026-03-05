package zadarma

import (
	"net/http"
	"telephony/internal/runner"
)

// Handler описывает HTTP-хендлер вебхуков Zadarma.
type Handler interface {
	handleZadarmaVoiceStatus(w http.ResponseWriter, r *http.Request)

	runner.Runner
}
