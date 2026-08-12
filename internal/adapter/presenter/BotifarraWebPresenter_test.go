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
// (空のトリックが `null` になる) を一度も踏まないまま緑になります。
func TestBotifarraWebPresenter_ArraysAreNeverNull(t *testing.T) {
	game := domain.NewDefaultBotifarra()
	game.Reset()

	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(new(BotifarraWebPresenter).Output(game, nil)), &raw))

	for _, key := range []string{"players", "validPlays", "currentTrick", "lastTrick", "roundPoints", "scores"} {
		require.Contains(t, raw, key, "%s が出力に無い", key)
		assert.NotEqual(t, "null", string(raw[key]), "%s が null で返っている", key)
		assert.Equal(t, byte('['), raw[key][0], "%s が配列でない", key)
	}
}

// **相手の手札は返さない。** 返すとそのまま覗けてしまいます。
func TestBotifarraWebPresenter_HidesOpponentHands(t *testing.T) {
	game := domain.NewDefaultBotifarra()
	game.Reset()

	var out struct {
		Players []struct {
			ID        int              `json:"id"`
			IsHuman   bool             `json:"isHuman"`
			Team      int              `json:"team"`
			CardCount int              `json:"cardCount"`
			Cards     []map[string]any `json:"cards"`
		} `json:"players"`
	}
	require.NoError(t, json.Unmarshal([]byte(new(BotifarraWebPresenter).Output(game, nil)), &out))
	require.Len(t, out.Players, domain.BotifarraPlayerCnt)

	for _, p := range out.Players {
		assert.Equal(t, domain.BotifarraTeamOf(p.ID), p.Team)
		assert.Equal(t, domain.BotifarraHandSize, p.CardCount, "枚数は全席ぶん見える")
		if p.IsHuman {
			assert.Len(t, p.Cards, domain.BotifarraHandSize, "自分の手札は見える")
			continue
		}
		assert.Empty(t, p.Cards, "席 %d の手札が漏れている", p.ID)
	}
}

func TestBotifarraWebPresenter_Fields(t *testing.T) {
	game := domain.NewDefaultBotifarra()
	game.Reset()
	require.NoError(t, game.Declare(domain.CardDesignHeart))

	var out struct {
		Phase       int  `json:"phase"`
		TrumpSuit   int  `json:"trumpSuit"`
		DeclarerIdx int  `json:"declarerIdx"`
		Multiplier  int  `json:"multiplier"`
		DealerIdx   int  `json:"dealerIdx"`
		WinnerTeam  int  `json:"winnerTeam"`
		GameEndFlag bool `json:"gameEndFlag"`
		Config      *struct {
			TargetScore   int  `json:"targetScore"`
			AllowDoubling bool `json:"allowDoubling"`
		} `json:"config"`
	}
	require.NoError(t, json.Unmarshal([]byte(new(BotifarraWebPresenter).Output(game, nil)), &out))

	assert.Equal(t, domain.CardDesignHeart, out.TrumpSuit)
	assert.Equal(t, 0, out.DeclarerIdx)
	assert.Equal(t, game.GetPhase(), out.Phase)
	assert.Equal(t, game.GetMultiplier(), out.Multiplier)
	assert.Equal(t, -1, out.WinnerTeam)
	assert.False(t, out.GameEndFlag)
	require.NotNil(t, out.Config)
	assert.Equal(t, domain.BotifarraDefaultTarget, out.Config.TargetScore)
}

// **切り札なしも素直に往復する。** -1 は「未設定」ではなく有効な宣言です。
func TestBotifarraWebPresenter_NoTrumpIsAValue(t *testing.T) {
	game := domain.NewDefaultBotifarra()
	game.Reset()
	require.NoError(t, game.Declare(domain.BotifarraNoTrump))

	var out struct {
		TrumpSuit   int `json:"trumpSuit"`
		DeclarerIdx int `json:"declarerIdx"`
	}
	require.NoError(t, json.Unmarshal([]byte(new(BotifarraWebPresenter).Output(game, nil)), &out))
	assert.Equal(t, domain.BotifarraNoTrump, out.TrumpSuit)
	assert.Equal(t, 0, out.DeclarerIdx)
}

func TestBotifarraWebPresenter_Error(t *testing.T) {
	game := domain.NewDefaultBotifarra()
	game.Reset()

	var out struct {
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal(
		[]byte(new(BotifarraWebPresenter).Output(game, errors.New("その札は出せません"))), &out))
	assert.Equal(t, "その札は出せません", out.Message)
}

func TestBotifarraWebPresenter_EndMessageCode(t *testing.T) {
	m := new(interfaces.MockBotifarraGame)
	fillBotifarraDefaults(m)
	m.ExpectedCalls = nil
	m.On("GetGameEndFlag").Return(true)
	m.On("GetWinnerTeam").Return(domain.BotifarraTeamOf(0))
	fillBotifarraDefaults(m)

	var out struct {
		MessageCode string `json:"messageCode"`
	}
	require.NoError(t, json.Unmarshal([]byte(new(BotifarraWebPresenter).Output(m, nil)), &out))
	assert.Equal(t, "botifarra.result.win", out.MessageCode)

	lose := new(interfaces.MockBotifarraGame)
	lose.On("GetGameEndFlag").Return(true)
	lose.On("GetWinnerTeam").Return(1)
	fillBotifarraDefaults(lose)

	require.NoError(t, json.Unmarshal([]byte(new(BotifarraWebPresenter).Output(lose, nil)), &out))
	assert.Equal(t, "botifarra.result.lose", out.MessageCode)
}

func TestBotifarraWebPresenter_HintAndActionLog(t *testing.T) {
	game := domain.NewDefaultBotifarra()
	game.Reset()

	p := new(BotifarraWebPresenter)
	var hint struct {
		Suit   *int   `json:"suit"`
		Reason string `json:"reason"`
	}
	require.NoError(t, json.Unmarshal([]byte(p.HintOutput(game)), &hint))
	require.NotNil(t, hint.Suit)
	assert.Equal(t, "botifarraDeclareLongest", hint.Reason)

	game.GiveUp()
	var none struct {
		Hint *string `json:"hint"`
	}
	require.NoError(t, json.Unmarshal([]byte(p.HintOutput(game)), &none))
	assert.Nil(t, none.Hint)

	assert.NotEmpty(t, p.ActionLogOutput(game))
}
