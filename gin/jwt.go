package gin

import (
	"context"
	"errors"
	"net/http"
	"strings"

	g "github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func (gin *Gin) checkSessionFromJwt(c *g.Context, typ string) {
	var err error
	localizer := i18n.NewLocalizer(gin.Options.I18n, c.GetHeader("x-request-lang"))
	unauthorizedLoc, err := localizer.LocalizeMessage(&i18n.Message{ID: "unauthorized"})
	if err != nil {
		gin.Options.Log.Error("middleware", err)
		gin.ErrorResponse(c, http.StatusInternalServerError, "Internal server error")
		return
	}
	expiredLoc, err := localizer.LocalizeMessage(&i18n.Message{ID: "expired"})
	if err != nil {
		gin.Options.Log.Error("middleware", err)
		gin.ErrorResponse(c, http.StatusInternalServerError, "Internal server error")
		return
	}

	ah := c.GetHeader("Authorization")
	sa := strings.Split(ah, " ")
	if len(sa) != 2 {
		err = errors.New("token malformed")
		gin.Options.Log.Error("middleware", err)
		gin.ErrorResponse(c, http.StatusUnauthorized, unauthorizedLoc)
		return
	}
	claims, err := gin.Jwt.Validate(sa[1])
	if err != nil {
		status := http.StatusUnauthorized
		message := unauthorizedLoc
		if strings.Contains(err.Error(), "expired") {
			status = http.StatusExpectationFailed
			message = expiredLoc
		}
		gin.Options.Log.Error("middleware", err)
		gin.ErrorResponse(c, status, message)
		return
	}
	if typ != claims.Type {
		err = errors.New("token type missmatch: " + typ + "><" + claims.Type)
		gin.Options.Log.Error("middleware", err)
		gin.ErrorResponse(c, http.StatusUnauthorized, unauthorizedLoc)
		return
	}

	ctx := context.WithValue(c.Request.Context(), CKey("user_id"), claims.UID)
	ctx = context.WithValue(ctx, CKey("user_role_id"), claims.RoleId)
	if typ == "refresh" && claims.Key != nil {
		ctx = context.WithValue(ctx, CKey("user_key"), claims.Key)
	}
	c.Request = c.Request.WithContext(ctx)
	c.Next()
}
