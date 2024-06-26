module github.com/CE-Thesis-2023/ltd

go 1.21.4

replace github.com/CE-Thesis-2023/backend => ../backend

require (
	github.com/CE-Thesis-2023/backend v1.0.0
	github.com/eclipse/paho.golang v0.12.0
	gonum.org/v1/gonum v0.15.0
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/pgvector/pgvector-go v0.1.1 // indirect
	github.com/rogpeppe/go-internal v1.10.0 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gorm.io/gorm v1.25.7-0.20240204074919-46816ad31dde // indirect
)

require (
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0
	gopkg.in/yaml.v3 v3.0.1
)
