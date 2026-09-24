//go:build js && wasm

package main

import (
	"log"
	"net/http"

	"github.com/syumai/workers-go"
	"github.com/syumai/workers-go/cloudflare"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/cors"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
	_ "github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games/extra11"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/recoverymw"
)

func main() {
	mux := http.NewServeMux()

	if err := games.RegisterCategory(mux, games.CategoryExtra11); err != nil {
		log.Fatal(err)
	}

	var handler http.Handler = recoverymw.Middleware(mux)
	if origins := cors.ParseOrigins(cloudflare.Getenv("CORS_ALLOWED_ORIGINS")); origins != nil {
		handler = cors.Middleware(origins, handler)
	}
	workers.Serve(handler)
}
