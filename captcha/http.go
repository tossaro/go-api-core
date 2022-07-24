package captcha

import (
	"net/http"

	"github.com/dchest/captcha"
	g "github.com/gin-gonic/gin"

	"github.com/tossaro/go-api-core/gin"
)

type httpV1 struct {
	g *gin.Gin
}

func New(g *gin.Gin) {
	a := &httpV1{g}
	h := g.R.Group("/captcha/v1")
	{
		h.GET("/generate", a.captchaGenerateV1)
		h.GET("/image/:id", a.captchaImageV1)
	}
}

// @Summary     Generate Captcha
// @Description Show Captcha Image to Secure
// @ID          captchaGenerateV1
// @Tags  	    captcha
// @Accept      json
// @Produce     json
// @Param		x-platform-lang header string true "Client Platform Lang" Enums(EN,ID)
// @Param		x-request-key header string true "Request Key"
// @Success     200 {string} 1a2b3c4d5e
// @Router      /captcha/v1/generate [get]
func (a *httpV1) captchaGenerateV1(c *g.Context) {
	c.JSON(http.StatusOK, captcha.NewLen(6))
}

// @Summary     Show Captcha Image
// @Description Show Captcha Image to Secure
// @ID          captchaImageV1
// @Tags  	    captcha
// @Accept      json
// @Produce     json
// @Param		x-platform-lang header string true "Client Platform Lang" Enums(EN,ID)
// @Param		x-request-key header string true "Request Key"
// @Param		id path string true "Captcha ID"
// @Success     200 "Show Captcha Image"
// @Failure     204 {object} gin.Error
// @Failure     400 {object} gin.Error
// @Router      /captcha/v1/image/{id} [get]
func (a *httpV1) captchaImageV1(c *g.Context) {
	id := c.Param("id")
	if id == "" {
		err := captcha.ErrNotFound
		a.g.ErrorResponse(c, http.StatusBadRequest, "captha error", "http-captchaImageV1", err)
		return
	}
	c.Set("Content-Type", "image/png")
	err := captcha.WriteImage(c.Writer, id, 120, 80)
	if err != nil {
		a.g.ErrorResponse(c, http.StatusBadRequest, "captha error", "http-captchaImageV1", err)
		return
	}
}
