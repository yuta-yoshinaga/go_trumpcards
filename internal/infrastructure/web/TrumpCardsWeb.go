package web

import (
	"context"
	"errors"
	"log/slog"
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
	htc *controller.HeartsWebController
	myc *controller.MemoryWebController
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
			cfg := domain.DefaultHoldemConfig()
			holdem := domain.NewHoldem(domain.NewTrumpCards(0), domain.NewPlayersForTable(cfg.TableSize), cfg)
			return usecase.NewHoldemInteractor(holdem, presenter.NewHoldemWebPresenter())
		}),
		htc: controller.NewHeartsWebController(func() usecase.HeartsInteractorIF {
			config := domain.DefaultHeartsConfig()
			players := []*domain.HeartsPlayer{
				domain.NewHeartsPlayer(true),
				domain.NewHeartsPlayer(false),
				domain.NewHeartsPlayer(false),
				domain.NewHeartsPlayer(false),
			}
			hearts := domain.NewHearts(domain.NewTrumpCards(0), players, config)
			return usecase.NewHeartsInteractor(hearts, presenter.NewHeartsWebPresenter())
		}),
		myc: controller.NewMemoryWebController(func() usecase.MemoryInteractorIF {
			config := domain.DefaultMemoryConfig()
			players := []*domain.MemoryPlayer{
				domain.NewMemoryPlayer(true),
				domain.NewMemoryPlayer(false),
				domain.NewMemoryPlayer(false),
				domain.NewMemoryPlayer(false),
			}
			memory := domain.NewMemory(domain.NewTrumpCards(0), players, config)
			return usecase.NewMemoryInteractor(memory, presenter.NewMemoryWebPresenter())
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
	type apiRoute struct {
		path    string
		handler rest.HandlerFunc
	}
	routes := []apiRoute{
		{"/blackjack/exec", web.bjc.Exec},
		{"/poker/exec", web.pkc.Exec},
		{"/oldmaid/exec", web.omc.Exec},
		{"/daifugo/exec", web.dgc.Exec},
		{"/sevens/exec", web.sgc.Exec},
		{"/doubt/exec", web.dwc.Exec},
		{"/holdem/exec", web.hmc.Exec},
		{"/hearts/exec", web.htc.Exec},
		{"/memory/exec", web.myc.Exec},
	}
	restRoutes := make([]*rest.Route, len(routes))
	for i, r := range routes {
		restRoutes[i] = rest.Post(r.path, r.handler)
	}
	router, err := rest.MakeRouter(restRoutes...)
	if err != nil {
		slog.Error("failed to create router", "error", err)
		os.Exit(1)
	}
	api.SetApp(router)
	mux := http.NewServeMux()
	apiHandler := api.MakeHandler()
	for _, r := range routes {
		mux.Handle(r.path, apiHandler)
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
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		slog.Info("shutting down server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("server shutdown error", "error", err)
		}
	}

	web.bjc.Stop()
	web.pkc.Stop()
	web.omc.Stop()
	web.dgc.Stop()
	web.sgc.Stop()
	web.dwc.Stop()
	web.hmc.Stop()
	web.htc.Stop()
	web.myc.Stop()
	slog.Info("server stopped")
}

func getListenPort() string {
	port := os.Getenv("PORT")
	if port != "" {
		return ":" + port
	}
	return ":80"
}
