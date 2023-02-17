package security

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/rlanhellas/aruna/config"
	"github.com/rlanhellas/aruna/logger"
)

type TokenJwt struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	Scope            string `json:"scope"`
}

type JwkPayload struct {
	Keys []*JwkKey `json:"keys"`
}

type JwkKey struct {
	Kid string   `json:"kid"`
	Kty string   `json:"kty"`
	Alg string   `json:"alg"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5C"`
}

// GetTokenJwtWithCreds from Access Token API using custom credentials
func GetTokenJwtWithCreds(clientid, clientsecret, tokenuri string) (*TokenJwt, error) {
	body := io.NopCloser(strings.NewReader(fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s", clientid, clientsecret)))
	resp, err := http.Post(tokenuri, "application/x-www-form-urlencoded", body)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	token := &TokenJwt{}
	err = json.Unmarshal(b, token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

// GetTokenJwt from Access Token API
func GetTokenJwt() (*TokenJwt, error) {
	return GetTokenJwtWithCreds(config.SecurityClientId(), config.SecurityClientSecret(), config.SecurityTokenUri())
}

// ValidateJwt against public certificates and other claims
func ValidateJwt(ctx context.Context, accessToken string) bool {
	accessToken = strings.ReplaceAll(accessToken, "Bearer ", "")
	r, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		rjwk, err := http.Get(config.SecurityJwkUri())
		if err != nil {
			logger.Error(ctx, "error calling jwk uri. %s", err.Error())
			return nil, err
		}

		jwkBody, err := io.ReadAll(rjwk.Body)
		if err != nil {
			logger.Error(ctx, "error reading jwk uri payload. %s", err.Error())
			return nil, err
		}

		jwkPayload := &JwkPayload{}
		err = json.Unmarshal(jwkBody, jwkPayload)
		if err != nil {
			logger.Error(ctx, "error unmarshalling jwk payload. %s", err.Error())
			return nil, err
		}

		kidToken := token.Header["kid"]
		for _, k := range jwkPayload.Keys {
			if k.Kid == kidToken {
				return publicKeyFrom64(k)
			}
		}

		return nil, errors.New("impossible to validate this token against jwk payload")
	})

	if err != nil {
		logger.Error(ctx, "Error validating JWT. %s", err.Error())
		return false
	}

	return r.Valid
}

func publicKeyFrom64(jwk *JwkKey) (*rsa.PublicKey, error) {

	// Create the RSA public key.
	publicKey := &rsa.PublicKey{}

	// Decode the exponent from Base64.
	//
	// According to RFC 7518, this is a Base64 URL unsigned integer.
	// https://tools.ietf.org/html/rfc7518#section-6.3
	if exponent, err := base64.RawURLEncoding.DecodeString(jwk.E); err != nil {
		return nil, err
	} else {
		// Turn the exponent into an integer.
		//
		// According to RFC 7517, these numbers are in big-endian format.
		// https://tools.ietf.org/html/rfc7517#appendix-A.1
		publicKey.E = int(big.NewInt(0).SetBytes(exponent).Uint64())
	}

	// Decode the modulus from Base64.
	if modulus, err := base64.RawURLEncoding.DecodeString(jwk.N); err != nil {
		return nil, err
	} else {
		// Turn the modulus into a *big.Int.
		publicKey.N = big.NewInt(0).SetBytes(modulus)
	}

	return publicKey, nil
}
