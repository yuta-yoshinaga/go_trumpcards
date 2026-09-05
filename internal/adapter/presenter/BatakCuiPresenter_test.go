package presenter_test

import (
	"errors"
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

func setupBatakCuiMock() *interfaces.MockBatakGame {
	m := new(interfaces.MockBatakGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetSpadesBroken").Return(false)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.BatakPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetDeclarerIdx").Return(-1)
	m.On("GetHighBid").Return(0)
	m.On("MinLegalBid").Return(domain.BatakMinBid)
	m.On("GetConfig").Return(domain.DefaultBatakConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// 合法手の目印 (#5605)。既定は「制限なし」= 目印を出さない状態にしておき、
	// 目印そのものを見るテストは本物のドメインで確かめる。
	m.On("GetValidPlayIndices", mock.Anything).Return([]int(nil))
	return m
}

func makeBatakPlayers() []*domain.BatakPlayer {
	return []*domain.BatakPlayer{
		domain.NewBatakPlayer(true),
		domain.NewBatakPlayer(false),
		domain.NewBatakPlayer(false),
		domain.NewBatakPlayer(false),
	}
}

func setupBatakCuiMockWithPlayers() (*interfaces.MockBatakGame, []*domain.BatakPlayer) {
	m := setupBatakCuiMock()
	players := makeBatakPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestBatakCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.BatakCuiPresenter)

	t.Run("initial header and player info", func(t *testing.T) {
		m, players := setupBatakCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Batak (バタック)")
		assert.Contains(t, result, "ラウンド: 1 / 5")
		assert.Contains(t, result, "トリック: 1")
		assert.Contains(t, result, "親: 未確定 (最高ビッド: -)")
		assert.Contains(t, result, "得点規則: 親は達成で +bid・未達で -bid、子は獲得トリック数が加点")
		assert.Contains(t, result, "スペードブレイク: なし")
		assert.Contains(t, result, "あなた: ビッド=未ビッド 獲得0トリック 累積0点 ラウンド0点 2枚")
		assert.Contains(t, result, "手番: あなた")
		assert.Contains(t, result, "play <idx>")
	})

	t.Run("raw integer score", func(t *testing.T) {
		m, players := setupBatakCuiMockWithPlayers()
		players[1].SetCumulativeScore(5)
		players[1].SetRoundScore(5)
		players[1].SetBid(5)
		players[1].AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})

		result := p.Output(m, nil)
		assert.Contains(t, result, "CPU 1: ビッド=5 獲得1トリック 累積5点 ラウンド5点")
	})

	t.Run("negative integer score", func(t *testing.T) {
		m, players := setupBatakCuiMockWithPlayers()
		players[1].SetCumulativeScore(-5)
		players[1].SetRoundScore(-5)
		players[1].SetBid(5)

		result := p.Output(m, nil)
		assert.Contains(t, result, "CPU 1: ビッド=5 獲得0トリック 累積-5点 ラウンド-5点")
	})

	t.Run("spades broken shows yes", func(t *testing.T) {
		m, _ := setupBatakCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetSpadesBroken")
		m.On("GetSpadesBroken").Return(true)
		assert.Contains(t, p.Output(m, nil), "スペードブレイク: あり")
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupBatakCuiMockWithPlayers()
		err := errors.New("invalid play")
		assert.Contains(t, p.Output(m, err), "invalid play")
	})

	t.Run("game ended human win", func(t *testing.T) {
		m, _ := setupBatakCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "あなたの勝利です！")
	})

	t.Run("bid phase with legal bid min", func(t *testing.T) {
		m, _ := setupBatakCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BatakPhaseBid)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ビッドフェーズ: あなたの番")
		assert.Contains(t, result, "b <n> (5-13) または pass (0)・・・ビッドを宣言")
	})

	t.Run("bid phase pass only when min legal bid is 0", func(t *testing.T) {
		m, _ := setupBatakCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "MinLegalBid")
		m.On("GetPhase").Return(domain.BatakPhaseBid)
		m.On("MinLegalBid").Return(domain.BatakPassBid)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ビッドフェーズ: あなたの番")
		assert.Contains(t, result, "パスのみ可能です (pass または b 0)")
	})

	t.Run("trick end and round end", func(t *testing.T) {
		m, _ := setupBatakCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BatakPhaseTrickEnd)
		assert.Contains(t, p.Output(m, nil), "トリック終了")
	})
}

func TestBatakCuiPresenter_ShowsDeclarerAndPass(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.BatakCuiPresenter)

	m, players := setupBatakCuiMockWithPlayers()
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDeclarerIdx")
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHighBid")
	m.On("GetDeclarerIdx").Return(1)
	m.On("GetHighBid").Return(6)

	players[1].SetBid(6)
	players[2].SetBid(domain.BatakPassBid)

	result := p.Output(m, nil)
	assert.Contains(t, result, "親: CPU 1 (最高ビッド: 6)")
	assert.Contains(t, result, "CPU 1 [親]: ビッド=6")
	assert.Contains(t, result, "CPU 2: ビッド=パス")
}

func TestBatakCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.BatakCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockBatakGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "played SPADE 5"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)
		// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
		m.On("GetPlayer", mock.Anything).Return(domain.NewBatakPlayer(true)).Maybe()
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "play")
		assert.Contains(t, result, "あなた", "棋譜の座席名が他の行と揃っていない")
	})

	t.Run("game not ended yields placeholder", func(t *testing.T) {
		m := new(interfaces.MockBatakGame)
		m.On("GetGameEndFlag").Return(false)
		assert.Contains(t, p.ActionLogOutput(m), "棋譜はありません")
	})
}

func TestBatakCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("no hint", func(t *testing.T) {
		m := new(interfaces.MockBatakGame)
		m.On("GetHint").Return((*domain.BatakHint)(nil))
		assert.Contains(t, (&presenter.BatakCuiPresenter{}).HintOutput(m), "ヒントはありません")
	})

	t.Run("bid hint", func(t *testing.T) {
		bid := 3
		m := new(interfaces.MockBatakGame)
		m.On("GetHint").Return(&domain.BatakHint{Bid: &bid, Reason: "strategic_bid"})
		result := (&presenter.BatakCuiPresenter{}).HintOutput(m)
		assert.Contains(t, result, "ビッド 3")
		assert.Contains(t, result, "戦略的なビッド")
	})

	t.Run("pass bid hint ja and en", func(t *testing.T) {
		origLang := i18n.Lang()
		defer i18n.SetLang(origLang)

		passBid := domain.BatakPassBid
		m := new(interfaces.MockBatakGame)
		m.On("GetHint").Return(&domain.BatakHint{Bid: &passBid, Reason: "pass_weak_hand"})

		// Japanese
		i18n.SetLang("ja")
		jaOut := (&presenter.BatakCuiPresenter{}).HintOutput(m)
		assert.Contains(t, jaOut, "[HINT: パスを推奨 (手が弱い)]")
		assert.NotContains(t, jaOut, "pass_weak_hand")
		assert.NotContains(t, jaOut, "ビッド 0")

		// English
		i18n.SetLang("en")
		enOut := (&presenter.BatakCuiPresenter{}).HintOutput(m)
		assert.Contains(t, enOut, "[HINT: pass (weak hand)]")
		assert.NotContains(t, enOut, "pass_weak_hand")
		assert.NotContains(t, enOut, "bid 0")
	})

	t.Run("nil bid and nil card index", func(t *testing.T) {
		m := new(interfaces.MockBatakGame)
		m.On("GetHint").Return(&domain.BatakHint{Reason: "unknown"})
		assert.Contains(t, (&presenter.BatakCuiPresenter{}).HintOutput(m), "ヒントはありません")
	})

	t.Run("play hint with trump_cut reason", func(t *testing.T) {
		idx := 0
		m := new(interfaces.MockBatakGame)
		m.On("GetHint").Return(&domain.BatakHint{CardIndex: &idx, Reason: "trump_cut"})
		player := domain.NewBatakPlayer(true)
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		m.On("GetPlayer", 0).Return(player)

		result := (&presenter.BatakCuiPresenter{}).HintOutput(m)
		assert.Contains(t, result, "スペードでカット")
	})

	t.Run("play hint shared reason fallback", func(t *testing.T) {
		idx := 0
		m := new(interfaces.MockBatakGame)
		m.On("GetHint").Return(&domain.BatakHint{CardIndex: &idx, Reason: "follow_suit"})
		player := domain.NewBatakPlayer(true)
		player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		m.On("GetPlayer", 0).Return(player)
		result := (&presenter.BatakCuiPresenter{}).HintOutput(m)
		assert.Contains(t, result, "リードスートに追随")
	})
}

// #5605: Web は validPlayIndices で出せない札を無効化しツールチップまで出すのに、
// CUI は番号付きの一覧を並べるだけだった。マストフォロー/マストトランプで
// 何が出せるかは、番号を打ってエラーを踏むまで分からない。
func TestBatakCuiPresenterMarksThePlayableCards(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.BatakCuiPresenter)

	// 人間が持っているスートをリードさせると、そのスートだけが合法になる。
	setup := func(t *testing.T) *domain.Batak {
		t.Helper()
		cb := domain.NewDefaultBatak()
		cb.Reset()
		cb.SetPhase(domain.BatakPhasePlay)
		cb.SetCurrentPlayerIdx(0)
		human := cb.GetPlayer(0)
		if human.GetCardsSize() == 0 {
			t.Fatal("前提: 人間に手札が配られていること")
		}
		leadSuit := human.GetCard(0).GetDesign()
		cb.SetCurrentTrick([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(leadSuit, 5, false)},
		})
		return cb
	}

	t.Run("human play turn marks only the legal cards", func(t *testing.T) {
		cb := setup(t)
		playable := cb.GetValidPlayIndices(0)
		// **配りに賭けてはいない。**リードスートは人間の1枚目のスートなので合法手は
		// 必ず1枚以上ある。「全部合法」になるのは13枚が同一スートの配りのときだけ。
		if len(playable) == 0 || len(playable) == cb.GetPlayer(0).GetCardsSize() {
			t.Skipf("13枚同一スートの配り (%d/%d) -- 目印の有無を区別できない",
				len(playable), cb.GetPlayer(0).GetCardsSize())
		}

		out := p.Output(cb, nil)
		assert.Equal(t, len(playable), strings.Count(out, presenter.CuiLegalMark),
			"目印の数が合法手の数と一致する")
	})

	// **目印を出さない側も踏む。**ビッド中は制限そのものが決まっていない。
	t.Run("bid phase leaves the hand unmarked", func(t *testing.T) {
		cb := setup(t)
		cb.SetPhase(domain.BatakPhaseBid)

		out := p.Output(cb, nil)
		assert.NotContains(t, out, presenter.CuiLegalMark, "ビッド中は目印を出さない")
	})

	// CPU の手番でも出さない。人間の手札に「今出せる札」の印が付くのは自分の番だけ。
	t.Run("another player's turn leaves the hand unmarked", func(t *testing.T) {
		cb := setup(t)
		cb.SetCurrentPlayerIdx(1)

		out := p.Output(cb, nil)
		assert.NotContains(t, out, presenter.CuiLegalMark, "他家の手番では目印を出さない")
	})
}
