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

// newKingoForPresenter は本物のドメインを返す。
func newKingoForPresenter(t *testing.T) *domain.Kingo {
	t.Helper()
	g := domain.NewDefaultKingo()
	g.Reset()
	return g
}

// kingoAsChild は人間が子の張り待ちになるまで進める。
func kingoAsChild(t *testing.T) *domain.Kingo {
	t.Helper()
	g := newKingoForPresenter(t)
	for range 10 {
		if !g.IsHumanBanker() {
			return g
		}
		require.NoError(t, g.Deal())
		require.NoError(t, g.NextRound())
	}
	t.Fatalf("10 ラウンド回しても人間が子にならなかった")
	return nil
}

// kingoSettled は決着まで進めた卓を返す。
func kingoSettled(t *testing.T) *domain.Kingo {
	t.Helper()
	g := newKingoForPresenter(t)
	if g.IsHumanBanker() {
		require.NoError(t, g.Deal())
	} else {
		require.NoError(t, g.PlaceBet(g.GetConfig().MinBet))
	}
	return g
}

// --- CUI ---

func TestKingoCuiPresenter_RendersJapaneseNotRawKeys(t *testing.T) {
	cp := new(KingoCuiPresenter)
	out := cp.Output(newKingoForPresenter(t), nil)

	assert.Contains(t, out, "ラウンド:")
	assert.Contains(t, out, "親:")
	assert.NotContains(t, out, "kingo.", "生の i18n キーが出力に混ざっている")
}

// **配る前は誰の手札も見えない。** キンゴに「自分だけ見える手札」は無い。
func TestKingoCuiPresenter_ShowsNoHandsBeforeTheDeal(t *testing.T) {
	cp := new(KingoCuiPresenter)
	g := newKingoForPresenter(t)
	require.Equal(t, domain.KingoPhaseBet, g.GetPhase())

	out := cp.Output(g, nil)
	// 役の名前はどれも出ていない。
	for _, r := range []domain.KingoRank{
		domain.KingoRankArashi, domain.KingoRankPair,
	} {
		assert.NotContains(t, out, kingoRankLabel(r),
			"配る前に役が出ている (%s)", domain.KingoRankName(r))
	}

	// 決着後は全員ぶんの役が出る。
	after := cp.Output(kingoSettled(t), nil)
	assert.Contains(t, after, "----------")
	shown := false
	for _, r := range []domain.KingoRank{
		domain.KingoRankArashi, domain.KingoRankPair, domain.KingoRankNone,
	} {
		if countOccurrences(after, kingoRankLabel(r)) > 0 {
			shown = true
		}
	}
	assert.True(t, shown, "決着後に役が出ていない")
}

func kingoRankLabel(r domain.KingoRank) string {
	switch r {
	case domain.KingoRankArashi:
		return "嵐"
	case domain.KingoRankPair:
		return "2枚そろい"
	default:
		return "役なし"
	}
}

// **親と子で求める操作が違う。** 親には配るよう、子には張るよう促す。
func TestKingoCuiPresenter_PromptsForTheRightAction(t *testing.T) {
	cp := new(KingoCuiPresenter)

	child := cp.Output(kingoAsChild(t), nil)
	assert.Contains(t, child, "張ってください")
	assert.NotContains(t, child, "deal で配ってください", "子に配る案内が出ている")

	// 人間が親の回。
	g := newKingoForPresenter(t)
	for range 10 {
		if g.IsHumanBanker() {
			break
		}
		require.NoError(t, g.PlaceBet(g.GetConfig().MinBet))
		require.NoError(t, g.NextRound())
	}
	require.True(t, g.IsHumanBanker(), "人間が親のラウンドに届かない")
	banker := cp.Output(g, nil)
	assert.Contains(t, banker, "あなたが親です")
	assert.NotContains(t, banker, "張ってください", "親に張りの案内が出ている")
}

func TestKingoCuiPresenter_ErrorsAndHint(t *testing.T) {
	cp := new(KingoCuiPresenter)
	g := kingoAsChild(t)
	assert.Contains(t, cp.Output(g, errors.New("張り額が範囲外です")), "張り額が範囲外です")
	assert.NotEmpty(t, cp.ActionLogOutput(g))

	hint := cp.HintOutput(g)
	assert.NotEmpty(t, hint)
	assert.NotContains(t, hint, "kingo.", "助言のキーが訳されていない")
}

// --- Web ---

func TestKingoWebPresenter_ArraysAreNeverNull(t *testing.T) {
	cp := new(KingoWebPresenter)
	var out map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(cp.Output(newKingoForPresenter(t), nil)), &out))
	assert.NotEqual(t, "null", string(out["seats"]))
}

// **張りの段階では誰の手札も送らない。** 配る前なので手札は存在しない。
func TestKingoWebPresenter_ShipsNoHandsBeforeTheDeal(t *testing.T) {
	cp := new(KingoWebPresenter)

	var got struct {
		Seats []struct {
			Cards []json.RawMessage `json:"cards"`
			Rank  int               `json:"rank"`
		} `json:"seats"`
		Phase int `json:"phase"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(newKingoForPresenter(t), nil)), &got))
	for i, s := range got.Seats {
		assert.Empty(t, s.Cards, "席 %d の手札が配る前にワイヤに乗っている", i)
		assert.Zero(t, s.Rank, "席 %d の役が配る前に出ている", i)
	}

	// 決着後は全員ぶんが載る。
	require.NoError(t, json.Unmarshal([]byte(cp.Output(kingoSettled(t), nil)), &got))
	for i, s := range got.Seats {
		assert.Len(t, s.Cards, domain.KingoHandSize, "席 %d の手札が開いていない", i)
	}
}

// **配当と枚数はサーバが載せる。** 画面に倍率を書き写させない。
func TestKingoWebPresenter_ShipsTheRules(t *testing.T) {
	cp := new(KingoWebPresenter)

	var got struct {
		HandSize     int  `json:"handSize"`
		PayoutArashi int  `json:"payoutArashi"`
		PayoutPair   int  `json:"payoutPair"`
		BankerSeat   int  `json:"bankerSeat"`
		RoundNumber  int  `json:"roundNumber"`
		Rounds       int  `json:"rounds"`
		IsHumanBankr bool `json:"isHumanBanker"`
		Seats        []struct {
			IsBanker bool `json:"isBanker"`
		} `json:"seats"`
	}
	g := newKingoForPresenter(t)
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))
	assert.Equal(t, domain.KingoHandSize, got.HandSize)
	assert.Equal(t, domain.KingoPayout(domain.KingoRankArashi), got.PayoutArashi)
	assert.Greater(t, got.PayoutArashi, got.PayoutPair, "嵐の配当が上でない")
	assert.Equal(t, g.GetBankerSeat(), got.BankerSeat)
	assert.Equal(t, 1, got.RoundNumber)
	assert.Equal(t, g.GetConfig().Rounds, got.Rounds)
	assert.Equal(t, g.IsHumanBanker(), got.IsHumanBankr)
	// 席の側にも親の印が立つ。
	assert.True(t, got.Seats[got.BankerSeat].IsBanker, "席の側に親の印が無い")
}

func TestKingoWebPresenter_ErrorAndHint(t *testing.T) {
	cp := new(KingoWebPresenter)
	g := kingoAsChild(t)

	var withErr struct {
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, errors.New("boom"))), &withErr))
	assert.Equal(t, "boom", withErr.Message)
	assert.NotEmpty(t, cp.ActionLogOutput(g))

	var hint struct {
		Action string `json:"action"`
		Amount int    `json:"amount"`
		Reason string `json:"reason"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.HintOutput(g)), &hint))
	assert.Equal(t, "bet", hint.Action)
	assert.Equal(t, g.GetConfig().MinBet, hint.Amount)
	assert.NotEmpty(t, hint.Reason)
}
