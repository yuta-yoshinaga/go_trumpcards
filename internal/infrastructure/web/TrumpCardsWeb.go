package web

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
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
	bjc  *controller.BlackJackWebController
	pkc  *controller.PokerWebController
	omc  *controller.OldMaidWebController
	dgc  *controller.DaifugoWebController
	sgc  *controller.SevensWebController
	dwc  *controller.DoubtWebController
	hmc  *controller.HoldemWebController
	ohc  *controller.OmahaWebController
	skc  *controller.ShortDeckWebController
	htc  *controller.HeartsWebController
	myc  *controller.MemoryWebController
	klc  *controller.KlondikeWebController
	fcc  *controller.FreeCellWebController
	bcc  *controller.BaccaratWebController
	spc  *controller.SpadesWebController
	cec  *controller.CrazyEightsWebController
	grc  *controller.GinRummyWebController
	sdc  *controller.SpiderWebController
	npc  *controller.NapoleonWebController
	ipc  *controller.IndianPokerWebController
	vpc  *controller.VideoPokerWebController
	dwwc *controller.VideoPokerWebController
	jpwc *controller.VideoPokerWebController
	euc  *controller.EuchreWebController
	pyc  *controller.PyramidWebController
	cbc  *controller.CribbageWebController
}

// NewTrumpCardsWeb コンストラクタ
func NewTrumpCardsWeb() *TrumpCardsWeb {
	return &TrumpCardsWeb{
		bjc: controller.NewBlackJackWebController(func() usecase.BlackJackInteractorIF {
			return usecase.NewBlackJackInteractor(
				domain.NewDefaultBlackJack(),
				new(presenter.BlackJackWebPresenter),
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
			return usecase.NewPokerInteractor(poker, new(presenter.PokerWebPresenter))
		}),
		omc: controller.NewOldMaidWebController(func() usecase.OldMaidInteractorIF {
			players := []*domain.OldMaidPlayer{
				domain.NewOldMaidPlayer(true),
				domain.NewOldMaidPlayer(false),
				domain.NewOldMaidPlayer(false),
				domain.NewOldMaidPlayer(false),
			}
			oldMaid := domain.NewOldMaid(domain.NewTrumpCards(1), players)
			return usecase.NewOldMaidInteractor(oldMaid, new(presenter.OldMaidWebPresenter))
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
			return usecase.NewDaifugoInteractor(daifugo, new(presenter.DaifugoWebPresenter))
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
			return usecase.NewSevensInteractor(sevens, new(presenter.SevensWebPresenter))
		}),
		dwc: controller.NewDoubtWebController(func() usecase.DoubtInteractorIF {
			players := []*domain.DoubtPlayer{
				domain.NewDoubtPlayer(true),
				domain.NewDoubtPlayer(false),
				domain.NewDoubtPlayer(false),
				domain.NewDoubtPlayer(false),
			}
			doubt := domain.NewDoubt(domain.NewTrumpCards(0), players)
			return usecase.NewDoubtInteractor(doubt, new(presenter.DoubtWebPresenter))
		}),
		hmc: controller.NewHoldemWebController(func() usecase.HoldemInteractorIF {
			cfg := domain.DefaultHoldemConfig()
			holdem := domain.NewHoldem(domain.NewTrumpCards(0), domain.NewPlayersForTable(cfg.TableSize), cfg)
			return usecase.NewHoldemInteractor(holdem, new(presenter.HoldemWebPresenter))
		}),
		ohc: controller.NewOmahaWebController(func() usecase.OmahaInteractorIF {
			cfg := domain.DefaultOmahaConfig()
			omaha := domain.NewOmaha(domain.NewTrumpCards(0), domain.NewOmahaPlayersForTable(cfg.TableSize), cfg)
			return usecase.NewOmahaInteractor(omaha, new(presenter.OmahaWebPresenter))
		}),
		skc: controller.NewShortDeckWebController(func() usecase.ShortDeckInteractorIF {
			cfg := domain.DefaultShortDeckConfig()
			sd := domain.NewShortDeck(domain.NewTrumpCardsShortDeck(), domain.NewShortDeckPlayersForTable(cfg.TableSize), cfg)
			return usecase.NewShortDeckInteractor(sd, new(presenter.ShortDeckWebPresenter))
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
			return usecase.NewHeartsInteractor(hearts, new(presenter.HeartsWebPresenter))
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
			return usecase.NewMemoryInteractor(memory, new(presenter.MemoryWebPresenter))
		}),
		klc: controller.NewKlondikeWebController(func() usecase.KlondikeInteractorIF {
			klondike := domain.NewKlondike(domain.NewTrumpCards(0))
			return usecase.NewKlondikeInteractor(klondike, new(presenter.KlondikeWebPresenter))
		}),
		fcc: controller.NewFreeCellWebController(func() usecase.FreeCellInteractorIF {
			freeCell := domain.NewFreeCell(domain.NewTrumpCards(0))
			return usecase.NewFreeCellInteractor(freeCell, new(presenter.FreeCellWebPresenter))
		}),
		bcc: controller.NewBaccaratWebController(func() usecase.BaccaratInteractorIF {
			baccarat := domain.NewDefaultBaccarat()
			return usecase.NewBaccaratInteractor(baccarat, new(presenter.BaccaratWebPresenter))
		}),
		spc: controller.NewSpadesWebController(func() usecase.SpadesInteractorIF {
			config := domain.DefaultSpadesConfig()
			players := []*domain.SpadesPlayer{
				domain.NewSpadesPlayer(true),
				domain.NewSpadesPlayer(false),
				domain.NewSpadesPlayer(false),
				domain.NewSpadesPlayer(false),
			}
			spades := domain.NewSpades(domain.NewTrumpCards(0), players, config)
			return usecase.NewSpadesInteractor(spades, new(presenter.SpadesWebPresenter))
		}),
		cec: controller.NewCrazyEightsWebController(func() usecase.CrazyEightsInteractorIF {
			config := domain.DefaultCrazyEightsConfig()
			players := []*domain.CrazyEightsPlayer{
				domain.NewCrazyEightsPlayer(true),
				domain.NewCrazyEightsPlayer(false),
				domain.NewCrazyEightsPlayer(false),
				domain.NewCrazyEightsPlayer(false),
			}
			ce := domain.NewCrazyEights(domain.NewTrumpCards(0), players, config)
			return usecase.NewCrazyEightsInteractor(ce, new(presenter.CrazyEightsWebPresenter))
		}),
		grc: controller.NewGinRummyWebController(func() usecase.GinRummyInteractorIF {
			config := domain.DefaultGinRummyConfig()
			players := []*domain.GinRummyPlayer{
				domain.NewGinRummyPlayer(true),
				domain.NewGinRummyPlayer(false),
			}
			gr := domain.NewGinRummy(domain.NewTrumpCards(0), players, config)
			return usecase.NewGinRummyInteractor(gr, new(presenter.GinRummyWebPresenter))
		}),
		sdc: controller.NewSpiderWebController(func() usecase.SpiderInteractorIF {
			spider := domain.NewSpider(domain.NewTrumpCardsWithSuits(domain.SpiderTotalCards, []int{domain.CardDesignSpade}))
			return usecase.NewSpiderInteractor(spider, new(presenter.SpiderWebPresenter))
		}),
		npc: controller.NewNapoleonWebController(func() usecase.NapoleonInteractorIF {
			config := domain.DefaultNapoleonConfig()
			players := []*domain.NapoleonPlayer{
				domain.NewNapoleonPlayer(true),
				domain.NewNapoleonPlayer(false),
				domain.NewNapoleonPlayer(false),
				domain.NewNapoleonPlayer(false),
			}
			napoleon := domain.NewNapoleon(domain.NewTrumpCards(1), players, config)
			return usecase.NewNapoleonInteractor(napoleon, new(presenter.NapoleonWebPresenter))
		}),
		ipc: controller.NewIndianPokerWebController(func() usecase.IndianPokerInteractorIF {
			cfg := domain.DefaultIndianPokerConfig()
			ip := domain.NewIndianPoker(domain.NewTrumpCards(0), domain.NewIndianPokerPlayers(), cfg)
			return usecase.NewIndianPokerInteractor(ip, new(presenter.IndianPokerWebPresenter))
		}),
		vpc: controller.NewVideoPokerWebController(func() usecase.VideoPokerInteractorIF {
			return usecase.NewVideoPokerInteractor(
				domain.NewDefaultVideoPoker(),
				new(presenter.VideoPokerWebPresenter),
			)
		}),
		dwwc: controller.NewVideoPokerWebController(func() usecase.VideoPokerInteractorIF {
			return usecase.NewVideoPokerInteractor(
				domain.NewDeucesWildVideoPoker(),
				new(presenter.VideoPokerWebPresenter),
			)
		}),
		jpwc: controller.NewVideoPokerWebController(func() usecase.VideoPokerInteractorIF {
			return usecase.NewVideoPokerInteractor(
				domain.NewJokerPokerVideoPoker(),
				new(presenter.VideoPokerWebPresenter),
			)
		}),
		euc: controller.NewEuchreWebController(func() usecase.EuchreInteractorIF {
			config := domain.DefaultEuchreConfig()
			players := []*domain.EuchrePlayer{
				domain.NewEuchrePlayer(true, 0),
				domain.NewEuchrePlayer(false, 1),
				domain.NewEuchrePlayer(false, 0),
				domain.NewEuchrePlayer(false, 1),
			}
			euchre := domain.NewEuchre(domain.NewTrumpCardsEuchre(), players, config)
			return usecase.NewEuchreInteractor(euchre, new(presenter.EuchreWebPresenter))
		}),
		pyc: controller.NewPyramidWebController(func() usecase.PyramidInteractorIF {
			pyramid := domain.NewPyramid(domain.NewTrumpCards(0))
			return usecase.NewPyramidInteractor(pyramid, new(presenter.PyramidWebPresenter))
		}),
		cbc: controller.NewCribbageWebController(func() usecase.CribbageInteractorIF {
			config := domain.DefaultCribbageConfig()
			players := []*domain.CribbagePlayer{
				domain.NewCribbagePlayer(true),
				domain.NewCribbagePlayer(false),
			}
			cribbage := domain.NewCribbage(domain.NewTrumpCards(0), players, config)
			return usecase.NewCribbageInteractor(cribbage, new(presenter.CribbageWebPresenter))
		}),
	}
}

// Exec ゲーム実行
func (web *TrumpCardsWeb) Exec() error {
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
		{"/omaha/exec", web.ohc.Exec},
		{"/shortdeck/exec", web.skc.Exec},
		{"/hearts/exec", web.htc.Exec},
		{"/memory/exec", web.myc.Exec},
		{"/klondike/exec", web.klc.Exec},
		{"/freecell/exec", web.fcc.Exec},
		{"/baccarat/exec", web.bcc.Exec},
		{"/spades/exec", web.spc.Exec},
		{"/crazyeights/exec", web.cec.Exec},
		{"/ginrummy/exec", web.grc.Exec},
		{"/spider/exec", web.sdc.Exec},
		{"/napoleon/exec", web.npc.Exec},
		{"/indianpoker/exec", web.ipc.Exec},
		{"/videopoker/exec", web.vpc.Exec},
		{"/deuceswild/exec", web.dwwc.Exec},
		{"/jokerpoker/exec", web.jpwc.Exec},
		{"/euchre/exec", web.euc.Exec},
		{"/pyramid/exec", web.pyc.Exec},
		{"/cribbage/exec", web.cbc.Exec},
	}
	restRoutes := make([]*rest.Route, len(routes))
	for i, r := range routes {
		restRoutes[i] = rest.Post(r.path, r.handler)
	}
	router, err := rest.MakeRouter(restRoutes...)
	if err != nil {
		slog.Error("failed to create router", "error", err)
		return fmt.Errorf("failed to create router: %w", err)
	}
	api.SetApp(router)
	mux := http.NewServeMux()
	apiHandler := api.MakeHandler()
	for _, r := range routes {
		mux.Handle(r.path, apiHandler)
	}
	RegisterSwaggerRoutes(mux)
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

	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		slog.Error("server listen error", "error", err)
		return fmt.Errorf("failed to listen on %s: %w", srv.Addr, err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	fmt.Printf("Server is running at http://localhost:%d\n", port)
	fmt.Println("Press Ctrl+C to stop")

	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(ln) }()

	var runErr error
	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			runErr = fmt.Errorf("server error: %w", err)
		}
	case <-ctx.Done():
		fmt.Println("\nShutting down server...")
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
	web.ohc.Stop()
	web.skc.Stop()
	web.htc.Stop()
	web.myc.Stop()
	web.klc.Stop()
	web.fcc.Stop()
	web.bcc.Stop()
	web.spc.Stop()
	web.cec.Stop()
	web.grc.Stop()
	web.sdc.Stop()
	web.npc.Stop()
	web.ipc.Stop()
	web.vpc.Stop()
	web.euc.Stop()
	fmt.Println("Server stopped.")
	slog.Info("server stopped")
	return runErr
}

func getListenPort() string {
	port := os.Getenv("PORT")
	if port != "" {
		return ":" + port
	}
	return ":8080"
}
