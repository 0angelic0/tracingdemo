package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo-contrib/jaegertracing"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
)

func main() {
	// Echo instance
	e := echo.New()

	// Enable tracing middleware
	// c := jaegertracing.New(e, nil)
	c := New(e, nil)
	defer c.Close()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", hello)
	e.GET("/users/:id", getUser)
	e.GET("/users", getUsers, track)

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
	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		panic("Could not initialize jaeger tracer: " + err.Error())
	}

	opentracing.SetGlobalTracer(tracer)
	e.Use(jaegertracing.TraceWithConfig(jaegertracing.TraceConfig{
		Tracer:     tracer,
		Skipper:    skipper,
		IsBodyDump: true,
	}))
	return closer
}

// Handler
func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func getUser(c echo.Context) error {
	// User ID from path `users/:id`
	id := c.Param("id")
	return c.String(http.StatusOK, id)
}

func getUsers(c echo.Context) error {
	return c.String(http.StatusOK, "/users")
}

// Route level middleware
func track(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		fmt.Println("request to /users")
		return next(c)
	}
}
