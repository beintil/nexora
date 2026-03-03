package country

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type country struct {
	client *http.Client
}

func NewCountry() Country {
	return &country{
		client: &http.Client{},
	}
}

func (c *country) GetAllCountries(ctx context.Context) ([]*GetAllCountriesResponse, error) {
	var url = "https://restcountries.com/v3.1/all?fields=cca2,name,translations"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCountryRequest, err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCountryRequest, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrCountryStatusCode, resp.StatusCode)
	}

	var countries []*GetAllCountriesResponse
	if err := json.NewDecoder(resp.Body).Decode(&countries); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCountryDecode, err)
	}
	return countries, nil
}
