package jwt

import (
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	j "github.com/golang-jwt/jwt/v4"
)

type (
	Jwt struct {
		PrivateKey *rsa.PrivateKey
		PublicKey  *rsa.PublicKey
		Options    *Options
	}

	Options struct {
		PrivateKeyPath       string
		PublicKeyPath        string
		AccessTokenLifetime  int
		RefreshTokenLifetime int
	}

	TokenClaimsV1 struct {
		UID  uint64
		Type string
		Key  *string
		j.RegisteredClaims
	}
)

func NewRSA(o *Options) *Jwt {
	if o.PrivateKeyPath == "" {
		log.Fatal("jwt - PrivateKeyPath option not provided")
	}
	if o.PublicKeyPath == "" {
		log.Fatal("jwt - PublicKeyPath option not provided")
	}
	if o.AccessTokenLifetime == 0 {
		log.Fatal("jwt - AccessTokenLifetime option not provided")
	}
	if o.RefreshTokenLifetime == 0 {
		log.Fatal("jwt - RefreshTokenLifetime option not provided")
	}

	vb, err := ioutil.ReadFile(o.PrivateKeyPath)
	if err != nil {
		log.Fatal("jwt - read private key error: %w", err)
	}

	vk, err := j.ParseRSAPrivateKeyFromPEM(vb)
	if err != nil {
		log.Fatal("jwt - parse private key error: %w", err)
	}

	cb, err := ioutil.ReadFile(o.PublicKeyPath)
	if err != nil {
		log.Fatal("jwt - read public key error:", err)
	}

	ck, err := j.ParseRSAPublicKeyFromPEM(cb)
	if err != nil {
		log.Fatal("jwt - parse public key error: %w", err)
	}

	return &Jwt{vk, ck, o}
}

func (jwt *Jwt) generateToken(typ string, exp int, uid uint64, key *string, iss string) (string, error) {
	c := TokenClaimsV1{
		uid,
		typ,
		key,
		j.RegisteredClaims{
			Issuer:    iss,
			IssuedAt:  j.NewNumericDate(time.Now()),
			ExpiresAt: j.NewNumericDate(time.Now().Add(time.Minute * time.Duration(exp))),
		},
	}
	token := j.NewWithClaims(j.SigningMethodRS256, c)
	return token.SignedString(jwt.PrivateKey)
}

func (jwt *Jwt) AccessToken(uid uint64, iss string) (*string, error) {
	tk, err := jwt.generateToken("access", jwt.Options.AccessTokenLifetime, uid, nil, iss)
	if err != nil {
		return nil, err
	}
	return &tk, nil
}

func (jwt *Jwt) RefreshToken(uid uint64, iss string) (*string, *string, error) {
	t := time.Unix(time.Now().UnixNano(), 0).String()
	s := strconv.FormatUint(uint64(uid), 10) + t
	k := hmac.New(sha256.New, []byte(s))
	hk := hex.EncodeToString(k.Sum(nil))
	tk, err := jwt.generateToken("refresh", jwt.Options.RefreshTokenLifetime, uid, &hk, iss)
	if err != nil {
		return nil, nil, err
	}
	return &tk, &hk, nil
}

func (jwt *Jwt) Validate(t string) (*TokenClaimsV1, error) {
	token, err := j.ParseWithClaims(t, &TokenClaimsV1{}, func(token *j.Token) (interface{}, error) {
		if _, ok := token.Method.(*j.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method in auth token")
		}
		return jwt.PublicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parse claims: %v", err)
	}

	claims, ok := token.Claims.(*TokenClaimsV1)
	if !ok || !token.Valid || claims.UID == 0 {
		return nil, fmt.Errorf("invalid token or missing uid")
	}
	if claims.Type == "refresh" && claims.Key == nil {
		return nil, fmt.Errorf("invalid token or missing key")
	}
	return claims, nil
}
