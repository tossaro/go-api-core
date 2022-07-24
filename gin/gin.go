package gin

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	g "github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	prm "github.com/prometheus/client_golang/prometheus/promhttp"
	sf "github.com/swaggo/files"
	gsw "github.com/swaggo/gin-swagger"
	"github.com/tossaro/go-api-core/jwt"
	"github.com/tossaro/go-api-core/logger"
)

type (
	CKey string

	Error struct {
		Msg string `json:"error" example:"message"`
	}

	Gin struct {
		L logger.Interface
		E *g.Engine
		R *g.RouterGroup
		r *redis.Client
		j *jwt.Jwt
	}

	TokenV1 struct {
		Access  string `json:"access"`
		Refresh string `json:"refresh"`
	}
)

func New(b string, l logger.Interface, rdb *redis.Client, at int, rt int) *Gin {
	gi := g.New()
	gi.Use(g.Logger())
	gi.Use(g.Recovery())

	j, err := jwt.New(at, rt)
	if err != nil {
		log.Fatalf("JWT error: %s", err)
	}

	ge := &Gin{l, gi, nil, rdb, j}

	sw := gsw.DisablingWrapHandler(sf.Handler, "DISABLE_SWAGGER_HTTP_HANDLER")
	r := gi.Group(b)
	{
		r.GET("/swagger/*any", sw)
		r.GET("/healthz", func(c *g.Context) { c.Status(http.StatusOK) })
		r.GET("/metrics", g.WrapH(prm.Handler()))
	}
	r.Use(headerCheck(ge))
	ge.R = r
	return ge
}

func headerCheck(ge *Gin) g.HandlerFunc {
	return func(c *g.Context) {
		l := c.GetHeader("x-platform-lang")
		if l == "" {
			ge.ErrorResponse(c, http.StatusBadRequest, "missing header param", "http-auth", nil)
			return
		}

		p := c.GetHeader("x-request-key")
		if p == "" {
			ge.ErrorResponse(c, http.StatusBadRequest, "missing header param", "http-auth", nil)
			return
		}
	}
}

func (ge *Gin) ErrorResponse(c *g.Context, code int, msg string, iss string, err error) {
	if err != nil {
		ge.L.Error(err, iss)
	}
	c.AbortWithStatusJSON(code, &Error{msg})
}

func (ge *Gin) AuthAccess() g.HandlerFunc {
	return ge.checkAuth("access")
}

func (ge *Gin) AuthRefresh() g.HandlerFunc {
	return ge.checkAuth("refresh")
}

func (ge *Gin) checkAuth(typ string) g.HandlerFunc {
	return func(c *g.Context) {
		a := c.GetHeader("Authorization")
		sa := strings.Split(a, " ")
		if len(sa) != 2 {
			ge.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "http-auth", nil)
			return
		}
		claims, err := ge.j.Validate(sa[1])
		if err != nil || typ != claims.Type {
			ge.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "http-auth", err)
			return
		}

		ctx := context.WithValue(c.Request.Context(), CKey("user_id"), claims.UID)
		if typ == "refresh" && claims.Key != nil {
			ctx = context.WithValue(ctx, CKey("user_key"), claims.Key)
		}
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func (a *Gin) NewSession(uid uint) (TokenV1, error) {
	var t TokenV1
	ac, err := a.j.AccessToken(uid, "signin")
	if err != nil {
		return t, err
	}

	rf, k, err := a.j.RefreshToken(uid, "signin")
	if err != nil {
		return t, err
	}

	err = a.r.Set(context.Background(), *(k), "0", 0).Err()
	if err != nil {
		return t, err
	}

	return TokenV1{Access: *(ac), Refresh: *(rf)}, nil
}

func (a *Gin) RefreshSession(uid uint, key string, req string) (TokenV1, error) {
	var t TokenV1
	v, _ := a.r.Get(context.Background(), key).Result()
	if v != "0" && v != req {
		return t, fmt.Errorf(key)
	}

	if v == req {
		nt, err := a.r.Get(context.Background(), key+"_issued").Result()
		if err != nil {
			return t, err
		}

		err = json.Unmarshal([]byte(nt), &t)
		if err != nil {
			return t, err
		}

		return t, nil
	}

	err := a.r.Set(context.Background(), key, req, 0).Err()
	if err != nil {
		return t, err
	}

	ses, err := a.NewSession(uid)
	if err != nil {
		return t, err
	}

	jses, err := json.Marshal(ses)
	if err != nil {
		return t, err
	}

	err = a.r.Set(context.Background(), key+"_issued", jses, 0).Err()
	if err != nil {
		return t, err
	}

	return ses, nil
}
