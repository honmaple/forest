package middleware

import (
	"time"

	"bytes"
	"github.com/honmaple/forest"
	"github.com/valyala/fasttemplate"
	"io"
	"strconv"
	"sync"
)

type (
	LoggerConfig struct {
		Format     string
		TimeFormat string
		template   *fasttemplate.Template
		pool       sync.Pool
	}
)

var DefaultLoggerConfig = &LoggerConfig{
	Format:     "[${time_local}] ${status} ${method} ${path} (${remote_addr}) ${latency_human}",
	TimeFormat: "2006-01-02 15:04:05.00000",
}

func LoggerWithConfig(config *LoggerConfig) forest.HandlerFunc {
	if config.Format == "" {
		config.Format = DefaultLoggerConfig.Format
	}
	if config.TimeFormat == "" {
		config.TimeFormat = DefaultLoggerConfig.TimeFormat
	}
	config.template = fasttemplate.New(config.Format, "${", "}")
	config.pool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 256))
		},
	}
	return func(c forest.Context) error {
		start := time.Now()
		req := c.Request()
		err := c.Next()
		res := c.Response()
		stop := time.Now()

		buf := config.pool.Get().(*bytes.Buffer)
		buf.Reset()
		defer config.pool.Put(buf)

		if _, err = config.template.ExecuteFunc(buf, func(w io.Writer, tag string) (int, error) {
			switch tag {
			case "remote_addr":
				return buf.WriteString(req.RemoteAddr)
			case "time_local":
				return buf.WriteString(time.Now().Format(config.TimeFormat))
			case "host":
				return buf.WriteString(req.Host)
			case "uri":
				return buf.WriteString(req.RequestURI)
			case "method":
				return buf.WriteString(req.Method)
			case "path":
				p := req.URL.Path
				if p == "" {
					p = "/"
				}
				return buf.WriteString(p)
			case "protocol":
				return buf.WriteString(req.Proto)
			case "referer":
				return buf.WriteString(req.Referer())
			case "user_agent":
				return buf.WriteString(req.UserAgent())
			case "status":
				return buf.WriteString(strconv.Itoa(res.Status))
			case "latency":
				l := stop.Sub(start)
				return buf.WriteString(strconv.FormatInt(int64(l), 10))
			case "latency_human":
				return buf.WriteString(stop.Sub(start).String())
			default:
			}
			return 0, nil
		}); err != nil {
			return err
		}
		c.Logger().Println(buf.String())
		return nil
	}
}

func Logger() forest.HandlerFunc {
	return LoggerWithConfig(DefaultLoggerConfig)
}
