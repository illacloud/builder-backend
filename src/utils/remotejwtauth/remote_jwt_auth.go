package remotejwtauth

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/illacloud/builder-backend/src/utils/config"
	"github.com/illacloud/builder-backend/src/utils/supervisor"
)

type AuthClaims struct {
	User   int       `json:"user"`
	UUID   uuid.UUID `json:"uuid"`
	Random string    `json:"rnd"`
	jwt.RegisteredClaims
}

func RemoteJWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// fetch content
		accessToken := c.Request.Header["Authorization"]
		var token string
		if len(accessToken) != 1 {
			c.AbortWithStatus(http.StatusUnauthorized)
		} else {
			token = accessToken[0]
		}

		sv := supervisor.NewSupervisor()

		validated, errInValidate := sv.ValidateUserAccount(token)
		fmt.Printf("token: %v\n", token)
		fmt.Printf("errInValidate: %v\n", errInValidate)
		if errInValidate != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			c.Next()
		}
		if !validated {
			c.AbortWithStatus(http.StatusUnauthorized)
			c.Next()
		}
		// ok set userID
		userID, userUID, errInExtractUserID := ExtractUserIDFromToken(token)
		if errInExtractUserID != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			c.Next()
		}
		c.Set("userID", userID)
		c.Set("userUID", userUID)
		c.Next()
	}
}

func ExtractUserIDFromToken(accessToken string) (int, uuid.UUID, error) {

	authClaims := &AuthClaims{}
	token, err := jwt.ParseWithClaims(accessToken, authClaims, func(token *jwt.Token) (interface{}, error) {
		conf := config.GetInstance()
		return []byte(conf.GetSecretKey()), nil
	})
	if err != nil {
		return 0, uuid.Nil, err
	}

	claims, ok := token.Claims.(*AuthClaims)
	if !(ok && token.Valid) {
		return 0, uuid.Nil, err
	}

	return claims.User, claims.UUID, nil
}
