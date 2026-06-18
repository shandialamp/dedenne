package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	"github.com/shandialamp/dedenne/config"
	"github.com/shandialamp/dedenne/response"
)

// AuthClaims JWT claims
type AuthClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// JWTMiddleware JWT 认证中间件
func JWTMiddleware(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token, err := extractToken(c)
			if err != nil {
				resp := response.Response{
					Code:    int(10007), // CodeInvalidToken
					Data:    nil,
					Message: err.Error(),
					Error:   "missing or invalid authorization header",
				}
				return c.JSON(http.StatusUnauthorized, resp)
			}

			// 验证 token
			claims := &AuthClaims{}
			parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
				// 验证签名算法
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(secret), nil
			})

			if err != nil || !parsedToken.Valid {
				resp := response.Response{
					Code:    int(10007), // CodeInvalidToken
					Data:    nil,
					Message: "Invalid token",
					Error:   err.Error(),
				}
				return c.JSON(http.StatusUnauthorized, resp)
			}

			// 将用户信息存入 context
			c.Set("auth_user", claims)
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)

			return next(c)
		}
	}
}

// extractToken 从请求头中提取 token
func extractToken(c echo.Context) (string, error) {
	auth := c.Request().Header.Get("Authorization")
	if auth == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	parts := strings.Split(auth, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return parts[1], nil
}

// GetAuthUser 从 context 中获取认证用户信息
func GetAuthUser(c echo.Context) *AuthClaims {
	user, ok := c.Get("auth_user").(*AuthClaims)
	if !ok {
		return nil
	}
	return user
}

// GetUserID 从 context 中获取用户 ID
func GetUserID(c echo.Context) int64 {
	userID, ok := c.Get("user_id").(int64)
	if !ok {
		return 0
	}
	return userID
}

// GetUsername 从 context 中获取用户名
func GetUsername(c echo.Context) string {
	username, ok := c.Get("username").(string)
	if !ok {
		return ""
	}
	return username
}

// GenerateToken 生成 JWT token
func GenerateToken(userID int64, username string) (string, error) {
	cfg := config.Get()
	
	claims := &AuthClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			// 设置过期时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(
				time.Duration(cfg.JWT.Expiration) * time.Second,
			)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWT.Secret))
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return tokenString, nil
}
