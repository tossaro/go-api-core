package gin

import (
	"log"
	"net/http"

	g "github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	prm "github.com/prometheus/client_golang/prometheus/promhttp"
	sf "github.com/swaggo/files"
	gsw "github.com/swaggo/gin-swagger"
	pAuth "github.com/tossaro/go-api-core/auth/proto"
	"github.com/tossaro/go-api-core/captcha"
	j "github.com/tossaro/go-api-core/jwt"
	"github.com/tossaro/go-api-core/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	AuthTypeGrpc  = "grpc"
	AuthTypeRedis = "redis"
)

type (
	CKey string

	Error struct {
		Message string `json:"error" example:"message"`
	}

	Gin struct {
		Gin        *g.Engine
		Router     *g.RouterGroup
		Jwt        *j.Jwt
		AuthClient *pAuth.AuthServiceV1Client
		*Options
	}

	Options struct {
		Mode         string
		Version      string
		BaseUrl      string
		Log          logger.Interface
		AuthType     string
		AuthService  *string
		Redis        *redis.Client
		Jwt          *j.Jwt
		AccessToken  int
		RefreshToken int
		Captcha      *bool
	}

	TokenV1 struct {
		Access  string `json:"access"`
		Refresh string `json:"refresh"`
	}
)

func New(o *Options) *Gin {
	if o.Mode == "" {
		log.Fatal("gin - option mode not found")
	}
	if o.Version == "" {
		log.Fatal("gin - option version not found")
	}
	if o.BaseUrl == "" {
		log.Fatal("gin - option base url not found")
	}
	if o.Log == nil {
		log.Fatal("gin - option log not found")
	}
	if o.AuthType == "" {
		log.Fatal("gin - option auth type not found")
	}

	g.SetMode(o.Mode)
	gEngine := g.New()
	gEngine.Use(g.Logger())
	gEngine.Use(g.Recovery())

	var authClient *pAuth.AuthServiceV1Client
	if o.AuthService != nil {
		conn, err := grpc.Dial(*o.AuthService, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			o.Log.Error("Gin init auth error: %s", err)
		}
		defer conn.Close()

		c := pAuth.NewAuthServiceV1Client(conn)
		authClient = &c
	}

	gin := &Gin{gEngine, nil, o.Jwt, authClient, o}

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
		l := c.GetHeader("x-platform-lang")
		if l == "" {
			gin.ErrorResponse(c, http.StatusBadRequest, "missing header param", "http-auth", nil)
			return
		}

		p := c.GetHeader("x-request-key")
		if p == "" {
			gin.ErrorResponse(c, http.StatusBadRequest, "missing header param", "http-auth", nil)
			return
		}
	}
}

func (gin *Gin) ErrorResponse(c *g.Context, code int, msg string, iss string, err error) {
	if err != nil {
		gin.Options.Log.Error(iss, err)
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
			if gin.AuthClient == nil {
				gin.ErrorResponse(c, http.StatusInternalServerError, "Missing auth processor", "http-auth", nil)
				return
			}
			gin.checkSessionFromGrpc(c, typ)
		}
		if gin.Options.AuthType == AuthTypeRedis {
			if gin.Redis == nil || gin.Jwt == nil {
				gin.ErrorResponse(c, http.StatusInternalServerError, "Missing auth processor", "http-auth", nil)
				return
			}
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
