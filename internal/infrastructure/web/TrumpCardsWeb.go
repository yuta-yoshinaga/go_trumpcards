package web

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
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
			config := domain.DefaultPokerConfig()
			players := []*domain.PokerPlayer{
				domain.NewPokerPlayer(true, domain.PokerStyleBalanced),
				domain.NewPokerPlayer(false, domain.PokerStyleConservative),
				domain.NewPokerPlayer(false, domain.PokerStyleAggressive),
				domain.NewPokerPlayer(false, domain.PokerStyleBluffer),
			}
			poker := domain.NewPoker(domain.NewTrumpCards(config.JokerCount), players, config)
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
	if allowedOriginsStr == "" && os.Getenv("APP_ENV") != "production" {
		allowedOriginsStr = "http://localhost:5173,http://localhost:8080"
	}
	if allowedOriginsStr != "" {
		allowedOrigins := make(map[string]bool, strings.Count(allowedOriginsStr, ",")+1)
		for _, origin := range strings.Split(allowedOriginsStr, ",") {
			if o := strings.TrimSpace(origin); o != "" {
				allowedOrigins[o] = true
			}
		}
		api.Use(&rest.CorsMiddleware{
			RejectNonCorsRequests: false,
			OriginValidator: func(origin string, request *rest.Request) bool {
				return allowedOrigins[origin]
			},
			AllowedMethods:                []string{"POST"},
			AllowedHeaders:                []string{"Content-Type"},
			AccessControlAllowCredentials: false,
			AccessControlMaxAge:           3600,
		})
	}
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
		readTimeout     = 10 * time.Second
		writeTimeout    = 30 * time.Second
		idleTimeout     = 60 * time.Second
		shutdownTimeout = 30 * time.Second
	)
	srv := &http.Server{
		Addr:         getListenPort(),
		Handler:      mux,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	case <-ctx.Done():
		log.Println("shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("server shutdown error: %v", err)
		}
	}

	web.bjc.Stop()
	web.pkc.Stop()
	web.omc.Stop()
	web.dgc.Stop()
	web.sgc.Stop()
	web.dwc.Stop()
	web.hmc.Stop()
	log.Println("server stopped")
}

func getListenPort() string {
	port := os.Getenv("PORT")
	if port != "" {
		return ":" + port
	}
	return ":80"
}
