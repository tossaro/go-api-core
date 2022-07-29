package module1

import (
	"fmt"
	"log"
	"net/http"

	g "github.com/gin-gonic/gin"
	"github.com/tossaro/go-api-core/gin"
	"github.com/tossaro/go-api-core/postgres"
)

type httpV1 struct {
	gin *gin.Gin
	pg  *postgres.Postgres
}

func NewHttpV1(args []interface{}) {
	if len(args) < 2 {
		log.Fatal("auth httpv1 args must be 2 but got: ", len(args))
	}

	var g *gin.Gin
	if fmt.Sprintf("%T", args[0]) == "*gin.Gin" {
		g = args[0].(*gin.Gin)
	}

	var pg *postgres.Postgres
	if fmt.Sprintf("%T", args[0]) == "*postgres.Postgres" {
		pg = args[0].(*postgres.Postgres)
	}

	m := &httpV1{g, pg}
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
func (*httpV1) api1(c *g.Context) {
	c.JSON(http.StatusOK, "API 1 Running")
}
