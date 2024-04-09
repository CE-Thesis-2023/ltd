module github.com/CE-Thesis-2023/ltd

go 1.21.4

replace github.com/CE-Thesis-2023/backend => ../backend

require (
	github.com/CE-Thesis-2023/backend v1.0.0
	github.com/eclipse/paho.golang v0.12.0
)

require (
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/rogpeppe/go-internal v1.10.0 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

require (
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.26.0
	gopkg.in/yaml.v3 v3.0.1
)
