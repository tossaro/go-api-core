package jwt

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"time"

	j "github.com/golang-jwt/jwt/v4"
)

type (
	Jwt struct {
		v *rsa.PrivateKey
		c *rsa.PublicKey
		a int
		r int
	}

	TokenClaims struct {
		UID  uint
		Type string
		j.RegisteredClaims
	}
)

func New(a int, r int) (*Jwt, error) {
	vb, err := ioutil.ReadFile("./key_private.pem")
	if err != nil {
		return nil, fmt.Errorf("Jwt : %w", err)
	}

	vk, err := j.ParseRSAPrivateKeyFromPEM(vb)
	if err != nil {
		return nil, fmt.Errorf("Jwt : %w", err)
	}

	cb, err := ioutil.ReadFile("./key_public.pem")
	if err != nil {
		return nil, fmt.Errorf("Jwt : %w", err)
	}

	ck, err := j.ParseRSAPublicKeyFromPEM(cb)
	if err != nil {
		return nil, fmt.Errorf("Jwt : %w", err)
	}

	return &Jwt{vk, ck, a, r}, nil
}

func (jwt *Jwt) generateToken(typ string, exp int, uid uint, iss string) (string, error) {
	c := TokenClaims{
		uid,
		typ,
		j.RegisteredClaims{
			ExpiresAt: j.NewNumericDate(time.Now().Add(time.Minute * time.Duration(jwt.a))),
			Issuer:    iss,
		},
	}
	token := j.NewWithClaims(j.SigningMethodRS256, c)
	return token.SignedString(jwt.v)
}

func (jwt *Jwt) AccessToken(uid uint, iss string) (string, error) {
	return jwt.generateToken("access", jwt.a, uid, iss)
}

func (jwt *Jwt) RefreshToken(uid uint, iss string) (string, error) {
	return jwt.generateToken("refresh", jwt.r, uid, iss)
}

func (jwt *Jwt) Validate(t string) (*uint, error) {
	token, err := j.ParseWithClaims(t, &TokenClaims{}, func(token *j.Token) (interface{}, error) {
		if _, ok := token.Method.(*j.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method in auth token")
		}
		return jwt.c, nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parse claims: %v", err)
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid || claims.UID == 0 || claims.Type != "access" {
		return nil, fmt.Errorf("invalid token: authentication failed")
	}
	return &claims.UID, nil
}
