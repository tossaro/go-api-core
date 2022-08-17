package gin

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	g "github.com/gin-gonic/gin"
	pAuth "github.com/tossaro/go-api-core/auth/proto"
	"go.elastic.co/apm"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (gin *Gin) checkSessionFromGrpc(c *g.Context, typ string, rid []int32) {
	span, _ := apm.StartSpan(c.Request.Context(), "checkSessionFromGrpc", "custom")
	defer span.End()

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

	conn, err := grpc.Dial(*gin.Options.AuthService, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		gin.Options.Log.Error("midleware", err)
		gin.ErrorResponse(c, http.StatusUnauthorized, unauthorizedLoc)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	svc := pAuth.NewAuthServiceV1Client(conn)
	resp, err := svc.CheckV1(ctx, &pAuth.CheckReqV1{Token: sa[1], Type: typ})
	if err != nil {
		status := http.StatusUnauthorized
		message := unauthorizedLoc
		if strings.Contains(err.Error(), "expired") && typ == "refresh" {
			status = http.StatusExpectationFailed
			message = expiredLoc
		}
		gin.Options.Log.Error("midleware", err)
		gin.ErrorResponse(c, status, message)
		return
	}

	if len(rid) != 0 {
		var allow bool
		for _, r := range rid {
			if r == resp.GetRid() {
				allow = true
			}
		}
		if !allow {
			forbiddenLoc, errL := localizer.LocalizeMessage(&i18n.Message{ID: "forbidden"})
			if errL != nil {
				gin.Log.Error("middleware", errL)
				gin.ErrorResponse(c, http.StatusInternalServerError, "Internal server error")
				return
			}
			err := errors.New(strconv.FormatUint(resp.GetUid(), 10) + " forbidden access")
			gin.Log.Error("middleware", err)
			gin.ErrorResponse(c, http.StatusForbidden, forbiddenLoc)
			return
		}
	}

	ctx2 := context.WithValue(c.Request.Context(), CKey("user_id"), resp.GetUid())
	ctx2 = context.WithValue(ctx2, CKey("user_role_id"), resp.GetRid())
	if typ == "refresh" && resp.GetKey() != "" {
		ctx2 = context.WithValue(ctx2, CKey("user_key"), resp.GetKey())
	}
	c.Request = c.Request.WithContext(ctx2)
	c.Next()
}
