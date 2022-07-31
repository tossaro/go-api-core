package gin

import (
	"log"
	"net/http"

	g "github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	prm "github.com/prometheus/client_golang/prometheus/promhttp"
	sf "github.com/swaggo/files"
	gsw "github.com/swaggo/gin-swagger"
	ch "github.com/tossaro/go-api-core/http"
	cj "github.com/tossaro/go-api-core/jwt"
	cl "github.com/tossaro/go-api-core/logger"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmgin"
)

const (
	AuthTypeGrpc = "grpc"
	AuthTypeJwt  = "jwt"
)

type (
	CKey string

	Error struct {
		Message string `json:"error" example:"message"`
	}

	Gin struct {
		Gin    *g.Engine
		Router *g.RouterGroup
		Jwt    *cj.Jwt
		*Options
	}

	Options struct {
		I18n        *i18n.Bundle
		Mode        string
		Version     string
		BaseUrl     string
		Log         cl.Interface
		AuthType    string
		AuthService *string
		Jwt         *cj.Jwt
		Captcha     *bool
	}

	TokenV1 struct {
		Access  string `json:"access"`
		Refresh string `json:"refresh"`
	}
)

func New(o *Options) *Gin {
	if o.I18n == nil {
		log.Fatal("gin - I18n option not provided")
	}
	if o.Mode == "" {
		log.Fatal("gin - Mode option not provided")
	}
	if o.Version == "" {
		log.Fatal("gin - Version option not provided")
	}
	if o.BaseUrl == "" {
		log.Fatal("gin - BaseUrl option not provided")
	}
	if o.Log == nil {
		log.Fatal("gin - Log option not provided")
	}
	if o.AuthType == "" {
		log.Fatal("gin - AuthType option auth type not provided")
	}
	if o.AuthType == AuthTypeGrpc && o.AuthService == nil {
		log.Fatal("gin - AuthTypeGrpc require AuthService option")
	}
	if o.AuthType == AuthTypeJwt && o.Jwt == nil {
		log.Fatal("gin - AuthTypeJwt require Jwt option")
	}

	g.SetMode(o.Mode)
	r := g.Default()
	r.Use(apmgin.Middleware(r))

	gin := &Gin{r, nil, o.Jwt, o}

	gRouter := r.Group(o.BaseUrl)
	{
		gRouter.GET("/version", gin.version)
		gRouter.GET("/metrics", g.WrapH(prm.Handler()))
		gRouter.GET("/swagger/*any", gsw.DisablingWrapHandler(sf.Handler, "HTTP_SWAGGER_DISABLED"))

		if o.Captcha != nil && *(o.Captcha) {
			ch.NewCaptchaV1(gRouter, o.Log)
		}
	}

	gRouter.Use(validateHeader(gin))
	gin.Router = gRouter
	return gin
}

func validateHeader(gin *Gin) g.HandlerFunc {
	return func(c *g.Context) {
		span, _ := apm.StartSpan(c.Request.Context(), "ValidateHeader", "custom")
		defer span.End()

		localizer := i18n.NewLocalizer(gin.Options.I18n, c.GetHeader("x-request-lang"))
		missHeader, err := localizer.LocalizeMessage(&i18n.Message{ID: "missing_header"})
		if err != nil {
			gin.Options.Log.Error("jwt-validate", err)
			gin.ErrorResponse(c, http.StatusInternalServerError, "Internal server error")
			return
		}
		l := c.GetHeader("x-request-lang")
		if l == "" {
			gin.ErrorResponse(c, http.StatusBadRequest, missHeader)
			return
		}

		p := c.GetHeader("x-request-key")
		if p == "" {
			gin.ErrorResponse(c, http.StatusBadRequest, missHeader)
			return
		}
	}
}

func (gin *Gin) ErrorResponse(c *g.Context, code int, msg string) {
	span, _ := apm.StartSpan(c.Request.Context(), "ErrorResponse", "error")
	defer span.End()
	c.AbortWithStatusJSON(code, &Error{msg})
}

func (gin *Gin) AuthAccessMiddleware() g.HandlerFunc {
	return gin.authCheck("access")
}

func (gin *Gin) AuthRefreshMiddleware() g.HandlerFunc {
	return gin.authCheck("refresh")
}

func (gin *Gin) authCheck(typ string) g.HandlerFunc {
	return func(c *g.Context) {
		switch gin.Options.AuthType {
		case AuthTypeGrpc:
			gin.checkSessionFromGrpc(c, typ)
		case AuthTypeJwt:
			gin.checkSessionFromJwt(c, typ)
		default:
			gin.checkSessionFromGrpc(c, typ)
		}
	}
}

// @Summary     Get Version
// @Description Get Version
// @ID          version
// @Tags  	    API
// @Accept      json
// @Produce     json
// @Success     200 {string} v1.0.0
// @Router      /version [get]
func (gin *Gin) version(c *g.Context) {
	span, _ := apm.StartSpan(c.Request.Context(), "version", "request")
	defer span.End()
	c.JSON(http.StatusOK, gin.Options.Version)
}
