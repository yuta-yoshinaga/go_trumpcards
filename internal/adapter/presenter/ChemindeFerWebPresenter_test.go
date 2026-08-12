//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// chemindeFerWebDecode は Output の JSON を素の map に戻す。
func chemindeFerWebDecode(t *testing.T, raw string) map[string]json.RawMessage {
	t.Helper()
	var out map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(raw), &out))
	return out
}

// **配っていないときも配列は配列。**
//
// null を返すと TS 側の非 optional な配列が破れてページが落ちる。
func TestChemindeFerWebPresenter_ArraysAreNeverNull(t *testing.T) {
	cp := new(ChemindeFerWebPresenter)
	g := newChemindeFerForPresenter(t)

	out := chemindeFerWebDecode(t, cp.Output(g, nil))
	// 配る前の手札は空。**それでも配列で返る。**
	for _, key := range []string{"bankerHand", "punterHand"} {
		assert.Equal(t, "[]", string(out[key]), "%s が配列で返っていない", key)
	}
	// 席はいつでも 6 つ揃っている (空配列で返るのは既定出力のときだけ)。
	var seats []json.RawMessage
	require.NoError(t, json.Unmarshal(out["players"], &seats))
	assert.Len(t, seats, domain.ChemindeFerSeatCnt)
}

func TestChemindeFerWebPresenter_CarriesTheTable(t *testing.T) {
	cp := new(ChemindeFerWebPresenter)
	g := newChemindeFerForPresenter(t)
	require.NoError(t, g.SetStake(300))

	var got struct {
		Players []struct {
			ID               int  `json:"id"`
			IsHuman          bool `json:"isHuman"`
			Chips            int  `json:"chips"`
			IsBanker         bool `json:"isBanker"`
			IsRepresentative bool `json:"isRepresentative"`
		} `json:"players"`
		Phase          int `json:"phase"`
		BankerIdx      int `json:"bankerIdx"`
		BetTurn        int `json:"betTurn"`
		Stake          int `json:"stake"`
		RemainingStake int `json:"remainingStake"`
		BetMax         int `json:"betMax"`
		StakeMax       int `json:"stakeMax"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))

	require.Len(t, got.Players, domain.ChemindeFerSeatCnt)
	assert.True(t, got.Players[0].IsHuman, "席 0 が人間")
	assert.Equal(t, int(domain.ChemindeFerPhaseBet), got.Phase)
	assert.Equal(t, 300, got.Stake)
	assert.Equal(t, 300, got.RemainingStake)
	assert.Equal(t, g.GetBetTurn(), got.BetTurn)

	// **親の印は席ではなくラウンドで決まる。**
	assert.True(t, got.Players[g.GetBankerIdx()].IsBanker)
	for i, p := range got.Players {
		if i != g.GetBankerIdx() {
			assert.False(t, p.IsBanker, "席 %d が親になっている", i)
		}
	}

	// 手番の子が賭けられる上限がそのまま出ている。
	_, wantBetMax := g.BetRangeFor(g.GetBetTurn())
	assert.Equal(t, wantBetMax, got.BetMax)
}

// **賭けが締まったら賭け可能額は 0。**
//
// 手番の居ない範囲を返すと、締め切った後の卓に入力欄が出る。
func TestChemindeFerWebPresenter_BetRangeIsZeroOnceBettingCloses(t *testing.T) {
	cp := new(ChemindeFerWebPresenter)
	g := chemindeFerPresenterPosition(t, 4, 5, domain.ChemindeFerPhaseBankerDraw)

	var got struct {
		BetTurn int `json:"betTurn"`
		BetMin  int `json:"betMin"`
		BetMax  int `json:"betMax"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))
	assert.Equal(t, -1, got.BetTurn)
	assert.Zero(t, got.BetMin)
	assert.Zero(t, got.BetMax)
}

// **選べるかどうかがそのままワイヤに乗る。** ページ側で計算し直させない。
func TestChemindeFerWebPresenter_PunterMayChooseIsOnTheWire(t *testing.T) {
	cp := new(ChemindeFerWebPresenter)

	for _, tt := range []struct {
		name  string
		total int
		want  bool
	}{
		{"0 は引かされる", 0, false},
		{"4 は引かされる", 4, false},
		{"5 だけが選べる", domain.ChemindeFerPunterFreeTotal, true},
		{"6 は立たされる", 6, false},
		{"7 は立たされる", 7, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := chemindeFerPresenterPosition(t, tt.total, 2, domain.ChemindeFerPhasePunterDraw)
			var got struct {
				PunterMayChoose bool `json:"punterMayChoose"`
				PunterTotal     int  `json:"punterTotal"`
			}
			require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))
			assert.Equal(t, tt.total, got.PunterTotal)
			assert.Equal(t, tt.want, got.PunterMayChoose)
		})
	}
}

func TestChemindeFerWebPresenter_ErrorGoesToMessage(t *testing.T) {
	cp := new(ChemindeFerWebPresenter)
	g := newChemindeFerForPresenter(t)

	var got struct {
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, errors.New("boom"))), &got))
	assert.Equal(t, "boom", got.Message)
}

// **勝敗はチップで決まる。** 親を何度取ったかではない。
func TestChemindeFerWebPresenter_EndMessageCode(t *testing.T) {
	cp := new(ChemindeFerWebPresenter)

	for _, tt := range []struct {
		name  string
		mine  int
		other int
		want  string
	}{
		{"いちばん多ければ勝ち", 2000, 500, "chemindefer.result.win"},
		{"並んでいれば引き分け", 1000, 1000, "chemindefer.result.draw"},
		{"上が居れば負け", 500, 2000, "chemindefer.result.lose"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := newChemindeFerForPresenter(t)
			for i := range domain.ChemindeFerSeatCnt {
				g.GetPlayer(i).SetChips(tt.other)
			}
			g.GetPlayer(0).SetChips(tt.mine)
			g.GiveUp()

			var got struct {
				MessageCode string `json:"messageCode"`
			}
			require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))
			assert.Equal(t, tt.want, got.MessageCode)
		})
	}
}

func TestChemindeFerWebPresenter_HintOutput(t *testing.T) {
	cp := new(ChemindeFerWebPresenter)

	t.Run("判断どころでなければ null", func(t *testing.T) {
		g := chemindeFerPresenterPosition(t, 3, 2, domain.ChemindeFerPhaseRoundEnd)
		assert.JSONEq(t, `{"hint":null}`, cp.HintOutput(g))
	})

	t.Run("親の判断には draw と reason", func(t *testing.T) {
		g := chemindeFerPresenterPosition(t, 4, 2, domain.ChemindeFerPhaseBankerDraw)
		var got struct {
			Draw   bool   `json:"draw"`
			Reason string `json:"reason"`
		}
		require.NoError(t, json.Unmarshal([]byte(cp.HintOutput(g)), &got))
		assert.True(t, got.Draw)
		assert.Equal(t, "bankerDraw", got.Reason)
	})
}

func TestChemindeFerWebPresenter_ActionLogOutput(t *testing.T) {
	cp := new(ChemindeFerWebPresenter)
	g := newChemindeFerForPresenter(t)
	assert.NotEmpty(t, cp.ActionLogOutput(g))
}
