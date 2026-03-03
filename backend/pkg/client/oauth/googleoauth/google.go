package googleoauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const userinfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"

type userinfoResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

type client struct {
	config      *oauth2.Config
	http        *http.Client
	userinfoURL string
}

// NewClient создаёт клиент Google OAuth. clientID, clientSecret и redirectURL обязательны.
// userinfoURL — URL userinfo API; если пустой, используется userinfoURL по умолчанию.
func NewClient(clientID, clientSecret, redirectURL, userinfoURL string) (Client, error) {
	if clientID == "" || clientSecret == "" || redirectURL == "" {
		return nil, fmt.Errorf("googleoauth: client_id, client_secret and redirect_url are required")
	}
	if userinfoURL == "" {
		userinfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	}
	cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
	return &client{
		config:      cfg,
		http:        &http.Client{},
		userinfoURL: userinfoURL,
	}, nil
}

// BuildRedirectURL собирает полный redirect_uri из baseURL (схема + хост) и path.
func BuildRedirectURL(baseURL, path string) (string, error) {
	baseURL = strings.TrimSuffix(baseURL, "/")
	if baseURL == "" || path == "" {
		return "", fmt.Errorf("googleoauth: baseURL and path required")
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	u.Path = ""
	u.RawPath = ""
	return strings.TrimSuffix(u.String(), "/") + path, nil
}

func (c *client) AuthCodeURL(state string) string {
	return c.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (c *client) ExchangeAndGetUserInfo(ctx context.Context, code string) (email, name, picture string, err error) {
	tok, err := c.config.Exchange(ctx, code)
	if err != nil {
		return "", "", "", fmt.Errorf("%w: %v", ErrGoogleExchange, err)
	}
	if tok.AccessToken == "" {
		return "", "", "", ErrGoogleEmptyToken
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.userinfoURL, nil)
	if err != nil {
		return "", "", "", fmt.Errorf("%w: %v", ErrGoogleUserinfo, err)
	}
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("%w: %v", ErrGoogleUserinfo, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("%w: status %d", ErrGoogleUserinfo, resp.StatusCode)
	}

	var info userinfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", "", "", fmt.Errorf("%w: %v", ErrGoogleUserinfoBody, err)
	}

	return strings.TrimSpace(info.Email), strings.TrimSpace(info.Name), strings.TrimSpace(info.Picture), nil
}
