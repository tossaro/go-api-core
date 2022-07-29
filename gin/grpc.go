package gin

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	g "github.com/gin-gonic/gin"
	pAuth "github.com/tossaro/go-api-core/auth/proto"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (gin *Gin) checkSessionFromGrpc(c *g.Context, typ string) {
	var err error
	localizer := i18n.NewLocalizer(gin.Options.I18n, c.GetHeader("x-request-lang"))
	unauthorizedLoc, err := localizer.LocalizeMessage(&i18n.Message{ID: "unauthorized"})
	if err != nil {
		gin.ErrorResponse(c, http.StatusInternalServerError, "Internal server error", "jwt-validate", err)
		return
	}
	expiredLoc, err := localizer.LocalizeMessage(&i18n.Message{ID: "expired"})
	if err != nil {
		gin.ErrorResponse(c, http.StatusInternalServerError, "Internal server error", "jwt-validate", err)
		return
	}

	ah := c.GetHeader("Authorization")
	sa := strings.Split(ah, " ")
	if len(sa) != 2 {
		err = errors.New("token malformed")
		gin.ErrorResponse(c, http.StatusUnauthorized, unauthorizedLoc, "http-auth", err)
		return
	}

	conn, err := grpc.Dial(*gin.Options.AuthService, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		gin.Options.Log.Error("Gin init auth error: %s", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	svc := pAuth.NewAuthServiceV1Client(conn)
	r, err := svc.CheckV1(ctx, &pAuth.AuthRequestV1{Token: sa[1], Type: typ})
	if err != nil {
		status := http.StatusUnauthorized
		message := unauthorizedLoc
		if strings.Contains(err.Error(), "expired") && typ == "refresh" {
			status = http.StatusExpectationFailed
			message = expiredLoc
		}
		gin.ErrorResponse(c, status, message, "jwt-validate", err)
		return
	}

	ctx2 := context.WithValue(c.Request.Context(), CKey("user_id"), r.GetUID())
	if typ == "refresh" && r.GetKey() != "" {
		ctx2 = context.WithValue(ctx2, CKey("user_key"), r.GetKey())
	}
	c.Request = c.Request.WithContext(ctx2)
	c.Next()
}
