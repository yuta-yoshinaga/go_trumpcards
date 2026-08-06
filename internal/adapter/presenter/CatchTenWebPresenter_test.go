package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupCatchTenWebMock() *interfaces.MockCatchTenGame {
	m := new(interfaces.MockCatchTenGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.CatchTenPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTeamScore", 0).Return(0)
	m.On("GetTeamScore", 1).Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultCatchTenConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()
	// validPlayIndices を出すようになった。既定は「人間の手番でない」= 制限なし。
	// 合法手を確かめるテストは自分で上書きする。
	m.On("IsHumanTurn").Return(false).Maybe()
	m.On("GetValidPlayIndices", 0).Return([]int(nil)).Maybe()

	return m
}

func makeCatchTenPlayers() []*domain.CatchTenPlayer {
	return []*domain.CatchTenPlayer{
		domain.NewCatchTenPlayer(true, 0),
		domain.NewCatchTenPlayer(false, 1),
		domain.NewCatchTenPlayer(false, 0),
		domain.NewCatchTenPlayer(false, 1),
	}
}

func setupCatchTenWebMockWithPlayers() (*interfaces.MockCatchTenGame, []*domain.CatchTenPlayer) {
	m := setupCatchTenWebMock()
	players := makeCatchTenPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestCatchTenWebPresenter_Output(t *testing.T) {
	p := new(presenter.CatchTenWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupCatchTenWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(m, nil)
		var resObj controller.CatchTenWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, 4, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, 0, resObj.Phase)
		assert.Equal(t, domain.CardDesignSpade, resObj.TrumpSuit)
		assert.Equal(t, -1, resObj.WinnerTeam)
		assert.Equal(t, "catchten.playPhase.lead", resObj.MessageCode)
	})

	t.Run("human cards shown, CPU hidden", func(t *testing.T) {
		m, players := setupCatchTenWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(m, nil)
		var resObj controller.CatchTenWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.Equal(t, 0, resObj.Players[0].Team)
		assert.Equal(t, 1, resObj.Players[1].Team)
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		result := p.Output(m, errors.New("test error"))
		var resObj controller.CatchTenWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "test error", resObj.Message)
	})

	t.Run("game end humanWin (team 0)", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetWinnerTeam").Return(0)

		result := p.Output(m, nil)
		var resObj controller.CatchTenWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "catchten.result.humanWin", resObj.MessageCode)
	})

	t.Run("game end cpuWin (team 1)", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetWinnerTeam").Return(1)

		result := p.Output(m, nil)
		var resObj controller.CatchTenWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "catchten.result.cpuWin", resObj.MessageCode)
	})

	t.Run("game end draw", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetWinnerTeam").Return(domain.CatchTenDrawTeam)

		result := p.Output(m, nil)
		var resObj controller.CatchTenWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "catchten.result.draw", resObj.MessageCode)
	})

	t.Run("play phase follow message", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 6, false)},
		})
		result := p.Output(m, nil)
		var resObj controller.CatchTenWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "catchten.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end message", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CatchTenPhaseTrickEnd)
		result := p.Output(m, nil)
		var resObj controller.CatchTenWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "catchten.trickEnd", resObj.MessageCode)
	})

	t.Run("round end message", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CatchTenPhaseRoundEnd)
		result := p.Output(m, nil)
		var resObj controller.CatchTenWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "catchten.roundEnd", resObj.MessageCode)
	})
}

// ドメインは合法手を判定済み (GetValidPlayIndices は「Web用」と明記されている)
// なのに Web に送っていなかった。違反札をクリックしてサーバーのエラーが返って
// 初めて出せないと分かる状態だった。
func TestCatchTenWebPresenter_ValidPlayIndices(t *testing.T) {
	p := new(presenter.CatchTenWebPresenter)

	decode := func(t *testing.T, m *interfaces.MockCatchTenGame) controller.CatchTenWebOutput {
		t.Helper()
		var out controller.CatchTenWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &out))
		return out
	}

	t.Run("human play turn carries the domain's answer", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsHumanTurn")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetValidPlayIndices")
		m.On("IsHumanTurn").Return(true)
		m.On("GetValidPlayIndices", 0).Return([]int{2})

		assert.Equal(t, []int{2}, decode(t, m).ValidPlayIndices)
	})

	// 合法手が1枚も無い局面も空で返る。呼び出し側はこれを「制限なし」と読まない。
	t.Run("no legal card still reports empty", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsHumanTurn")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetValidPlayIndices")
		m.On("IsHumanTurn").Return(true)
		m.On("GetValidPlayIndices", 0).Return([]int{})

		assert.Empty(t, decode(t, m).ValidPlayIndices)
	})

	// プレイフェーズ以外も踏む。トリック終了中は制限が決まっていないので、
	// ここで合法手を送ると「いま出せる」と誤って伝えることになる。
	t.Run("trick end reports nothing", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsHumanTurn")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetValidPlayIndices")
		m.On("GetPhase").Return(domain.CatchTenPhaseTrickEnd)
		m.On("IsHumanTurn").Return(true)
		m.On("GetValidPlayIndices", 0).Return([]int{2})

		assert.Empty(t, decode(t, m).ValidPlayIndices)
	})

	// nil を返す実装でも JSON が null にならないこと。GetValidPlayIndices は
	// インターフェース越しの呼び出しなので、具象型の性質には依存できない。
	t.Run("nil from the domain serialises as an empty array", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsHumanTurn")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetValidPlayIndices")
		m.On("IsHumanTurn").Return(true)
		m.On("GetValidPlayIndices", 0).Return([]int(nil))

		raw := p.Output(m, nil)
		assert.Contains(t, raw, `"validPlayIndices":[]`, "null ではなく空配列で出す")
		assert.Empty(t, decode(t, m).ValidPlayIndices)
	})

	t.Run("cpu turn reports nothing", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetValidPlayIndices")
		m.On("GetValidPlayIndices", 0).Return([]int{2})

		assert.Empty(t, decode(t, m).ValidPlayIndices)
	})
}

func TestCatchTenWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.CatchTenWebPresenter)

	t.Run("with hint", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		cardIdx := 3
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.CatchTenHint{CardIndex: &cardIdx, Reason: "follow_suit"})
		result := p.HintOutput(m)
		var resObj controller.CatchTenWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, 3, *resObj.Hint.CardIndex)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.CatchTenHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.CatchTenWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestCatchTenWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.CatchTenWebPresenter)
	m := setupCatchTenWebMock()
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "test"},
	})
	assert.NotEmpty(t, p.ActionLogOutput(m))
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestCatchTenWebPresenterOutputCarriesTheHint(t *testing.T) {
	idx := 0
	ctg, _ := setupCatchTenWebMockWithPlayers()
	ctg.ExpectedCalls = removeMockCall(ctg.ExpectedCalls, "GetHint")
	ctg.On("GetHint").Return(&domain.CatchTenHint{CardIndex: &idx, Reason: "follow_suit"})

	result := new(presenter.CatchTenWebPresenter).Output(ctg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	assert.NotContains(t, result, "catchten.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestCatchTenWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	idx := 0
	ctg, _ := setupCatchTenWebMockWithPlayers()
	ctg.ExpectedCalls = removeMockCall(ctg.ExpectedCalls, "GetHint")
	ctg.On("GetHint").Return(&domain.CatchTenHint{CardIndex: &idx, Reason: "follow_suit"})
	assert.Contains(t, new(presenter.CatchTenWebPresenter).HintOutput(ctg), "catchten.hintRequested")

	none, _ := setupCatchTenWebMockWithPlayers()
	none.ExpectedCalls = removeMockCall(none.ExpectedCalls, "GetHint")
	none.On("GetHint").Return((*domain.CatchTenHint)(nil))
	assert.Contains(t, new(presenter.CatchTenWebPresenter).HintOutput(none), "catchten.noHint")
}
