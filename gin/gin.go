package gin

import (
	"log"
	"net/http"

	g "github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	prm "github.com/prometheus/client_golang/prometheus/promhttp"
	sf "github.com/swaggo/files"
	gsw "github.com/swaggo/gin-swagger"
	"github.com/tossaro/go-api-core/captcha"
	j "github.com/tossaro/go-api-core/jwt"
	"github.com/tossaro/go-api-core/logger"
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
		Jwt    *j.Jwt
		*Options
	}

	Options struct {
		I18n        *i18n.Bundle
		Mode        string
		Version     string
		BaseUrl     string
		Log         logger.Interface
		AuthType    string
		AuthService *string
		Jwt         *j.Jwt
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
	gEngine := g.New()
	gEngine.Use(g.Logger())
	gEngine.Use(g.Recovery())

	gin := &Gin{gEngine, nil, o.Jwt, o}

	gRouter := gEngine.Group(o.BaseUrl)
	{
		gRouter.GET("/version", gin.version)
		gRouter.GET("/metrics", g.WrapH(prm.Handler()))
		gRouter.GET("/swagger/*any", gsw.DisablingWrapHandler(sf.Handler, "HTTP_SWAGGER_DISABLED"))

		if o.Captcha != nil && *(o.Captcha) {
			captcha.New(gRouter, o.Log)
		}
	}

	gRouter.Use(headerCheck(gin))
	gin.Router = gRouter
	return gin
}

func headerCheck(gin *Gin) g.HandlerFunc {
	return func(c *g.Context) {
		localizer := i18n.NewLocalizer(gin.Options.I18n, c.GetHeader("x-request-lang"))
		missHeader, err := localizer.LocalizeMessage(&i18n.Message{ID: "missing_header"})
		if err != nil {
			gin.ErrorResponse(c, http.StatusInternalServerError, "Internal server error", "jwt-validate", err)
			return
		}
		l := c.GetHeader("x-request-lang")
		if l == "" {
			gin.ErrorResponse(c, http.StatusBadRequest, missHeader, "http-auth", nil)
			return
		}

		p := c.GetHeader("x-request-key")
		if p == "" {
			gin.ErrorResponse(c, http.StatusBadRequest, missHeader, "http-auth", nil)
			return
		}
	}
}

func (gin *Gin) ErrorResponse(c *g.Context, code int, msg string, iss string, err error) {
	if err != nil {
		gin.Options.Log.Error(iss + ": " + err.Error())
	}
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
		if gin.Options.AuthType == AuthTypeGrpc {
			gin.checkSessionFromGrpc(c, typ)
		}
		if gin.Options.AuthType == AuthTypeJwt {
			gin.checkSessionFromJwt(c, typ)
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
	c.JSON(http.StatusOK, gin.Options.Version)
}
