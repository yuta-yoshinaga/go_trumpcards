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
func TestRikkenWebPresenter_ArraysAreNeverNull(t *testing.T) {
	game := domain.NewDefaultRikken()
	game.Reset()

	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(new(RikkenWebPresenter).Output(game, nil)), &raw))

	for _, key := range []string{"players", "validPlays", "currentTrick", "lastTrick"} {
		require.Contains(t, raw, key, "%s が出力に無い", key)
		assert.NotEqual(t, "null", string(raw[key]), "%s が null で返っている", key)
		assert.Equal(t, byte('['), raw[key][0], "%s が配列でない", key)
	}
}

// **相手の手札は返さない。**
func TestRikkenWebPresenter_HidesOpponentHands(t *testing.T) {
	game := domain.NewDefaultRikken()
	game.Reset()

	var out struct {
		Players []struct {
			ID        int              `json:"id"`
			IsHuman   bool             `json:"isHuman"`
			CardCount int              `json:"cardCount"`
			Cards     []map[string]any `json:"cards"`
			Score     int              `json:"score"`
		} `json:"players"`
	}
	require.NoError(t, json.Unmarshal([]byte(new(RikkenWebPresenter).Output(game, nil)), &out))
	require.Len(t, out.Players, domain.RikkenPlayerCnt)

	for _, p := range out.Players {
		assert.Equal(t, domain.RikkenHandSize, p.CardCount, "枚数は全席ぶん見える")
		if p.IsHuman {
			assert.Len(t, p.Cards, domain.RikkenHandSize, "自分の手札は見える")
			continue
		}
		assert.Empty(t, p.Cards, "席 %d の手札が漏れている", p.ID)
	}
}

// **相方は公開されるまで -1。** 席が漏れると誰と組んでいるか分かってしまいます。
func TestRikkenWebPresenter_HidesThePartnerUntilRevealed(t *testing.T) {
	m := new(interfaces.MockRikkenGame)
	m.On("GetPartnerIdx").Return(-1)
	fillRikkenDefaults(m)

	var out struct {
		PartnerIdx int `json:"partnerIdx"`
	}
	require.NoError(t, json.Unmarshal([]byte(new(RikkenWebPresenter).Output(m, nil)), &out))
	assert.Equal(t, -1, out.PartnerIdx)
}

func TestRikkenWebPresenter_Fields(t *testing.T) {
	m := new(interfaces.MockRikkenGame)
	m.On("GetContract").Return(domain.RikkenContractSolo)
	m.On("GetTrumpSuit").Return(domain.CardDesignDiamond)
	m.On("GetDeclarerTricks").Return(5)
	m.On("GetRoundNumber").Return(3)
	fillRikkenDefaults(m)

	var out struct {
		Contract       int `json:"contract"`
		TrumpSuit      int `json:"trumpSuit"`
		DeclarerTricks int `json:"declarerTricks"`
		RoundNumber    int `json:"roundNumber"`
		Config         *struct {
			Rounds int `json:"rounds"`
		} `json:"config"`
	}
	require.NoError(t, json.Unmarshal([]byte(new(RikkenWebPresenter).Output(m, nil)), &out))
	assert.Equal(t, domain.RikkenContractSolo, out.Contract)
	assert.Equal(t, domain.CardDesignDiamond, out.TrumpSuit)
	assert.Equal(t, 5, out.DeclarerTricks)
	assert.Equal(t, 3, out.RoundNumber)
	require.NotNil(t, out.Config)
	assert.Equal(t, domain.RikkenDefaultRounds, out.Config.Rounds)
}

func TestRikkenWebPresenter_EndMessageCode(t *testing.T) {
	win := new(interfaces.MockRikkenGame)
	win.On("GetGameEndFlag").Return(true)
	win.On("GetWinnerIdx").Return(0)
	fillRikkenDefaults(win)

	var out struct {
		MessageCode string `json:"messageCode"`
	}
	require.NoError(t, json.Unmarshal([]byte(new(RikkenWebPresenter).Output(win, nil)), &out))
	assert.Equal(t, "rikken.result.win", out.MessageCode)

	lose := new(interfaces.MockRikkenGame)
	lose.On("GetGameEndFlag").Return(true)
	lose.On("GetWinnerIdx").Return(2)
	fillRikkenDefaults(lose)
	require.NoError(t, json.Unmarshal([]byte(new(RikkenWebPresenter).Output(lose, nil)), &out))
	assert.Equal(t, "rikken.result.lose", out.MessageCode)
}

func TestRikkenWebPresenter_Error(t *testing.T) {
	m := new(interfaces.MockRikkenGame)
	fillRikkenDefaults(m)

	var out struct {
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal(
		[]byte(new(RikkenWebPresenter).Output(m, errors.New("その札は出せません"))), &out))
	assert.Equal(t, "その札は出せません", out.Message)
}

func TestRikkenWebPresenter_HintAndActionLog(t *testing.T) {
	game := domain.NewDefaultRikken()
	game.Reset()

	p := new(RikkenWebPresenter)
	var hint struct {
		Contract *int   `json:"contract"`
		Reason   string `json:"reason"`
	}
	require.NoError(t, json.Unmarshal([]byte(p.HintOutput(game)), &hint))
	if game.IsHumanTurn() {
		require.NotNil(t, hint.Contract)
		assert.Equal(t, "rikkenBidStrength", hint.Reason)
	}

	game.GiveUp()
	var none struct {
		Hint *string `json:"hint"`
	}
	require.NoError(t, json.Unmarshal([]byte(p.HintOutput(game)), &none))
	assert.Nil(t, none.Hint)

	assert.NotEmpty(t, p.ActionLogOutput(game))
}

// **オープンミゼールは宣言者の手札を公開する。** 名前どおりの仕掛けです。
//
// これが無いと、オープンミゼールは「ミゼールだが点が高いだけ」になります。
func TestRikkenWebPresenter_OpenMisereRevealsTheDeclarerHand(t *testing.T) {
	hand := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignHeart, 3, false),
	}
	cpu := domain.NewRikkenPlayer(false)
	for _, c := range hand {
		cpu.AddCard(c)
	}

	m := new(interfaces.MockRikkenGame)
	m.On("GetContract").Return(domain.RikkenContractOpenMisere)
	m.On("GetDeclarerIdx").Return(1)
	m.On("GetTrumpSuit").Return(domain.RikkenNoTrump)
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(domain.NewRikkenPlayer(true))
	m.On("GetPlayer", 1).Return(cpu)
	fillRikkenDefaults(m)

	var out struct {
		Players []struct {
			ID    int              `json:"id"`
			Cards []map[string]any `json:"cards"`
		} `json:"players"`
	}
	require.NoError(t, json.Unmarshal([]byte(new(RikkenWebPresenter).Output(m, nil)), &out))
	require.Len(t, out.Players, 2)
	assert.Len(t, out.Players[1].Cards, len(hand), "宣言者の手札が公開されていない")
}

// **ミゼール（非公開）では宣言者の手札は伏せたまま。** 負のコントロールです。
func TestRikkenWebPresenter_PlainMisereKeepsTheDeclarerHandHidden(t *testing.T) {
	cpu := domain.NewRikkenPlayer(false)
	cpu.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))

	m := new(interfaces.MockRikkenGame)
	m.On("GetContract").Return(domain.RikkenContractMisere)
	m.On("GetDeclarerIdx").Return(1)
	m.On("GetTrumpSuit").Return(domain.RikkenNoTrump)
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(domain.NewRikkenPlayer(true))
	m.On("GetPlayer", 1).Return(cpu)
	fillRikkenDefaults(m)

	var out struct {
		Players []struct {
			Cards []map[string]any `json:"cards"`
		} `json:"players"`
	}
	require.NoError(t, json.Unmarshal([]byte(new(RikkenWebPresenter).Output(m, nil)), &out))
	require.Len(t, out.Players, 2)
	assert.Empty(t, out.Players[1].Cards, "ミゼールで手札が漏れている")
}
