package response

import (
	"encoding/json"
	"net/http"
	"telephony/models"
	"telephony/pkg/logger"
	"telephony/pkg/trimmed"
)

type httpResponse struct {
	log     logger.Logger
	isDebug bool
}

func (m *httpResponse) ErrorResponse(w http.ResponseWriter, r *http.Request, err *models.TransportError) {
	m.WriteResponse(w, r, int(*err.Code), err)
}

func (m *httpResponse) WriteResponse(w http.ResponseWriter, r *http.Request, code int, resp any) {
	raw, err := json.Marshal(resp)
	if err != nil {
		m.log.Errorf("error marshalling response: %v", err)
		return
	}
	body, err := json.Marshal(resp)
	if err != nil {
		m.log.Errorf("error marshalling response: %v", err)
		return
	}

	pretty, _ := json.MarshalIndent(body, "", "  ")
	if m.isDebug || code >= 500 {
		if code >= 500 {
			stack := trimmed.GetTrimmedStack(2)

			m.log.Errorf("[%s] '%s' [%s] Response: %s", r.Method, r.URL.String(), stack, pretty)
		} else {
			m.log.Debugf("[%s] '%s' Response: %s", r.Method, r.URL.String(), pretty)
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_, _ = w.Write(raw)
}

func NewHTTPResponse(log logger.Logger, isDebag bool) HttpResponse {
	return &httpResponse{
		log:     log,
		isDebug: isDebag,
	}
}
