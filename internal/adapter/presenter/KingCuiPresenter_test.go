package presenter_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestKingCuiPresenter_Output(t *testing.T) {
	g := domain.NewDefaultKing()
	g.Reset()
	p := new(presenter.KingCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "1/7") // deal line
	assert.Contains(t, out, "[0]") // human indexed hand
}

// kingLineContaining returns the first line of out that contains marker.
func kingLineContaining(out, marker string) string {
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, marker) {
			return line
		}
	}
	return ""
}

func TestKingCuiPresenter_RemainingContracts(t *testing.T) {
	p := new(presenter.KingCuiPresenter)
	prefix := strings.Split(i18n.T("king.remainingContracts"), "{{")[0]

	// Right after reset every contract is available.
	g := domain.NewDefaultKing()
	g.Reset()
	remaining := kingLineContaining(p.Output(g, nil), prefix)
	assert.Contains(t, remaining, "[0]")
	assert.Contains(t, remaining, "[6]")

	// Once the dealer selects a contract it drops out of the remaining list.
	g2 := domain.NewDefaultKing()
	g2.Reset()
	g2.SetDealerIdx(0) // human dealer
	require.NoError(t, g2.SelectContract(domain.KingContractNoTricks, 0))
	g2.SetPhase(domain.KingPhaseSelectContract) // force back to the selection view
	remaining2 := kingLineContaining(p.Output(g2, nil), prefix)
	assert.NotContains(t, remaining2, "[0]") // used contract 0 excluded
	assert.Contains(t, remaining2, "[6]")
}

func TestKingCuiPresenter_ContractAndTrick(t *testing.T) {
	g := domain.NewDefaultKing()
	g.Reset()
	g.SetCurrentContract(domain.KingContractKingTrump)
	g.SetTrumpSuit(domain.CardDesignSpade)
	g.SetPhase(domain.KingPhasePlay)
	g.SetCurrentTurn(0)
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: bcard(domain.CardDesignHeart, 9)}})
	p := new(presenter.KingCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
}

func TestKingCuiPresenter_DealEndAndGameEnd(t *testing.T) {
	p := new(presenter.KingCuiPresenter)

	g := domain.NewDefaultKing()
	g.Reset()
	g.SetPhase(domain.KingPhaseDealEnd)
	assert.NotEmpty(t, p.Output(g, nil))

	// Error output.
	assert.NotEmpty(t, p.Output(g, errors.New("boom")))
}

func TestKingCuiPresenter_HintOutput(t *testing.T) {
	p := new(presenter.KingCuiPresenter)

	t.Run("negative contract avoid", func(t *testing.T) {
		g := domain.NewDefaultKing()
		g.Reset()
		g.SetCurrentContract(domain.KingContractNoTricks)
		g.SetTrumpSuit(-1)
		g.SetPhase(domain.KingPhasePlay)
		g.SetCurrentTurn(0)
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(bcard(domain.CardDesignSpade, 2))
		g.GetPlayer(0).AddCard(bcard(domain.CardDesignHeart, 13))
		out := p.HintOutput(g)
		assert.NotEmpty(t, out)
	})

	t.Run("positive contract win", func(t *testing.T) {
		g := domain.NewDefaultKing()
		g.Reset()
		g.SetCurrentContract(domain.KingContractKingTrump)
		g.SetTrumpSuit(domain.CardDesignSpade)
		g.SetPhase(domain.KingPhasePlay)
		g.SetCurrentTurn(0)
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(bcard(domain.CardDesignSpade, 1))
		out := p.HintOutput(g)
		assert.NotEmpty(t, out)
	})

	t.Run("no hint outside play phase", func(t *testing.T) {
		g := domain.NewDefaultKing()
		g.Reset()
		g.SetPhase(domain.KingPhaseSelectContract)
		assert.Contains(t, p.HintOutput(g), "")
	})
}

func TestKingCuiPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultKing()
	g.Reset()
	p := new(presenter.KingCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// kingPlayToDealEnd plays one full deal of the given contract and returns the game
// sitting in DealEnd. The deal is shuffled, so the *gains* differ run to run — the
// assertions read them back from the domain rather than hard-coding them.
func kingPlayToDealEnd(t *testing.T, contract, trump int) *domain.King {
	t.Helper()
	g := domain.NewDefaultKing()
	g.Reset()
	require.True(t, g.GetPlayer(g.GetDealerIdx()).GetIsHuman(), "deal 1 is dealt by the human")
	require.NoError(t, g.SelectContract(contract, trump))
	for i := 0; i < 200 && g.GetPhase() == domain.KingPhasePlay; i++ {
		if g.IsHumanTurn() {
			valid := g.GetPlayableIndices(g.GetCurrentTurn())
			require.NotEmpty(t, valid)
			require.NoError(t, g.PlayerPlay(valid[0]))
		} else {
			g.CpuPlay()
		}
	}
	require.Equal(t, domain.KingPhaseDealEnd, g.GetPhase())
	return g
}

// #5691: Web の king-deal-breakdown はコントラクト名・切り札・各人の獲得点を出すのに、
// CUI は「ディール終了。n で次のディールへ。」という固定案内だけで、何のコントラクトを
// 誰が何点で終えたのかはその場では分からなかった。
func TestKingCuiPresenter_DealEndBreakdown(t *testing.T) {
	p := new(presenter.KingCuiPresenter)

	gainedPrefix := strings.Split(i18n.T("king.dealResultGained"), "{{")[0]

	t.Run("lists what each player gained on this deal", func(t *testing.T) {
		g := kingPlayToDealEnd(t, domain.KingContractNoHearts, -1)
		detail := g.GetLastDealDetail()
		require.NotNil(t, detail)

		line := kingLineContaining(p.Output(g, nil), gainedPrefix)

		for i := 0; i < domain.KingPlayerCnt; i++ {
			assert.Contains(t, line, strconv.Itoa(detail.Gained[i]),
				"player %d gained %d", i, detail.Gained[i])
		}
	})

	// 累計得点はプレイヤー行に出ているので、**このディールぶん**と取り違えないこと。
	t.Run("is the deal's gain, not the running total", func(t *testing.T) {
		g := kingPlayToDealEnd(t, domain.KingContractKingTrump, domain.CardDesignHeart)
		gained := g.GetLastDealDetail().Gained
		// 累計だけを、桁数の違う一意な値へ動かす。0 点のディールだと
		// 「累計を出す」実装でも `0` が部分一致で通ってしまうため。
		for i := 0; i < domain.KingPlayerCnt; i++ {
			g.GetPlayer(i).AddScore(5000 + i*111)
		}

		line := kingLineContaining(p.Output(g, nil), gainedPrefix)

		for i := 0; i < domain.KingPlayerCnt; i++ {
			assert.Contains(t, line, strconv.Itoa(gained[i]))
			assert.NotContains(t, line, strconv.Itoa(g.GetPlayer(i).GetTotalScore()),
				"the running total must not be here")
		}
	})

	t.Run("says nothing before the first deal settles", func(t *testing.T) {
		g := domain.NewDefaultKing()
		g.Reset()

		assert.NotContains(t, p.Output(g, nil), gainedPrefix)
	})
}

// レビュー指摘 (#6026): 7 ディール目は finishDeal が gameEndFlag を立てて
// DealEnd を飛ばすので、DealEnd 分岐だけに置くと**最後のディールの内訳が出ない**。
// 決着を決めた 1 ディールこそ見たいので、GameEnd でも出す。
func TestKingCuiPresenter_DealEndBreakdownOnTheDecidingDeal(t *testing.T) {
	p := new(presenter.KingCuiPresenter)
	g := domain.NewDefaultKing()
	g.Reset()
	for i := 0; i < 4000 && !g.GetGameEndFlag(); i++ {
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
	gained := g.GetLastDealDetail().Gained

	line := kingLineContaining(p.Output(g, nil),
		strings.Split(i18n.T("king.dealResultGained"), "{{")[0])

	require.NotEmpty(t, line, "the deciding deal must show its breakdown too")
	for i := 0; i < domain.KingPlayerCnt; i++ {
		assert.Contains(t, line, strconv.Itoa(gained[i]))
	}
}
