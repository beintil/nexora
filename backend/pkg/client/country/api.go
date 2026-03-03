package country

import "context"

type Country interface {
	GetAllCountries(ctx context.Context) ([]*GetAllCountriesResponse, error)
}

type NameTranslation struct {
	Official string `json:"official"`
	Common   string `json:"common"`
}

type NativeName struct {
	Eng NameTranslation `json:"eng,omitempty"`
}

type CountryName struct {
	Common     string     `json:"common"`
	Official   string     `json:"official"`
	NativeName NativeName `json:"nativeName"`
}

type GetAllCountriesResponse struct {
	Name CountryName `json:"name"`
	Cca2 string      `json:"cca2"`
}
