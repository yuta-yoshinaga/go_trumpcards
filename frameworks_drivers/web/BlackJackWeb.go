package web

import (
	"go_trumpcards/interface_adapters/controllers"
	"go_trumpcards/interface_adapters/presenters"
	"go_trumpcards/usecases"
	"log"
	"net/http"
	"os"

	"github.com/ant0ine/go-json-rest/rest"
)

// BlackJackWeb ブラックジャックWebクラス
type BlackJackWeb struct {
	bjc *controllers.BlackJackWebController
}

// NewBlackJackWeb コンストラクタ
func NewBlackJackWeb() *BlackJackWeb {
	return &BlackJackWeb{
		bjc: controllers.NewBlackJackWebController(usecases.NewBlackJackInteractor(presenters.NewBlackJackWebPresenter())),
	}
}

// Exec ゲーム実行
func (web *BlackJackWeb) Exec() {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Post("/blackjac/exec", web.bjc.Exec),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	http.Handle("/", http.FileServer(http.Dir("public")))
	http.Handle("/blackjac/exec", api.MakeHandler())
	log.Fatal(http.ListenAndServe(getListenPort(), nil))
}

func getListenPort() string {
	port := os.Getenv("PORT")
	if port != "" {
		return ":" + port
	}
	return ":80"
}
