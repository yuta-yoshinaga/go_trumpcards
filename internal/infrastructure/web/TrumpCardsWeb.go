package web

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

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
	dwc *controller.DoubtWebController
	hmc *controller.HoldemWebController
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
			poker := domain.NewPoker(domain.NewTrumpCards(0), domain.NewPokerPlayer(), domain.NewPokerPlayer())
			return usecase.NewPokerInteractor(poker, presenter.NewPokerWebPresenter())
		}),
		omc: controller.NewOldMaidWebController(func() usecase.OldMaidInteractorIF {
			players := []*domain.OldMaidPlayer{
				domain.NewOldMaidPlayer(true),
				domain.NewOldMaidPlayer(false),
				domain.NewOldMaidPlayer(false),
				domain.NewOldMaidPlayer(false),
			}
			oldMaid := domain.NewOldMaid(domain.NewTrumpCards(1), players)
			return usecase.NewOldMaidInteractor(oldMaid, presenter.NewOldMaidWebPresenter())
		}),
		dgc: controller.NewDaifugoWebController(func() usecase.DaifugoInteractorIF {
			config := domain.DefaultDaifugoConfig()
			players := []*domain.DaifugoPlayer{
				domain.NewDaifugoPlayer(true),
				domain.NewDaifugoPlayer(false),
				domain.NewDaifugoPlayer(false),
				domain.NewDaifugoPlayer(false),
			}
			daifugo := domain.NewDaifugo(domain.NewTrumpCards(config.JokerCount), players, config)
			return usecase.NewDaifugoInteractor(daifugo, presenter.NewDaifugoWebPresenter())
		}),
		sgc: controller.NewSevensWebController(func() usecase.SevensInteractorIF {
			config := domain.DefaultSevensConfig()
			players := []*domain.SevensPlayer{
				domain.NewSevensPlayer(true),
				domain.NewSevensPlayer(false),
				domain.NewSevensPlayer(false),
				domain.NewSevensPlayer(false),
			}
			sevens := domain.NewSevens(domain.NewTrumpCards(config.JokerCount), players, config)
			return usecase.NewSevensInteractor(sevens, presenter.NewSevensWebPresenter())
		}),
		dwc: controller.NewDoubtWebController(func() usecase.DoubtInteractorIF {
			players := []*domain.DoubtPlayer{
				domain.NewDoubtPlayer(true),
				domain.NewDoubtPlayer(false),
				domain.NewDoubtPlayer(false),
				domain.NewDoubtPlayer(false),
			}
			doubt := domain.NewDoubt(domain.NewTrumpCards(0), players)
			return usecase.NewDoubtInteractor(doubt, presenter.NewDoubtWebPresenter())
		}),
		hmc: controller.NewHoldemWebController(func() usecase.HoldemInteractorIF {
			players := []*domain.HoldemPlayer{
				domain.NewHoldemPlayer(true, domain.HoldemStyleTAG),
				domain.NewHoldemPlayer(false, domain.HoldemStyleLAP),
				domain.NewHoldemPlayer(false, domain.HoldemStyleTAP),
				domain.NewHoldemPlayer(false, domain.HoldemStyleLAG),
			}
			holdem := domain.NewHoldem(domain.NewTrumpCards(0), players, domain.DefaultHoldemConfig())
			return usecase.NewHoldemInteractor(holdem, presenter.NewHoldemWebPresenter())
		}),
	}
}

// Exec ゲーム実行
func (web *TrumpCardsWeb) Exec() {
	api := rest.NewApi()
	allowedOriginsStr := os.Getenv("CORS_ALLOWED_ORIGINS")
	if allowedOriginsStr == "" {
		allowedOriginsStr = "http://localhost:5173,http://localhost:8080"
	}
	allowedOrigins := make(map[string]bool, strings.Count(allowedOriginsStr, ",")+1)
	for _, origin := range strings.Split(allowedOriginsStr, ",") {
		allowedOrigins[origin] = true
	}
	api.Use(&rest.CorsMiddleware{
		RejectNonCorsRequests: false,
		OriginValidator: func(origin string, request *rest.Request) bool {
			return allowedOrigins[origin]
		},
		AllowedMethods:                []string{"GET", "POST"},
		AllowedHeaders:                []string{"Content-Type"},
		AccessControlAllowCredentials: false,
		AccessControlMaxAge:           3600,
	})
	stack := rest.DefaultDevStack
	if os.Getenv("APP_ENV") == "production" {
		stack = rest.DefaultProdStack
	}
	api.Use(stack...)
	router, err := rest.MakeRouter(
		rest.Post("/blackjack/exec", web.bjc.Exec),
		rest.Post("/poker/exec", web.pkc.Exec),
		rest.Post("/oldmaid/exec", web.omc.Exec),
		rest.Post("/daifugo/exec", web.dgc.Exec),
		rest.Post("/sevens/exec", web.sgc.Exec),
		rest.Post("/doubt/exec", web.dwc.Exec),
		rest.Post("/holdem/exec", web.hmc.Exec),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	mux := http.NewServeMux()
	apiHandler := api.MakeHandler()
	apiPaths := []string{
		"/blackjack/exec",
		"/poker/exec",
		"/oldmaid/exec",
		"/daifugo/exec",
		"/sevens/exec",
		"/doubt/exec",
		"/holdem/exec",
	}
	for _, path := range apiPaths {
		mux.Handle(path, apiHandler)
	}
	mux.Handle("/", http.FileServer(http.Dir("public")))
	const (
		readTimeout  = 10 * time.Second
		writeTimeout = 30 * time.Second
		idleTimeout  = 60 * time.Second
	)
	srv := &http.Server{
		Addr:         getListenPort(),
		Handler:      mux,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}
	log.Fatal(srv.ListenAndServe())
}

func getListenPort() string {
	port := os.Getenv("PORT")
	if port != "" {
		return ":" + port
	}
	return ":80"
}
