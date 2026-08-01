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

func TestKingWebPresenter_Output(t *testing.T) {
	g := domain.NewDefaultKing()
	g.Reset()
	g.SetCurrentContract(domain.KingContractKingTrump)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetPhase(domain.KingPhasePlay)
	g.SetCurrentTurn(0)
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: bcard(domain.CardDesignHeart, 9)}})
	p := new(presenter.KingWebPresenter)
	out := p.Output(g, nil)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, domain.KingPhasePlay, decoded["phase"])
	assert.Equal(t, float64(domain.KingTotalDeals), decoded["totalDeals"])
	players, ok := decoded["players"].([]any)
	require.True(t, ok)
	assert.Len(t, players, domain.KingPlayerCnt)
	assert.Contains(t, decoded, "usedContracts")
	assert.Contains(t, decoded, "currentTrick")
}

func TestKingWebPresenter_Error(t *testing.T) {
	g := domain.NewDefaultKing()
	g.Reset()
	p := new(presenter.KingWebPresenter)
	out := p.Output(g, errors.New("boom"))
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "boom", decoded["message"])
}

func TestKingWebPresenter_GameEnd(t *testing.T) {
	g := domain.NewDefaultKing()
	cfg := g.GetConfig()
	cfg.CpuDifficulty = domain.KingDifficultyEasy
	g.SetConfig(cfg)
	g.Reset()

	// Drive a CPU-only game to completion (dealer 0 is human; still pick lowest).
	guard := 0
	for !g.GetGameEndFlag() && guard < 20000 {
		guard++
		switch g.GetPhase() {
		case domain.KingPhaseSelectContract:
			if g.GetPlayer(g.GetDealerIdx()).GetIsHuman() {
				used := g.GetUsedContracts()
				c := 0
				for c < domain.KingContractCnt && used[c] {
					c++
				}
				trump := -1
				if c == domain.KingContractKingTrump {
					trump = domain.CardDesignSpade
				}
				require.NoError(t, g.SelectContract(c, trump))
			} else {
				g.CpuPlay()
			}
		case domain.KingPhasePlay:
			if g.IsHumanTurn() {
				valid := g.GetPlayableIndices(g.GetCurrentTurn())
				require.NotEmpty(t, valid)
				require.NoError(t, g.PlayerPlay(valid[0]))
			} else {
				g.CpuPlay()
			}
		case domain.KingPhaseDealEnd:
			g.NextDeal()
		}
	}
	require.True(t, g.GetGameEndFlag())

	p := new(presenter.KingWebPresenter)
	out := p.Output(g, nil)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, true, decoded["gameEndFlag"])
	assert.Equal(t, "king.result.scores", decoded["messageCode"])
	assert.NotEmpty(t, decoded["roundWinners"])
	assert.NotNil(t, decoded["lastDealDetail"])
}

func TestKingWebPresenter_HintOutput(t *testing.T) {
	g := domain.NewDefaultKing()
	g.Reset()
	g.SetCurrentContract(domain.KingContractNoTricks)
	g.SetTrumpSuit(-1)
	g.SetPhase(domain.KingPhasePlay)
	g.SetCurrentTurn(0)
	p := new(presenter.KingWebPresenter)
	out := p.HintOutput(g)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	// Hint present when it's the human's turn in play phase.
	assert.Contains(t, decoded, "hint")
}

func TestKingWebPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultKing()
	g.Reset()
	p := new(presenter.KingWebPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestKingWebPresenterOutputCarriesTheHint(t *testing.T) {
	// 既存の HintOutput テストと同じ状態。人間の手番のプレイフェーズ。
	g := domain.NewDefaultKing()
	g.Reset()
	g.SetCurrentContract(domain.KingContractNoTricks)
	g.SetTrumpSuit(-1)
	g.SetPhase(domain.KingPhasePlay)
	g.SetCurrentTurn(0)
	require.NotNil(t, g.GetHint(), "fixture must actually produce a hint")

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(new(presenter.KingWebPresenter).Output(g, nil)), &decoded))
	assert.Contains(t, decoded, "hint", "Output must carry the hint -- the frontend reads state.hint")
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	assert.NotEqual(t, "king.hintRequested", decoded["messageCode"])
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestKingWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	g := domain.NewDefaultKing()
	g.Reset()
	g.SetCurrentContract(domain.KingContractNoTricks)
	g.SetTrumpSuit(-1)
	g.SetPhase(domain.KingPhasePlay)
	g.SetCurrentTurn(0)

	assert.Contains(t, new(presenter.KingWebPresenter).HintOutput(g), "king.hintRequested")
}
