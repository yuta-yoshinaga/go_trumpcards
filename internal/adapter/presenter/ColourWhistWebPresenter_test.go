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

// **本物のドメインを通します。**
func TestColourWhistWebPresenter_ArraysAreNeverNull(t *testing.T) {
	game := domain.NewDefaultColourWhist()
	game.Reset()

	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(new(ColourWhistWebPresenter).Output(game, nil)), &raw))

	for _, key := range []string{"players", "validPlays", "currentTrick", "lastTrick"} {
		require.Contains(t, raw, key, "%s が出力に無い", key)
		assert.NotEqual(t, "null", string(raw[key]), "%s が null で返っている", key)
		assert.Equal(t, byte('['), raw[key][0], "%s が配列でない", key)
	}
}

// **相手の手札は返さない。**
func TestColourWhistWebPresenter_HidesOpponentHands(t *testing.T) {
	game := domain.NewDefaultColourWhist()
	game.Reset()

	var out struct {
		Players []struct {
			ID        int              `json:"id"`
			IsHuman   bool             `json:"isHuman"`
			CardCount int              `json:"cardCount"`
			Cards     []map[string]any `json:"cards"`
		} `json:"players"`
	}
	require.NoError(t, json.Unmarshal([]byte(new(ColourWhistWebPresenter).Output(game, nil)), &out))
	require.Len(t, out.Players, domain.ColourWhistPlayerCnt)

	for _, p := range out.Players {
		if p.IsHuman {
			assert.Len(t, p.Cards, p.CardCount, "自分の手札は見える")
			continue
		}
		assert.Empty(t, p.Cards, "席 %d の手札が漏れている", p.ID)
	}
}

// **Troel が強制成立したことがワイヤに載る。** 競りで選んだ契約と区別が要ります。
func TestColourWhistWebPresenter_ReportsTroelForced(t *testing.T) {
	m := new(interfaces.MockColourWhistGame)
	m.On("IsTroelForced").Return(true)
	m.On("GetContract").Return(domain.ColourWhistContractTroel)
	m.On("GetDeclarerIdx").Return(1)
	m.On("GetPartnerIdx").Return(3)
	fillColourWhistDefaults(m)

	var out struct {
		TroelForced bool `json:"troelForced"`
		Contract    int  `json:"contract"`
		PartnerIdx  int  `json:"partnerIdx"`
	}
	require.NoError(t, json.Unmarshal([]byte(new(ColourWhistWebPresenter).Output(m, nil)), &out))
	assert.True(t, out.TroelForced)
	assert.Equal(t, domain.ColourWhistContractTroel, out.Contract)
	// **Troel の相方は最初から分かっています。** 配りで決まるためです。
	assert.Equal(t, 3, out.PartnerIdx)
}

// **競りで決まった契約では troelForced が立たない。** 負のコントロールです。
func TestColourWhistWebPresenter_BidContractIsNotTroelForced(t *testing.T) {
	m := new(interfaces.MockColourWhistGame)
	fillColourWhistDefaults(m)

	var out struct {
		TroelForced bool `json:"troelForced"`
	}
	require.NoError(t, json.Unmarshal([]byte(new(ColourWhistWebPresenter).Output(m, nil)), &out))
	assert.False(t, out.TroelForced)
}

func TestColourWhistWebPresenter_Fields(t *testing.T) {
	m := new(interfaces.MockColourWhistGame)
	m.On("GetContract").Return(domain.ColourWhistContractMiserie)
	m.On("GetTrumpSuit").Return(domain.ColourWhistNoTrump)
	m.On("GetDeclarerTricks").Return(0)
	m.On("GetRoundNumber").Return(4)
	fillColourWhistDefaults(m)

	var out struct {
		Contract    int `json:"contract"`
		TrumpSuit   int `json:"trumpSuit"`
		RoundNumber int `json:"roundNumber"`
		Config      *struct {
			Rounds int `json:"rounds"`
		} `json:"config"`
	}
	require.NoError(t, json.Unmarshal([]byte(new(ColourWhistWebPresenter).Output(m, nil)), &out))
	assert.Equal(t, domain.ColourWhistContractMiserie, out.Contract)
	assert.Equal(t, domain.ColourWhistNoTrump, out.TrumpSuit, "Miserie に切り札は無い")
	assert.Equal(t, 4, out.RoundNumber)
	require.NotNil(t, out.Config)
	assert.Equal(t, domain.ColourWhistDefaultRounds, out.Config.Rounds)
}

func TestColourWhistWebPresenter_EndMessageCode(t *testing.T) {
	win := new(interfaces.MockColourWhistGame)
	win.On("GetGameEndFlag").Return(true)
	win.On("GetWinnerIdx").Return(0)
	fillColourWhistDefaults(win)

	var out struct {
		MessageCode string `json:"messageCode"`
	}
	require.NoError(t, json.Unmarshal([]byte(new(ColourWhistWebPresenter).Output(win, nil)), &out))
	assert.Equal(t, "colourwhist.result.win", out.MessageCode)

	lose := new(interfaces.MockColourWhistGame)
	lose.On("GetGameEndFlag").Return(true)
	lose.On("GetWinnerIdx").Return(3)
	fillColourWhistDefaults(lose)
	require.NoError(t, json.Unmarshal([]byte(new(ColourWhistWebPresenter).Output(lose, nil)), &out))
	assert.Equal(t, "colourwhist.result.lose", out.MessageCode)
}

func TestColourWhistWebPresenter_ErrorAndHint(t *testing.T) {
	m := new(interfaces.MockColourWhistGame)
	fillColourWhistDefaults(m)

	var out struct {
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal(
		[]byte(new(ColourWhistWebPresenter).Output(m, errors.New("その札は出せません"))), &out))
	assert.Equal(t, "その札は出せません", out.Message)

	p := new(ColourWhistWebPresenter)
	var none struct {
		Hint *string `json:"hint"`
	}
	require.NoError(t, json.Unmarshal([]byte(p.HintOutput(m)), &none))
	assert.Nil(t, none.Hint)
	assert.NotEmpty(t, p.ActionLogOutput(m))
}

// **未公開の相方はワイヤにも出ない。** ページは isDeclarerSide をそのまま印にします。
func TestColourWhistWebPresenter_HidesTheUnrevealedPartner(t *testing.T) {
	m := new(interfaces.MockColourWhistGame)
	m.On("GetContract").Return(domain.ColourWhistContractSamen)
	m.On("GetPartnerIdx").Return(-1)
	// 公開用アクセサは相方を false にする（内部の真値とは別物）。
	m.On("IsDeclarerSideVisible", 0).Return(true)
	m.On("IsDeclarerSideVisible", 1).Return(false)
	m.On("IsDeclarerSideVisible", 2).Return(false)
	m.On("IsDeclarerSideVisible", 3).Return(false)
	fillColourWhistDefaults(m)

	var out struct {
		PartnerIdx int `json:"partnerIdx"`
		Players    []struct {
			ID             int  `json:"id"`
			IsDeclarerSide bool `json:"isDeclarerSide"`
		} `json:"players"`
	}
	require.NoError(t, json.Unmarshal([]byte(new(ColourWhistWebPresenter).Output(m, nil)), &out))
	assert.Equal(t, -1, out.PartnerIdx)
	for _, p := range out.Players {
		if p.ID == 0 {
			assert.True(t, p.IsDeclarerSide, "宣言者は常に見える")
			continue
		}
		assert.False(t, p.IsDeclarerSide, "席 %d で相方が漏れている", p.ID)
	}
}
