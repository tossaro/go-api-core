package http

import (
	"net/http"

	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"
	"github.com/tossaro/go-api-core/logger"
)

type captchaV1 struct {
	r *gin.RouterGroup
	l logger.Interface
}

type Error struct {
	Msg string `json:"error" example:"message"`
}

func NewCaptchaV1(r *gin.RouterGroup, l logger.Interface) {
	a := &captchaV1{r, l}
	rc := r.Group("/captcha/v1")
	{
		rc.GET("/generate", a.generate)
		rc.GET("/image/:id", a.image)
	}
}

func (a *captchaV1) ErrorResponse(c *gin.Context, code int, msg string) {
	c.AbortWithStatusJSON(code, &Error{msg})
}

// @Summary     Generate Captcha
// @Description Show Captcha Image to Secure
// @ID          captchaV1Generate
// @Tags  	    Captcha
// @Accept      json
// @Produce     json
// @Success     200 {string} 1a2b3c4d5e
// @Router      /captcha/v1/generate [get]
func (a *captchaV1) generate(c *gin.Context) {
	c.JSON(http.StatusOK, captcha.NewLen(6))
}

// @Summary     Show Captcha Image
// @Description Show Captcha Image to Secure
// @ID          captchaV1Image
// @Tags  	    Captcha
// @Accept      json
// @Produce     json
// @Param		id path string true "Captcha ID"
// @Success     200 "Show Captcha Image"
// @Failure     204 {string} Message
// @Failure     400 {string} Message
// @Router      /captcha/v1/image/{id} [get]
func (a *captchaV1) image(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		err := captcha.ErrNotFound
		a.l.Error("captchaV1-image", err)
		a.ErrorResponse(c, http.StatusBadRequest, "captha error")
		return
	}
	c.Set("Content-Type", "image/png")
	err := captcha.WriteImage(c.Writer, id, 120, 80)
	if err != nil {
		a.l.Error("captchaV1-image", err)
		a.ErrorResponse(c, http.StatusBadRequest, "captha error")
		return
	}
}
