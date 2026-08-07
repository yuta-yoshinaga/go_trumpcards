//go:build test

package usecase

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// --- stub presenters ---

// stubBlackJackPresenter implements presenter.BlackJackPresenter (= GamePresenter[interfaces.BlackJackGame])
type stubBlackJackPresenter struct{}

func (s *stubBlackJackPresenter) Output(_ interfaces.BlackJackGame, _ error) string { return `{}` }
func (s *stubBlackJackPresenter) ActionLogOutput(_ interfaces.BlackJackGame) string { return `{}` }

// stubPokerPresenter implements presenter.PokerPresenter (GamePresenter + OutputWithOdds)
type stubPokerPresenter struct{}

func (s *stubPokerPresenter) Output(_ interfaces.PokerGame, _ error) string { return `{}` }
func (s *stubPokerPresenter) ActionLogOutput(_ interfaces.PokerGame) string { return `{}` }
func (s *stubPokerPresenter) OutputWithOdds(_ interfaces.PokerGame, _ error, _ []domain.PokerDrawOdds) string {
	return `{}`
}

// stubHoldemPresenter implements presenter.HoldemPresenter (= GamePresenter[interfaces.HoldemGame])
type stubHoldemPresenter struct{}

func (s *stubHoldemPresenter) Output(_ interfaces.HoldemGame, _ error) string { return `{}` }
func (s *stubHoldemPresenter) ActionLogOutput(_ interfaces.HoldemGame) string { return `{}` }

// stubOmahaPresenter implements presenter.OmahaPresenter (= GamePresenter[interfaces.OmahaGame])
type stubOmahaPresenter struct{}

func (s *stubOmahaPresenter) Output(_ interfaces.OmahaGame, _ error) string { return `{}` }
func (s *stubOmahaPresenter) ActionLogOutput(_ interfaces.OmahaGame) string { return `{}` }

// stubShortDeckPresenter implements presenter.ShortDeckPresenter (= GamePresenter[interfaces.ShortDeckGame])
type stubShortDeckPresenter struct{}

func (s *stubShortDeckPresenter) Output(_ interfaces.ShortDeckGame, _ error) string { return `{}` }
func (s *stubShortDeckPresenter) ActionLogOutput(_ interfaces.ShortDeckGame) string { return `{}` }

// stubIndianPokerPresenter implements presenter.IndianPokerPresenter (= GamePresenter[interfaces.IndianPokerGame])
type stubIndianPokerPresenter struct{}

func (s *stubIndianPokerPresenter) Output(_ interfaces.IndianPokerGame, _ error) string {
	return `{}`
}
func (s *stubIndianPokerPresenter) ActionLogOutput(_ interfaces.IndianPokerGame) string {
	return `{}`
}

// stubVideoPokerPresenter implements presenter.VideoPokerPresenter (= GamePresenter[interfaces.VideoPokerGame])
type stubVideoPokerPresenter struct{}

func (s *stubVideoPokerPresenter) Output(_ interfaces.VideoPokerGame, _ error) string { return `{}` }
func (s *stubVideoPokerPresenter) ActionLogOutput(_ interfaces.VideoPokerGame) string { return `{}` }
func (s *stubVideoPokerPresenter) HintOutput(_ interfaces.VideoPokerGame) string      { return `{}` }

// stubHeartsPresenter implements presenter.HeartsPresenter (GamePresenter + HintOutput)
type stubHeartsPresenter struct{}

func (s *stubHeartsPresenter) Output(_ interfaces.HeartsGame, _ error) string { return `{}` }
func (s *stubHeartsPresenter) ActionLogOutput(_ interfaces.HeartsGame) string { return `{}` }
func (s *stubHeartsPresenter) HintOutput(_ interfaces.HeartsGame) string      { return `{}` }

// stubSpadesPresenter implements presenter.SpadesPresenter (GamePresenter + HintOutput)
type stubSpadesPresenter struct{}

func (s *stubSpadesPresenter) Output(_ interfaces.SpadesGame, _ error) string { return `{}` }
func (s *stubSpadesPresenter) ActionLogOutput(_ interfaces.SpadesGame) string { return `{}` }
func (s *stubSpadesPresenter) HintOutput(_ interfaces.SpadesGame) string      { return `{}` }

// stubEuchrePresenter implements presenter.EuchrePresenter (GamePresenter + HintOutput)
type stubEuchrePresenter struct{}

func (s *stubEuchrePresenter) Output(_ interfaces.EuchreGame, _ error) string { return `{}` }
func (s *stubEuchrePresenter) ActionLogOutput(_ interfaces.EuchreGame) string { return `{}` }
func (s *stubEuchrePresenter) HintOutput(_ interfaces.EuchreGame) string      { return `{}` }

// stubNapoleonPresenter implements presenter.NapoleonPresenter (GamePresenter + HintOutput)
type stubNapoleonPresenter struct{}

func (s *stubNapoleonPresenter) Output(_ interfaces.NapoleonGame, _ error) string { return `{}` }
func (s *stubNapoleonPresenter) ActionLogOutput(_ interfaces.NapoleonGame) string { return `{}` }
func (s *stubNapoleonPresenter) HintOutput(_ interfaces.NapoleonGame) string      { return `{}` }

// stubOldMaidPresenter implements presenter.OldMaidPresenter (= GamePresenter[interfaces.OldMaidGame])
type stubOldMaidPresenter struct{}

func (s *stubOldMaidPresenter) Output(_ interfaces.OldMaidGame, _ error) string { return `{}` }
func (s *stubOldMaidPresenter) ActionLogOutput(_ interfaces.OldMaidGame) string { return `{}` }

// stubDoubtPresenter implements presenter.DoubtPresenter (= GamePresenter[interfaces.DoubtGame])
type stubDoubtPresenter struct{}

func (s *stubDoubtPresenter) Output(_ interfaces.DoubtGame, _ error) string { return `{}` }
func (s *stubDoubtPresenter) ActionLogOutput(_ interfaces.DoubtGame) string { return `{}` }

// stubDaifugoPresenter implements presenter.DaifugoPresenter (= GamePresenter[interfaces.DaifugoGame])
type stubDaifugoPresenter struct{}

func (s *stubDaifugoPresenter) Output(_ interfaces.DaifugoGame, _ error) string { return `{}` }
func (s *stubDaifugoPresenter) ActionLogOutput(_ interfaces.DaifugoGame) string { return `{}` }

// stubSevensPresenter implements presenter.SevensPresenter (= GamePresenter[interfaces.SevensGame])
type stubSevensPresenter struct{}

func (s *stubSevensPresenter) Output(_ interfaces.SevensGame, _ error) string { return `{}` }
func (s *stubSevensPresenter) ActionLogOutput(_ interfaces.SevensGame) string { return `{}` }

// stubCrazyEightsPresenter implements presenter.CrazyEightsPresenter
type stubCrazyEightsPresenter struct{}

func (s *stubCrazyEightsPresenter) Output(_ interfaces.CrazyEightsGame, _ error) string {
	return `{}`
}
func (s *stubCrazyEightsPresenter) ActionLogOutput(_ interfaces.CrazyEightsGame) string {
	return `{}`
}
func (s *stubCrazyEightsPresenter) HintOutput(_ interfaces.CrazyEightsGame) string {
	return `{}`
}

// stubKlondikePresenter implements presenter.KlondikePresenter (GamePresenter + HintOutput)
type stubKlondikePresenter struct{}

func (s *stubKlondikePresenter) Output(_ interfaces.KlondikeGame, _ error) string { return `{}` }
func (s *stubKlondikePresenter) ActionLogOutput(_ interfaces.KlondikeGame) string { return `{}` }
func (s *stubKlondikePresenter) HintOutput(_ interfaces.KlondikeGame) string      { return `{}` }

// stubFreeCellPresenter implements presenter.FreeCellPresenter (GamePresenter + HintOutput)
type stubFreeCellPresenter struct{}

func (s *stubFreeCellPresenter) Output(_ interfaces.FreeCellGame, _ error) string { return `{}` }
func (s *stubFreeCellPresenter) ActionLogOutput(_ interfaces.FreeCellGame) string { return `{}` }
func (s *stubFreeCellPresenter) HintOutput(_ interfaces.FreeCellGame) string      { return `{}` }

// stubSpiderPresenter implements presenter.SpiderPresenter (GamePresenter + HintOutput)
type stubSpiderPresenter struct{}

func (s *stubSpiderPresenter) Output(_ interfaces.SpiderGame, _ error) string { return `{}` }
func (s *stubSpiderPresenter) ActionLogOutput(_ interfaces.SpiderGame) string { return `{}` }
func (s *stubSpiderPresenter) HintOutput(_ interfaces.SpiderGame) string      { return `{}` }

// stubSpiderettePresenter implements presenter.SpiderettePresenter.
type stubSpiderettePresenter struct{}

func (s *stubSpiderettePresenter) Output(_ interfaces.SpideretteGame, _ error) string { return `{}` }
func (s *stubSpiderettePresenter) ActionLogOutput(_ interfaces.SpideretteGame) string { return `{}` }
func (s *stubSpiderettePresenter) HintOutput(_ interfaces.SpideretteGame) string      { return `{}` }

// stubPyramidPresenter implements presenter.PyramidPresenter (GamePresenter + HintOutput)
type stubPyramidPresenter struct{}

func (s *stubPyramidPresenter) Output(_ interfaces.PyramidGame, _ error) string { return `{}` }
func (s *stubPyramidPresenter) ActionLogOutput(_ interfaces.PyramidGame) string { return `{}` }
func (s *stubPyramidPresenter) HintOutput(_ interfaces.PyramidGame) string      { return `{}` }

// stubMemoryPresenter implements presenter.MemoryPresenter (= GamePresenter[interfaces.MemoryGame])
type stubMemoryPresenter struct{}

func (s *stubMemoryPresenter) Output(_ interfaces.MemoryGame, _ error) string { return `{}` }
func (s *stubMemoryPresenter) ActionLogOutput(_ interfaces.MemoryGame) string { return `{}` }

// stubGinRummyPresenter implements presenter.GinRummyPresenter (= GamePresenter[interfaces.GinRummyGame])
type stubGinRummyPresenter struct{}

func (s *stubGinRummyPresenter) Output(_ interfaces.GinRummyGame, _ error) string { return `{}` }
func (s *stubGinRummyPresenter) ActionLogOutput(_ interfaces.GinRummyGame) string { return `{}` }

// stubCribbagePresenter implements presenter.CribbagePresenter
type stubCribbagePresenter struct{}

func (s *stubCribbagePresenter) Output(_ interfaces.CribbageGame, _ error) string { return `{}` }
func (s *stubCribbagePresenter) ActionLogOutput(_ interfaces.CribbageGame) string { return `{}` }
func (s *stubCribbagePresenter) HintOutput(_ interfaces.CribbageGame) string      { return `{}` }

// stubPaiGowPresenter implements presenter.PaiGowPresenter (= GamePresenter[interfaces.PaiGowGame])
type stubPaiGowPresenter struct{}

func (s *stubPaiGowPresenter) Output(_ interfaces.PaiGowGame, _ error) string { return `{}` }
func (s *stubPaiGowPresenter) ActionLogOutput(_ interfaces.PaiGowGame) string { return `{}` }

// --- tests ---

func TestBlackJackInteractor_SnapshotRestore(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	bi := NewBlackJackInteractor(bj, new(stubBlackJackPresenter))

	data, err := bi.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreBlackJackInteractor(data, new(stubBlackJackPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
}

func TestPokerInteractor_SnapshotRestore(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.PokerPlayer{
		domain.NewPokerPlayer(true, domain.PokerStyleBalanced),
		domain.NewPokerPlayer(false, domain.PokerStyleConservative),
	}
	p := domain.NewPoker(tc, players, domain.DefaultPokerConfig())
	pi := NewPokerInteractor(p, new(stubPokerPresenter))

	data, err := pi.Snapshot()
	require.NoError(t, err)

	restored, err := RestorePokerInteractor(data, new(stubPokerPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
}

func TestHoldemInteractor_SnapshotRestore(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.HoldemPlayer{
		domain.NewHoldemPlayer(true, domain.HoldemStyleTAG),
		domain.NewHoldemPlayer(false, domain.HoldemStyleLAP),
	}
	h := domain.NewHoldem(tc, players, domain.DefaultHoldemConfig())
	hi := NewHoldemInteractor(h, new(stubHoldemPresenter))

	data, err := hi.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreHoldemInteractor(data, new(stubHoldemPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
}

func TestOmahaInteractor_SnapshotRestore(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.OmahaPlayer{
		domain.NewOmahaPlayer(true, domain.HoldemStyleTAG),
		domain.NewOmahaPlayer(false, domain.HoldemStyleLAP),
	}
	o := domain.NewOmaha(tc, players, domain.DefaultOmahaConfig())
	oi := NewOmahaInteractor(o, new(stubOmahaPresenter))

	data, err := oi.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreOmahaInteractor(data, new(stubOmahaPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
}

func TestShortDeckInteractor_SnapshotRestore(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.ShortDeckPlayer{
		domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG),
		domain.NewShortDeckPlayer(false, domain.HoldemStyleLAP),
	}
	sd := domain.NewShortDeck(tc, players, domain.DefaultShortDeckConfig())
	si := NewShortDeckInteractor(sd, new(stubShortDeckPresenter))

	data, err := si.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreShortDeckInteractor(data, new(stubShortDeckPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
}

func TestIndianPokerInteractor_SnapshotRestore(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.IndianPokerPlayer{
		domain.NewIndianPokerPlayer(true, domain.HoldemStyleTAG),
		domain.NewIndianPokerPlayer(false, domain.HoldemStyleLAP),
	}
	ip := domain.NewIndianPoker(tc, players, domain.DefaultIndianPokerConfig())
	ii := NewIndianPokerInteractor(ip, new(stubIndianPokerPresenter))

	data, err := ii.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreIndianPokerInteractor(data, new(stubIndianPokerPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
}

func TestVideoPokerInteractor_SnapshotRestore(t *testing.T) {
	vp := domain.NewDefaultVideoPoker()
	vi := NewVideoPokerInteractor(vp, new(stubVideoPokerPresenter))

	data, err := vi.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreVideoPokerInteractor(data, new(stubVideoPokerPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
}

func TestHeartsInteractor_SnapshotRestore(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.HeartsPlayer{
		domain.NewHeartsPlayer(true),
		domain.NewHeartsPlayer(false),
		domain.NewHeartsPlayer(false),
		domain.NewHeartsPlayer(false),
	}
	h := domain.NewHearts(tc, players, domain.DefaultHeartsConfig())
	hi := NewHeartsInteractor(h, new(stubHeartsPresenter))

	data, err := hi.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreHeartsInteractor(data, new(stubHeartsPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
}

func TestSpadesInteractor_SnapshotRestore(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.SpadesPlayer{
		domain.NewSpadesPlayer(true),
		domain.NewSpadesPlayer(false),
		domain.NewSpadesPlayer(false),
		domain.NewSpadesPlayer(false),
	}
	s := domain.NewSpades(tc, players, domain.DefaultSpadesConfig())
	si := NewSpadesInteractor(s, new(stubSpadesPresenter))

	data, err := si.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreSpadesInteractor(data, new(stubSpadesPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
}

func TestEuchreInteractor_SnapshotRestore(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.EuchrePlayer{
		domain.NewEuchrePlayer(true, 0),
		domain.NewEuchrePlayer(false, 1),
		domain.NewEuchrePlayer(false, 0),
		domain.NewEuchrePlayer(false, 1),
	}
	e := domain.NewEuchre(tc, players, domain.DefaultEuchreConfig())
	ei := NewEuchreInteractor(e, new(stubEuchrePresenter))

	data, err := ei.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreEuchreInteractor(data, new(stubEuchrePresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
}

func TestNapoleonInteractor_SnapshotRestore(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.NapoleonPlayer{
		domain.NewNapoleonPlayer(true),
		domain.NewNapoleonPlayer(false),
		domain.NewNapoleonPlayer(false),
		domain.NewNapoleonPlayer(false),
		domain.NewNapoleonPlayer(false),
	}
	n := domain.NewNapoleon(tc, players, domain.DefaultNapoleonConfig())
	ni := NewNapoleonInteractor(n, new(stubNapoleonPresenter))

	data, err := ni.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreNapoleonInteractor(data, new(stubNapoleonPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
}

func TestOldMaidInteractor_SnapshotRestore(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.OldMaidPlayer{
		domain.NewOldMaidPlayer(true),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
	}
	om := domain.NewOldMaid(tc, players)
	oi := NewOldMaidInteractor(om, new(stubOldMaidPresenter))

	data, err := oi.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreOldMaidInteractor(data, new(stubOldMaidPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
}

func TestDoubtInteractor_SnapshotRestore(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.DoubtPlayer{
		domain.NewDoubtPlayer(true),
		domain.NewDoubtPlayer(false),
	}
	d := domain.NewDoubt(tc, players)
	di := NewDoubtInteractor(d, new(stubDoubtPresenter))

	data, err := di.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreDoubtInteractor(data, new(stubDoubtPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
}

func TestDaifugoInteractor_SnapshotRestore(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.DaifugoPlayer{
		domain.NewDaifugoPlayer(true),
		domain.NewDaifugoPlayer(false),
		domain.NewDaifugoPlayer(false),
	}
	dg := domain.NewDaifugo(tc, players, domain.DefaultDaifugoConfig())
	di := NewDaifugoInteractor(dg, new(stubDaifugoPresenter))

	data, err := di.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreDaifugoInteractor(data, new(stubDaifugoPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
}

func TestSevensInteractor_SnapshotRestore(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.SevensPlayer{
		domain.NewSevensPlayer(true),
		domain.NewSevensPlayer(false),
		domain.NewSevensPlayer(false),
	}
	s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
	si := NewSevensInteractor(s, new(stubSevensPresenter))

	data, err := si.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreSevensInteractor(data, new(stubSevensPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
}

func TestCrazyEightsInteractor_SnapshotRestore(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.CrazyEightsPlayer{
		domain.NewCrazyEightsPlayer(true),
		domain.NewCrazyEightsPlayer(false),
	}
	ce := domain.NewCrazyEights(tc, players, domain.DefaultCrazyEightsConfig())
	ci := NewCrazyEightsInteractor(ce, new(stubCrazyEightsPresenter))

	data, err := ci.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreCrazyEightsInteractor(data, new(stubCrazyEightsPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
}

func TestKlondikeInteractor_SnapshotRestore(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	k := domain.NewKlondike(tc)
	ki := NewKlondikeInteractor(k, new(stubKlondikePresenter))

	data, err := ki.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreKlondikeInteractor(data, new(stubKlondikePresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
}

func TestFreeCellInteractor_SnapshotRestore(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	f := domain.NewFreeCell(tc)
	fi := NewFreeCellInteractor(f, new(stubFreeCellPresenter))

	data, err := fi.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreFreeCellInteractor(data, new(stubFreeCellPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
}

func TestSpiderInteractor_SnapshotRestore(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	s := domain.NewSpider(tc)
	si := NewSpiderInteractor(s, new(stubSpiderPresenter))

	data, err := si.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreSpiderInteractor(data, new(stubSpiderPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
}

func TestSpideretteInteractor_SnapshotRestore(t *testing.T) {
	s := domain.NewDefaultSpiderette()
	si := NewSpideretteInteractor(s, new(stubSpiderettePresenter))

	data, err := si.Snapshot()
	require.NoError(t, err)
	require.NotEmpty(t, data)

	restored, err := RestoreSpideretteInteractor(data, new(stubSpiderettePresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
}

func TestSpideretteInteractor_RestoreInvalidJSON(t *testing.T) {
	_, err := RestoreSpideretteInteractor([]byte(`not json`), new(stubSpiderettePresenter))
	require.Error(t, err)
}

func TestPyramidInteractor_SnapshotRestore(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	p := domain.NewPyramid(tc)
	pi := NewPyramidInteractor(p, new(stubPyramidPresenter))

	data, err := pi.Snapshot()
	require.NoError(t, err)

	restored, err := RestorePyramidInteractor(data, new(stubPyramidPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
}

func TestMemoryInteractor_SnapshotRestore(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.MemoryPlayer{
		domain.NewMemoryPlayer(true),
		domain.NewMemoryPlayer(false),
	}
	m := domain.NewMemory(tc, players, domain.DefaultMemoryConfig())
	mi := NewMemoryInteractor(m, new(stubMemoryPresenter))

	data, err := mi.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreMemoryInteractor(data, new(stubMemoryPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
}

func TestGinRummyInteractor_SnapshotRestore(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.GinRummyPlayer{
		domain.NewGinRummyPlayer(true),
		domain.NewGinRummyPlayer(false),
	}
	gr := domain.NewGinRummy(tc, players, domain.DefaultGinRummyConfig())
	gi := NewGinRummyInteractor(gr, new(stubGinRummyPresenter))

	data, err := gi.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreGinRummyInteractor(data, new(stubGinRummyPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
}

func TestCribbageInteractor_SnapshotRestore(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.CribbagePlayer{
		domain.NewCribbagePlayer(true),
		domain.NewCribbagePlayer(false),
	}
	cr := domain.NewCribbage(tc, players, domain.DefaultCribbageConfig())
	ci := NewCribbageInteractor(cr, new(stubCribbagePresenter))

	data, err := ci.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreCribbageInteractor(data, new(stubCribbagePresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
}

func TestPaiGowInteractor_SnapshotRestore(t *testing.T) {
	pg := domain.NewDefaultPaiGow()
	pi := NewPaiGowInteractor(pg, new(stubPaiGowPresenter))

	data, err := pi.Snapshot()
	require.NoError(t, err)

	restored, err := RestorePaiGowInteractor(data, new(stubPaiGowPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
}
