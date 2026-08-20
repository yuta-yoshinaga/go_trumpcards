package presenter_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newSakuraForPresenter() *domain.Sakura {
	g := domain.NewDefaultSakura()
	g.Reset()
	return g
}

func TestSakuraWebPresenter_Output(t *testing.T) {
	g := newSakuraForPresenter()
	out := new(presenter.SakuraWebPresenter).Output(g, nil)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, float64(domain.SakuraPhasePlay), decoded["phase"])
	assert.Equal(t, float64(1), decoded["round"])
	assert.Equal(t, float64(domain.SakuraDefaultRounds), decoded["totalRounds"])
	assert.Equal(t, float64(domain.SakuraFieldSize), float64(len(decoded["fieldCards"].([]any))))
	assert.Equal(t, float64(g.GetStockCount()), decoded["stockCount"])
	assert.True(t, decoded["isHumanTurn"].(bool))
	players := decoded["players"].([]any)
	assert.Len(t, players, domain.SakuraDefaultSeats)
	assert.Contains(t, decoded, "captureOptions")
	// 未決着なので勝者は -1。
	assert.Equal(t, float64(-1), decoded["winner"])
}

// **人間の手札だけが表向き。** 相手の手札が見えると勝負にならない。
func TestSakuraWebPresenter_HidesOpponentHands(t *testing.T) {
	g := newSakuraForPresenter()
	out := new(presenter.SakuraWebPresenter).Output(g, nil)

	type webCard struct {
		Deck  string `json:"deck"`
		Glyph string `json:"glyph"`
		Label string `json:"label"`
		Draw  bool   `json:"draw"`
	}
	var parsed struct {
		Players []struct {
			IsHuman   bool       `json:"isHuman"`
			CardCount int        `json:"cardCount"`
			Cards     []*webCard `json:"cards"`
		} `json:"players"`
		FieldCards []*webCard `json:"fieldCards"`
	}
	require.NoError(t, json.Unmarshal([]byte(out), &parsed))

	require.Len(t, parsed.FieldCards, domain.SakuraFieldSize)
	for _, c := range parsed.FieldCards {
		assert.Equal(t, "hanafuda", c.Deck, "花札は手続き描画 (ADR-0033)")
		assert.NotEmpty(t, c.Glyph)
		assert.NotEmpty(t, c.Label)
	}
	humanSeen, cpuSeen := false, false
	for _, p := range parsed.Players {
		// 枚数はどちらの席も出す。伏せた札そのものは相手の席には出さない。
		assert.Equal(t, domain.SakuraHandSize, p.CardCount)
		if p.IsHuman {
			humanSeen = true
			require.Len(t, p.Cards, domain.SakuraHandSize, "自分の手札が見えない")
			for _, c := range p.Cards {
				assert.NotEmpty(t, c.Glyph)
				assert.Equal(t, "hanafuda", c.Deck)
			}
			continue
		}
		cpuSeen = true
		assert.Empty(t, p.Cards, "相手の手札が出力に載っている")
	}
	assert.True(t, humanSeen && cpuSeen, "両方の席を確かめていない")
}

func TestSakuraWebPresenter_CaptureOptionsOnlyOnTheHumanTurn(t *testing.T) {
	g := newSakuraForPresenter()
	p := new(presenter.SakuraWebPresenter)

	var human struct {
		CaptureOptions map[string][]int `json:"captureOptions"`
	}
	require.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &human))
	assert.Equal(t, len(g.GetValidFieldIndices()), len(human.CaptureOptions))

	// CPU の手番では合わせ先を配らない (相手の手札が推測できてしまう)。
	for range domain.SakuraHandSize {
		if !g.IsHumanTurn() {
			break
		}
		h := g.GetHint()
		require.NoError(t, g.PlayerPlay(h.CardIndex, h.FieldIndex))
	}
	g.SetConfig(g.GetConfig()) // 状態は変えない。
	if !g.IsHumanTurn() {
		var cpu struct {
			CaptureOptions map[string][]int `json:"captureOptions"`
		}
		require.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &cpu))
		assert.Empty(t, cpu.CaptureOptions)
	}
}

func TestSakuraWebPresenter_Error(t *testing.T) {
	g := newSakuraForPresenter()
	out := new(presenter.SakuraWebPresenter).Output(g, errors.New("boom"))

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "boom", decoded["message"])
}

func TestSakuraWebPresenter_GameEndCarriesScores(t *testing.T) {
	g := domain.NewSakura(domain.NewSakuraPlayersForTable(2), domain.SakuraConfig{Seats: 2, Rounds: 1})
	g.Reset()
	for range 500 {
		if g.GetGameEndFlag() {
			break
		}
		if g.IsHumanTurn() {
			h := g.GetHint()
			require.NoError(t, g.PlayerPlay(h.CardIndex, h.FieldIndex))
			continue
		}
		g.CpuPlay()
	}
	require.True(t, g.GetGameEndFlag())

	var decoded struct {
		GameEndFlag   bool              `json:"gameEndFlag"`
		MessageCode   string            `json:"messageCode"`
		MessageParams map[string]string `json:"messageParams"`
		LastResult    struct {
			Round  int `json:"round"`
			Winner int `json:"winner"`
			Seats  []struct {
				CardPoints  int `json:"cardPoints"`
				BonusPoints int `json:"bonusPoints"`
				Total       int `json:"total"`
			} `json:"seats"`
		} `json:"lastResult"`
	}
	require.NoError(t, json.Unmarshal([]byte(new(presenter.SakuraWebPresenter).Output(g, nil)), &decoded))
	assert.True(t, decoded.GameEndFlag)
	assert.Equal(t, "sakura.result.scores", decoded.MessageCode)
	assert.Contains(t, decoded.MessageParams["scores"], "0:")
	require.Len(t, decoded.LastResult.Seats, 2)
	for i, s := range decoded.LastResult.Seats {
		assert.Equal(t, s.CardPoints+s.BonusPoints, s.Total, "席 %d の内訳が合計と合わない", i)
		assert.Equal(t, g.GetPlayer(i).TotalPoints(), s.Total)
	}
}

// 追加役が出力に乗る。
func TestSakuraWebPresenter_ReportsBonuses(t *testing.T) {
	g := newSakuraForPresenter()
	g.GetPlayer(0).AddTaken(
		domain.NewCard(domain.SakuraCurtainMonth, domain.SakuraCurtainIndex, false),
		domain.NewCard(domain.SakuraMoonMonth, domain.SakuraMoonIndex, false),
	)

	var decoded struct {
		Players []struct {
			Bonuses []struct {
				Key    string `json:"key"`
				Points int    `json:"points"`
			} `json:"bonuses"`
			BonusPoints int `json:"bonusPoints"`
			CardPoints  int `json:"cardPoints"`
			TotalPoints int `json:"totalPoints"`
		} `json:"players"`
	}
	require.NoError(t, json.Unmarshal([]byte(new(presenter.SakuraWebPresenter).Output(g, nil)), &decoded))
	require.NotEmpty(t, decoded.Players)
	require.Len(t, decoded.Players[0].Bonuses, 1)
	assert.Equal(t, "sakuraSake", decoded.Players[0].Bonuses[0].Key)
	assert.Equal(t, 30, decoded.Players[0].Bonuses[0].Points)
	assert.Equal(t, 30, decoded.Players[0].BonusPoints)
	assert.Equal(t, 40, decoded.Players[0].CardPoints)
	assert.Equal(t, 70, decoded.Players[0].TotalPoints)
}

func TestSakuraWebPresenter_HintOutput(t *testing.T) {
	g := newSakuraForPresenter()
	var decoded struct {
		Hint *struct {
			CardIndex  int    `json:"cardIndex"`
			FieldIndex int    `json:"fieldIndex"`
			Reason     string `json:"reason"`
		} `json:"hint"`
	}
	require.NoError(t, json.Unmarshal([]byte(new(presenter.SakuraWebPresenter).HintOutput(g)), &decoded))
	require.NotNil(t, decoded.Hint)
	assert.GreaterOrEqual(t, decoded.Hint.CardIndex, 0)
	assert.Contains(t, []string{"capture", "discard"}, decoded.Hint.Reason)
}

// 人間の手番でなければヒントを載せない。
func TestSakuraWebPresenter_NoHintOutsideTheHumanTurn(t *testing.T) {
	g := newSakuraForPresenter()
	for range 500 {
		if !g.IsHumanTurn() || g.GetPhase() != domain.SakuraPhasePlay {
			break
		}
		h := g.GetHint()
		require.NoError(t, g.PlayerPlay(h.CardIndex, h.FieldIndex))
	}
	out := new(presenter.SakuraWebPresenter).Output(g, nil)
	assert.NotContains(t, out, `"hint"`)
}

// 棋譜は終局まで伏せる (途中で配り札が読めてしまうため)。
func TestSakuraWebPresenter_ActionLogOutput(t *testing.T) {
	g := newSakuraForPresenter()
	p := new(presenter.SakuraWebPresenter)

	var midway struct {
		Entries []map[string]any `json:"entries"`
	}
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(g)), &midway))
	assert.Empty(t, midway.Entries, "終局前に棋譜を出している")

	sakuraPlayOut(t, g)
	var ended struct {
		Entries []struct {
			ActionType string `json:"actionType"`
			Detail     string `json:"detail"`
		} `json:"entries"`
	}
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(g)), &ended))
	require.NotEmpty(t, ended.Entries)
	types := map[string]int{}
	for _, e := range ended.Entries {
		types[e.ActionType]++
	}
	assert.Positive(t, types["deal"], "配りが記録されていない")
	assert.Positive(t, types["play"], "着手が記録されていない")
	assert.Positive(t, types["round"], "ラウンド結果が記録されていない")
}

// sakuraPlayOut は終局まで打ち切る。
func sakuraPlayOut(t *testing.T, g *domain.Sakura) {
	t.Helper()
	for range 2000 {
		if g.GetGameEndFlag() {
			return
		}
		if g.GetPhase() == domain.SakuraPhaseRoundEnd {
			g.NextRound()
			continue
		}
		if g.IsHumanTurn() {
			h := g.GetHint()
			require.NoError(t, g.PlayerPlay(h.CardIndex, h.FieldIndex))
			continue
		}
		g.CpuPlay()
	}
	t.Fatal("終局しなかった")
}

// 出力に「捨てた」情報がそのまま残ると、相手の手札が読めてしまう。
func TestSakuraWebPresenter_DoesNotLeakOpponentHandInTheLog(t *testing.T) {
	g := newSakuraForPresenter()
	h := g.GetHint()
	require.NoError(t, g.PlayerPlay(h.CardIndex, h.FieldIndex))
	sakuraPlayOut(t, g)
	out := new(presenter.SakuraWebPresenter).ActionLogOutput(g)
	// 棋譜は「出した/めくった」札しか含まない。伏せたままの手札は載らない。
	assert.NotContains(t, out, `"cards":null`)
	assert.Positive(t, strings.Count(out, `"actionType"`))
}

// **点数を札そのものに書く** (#5785)。さくらは役ではなく点数の合計で競うので、
// どの札が何点かが読めないと打ち手を決められない——CUI は最初からそう出している。
func TestSakuraWebPresenter_SendsPointsWithEveryCard(t *testing.T) {
	g := newSakuraForPresenter()
	var out struct {
		FieldCards []struct {
			Points *int `json:"points"`
		} `json:"fieldCards"`
		Players []struct {
			IsHuman bool `json:"isHuman"`
			Cards   []struct {
				Points *int `json:"points"`
			} `json:"cards"`
		} `json:"players"`
	}
	require.NoError(t, json.Unmarshal([]byte(new(presenter.SakuraWebPresenter).Output(g, nil)), &out))

	require.NotEmpty(t, out.FieldCards)
	for i, c := range out.FieldCards {
		require.NotNil(t, c.Points, "場札 %d に点数が無い", i)
		assert.Equal(t, domain.SakuraCardPoints(g.GetField()[i]), *c.Points)
	}
	humans := 0
	for _, p := range out.Players {
		if !p.IsHuman {
			assert.Empty(t, p.Cards, "CPU の手札は出さない")
			continue
		}
		humans++
		require.NotEmpty(t, p.Cards)
		for i, c := range p.Cards {
			require.NotNil(t, c.Points, "手札 %d に点数が無い", i)
			assert.Equal(t, domain.SakuraCardPoints(g.GetPlayer(0).GetCard(i)), *c.Points)
		}
	}
	require.Equal(t, 1, humans, "人間の席が 1 つでない")
}
