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
	"syscall"
	"time"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	corsmw "github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/cors"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// webController is the common interface for all game web controllers.
type webController interface {
	Exec(http.ResponseWriter, *http.Request)
	Stop()
}

// gameEntry pairs a route name with its controller constructor.
type gameEntry struct {
	name       string
	controller webController
}

// TrumpCardsWeb トランプカードゲームWebクラス
type TrumpCardsWeb struct {
	games []gameEntry
}

// NewTrumpCardsWeb コンストラクタ
func NewTrumpCardsWeb() *TrumpCardsWeb {
	web := &TrumpCardsWeb{}
	web.registerAll()
	return web
}

// register adds a game controller to the registry.
func (web *TrumpCardsWeb) register(name string, c webController) {
	web.games = append(web.games, gameEntry{name: name, controller: c})
}

// registerAll registers all game controllers.
func (web *TrumpCardsWeb) registerAll() {
	web.register("blackjack", controller.NewBlackJackWebController(func() usecase.BlackJackInteractorIF {
		return usecase.NewBlackJackInteractor(
			domain.NewDefaultBlackJack(),
			new(presenter.BlackJackWebPresenter),
		)
	}))
	web.register("poker", controller.NewPokerWebController(func() usecase.PokerInteractorIF {
		return usecase.NewPokerInteractor(domain.NewDefaultPoker(), new(presenter.PokerWebPresenter))
	}))
	web.register("oldmaid", controller.NewOldMaidWebController(func() usecase.OldMaidInteractorIF {
		return usecase.NewOldMaidInteractor(domain.NewDefaultOldMaid(), new(presenter.OldMaidWebPresenter))
	}))
	web.register("daifugo", controller.NewDaifugoWebController(func() usecase.DaifugoInteractorIF {
		return usecase.NewDaifugoInteractor(domain.NewDefaultDaifugo(), new(presenter.DaifugoWebPresenter))
	}))
	web.register("sevens", controller.NewSevensWebController(func() usecase.SevensInteractorIF {
		return usecase.NewSevensInteractor(domain.NewDefaultSevens(), new(presenter.SevensWebPresenter))
	}))
	web.register("doubt", controller.NewDoubtWebController(func() usecase.DoubtInteractorIF {
		return usecase.NewDoubtInteractor(domain.NewDefaultDoubt(), new(presenter.DoubtWebPresenter))
	}))
	web.register("holdem", controller.NewHoldemWebController(func() usecase.HoldemInteractorIF {
		return usecase.NewHoldemInteractor(domain.NewDefaultHoldem(), new(presenter.HoldemWebPresenter))
	}))
	web.register("omaha", controller.NewOmahaWebController(func() usecase.OmahaInteractorIF {
		return usecase.NewOmahaInteractor(domain.NewDefaultOmaha(), new(presenter.OmahaWebPresenter))
	}))
	web.register("shortdeck", controller.NewShortDeckWebController(func() usecase.ShortDeckInteractorIF {
		return usecase.NewShortDeckInteractor(domain.NewDefaultShortDeck(), new(presenter.ShortDeckWebPresenter))
	}))
	web.register("hearts", controller.NewHeartsWebController(func() usecase.HeartsInteractorIF {
		return usecase.NewHeartsInteractor(domain.NewDefaultHearts(), new(presenter.HeartsWebPresenter))
	}))
	web.register("memory", controller.NewMemoryWebController(func() usecase.MemoryInteractorIF {
		return usecase.NewMemoryInteractor(domain.NewDefaultMemory(), new(presenter.MemoryWebPresenter))
	}))
	web.register("klondike", controller.NewKlondikeWebController(func() usecase.KlondikeInteractorIF {
		return usecase.NewKlondikeInteractor(domain.NewDefaultKlondike(), new(presenter.KlondikeWebPresenter))
	}))
	web.register("freecell", controller.NewFreeCellWebController(func() usecase.FreeCellInteractorIF {
		return usecase.NewFreeCellInteractor(domain.NewDefaultFreeCell(), new(presenter.FreeCellWebPresenter))
	}))
	web.register("baccarat", controller.NewBaccaratWebController(func() usecase.BaccaratInteractorIF {
		return usecase.NewBaccaratInteractor(domain.NewDefaultBaccarat(), new(presenter.BaccaratWebPresenter))
	}))
	web.register("spades", controller.NewSpadesWebController(func() usecase.SpadesInteractorIF {
		return usecase.NewSpadesInteractor(domain.NewDefaultSpades(), new(presenter.SpadesWebPresenter))
	}))
	web.register("crazyeights", controller.NewCrazyEightsWebController(func() usecase.CrazyEightsInteractorIF {
		return usecase.NewCrazyEightsInteractor(domain.NewDefaultCrazyEights(), new(presenter.CrazyEightsWebPresenter))
	}))
	web.register("ginrummy", controller.NewGinRummyWebController(func() usecase.GinRummyInteractorIF {
		return usecase.NewGinRummyInteractor(domain.NewDefaultGinRummy(), new(presenter.GinRummyWebPresenter))
	}))
	web.register("spider", controller.NewSpiderWebController(func() usecase.SpiderInteractorIF {
		return usecase.NewSpiderInteractor(domain.NewDefaultSpider(), new(presenter.SpiderWebPresenter))
	}))
	web.register("napoleon", controller.NewNapoleonWebController(func() usecase.NapoleonInteractorIF {
		return usecase.NewNapoleonInteractor(domain.NewDefaultNapoleon(), new(presenter.NapoleonWebPresenter))
	}))
	web.register("indianpoker", controller.NewIndianPokerWebController(func() usecase.IndianPokerInteractorIF {
		return usecase.NewIndianPokerInteractor(domain.NewDefaultIndianPoker(), new(presenter.IndianPokerWebPresenter))
	}))
	web.register("videopoker", controller.NewVideoPokerWebController(func() usecase.VideoPokerInteractorIF {
		return usecase.NewVideoPokerInteractor(
			domain.NewDefaultVideoPoker(),
			new(presenter.VideoPokerWebPresenter),
		)
	}))
	web.register("deuceswild", controller.NewVideoPokerWebController(func() usecase.VideoPokerInteractorIF {
		return usecase.NewVideoPokerInteractor(
			domain.NewDeucesWildVideoPoker(),
			new(presenter.VideoPokerWebPresenter),
		)
	}))
	web.register("jokerpoker", controller.NewVideoPokerWebController(func() usecase.VideoPokerInteractorIF {
		return usecase.NewVideoPokerInteractor(
			domain.NewJokerPokerVideoPoker(),
			new(presenter.VideoPokerWebPresenter),
		)
	}))
	web.register("euchre", controller.NewEuchreWebController(func() usecase.EuchreInteractorIF {
		return usecase.NewEuchreInteractor(domain.NewDefaultEuchre(), new(presenter.EuchreWebPresenter))
	}))
	web.register("pyramid", controller.NewPyramidWebController(func() usecase.PyramidInteractorIF {
		return usecase.NewPyramidInteractor(domain.NewDefaultPyramid(), new(presenter.PyramidWebPresenter))
	}))
	web.register("tripeaks", controller.NewTriPeaksWebController(func() usecase.TriPeaksInteractorIF {
		return usecase.NewTriPeaksInteractor(domain.NewDefaultTriPeaks(), new(presenter.TriPeaksWebPresenter))
	}))
	web.register("cribbage", controller.NewCribbageWebController(func() usecase.CribbageInteractorIF {
		return usecase.NewCribbageInteractor(domain.NewDefaultCribbage(), new(presenter.CribbageWebPresenter))
	}))
	web.register("threecard", controller.NewThreeCardWebController(func() usecase.ThreeCardInteractorIF {
		return usecase.NewThreeCardInteractor(
			domain.NewDefaultThreeCard(),
			new(presenter.ThreeCardWebPresenter),
		)
	}))
	web.register("ohhell", controller.NewOhHellWebController(func() usecase.OhHellInteractorIF {
		return usecase.NewOhHellInteractor(domain.NewDefaultOhHell(), new(presenter.OhHellWebPresenter))
	}))
	web.register("bridge", controller.NewBridgeWebController(func() usecase.BridgeInteractorIF {
		return usecase.NewBridgeInteractor(domain.NewDefaultBridge(), new(presenter.BridgeWebPresenter))
	}))
	web.register("pineapple", controller.NewPineappleWebController(func() usecase.PineappleInteractorIF {
		return usecase.NewPineappleInteractor(domain.NewDefaultPineapple(), new(presenter.PineappleWebPresenter))
	}))
	web.register("speed", controller.NewSpeedWebController(func() usecase.SpeedInteractorIF {
		return usecase.NewSpeedInteractor(domain.NewDefaultSpeed(), new(presenter.SpeedWebPresenter))
	}))
	web.register("gofish", controller.NewGoFishWebController(func() usecase.GoFishInteractorIF {
		return usecase.NewGoFishInteractor(domain.NewDefaultGoFish(), new(presenter.GoFishWebPresenter))
	}))
	web.register("canasta", controller.NewCanastaWebController(func() usecase.CanastaInteractorIF {
		return usecase.NewCanastaInteractor(domain.NewDefaultCanasta(), new(presenter.CanastaWebPresenter))
	}))
	web.register("pinochle", controller.NewPinochleWebController(func() usecase.PinochleInteractorIF {
		return usecase.NewPinochleInteractor(domain.NewDefaultPinochle(), new(presenter.PinochleWebPresenter))
	}))
	web.register("golf", controller.NewGolfWebController(func() usecase.GolfInteractorIF {
		return usecase.NewGolfInteractor(domain.NewDefaultGolf(), new(presenter.GolfWebPresenter))
	}))
	web.register("pigtail", controller.NewPigsTailWebController(func() usecase.PigsTailInteractorIF {
		return usecase.NewPigsTailInteractor(domain.NewDefaultPigsTail(), new(presenter.PigsTailWebPresenter))
	}))
	web.register("sevencardstud", controller.NewSevenCardStudWebController(func() usecase.SevenCardStudInteractorIF {
		return usecase.NewSevenCardStudInteractor(domain.NewDefaultSevenCardStud(), new(presenter.SevenCardStudWebPresenter))
	}))
	web.register("razz", controller.NewSevenCardStudWebController(func() usecase.SevenCardStudInteractorIF {
		return usecase.NewSevenCardStudInteractor(domain.NewDefaultRazz(), new(presenter.SevenCardStudWebPresenter))
	}))
	web.register("clocksolitaire", controller.NewClockSolitaireWebController(func() usecase.ClockSolitaireInteractorIF {
		return usecase.NewClockSolitaireInteractor(domain.NewDefaultClockSolitaire(), new(presenter.ClockSolitaireWebPresenter))
	}))
	web.register("durak", controller.NewDurakWebController(func() usecase.DurakInteractorIF {
		return usecase.NewDurakInteractor(domain.NewDefaultDurak(), new(presenter.DurakWebPresenter))
	}))
	web.register("fortythieves", controller.NewFortyThievesWebController(func() usecase.FortyThievesInteractorIF {
		return usecase.NewFortyThievesInteractor(domain.NewDefaultFortyThieves(), new(presenter.FortyThievesWebPresenter))
	}))
	web.register("paigow", controller.NewPaiGowWebController(func() usecase.PaiGowInteractorIF {
		return usecase.NewPaiGowInteractor(
			domain.NewDefaultPaiGow(),
			new(presenter.PaiGowWebPresenter),
		)
	}))
	web.register("twotenjack", controller.NewTwoTenJackWebController(func() usecase.TwoTenJackInteractorIF {
		return usecase.NewTwoTenJackInteractor(domain.NewDefaultTwoTenJack(), new(presenter.TwoTenJackWebPresenter))
	}))
	web.register("caribbeanstud", controller.NewCaribbeanStudWebController(func() usecase.CaribbeanStudInteractorIF {
		return usecase.NewCaribbeanStudInteractor(
			domain.NewDefaultCaribbeanStud(),
			new(presenter.CaribbeanStudWebPresenter),
		)
	}))
	web.register("war", controller.NewWarWebController(func() usecase.WarInteractorIF {
		return usecase.NewWarInteractor(domain.NewDefaultWar(), new(presenter.WarWebPresenter))
	}))
	web.register("canfield", controller.NewCanfieldWebController(func() usecase.CanfieldInteractorIF {
		return usecase.NewCanfieldInteractor(domain.NewDefaultCanfield(), new(presenter.CanfieldWebPresenter))
	}))
	web.register("fiftyone", controller.NewFiftyOneWebController(func() usecase.FiftyOneInteractorIF {
		return usecase.NewFiftyOneInteractor(domain.NewDefaultFiftyOne(), new(presenter.FiftyOneWebPresenter))
	}))
	web.register("yukon", controller.NewYukonWebController(func() usecase.YukonInteractorIF {
		return usecase.NewYukonInteractor(domain.NewDefaultYukon(), new(presenter.YukonWebPresenter))
	}))
	web.register("scorpion", controller.NewScorpionWebController(func() usecase.ScorpionInteractorIF {
		return usecase.NewScorpionInteractor(domain.NewDefaultScorpion(), new(presenter.ScorpionWebPresenter))
	}))
	web.register("whist", controller.NewWhistWebController(func() usecase.WhistInteractorIF {
		return usecase.NewWhistInteractor(domain.NewDefaultWhist(), new(presenter.WhistWebPresenter))
	}))
	web.register("letitride", controller.NewLetItRideWebController(func() usecase.LetItRideInteractorIF {
		return usecase.NewLetItRideInteractor(
			domain.NewDefaultLetItRide(),
			new(presenter.LetItRideWebPresenter),
		)
	}))
	web.register("pokersquares", controller.NewPokerSquaresWebController(func() usecase.PokerSquaresInteractorIF {
		return usecase.NewPokerSquaresInteractor(domain.NewDefaultPokerSquares(), new(presenter.PokerSquaresWebPresenter))
	}))
	web.register("pageone", controller.NewPageOneWebController(func() usecase.PageOneInteractorIF {
		return usecase.NewPageOneInteractor(domain.NewDefaultPageOne(), new(presenter.PageOneWebPresenter))
	}))
	web.register("reddog", controller.NewRedDogWebController(func() usecase.RedDogInteractorIF {
		return usecase.NewRedDogInteractor(
			domain.NewDefaultRedDog(),
			new(presenter.RedDogWebPresenter),
		)
	}))
}

// Exec ゲーム実行
func (web *TrumpCardsWeb) Exec() error {
	mux := http.NewServeMux()

	for _, g := range web.games {
		mux.HandleFunc("POST /"+g.name+"/exec", g.controller.Exec)
	}
	RegisterSwaggerRoutes(mux)
	mux.Handle("/", http.FileServer(http.Dir("public")))

	// Apply CORS middleware if allowed origins are configured.
	allowedOriginsStr := os.Getenv("CORS_ALLOWED_ORIGINS")
	if allowedOriginsStr == "" && os.Getenv("APP_ENV") != "production" {
		allowedOriginsStr = "http://localhost:5173,http://localhost:8080"
	}
	var handler http.Handler = mux
	if origins := corsmw.ParseOrigins(allowedOriginsStr); origins != nil {
		handler = corsmw.Middleware(origins, mux)
	}
	const (
		readTimeout     = 10 * time.Second
		writeTimeout    = 30 * time.Second
		idleTimeout     = 60 * time.Second
		shutdownTimeout = 30 * time.Second
	)
	srv := &http.Server{
		Addr:         getListenAddr(),
		Handler:      handler,
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
	fmt.Fprintln(os.Stderr, i18n.Tf("webServerRunning", "addr", ln.Addr().String()))
	fmt.Fprintln(os.Stderr, i18n.T("webServerStop"))

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
		fmt.Fprintln(os.Stderr, "\n"+i18n.T("webServerShutdown"))
		slog.Info("shutting down server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("server shutdown error", "error", err)
		}
	}

	for _, g := range web.games {
		g.controller.Stop()
	}
	fmt.Fprintln(os.Stderr, i18n.T("webServerStopped"))
	slog.Info("server stopped")
	return runErr
}

func getListenAddr() string {
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return net.JoinHostPort(host, port)
}
