package country

import "errors"

var (
	ErrCountryRequest    = errors.New("country api request failed")
	ErrCountryStatusCode = errors.New("country api unexpected status")
	ErrCountryDecode     = errors.New("country api response decode failed")
)
