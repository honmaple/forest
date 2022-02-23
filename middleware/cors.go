package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/honmaple/forest"
)

type (
	CorsConfig struct {
		Skipper          Skipper
		AllowOrigins     []string
		AllowMethods     []string
		AllowHeaders     []string
		AllowCredentials bool
		AllowOriginFunc  func(*http.Request, string) bool
		ExposeHeaders    []string
		MaxAge           time.Duration
	}
)

var (
	DefaultCorsConfig = CorsConfig{
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders: []string{"Origin", "Content-Length", "Content-Type"},
	}
)

// *.example.com match a.example.com b.example.com
func matchSubdomain(domain, origin string) bool {
	domainScheme := ""
	if index := strings.Index(domain, "://"); index > -1 {
		domainScheme = domain[:index]
		domain = domain[index+3:]
	}
	originScheme := ""
	if index := strings.Index(origin, "://"); index > -1 {
		originScheme = origin[:index]
		origin = origin[index+3:]
	}
	if domainScheme != originScheme {
		return false
	}
	domains := strings.Split(domain, ".")
	origins := strings.Split(origin, ".")
	if len(domains) != len(origins) {
		return false
	}
	for i, s := range domains {
		if s != "*" && origins[i] != s {
			return false
		}
	}
	return true
}

func Cors() forest.HandlerFunc {
	return CorsWithConfig(DefaultCorsConfig)
}

func CorsWithConfig(config CorsConfig) forest.HandlerFunc {
	normalHeaders := make(map[string]string)
	preflightHeaders := make(map[string]string)

	if len(config.ExposeHeaders) > 0 {
		normalHeaders["Access-Control-Expose-Headers"] = strings.Join(config.ExposeHeaders, ",")
	}
	if config.AllowCredentials {
		normalHeaders["Access-Control-Allow-Credentials"] = "true"
		preflightHeaders["Access-Control-Allow-Credentials"] = "true"
	}
	if len(config.AllowHeaders) > 0 {
		preflightHeaders["Access-Control-Allow-Headers"] = strings.Join(config.AllowHeaders, ",")
	}
	if len(config.AllowMethods) > 0 {
		preflightHeaders["Access-Control-Allow-Methods"] = strings.Join(config.AllowMethods, ",")
	}
	if config.MaxAge > 0 {
		preflightHeaders["Access-Control-Max-Age"] = config.MaxAge.String()
	}

	return func(c forest.Context) error {
		if config.Skipper != nil && config.Skipper(c) {
			return c.Next()
		}
		r := c.Request()
		w := c.Response()

		origin := r.Header.Get("Origin")
		allowedOrigin := ""

		if origin != "" {
			if matchSubdomain(r.Host, origin) {
				return c.Next()
			}
			if config.AllowOriginFunc != nil && config.AllowOriginFunc(r, origin) {
				allowedOrigin = origin
			} else {
				for _, o := range config.AllowOrigins {
					if o == "*" && config.AllowCredentials {
						allowedOrigin = origin
						break
					}
					if o == "*" || o == origin {
						allowedOrigin = o
						break
					}
					if matchSubdomain(o, origin) {
						allowedOrigin = origin
						break
					}
				}
			}
		}

		preflight := r.Method == http.MethodOptions
		if origin == "" || allowedOrigin == "" {
			if !preflight {
				return c.Next()
			}
			return c.Status(http.StatusNoContent)
		}

		header := w.Header()
		header.Add("Vary", "Origin")
		header.Set("Access-Control-Allow-Origin", allowedOrigin)

		if !preflight {
			for k, v := range normalHeaders {
				header.Set(k, v)
			}
			return c.Next()
		}

		header.Add("Vary", "Access-Control-Request-Method")
		header.Add("Vary", "Access-Control-Request-Headers")
		for k, v := range preflightHeaders {
			header.Set(k, v)
		}
		return c.Status(http.StatusNoContent)
	}
}
