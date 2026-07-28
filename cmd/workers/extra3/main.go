//go:build js && wasm

package main

import (
	"log"
	"net/http"

	"github.com/syumai/workers"
	"github.com/syumai/workers/cloudflare"

	corsmw "github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/cors"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
	_ "github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games/extra3"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/recoverymw"
)

func main() {
	mux := http.NewServeMux()

	if err := games.RegisterCategory(mux, games.CategoryExtra3); err != nil {
		log.Fatal(err)
	}

	// Wrap the mux with panic recovery so any handler panic produces a JSON
	// 500 instead of Cloudflare's generic 1101 HTML error page (which loses
	// CORS headers and shows up as "通信エラー" with no actionable detail).
	var handler http.Handler = recoverymw.Middleware(mux)
	if origins := corsmw.ParseOrigins(cloudflare.Getenv("CORS_ALLOWED_ORIGINS")); origins != nil {
		handler = corsmw.Middleware(origins, handler)
	}
	workers.Serve(handler)
}
