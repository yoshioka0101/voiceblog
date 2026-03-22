package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func configureGinLogging(log *slog.Logger) {
	gin.DefaultWriter = &ginLogWriter{
		log:   log,
		level: slog.LevelDebug,
		event: "gin debug",
	}
	gin.DefaultErrorWriter = &ginLogWriter{
		log:   log,
		level: slog.LevelError,
		event: "gin error",
	}
	gin.DebugPrintFunc = func(format string, values ...any) {
		msg := strings.TrimSpace(fmt.Sprintf(format, values...))
		if msg == "" {
			return
		}

		if strings.Contains(msg, `Running in "debug" mode`) {
			log.Warn("gin debug mode enabled",
				slog.String("gin_mode", gin.Mode()),
			)
			return
		}

		log.Debug("gin debug",
			slog.String("message", msg),
		)
	}
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, numHandlers int) {
		log.Debug("gin route registered",
			slog.String("method", httpMethod),
			slog.String("path", absolutePath),
			slog.String("handler", handlerName),
			slog.Int("handlers", numHandlers),
		)
	}
}

func requestLogger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		status := c.Writer.Status()
		level := slog.LevelInfo
		switch {
		case status >= http.StatusInternalServerError:
			level = slog.LevelError
		case status >= http.StatusBadRequest:
			level = slog.LevelWarn
		}

		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		attrs := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("route", route),
			slog.Int("status", status),
			slog.Int64("latency_ms", time.Since(start).Milliseconds()),
			slog.String("client_ip", c.ClientIP()),
		}
		if size := c.Writer.Size(); size >= 0 {
			attrs = append(attrs, slog.Int("response_bytes", size))
		}
		if userAgent := c.Request.UserAgent(); userAgent != "" {
			attrs = append(attrs, slog.String("user_agent", userAgent))
		}
		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("errors", c.Errors.String()))
		}

		logWithLevel(log, level, "request completed", attrs...)
	}
}

func recoveryLogger(log *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, recovered any) {
		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		attrs := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("route", route),
			slog.String("client_ip", c.ClientIP()),
			slog.String("panic_type", fmt.Sprintf("%T", recovered)),
			slog.String("panic", fmt.Sprint(recovered)),
		}
		if userAgent := c.Request.UserAgent(); userAgent != "" {
			attrs = append(attrs, slog.String("user_agent", userAgent))
		}

		var err error
		if recoveredErr, ok := recovered.(error); ok {
			err = recoveredErr
		}
		if isBrokenPipeError(err) {
			logWithLevel(log, slog.LevelWarn, "client connection error", attrs...)
			if err != nil {
				_ = c.Error(err)
			}
			c.Abort()
			return
		}

		attrs = append(attrs,
			slog.Int("status", http.StatusInternalServerError),
			slog.String("stack", string(debug.Stack())),
		)
		logWithLevel(log, slog.LevelError, "panic recovered", attrs...)
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

func logWithLevel(log *slog.Logger, level slog.Level, msg string, attrs ...slog.Attr) {
	log.LogAttrs(context.Background(), level, msg, attrs...)
}

func isBrokenPipeError(err error) bool {
	if err == nil {
		return false
	}

	var netErr *net.OpError
	if !errors.As(err, &netErr) {
		return false
	}

	var syscallErr *os.SyscallError
	if !errors.As(netErr, &syscallErr) {
		return false
	}

	errMessage := strings.ToLower(syscallErr.Error())
	return strings.Contains(errMessage, "broken pipe") || strings.Contains(errMessage, "connection reset by peer")
}

type ginLogWriter struct {
	log   *slog.Logger
	level slog.Level
	event string
}

func (w *ginLogWriter) Write(p []byte) (int, error) {
	msg := strings.TrimSpace(string(p))
	if msg == "" {
		return len(p), nil
	}

	logWithLevel(w.log, w.level, w.event, slog.String("message", msg))
	return len(p), nil
}
