//go:build js && wasm

// Package extra6 reserves the ninth Cloudflare Worker size bucket.
package extra6

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func init() {
	games.RegisterKVGame("bourre", games.CategoryExtra6,
		func() usecase.BourreInteractorIF {
			return usecase.NewBourreInteractor(domain.NewDefaultBourre(), new(presenter.BourreWebPresenter))
		},
		func(data []byte) (usecase.BourreInteractorIF, error) {
			return usecase.RestoreBourreInteractor(data, new(presenter.BourreWebPresenter))
		},
		controller.NewBourreWebControllerWithProvider)
	games.RegisterKVGame("doppelkopf", games.CategoryExtra6,
		func() usecase.DoppelkopfInteractorIF {
			return usecase.NewDoppelkopfInteractor(domain.NewDefaultDoppelkopf(), new(presenter.DoppelkopfWebPresenter))
		},
		func(data []byte) (usecase.DoppelkopfInteractorIF, error) {
			return usecase.RestoreDoppelkopfInteractor(data, new(presenter.DoppelkopfWebPresenter))
		},
		controller.NewDoppelkopfWebControllerWithProvider)
	games.RegisterKVGame("ecarte", games.CategoryExtra6,
		func() usecase.EcarteInteractorIF {
			return usecase.NewEcarteInteractor(domain.NewDefaultEcarte(), new(presenter.EcarteWebPresenter))
		},
		func(data []byte) (usecase.EcarteInteractorIF, error) {
			return usecase.RestoreEcarteInteractor(data, new(presenter.EcarteWebPresenter))
		},
		controller.NewEcarteWebControllerWithProvider)
	games.RegisterKVGame("mus", games.CategoryExtra6,
		func() usecase.MusInteractorIF {
			return usecase.NewMusInteractor(domain.NewDefaultMus(), new(presenter.MusWebPresenter))
		},
		func(data []byte) (usecase.MusInteractorIF, error) {
			return usecase.RestoreMusInteractor(data, new(presenter.MusWebPresenter))
		},
		controller.NewMusWebControllerWithProvider)
	games.RegisterKVGame("tressette", games.CategoryExtra6,
		func() usecase.TressetteInteractorIF {
			return usecase.NewTressetteInteractor(domain.NewDefaultTressette(), new(presenter.TressetteWebPresenter))
		},
		func(data []byte) (usecase.TressetteInteractorIF, error) {
			return usecase.RestoreTressetteInteractor(data, new(presenter.TressetteWebPresenter))
		},
		controller.NewTressetteWebControllerWithProvider)
	games.RegisterKVGame("sueca", games.CategoryExtra6,
		func() usecase.SuecaInteractorIF {
			return usecase.NewSuecaInteractor(domain.NewDefaultSueca(), new(presenter.SuecaWebPresenter))
		},
		func(data []byte) (usecase.SuecaInteractorIF, error) {
			return usecase.RestoreSuecaInteractor(data, new(presenter.SuecaWebPresenter))
		},
		controller.NewSuecaWebControllerWithProvider)
	games.RegisterKVGame("fortyfives", games.CategoryExtra6,
		func() usecase.FortyFivesInteractorIF {
			return usecase.NewFortyFivesInteractor(domain.NewDefaultFortyFives(), new(presenter.FortyFivesWebPresenter))
		},
		func(data []byte) (usecase.FortyFivesInteractorIF, error) {
			return usecase.RestoreFortyFivesInteractor(data, new(presenter.FortyFivesWebPresenter))
		},
		controller.NewFortyFivesWebControllerWithProvider)
	games.RegisterKVGame("twentynine", games.CategoryExtra6,
		func() usecase.TwentyNineInteractorIF {
			return usecase.NewTwentyNineInteractor(domain.NewDefaultTwentyNine(), new(presenter.TwentyNineWebPresenter))
		},
		func(data []byte) (usecase.TwentyNineInteractorIF, error) {
			return usecase.RestoreTwentyNineInteractor(data, new(presenter.TwentyNineWebPresenter))
		},
		controller.NewTwentyNineWebControllerWithProvider)
	games.RegisterKVGame("tute", games.CategoryExtra6,
		func() usecase.TuteInteractorIF {
			return usecase.NewTuteInteractor(domain.NewDefaultTute(), new(presenter.TuteWebPresenter))
		},
		func(data []byte) (usecase.TuteInteractorIF, error) {
			return usecase.RestoreTuteInteractor(data, new(presenter.TuteWebPresenter))
		},
		controller.NewTuteWebControllerWithProvider)
	games.RegisterKVGame("montebank", games.CategoryExtra6,
		func() usecase.MonteBankInteractorIF {
			return usecase.NewMonteBankInteractor(domain.NewDefaultMonteBank(), new(presenter.MonteBankWebPresenter))
		},
		func(data []byte) (usecase.MonteBankInteractorIF, error) {
			return usecase.RestoreMonteBankInteractor(data, new(presenter.MonteBankWebPresenter))
		},
		controller.NewMonteBankWebControllerWithProvider)
	games.RegisterKVGame("tarneeb", games.CategoryExtra6,
		func() usecase.TarneebInteractorIF {
			return usecase.NewTarneebInteractor(domain.NewDefaultTarneeb(), new(presenter.TarneebWebPresenter))
		},
		func(data []byte) (usecase.TarneebInteractorIF, error) {
			return usecase.RestoreTarneebInteractor(data, new(presenter.TarneebWebPresenter))
		},
		controller.NewTarneebWebControllerWithProvider)
	games.RegisterKVGame("courtpiece", games.CategoryExtra6,
		func() usecase.CourtPieceInteractorIF {
			return usecase.NewCourtPieceInteractor(domain.NewDefaultCourtPiece(), new(presenter.CourtPieceWebPresenter))
		},
		func(data []byte) (usecase.CourtPieceInteractorIF, error) {
			return usecase.RestoreCourtPieceInteractor(data, new(presenter.CourtPieceWebPresenter))
		},
		controller.NewCourtPieceWebControllerWithProvider)
	games.RegisterKVGame("caribbeandraw", games.CategoryExtra6,
		func() usecase.CaribbeanDrawInteractorIF {
			return usecase.NewCaribbeanDrawInteractor(domain.NewDefaultCaribbeanDraw(), new(presenter.CaribbeanDrawWebPresenter))
		},
		func(data []byte) (usecase.CaribbeanDrawInteractorIF, error) {
			return usecase.RestoreCaribbeanDrawInteractor(data, new(presenter.CaribbeanDrawWebPresenter))
		},
		controller.NewCaribbeanDrawWebControllerWithProvider)
	games.RegisterKVGame("oasispoker", games.CategoryExtra6,
		func() usecase.OasisPokerInteractorIF {
			return usecase.NewOasisPokerInteractor(domain.NewDefaultOasisPoker(), new(presenter.OasisPokerWebPresenter))
		},
		func(data []byte) (usecase.OasisPokerInteractorIF, error) {
			return usecase.RestoreOasisPokerInteractor(data, new(presenter.OasisPokerWebPresenter))
		},
		controller.NewOasisPokerWebControllerWithProvider)
	games.RegisterKVGame("paigow", games.CategoryExtra6,
		func() usecase.PaiGowInteractorIF {
			return usecase.NewPaiGowInteractor(domain.NewDefaultPaiGow(), new(presenter.PaiGowWebPresenter))
		},
		func(data []byte) (usecase.PaiGowInteractorIF, error) {
			return usecase.RestorePaiGowInteractor(data, new(presenter.PaiGowWebPresenter))
		},
		controller.NewPaiGowWebControllerWithProvider)
	games.RegisterKVGame("mississippistud", games.CategoryExtra6,
		func() usecase.MississippiStudInteractorIF {
			return usecase.NewMississippiStudInteractor(domain.NewDefaultMississippiStud(), new(presenter.MississippiStudWebPresenter))
		},
		func(data []byte) (usecase.MississippiStudInteractorIF, error) {
			return usecase.RestoreMississippiStudInteractor(data, new(presenter.MississippiStudWebPresenter))
		},
		controller.NewMississippiStudWebControllerWithProvider)
	games.RegisterKVGame("daifugo", games.CategoryExtra6,
		func() usecase.DaifugoInteractorIF {
			return usecase.NewDaifugoInteractor(domain.NewDefaultDaifugo(), new(presenter.DaifugoWebPresenter))
		},
		func(data []byte) (usecase.DaifugoInteractorIF, error) {
			return usecase.RestoreDaifugoInteractor(data, new(presenter.DaifugoWebPresenter))
		},
		controller.NewDaifugoWebControllerWithProvider)
	games.RegisterKVGame("costlycolours", games.CategoryExtra6,
		func() usecase.CostlyColoursInteractorIF {
			return usecase.NewCostlyColoursInteractor(domain.NewDefaultCostlyColours(), new(presenter.CostlyColoursWebPresenter))
		},
		func(data []byte) (usecase.CostlyColoursInteractorIF, error) {
			return usecase.RestoreCostlyColoursInteractor(data, new(presenter.CostlyColoursWebPresenter))
		},
		controller.NewCostlyColoursWebControllerWithProvider)
	games.RegisterKVGame("cribbage", games.CategoryExtra6,
		func() usecase.CribbageInteractorIF {
			return usecase.NewCribbageInteractor(domain.NewDefaultCribbage(), new(presenter.CribbageWebPresenter))
		},
		func(data []byte) (usecase.CribbageInteractorIF, error) {
			return usecase.RestoreCribbageInteractor(data, new(presenter.CribbageWebPresenter))
		},
		controller.NewCribbageWebControllerWithProvider)
	games.RegisterKVGame("marjapussi", games.CategoryExtra6,
		func() usecase.MarjapussiInteractorIF {
			return usecase.NewMarjapussiInteractor(domain.NewDefaultMarjapussi(), new(presenter.MarjapussiWebPresenter))
		},
		func(data []byte) (usecase.MarjapussiInteractorIF, error) {
			return usecase.RestoreMarjapussiInteractor(data, new(presenter.MarjapussiWebPresenter))
		},
		controller.NewMarjapussiWebControllerWithProvider)
	games.RegisterKVGame("marriage", games.CategoryExtra6,
		func() usecase.MarriageInteractorIF {
			return usecase.NewMarriageInteractor(domain.NewDefaultMarriage(), new(presenter.MarriageWebPresenter))
		},
		func(data []byte) (usecase.MarriageInteractorIF, error) {
			return usecase.RestoreMarriageInteractor(data, new(presenter.MarriageWebPresenter))
		},
		controller.NewMarriageWebControllerWithProvider)
	games.RegisterKVGame("kingo", games.CategoryExtra6,
		func() usecase.KingoInteractorIF {
			return usecase.NewKingoInteractor(domain.NewDefaultKingo(), new(presenter.KingoWebPresenter))
		},
		func(data []byte) (usecase.KingoInteractorIF, error) {
			return usecase.RestoreKingoInteractor(data, new(presenter.KingoWebPresenter))
		},
		controller.NewKingoWebControllerWithProvider)
	games.RegisterKVGame("oichokabu", games.CategoryExtra6,
		func() usecase.OichoKabuInteractorIF {
			return usecase.NewOichoKabuInteractor(domain.NewDefaultOichoKabu(), new(presenter.OichoKabuWebPresenter))
		},
		func(data []byte) (usecase.OichoKabuInteractorIF, error) {
			return usecase.RestoreOichoKabuInteractor(data, new(presenter.OichoKabuWebPresenter))
		},
		controller.NewOichoKabuWebControllerWithProvider)
	games.RegisterKVGame("tusac", games.CategoryExtra6,
		func() usecase.TuSacInteractorIF {
			return usecase.NewTuSacInteractor(domain.NewDefaultTuSac(), new(presenter.TuSacWebPresenter))
		},
		func(data []byte) (usecase.TuSacInteractorIF, error) {
			return usecase.RestoreTuSacInteractor(data, new(presenter.TuSacWebPresenter))
		},
		controller.NewTuSacWebControllerWithProvider)
	games.RegisterKVGame("mighty", games.CategoryExtra6,
		func() usecase.MightyInteractorIF {
			return usecase.NewMightyInteractor(domain.NewDefaultMighty(), new(presenter.MightyWebPresenter))
		},
		func(data []byte) (usecase.MightyInteractorIF, error) {
			return usecase.RestoreMightyInteractor(data, new(presenter.MightyWebPresenter))
		},
		controller.NewMightyWebControllerWithProvider)
	games.RegisterKVGame("chineseten", games.CategoryExtra6,
		func() usecase.ChineseTenInteractorIF {
			return usecase.NewChineseTenInteractor(domain.NewDefaultChineseTen(), new(presenter.ChineseTenWebPresenter))
		},
		func(data []byte) (usecase.ChineseTenInteractorIF, error) {
			return usecase.RestoreChineseTenInteractor(data, new(presenter.ChineseTenWebPresenter))
		},
		controller.NewChineseTenWebControllerWithProvider)
	games.RegisterKVGame("cuckoo", games.CategoryExtra6,
		func() usecase.CuckooInteractorIF {
			return usecase.NewCuckooInteractor(domain.NewDefaultCuckoo(), new(presenter.CuckooWebPresenter))
		},
		func(data []byte) (usecase.CuckooInteractorIF, error) {
			return usecase.RestoreCuckooInteractor(data, new(presenter.CuckooWebPresenter))
		},
		controller.NewCuckooWebControllerWithProvider)
	games.RegisterKVGame("mushi", games.CategoryExtra6,
		func() usecase.MushiInteractorIF {
			return usecase.NewMushiInteractor(domain.NewDefaultMushi(), new(presenter.MushiWebPresenter))
		},
		func(data []byte) (usecase.MushiInteractorIF, error) {
			return usecase.RestoreMushiInteractor(data, new(presenter.MushiWebPresenter))
		},
		controller.NewMushiWebControllerWithProvider)
	games.RegisterKVGame("pontoon", games.CategoryExtra6,
		func() usecase.PontoonInteractorIF {
			return usecase.NewPontoonInteractor(domain.NewDefaultPontoon(), new(presenter.PontoonWebPresenter))
		},
		func(data []byte) (usecase.PontoonInteractorIF, error) {
			return usecase.RestorePontoonInteractor(data, new(presenter.PontoonWebPresenter))
		},
		controller.NewPontoonWebControllerWithProvider)
	games.RegisterKVGame("quinze", games.CategoryExtra6,
		func() usecase.QuinzeInteractorIF {
			return usecase.NewQuinzeInteractor(domain.NewDefaultQuinze(), new(presenter.QuinzeWebPresenter))
		},
		func(data []byte) (usecase.QuinzeInteractorIF, error) {
			return usecase.RestoreQuinzeInteractor(data, new(presenter.QuinzeWebPresenter))
		},
		controller.NewQuinzeWebControllerWithProvider)
	games.RegisterKVGame("settemezzo", games.CategoryExtra6,
		func() usecase.SetteEMezzoInteractorIF {
			return usecase.NewSetteEMezzoInteractor(domain.NewDefaultSetteEMezzo(), new(presenter.SetteEMezzoWebPresenter))
		},
		func(data []byte) (usecase.SetteEMezzoInteractorIF, error) {
			return usecase.RestoreSetteEMezzoInteractor(data, new(presenter.SetteEMezzoWebPresenter))
		},
		controller.NewSetteEMezzoWebControllerWithProvider)
	games.RegisterKVGame("pig", games.CategoryExtra6,
		func() usecase.PigInteractorIF {
			return usecase.NewPigInteractor(domain.NewDefaultPig(), new(presenter.PigWebPresenter))
		},
		func(data []byte) (usecase.PigInteractorIF, error) {
			return usecase.RestorePigInteractor(data, new(presenter.PigWebPresenter))
		},
		controller.NewPigWebControllerWithProvider)
	games.RegisterKVGame("spoons", games.CategoryExtra6,
		func() usecase.SpoonsInteractorIF {
			return usecase.NewSpoonsInteractor(domain.NewDefaultSpoons(), new(presenter.SpoonsWebPresenter))
		},
		func(data []byte) (usecase.SpoonsInteractorIF, error) {
			return usecase.RestoreSpoonsInteractor(data, new(presenter.SpoonsWebPresenter))
		},
		controller.NewSpoonsWebControllerWithProvider)
	games.RegisterKVGame("faro", games.CategoryExtra6,
		func() usecase.FaroInteractorIF {
			return usecase.NewFaroInteractor(domain.NewDefaultFaro(), new(presenter.FaroWebPresenter))
		},
		func(data []byte) (usecase.FaroInteractorIF, error) {
			return usecase.RestoreFaroInteractor(data, new(presenter.FaroWebPresenter))
		},
		controller.NewFaroWebControllerWithProvider)
	games.RegisterKVGame("sutda", games.CategoryExtra6,
		func() usecase.SutdaInteractorIF {
			return usecase.NewSutdaInteractor(domain.NewDefaultSutda(), new(presenter.SutdaWebPresenter))
		},
		func(data []byte) (usecase.SutdaInteractorIF, error) {
			return usecase.RestoreSutdaInteractor(data, new(presenter.SutdaWebPresenter))
		},
		controller.NewSutdaWebControllerWithProvider)
	games.RegisterKVGame("tichu", games.CategoryExtra6,
		func() usecase.TichuInteractorIF {
			return usecase.NewTichuInteractor(domain.NewDefaultTichu(), new(presenter.TichuWebPresenter))
		},
		func(data []byte) (usecase.TichuInteractorIF, error) {
			return usecase.RestoreTichuInteractor(data, new(presenter.TichuWebPresenter))
		},
		controller.NewTichuWebControllerWithProvider)
	games.RegisterKVGame("belote", games.CategoryExtra6,
		func() usecase.BeloteInteractorIF {
			return usecase.NewBeloteInteractor(domain.NewDefaultBelote(), new(presenter.BeloteWebPresenter))
		},
		func(data []byte) (usecase.BeloteInteractorIF, error) {
			return usecase.RestoreBeloteInteractor(data, new(presenter.BeloteWebPresenter))
		},
		controller.NewBeloteWebControllerWithProvider)
	games.RegisterKVGame("coinche", games.CategoryExtra6,
		func() usecase.CoincheInteractorIF {
			return usecase.NewCoincheInteractor(domain.NewDefaultCoinche(), new(presenter.CoincheWebPresenter))
		},
		func(data []byte) (usecase.CoincheInteractorIF, error) {
			return usecase.RestoreCoincheInteractor(data, new(presenter.CoincheWebPresenter))
		},
		controller.NewCoincheWebControllerWithProvider)
	games.RegisterKVGame("jass", games.CategoryExtra6,
		func() usecase.JassInteractorIF {
			return usecase.NewJassInteractor(domain.NewDefaultJass(), new(presenter.JassWebPresenter))
		},
		func(data []byte) (usecase.JassInteractorIF, error) {
			return usecase.RestoreJassInteractor(data, new(presenter.JassWebPresenter))
		},
		controller.NewJassWebControllerWithProvider)
	games.RegisterKVGame("klaberjass", games.CategoryExtra6,
		func() usecase.KlaberjassInteractorIF {
			return usecase.NewKlaberjassInteractor(domain.NewDefaultKlaberjass(), new(presenter.KlaberjassWebPresenter))
		},
		func(data []byte) (usecase.KlaberjassInteractorIF, error) {
			return usecase.RestoreKlaberjassInteractor(data, new(presenter.KlaberjassWebPresenter))
		},
		controller.NewKlaberjassWebControllerWithProvider)
	games.RegisterKVGame("ramsch", games.CategoryExtra6,
		func() usecase.RamschInteractorIF {
			return usecase.NewRamschInteractor(domain.NewDefaultRamsch(), new(presenter.RamschWebPresenter))
		},
		func(data []byte) (usecase.RamschInteractorIF, error) {
			return usecase.RestoreRamschInteractor(data, new(presenter.RamschWebPresenter))
		},
		controller.NewRamschWebControllerWithProvider)
	games.RegisterKVGame("skat", games.CategoryExtra6,
		func() usecase.SkatInteractorIF {
			return usecase.NewSkatInteractor(domain.NewDefaultSkat(), new(presenter.SkatWebPresenter))
		},
		func(data []byte) (usecase.SkatInteractorIF, error) {
			return usecase.RestoreSkatInteractor(data, new(presenter.SkatWebPresenter))
		},
		controller.NewSkatWebControllerWithProvider)
	games.RegisterKVGame("tarabish", games.CategoryExtra6,
		func() usecase.TarabishInteractorIF {
			return usecase.NewTarabishInteractor(domain.NewDefaultTarabish(), new(presenter.TarabishWebPresenter))
		},
		func(data []byte) (usecase.TarabishInteractorIF, error) {
			return usecase.RestoreTarabishInteractor(data, new(presenter.TarabishWebPresenter))
		},
		controller.NewTarabishWebControllerWithProvider)
}
