//go:build test

package presenter_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupContractRummyCuiMock(phase domain.ContractRummyPhase, gameEnd bool) (*interfaces.MockContractRummyGame, []*domain.ContractRummyPlayer) {
	m := new(interfaces.MockContractRummyGame)
	players := []*domain.ContractRummyPlayer{
		domain.NewContractRummyPlayer(true),
		domain.NewContractRummyPlayer(false),
		domain.NewContractRummyPlayer(false),
	}
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(60)
	m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 7, false))
	m.On("GetGameEndFlag").Return(gameEnd)
	m.On("GetPhase").Return(phase)
	m.On("GetCurrentPlayerIdx").Return(0)
	winner := -1
	if gameEnd {
		winner = 0
	}
	m.On("GetWinnerIdx").Return(winner)
	m.On("GetConfig").Return(domain.DefaultContractRummyConfig())
	m.On("GetRoundWinnerIdx").Return(-1)
	m.On("GetCurrentContract").Return(domain.ContractForRound(1))
	m.On("GetPlayerCnt").Return(3)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	return m, players
}

func TestContractRummyCuiPresenter_Output(t *testing.T) {
	p := new(presenter.ContractRummyCuiPresenter)

	t.Run("draw phase", func(t *testing.T) {
		m, _ := setupContractRummyCuiMock(domain.ContractRummyPhaseDraw, false)
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("play phase", func(t *testing.T) {
		m, _ := setupContractRummyCuiMock(domain.ContractRummyPhasePlay, false)
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
		// Contract not yet met → the help emphasizes meeting it first.
		assert.Contains(t, out, "コントラクト達成が必須")
	})

	t.Run("round end", func(t *testing.T) {
		m, _ := setupContractRummyCuiMock(domain.ContractRummyPhaseRoundEnd, false)
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("game end", func(t *testing.T) {
		m, _ := setupContractRummyCuiMock(domain.ContractRummyPhaseGameEnd, true)
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupContractRummyCuiMock(domain.ContractRummyPhaseDraw, false)
		out := p.Output(m, errors.New("err"))
		assert.NotEmpty(t, out)
	})

	t.Run("contract met player", func(t *testing.T) {
		m, players := setupContractRummyCuiMock(domain.ContractRummyPhasePlay, false)
		players[0].SetContractMet(true)
		players[0].AppendMeld([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 5, false),
			domain.NewCard(domain.CardDesignHeart, 5, false),
			domain.NewCard(domain.CardDesignDiamond, 5, false),
		})
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
		// Contract met → optional extra melds / discard-to-end help, not the required prompt.
		assert.Contains(t, out, "コントラクト達成済み")
		assert.NotContains(t, out, "コントラクト達成が必須")
	})
}

// **`lo <cardIdx> <playerIdx> <meldIdx>` が要求する番号が場に出ていなかった** ──
// 札が並ぶだけなので、狙うメルドの添字は数えて当てるしかない (#6849)。
// Web の CLI モードは既に `M0:` と番号を出していて、無いのは CUI だけだった。
func TestContractRummyCuiPresenter_NumbersEachMeldForLayoff(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)

	p := new(presenter.ContractRummyCuiPresenter)

	// 黒スートだけで組む ── 赤スートは色コードで包まれ、期待値が装飾に依存する。
	// 中身で見分けが付く 2 つのメルドを使い、番号だけでなく **その番号のすぐ後ろに
	// そのメルドの札が並ぶこと**を見る。`mi + 1` にしたり順序を入れ替えたりする
	// 変異はこれで落ちる。
	melds := [][]*domain.Card{
		{
			domain.NewCard(domain.CardDesignSpade, 3, false),
			domain.NewCard(domain.CardDesignSpade, 4, false),
			domain.NewCard(domain.CardDesignSpade, 5, false),
		},
		{
			domain.NewCard(domain.CardDesignClover, 9, false),
			domain.NewCard(domain.CardDesignClover, 10, false),
			domain.NewCard(domain.CardDesignClover, 11, false),
		},
	}
	want := []string{
		"[0] SPADE 3 SPADE 4 SPADE 5",
		"[1] CLOVER 9 CLOVER 10 CLOVER 11",
	}

	withMelds := func(seat int) string {
		m, players := setupContractRummyCuiMock(domain.ContractRummyPhasePlay, false)
		players[seat].SetContractMet(true)
		for _, meld := range melds {
			players[seat].AppendMeld(meld)
		}
		return p.Output(m, nil)
	}

	t.Run("numbers the human melds from zero", func(t *testing.T) {
		out := withMelds(0)
		for mi, line := range want {
			assert.Contains(t, out, line,
				"meld %d must be labelled with the index `layoff` takes", mi)
		}
		// 0 始まりであること。`mi + 1` にすると最後の番号がここに現れる。
		assert.NotContains(t, out, "["+strconv.Itoa(len(melds))+"] ")
	})

	// layoff の相手は他家なので、番号付けが人間の場だけだと使えない。
	t.Run("numbers another player's melds the same way", func(t *testing.T) {
		out := withMelds(1)
		for _, line := range want {
			assert.Contains(t, out, line)
		}
	})
}

func TestContractRummyCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.ContractRummyCuiPresenter)
	m, _ := setupContractRummyCuiMock(domain.ContractRummyPhaseDraw, false)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	out := p.ActionLogOutput(m)
	assert.NotEmpty(t, out)
}

func TestContractRummyCuiPresenter_LayoffTargets(t *testing.T) {
	p := new(presenter.ContractRummyCuiPresenter)
	m, players := setupContractRummyCuiMock(domain.ContractRummyPhasePlay, false)
	players[0].SetContractMet(true)
	players[0].Reset()
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	players[1].SetContractMet(true)
	players[1].AppendMeld([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 5, false),
		domain.NewCard(domain.CardDesignClover, 5, false),
	})
	players[2].SetContractMet(true)
	players[2].AppendMeld([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 8, false),
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignSpade, 10, false),
	})

	out := p.Output(m, nil)
	assert.Contains(t, out, "レイオフ可能なメルド: 1/0 (プレイヤー/メルド)")
	assert.NotContains(t, out, "2/0")
}

func TestContractRummyCuiPresenter_OmitsLayoffTargetsWhenNoneAccept(t *testing.T) {
	p := new(presenter.ContractRummyCuiPresenter)
	m, players := setupContractRummyCuiMock(domain.ContractRummyPhasePlay, false)
	players[0].SetContractMet(true)
	players[0].Reset()
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	players[1].SetContractMet(true)
	players[1].AppendMeld([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 8, false),
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignSpade, 10, false),
	})

	out := p.Output(m, nil)
	assert.NotContains(t, out, "レイオフ可能なメルド:")
}

func TestContractRummyCuiPresenter_LayoffTargetsUseDomainRule(t *testing.T) {
	tests := []struct {
		name string
		card *domain.Card
		meld []*domain.Card
		want bool
	}{
		{
			name: "run extends at high end",
			card: domain.NewCard(domain.CardDesignSpade, 8, false),
			meld: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 5, false),
				domain.NewCard(domain.CardDesignSpade, 6, false),
				domain.NewCard(domain.CardDesignSpade, 7, false),
			},
			want: true,
		},
		{
			name: "run rejects a different suit",
			card: domain.NewCard(domain.CardDesignHeart, 8, false),
			meld: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 5, false),
				domain.NewCard(domain.CardDesignSpade, 6, false),
				domain.NewCard(domain.CardDesignSpade, 7, false),
			},
			want: false,
		},
		{
			name: "run accepts ace above king",
			card: domain.NewCard(domain.CardDesignSpade, 1, false),
			meld: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 11, false),
				domain.NewCard(domain.CardDesignSpade, 12, false),
				domain.NewCard(domain.CardDesignSpade, 13, false),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, players := setupContractRummyCuiMock(domain.ContractRummyPhasePlay, false)
			players[0].SetContractMet(true)
			players[0].Reset()
			players[0].AddCard(tt.card)
			players[1].SetContractMet(true)
			players[1].AppendMeld(tt.meld)

			out := (&presenter.ContractRummyCuiPresenter{}).Output(m, nil)
			assert.Equal(t, tt.want, strings.Contains(out, "1/0 (プレイヤー/メルド)"))
		})
	}
}
