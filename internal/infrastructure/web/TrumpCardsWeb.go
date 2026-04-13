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
	"strconv"
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
		config := domain.DefaultPokerConfig()
		players := []*domain.PokerPlayer{
			domain.NewPokerPlayer(true, domain.PokerStyleBalanced),
			domain.NewPokerPlayer(false, domain.PokerStyleConservative),
			domain.NewPokerPlayer(false, domain.PokerStyleAggressive),
			domain.NewPokerPlayer(false, domain.PokerStyleBluffer),
		}
		poker := domain.NewPoker(domain.NewTrumpCards(config.JokerCount), players, config)
		return usecase.NewPokerInteractor(poker, new(presenter.PokerWebPresenter))
	}))
	web.register("oldmaid", controller.NewOldMaidWebController(func() usecase.OldMaidInteractorIF {
		players := []*domain.OldMaidPlayer{
			domain.NewOldMaidPlayer(true),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
		}
		oldMaid := domain.NewOldMaid(domain.NewTrumpCards(1), players)
		return usecase.NewOldMaidInteractor(oldMaid, new(presenter.OldMaidWebPresenter))
	}))
	web.register("daifugo", controller.NewDaifugoWebController(func() usecase.DaifugoInteractorIF {
		config := domain.DefaultDaifugoConfig()
		players := []*domain.DaifugoPlayer{
			domain.NewDaifugoPlayer(true),
			domain.NewDaifugoPlayer(false),
			domain.NewDaifugoPlayer(false),
			domain.NewDaifugoPlayer(false),
		}
		daifugo := domain.NewDaifugo(domain.NewTrumpCards(config.JokerCount), players, config)
		return usecase.NewDaifugoInteractor(daifugo, new(presenter.DaifugoWebPresenter))
	}))
	web.register("sevens", controller.NewSevensWebController(func() usecase.SevensInteractorIF {
		config := domain.DefaultSevensConfig()
		players := []*domain.SevensPlayer{
			domain.NewSevensPlayer(true),
			domain.NewSevensPlayer(false),
			domain.NewSevensPlayer(false),
			domain.NewSevensPlayer(false),
		}
		sevens := domain.NewSevens(domain.NewTrumpCards(config.JokerCount), players, config)
		return usecase.NewSevensInteractor(sevens, new(presenter.SevensWebPresenter))
	}))
	web.register("doubt", controller.NewDoubtWebController(func() usecase.DoubtInteractorIF {
		players := []*domain.DoubtPlayer{
			domain.NewDoubtPlayer(true),
			domain.NewDoubtPlayer(false),
			domain.NewDoubtPlayer(false),
			domain.NewDoubtPlayer(false),
		}
		doubt := domain.NewDoubt(domain.NewTrumpCards(0), players)
		return usecase.NewDoubtInteractor(doubt, new(presenter.DoubtWebPresenter))
	}))
	web.register("holdem", controller.NewHoldemWebController(func() usecase.HoldemInteractorIF {
		cfg := domain.DefaultHoldemConfig()
		holdem := domain.NewHoldem(domain.NewTrumpCards(0), domain.NewPlayersForTable(cfg.TableSize), cfg)
		return usecase.NewHoldemInteractor(holdem, new(presenter.HoldemWebPresenter))
	}))
	web.register("omaha", controller.NewOmahaWebController(func() usecase.OmahaInteractorIF {
		cfg := domain.DefaultOmahaConfig()
		omaha := domain.NewOmaha(domain.NewTrumpCards(0), domain.NewOmahaPlayersForTable(cfg.TableSize), cfg)
		return usecase.NewOmahaInteractor(omaha, new(presenter.OmahaWebPresenter))
	}))
	web.register("shortdeck", controller.NewShortDeckWebController(func() usecase.ShortDeckInteractorIF {
		cfg := domain.DefaultShortDeckConfig()
		sd := domain.NewShortDeck(domain.NewTrumpCardsShortDeck(), domain.NewShortDeckPlayersForTable(cfg.TableSize), cfg)
		return usecase.NewShortDeckInteractor(sd, new(presenter.ShortDeckWebPresenter))
	}))
	web.register("hearts", controller.NewHeartsWebController(func() usecase.HeartsInteractorIF {
		config := domain.DefaultHeartsConfig()
		players := []*domain.HeartsPlayer{
			domain.NewHeartsPlayer(true),
			domain.NewHeartsPlayer(false),
			domain.NewHeartsPlayer(false),
			domain.NewHeartsPlayer(false),
		}
		hearts := domain.NewHearts(domain.NewTrumpCards(0), players, config)
		return usecase.NewHeartsInteractor(hearts, new(presenter.HeartsWebPresenter))
	}))
	web.register("memory", controller.NewMemoryWebController(func() usecase.MemoryInteractorIF {
		config := domain.DefaultMemoryConfig()
		players := []*domain.MemoryPlayer{
			domain.NewMemoryPlayer(true),
			domain.NewMemoryPlayer(false),
			domain.NewMemoryPlayer(false),
			domain.NewMemoryPlayer(false),
		}
		memory := domain.NewMemory(domain.NewTrumpCards(0), players, config)
		return usecase.NewMemoryInteractor(memory, new(presenter.MemoryWebPresenter))
	}))
	web.register("klondike", controller.NewKlondikeWebController(func() usecase.KlondikeInteractorIF {
		klondike := domain.NewKlondike(domain.NewTrumpCards(0))
		return usecase.NewKlondikeInteractor(klondike, new(presenter.KlondikeWebPresenter))
	}))
	web.register("freecell", controller.NewFreeCellWebController(func() usecase.FreeCellInteractorIF {
		freeCell := domain.NewFreeCell(domain.NewTrumpCards(0))
		return usecase.NewFreeCellInteractor(freeCell, new(presenter.FreeCellWebPresenter))
	}))
	web.register("baccarat", controller.NewBaccaratWebController(func() usecase.BaccaratInteractorIF {
		baccarat := domain.NewDefaultBaccarat()
		return usecase.NewBaccaratInteractor(baccarat, new(presenter.BaccaratWebPresenter))
	}))
	web.register("spades", controller.NewSpadesWebController(func() usecase.SpadesInteractorIF {
		config := domain.DefaultSpadesConfig()
		players := []*domain.SpadesPlayer{
			domain.NewSpadesPlayer(true),
			domain.NewSpadesPlayer(false),
			domain.NewSpadesPlayer(false),
			domain.NewSpadesPlayer(false),
		}
		spades := domain.NewSpades(domain.NewTrumpCards(0), players, config)
		return usecase.NewSpadesInteractor(spades, new(presenter.SpadesWebPresenter))
	}))
	web.register("crazyeights", controller.NewCrazyEightsWebController(func() usecase.CrazyEightsInteractorIF {
		config := domain.DefaultCrazyEightsConfig()
		players := []*domain.CrazyEightsPlayer{
			domain.NewCrazyEightsPlayer(true),
			domain.NewCrazyEightsPlayer(false),
			domain.NewCrazyEightsPlayer(false),
			domain.NewCrazyEightsPlayer(false),
		}
		ce := domain.NewCrazyEights(domain.NewTrumpCards(0), players, config)
		return usecase.NewCrazyEightsInteractor(ce, new(presenter.CrazyEightsWebPresenter))
	}))
	web.register("ginrummy", controller.NewGinRummyWebController(func() usecase.GinRummyInteractorIF {
		config := domain.DefaultGinRummyConfig()
		players := []*domain.GinRummyPlayer{
			domain.NewGinRummyPlayer(true),
			domain.NewGinRummyPlayer(false),
		}
		gr := domain.NewGinRummy(domain.NewTrumpCards(0), players, config)
		return usecase.NewGinRummyInteractor(gr, new(presenter.GinRummyWebPresenter))
	}))
	web.register("spider", controller.NewSpiderWebController(func() usecase.SpiderInteractorIF {
		spider := domain.NewSpider(domain.NewTrumpCardsWithSuits(domain.SpiderTotalCards, []int{domain.CardDesignSpade}))
		return usecase.NewSpiderInteractor(spider, new(presenter.SpiderWebPresenter))
	}))
	web.register("napoleon", controller.NewNapoleonWebController(func() usecase.NapoleonInteractorIF {
		config := domain.DefaultNapoleonConfig()
		players := []*domain.NapoleonPlayer{
			domain.NewNapoleonPlayer(true),
			domain.NewNapoleonPlayer(false),
			domain.NewNapoleonPlayer(false),
			domain.NewNapoleonPlayer(false),
		}
		napoleon := domain.NewNapoleon(domain.NewTrumpCards(1), players, config)
		return usecase.NewNapoleonInteractor(napoleon, new(presenter.NapoleonWebPresenter))
	}))
	web.register("indianpoker", controller.NewIndianPokerWebController(func() usecase.IndianPokerInteractorIF {
		cfg := domain.DefaultIndianPokerConfig()
		ip := domain.NewIndianPoker(domain.NewTrumpCards(0), domain.NewIndianPokerPlayers(), cfg)
		return usecase.NewIndianPokerInteractor(ip, new(presenter.IndianPokerWebPresenter))
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
		config := domain.DefaultEuchreConfig()
		players := []*domain.EuchrePlayer{
			domain.NewEuchrePlayer(true, 0),
			domain.NewEuchrePlayer(false, 1),
			domain.NewEuchrePlayer(false, 0),
			domain.NewEuchrePlayer(false, 1),
		}
		euchre := domain.NewEuchre(domain.NewTrumpCardsEuchre(), players, config)
		return usecase.NewEuchreInteractor(euchre, new(presenter.EuchreWebPresenter))
	}))
	web.register("pyramid", controller.NewPyramidWebController(func() usecase.PyramidInteractorIF {
		pyramid := domain.NewPyramid(domain.NewTrumpCards(0))
		return usecase.NewPyramidInteractor(pyramid, new(presenter.PyramidWebPresenter))
	}))
	web.register("tripeaks", controller.NewTriPeaksWebController(func() usecase.TriPeaksInteractorIF {
		triPeaks := domain.NewTriPeaks(domain.NewTrumpCards(0))
		return usecase.NewTriPeaksInteractor(triPeaks, new(presenter.TriPeaksWebPresenter))
	}))
	web.register("cribbage", controller.NewCribbageWebController(func() usecase.CribbageInteractorIF {
		config := domain.DefaultCribbageConfig()
		players := []*domain.CribbagePlayer{
			domain.NewCribbagePlayer(true),
			domain.NewCribbagePlayer(false),
		}
		cribbage := domain.NewCribbage(domain.NewTrumpCards(0), players, config)
		return usecase.NewCribbageInteractor(cribbage, new(presenter.CribbageWebPresenter))
	}))
	web.register("threecard", controller.NewThreeCardWebController(func() usecase.ThreeCardInteractorIF {
		return usecase.NewThreeCardInteractor(
			domain.NewDefaultThreeCard(),
			new(presenter.ThreeCardWebPresenter),
		)
	}))
	web.register("ohhell", controller.NewOhHellWebController(func() usecase.OhHellInteractorIF {
		config := domain.DefaultOhHellConfig()
		players := []*domain.OhHellPlayer{
			domain.NewOhHellPlayer(true),
			domain.NewOhHellPlayer(false),
			domain.NewOhHellPlayer(false),
			domain.NewOhHellPlayer(false),
		}
		ohHell := domain.NewOhHell(domain.NewTrumpCards(0), players, config)
		return usecase.NewOhHellInteractor(ohHell, new(presenter.OhHellWebPresenter))
	}))
	web.register("bridge", controller.NewBridgeWebController(func() usecase.BridgeInteractorIF {
		config := domain.DefaultBridgeConfig()
		players := []*domain.BridgePlayer{
			domain.NewBridgePlayer(true, 0),
			domain.NewBridgePlayer(false, 1),
			domain.NewBridgePlayer(false, 0),
			domain.NewBridgePlayer(false, 1),
		}
		bridge := domain.NewBridge(domain.NewTrumpCards(0), players, config)
		return usecase.NewBridgeInteractor(bridge, new(presenter.BridgeWebPresenter))
	}))
	web.register("pineapple", controller.NewPineappleWebController(func() usecase.PineappleInteractorIF {
		cfg := domain.DefaultPineappleConfig()
		pineapple := domain.NewPineapple(domain.NewTrumpCards(0), domain.NewPineapplePlayersForTable(cfg.TableSize), cfg)
		return usecase.NewPineappleInteractor(pineapple, new(presenter.PineappleWebPresenter))
	}))
	web.register("speed", controller.NewSpeedWebController(func() usecase.SpeedInteractorIF {
		config := domain.DefaultSpeedConfig()
		players := []*domain.SpeedPlayer{
			domain.NewSpeedPlayer(true),
			domain.NewSpeedPlayer(false),
		}
		speed := domain.NewSpeed(domain.NewTrumpCards(0), players, config)
		return usecase.NewSpeedInteractor(speed, new(presenter.SpeedWebPresenter))
	}))
	web.register("gofish", controller.NewGoFishWebController(func() usecase.GoFishInteractorIF {
		players := []*domain.GoFishPlayer{
			domain.NewGoFishPlayer(true),
			domain.NewGoFishPlayer(false),
			domain.NewGoFishPlayer(false),
			domain.NewGoFishPlayer(false),
		}
		goFish := domain.NewGoFish(domain.NewTrumpCards(0), players)
		return usecase.NewGoFishInteractor(goFish, new(presenter.GoFishWebPresenter))
	}))
	web.register("canasta", controller.NewCanastaWebController(func() usecase.CanastaInteractorIF {
		config := domain.DefaultCanastaConfig()
		players := []*domain.CanastaPlayer{
			domain.NewCanastaPlayer(true),
			domain.NewCanastaPlayer(false),
		}
		canasta := domain.NewCanasta(domain.NewTrumpCardsWithDecks(2, 4), players, config)
		return usecase.NewCanastaInteractor(canasta, new(presenter.CanastaWebPresenter))
	}))
	web.register("pinochle", controller.NewPinochleWebController(func() usecase.PinochleInteractorIF {
		config := domain.DefaultPinochleConfig()
		players := []*domain.PinochlePlayer{
			domain.NewPinochlePlayer(true, 0),
			domain.NewPinochlePlayer(false, 1),
			domain.NewPinochlePlayer(false, 0),
			domain.NewPinochlePlayer(false, 1),
		}
		pinochle := domain.NewPinochle(domain.NewTrumpCardsPinochle(), players, config)
		return usecase.NewPinochleInteractor(pinochle, new(presenter.PinochleWebPresenter))
	}))
	web.register("golf", controller.NewGolfWebController(func() usecase.GolfInteractorIF {
		golf := domain.NewGolf(domain.NewTrumpCards(0))
		return usecase.NewGolfInteractor(golf, new(presenter.GolfWebPresenter))
	}))
	web.register("pigtail", controller.NewPigsTailWebController(func() usecase.PigsTailInteractorIF {
		players := []*domain.PigsTailPlayer{
			domain.NewPigsTailPlayer(true),
			domain.NewPigsTailPlayer(false),
			domain.NewPigsTailPlayer(false),
			domain.NewPigsTailPlayer(false),
		}
		pigsTail := domain.NewPigsTail(domain.NewTrumpCards(0), players)
		return usecase.NewPigsTailInteractor(pigsTail, new(presenter.PigsTailWebPresenter))
	}))
	web.register("sevencardstud", controller.NewSevenCardStudWebController(func() usecase.SevenCardStudInteractorIF {
		cfg := domain.DefaultSevenCardStudConfig()
		scs := domain.NewSevenCardStud(domain.NewTrumpCards(0), domain.NewSevenCardStudPlayersForTable(cfg.TableSize), cfg)
		return usecase.NewSevenCardStudInteractor(scs, new(presenter.SevenCardStudWebPresenter))
	}))
	web.register("clocksolitaire", controller.NewClockSolitaireWebController(func() usecase.ClockSolitaireInteractorIF {
		cs := domain.NewClockSolitaire(domain.NewTrumpCards(0))
		return usecase.NewClockSolitaireInteractor(cs, new(presenter.ClockSolitaireWebPresenter))
	}))
	web.register("durak", controller.NewDurakWebController(func() usecase.DurakInteractorIF {
		players := []*domain.DurakPlayer{
			domain.NewDurakPlayer(true),
			domain.NewDurakPlayer(false),
			domain.NewDurakPlayer(false),
			domain.NewDurakPlayer(false),
		}
		d := domain.NewDurak(domain.NewTrumpCardsShortDeck(), players)
		return usecase.NewDurakInteractor(d, new(presenter.DurakWebPresenter))
	}))
	web.register("fortythieves", controller.NewFortyThievesWebController(func() usecase.FortyThievesInteractorIF {
		ft := domain.NewFortyThieves(domain.NewTrumpCardsWithDecks(2, 0))
		return usecase.NewFortyThievesInteractor(ft, new(presenter.FortyThievesWebPresenter))
	}))
	web.register("paigow", controller.NewPaiGowWebController(func() usecase.PaiGowInteractorIF {
		return usecase.NewPaiGowInteractor(
			domain.NewDefaultPaiGow(),
			new(presenter.PaiGowWebPresenter),
		)
	}))
	web.register("twotenjack", controller.NewTwoTenJackWebController(func() usecase.TwoTenJackInteractorIF {
		config := domain.DefaultTwoTenJackConfig()
		players := []*domain.TwoTenJackPlayer{
			domain.NewTwoTenJackPlayer(true),
			domain.NewTwoTenJackPlayer(false),
			domain.NewTwoTenJackPlayer(false),
			domain.NewTwoTenJackPlayer(false),
		}
		ttj := domain.NewTwoTenJack(domain.NewTrumpCards(0), players, config)
		return usecase.NewTwoTenJackInteractor(ttj, new(presenter.TwoTenJackWebPresenter))
	}))
	web.register("caribbeanstud", controller.NewCaribbeanStudWebController(func() usecase.CaribbeanStudInteractorIF {
		return usecase.NewCaribbeanStudInteractor(
			domain.NewDefaultCaribbeanStud(),
			new(presenter.CaribbeanStudWebPresenter),
		)
	}))
	web.register("war", controller.NewWarWebController(func() usecase.WarInteractorIF {
		config := domain.DefaultWarConfig()
		players := []*domain.WarPlayer{
			domain.NewWarPlayer(true),
			domain.NewWarPlayer(false),
		}
		war := domain.NewWar(domain.NewTrumpCards(0), players, config)
		return usecase.NewWarInteractor(war, new(presenter.WarWebPresenter))
	}))
	web.register("canfield", controller.NewCanfieldWebController(func() usecase.CanfieldInteractorIF {
		canfield := domain.NewCanfield(domain.NewTrumpCards(0))
		return usecase.NewCanfieldInteractor(canfield, new(presenter.CanfieldWebPresenter))
	}))
	web.register("fiftyone", controller.NewFiftyOneWebController(func() usecase.FiftyOneInteractorIF {
		players := []*domain.FiftyOnePlayer{
			domain.NewFiftyOnePlayer(true),
			domain.NewFiftyOnePlayer(false),
			domain.NewFiftyOnePlayer(false),
			domain.NewFiftyOnePlayer(false),
		}
		fo := domain.NewFiftyOne(domain.NewTrumpCards(0), players)
		return usecase.NewFiftyOneInteractor(fo, new(presenter.FiftyOneWebPresenter))
	}))
	web.register("yukon", controller.NewYukonWebController(func() usecase.YukonInteractorIF {
		yukon := domain.NewYukon(domain.NewTrumpCards(0))
		return usecase.NewYukonInteractor(yukon, new(presenter.YukonWebPresenter))
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
		Addr:         getListenPort(),
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
	port := ln.Addr().(*net.TCPAddr).Port
	fmt.Println(i18n.Tf("webServerRunning", "port", strconv.Itoa(port)))
	fmt.Println(i18n.T("webServerStop"))

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
		fmt.Println("\n" + i18n.T("webServerShutdown"))
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
	fmt.Println(i18n.T("webServerStopped"))
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
