//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// speedWithBoard rebuilds a Speed game with an exact hand and centre piles, so
// the playable/unplayable branches are both reachable without depending on the
// shuffle.
func speedWithBoard(t *testing.T, hand []*domain.Card, piles [2]*domain.Card) *domain.Speed {
	t.Helper()
	s := setupSpeedWebTest()
	data, err := json.Marshal(s)
	assert.NoError(t, err)
	var raw map[string]json.RawMessage
	assert.NoError(t, json.Unmarshal(data, &raw))

	var players []json.RawMessage
	assert.NoError(t, json.Unmarshal(raw["ps"], &players))
	// 手札は SpeedPlayer -> GamePlayer("gp") -> Player("p") -> Cards("c") の奥。
	var human map[string]json.RawMessage
	assert.NoError(t, json.Unmarshal(players[0], &human))
	var gp map[string]json.RawMessage
	assert.NoError(t, json.Unmarshal(human["gp"], &gp))
	var pl map[string]json.RawMessage
	assert.NoError(t, json.Unmarshal(gp["p"], &pl))
	pl["c"], err = json.Marshal(hand)
	assert.NoError(t, err)
	gp["p"], err = json.Marshal(pl)
	assert.NoError(t, err)
	human["gp"], err = json.Marshal(gp)
	assert.NoError(t, err)
	players[0], err = json.Marshal(human)
	assert.NoError(t, err)
	raw["ps"], err = json.Marshal(players)
	assert.NoError(t, err)
	raw["cp"], err = json.Marshal(piles)
	assert.NoError(t, err)

	newData, err := json.Marshal(raw)
	assert.NoError(t, err)
	assert.NoError(t, json.Unmarshal(newData, s))
	return s
}

// **出せる札に印を付ける。**Web は PLAY フェーズ中ずっと出せる札にリングを出す
// のに、CUI は番号付き一覧だけで、2 枚の台札に対する ±1 (と K↔A) を毎回自分で
// 比べる必要があった (#4861)。
func TestSpeedCuiPresenter_MarksPlayableCards(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.SpeedCuiPresenter)

	t.Run("marks only the cards that fit a pile", func(t *testing.T) {
		s := speedWithBoard(t, []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 6, false),  // 台札 7 の隣 -> 出せる
			domain.NewCard(domain.CardDesignHeart, 10, false), // どちらにも付かない
			domain.NewCard(domain.CardDesignClover, 1, false), // K の隣 (ラップ) -> 出せる
		}, [2]*domain.Card{
			domain.NewCard(domain.CardDesignDiamond, 7, false),
			domain.NewCard(domain.CardDesignSpade, 13, false),
		})
		out := p.Output(s, nil)
		assert.Contains(t, out, "[0]SPADE 6*")
		assert.Contains(t, out, "[1]HEART 10 ")
		assert.NotContains(t, out, "[1]HEART 10*")
		assert.Contains(t, out, "[2]CLOVER 1*")
	})

	t.Run("marks nothing when the hand is stuck", func(t *testing.T) {
		s := speedWithBoard(t, []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 4, false),
			domain.NewCard(domain.CardDesignHeart, 10, false),
		}, [2]*domain.Card{
			domain.NewCard(domain.CardDesignDiamond, 7, false),
			domain.NewCard(domain.CardDesignSpade, 7, false),
		})
		assert.NotContains(t, p.Output(s, nil), "*")
	})
}

func TestSpeedCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	p := new(presenter.SpeedCuiPresenter)

	t.Run("initial state", func(t *testing.T) {
		s := setupSpeedWebTest()
		result := p.Output(s, nil)
		assert.Contains(t, result, "Speed (スピード)")
		assert.Contains(t, result, "CPU:")
		assert.Contains(t, result, "[台札]")
		assert.Contains(t, result, "あなた:")
	})

	t.Run("shows human hand cards", func(t *testing.T) {
		s := setupSpeedWebTest()
		result := p.Output(s, nil)
		assert.Contains(t, result, "手札")
		assert.Contains(t, result, "山札")
	})

	t.Run("shows error", func(t *testing.T) {
		s := setupSpeedWebTest()
		result := p.Output(s, errors.New("bad play"))
		assert.Contains(t, result, "bad play")
	})

	t.Run("shows win message", func(t *testing.T) {
		s := setupSpeedWebTest()
		data, _ := json.Marshal(s)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ge"], _ = json.Marshal(true)
		raw["wi"], _ = json.Marshal(0)
		raw["ph"], _ = json.Marshal(domain.SpeedPhaseGameEnd)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, s)

		result := p.Output(s, nil)
		assert.Contains(t, result, "あなたの勝ちです")
	})

	t.Run("shows lose message", func(t *testing.T) {
		s := setupSpeedWebTest()
		data, _ := json.Marshal(s)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ge"], _ = json.Marshal(true)
		raw["wi"], _ = json.Marshal(1)
		raw["ph"], _ = json.Marshal(domain.SpeedPhaseGameEnd)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, s)

		result := p.Output(s, nil)
		assert.Contains(t, result, "CPUの勝ちです")
	})

	t.Run("shows stuck message", func(t *testing.T) {
		s := setupSpeedWebTest()
		data, _ := json.Marshal(s)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ph"], _ = json.Marshal(domain.SpeedPhaseStuck)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, s)

		result := p.Output(s, nil)
		assert.Contains(t, result, "膠着状態")
		// The flip-command help line accompanies the stuck message.
		assert.Contains(t, result, "f / flip")
	})

	t.Run("does not show stuck help during normal play", func(t *testing.T) {
		s := setupSpeedWebTest()
		// Force the play phase: a freshly-dealt board can occasionally be stuck,
		// which would non-deterministically surface the stuck help.
		data, _ := json.Marshal(s)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ph"], _ = json.Marshal(domain.SpeedPhasePlay)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, s)

		result := p.Output(s, nil)
		assert.NotContains(t, result, "f / flip")
	})
}

func TestSpeedCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SpeedCuiPresenter)
	s := setupSpeedWebTest()
	result := p.ActionLogOutput(s)
	assert.NotEmpty(t, result)
}

func TestSpeedCuiPresenter_Timer(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	p := new(presenter.SpeedCuiPresenter)
	mockTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	presenter.SetSpeedCuiPresenterClock(p, func() time.Time {
		return mockTime
	})

	t.Run("tracks elapsed time and session best", func(t *testing.T) {
		s := setupSpeedWebTest()
		// Initial output: game starts
		out1 := p.Output(s, nil)
		assert.Contains(t, out1, "経過時間: 00:00")
		assert.NotContains(t, out1, "セッション自己ベスト")
		assert.NotContains(t, out1, "{{")

		// Advance time by 45 seconds
		mockTime = mockTime.Add(45 * time.Second)
		out2 := p.Output(s, nil)
		assert.Contains(t, out2, "経過時間: 00:45")
		assert.NotContains(t, out2, "{{")

		// Game ends at 45 seconds
		data, _ := json.Marshal(s)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ge"], _ = json.Marshal(true)
		raw["wi"], _ = json.Marshal(0)
		raw["ph"], _ = json.Marshal(domain.SpeedPhaseGameEnd)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, s)

		out3 := p.Output(s, nil)
		assert.Contains(t, out3, "経過時間: 00:45")
		assert.Contains(t, out3, "セッション自己ベスト: 00:45")
		assert.NotContains(t, out3, "{{")

		// Restart game
		mockTime = mockTime.Add(10 * time.Second)
		s.Reset()

		out4 := p.Output(s, nil)
		// Elapsed should reset to 00:00, best should be 00:45
		assert.Contains(t, out4, "経過時間: 00:00")
		assert.Contains(t, out4, "セッション自己ベスト: 00:45")

		// Advance by 30 seconds
		mockTime = mockTime.Add(30 * time.Second)
		out5 := p.Output(s, nil)
		assert.Contains(t, out5, "経過時間: 00:30")

		// End game again, beating previous best
		data, _ = json.Marshal(s)
		_ = json.Unmarshal(data, &raw)
		raw["ge"], _ = json.Marshal(true)
		raw["wi"], _ = json.Marshal(0)
		raw["ph"], _ = json.Marshal(domain.SpeedPhaseGameEnd)
		newData, _ = json.Marshal(raw)
		_ = json.Unmarshal(newData, s)

		out6 := p.Output(s, nil)
		assert.Contains(t, out6, "経過時間: 00:30")
		assert.Contains(t, out6, "セッション自己ベスト: 00:30")
	})
}
