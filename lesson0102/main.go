package main

import (
	"context"
	"fmt"
	"os"
	"tracingdemo/tracing"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

func main() {
	if len(os.Args) != 2 {
		panic("ERROR: Expecting one argument")
	}
	helloTo := os.Args[1]

	tracer, closer := tracing.Init("hello-world")
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	span := tracer.StartSpan("say-hello")
	span.SetTag("hello-to", helloTo)
	defer span.Finish()

	ctx := context.Background()
	spanCtx := opentracing.ContextWithSpan(ctx, span)

	helloStr := formatString(spanCtx, helloTo)
	printHello(spanCtx, helloStr)
}

func formatString(ctx context.Context, helloTo string) string {
	// span := rootSpan.Tracer().StartSpan("formatString", opentracing.ChildOf(rootSpan.Context()))
	span, _ := opentracing.StartSpanFromContext(ctx, "formatString")
	defer span.Finish()

	helloStr := fmt.Sprintf("Hello, %s!", helloTo)
	span.LogFields(
		log.String("event", "string-format"),
		log.String("helloStr", helloStr),
	)

	return helloStr
}

func printHello(ctx context.Context, helloStr string) {
	// span := rootSpan.Tracer().StartSpan("printHello", opentracing.ChildOf(rootSpan.Context()))
	span, _ := opentracing.StartSpanFromContext(ctx, "printHello")
	defer span.Finish()

	println(helloStr)
	span.LogKV("event", "println")
}
