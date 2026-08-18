package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func makeGongZhuPlayers() []*domain.GongZhuPlayer {
	return []*domain.GongZhuPlayer{
		domain.NewGongZhuPlayer(true),
		domain.NewGongZhuPlayer(false),
		domain.NewGongZhuPlayer(false),
		domain.NewGongZhuPlayer(false),
	}
}

func setupGongZhuWebMock() *interfaces.MockGongZhuGame {
	m := new(interfaces.MockGongZhuGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetHeartsBroken").Return(false)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.GongZhuPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("IsHumanTurn").Return(true).Maybe()
	m.On("GetPlayableIndices", mock.Anything).Return(([]int)(nil)).Maybe()
	m.On("GetExposure").Return(domain.GongZhuExposure{})
	m.On("GetExposableIndices", 0).Return([]int{})
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultGongZhuConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// 得点内訳の既定値 (#5630)。中身を見るテストは removeMockCall で上書きする。
	m.On("ScoreBreakdownFor", mock.Anything).Return(domain.GongZhuScoreBreakdown{}).Maybe()
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupGongZhuWebMockWithPlayers() (*interfaces.MockGongZhuGame, []*domain.GongZhuPlayer) {
	m := setupGongZhuWebMock()
	players := makeGongZhuPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestGongZhuWebPresenter_Output(t *testing.T) {
	p := new(presenter.GongZhuWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupGongZhuWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.GongZhuWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, 4)
		assert.Equal(t, 1, resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerIdx)
		assert.Equal(t, "gongzhu.playPhase.lead", resObj.MessageCode)
		// human cards visible, cpu hidden
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
	})

	t.Run("captured point cards populated and filtered", func(t *testing.T) {
		m, players := setupGongZhuWebMockWithPlayers()
		// Player 0 captures a trick with the pig (♠Q), two hearts, and a non-point card.
		players[0].AddTrick([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 12, false), // ♠Q pig -> point
			domain.NewCard(domain.CardDesignHeart, 5, false),  // heart -> point
			domain.NewCard(domain.CardDesignHeart, 13, false), // heart -> point
			domain.NewCard(domain.CardDesignClover, 3, false), // non-point -> filtered out
		})
		// Player 1 captures the sheep (♦J) and the doubler (♣10).
		players[1].AddTrick([]*domain.Card{
			domain.NewCard(domain.CardDesignDiamond, 11, false), // ♦J sheep -> point
			domain.NewCard(domain.CardDesignClover, 10, false),  // ♣10 doubler -> point
			domain.NewCard(domain.CardDesignSpade, 3, false),    // non-point -> filtered out
		})

		result := p.Output(m, nil)
		var resObj controller.GongZhuWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))

		assert.Len(t, resObj.Players[0].CapturedPointCards, 3)
		assert.Len(t, resObj.Players[1].CapturedPointCards, 2)
		// Players who have taken no tricks report an empty (non-nil) slice.
		assert.NotNil(t, resObj.Players[2].CapturedPointCards)
		assert.Len(t, resObj.Players[2].CapturedPointCards, 0)

		// Player 0's point cards are exactly the pig and two hearts (clover 3 dropped).
		got := resObj.Players[0].CapturedPointCards
		assert.Equal(t, "SPADE", got[0].Design)
		assert.Equal(t, 12, got[0].Value)
		assert.Equal(t, "HEART", got[1].Design)
		assert.Equal(t, "HEART", got[2].Design)
	})

	t.Run("exposure flags & exposable indices", func(t *testing.T) {
		m, _ := setupGongZhuWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetExposure")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetExposableIndices")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetExposure").Return(domain.GongZhuExposure{Pig: true, Sheep: true})
		m.On("GetExposableIndices", 0).Return([]int{1, 4})
		m.On("GetPhase").Return(domain.GongZhuPhaseExpose)

		result := p.Output(m, nil)
		var resObj controller.GongZhuWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.Exposed.Pig)
		assert.True(t, resObj.Exposed.Sheep)
		assert.False(t, resObj.Exposed.Ace)
		assert.Equal(t, []int{1, 4}, resObj.ExposableIndices)
		assert.Equal(t, "gongzhu.exposePhase", resObj.MessageCode)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupGongZhuWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.GongZhuWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, int(domain.GongZhuCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, 1000, resObj.Config.PointLimit)
	})

	t.Run("play phase follow when trick has cards", func(t *testing.T) {
		m, _ := setupGongZhuWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
		})
		result := p.Output(m, nil)
		var resObj controller.GongZhuWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.CurrentTrick, 1)
		assert.Equal(t, "gongzhu.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end / round end message codes", func(t *testing.T) {
		for phase, code := range map[domain.GongZhuPhase]string{
			domain.GongZhuPhaseTrickEnd: "gongzhu.trickEnd",
			domain.GongZhuPhaseRoundEnd: "gongzhu.roundEnd",
		} {
			m, _ := setupGongZhuWebMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
			m.On("GetPhase").Return(phase)
			result := p.Output(m, nil)
			var resObj controller.GongZhuWebOutput
			assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
			assert.Equal(t, code, resObj.MessageCode)
		}
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupGongZhuWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.GongZhuWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human wins", func(t *testing.T) {
		m, _ := setupGongZhuWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		result := p.Output(m, nil)
		var resObj controller.GongZhuWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "gongzhu.result.humanWin", resObj.MessageCode)
	})

	t.Run("game end cpu wins", func(t *testing.T) {
		m, _ := setupGongZhuWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(2)
		result := p.Output(m, nil)
		var resObj controller.GongZhuWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "gongzhu.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"cpuId": "2"}, resObj.MessageParams)
	})
}

func TestGongZhuWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.GongZhuWebPresenter)

	t.Run("hint available", func(t *testing.T) {
		m, _ := setupGongZhuWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.GongZhuHint{CardIndices: []int{2}, Reason: "follow_suit"})
		result := p.HintOutput(m)
		var resObj controller.GongZhuWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, []int{2}, resObj.Hint.CardIndices)
		assert.Equal(t, "follow_suit", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupGongZhuWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.GongZhuHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.GongZhuWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestGongZhuWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.GongZhuWebPresenter)
	m := new(interfaces.MockGongZhuGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "plays ♠5"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, `"actionType":"play"`)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestGongZhuWebPresenterOutputCarriesTheHint(t *testing.T) {
	gzg, _ := setupGongZhuWebMockWithPlayers()
	gzg.ExpectedCalls = removeMockCall(gzg.ExpectedCalls, "GetHint")
	gzg.On("GetHint").Return(&domain.GongZhuHint{CardIndices: []int{0}, Reason: "follow_suit"})

	result := new(presenter.GongZhuWebPresenter).Output(gzg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	assert.NotContains(t, result, "gongzhu.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestGongZhuWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	gzg, _ := setupGongZhuWebMockWithPlayers()
	gzg.ExpectedCalls = removeMockCall(gzg.ExpectedCalls, "GetHint")
	gzg.On("GetHint").Return(&domain.GongZhuHint{CardIndices: []int{0}, Reason: "follow_suit"})
	assert.Contains(t, new(presenter.GongZhuWebPresenter).HintOutput(gzg), "gongzhu.hintRequested")

	none, _ := setupGongZhuWebMockWithPlayers()
	none.ExpectedCalls = removeMockCall(none.ExpectedCalls, "GetHint")
	none.On("GetHint").Return((*domain.GongZhuHint)(nil))
	assert.Contains(t, new(presenter.GongZhuWebPresenter).HintOutput(none), "gongzhu.noHint")
}

// **マストフォローの可視化 (#4812)。**人間の手番のプレイフェーズだけ、
// 出せる手札の位置を載せる。
func TestGongZhuWebPresenter_PlayableIndices(t *testing.T) {
	m, _ := setupGongZhuWebMockWithPlayers()
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayableIndices")
	m.On("GetPlayableIndices", 0).Return([]int{1, 3})

	var out controller.GongZhuWebOutput
	assert.NoError(t, json.Unmarshal([]byte(new(presenter.GongZhuWebPresenter).Output(m, nil)), &out))
	assert.Equal(t, []int{1, 3}, out.PlayableIndices)

	// CPU の手番では空 (null ではない)。
	m2, _ := setupGongZhuWebMockWithPlayers()
	m2.ExpectedCalls = removeMockCall(m2.ExpectedCalls, "IsHumanTurn")
	m2.On("IsHumanTurn").Return(false)

	var out2 controller.GongZhuWebOutput
	assert.NoError(t, json.Unmarshal([]byte(new(presenter.GongZhuWebPresenter).Output(m2, nil)), &out2))
	assert.Empty(t, out2.PlayableIndices)
}

// #5630: ラウンド終了時の得点内訳を Web にも運ぶ。数字だけでは「なぜその点か」
// を確かめられない。
func TestGongZhuWebPresenterCarriesTheBreakdownAtRoundEnd(t *testing.T) {
	m, _ := setupGongZhuWebMockWithPlayers()
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
	m.On("GetPhase").Return(domain.GongZhuPhaseRoundEnd)
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "ScoreBreakdownFor")
	m.On("ScoreBreakdownFor", mock.Anything).Return(domain.GongZhuScoreBreakdown{
		HeartCount: 3, HeartsSum: -120, HasPig: true, PigExposed: true, Subtotal: -320, Total: -320,
	})

	var out controller.GongZhuWebOutput
	require.NoError(t, json.Unmarshal([]byte(new(presenter.GongZhuWebPresenter).Output(m, nil)), &out))

	require.Len(t, out.ScoreBreakdowns, 4, "プレイヤー全員分")
	assert.Equal(t, 3, out.ScoreBreakdowns[0].HeartCount)
	assert.True(t, out.ScoreBreakdowns[0].PigExposed)
	assert.Equal(t, -320, out.ScoreBreakdowns[0].Total)
}

// プレイ中は出さない。まだ確定していない数字を並べても読めない。
func TestGongZhuWebPresenterOmitsTheBreakdownDuringPlay(t *testing.T) {
	m, _ := setupGongZhuWebMockWithPlayers()

	var out controller.GongZhuWebOutput
	require.NoError(t, json.Unmarshal([]byte(new(presenter.GongZhuWebPresenter).Output(m, nil)), &out))
	assert.Empty(t, out.ScoreBreakdowns)
}
