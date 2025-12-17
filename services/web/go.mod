module github.com/microserviceboilerplate/web

go 1.22

require (
	github.com/go-chi/chi/v5 v5.1.0
	github.com/gorilla/sessions v1.2.2
	github.com/nats-io/nats.go v1.38.0
	github.com/nikolapavicevic-001/CommonGo v0.0.0
	github.com/rs/zerolog v1.33.0
	golang.org/x/oauth2 v0.15.0
)

require (
	github.com/go-chi/cors v1.2.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/nats-io/nkeys v0.4.9 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)

replace github.com/gorilla/mux => github.com/go-chi/chi/v5 v5.1.0

replace github.com/nikolapavicevic-001/CommonGo => ../../../CommonGo
