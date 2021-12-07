package main

import (
	"fmt"
	"os"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

func main() {
	if len(os.Args) != 2 {
		panic("ERROR: Expecting one argument")
	}
	helloTo := os.Args[1]

	// Create Jaeger Tracer
	// cfg := &config.Configuration{
	// 	ServiceName: "king-tracingdemo",
	// 	Sampler: &config.SamplerConfig{
	// 		Type:  "const",
	// 		Param: 1,
	// 	},
	// 	Reporter: &config.ReporterConfig{
	// 		LogSpans: true,
	// 	},
	// }

	cfg := &config.Configuration{
		ServiceName: "king-tracingdemo",
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans: true,
		},
	}
	fmt.Printf("1st cfg = %+v", cfg)

	// Override configurations by ENV
	cfg, err := cfg.FromEnv()
	if err != nil {
		fmt.Errorf("error = %+v", err)
	}
	fmt.Printf("2nd cfg = %+v", cfg)

	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	defer closer.Close()

	// Start Span
	span := tracer.StartSpan("say-hello")
	span.SetTag("hello-to", helloTo)
	defer span.Finish()

	// Start child span, operation name = formatString
	spanFormatString := tracer.StartSpan("formatString", opentracing.ChildOf(span.Context()))
	helloStr := fmt.Sprintf("Hello, %s!", helloTo)
	spanFormatString.LogFields(
		log.String("event", "string-format"),
		log.String("helloStr", helloStr),
	)
	spanFormatString.Finish()

	// Start another child span, operation name = printHello
	spanPrintHello := tracer.StartSpan("printHello", opentracing.ChildOf(span.Context()))
	println(helloStr)
	for i := 0; i < 10000; i++ {
	}
	spanPrintHello.LogKV("event", "println")
	spanPrintHello.Finish()
}
