package presenter_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestFollowTheQueenCuiPresenter_Output_BettingLimitDisplay(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.FollowTheQueenCuiPresenter)

	t.Run("displays Fixed limit", func(t *testing.T) {
		s, _ := makeFollowTheQueenForPresenter()
		s.SetPhase(domain.FollowTheQueenPhaseThirdStreet)
		result := p.Output(s, nil)
		assert.Contains(t, result, "Fixed")
	})
}

func TestFollowTheQueenCuiPresenter_Output_TableSize(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.FollowTheQueenCuiPresenter)

	t.Run("displays 4-max", func(t *testing.T) {
		s, _ := makeFollowTheQueenForPresenter()
		s.SetPhase(domain.FollowTheQueenPhaseThirdStreet)
		result := p.Output(s, nil)
		assert.Contains(t, result, "テーブル: 4-max")
	})
}

func TestFollowTheQueenCuiPresenter_Output_RebuyAddon(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.FollowTheQueenCuiPresenter)

	t.Run("tournament mode with rebuy enabled shows rebuy info", func(t *testing.T) {
		s, _ := makeFollowTheQueenForPresenter()
		s.SetPhase(domain.FollowTheQueenPhaseThirdStreet)
		cfg := domain.FollowTheQueenConfig{
			Ante:             5,
			BringIn:          10,
			SmallBet:         10,
			BigBet:           20,
			InitChips:        1000,
			TournamentMode:   true,
			AnteLevelHands:   5,
			AnteMultiplier:   200,
			RebuyEnabled:     true,
			RebuyChips:       1000,
			RebuyMaxCount:    3,
			RebuyPeriodHands: 20,
		}
		s.SetConfig(cfg)
		s.SetHandCount(2)

		result := p.Output(s, nil)
		assert.Contains(t, result, "リバイ: 1000チップ (最大3回, 20ハンド目まで)")
	})

	t.Run("tournament mode with addon enabled shows addon info", func(t *testing.T) {
		s, _ := makeFollowTheQueenForPresenter()
		s.SetPhase(domain.FollowTheQueenPhaseThirdStreet)
		cfg := domain.FollowTheQueenConfig{
			Ante:           5,
			BringIn:        10,
			SmallBet:       10,
			BigBet:         20,
			InitChips:      1000,
			TournamentMode: true,
			AnteLevelHands: 5,
			AnteMultiplier: 200,
			AddonEnabled:   true,
			AddonChips:     1500,
			AddonAfterHand: 20,
		}
		s.SetConfig(cfg)
		s.SetHandCount(2)

		result := p.Output(s, nil)
		assert.Contains(t, result, "アドオン: 1500チップ (20ハンド目に提供)")
	})

	t.Run("rebuy phase type 1 shows rebuy prompt", func(t *testing.T) {
		s, _ := makeFollowTheQueenForPresenter()
		cfg := domain.FollowTheQueenConfig{
			Ante:             5,
			BringIn:          10,
			SmallBet:         10,
			BigBet:           20,
			InitChips:        1000,
			TournamentMode:   true,
			AnteLevelHands:   5,
			AnteMultiplier:   200,
			RebuyEnabled:     true,
			RebuyChips:       1000,
			RebuyMaxCount:    3,
			RebuyPeriodHands: 20,
		}
		s.SetConfig(cfg)
		s.SetPhase(domain.FollowTheQueenPhaseRebuy)
		s.SetRebuyPhaseType(1)
		s.SetRebuyCounts([]int{1, 0, 0, 0})

		result := p.Output(s, nil)
		assert.Contains(t, result, "リバイしますか?")
		assert.Contains(t, result, "1000チップ")
		assert.Contains(t, result, "1/3回使用済")
	})

	t.Run("rebuy phase type 2 shows addon prompt", func(t *testing.T) {
		s, _ := makeFollowTheQueenForPresenter()
		cfg := domain.FollowTheQueenConfig{
			Ante:           5,
			BringIn:        10,
			SmallBet:       10,
			BigBet:         20,
			InitChips:      1000,
			TournamentMode: true,
			AnteLevelHands: 5,
			AnteMultiplier: 200,
			AddonEnabled:   true,
			AddonChips:     1500,
			AddonAfterHand: 20,
		}
		s.SetConfig(cfg)
		s.SetPhase(domain.FollowTheQueenPhaseRebuy)
		s.SetRebuyPhaseType(2)

		result := p.Output(s, nil)
		assert.Contains(t, result, "アドオンしますか?")
		assert.Contains(t, result, "1500チップ")
	})
}

func TestFollowTheQueenCuiPresenter_Output_Muck(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.FollowTheQueenCuiPresenter)

	t.Run("muck prompt displayed when available", func(t *testing.T) {
		s, _ := makeFollowTheQueenForPresenter()
		s.SetPhase(domain.FollowTheQueenPhaseShowdown)
		s.SetRoundResults([]domain.FollowTheQueenResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandOnePair, HandName: "One Pair", WonAmount: 0},
			{PlayerIdx: 1, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100},
		})

		result := p.Output(s, nil)
		assert.Contains(t, result, "マックしますか? (m=マック / sh=ショー)")
	})
}

func TestFollowTheQueenCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.FollowTheQueenCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		mockGame := new(interfaces.MockFollowTheQueenGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "raise", Detail: "raised to 100"},
		}
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(entries)
		// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
		mockGame.On("GetPlayer", mock.Anything).Return(domain.NewFollowTheQueenPlayer(true, domain.FollowTheQueenPlayStyle(0))).Maybe()

		result := p.ActionLogOutput(mockGame)
		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "raise")
		mockGame.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		mockGame := new(interfaces.MockFollowTheQueenGame)
		mockGame.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(mockGame)
		assert.Contains(t, result, "棋譜はありません")
		mockGame.AssertExpectations(t)
	})
}

// #5542: Web は 3rd street でブリングイン (強制ベットを払い最初に動く席) に
// バッジを出すのに、CUI は誰なのかを知る手段が無かった。
func TestFollowTheQueenCuiPresenter_Output_BringIn(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.FollowTheQueenCuiPresenter)

	outputWith := func(phase, bringIn int) string {
		s, players := makeFollowTheQueenForPresenter()
		s.SetPhase(phase)
		s.SetBringInPlayerIdx(bringIn)
		players[0].AddDoorCard(domain.NewCard(domain.CardDesignClover, 5, false))
		return p.Output(s, nil)
	}

	line := func(idx int) string {
		return i18n.Tf("followthequeen.bringInLine", "name", i18n.Tf("cuiPlayerCpu", "idx", strconv.Itoa(idx)))
	}

	out := outputWith(domain.FollowTheQueenPhaseThirdStreet, 2)
	assert.Contains(t, out, line(2))

	// **他ストリートでは出さない。**強制ベットは 3rd street だけの話。
	header := strings.SplitN(i18n.T("followthequeen.bringInLine"), "{{", 2)[0]
	assert.NotContains(t, outputWith(domain.FollowTheQueenPhaseFourthStreet, 2), header)
	// 未確定 (-1) のときも出さない。
	assert.NotContains(t, outputWith(domain.FollowTheQueenPhaseThirdStreet, -1), header)
}

// **いま何がワイルドかを画面に出しているか。** これがフォロー・ザ・クイーンの
// ゲームそのもので、出ていなければプレイヤーは自分の役すら数えられない。
// 「まだ出ていない」も明示させる —— 行ごと消えると「無いのか出し忘れなのか」が
// 区別できず、負のコントロールとしても効かない。
func TestFollowTheQueenCuiPresenter_Output_WildRankLine(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.FollowTheQueenCuiPresenter)

	outputWithWild := func(rank int) string {
		s, _ := makeFollowTheQueenForPresenter()
		s.SetPhase(domain.FollowTheQueenPhaseFourthStreet)
		s.SetWildRankForTest(rank)
		return p.Output(s, nil)
	}

	t.Run("names the wild rank once a queen has turned one up", func(t *testing.T) {
		out := outputWithWild(7)
		assert.Contains(t, out, i18n.Tf("followthequeen.wildLine", "rank", "7"))
		assert.NotContains(t, out, i18n.T("followthequeen.wildNone"))
	})

	t.Run("renders face ranks as letters rather than numbers", func(t *testing.T) {
		out := outputWithWild(13)
		assert.Contains(t, out, i18n.Tf("followthequeen.wildLine", "rank", "K"))
		// 「13」がそのまま出ていないこと。cuiRankName を strconv.Itoa に
		// 差し替えるとここで落ちる。
		assert.NotContains(t, out, i18n.Tf("followthequeen.wildLine", "rank", "13"))
	})

	t.Run("says none while no queen has shown", func(t *testing.T) {
		out := outputWithWild(0)
		assert.Contains(t, out, i18n.T("followthequeen.wildNone"))
	})
}

// **ヒントがワイルドを見ているか。** ワイルドを 1 枚持っていれば実質ペア以上、
// 2 枚ならスリーカード以上が確定する。素のスタッド用の助言をそのまま出すと、
// 同じ手を「ワンペア」あるいは「降りろ」と評価してしまう。
func TestFollowTheQueenCuiPresenter_Hint_CountsWilds(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.FollowTheQueenCuiPresenter)

	// バラバラの 3 枚。ワイルド無しなら「降りろ」になる手を土台にする。
	hintFor := func(wildRank int, cards ...*domain.Card) string {
		s, players := makeFollowTheQueenForPresenter()
		s.SetPhase(domain.FollowTheQueenPhaseThirdStreet)
		s.SetCurrentTurn(0)
		s.SetWildRankForTest(wildRank)
		for i, c := range cards {
			if i == len(cards)-1 {
				players[0].AddDoorCard(c)
				continue
			}
			players[0].AddHoleCard(c)
		}
		return p.HintOutput(s)
	}

	junk := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, false),
		domain.NewCard(domain.CardDesignHeart, 6, false),
		domain.NewCard(domain.CardDesignClover, 9, false),
	}

	t.Run("folds the same three cards when nothing is wild", func(t *testing.T) {
		out := hintFor(0, junk...)
		assert.Contains(t, out, i18n.T("followthequeen.hintReasonFold"))
	})

	t.Run("plays on one wild", func(t *testing.T) {
		// 9 をワイルドにするだけで、同じ 3 枚が「実質ペア以上」に変わる。
		out := hintFor(9, junk...)
		assert.Contains(t, out, i18n.T("followthequeen.hintReasonWildOne"))
		assert.Contains(t, out, i18n.T("followthequeen.hintContinue"))
	})

	t.Run("counts a queen as wild without any wild rank set", func(t *testing.T) {
		// wildRank=0 でもクイーンは常にワイルド。presenter が自前で
		// ランク判定を書き直すと、この subtest が落ちる。
		out := hintFor(0,
			domain.NewCard(domain.CardDesignSpade, 3, false),
			domain.NewCard(domain.CardDesignHeart, 6, false),
			domain.NewCard(domain.CardDesignClover, domain.FollowTheQueenQueenValue, false))
		assert.Contains(t, out, i18n.T("followthequeen.hintReasonWildOne"))
	})

	t.Run("says trips-or-better on two wilds", func(t *testing.T) {
		out := hintFor(9,
			domain.NewCard(domain.CardDesignSpade, 9, false),
			domain.NewCard(domain.CardDesignHeart, 6, false),
			domain.NewCard(domain.CardDesignClover, 9, false))
		assert.Contains(t, out, i18n.T("followthequeen.hintReasonWildTwo"))
	})

	t.Run("prefers the wild reading over the pair reading", func(t *testing.T) {
		// 9-9 は素のスタッドならワンペア。ワイルドが 9 なら実際はスリーカード
		// 以上なので、ペア分岐に先に落ちてはいけない。
		out := hintFor(9,
			domain.NewCard(domain.CardDesignSpade, 9, false),
			domain.NewCard(domain.CardDesignHeart, 9, false),
			domain.NewCard(domain.CardDesignClover, 4, false))
		assert.NotContains(t, out, i18n.T("followthequeen.hintReasonPair"))
	})
}
