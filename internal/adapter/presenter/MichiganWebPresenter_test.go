package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// michiganWebCard は指定デザイン・値のカードを生成するテストヘルパー。
func michiganWebCard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

// michiganWebSetHand は player の手札を差し替える (プレゼンターテスト用ヘルパー)。
func michiganWebSetHand(p *domain.MichiganPlayer, cards ...*domain.Card) {
	p.ClearHand()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// michiganResultGame は決定的な結果フェーズのゲームを組み立てる。
// humanWins=true なら seat0 (人間) がブードルを取って上がり、false なら seat1 (CPU) が上がる。
func michiganResultGame(humanWins bool) *domain.Michigan {
	g := domain.NewDefaultMichigan()
	cfg := g.GetConfig()
	cfg.PlayerCount = 3
	g.SetConfig(cfg)
	g.Reset()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).ClearHand()
		g.GetPlayer(i).SetChips(100)
	}
	g.SetRoundStartChipsForTest([]int{100, 100, 100})
	g.SetBoodleForTest(0, 30, -1) // boodle 0 = A of hearts, 30 chips
	if humanWins {
		michiganWebSetHand(g.GetPlayer(0), michiganWebCard(domain.CardDesignHeart, 1))
		michiganWebSetHand(g.GetPlayer(1), michiganWebCard(domain.CardDesignSpade, 5))
		michiganWebSetHand(g.GetPlayer(2), michiganWebCard(domain.CardDesignClover, 9))
		g.SetPlayStateForTest(0, 0, 0, 0)
		g.DoPlayForTest(0, 0)
	} else {
		michiganWebSetHand(g.GetPlayer(1), michiganWebCard(domain.CardDesignHeart, 1))
		michiganWebSetHand(g.GetPlayer(0), michiganWebCard(domain.CardDesignSpade, 5))
		michiganWebSetHand(g.GetPlayer(2), michiganWebCard(domain.CardDesignClover, 9))
		g.SetPlayStateForTest(0, 0, 1, 1)
		g.DoPlayForTest(1, 0)
	}
	return g
}

func TestMichiganWebPresenter_OutputBetPhase(t *testing.T) {
	g := domain.NewDefaultMichigan()
	p := new(presenter.MichiganWebPresenter)
	out := p.Output(g, nil)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Contains(t, decoded, "players")
	assert.Contains(t, decoded, "boodles")
	assert.Contains(t, decoded, "config")
	assert.Contains(t, decoded, "betBudget")
	assert.Contains(t, decoded, "playableIndices")
	assert.Contains(t, decoded, "currentPlayerIdx")
	boodles, ok := decoded["boodles"].([]any)
	require.True(t, ok)
	assert.Len(t, boodles, domain.MichiganBoodleCount)
	assert.Equal(t, "michigan.betPhase", decoded["messageCode"])
}

func TestMichiganWebPresenter_Error(t *testing.T) {
	g := domain.NewDefaultMichigan()
	p := new(presenter.MichiganWebPresenter)
	out := p.Output(g, errors.New("boom"))

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "boom", decoded["message"])
}

func TestMichiganWebPresenter_ResultHumanLose(t *testing.T) {
	g := michiganResultGame(false)
	p := new(presenter.MichiganWebPresenter)
	out := p.Output(g, nil)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, float64(domain.MichiganPhaseResult), decoded["phase"])
	assert.Equal(t, "michigan.roundEndHumanLose", decoded["messageCode"])
	assert.Equal(t, float64(1), decoded["winnerIdx"])
	// At result phase every player's hand is revealed.
	players, ok := decoded["players"].([]any)
	require.True(t, ok)
	human := players[0].(map[string]any)
	humanCards, ok := human["cards"].([]any)
	require.True(t, ok)
	assert.NotNil(t, humanCards)
}

func TestMichiganWebPresenter_ResultHumanWin(t *testing.T) {
	g := michiganResultGame(true)
	p := new(presenter.MichiganWebPresenter)
	out := p.Output(g, nil)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "michigan.roundEndHumanWin", decoded["messageCode"])
	assert.Equal(t, float64(0), decoded["winnerIdx"])
	// The boodle was collected by seat 0.
	boodles := decoded["boodles"].([]any)
	b0 := boodles[0].(map[string]any)
	assert.Equal(t, float64(0), b0["claimedBy"])
	assert.Equal(t, float64(0), b0["chips"])
}

// **配り方に依存させない。**PlaceHumanBet は配ってから CPU を回すので、人間が
// 一度も出せない配りを引くと、プレゼンターに渡る前にラウンドが終わってしまい
// michigan.roundEndHumanLose になる (#4506)。プレイフェーズに入った局だけを
// 対象にし、それが一度も起きなければ落とす。
func TestMichiganWebPresenter_PlayPhase(t *testing.T) {
	const attempts = 50
	p := new(presenter.MichiganWebPresenter)

	for i := 0; i < attempts; i++ {
		g := domain.NewDefaultMichigan()
		require.NoError(t, g.PlaceHumanBet(michiganWebEvenBet(g.GetBetBudget())))
		if g.GetPhase() != domain.MichiganPhasePlay {
			// この配りは人間の手番を迎えずにラウンドが終わった。配り直す。
			continue
		}

		out := p.Output(g, nil)
		var decoded map[string]any
		require.NoError(t, json.Unmarshal([]byte(out), &decoded))
		assert.Equal(t, float64(domain.MichiganPhasePlay), decoded["phase"])
		assert.Equal(t, "michigan.playPhase", decoded["messageCode"])
		return
	}
	t.Fatalf("no deal out of %d reached the play phase — the CPU drive may never be yielding", attempts)
}

// michiganWebEvenBet は budget を 4 分割した賭けスライスを返す。
func michiganWebEvenBet(budget int) []int {
	dist := make([]int, domain.MichiganBoodleCount)
	q := budget / domain.MichiganBoodleCount
	r := budget % domain.MichiganBoodleCount
	for i := range dist {
		dist[i] = q
		if i < r {
			dist[i]++
		}
	}
	return dist
}

func TestMichiganWebPresenter_GameEnd(t *testing.T) {
	g := michiganResultGame(true)
	cfg := g.GetConfig()
	cfg.TargetRounds = 1
	g.SetConfig(cfg)
	// Re-run a single-round game to completion.
	g.Reset()
	require.NoError(t, g.PlaceHumanBet(michiganWebEvenBet(g.GetBetBudget())))
	for i := 0; i < 300 && g.GetPhase() == domain.MichiganPhasePlay && g.IsHumanTurn(); i++ {
		pi := g.GetPlayableIndices()
		require.NotEmpty(t, pi)
		require.NoError(t, g.PlayCard(pi[0]))
	}
	require.True(t, g.GetGameEndFlag())

	p := new(presenter.MichiganWebPresenter)
	out := p.Output(g, nil)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, true, decoded["gameEndFlag"])
	assert.Contains(t, decoded["messageCode"], "michigan.result.")
}

func TestMichiganWebPresenter_HintOutput(t *testing.T) {
	g := domain.NewDefaultMichigan()
	require.NoError(t, g.PlaceHumanBet(michiganWebEvenBet(g.GetBetBudget())))
	p := new(presenter.MichiganWebPresenter)
	out := p.HintOutput(g)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Contains(t, decoded, "players")
}

func TestMichiganWebPresenter_ActionLog(t *testing.T) {
	g := michiganResultGame(false)
	p := new(presenter.MichiganWebPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestMichiganWebPresenterOutputCarriesTheHint(t *testing.T) {
	// **配りに依存させない。**ベット直後にヒントが出ない配りが 300 回中 1 回ある
	// ので、出る配りを引くまで回す。1000 回引けなければヒントが壊れている。
	for range 1000 {
		g := domain.NewDefaultMichigan()
		require.NoError(t, g.PlaceHumanBet(michiganWebEvenBet(g.GetBetBudget())))
		if g.GetHint() == nil {
			continue
		}

		var decoded map[string]any
		require.NoError(t, json.Unmarshal([]byte(new(presenter.MichiganWebPresenter).Output(g, nil)), &decoded))
		assert.Contains(t, decoded, "hint", "Output must carry the hint -- the frontend reads state.hint")
		// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
		assert.NotEqual(t, "michigan.hintRequested", decoded["messageCode"])
		return
	}
	t.Fatal("1000 deals and never once did a hint come back")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestMichiganWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	for range 1000 {
		g := domain.NewDefaultMichigan()
		require.NoError(t, g.PlaceHumanBet(michiganWebEvenBet(g.GetBetBudget())))
		if g.GetHint() == nil {
			continue
		}
		assert.Contains(t, new(presenter.MichiganWebPresenter).HintOutput(g), "michigan.hintRequested")
		return
	}
	t.Fatal("1000 deals and never once did a hint come back")
}
