docker run --rm --name jaeger ^
  -e SPAN_STORAGE_TYPE=badger ^
  -e BADGER_EPHEMERAL=false ^
  -e BADGER_DIRECTORY_VALUE=/badger/data ^
  -e BADGER_DIRECTORY_KEY=/badger/key ^
  -v C:\Users\Freshket\Projects\tracingdemo\docker_volume\badger:/badger ^
  -p 6831:6831/udp ^
  -p 6832:6832/udp ^
  -p 5778:5778 ^
  -p 16686:16686 ^
  -p 14268:14268 ^
  -p 14250:14250 ^
  jaegertracing/all-in-one:1.27 ^
  --log-level=debug