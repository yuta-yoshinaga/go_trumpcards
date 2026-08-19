//go:build test

package presenter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newMendikotForWeb(t *testing.T) *domain.Mendikot {
	t.Helper()
	m := domain.NewDefaultMendikot()
	m.Reset()
	return m
}

func decodeMendikot(t *testing.T, str string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(str), &m))
	return m
}

func TestMendikotWebPresenterOutput(t *testing.T) {
	p := new(MendikotWebPresenter)
	m := decodeMendikot(t, new(MendikotWebPresenter).Output(newMendikotForWeb(t), nil))
	_ = p

	assert.Equal(t, float64(domain.MendikotPhasePlay), m["phase"], "配り終えたら即プレイ、切り札フェーズは無い")
	assert.Equal(t, float64(1), m["handNumber"])
	assert.Equal(t, float64(0), m["trumpSuit"], "切り札はまだ決まっていない")
	assert.Equal(t, float64(-1), m["trumpChooserIdx"])
	assert.Equal(t, float64(domain.MendikotTensInDeck), m["tensInDeck"])
	assert.Equal(t, float64(-1), m["lastHandWinner"])

	players := m["players"].([]any)
	require.Len(t, players, domain.MendikotPlayerCnt)
	human := players[0].(map[string]any)
	assert.True(t, human["isHuman"].(bool))
	assert.Equal(t, float64(domain.MendikotHandSize), human["cardCount"])
	assert.Equal(t, float64(0), human["team"])
	assert.Equal(t, float64(0), players[2].(map[string]any)["team"], "0 と 2 が味方")
	assert.Equal(t, float64(1), players[1].(map[string]any)["team"])
	assert.Len(t, human["cards"], domain.MendikotHandSize, "人間の手札だけ見える")
	assert.Empty(t, players[1].(map[string]any)["cards"], "CPU の手札は伏せる")
}

// **勝敗を決めるのは 10 の枚数。** 盤面から読めないので必ずワイヤに載せる。
func TestMendikotWebPresenterCarriesTheTenCount(t *testing.T) {
	p := new(MendikotWebPresenter)
	m := newMendikotForWeb(t)
	m.GetPlayer(0).SetTens(2)
	m.GetPlayer(2).SetTens(1)
	m.GetPlayer(1).SetTens(1)

	out := decodeMendikot(t, p.Output(m, nil))
	tens := out["teamTens"].([]any)
	require.Len(t, tens, domain.MendikotTeamCnt)
	assert.Equal(t, float64(3), tens[0], "味方 2 人分を合算する")
	assert.Equal(t, float64(1), tens[1])
}

// **同点でトリック数が効く経路もワイヤに載る。**
func TestMendikotWebPresenterCarriesTrickCounts(t *testing.T) {
	p := new(MendikotWebPresenter)
	m := newMendikotForWeb(t)
	m.GiveTricksForTest(0, 3)
	m.GiveTricksForTest(3, 2)

	tricks := decodeMendikot(t, p.Output(m, nil))["teamTricks"].([]any)
	assert.Equal(t, float64(3), tricks[0])
	assert.Equal(t, float64(2), tricks[1])
}

// **決まり方でハンドの勝ち点が 1/2/3 と変わる。** どの決まり方かをコードで返す。
func TestMendikotWebPresenterReportsTheHandOutcome(t *testing.T) {
	p := new(MendikotWebPresenter)
	m := newMendikotForWeb(t)
	m.GetPlayer(0).SetTens(domain.MendikotTensInDeck)
	m.SetTrickNumberForTest(domain.MendikotTricksPerRound - 1)
	m.FinishHandForTest()

	out := decodeMendikot(t, p.Output(m, nil))
	assert.Equal(t, "mendikot", out["lastHandKind"])
	assert.Equal(t, float64(0), out["lastHandWinner"])
	assert.Equal(t, "mendikot.handEnd.mendikot", out["messageCode"])
	assert.Equal(t, float64(domain.MendikotMendikotPoints), out["scores"].([]any)[0])
}

// **切り札が決まる前と後で案内が違う。**
func TestMendikotWebPresenterMessageTracksTrump(t *testing.T) {
	p := new(MendikotWebPresenter)

	fresh := newMendikotForWeb(t)
	assert.Equal(t, "mendikot.play.noTrump", decodeMendikot(t, p.Output(fresh, nil))["messageCode"])

	decided := newMendikotForWeb(t)
	decided.SetTrumpForTest(domain.CardDesignHeart, 2)
	out := decodeMendikot(t, p.Output(decided, nil))
	assert.Equal(t, "mendikot.play", out["messageCode"])
	assert.Equal(t, float64(domain.CardDesignHeart), out["trumpSuit"])
	assert.Equal(t, float64(2), out["trumpChooserIdx"])
}

func TestMendikotWebPresenterSurfacesErrors(t *testing.T) {
	p := new(MendikotWebPresenter)
	m := newMendikotForWeb(t)
	err := m.PlayerPlay(999)
	require.Error(t, err)

	out := decodeMendikot(t, p.Output(m, err))
	assert.Equal(t, err.Error(), out["message"])
	assert.Empty(t, out["messageCode"], "エラー時はコードを立てない")
}

func TestMendikotWebPresenterGameEnd(t *testing.T) {
	p := new(MendikotWebPresenter)
	m := newMendikotForWeb(t)
	// **勝者は得点から決まる。** winnerTeam を直接置いても finishGame が上書きする。
	m.SetScoreForTestUse(0, 2)
	m.SetScoreForTestUse(1, m.GetConfig().Target)
	m.FinishGameForTest()

	out := decodeMendikot(t, p.Output(m, nil))
	assert.True(t, out["gameEndFlag"].(bool))
	assert.Equal(t, float64(1), out["winnerTeam"])
	assert.Equal(t, "mendikot.result.team1", out["messageCode"])
}

func TestMendikotWebPresenterHintAndLog(t *testing.T) {
	p := new(MendikotWebPresenter)
	m := newMendikotForWeb(t)
	m.SetCurrentPlayerIdxForTest(0)

	hint := decodeMendikot(t, p.HintOutput(m))
	require.NotNil(t, hint["hint"])
	h := hint["hint"].(map[string]any)
	assert.NotEmpty(t, h["reason"])
	idx := int(h["cardIndex"].(float64))
	assert.Contains(t, m.GetValidPlayIndices(0), idx, "勧める札は必ず合法")

	var logOut map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(m)), &logOut))
	assert.Contains(t, logOut, "entries")
}

// **切り札は宣言ではなく事故で決まる** (#5755)。画面で作り直すのではなく、
// ドメインの判定をそのまま配る。
func TestMendikotWebPresenterFlagsTheTrumpSettingTurn(t *testing.T) {
	setup := func(hand []*domain.Card, trump int) *domain.Mendikot {
		m := newMendikotForWeb(t)
		m.SetPhaseForTest(domain.MendikotPhasePlay)
		m.SetTrumpForTest(trump, -1)
		m.SetCurrentPlayerIdxForTest(0)
		m.SetCurrentTrickForTest([]*domain.TrickCard{
			{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 9, true)},
		})
		human := m.GetPlayer(0)
		human.ResetRound()
		for _, c := range hand {
			human.AddCard(c)
		}
		return m
	}

	cannotFollow := setup([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 5, true),
		domain.NewCard(domain.CardDesignClover, 8, true),
	}, 0)
	assert.True(t, decodeMendikot(t, new(MendikotWebPresenter).Output(cannotFollow, nil))["willSetTrump"].(bool))

	canFollow := setup([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, true),
		domain.NewCard(domain.CardDesignHeart, 5, true),
	}, 0)
	assert.False(t, decodeMendikot(t, new(MendikotWebPresenter).Output(canFollow, nil))["willSetTrump"].(bool))

	decided := setup([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 5, true),
		domain.NewCard(domain.CardDesignClover, 8, true),
	}, domain.CardDesignHeart)
	assert.False(t, decodeMendikot(t, new(MendikotWebPresenter).Output(decided, nil))["willSetTrump"].(bool))
}
