package middleware

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/honmaple/forest"
)

const (
	greenColor  = "\033[0;32m"
	cyanColor   = "\033[0;36m"
	yellowColor = "\033[0;33m"
	redColor    = "\033[0;31m"
	resetColor  = "\033[0m"
)

type (
	LoggerConfig struct {
		Skipper   Skipper
		Output    io.Writer
		Formatter func() LoggerFormatter
	}
	LoggerFormatter interface {
		Reset()
		Format(*http.Request, http.ResponseWriter, int) string
	}
	loggerFormatter struct {
		start time.Time
	}
)

var (
	DefaultLoggerConfig = LoggerConfig{
		Output:    os.Stdout,
		Formatter: newLoggerFormatter,
	}
)

func newLoggerFormatter() LoggerFormatter {
	return &loggerFormatter{start: time.Now()}
}

func (f *loggerFormatter) Reset() {
	f.start = time.Now()
}

func (f *loggerFormatter) Format(req *http.Request, resp http.ResponseWriter, status int) string {
	statusColor := greenColor
	if status >= 300 && status < 400 {
		statusColor = cyanColor
	} else if status >= 400 && status < 500 {
		statusColor = yellowColor
	} else if status >= 500 {
		statusColor = redColor
	}

	end := time.Now()
	return fmt.Sprintf("%s - [%s] \"%s %s%s %s\" %s%03d%s - %s\n",
		req.RemoteAddr,
		f.start.Format("02/Jan/2006 15:04:05.00000"),
		req.Method, req.Host, req.RequestURI, req.Proto,
		statusColor, status, resetColor,
		end.Sub(f.start).String(),
	)
}

func Logger() forest.HandlerFunc {
	return LoggerWithConfig(DefaultLoggerConfig)
}

func LoggerWithConfig(config LoggerConfig) forest.HandlerFunc {
	if config.Output == nil {
		config.Output = DefaultLoggerConfig.Output
	}
	if config.Formatter == nil {
		config.Formatter = DefaultLoggerConfig.Formatter
	}
	f := config.Formatter()
	return func(c forest.Context) error {
		if config.Skipper != nil && config.Skipper(c) {
			return c.Next()
		}
		f.Reset()

		req := c.Request()
		err := c.Next()
		resp := c.Response()
		fmt.Fprint(config.Output, f.Format(req, resp, resp.Status))
		return err
	}
}
