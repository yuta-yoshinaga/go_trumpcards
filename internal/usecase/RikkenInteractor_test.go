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

func newRikkenInteractorForTest() (*interfaces.MockRikkenGame,
	*presenter.MockRikkenPresenter, *RikkenInteractor,
) {
	mg := new(interfaces.MockRikkenGame)
	mp := new(presenter.MockRikkenPresenter)
	return mg, mp, NewRikkenInteractor(mg, mp)
}

func TestNewRikkenInteractor(t *testing.T) {
	_, _, ri := newRikkenInteractorForTest()
	assert.NotNil(t, ri)
}

func TestNewRikkenInteractor_NilPanics(t *testing.T) {
	mp := new(presenter.MockRikkenPresenter)
	assert.Panics(t, func() { NewRikkenInteractor(nil, mp) })

	mg := new(interfaces.MockRikkenGame)
	assert.Panics(t, func() { NewRikkenInteractor(mg, nil) })
}

func TestRikkenInteractor_Reset(t *testing.T) {
	mg, mp, ri := newRikkenInteractorForTest()
	mg.On("Reset").Return()
	mp.On("Output", mg, nil).Return("reset output")

	assert.Equal(t, "reset output", ri.Reset())
	mg.AssertCalled(t, "Reset")
}

// **どの操作もドメインへそのまま渡る。** 引数はどれも int なので、取り違えても
// コンパイルは通ります。呼び出しの中身まで固定します。
func TestRikkenInteractor_ActionsReachTheDomain(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*interfaces.MockRikkenGame)
		invoke func(*RikkenInteractor) string
		assert func(*testing.T, *interfaces.MockRikkenGame)
	}{
		{
			name:   "Bid",
			setup:  func(m *interfaces.MockRikkenGame) { m.On("Bid", domain.RikkenContractSolo).Return(nil) },
			invoke: func(ri *RikkenInteractor) string { return ri.Bid(domain.RikkenContractSolo) },
			assert: func(t *testing.T, m *interfaces.MockRikkenGame) {
				m.AssertCalled(t, "Bid", domain.RikkenContractSolo)
			},
		},
		{
			// **パスも契約値のひとつ (0)。** 「宣言しない」を別経路にしていません。
			name:   "Bid pass",
			setup:  func(m *interfaces.MockRikkenGame) { m.On("Bid", domain.RikkenContractNone).Return(nil) },
			invoke: func(ri *RikkenInteractor) string { return ri.Bid(domain.RikkenContractNone) },
			assert: func(t *testing.T, m *interfaces.MockRikkenGame) {
				m.AssertCalled(t, "Bid", domain.RikkenContractNone)
			},
		},
		{
			name:   "Call",
			setup:  func(m *interfaces.MockRikkenGame) { m.On("Call", domain.CardDesignHeart).Return(nil) },
			invoke: func(ri *RikkenInteractor) string { return ri.Call(domain.CardDesignHeart) },
			assert: func(t *testing.T, m *interfaces.MockRikkenGame) {
				m.AssertCalled(t, "Call", domain.CardDesignHeart)
			},
		},
		{
			name:   "PlayCard",
			setup:  func(m *interfaces.MockRikkenGame) { m.On("PlayCard", 7).Return(nil) },
			invoke: func(ri *RikkenInteractor) string { return ri.PlayCard(7) },
			assert: func(t *testing.T, m *interfaces.MockRikkenGame) { m.AssertCalled(t, "PlayCard", 7) },
		},
		{
			name:   "NextRound",
			setup:  func(m *interfaces.MockRikkenGame) { m.On("NextRound").Return(nil) },
			invoke: func(ri *RikkenInteractor) string { return ri.NextRound() },
			assert: func(t *testing.T, m *interfaces.MockRikkenGame) { m.AssertCalled(t, "NextRound") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mg, mp, ri := newRikkenInteractorForTest()
			mg.On("GetGameEndFlag").Return(false)
			tt.setup(mg)
			mp.On("Output", mg, nil).Return("ok output")

			assert.Equal(t, "ok output", tt.invoke(ri))
			tt.assert(t, mg)
		})
	}
}

func TestRikkenInteractor_ErrorsReachThePresenter(t *testing.T) {
	mg, mp, ri := newRikkenInteractorForTest()
	bidErr := errors.New("いまの契約より強い宣言が要ります")
	mg.On("GetGameEndFlag").Return(false)
	mg.On("Bid", domain.RikkenContractRik).Return(bidErr)
	mp.On("Output", mg, bidErr).Return("error output")

	assert.Equal(t, "error output", ri.Bid(domain.RikkenContractRik))
	mp.AssertCalled(t, "Output", mg, bidErr)
}

// **終局後の操作はドメインまで届かない。**
func TestRikkenInteractor_BlocksAfterGameEnd(t *testing.T) {
	for _, tt := range []struct {
		name   string
		invoke func(*RikkenInteractor) string
		method string
	}{
		{"Bid", func(ri *RikkenInteractor) string { return ri.Bid(1) }, "Bid"},
		{"Call", func(ri *RikkenInteractor) string { return ri.Call(1) }, "Call"},
		{"PlayCard", func(ri *RikkenInteractor) string { return ri.PlayCard(0) }, "PlayCard"},
		{"NextRound", func(ri *RikkenInteractor) string { return ri.NextRound() }, "NextRound"},
		{"GiveUp", func(ri *RikkenInteractor) string { return ri.GiveUp() }, "GiveUp"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mg, mp, ri := newRikkenInteractorForTest()
			mg.On("GetGameEndFlag").Return(true)
			mp.On("Output", mg, nil).Return("game over")

			assert.NotEmpty(t, tt.invoke(ri))
			mg.AssertNotCalled(t, tt.method)
		})
	}
}

func TestRikkenInteractor_GiveUp(t *testing.T) {
	mg, mp, ri := newRikkenInteractorForTest()
	mg.On("GetGameEndFlag").Return(false)
	mg.On("GiveUp").Return()
	mp.On("Output", mg, nil).Return("gave up")

	assert.Equal(t, "gave up", ri.GiveUp())
	mg.AssertCalled(t, "GiveUp")
}

func TestRikkenInteractor_ConfigHintAndLog(t *testing.T) {
	mg, mp, ri := newRikkenInteractorForTest()
	cfg := domain.DefaultRikkenConfig()
	mg.On("GetConfig").Return(cfg)
	mp.On("HintOutput", mg).Return("hint output")
	mp.On("ActionLogOutput", mg).Return("log output")

	assert.Equal(t, cfg, ri.GetConfig())
	assert.Equal(t, "hint output", ri.Hint())
	assert.Equal(t, "log output", ri.ActionLog())
}

// **保存して読み直しても同じ盤面になる。** Worker では毎リクエストこれが走ります。
func TestRestoreRikkenInteractor(t *testing.T) {
	mp := new(presenter.MockRikkenPresenter)

	game := domain.NewDefaultRikken()
	game.Reset()
	data, err := json.Marshal(game)
	require.NoError(t, err)

	ri, err := RestoreRikkenInteractor(data, mp)
	require.NoError(t, err)
	require.NotNil(t, ri)
	assert.Equal(t, game.GetPhase(), ri.Game.GetPhase())
	assert.Equal(t, game.GetContract(), ri.Game.GetContract())
	assert.Equal(t, game.GetRoundNumber(), ri.Game.GetRoundNumber())

	snap, err := ri.Snapshot()
	require.NoError(t, err)
	again, err := RestoreRikkenInteractor(snap, mp)
	require.NoError(t, err)
	assert.Equal(t, game.GetTrumpSuit(), again.Game.GetTrumpSuit())
}

func TestRestoreRikkenInteractor_RejectsBrokenData(t *testing.T) {
	mp := new(presenter.MockRikkenPresenter)

	_, err := RestoreRikkenInteractor([]byte(`{`), mp)
	assert.Error(t, err, "壊れた JSON")

	// **改竄された盤面はドメインが弾く。**
	_, err = RestoreRikkenInteractor([]byte(`{"ph":99}`), mp)
	assert.Error(t, err, "フェーズが範囲外")
}
