package web

import (
	"log"
	"net/http"
	"os"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/ant0ine/go-json-rest/rest"
)

// TrumpCardsWeb トランプカードゲームWebクラス
type TrumpCardsWeb struct {
	bjc *controller.BlackJackWebController
	pkc *controller.PokerWebController
	omc *controller.OldMaidWebController
	dgc *controller.DaifugoWebController
	sgc *controller.SevensWebController
}

// NewTrumpCardsWeb コンストラクタ
func NewTrumpCardsWeb() *TrumpCardsWeb {
	return &TrumpCardsWeb{
		bjc: controller.NewBlackJackWebController(func() usecase.BlackJackInteractorIF {
			return usecase.NewBlackJackInteractor(
				domain.NewDefaultBlackJack(),
				presenter.NewBlackJackWebPresenter(),
			)
		}),
		pkc: controller.NewPokerWebController(func() usecase.PokerInteractorIF {
			return usecase.NewPokerInteractor(presenter.NewPokerWebPresenter())
		}),
		omc: controller.NewOldMaidWebController(func() usecase.OldMaidInteractorIF {
			return usecase.NewOldMaidInteractor(presenter.NewOldMaidWebPresenter())
		}),
		dgc: controller.NewDaifugoWebController(func() usecase.DaifugoInteractorIF {
			return usecase.NewDaifugoInteractor(presenter.NewDaifugoWebPresenter())
		}),
		sgc: controller.NewSevensWebController(func() usecase.SevensInteractorIF {
			return usecase.NewSevensInteractor(presenter.NewSevensWebPresenter())
		}),
	}
}

// Exec ゲーム実行
func (web *TrumpCardsWeb) Exec() {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Post("/blackjack/exec", web.bjc.Exec),
		rest.Post("/poker/exec", web.pkc.Exec),
		rest.Post("/oldmaid/exec", web.omc.Exec),
		rest.Post("/daifugo/exec", web.dgc.Exec),
		rest.Post("/sevens/exec", web.sgc.Exec),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	http.Handle("/", http.FileServer(http.Dir("public")))
	http.Handle("/blackjack/exec", api.MakeHandler())
	http.Handle("/poker/exec", api.MakeHandler())
	http.Handle("/oldmaid/exec", api.MakeHandler())
	http.Handle("/daifugo/exec", api.MakeHandler())
	http.Handle("/sevens/exec", api.MakeHandler())
	log.Fatal(http.ListenAndServe(getListenPort(), nil))
}

func getListenPort() string {
	port := os.Getenv("PORT")
	if port != "" {
		return ":" + port
	}
	return ":80"
}
