package usecase

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newBotifarraInteractorForTest() (*interfaces.MockBotifarraGame,
	*presenter.MockBotifarraPresenter, *BotifarraInteractor,
) {
	mg := new(interfaces.MockBotifarraGame)
	mp := new(presenter.MockBotifarraPresenter)
	return mg, mp, NewBotifarraInteractor(mg, mp)
}

func TestNewBotifarraInteractor(t *testing.T) {
	_, _, bi := newBotifarraInteractorForTest()
	assert.NotNil(t, bi)
}

func TestNewBotifarraInteractor_NilPanics(t *testing.T) {
	mp := new(presenter.MockBotifarraPresenter)
	assert.Panics(t, func() { NewBotifarraInteractor(nil, mp) })

	mg := new(interfaces.MockBotifarraGame)
	assert.Panics(t, func() { NewBotifarraInteractor(mg, nil) })
}

func TestBotifarraInteractor_Reset(t *testing.T) {
	mg, mp, bi := newBotifarraInteractorForTest()
	mg.On("Reset").Return()
	mp.On("Output", mg, nil).Return("reset output")

	assert.Equal(t, "reset output", bi.Reset())
	mg.AssertCalled(t, "Reset")
}

// **どの操作もドメインへそのまま渡る。** 引数を取り違えても型が同じで通るので、
// 呼び出しの中身まで固定します。
func TestBotifarraInteractor_ActionsReachTheDomain(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*interfaces.MockBotifarraGame)
		invoke func(*BotifarraInteractor) string
		assert func(*testing.T, *interfaces.MockBotifarraGame)
	}{
		{
			name:   "Declare",
			setup:  func(m *interfaces.MockBotifarraGame) { m.On("Declare", domain.CardDesignHeart).Return(nil) },
			invoke: func(bi *BotifarraInteractor) string { return bi.Declare(domain.CardDesignHeart) },
			assert: func(t *testing.T, m *interfaces.MockBotifarraGame) {
				m.AssertCalled(t, "Declare", domain.CardDesignHeart)
			},
		},
		{
			name:   "Declare no trump",
			setup:  func(m *interfaces.MockBotifarraGame) { m.On("Declare", domain.BotifarraNoTrump).Return(nil) },
			invoke: func(bi *BotifarraInteractor) string { return bi.Declare(domain.BotifarraNoTrump) },
			assert: func(t *testing.T, m *interfaces.MockBotifarraGame) {
				m.AssertCalled(t, "Declare", domain.BotifarraNoTrump)
			},
		},
		{
			name:   "Delegate",
			setup:  func(m *interfaces.MockBotifarraGame) { m.On("Delegate").Return(nil) },
			invoke: func(bi *BotifarraInteractor) string { return bi.Delegate() },
			assert: func(t *testing.T, m *interfaces.MockBotifarraGame) { m.AssertCalled(t, "Delegate") },
		},
		{
			name:   "Double",
			setup:  func(m *interfaces.MockBotifarraGame) { m.On("Double").Return(nil) },
			invoke: func(bi *BotifarraInteractor) string { return bi.Double() },
			assert: func(t *testing.T, m *interfaces.MockBotifarraGame) { m.AssertCalled(t, "Double") },
		},
		{
			name:   "PassDouble",
			setup:  func(m *interfaces.MockBotifarraGame) { m.On("PassDouble").Return(nil) },
			invoke: func(bi *BotifarraInteractor) string { return bi.PassDouble() },
			assert: func(t *testing.T, m *interfaces.MockBotifarraGame) { m.AssertCalled(t, "PassDouble") },
		},
		{
			name:   "PlayCard",
			setup:  func(m *interfaces.MockBotifarraGame) { m.On("PlayCard", 5).Return(nil) },
			invoke: func(bi *BotifarraInteractor) string { return bi.PlayCard(5) },
			assert: func(t *testing.T, m *interfaces.MockBotifarraGame) { m.AssertCalled(t, "PlayCard", 5) },
		},
		{
			name:   "NextRound",
			setup:  func(m *interfaces.MockBotifarraGame) { m.On("NextRound").Return(nil) },
			invoke: func(bi *BotifarraInteractor) string { return bi.NextRound() },
			assert: func(t *testing.T, m *interfaces.MockBotifarraGame) { m.AssertCalled(t, "NextRound") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mg, mp, bi := newBotifarraInteractorForTest()
			mg.On("GetGameEndFlag").Return(false)
			tt.setup(mg)
			mp.On("Output", mg, nil).Return("ok output")

			assert.Equal(t, "ok output", tt.invoke(bi))
			tt.assert(t, mg)
		})
	}
}

// **ドメインのエラーはそのままプレゼンタへ渡す。**
func TestBotifarraInteractor_ErrorsReachThePresenter(t *testing.T) {
	mg, mp, bi := newBotifarraInteractorForTest()
	playErr := errors.New("その札は出せません")
	mg.On("GetGameEndFlag").Return(false)
	mg.On("PlayCard", 2).Return(playErr)
	mp.On("Output", mg, playErr).Return("error output")

	assert.Equal(t, "error output", bi.PlayCard(2))
	mp.AssertCalled(t, "Output", mg, playErr)
}

// **終局後の操作はドメインまで届かない。**
func TestBotifarraInteractor_BlocksAfterGameEnd(t *testing.T) {
	for _, tt := range []struct {
		name   string
		invoke func(*BotifarraInteractor) string
		method string
	}{
		{"PlayCard", func(bi *BotifarraInteractor) string { return bi.PlayCard(0) }, "PlayCard"},
		{"Declare", func(bi *BotifarraInteractor) string { return bi.Declare(0) }, "Declare"},
		{"Delegate", func(bi *BotifarraInteractor) string { return bi.Delegate() }, "Delegate"},
		{"Double", func(bi *BotifarraInteractor) string { return bi.Double() }, "Double"},
		{"PassDouble", func(bi *BotifarraInteractor) string { return bi.PassDouble() }, "PassDouble"},
		{"NextRound", func(bi *BotifarraInteractor) string { return bi.NextRound() }, "NextRound"},
		{"GiveUp", func(bi *BotifarraInteractor) string { return bi.GiveUp() }, "GiveUp"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mg, mp, bi := newBotifarraInteractorForTest()
			mg.On("GetGameEndFlag").Return(true)
			mp.On("Output", mg, nil).Return("game over")

			assert.NotEmpty(t, tt.invoke(bi))
			mg.AssertNotCalled(t, tt.method)
		})
	}
}

func TestBotifarraInteractor_GiveUp(t *testing.T) {
	mg, mp, bi := newBotifarraInteractorForTest()
	mg.On("GetGameEndFlag").Return(false)
	mg.On("GiveUp").Return()
	mp.On("Output", mg, nil).Return("gave up")

	assert.Equal(t, "gave up", bi.GiveUp())
	mg.AssertCalled(t, "GiveUp")
}

func TestBotifarraInteractor_ConfigHintAndLog(t *testing.T) {
	mg, mp, bi := newBotifarraInteractorForTest()
	cfg := domain.DefaultBotifarraConfig()
	mg.On("GetConfig").Return(cfg)
	mp.On("HintOutput", mg).Return("hint output")
	mp.On("ActionLogOutput", mg).Return("log output")

	assert.Equal(t, cfg, bi.GetConfig())
	assert.Equal(t, "hint output", bi.Hint())
	assert.Equal(t, "log output", bi.ActionLog())
}

// **保存して読み直しても同じ盤面になる。** Worker では毎リクエストこれが走ります。
func TestRestoreBotifarraInteractor(t *testing.T) {
	mp := new(presenter.MockBotifarraPresenter)

	game := domain.NewDefaultBotifarra()
	game.Reset()
	require.NoError(t, game.Declare(domain.CardDesignSpade))
	data, err := json.Marshal(game)
	require.NoError(t, err)

	bi, err := RestoreBotifarraInteractor(data, mp)
	require.NoError(t, err)
	require.NotNil(t, bi)
	assert.Equal(t, game.GetTrumpSuit(), bi.Game.GetTrumpSuit())
	assert.Equal(t, game.GetPhase(), bi.Game.GetPhase())
	assert.Equal(t, game.GetDeclarerIdx(), bi.Game.GetDeclarerIdx())

	snap, err := bi.Snapshot()
	require.NoError(t, err)
	again, err := RestoreBotifarraInteractor(snap, mp)
	require.NoError(t, err)
	assert.Equal(t, game.GetTrickCount(), again.Game.GetTrickCount())
}

func TestRestoreBotifarraInteractor_RejectsBrokenData(t *testing.T) {
	mp := new(presenter.MockBotifarraPresenter)

	_, err := RestoreBotifarraInteractor([]byte(`{`), mp)
	assert.Error(t, err, "壊れた JSON")

	// **改竄された盤面はドメインが弾く。** インタラクタが素通ししないことを確かめます。
	_, err = RestoreBotifarraInteractor([]byte(`{"ph":99}`), mp)
	assert.Error(t, err, "フェーズが範囲外")
}
