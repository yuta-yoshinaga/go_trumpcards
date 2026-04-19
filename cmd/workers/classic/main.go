//go:build js && wasm

package main

import (
	"log"
	"net/http"

	"github.com/syumai/workers"
	"github.com/syumai/workers/cloudflare"

	corsmw "github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/cors"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
	_ "github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games/classic"
)

func main() {
	mux := http.NewServeMux()

	if err := games.RegisterCategory(mux, games.CategoryClassic); err != nil {
		log.Fatal(err)
	}

	var handler http.Handler = mux
	if origins := corsmw.ParseOrigins(cloudflare.Getenv("CORS_ALLOWED_ORIGINS")); origins != nil {
		handler = corsmw.Middleware(origins, mux)
	}
	workers.Serve(handler)
}
