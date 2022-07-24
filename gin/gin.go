package gin

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	g "github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/tossaro/go-api-core/jwt"
	"github.com/tossaro/go-api-core/logger"
)

type (
	Error struct {
		Msg string `json:"error" example:"message"`
	}

	Gin struct {
		R *g.RouterGroup
		L logger.Interface
	}

	CKey string
)

func New(g *g.RouterGroup, l logger.Interface) *Gin {
	return &Gin{g, l}
}

func (ge *Gin) ErrorResponse(c *g.Context, code int, msg string, iss string, err error) {
	ge.L.Error(err, iss)
	c.AbortWithStatusJSON(code, &Error{msg})
}

func (ge *Gin) checkAuth(rdb *redis.Client, jwt *jwt.Jwt, typ string) g.HandlerFunc {
	return func(c *g.Context) {
		a := c.GetHeader("Authorization")
		sa := strings.Split(a, " ")
		if len(sa) != 2 {
			ge.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "http-auth", fmt.Errorf("token not provided or malformed"))
			return
		}
		claims, err := jwt.Validate(sa[1])
		if err != nil || typ != claims.Type {
			ge.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "http-auth", err)
			return
		}

		if typ == "refresh" {
			_, err := rdb.Get(context.Background(), *claims.Key).Result()
			if err != nil {
				ge.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "http-auth", err)
				return
			}
		}

		ctx := context.WithValue(c.Request.Context(), CKey("user_id"), claims.UID)
		if typ == "refresh" && claims.Key != nil {
			ctx = context.WithValue(ctx, CKey("user_key"), &claims.Key)
		}
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func (ge *Gin) AuthAccess(rdb *redis.Client, jwt *jwt.Jwt) g.HandlerFunc {
	return ge.checkAuth(rdb, jwt, "access")
}

func (ge *Gin) AuthRefresh(rdb *redis.Client, jwt *jwt.Jwt) g.HandlerFunc {
	return ge.checkAuth(rdb, jwt, "refresh")
}
