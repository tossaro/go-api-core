package gin

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

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
		Logger       logger.Interface
		AuthService  *string
		Redis        *redis.Client
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
	g.SetMode(o.Mode)
	gEngine := g.New()
	gEngine.Use(g.Logger())
	gEngine.Use(g.Recovery())

	var jwt *j.Jwt
	if o.Redis != nil {
		jNew, err := j.New(o.AccessToken, o.RefreshToken)
		if err != nil {
			log.Printf("JWT error: %s", err)
		}
		jwt = jNew
	}

	var authClient *pAuth.AuthServiceV1Client
	if o.AuthService != nil {
		conn, err := grpc.Dial(*o.AuthService, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("GRPC error: %s", err)
		}
		defer conn.Close()

		c := pAuth.NewAuthServiceV1Client(conn)
		authClient = &c
	}

	gin := &Gin{gEngine, nil, jwt, authClient, o}

	gRouter := gEngine.Group(o.BaseUrl)
	{
		gRouter.GET("/version", gin.version)
		gRouter.GET("/metrics", g.WrapH(prm.Handler()))
		gRouter.GET("/swagger/*any", gsw.DisablingWrapHandler(sf.Handler, "HTTP_SWAGGER_DISABLED"))

		if o.Captcha != nil && *(o.Captcha) {
			captcha.New(gRouter, o.Logger)
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

func (gin *Gin) ErrorResponse(c *g.Context, code int, msg string, iss string, err error) {
	if err != nil {
		gin.Options.Logger.Error(err, iss)
	}
	c.AbortWithStatusJSON(code, &Error{msg})
}

func (gin *Gin) AuthAccessMiddleware() g.HandlerFunc {
	if gin.Options.Redis != nil {
		return gin.checkSessionFromJwt("access")
	}
	return gin.checkSessionFromService("access")
}

func (gin *Gin) AuthRefreshMiddleware() g.HandlerFunc {
	if gin.Options.Redis != nil {
		return gin.checkSessionFromJwt("refresh")
	}
	return gin.checkSessionFromService("refresh")
}

func (gin *Gin) checkSessionFromService(typ string) g.HandlerFunc {
	return func(c *g.Context) {
		if gin.Options.AuthService == nil {
			gin.ErrorResponse(c, http.StatusUnauthorized, "Authentication Error", "http-auth", nil)
			return
		}

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
}

func (gin *Gin) checkSessionFromJwt(typ string) g.HandlerFunc {
	return func(c *g.Context) {
		if gin.Options.Redis == nil {
			gin.ErrorResponse(c, http.StatusUnauthorized, "Authentication Error", "http-auth", nil)
			return
		}

		ah := c.GetHeader("Authorization")
		sa := strings.Split(ah, " ")
		if len(sa) != 2 {
			gin.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "http-auth", nil)
			return
		}
		claims, err := gin.Jwt.Validate(sa[1])
		if err != nil || typ != claims.Type {
			gin.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "http-auth", err)
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

func (gin *Gin) CreateSessionJwt(uid uint64, iss string) (TokenV1, error) {
	var t TokenV1
	ac, err := gin.Jwt.AccessToken(uid, iss)
	if err != nil {
		return t, err
	}

	rf, k, err := gin.Jwt.RefreshToken(uid, iss)
	if err != nil {
		return t, err
	}

	err = gin.Redis.Set(context.Background(), *(k), "0", 0).Err()
	if err != nil {
		return t, err
	}

	return TokenV1{Access: *(ac), Refresh: *(rf)}, nil
}

func (gin *Gin) RefreshSessionJwt(uid uint64, key string, req string) (TokenV1, error) {
	var t TokenV1
	v, _ := gin.Redis.Get(context.Background(), key).Result()
	if v != "0" && v != req {
		return t, fmt.Errorf(key)
	}

	if v == req {
		nt, err := gin.Redis.Get(context.Background(), key+"_issued").Result()
		if err != nil {
			return t, err
		}

		err = json.Unmarshal([]byte(nt), &t)
		if err != nil {
			return t, err
		}

		return t, nil
	}

	err := gin.Redis.Set(context.Background(), key, req, 0).Err()
	if err != nil {
		return t, err
	}

	ses, err := gin.CreateSessionJwt(uid, req)
	if err != nil {
		return t, err
	}

	jses, err := json.Marshal(ses)
	if err != nil {
		return t, err
	}

	err = gin.Redis.Set(context.Background(), key+"_issued", jses, 0).Err()
	if err != nil {
		return t, err
	}

	return ses, nil
}
