module github.com/microserviceboilerplate/web

go 1.21

require (
	github.com/go-chi/chi/v5 v5.2.3
	github.com/gorilla/sessions v1.2.2
	github.com/nats-io/nats.go v1.31.0
	golang.org/x/oauth2 v0.15.0
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/klauspost/compress v1.17.0 // indirect
	github.com/nats-io/nkeys v0.4.5 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.16.0 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)

replace github.com/gorilla/mux => github.com/go-chi/chi/v5 v5.2.3
