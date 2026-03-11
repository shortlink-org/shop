package service

import (
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/vektah/gqlparser/v2/ast"
)

func newGraphQLHandler(svc *Service) http.Handler {
	server := handler.New(NewExecutableSchema(Config{
		Resolvers: &Resolver{Service: svc},
	}))

	server.AddTransport(transport.Options{})
	server.AddTransport(transport.GET{})
	server.AddTransport(transport.POST{})
	server.AddTransport(transport.MultipartForm{})
	server.AddTransport(transport.SSE{
		KeepAlivePingInterval: 15 * time.Second,
	})
	server.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	server.Use(extension.Introspection{})

	return server
}
