docker run ^
  --rm ^
  --env JAEGER_AGENT_HOST=10.254.0.119 ^
  --env JAEGER_AGENT_PORT=6831 ^
  -p8080-8083:8080-8083 ^
  jaegertracing/example-hotrod:latest ^
  all