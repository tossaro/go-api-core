package module1

import (
	"net/http"

	g "github.com/gin-gonic/gin"
	"github.com/tossaro/go-api-core/gin"
	"github.com/tossaro/go-api-core/postgres"
	"github.com/tossaro/go-api-core/twilio"
)

type httpV1 struct {
	gin *gin.Gin
	pg  *postgres.Postgres
	twl *twilio.Twilio
}

func NewHttpV1(g *gin.Gin, pg *postgres.Postgres, twl *twilio.Twilio) {
	m := &httpV1{g, pg, twl}
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
