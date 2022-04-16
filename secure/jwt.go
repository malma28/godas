package secure

import (
	"errors"
	"godas/model/web"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const DefaultJWTExpiration = time.Hour * 24 * 30

type JWTProvider struct {
	expiration   time.Duration
	issuer       string
	signatureKey []byte
}

func NewJWTProvider(expiration time.Duration, issuer string, signatureKey string) *JWTProvider {
	jwtProvider := new(JWTProvider)
	jwtProvider.expiration = expiration
	jwtProvider.issuer = issuer
	jwtProvider.signatureKey = []byte(signatureKey)

	return jwtProvider
}

// Create token
func (provider *JWTProvider) Token(claims web.JwtClaims) (string, error) {
	now := time.Now()

	claims.RegisteredClaims = jwt.RegisteredClaims{
		Subject:   claims.UserID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(provider.expiration)),
		Issuer:    provider.issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(provider.signatureKey)
}

func (provider *JWTProvider) Validate(tokenString string) (web.JwtClaims, error) {
	claims := web.JwtClaims{}

	parsedToken, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		method, isHS256 := token.Method.(*jwt.SigningMethodHMAC)
		if !isHS256 || method != jwt.SigningMethodHS256 {
			return nil, errors.New("invalid signing method")
		}
		return provider.signatureKey, nil
	})
	if err != nil {
		return claims, err
	}

	if !parsedToken.Valid {
		return claims, errors.New("invalid claims")
	}

	return claims, nil
}
