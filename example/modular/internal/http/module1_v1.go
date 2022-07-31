package http

import (
	"fmt"
	"log"
	"net/http"

	g "github.com/gin-gonic/gin"
	"github.com/tossaro/go-api-core/config"
	"github.com/tossaro/go-api-core/gin"
	"github.com/tossaro/go-api-core/postgres"
)

type module1V1 struct {
	gin *gin.Gin
	cfg config.Config
	pg  *postgres.Postgres
}

func NewModule1V1(args []interface{}) {
	var g *gin.Gin
	var cfg config.Config
	var pg *postgres.Postgres
	for _, a := range args {
		switch fmt.Sprintf("%T", a) {
		case "*gin.Gin":
			g = a.(*gin.Gin)
		case "config.Config":
			cfg = a.(config.Config)
		case "*postgres.Postgres":
			pg = a.(*postgres.Postgres)
		}
	}

	if g == nil || cfg.App.Name == "" || pg == nil {
		log.Fatal("Auth args incomplete: ", g, cfg, pg)
	}

	m := &module1V1{g, cfg, pg}
	h := g.Router.Group("module1")
	{
		h.GET("api1", m.api1)
	}
}

// @Summary     API 1
// @Description Provide API 1
// @ID          api1
// @Tags  	    Module 1
// @Accept      json
// @Produce     json
// @Param		x-request-lang header string true "Client Request Lang" Enums(EN,ID)
// @Param		x-request-key header string true "Request Key"
// @Success     200 {string} Status
// @Failure     401 {object} gin.Error
// @Failure     417 {object} gin.Error
// @Failure     500 {object} gin.Error
// @Router      /module1/api1 [get]
func (m *module1V1) api1(c *g.Context) {
	c.JSON(http.StatusOK, "API 1 Running")
}
