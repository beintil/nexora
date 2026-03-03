package appleoauth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultAuthURL = "https://appleid.apple.com/auth/authorize"
	defaultJWKSURL = "https://appleid.apple.com/auth/keys"
	defaultIssuer  = "https://appleid.apple.com"
)

type jwksResponse struct {
	Keys []jwkKey `json:"keys"`
}

type jwkKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type idTokenClaims struct {
	jwt.RegisteredClaims
	Email          string `json:"email"`
	EmailVerified  string `json:"email_verified"`
	Nonce          string `json:"nonce"`
	RealUserStatus int    `json:"real_user_status,omitempty"`
}

type client struct {
	clientID string
	authURL  string
	issuer   string
	jwks     *jwksCache
}

type jwksCache struct {
	keys    []jwkKey
	jwksURL string
}

// NewClient создаёт клиент Apple Sign in. clientID (Service ID) обязателен.
// authURL, jwksURL, issuer — URL авторизации, ключей и issuer; если пустые, используются дефолты.
func NewClient(clientID, authURL, jwksURL, issuer string) (Client, error) {
	if clientID == "" {
		return nil, fmt.Errorf("appleoauth: client_id is required")
	}
	if authURL == "" {
		authURL = defaultAuthURL
	}
	if jwksURL == "" {
		jwksURL = defaultJWKSURL
	}
	if issuer == "" {
		issuer = defaultIssuer
	}
	return &client{
		clientID: clientID,
		authURL:  authURL,
		issuer:   issuer,
		jwks:     &jwksCache{jwksURL: jwksURL},
	}, nil
}

func (c *client) AuthCodeURL(redirectURI, state string) string {
	v := url.Values{}
	v.Set("client_id", c.clientID)
	v.Set("redirect_uri", redirectURI)
	v.Set("response_type", "code id_token")
	v.Set("response_mode", "form_post")
	v.Set("scope", "name email")
	v.Set("state", state)
	return c.authURL + "?" + v.Encode()
}

func (c *client) VerifyIDToken(ctx context.Context, idToken string) (email, sub string, err error) {
	token, err := jwt.ParseWithClaims(idToken, &idTokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		kid, ok := t.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("%w: missing kid", ErrAppleInvalidToken)
		}
		if t.Method != jwt.SigningMethodRS256 {
			return nil, fmt.Errorf("%w: unexpected alg", ErrAppleInvalidToken)
		}
		key, err := c.jwks.getKey(ctx, kid)
		if err != nil {
			return nil, err
		}
		return key, nil
	})
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", ErrAppleInvalidToken, err)
	}
	claims, ok := token.Claims.(*idTokenClaims)
	if !ok || !token.Valid {
		return "", "", ErrAppleInvalidToken
	}
	if claims.Issuer != c.issuer {
		return "", "", fmt.Errorf("%w: wrong issuer", ErrAppleInvalidToken)
	}
	if claims.Audience != nil {
		audOK := false
		for _, a := range claims.Audience {
			if a == c.clientID {
				audOK = true
				break
			}
		}
		if !audOK {
			return "", "", fmt.Errorf("%w: audience mismatch", ErrAppleInvalidToken)
		}
	}
	return strings.TrimSpace(claims.Email), strings.TrimSpace(claims.Subject), nil
}

func (c *jwksCache) getKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	if len(c.keys) == 0 {
		if err := c.fetch(ctx); err != nil {
			return nil, err
		}
	}
	for _, k := range c.keys {
		if k.Kid == kid {
			return jwkToRSA(k)
		}
	}
	return nil, fmt.Errorf("%w: key kid %s", ErrAppleJWKS, kid)
}

func (c *jwksCache) fetch(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.jwksURL, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrAppleJWKS, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: status %d", ErrAppleJWKS, resp.StatusCode)
	}
	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("%w: %v", ErrAppleJWKS, err)
	}
	c.keys = jwks.Keys
	return nil
}

func jwkToRSA(k jwkKey) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, fmt.Errorf("%w: n decode: %v", ErrAppleJWKS, err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, fmt.Errorf("%w: e decode: %v", ErrAppleJWKS, err)
	}
	if len(eBytes) < 4 {
		return nil, fmt.Errorf("%w: e too short", ErrAppleJWKS)
	}
	var e int
	for _, b := range eBytes {
		e = e<<8 | int(b)
	}
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: e,
	}, nil
}
