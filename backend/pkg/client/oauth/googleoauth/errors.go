package googleoauth

import "errors"

var (
	ErrGoogleExchange     = errors.New("google oauth exchange failed")
	ErrGoogleEmptyToken   = errors.New("google oauth empty access token")
	ErrGoogleUserinfo     = errors.New("google userinfo request failed")
	ErrGoogleUserinfoBody = errors.New("google userinfo decode failed")
)
