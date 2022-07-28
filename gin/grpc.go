package gin

import (
	"context"
	"net/http"
	"strings"
	"time"

	g "github.com/gin-gonic/gin"
	pAuth "github.com/tossaro/go-api-core/auth/proto"
)

func (gin *Gin) checkSessionFromGrpc(c *g.Context, typ string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ah := c.GetHeader("Authorization")
	sa := strings.Split(ah, " ")
	if len(sa) != 2 {
		gin.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "http-auth", nil)
		return
	}

	a := *gin.AuthClient
	r, err := a.CheckV1(ctx, &pAuth.AuthRequestV1{Token: sa[1], Type: typ})
	if err != nil {
		gin.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "http-auth", nil)
		return
	}

	ctx2 := context.WithValue(c.Request.Context(), CKey("user_id"), r.GetUID())
	if typ == "refresh" && r.GetKey() != "" {
		ctx2 = context.WithValue(ctx2, CKey("user_key"), r.GetKey())
	}
	c.Request = c.Request.WithContext(ctx2)
	c.Next()
}
