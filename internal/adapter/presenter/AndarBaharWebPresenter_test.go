package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// **本物のドメインを通します。** モックに配列を持たせると、サーバが実際に返す形
// (空スライスが `null` になる) を一度も踏まないまま緑になります。
func TestAndarBaharWebPresenter_ArraysAreNeverNull(t *testing.T) {
	p := new(AndarBaharWebPresenter)

	t.Run("ベット前は両列とも空", func(t *testing.T) {
		game := domain.NewDefaultAndarBahar()

		var raw map[string]json.RawMessage
		require.NoError(t, json.Unmarshal([]byte(p.Output(game, nil)), &raw))

		// **`null` は型契約違反。** TS 側は `Card[]` を非 optional で約束しています。
		for _, key := range []string{"andarCards", "baharCards", "history"} {
			require.Contains(t, raw, key, "%s が出力に無い", key)
			assert.NotEqual(t, "null", string(raw[key]), "%s が null で返っている", key)
			assert.Equal(t, byte('['), raw[key][0], "%s が配列でない", key)
		}
	})

	t.Run("決着後も配列で返る", func(t *testing.T) {
		game := domain.NewDefaultAndarBahar()
		require.NoError(t, game.Bet(100, domain.AndarBaharBetAndar, 50, domain.AndarBaharSide2To5))

		var out struct {
			AndarCards  []map[string]any `json:"andarCards"`
			BaharCards  []map[string]any `json:"baharCards"`
			History     []int            `json:"history"`
			DealtCount  int              `json:"dealtCount"`
			Winner      int              `json:"winner"`
			Payout      int              `json:"payout"`
			Chips       int              `json:"chips"`
			FirstColumn int              `json:"firstColumn"`
			Phase       int              `json:"phase"`
		}
		require.NoError(t, json.Unmarshal([]byte(p.Output(game, nil)), &out))

		assert.NotNil(t, out.AndarCards)
		assert.NotNil(t, out.BaharCards)
		assert.Len(t, out.History, 1)
		assert.Equal(t, game.DealtCount(), out.DealtCount)
		assert.Equal(t, len(out.AndarCards)+len(out.BaharCards), out.DealtCount,
			"2 列の合計が配った枚数と合わない")
		assert.Equal(t, game.GetWinner(), out.Winner)
		assert.Equal(t, game.GetPayout(), out.Payout)
		assert.Equal(t, game.GetChips(), out.Chips)
		assert.Equal(t, game.GetFirstColumn(), out.FirstColumn)
		assert.Equal(t, domain.AndarBaharPhaseEnd, out.Phase)
	})
}

func TestAndarBaharWebPresenter_Output_Fields(t *testing.T) {
	m := new(interfaces.MockAndarBaharGame)
	m.On("GetJoker").Return(domain.NewCard(domain.CardDesignSpade, 11, false))
	m.On("GetPhase").Return(domain.AndarBaharPhaseEnd)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetResult").Return(domain.GameResultWin)
	m.On("GetBetTarget").Return(domain.AndarBaharBetBahar)
	m.On("GetSideBand").Return(domain.AndarBaharSide11To15)
	m.On("GetSideAmount").Return(20)
	m.On("GetBetAmount").Return(100)
	m.On("GetPayout").Return(190)
	m.On("GetWinner").Return(domain.AndarBaharBetBahar)
	fillAndarBaharCuiDefaults(m)

	var out struct {
		Joker       *map[string]any `json:"joker"`
		BetTarget   int             `json:"betTarget"`
		SideBand    int             `json:"sideBand"`
		SideAmount  int             `json:"sideAmount"`
		BetAmount   int             `json:"betAmount"`
		Payout      int             `json:"payout"`
		Result      int             `json:"result"`
		MessageCode string          `json:"messageCode"`
	}
	require.NoError(t, json.Unmarshal([]byte(new(AndarBaharWebPresenter).Output(m, nil)), &out))

	assert.NotNil(t, out.Joker)
	assert.Equal(t, domain.AndarBaharBetBahar, out.BetTarget)
	assert.Equal(t, domain.AndarBaharSide11To15, out.SideBand)
	assert.Equal(t, 20, out.SideAmount)
	assert.Equal(t, 100, out.BetAmount)
	assert.Equal(t, 190, out.Payout)
	assert.Equal(t, int(domain.GameResultWin), out.Result)
	assert.Equal(t, "andarbahar.result.win", out.MessageCode)
}

func TestAndarBaharWebPresenter_Output_LoseMessage(t *testing.T) {
	m := new(interfaces.MockAndarBaharGame)
	m.On("GetPhase").Return(domain.AndarBaharPhaseEnd)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetResult").Return(domain.GameResultLose)
	fillAndarBaharCuiDefaults(m)

	var out struct {
		MessageCode string `json:"messageCode"`
	}
	require.NoError(t, json.Unmarshal([]byte(new(AndarBaharWebPresenter).Output(m, nil)), &out))
	assert.Equal(t, "andarbahar.result.lose", out.MessageCode)
}

func TestAndarBaharWebPresenter_Output_Error(t *testing.T) {
	m := new(interfaces.MockAndarBaharGame)
	fillAndarBaharCuiDefaults(m)

	var out struct {
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal(
		[]byte(new(AndarBaharWebPresenter).Output(m, errors.New("Insufficient chips."))), &out))
	assert.Equal(t, "Insufficient chips.", out.Message)
}

func TestAndarBaharWebPresenter_HintAndActionLog(t *testing.T) {
	m := new(interfaces.MockAndarBaharGame)
	fillAndarBaharCuiDefaults(m)

	p := new(AndarBaharWebPresenter)
	var hint struct {
		Hint string `json:"hint"`
	}
	require.NoError(t, json.Unmarshal([]byte(p.HintOutput(m)), &hint))
	assert.Equal(t, "andarBaharHintAndarFirst", hint.Hint)

	assert.NotEmpty(t, p.ActionLogOutput(m))
}
