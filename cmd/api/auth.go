package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Auth struct {
	Issuer        string
	Audience      string
	Secret        string
	TokenExpiry   time.Duration
	RefreshExpiry time.Duration
	CookieDomain  string
	CookiePath    string
	CookieName    string
}

type jwtUSer struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type TokenPairs struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type Claims struct {
	jwt.RegisteredClaims
}

func (j *Auth) GenerateTokens(user *jwtUSer) (TokenPairs, error) {
	//create a token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set the token claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	claims["sub"] = fmt.Sprint(user.ID)
	claims["aud"] = j.Audience
	claims["iss"] = j.Issuer
	claims["iat"] = time.Now().UTC().Unix()
	claims["type"] = "JWT"

	// set expirty for token
	claims["exp"] = time.Now().UTC().Add(j.TokenExpiry).Unix()

	// create a signed access token
	signedToken, err := token.SignedString([]byte(j.Secret))
	if err != nil {
		return TokenPairs{}, err
	}

	// same step as above for refresh token
	// create refresh token
	refreshToken := jwt.New(jwt.SigningMethodHS256)

	// set refresh token claims
	refreshClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshClaims["sub"] = fmt.Sprint(user.ID)
	refreshClaims["iat"] = time.Now().UTC().Unix()

	// set expiry for token
	refreshClaims["exp"] = time.Now().UTC().Add(j.RefreshExpiry).Unix()

	// create a signed refreshtoken
	signedRefreshToken, err := refreshToken.SignedString([]byte(j.Secret))
	if err != nil {
		return TokenPairs{}, err
	}

	// create token pairs
	tokenPairs := TokenPairs{
		Token:        signedToken,
		RefreshToken: signedRefreshToken,
	}

	// return token pairs
	return tokenPairs, nil
}

// create a function to generate refresh token
func (j *Auth) GetRefreshCookie(refreshToken string) *http.Cookie {
	return &http.Cookie{
		Name:     j.CookieName,
		Path:     j.CookiePath,
		Value:    refreshToken,
		Expires:  time.Now().Add(j.RefreshExpiry),
		MaxAge:   int(j.RefreshExpiry.Seconds()),
		SameSite: http.SameSiteStrictMode,
		Domain:   j.CookieDomain,
		HttpOnly: true,
		Secure:   true,
	}
}

// delete the token
func (j *Auth) GetExpiredRefreshCookie() *http.Cookie {
	return &http.Cookie{
		Name:     j.CookieName,
		Path:     j.CookiePath,
		Value:    "",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		SameSite: http.SameSiteStrictMode,
		Domain:   j.CookieDomain,
		HttpOnly: true,
		Secure:   true,
	}
}

// create a function to verify the jwt token
func (j *Auth) GetTokenFromHeaderAndVerify(w http.ResponseWriter, r *http.Request) (string, *Claims, error) {
	// first add a header
	w.Header().Add("Vary", "Authorization")

	// get auth header
	authHeader := w.Header().Get("Authorization")

	// check for things in the header
	if authHeader == "" {
		return "", nil, errors.New("no auth header provided")
	}

	// split auth header to get expected stuffs
	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 {
		return "", nil, errors.New("invalid auth header provided")
	}

	// check if part one of the header is Bearer
	if headerParts[0] != "Bearer" {
		return "", nil, errors.New("invalid auth header provided")
	}

	token := headerParts[1]

	// create a store for  claims
	claims := &Claims{}

	// parse token
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		// first validate signing algo
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", token.Header["alg"])
		}

		return []byte(j.Secret), nil
	})

	if err != nil {
		if strings.HasPrefix(err.Error(), "token is expired by") {
			return "", nil, errors.New("expired token")
		}
		return "", nil, err
	}

	// check to see if I issued the token
	if claims.Issuer != j.Issuer {
		return "", nil, errors.New("invalid issuer")
	}

	return token, claims, nil
}
