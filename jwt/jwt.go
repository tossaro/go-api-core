package jwt

import (
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"strconv"
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
		Key  *string
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

func (jwt *Jwt) generateToken(typ string, exp int, uid uint, key *string, iss string) (string, error) {
	c := TokenClaims{
		uid,
		typ,
		key,
		j.RegisteredClaims{
			Issuer:    iss,
			IssuedAt:  j.NewNumericDate(time.Now()),
			ExpiresAt: j.NewNumericDate(time.Now().Add(time.Minute * time.Duration(jwt.a))),
		},
	}
	token := j.NewWithClaims(j.SigningMethodRS256, c)
	return token.SignedString(jwt.v)
}

func (jwt *Jwt) AccessToken(uid uint, iss string) (*string, error) {
	tk, err := jwt.generateToken("access", jwt.r, uid, nil, iss)
	if err != nil {
		return nil, fmt.Errorf("Jwt : %w", err)
	}
	return &tk, nil
}

func (jwt *Jwt) RefreshToken(uid uint, iss string) (*string, *string, error) {
	t := time.Unix(time.Now().UnixNano(), 0).String()
	s := strconv.FormatUint(uint64(uid), 10) + t
	k := hmac.New(sha256.New, []byte(s))
	hk := hex.EncodeToString(k.Sum(nil))
	tk, err := jwt.generateToken("refresh", jwt.r, uid, &hk, iss)
	if err != nil {
		return nil, nil, fmt.Errorf("Jwt : %w", err)
	}
	return &tk, &hk, nil
}

func (jwt *Jwt) Validate(t string) (*TokenClaims, error) {
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
	if !ok || !token.Valid || claims.UID == 0 {
		return nil, fmt.Errorf("invalid token or missing uid")
	}
	if claims.Type == "refresh" && claims.Key == nil {
		return nil, fmt.Errorf("invalid token or missing key")
	}
	return claims, nil
}
