module github.com/xxnuo/MTranCore

go 1.25.0

require (
	github.com/Andrew-M-C/go.emoji v1.1.4
	github.com/fasthttp/websocket v1.5.12
	github.com/gofiber/fiber/v3 v3.0.0-rc.2
	github.com/jerbob92/wazero-emscripten-embind v1.5.2
	github.com/kiuber/gofiber3-contrib/websocket v0.1.1-0.20250623070125-7ef7e8d8a964
	github.com/soheilhy/cmux v0.1.5
	github.com/tetratelabs/wazero v1.9.0
	google.golang.org/grpc v1.75.1
	google.golang.org/protobuf v1.36.9
	sigs.k8s.io/yaml v1.6.0
)

require (
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/gofiber/schema v1.6.0 // indirect
	github.com/gofiber/utils/v2 v2.0.0-rc.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/savsgio/gotils v0.0.0-20250924091648-bce9a52d7761 // indirect
	github.com/tinylib/msgp v1.4.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.66.0 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	golang.org/x/crypto v0.42.0 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250929231259-57b25ae835d4 // indirect
)

retract (
	v0.1.2 // contains only retractions
	v0.1.1 // contains invalid module name
)
