package module1

import (
	"fmt"
	"log"
	"net/http"

	g "github.com/gin-gonic/gin"
	"github.com/tossaro/go-api-core/gin"
)

type httpV1 struct {
	gin *gin.Gin
}

func NewHttpV1(args ...interface{}) {
	if len(args) == 0 {
		log.Fatal("missing module1 args")
	}

	var g *gin.Gin
	t := fmt.Sprintf("%T", args[0])
	log.Print(t)
	if t == "*gin.Gin" {
		g = args[0].(*gin.Gin)
	}

	m := &httpV1{g}
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
// @Success     200 {object} domain.UserV1
// @Failure     401 {object} gin.Error
// @Failure     417 {object} gin.Error
// @Failure     500 {object} gin.Error
// @Router      /module1/api1 [get]
func (*httpV1) api1(c *g.Context) {
	c.JSON(http.StatusOK, "API 1 Running")
}
