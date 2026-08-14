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

func newAndarBaharInteractorForTest() (*interfaces.MockAndarBaharGame,
	*presenter.MockAndarBaharPresenter, *AndarBaharInteractor,
) {
	mg := new(interfaces.MockAndarBaharGame)
	mp := new(presenter.MockAndarBaharPresenter)
	return mg, mp, NewAndarBaharInteractor(mg, mp)
}

func TestNewAndarBaharInteractor(t *testing.T) {
	_, _, ai := newAndarBaharInteractorForTest()
	assert.NotNil(t, ai)
}

func TestNewAndarBaharInteractor_NilPanics(t *testing.T) {
	mp := new(presenter.MockAndarBaharPresenter)
	assert.Panics(t, func() { NewAndarBaharInteractor(nil, mp) })

	mg := new(interfaces.MockAndarBaharGame)
	assert.Panics(t, func() { NewAndarBaharInteractor(mg, nil) })
}

func TestAndarBaharInteractor_Reset(t *testing.T) {
	mg, mp, ai := newAndarBaharInteractorForTest()
	mg.On("Reset").Return()
	mp.On("Output", mg, nil).Return("reset output")

	assert.Equal(t, "reset output", ai.Reset())
	mg.AssertCalled(t, "Reset")
}

// **4 つの引数がそのままドメインへ渡る。** 並びを取り違えるとサイドベット額が
// メインベットに化けるので、呼び出しの中身まで固定します。
func TestAndarBaharInteractor_BetPassesEveryArgument(t *testing.T) {
	mg, mp, ai := newAndarBaharInteractorForTest()
	mg.On("Bet", 100, domain.AndarBaharBetBahar, 50, domain.AndarBaharSide6To10).Return(nil)
	mp.On("Output", mg, nil).Return("bet output")

	assert.Equal(t, "bet output", ai.Bet(100, domain.AndarBaharBetBahar, 50, domain.AndarBaharSide6To10))
	mg.AssertCalled(t, "Bet", 100, domain.AndarBaharBetBahar, 50, domain.AndarBaharSide6To10)
}

func TestAndarBaharInteractor_BetErrorReachesThePresenter(t *testing.T) {
	mg, mp, ai := newAndarBaharInteractorForTest()
	betErr := errors.New("Insufficient chips.")
	mg.On("Bet", 100, domain.AndarBaharBetAndar, 0, domain.AndarBaharSideNone).Return(betErr)
	mp.On("Output", mg, betErr).Return("error output")

	assert.Equal(t, "error output", ai.Bet(100, domain.AndarBaharBetAndar, 0, domain.AndarBaharSideNone))
	mp.AssertCalled(t, "Output", mg, betErr)
}

func TestAndarBaharInteractor_ClearHistory(t *testing.T) {
	mg, mp, ai := newAndarBaharInteractorForTest()
	mg.On("ClearHistory").Return()
	mp.On("Output", mg, nil).Return("cleared")

	assert.Equal(t, "cleared", ai.ClearHistory())
	mg.AssertCalled(t, "ClearHistory")
}

func TestAndarBaharInteractor_Hint(t *testing.T) {
	mg, mp, ai := newAndarBaharInteractorForTest()
	mp.On("HintOutput", mg).Return("hint output")

	assert.Equal(t, "hint output", ai.Hint())
}

func TestAndarBaharInteractor_ActionLog(t *testing.T) {
	mg, mp, ai := newAndarBaharInteractorForTest()
	mp.On("ActionLogOutput", mg).Return("log output")

	assert.Equal(t, "log output", ai.ActionLog())
}

// **保存して読み直しても同じ盤面になる。** Worker では毎リクエストこれが走ります。
func TestRestoreAndarBaharInteractor(t *testing.T) {
	mp := new(presenter.MockAndarBaharPresenter)

	game := domain.NewDefaultAndarBahar()
	require.NoError(t, game.Bet(100, domain.AndarBaharBetAndar, 50, domain.AndarBaharSide2To5))
	data, err := json.Marshal(game)
	require.NoError(t, err)

	ai, err := RestoreAndarBaharInteractor(data, mp)
	require.NoError(t, err)
	require.NotNil(t, ai)
	assert.Equal(t, game.GetWinner(), ai.Game.GetWinner())
	assert.Equal(t, game.DealtCount(), ai.Game.DealtCount())
	assert.Equal(t, game.GetChips(), ai.Game.GetChips())

	// Snapshot は復元できる形で返る。
	snap, err := ai.Snapshot()
	require.NoError(t, err)
	again, err := RestoreAndarBaharInteractor(snap, mp)
	require.NoError(t, err)
	assert.Equal(t, game.GetPayout(), again.Game.GetPayout())
}

func TestRestoreAndarBaharInteractor_RejectsBrokenData(t *testing.T) {
	mp := new(presenter.MockAndarBaharPresenter)

	_, err := RestoreAndarBaharInteractor([]byte(`{`), mp)
	assert.Error(t, err, "壊れた JSON")

	// **改竄された盤面はドメインが弾く。** インタラクタが素通ししないことを確かめます。
	_, err = RestoreAndarBaharInteractor([]byte(`{"ps":9}`), mp)
	assert.Error(t, err, "フェーズが範囲外")
}
