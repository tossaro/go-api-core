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
		h.GET("/image/:id", a.captchaImage)
	}
}

// @Summary     Generate Captcha
// @Description Show Captcha Image to Secure
// @ID          captchaGenerateV1
// @Tags  	    captcha
// @Accept      json
// @Produce     json
// @Param		x-platform-lang header string true "Client Platform Lang" Enums(EN,ID)
// @Success     200 {string} 123456
// @Failure     204 {object} gin.Error
// @Failure     500 {object} gin.Error
// @Router      /captcha/v1/generate [get]
func (a *httpV1) captchaGenerateV1(c *g.Context) {
	c.JSON(http.StatusOK, captcha.NewLen(6))
}

// @Summary     Show Captcha Image
// @Description Show Captcha Image to Secure
// @ID          captchaImage
// @Tags  	    captcha
// @Accept      json
// @Produce     json
// @Param		x-platform-lang header string true "Client Platform Lang" Enums(EN,ID)
// @Param		id path string true "Captcha ID"
// @Success     200 "Show Captcha Image"
// @Failure     204 {object} gin.Error
// @Failure     500 {object} gin.Error
// @Router      /captcha/v1/image/{id} [get]
func (a *httpV1) captchaImage(c *g.Context) {
	id := c.Param("id")
	if id == "" {
		err := captcha.ErrNotFound
		a.g.ErrorResponse(c, http.StatusBadRequest, "captha error", "http-captchaImage", err)
		return
	}
	c.Set("Content-Type", "image/png")
	err := captcha.WriteImage(c.Writer, id, 120, 80)
	if err != nil {
		a.g.ErrorResponse(c, http.StatusBadRequest, "captha error", "http-captchaImage", err)
		return
	}
}
