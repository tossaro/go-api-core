package gin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	g "github.com/gin-gonic/gin"
)

func (gin *Gin) checkSessionFromJwt(c *g.Context, typ string) {
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
