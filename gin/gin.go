package gin

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	g "github.com/gin-gonic/gin"
	"github.com/tossaro/go-api-core/jwt"
	"github.com/tossaro/go-api-core/logger"
)

type (
	Error struct {
		Error string `json:"error" example:"message"`
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

func (ge *Gin) AuthCheck(jwt *jwt.Jwt) g.HandlerFunc {
	return func(c *g.Context) {
		a := c.GetHeader("Authorization")
		sa := strings.Split(a, " ")
		if len(sa) != 2 {
			ge.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "http-auth", fmt.Errorf("token not provided or malformed"))
			return
		}
		uid, err := jwt.Validate(sa[1])
		if err != nil {
			ge.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "http-auth", err)
			return
		}

		ctx := context.WithValue(c.Request.Context(), CKey("user_id"), uid)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
