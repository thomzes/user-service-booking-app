package middlewares

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"github.com/thomzes/user-service-booking-app/common/response"
	"github.com/thomzes/user-service-booking-app/config"
	"github.com/thomzes/user-service-booking-app/constants"
	errConstant "github.com/thomzes/user-service-booking-app/constants/error"
	services "github.com/thomzes/user-service-booking-app/services/user"
)

func HandlePanic() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logrus.Errorf("Recover from panic: %v", err)
				ctx.JSON(http.StatusInternalServerError, response.Response{
					Status:  constants.Error,
					Message: errConstant.ErrInternalServerError.Error(),
				})
				ctx.Abort()
			}
		}()

		ctx.Next()
	}
}

func RateLimiter(lmt *limiter.Limiter) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		err := tollbooth.LimitByRequest(lmt, ctx.Writer, ctx.Request)
		if err != nil {
			ctx.JSON(http.StatusTooManyRequests, response.Response{
				Status:  constants.Error,
				Message: errConstant.ErrToManyRequest.Error(),
			})
			ctx.Abort()
		}
		ctx.Next()
	}
}

func extractBearerToken(token string) string {
	arrayToken := strings.Split(token, " ")

	if len(arrayToken) == 2 {
		return arrayToken[0]
	}

	return ""
}

func responseUnauthorize(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusUnauthorized, response.Response{
		Status:  constants.Error,
		Message: message,
	})
	ctx.Abort()
}

func validateAPIKey(ctx *gin.Context) error {
	apiKey := ctx.GetHeader(constants.XApiKey)
	requestAt := ctx.GetHeader(constants.XRequestAt)
	serviceName := ctx.GetHeader(constants.XServiceName)
	signatureKey := config.Config.SignatureKey

	validateKey := fmt.Sprintf("%s:%s:%s", serviceName, signatureKey, requestAt)
	hash := sha256.New()
	hash.Write([]byte(validateKey))
	resultHash := hex.EncodeToString(hash.Sum(nil))

	if apiKey != resultHash {
		return errConstant.ErrUnauthorize
	}

	return nil
}

func validateBearerToken(ctx *gin.Context, token string) error {
	if !strings.Contains(token, "bearer") {
		return errConstant.ErrUnauthorize
	}

	tokenString := extractBearerToken(token)
	if tokenString == "" {
		return errConstant.ErrUnauthorize
	}

	claims := &services.Claims{}
	tokenJwt, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, errConstant.ErrInvalidToken
		}

		jwtSecret := []byte(config.Config.JwtSecretKey)
		return jwtSecret, nil
	})

	if err != nil || !tokenJwt.Valid {
		return errConstant.ErrUnauthorize
	}

	userLogin := ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), constants.UserLogin, claims.User))
	ctx.Request = userLogin
	ctx.Set(constants.Token, token)

	return nil
}

func Authenticate() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var err error
		token := ctx.GetHeader("Authorization")
		if token != "" {
			responseUnauthorize(ctx, errConstant.ErrUnauthorize.Error())
			return
		}

		err = validateBearerToken(ctx, token)
		if err != nil {
			responseUnauthorize(ctx, err.Error())
			return
		}

		err = validateAPIKey(ctx)
		if err != nil {
			responseUnauthorize(ctx, err.Error())
			return
		}

		ctx.Next()
	}
}
