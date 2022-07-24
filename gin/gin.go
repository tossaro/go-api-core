package gin

import (
	"context"
	"net/http"
	"strings"

	g "github.com/gin-gonic/gin"
	prm "github.com/prometheus/client_golang/prometheus/promhttp"
	sf "github.com/swaggo/files"
	gsw "github.com/swaggo/gin-swagger"
	"github.com/tossaro/go-api-core/jwt"
	"github.com/tossaro/go-api-core/logger"
)

type (
	Error struct {
		Msg string `json:"error" example:"message"`
	}

	Gin struct {
		L logger.Interface
		E *g.Engine
		R *g.RouterGroup
	}

	CKey string
)

func New(b string, l logger.Interface) *Gin {
	gi := g.New()
	gi.Use(g.Logger())
	gi.Use(g.Recovery())
	sw := gsw.DisablingWrapHandler(sf.Handler, "DISABLE_SWAGGER_HTTP_HANDLER")
	ge := &Gin{l, gi, nil}
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

func (ge *Gin) AuthAccess(jwt *jwt.Jwt) g.HandlerFunc {
	return ge.checkAuth(jwt, "access")
}

func (ge *Gin) AuthRefresh(jwt *jwt.Jwt) g.HandlerFunc {
	return ge.checkAuth(jwt, "refresh")
}

func (ge *Gin) checkAuth(jwt *jwt.Jwt, typ string) g.HandlerFunc {
	return func(c *g.Context) {
		a := c.GetHeader("Authorization")
		sa := strings.Split(a, " ")
		if len(sa) != 2 {
			ge.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "http-auth", nil)
			return
		}
		claims, err := jwt.Validate(sa[1])
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
