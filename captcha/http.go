package captcha

import (
	"net/http"

	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"
	"github.com/tossaro/go-api-core/logger"
)

type (
	httpV1 struct {
		r *gin.RouterGroup
		l logger.Interface
	}

	Error struct {
		Msg string `json:"error" example:"message"`
	}
)

func New(r *gin.RouterGroup, l logger.Interface) {
	a := &httpV1{r, l}
	rc := r.Group("/captcha/v1")
	{
		rc.GET("/generate", a.captchaGenerateV1)
		rc.GET("/image/:id", a.captchaImageV1)
	}
}

func (a *httpV1) ErrorResponse(c *gin.Context, code int, msg string, iss string, err error) {
	if err != nil {
		a.l.Error(err, iss)
	}
	c.AbortWithStatusJSON(code, &Error{msg})
}

// @Summary     Generate Captcha
// @Description Show Captcha Image to Secure
// @ID          captchaGenerateV1
// @Tags  	    Captcha
// @Accept      json
// @Produce     json
// @Success     200 {string} 1a2b3c4d5e
// @Router      /captcha/v1/generate [get]
func (a *httpV1) captchaGenerateV1(c *gin.Context) {
	c.JSON(http.StatusOK, captcha.NewLen(6))
}

// @Summary     Show Captcha Image
// @Description Show Captcha Image to Secure
// @ID          captchaImageV1
// @Tags  	    Captcha
// @Accept      json
// @Produce     json
// @Param		id path string true "Captcha ID"
// @Success     200 "Show Captcha Image"
// @Failure     204 {string} Message
// @Failure     400 {string} Message
// @Router      /captcha/v1/image/{id} [get]
func (a *httpV1) captchaImageV1(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		err := captcha.ErrNotFound
		a.ErrorResponse(c, http.StatusBadRequest, "captha error", "http-captchaImageV1", err)
		return
	}
	c.Set("Content-Type", "image/png")
	err := captcha.WriteImage(c.Writer, id, 120, 80)
	if err != nil {
		a.ErrorResponse(c, http.StatusBadRequest, "captha error", "http-captchaImageV1", err)
		return
	}
}
