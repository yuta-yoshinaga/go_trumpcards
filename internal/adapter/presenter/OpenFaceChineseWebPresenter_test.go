//go:build test

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

func setupOpenFaceChineseWebMock() (*interfaces.MockOpenFaceChineseGame, []*domain.OpenFaceChinesePlayer) {
	m := new(interfaces.MockOpenFaceChineseGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.OpenFaceChinesePhasePlacing)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("IsHumanTurn").Return(true)
	m.On("GetCurrentCard").Return(domain.NewCard(domain.CardDesignSpade, 13, false))
	m.On("GetConfig").Return(domain.DefaultOpenFaceChineseConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	human := domain.NewOpenFaceChinesePlayer(true)
	human.SetPending([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 13, false)})
	cpu := domain.NewOpenFaceChinesePlayer(false)
	players := []*domain.OpenFaceChinesePlayer{human, cpu}
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(human)
	m.On("GetPlayer", 1).Return(cpu)
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()

	return m, players
}

func TestOpenFaceChineseWebPresenter_Output(t *testing.T) {
	p := new(presenter.OpenFaceChineseWebPresenter)

	t.Run("placing phase", func(t *testing.T) {
		m, _ := setupOpenFaceChineseWebMock()
		result := p.Output(m, nil)
		var out controller.OpenFaceChineseWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &out))
		assert.Equal(t, int(domain.OpenFaceChinesePhasePlacing), out.Phase)
		assert.Len(t, out.Players, 2)
		assert.Equal(t, "openfacechinese.placing", out.MessageCode)
		// human pending visible, cpu pending hidden.
		assert.Len(t, out.Players[0].Pending, 1)
		assert.Len(t, out.Players[1].Pending, 0)
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupOpenFaceChineseWebMock()
		result := p.Output(m, errors.New("boom"))
		var out controller.OpenFaceChineseWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &out))
		assert.Equal(t, "boom", out.Message)
	})

	t.Run("round end message", func(t *testing.T) {
		m, _ := setupOpenFaceChineseWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.OpenFaceChinesePhaseRoundEnd)
		result := p.Output(m, nil)
		var out controller.OpenFaceChineseWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &out))
		assert.Equal(t, "openfacechinese.roundEnd", out.MessageCode)
	})

	t.Run("game end human win", func(t *testing.T) {
		m, _ := setupOpenFaceChineseWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		result := p.Output(m, nil)
		var out controller.OpenFaceChineseWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &out))
		assert.Equal(t, "openfacechinese.result.humanWin", out.MessageCode)
	})

	t.Run("game end cpu win", func(t *testing.T) {
		m, _ := setupOpenFaceChineseWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(1)
		result := p.Output(m, nil)
		var out controller.OpenFaceChineseWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &out))
		assert.Equal(t, "openfacechinese.result.cpuWin", out.MessageCode)
	})

	t.Run("game end draw", func(t *testing.T) {
		m, _ := setupOpenFaceChineseWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		result := p.Output(m, nil)
		var out controller.OpenFaceChineseWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &out))
		assert.Equal(t, "openfacechinese.result.draw", out.MessageCode)
	})
}

func TestOpenFaceChineseWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.OpenFaceChineseWebPresenter)

	t.Run("with hint", func(t *testing.T) {
		m, _ := setupOpenFaceChineseWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.OpenFaceChineseHint{Row: domain.OpenFaceChineseRowBack, Reason: "strong_back"})
		result := p.HintOutput(m)
		var out controller.OpenFaceChineseWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &out))
		assert.NotNil(t, out.Hint)
		assert.Equal(t, domain.OpenFaceChineseRowBack, out.Hint.Row)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupOpenFaceChineseWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.OpenFaceChineseHint)(nil))
		result := p.HintOutput(m)
		var out controller.OpenFaceChineseWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &out))
		assert.Nil(t, out.Hint)
	})
}

func TestOpenFaceChineseWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.OpenFaceChineseWebPresenter)
	m, _ := setupOpenFaceChineseWebMock()
	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
//
// Output 側にゲートは置きません。OpenFaceChinese.GetHint() が「人間の手番で、かつ
// 行動を選べる状態か」を自分で確かめて nil を返します。
func TestOpenFaceChineseWebPresenterOutputCarriesTheHint(t *testing.T) {
	ofc, _ := setupOpenFaceChineseWebMock()
	ofc.ExpectedCalls = removeMockCall(ofc.ExpectedCalls, "GetHint")
	ofc.On("GetHint").Return(&domain.OpenFaceChineseHint{Row: 1, Reason: "balance"})

	result := new(presenter.OpenFaceChineseWebPresenter).Output(ofc, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
// ページは `isRequestedHint` でこのコードを見てからバナーを出すので (#4605)、
// 付いていないとヒントを押しても画面に何も出ない。
func TestOpenFaceChineseWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	g := domain.NewDefaultOpenFaceChinese()
	g.Reset()
	// **Reset 直後は人間の手番とは限らない。**GetHint は手番でなければ nil を
	// 返すので、席を人間に固定しないとこのテストは前提で落ちる。
	g.SetCurrentPlayerIdx(0)
	require.NotNil(t, g.GetHint(), "fixture must actually produce a hint")
	assert.Contains(t, new(presenter.OpenFaceChineseWebPresenter).HintOutput(g), "openfacechinese.hintRequested")
}
