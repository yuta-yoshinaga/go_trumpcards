package web

import (
	"log"
	"net/http"
	"os"

	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/presenters"
	"github.com/yuta-yoshinaga/go_trumpcards/usecases"

	"github.com/ant0ine/go-json-rest/rest"
)

// TrumpCardsWeb トランプカードゲームWebクラス
type TrumpCardsWeb struct {
	bjc *controllers.BlackJackWebController
	pkc *controllers.PokerWebController
	omc *controllers.OldMaidWebController
}

// NewTrumpCardsWeb コンストラクタ
func NewTrumpCardsWeb() *TrumpCardsWeb {
	return &TrumpCardsWeb{
		bjc: controllers.NewBlackJackWebController(usecases.NewBlackJackInteractor(presenters.NewBlackJackWebPresenter())),
		pkc: controllers.NewPokerWebController(usecases.NewPokerInteractor(presenters.NewPokerWebPresenter())),
		omc: controllers.NewOldMaidWebController(usecases.NewOldMaidInteractor(presenters.NewOldMaidWebPresenter())),
	}
}

// Exec ゲーム実行
func (web *TrumpCardsWeb) Exec() {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Post("/blackjac/exec", web.bjc.Exec),
		rest.Post("/poker/exec", web.pkc.Exec),
		rest.Post("/oldmaid/exec", web.omc.Exec),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	http.Handle("/", http.FileServer(http.Dir("public")))
	http.Handle("/blackjac/exec", api.MakeHandler())
	http.Handle("/poker/exec", api.MakeHandler())
	http.Handle("/oldmaid/exec", api.MakeHandler())
	log.Fatal(http.ListenAndServe(getListenPort(), nil))
}

func getListenPort() string {
	port := os.Getenv("PORT")
	if port != "" {
		return ":" + port
	}
	return ":80"
}
