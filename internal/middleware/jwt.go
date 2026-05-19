package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/imkerbos/mxcmdb/internal/pkg/response"
)

// JWTAuth JWT 认证中间件
func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := ""
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenStr = parts[1]
			}
		}
		// WebSocket 不支持自定义 Header，从 query param 获取 token
		if tokenStr == "" {
			tokenStr = c.Query("token")
		}
		if tokenStr == "" {
			response.Error(c, http.StatusUnauthorized, 40100, "缺少认证信息")
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			response.Error(c, http.StatusUnauthorized, 40101, "Token 无效或已过期")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.Error(c, http.StatusUnauthorized, 40101, "Token 解析失败")
			c.Abort()
			return
		}

		c.Set("user_id", uint(claims["user_id"].(float64)))
		c.Set("username", claims["username"].(string))
		c.Set("role", claims["role"].(string))
		c.Next()
	}
}
