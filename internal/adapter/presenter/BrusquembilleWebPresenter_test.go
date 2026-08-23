//go:build test && (!js || !wasm || classic)

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupBrusquembilleWebMock(trumpCard *domain.Card) *interfaces.MockBrusquembilleGame {
	m := new(interfaces.MockBrusquembilleGame)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.BrusquembillePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetTrumpCard").Return(trumpCard)
	m.On("GetDealerIdx").Return(0)
	m.On("GetLeadPlayerIdx").Return(1)
	m.On("GetStockRemaining").Return(25)
	// 前半 (山札あり) を既定に。合法手は手札全部。
	m.On("IsFollowRequired").Return(false).Maybe()
	m.On("GetValidPlayIndices", 0).Return([]int{0, 1, 2}).Maybe()
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultBrusquembilleConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupBrusquembilleWebMockWithPlayers(trumpCard *domain.Card) (*interfaces.MockBrusquembilleGame, []*domain.BrusquembillePlayer) {
	m := setupBrusquembilleWebMock(trumpCard)
	players := []*domain.BrusquembillePlayer{
		domain.NewBrusquembillePlayer(true),
		domain.NewBrusquembillePlayer(false),
	}
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayerPoints", 0).Return(15)
	m.On("GetPlayerPoints", 1).Return(5)
	return m, players
}

func TestBrusquembilleWebPresenter_Output_InitialState(t *testing.T) {
	p := new(presenter.BrusquembilleWebPresenter)
	trump := domain.NewCard(domain.CardDesignSpade, 13, false)
	m, players := setupBrusquembilleWebMockWithPlayers(trump)
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))

	got := p.Output(m, nil)
	assert.NotEmpty(t, got)

	var out controller.BrusquembilleWebOutput
	assert.NoError(t, json.Unmarshal([]byte(got), &out))
	assert.Equal(t, 2, len(out.Players))
	assert.False(t, out.GameEndFlag)
	assert.Equal(t, 0, out.Phase)
	assert.Equal(t, 1, out.TrickNumber)
	assert.Equal(t, domain.CardDesignSpade, out.TrumpSuit)
	assert.NotNil(t, out.TrumpCard)
	// 32 枚デッキ (クローン元のブリスコラは 40 枚)。
	assert.Equal(t, 25, out.StockRemaining)
	// **合法手と追従義務を必ず渡す。** 渡さないと、画面は押せるように
	// 見せておいて実行時にだけ拒否することになる。
	assert.Equal(t, []int{0, 1, 2}, out.ValidIndices)
	assert.False(t, out.FollowRequired, "山札が残っているうちは自由出し")
	assert.Equal(t, -1, out.WinnerIdx)
	assert.Equal(t, 15, out.Players[0].Points)
	assert.Equal(t, 5, out.Players[1].Points)
}

func TestBrusquembilleWebPresenter_Output_HumanCardsShownCPUHidden(t *testing.T) {
	p := new(presenter.BrusquembilleWebPresenter)
	trump := domain.NewCard(domain.CardDesignSpade, 13, false)
	m, players := setupBrusquembilleWebMockWithPlayers(trump)
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))

	got := p.Output(m, nil)
	var out controller.BrusquembilleWebOutput
	_ = json.Unmarshal([]byte(got), &out)

	human := out.Players[0]
	assert.True(t, human.IsHuman)
	assert.Equal(t, 1, human.CardCount)
	assert.Len(t, human.Cards, 1)

	cpu := out.Players[1]
	assert.False(t, cpu.IsHuman)
	assert.Equal(t, 1, cpu.CardCount)
	for _, c := range cpu.Cards {
		assert.Empty(t, c.Design, "CPU card design should be hidden")
	}
}

func TestBrusquembilleWebPresenter_Output_GameEnd(t *testing.T) {
	p := new(presenter.BrusquembilleWebPresenter)
	m := setupBrusquembilleWebMock(nil)
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(domain.NewBrusquembillePlayer(true))
	m.On("GetPlayer", 1).Return(domain.NewBrusquembillePlayer(false))
	m.On("GetPlayerPoints", 0).Return(70)
	m.On("GetPlayerPoints", 1).Return(50)
	// Override: game ended, p0 wins
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
	m.On("GetGameEndFlag").Return(true)
	m.On("GetWinnerIdx").Return(0)

	got := p.Output(m, nil)
	var out controller.BrusquembilleWebOutput
	_ = json.Unmarshal([]byte(got), &out)
	assert.True(t, out.GameEndFlag)
	assert.Equal(t, 0, out.WinnerIdx)
	assert.Equal(t, "brusquembille.result.p0Win", out.MessageCode)
}

func TestBrusquembilleWebPresenter_Output_GameEndTie(t *testing.T) {
	p := new(presenter.BrusquembilleWebPresenter)
	m := setupBrusquembilleWebMock(nil)
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(domain.NewBrusquembillePlayer(true))
	m.On("GetPlayer", 1).Return(domain.NewBrusquembillePlayer(false))
	m.On("GetPlayerPoints", 0).Return(60)
	m.On("GetPlayerPoints", 1).Return(60)
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
	m.On("GetGameEndFlag").Return(true)
	m.On("GetWinnerIdx").Return(-1)

	got := p.Output(m, nil)
	var out controller.BrusquembilleWebOutput
	_ = json.Unmarshal([]byte(got), &out)
	assert.Equal(t, "brusquembille.result.tie", out.MessageCode)
}

func TestBrusquembilleWebPresenter_Output_Error(t *testing.T) {
	p := new(presenter.BrusquembilleWebPresenter)
	m, _ := setupBrusquembilleWebMockWithPlayers(nil)
	got := p.Output(m, errors.New("boom"))
	var out controller.BrusquembilleWebOutput
	_ = json.Unmarshal([]byte(got), &out)
	assert.Equal(t, "boom", out.Message)
}

func TestBrusquembilleWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.BrusquembilleWebPresenter)
	trump := domain.NewCard(domain.CardDesignSpade, 13, false)
	m, players := setupBrusquembilleWebMockWithPlayers(trump)
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
	idx := 0
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
	m.On("GetHint").Return(&domain.BrusquembilleHint{CardIndex: &idx, Reason: "lead_low"})

	got := p.HintOutput(m)
	var out controller.BrusquembilleWebOutput
	_ = json.Unmarshal([]byte(got), &out)
	assert.NotNil(t, out.Hint)
	assert.Equal(t, 0, *out.Hint.CardIndex)
	assert.Equal(t, "lead_low", out.Hint.Reason)
}

func TestBrusquembilleWebPresenter_HintOutput_None(t *testing.T) {
	p := new(presenter.BrusquembilleWebPresenter)
	m, _ := setupBrusquembilleWebMockWithPlayers(nil)
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
	m.On("GetHint").Return((*domain.BrusquembilleHint)(nil))
	got := p.HintOutput(m)
	var out controller.BrusquembilleWebOutput
	_ = json.Unmarshal([]byte(got), &out)
	assert.Nil(t, out.Hint)
}

func TestBrusquembilleWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.BrusquembilleWebPresenter)
	m, _ := setupBrusquembilleWebMockWithPlayers(nil)
	got := p.ActionLogOutput(m)
	assert.NotEmpty(t, got)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
//
// トリックテイキング系はソリティア系と違い、**Output 側のゲートが要りません。**
// Brusquembille.GetHint() が「プレイ中かつ人間の手番」を自分で確かめて nil を返すので、
// フェーズ判定を持ち込むとドメインの判定を二重に書くことになります。
func TestBrusquembilleWebPresenterOutputCarriesTheHint(t *testing.T) {
	trump := domain.NewCard(domain.CardDesignSpade, 13, false)
	brg, players := setupBrusquembilleWebMockWithPlayers(trump)
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
	idx := 0
	brg.ExpectedCalls = removeMockCall(brg.ExpectedCalls, "GetHint")
	brg.On("GetHint").Return(&domain.BrusquembilleHint{CardIndex: &idx, Reason: "lead_low"})

	result := new(presenter.BrusquembilleWebPresenter).Output(brg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
// ページは `isRequestedHint` でこのコードを見てからバナーを出すので (#4605)、
// 付いていないとヒントを押しても画面に何も出ない。
func TestBrusquembilleWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	g := domain.NewDefaultBrusquembille()
	g.Reset()
	// **Reset 直後は人間の手番とは限らない。**GetHint は手番でなければ nil を
	// 返すので、席を人間に固定しないとこのテストは前提で落ちる。
	g.SetCurrentPlayerIdx(0)
	require.NotNil(t, g.GetHint(), "fixture must actually produce a hint")
	assert.Contains(t, new(presenter.BrusquembilleWebPresenter).HintOutput(g), "brusquembille.hintRequested")
}
