package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo-contrib/jaegertracing"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

func main() {
	// Echo instance
	e := echo.New()

	// e.Use(middleware.Logger())
	logger := logrus.New()
	logger.SetFormatter(new(logrus.JSONFormatter))

	// Enable tracing middleware
	// c := jaegertracing.New(e, nil)
	c := New(e, nil)
	defer c.Close()

	e.Use(LogRequestResponse(logger))

	// Routes
	e.GET("/name", nameHandler)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}

func New(e *echo.Echo, skipper middleware.Skipper) io.Closer {
	// Add Opentracing instrumentation
	defcfg := config.Configuration{
		ServiceName: "echo-tracer",
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
		},
	}
	cfg, err := defcfg.FromEnv()
	if err != nil {
		panic("Could not parse Jaeger env vars: " + err.Error())
	}
	// tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		panic("Could not initialize jaeger tracer: " + err.Error())
	}

	opentracing.SetGlobalTracer(tracer)
	e.Use(jaegertracing.TraceWithConfig(jaegertracing.TraceConfig{
		Tracer:     tracer,
		Skipper:    skipper,
		IsBodyDump: false,
	}))
	return closer
}

// Handler
func nameHandler(c echo.Context) error {
	url := "http://localhost:1324/namelist"

	span := opentracing.SpanFromContext(c.Request().Context())
	req, err := jaegertracing.NewTracedRequest("GET", url, nil, span)
	if err != nil {
		panic(err.Error())
	}

	resp, err := do(req)
	if err != nil {
		panic(err.Error())
	}
	namelist := string(resp)
	return c.String(http.StatusOK, randomName(namelist))
}

func randomName(namelist string) string {
	names := strings.Split(namelist, ",")
	return names[rand.Intn(3)]
}

// Route level middleware
func track(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		fmt.Printf("request to %v\n", c.Request().URL)
		return next(c)
	}
}

// Logrus : implement applog
type applog struct {
	logger *logrus.Logger
}

func (l *applog) handleLog(c echo.Context, reqBody, resBody []byte) {
	req := c.Request()
	res := c.Response()
	bytesIn := req.Header.Get(echo.HeaderContentLength)

	sp := opentracing.SpanFromContext(req.Context())
	sc := sp.Context().(jaeger.SpanContext)

	l.logger.WithFields(logrus.Fields{
		"timeRfc3339": time.Now().Format(time.RFC3339),
		"remoteIp":    c.RealIP(),
		"host":        req.Host,
		"reqMethod":   req.Method,
		"reqUri":      req.RequestURI,
		"reqPath":     req.URL.Path,
		"reqHeaders":  req.Header,
		"reqBody":     string(reqBody),
		"resStatus":   res.Status,
		"resHeaders":  res.Header(),
		"resBody":     string(resBody),
		"bytesIn":     bytesIn,
		"bytesOut":    strconv.FormatInt(res.Size, 10),
		"traceId":     sc.TraceID().String(),
		"spanId":      sc.SpanID().String(),
		"parentId":    sc.ParentID().String(),
	}).Info("Handled request")
}

// NewLogger func
func LogRequestResponse(logger *logrus.Logger) echo.MiddlewareFunc {
	l := &applog{logger: logger}
	return middleware.BodyDump(l.handleLog)
}

// Do executes an HTTP request and returns the response body.
// Any errors or non-200 status code result in an error.
func do(req *http.Request) ([]byte, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("StatusCode: %d, Body: %s", resp.StatusCode, body)
	}

	return body, nil
}
