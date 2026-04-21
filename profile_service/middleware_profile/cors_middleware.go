package middleware_profile

import (
	"github.com/gin-gonic/gin"
)

type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
}

// NewCORS возвращает middleware для обработки CORS
func NewCORS(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Проверяем, разрешён ли origin
		allowedOrigin := ""
		for _, o := range config.AllowOrigins {
			if o == "*" {
				allowedOrigin = "*"
				break
			}
			if o == origin {
				allowedOrigin = origin
				break
			}
		}

		if allowedOrigin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			c.Writer.Header().Set("Access-Control-Allow-Methods", joinStrings(config.AllowMethods))
			c.Writer.Header().Set("Access-Control-Allow-Headers", joinStrings(config.AllowHeaders))
			c.Writer.Header().Set("Access-Control-Expose-Headers", joinStrings(config.ExposeHeaders))
			if config.AllowCredentials {
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}

		// Preflight запросы обрабатываем сразу
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func joinStrings(strs []string) string {
	res := ""
	for i, s := range strs {
		if i > 0 {
			res += ", "
		}
		res += s
	}
	return res
}
