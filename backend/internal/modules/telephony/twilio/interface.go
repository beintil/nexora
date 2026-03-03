package twilio

import (
	"context"
	"net/http"
	"telephony/internal/domain"
	"telephony/internal/runner"
	srverr "telephony/internal/shared/server_error"
)

type Handler interface {
	handleTwilioVoiceStatus(w http.ResponseWriter, r *http.Request)

	runner.Runner
}

type Service interface {
	VoiceStatus(ctx context.Context, req *domain.TwilioCallStatusCallback) srverr.ServerError
}

type Repository interface {
}
