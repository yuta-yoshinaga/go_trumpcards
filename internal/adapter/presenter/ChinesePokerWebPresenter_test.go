//go:build test

package presenter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestChinesePokerWebPresenter_Output(t *testing.T) {
	pp := &ChinesePokerWebPresenter{}
	cp := domain.NewDefaultChinesePoker()
	_ = cp.Bet(100)

	result := pp.Output(cp, nil)
	var output controller.ChinesePokerWebOutput
	err := json.Unmarshal([]byte(result), &output)
	require.NoError(t, err)
	assert.Equal(t, domain.ChinesePokerPhaseSetHands, output.Phase)
	assert.Equal(t, 100, output.Bet)
	assert.Len(t, output.PlayerCards, 13)
}

func TestChinesePokerWebPresenter_Output_WithError(t *testing.T) {
	pp := &ChinesePokerWebPresenter{}
	cp := domain.NewDefaultChinesePoker()
	testErr := domain.NewDomainError(domain.ErrWrongPhase, "test error")

	result := pp.Output(cp, testErr)
	var output controller.ChinesePokerWebOutput
	err := json.Unmarshal([]byte(result), &output)
	require.NoError(t, err)
	assert.Contains(t, output.Message, "test error")
}

func TestChinesePokerWebPresenter_Output_EndPhaseWin(t *testing.T) {
	pp := &ChinesePokerWebPresenter{}
	cp := domain.NewDefaultChinesePoker()
	cp.SetGameEndFlag(true)
	cp.SetResult(domain.GameResultWin)
	cp.SetScoop(false)

	result := pp.Output(cp, nil)
	var output controller.ChinesePokerWebOutput
	err := json.Unmarshal([]byte(result), &output)
	require.NoError(t, err)
	assert.Equal(t, "chinesepoker.result.playerWins", output.MessageCode)
}

func TestChinesePokerWebPresenter_Output_EndPhaseScoopWin(t *testing.T) {
	pp := &ChinesePokerWebPresenter{}
	cp := domain.NewDefaultChinesePoker()
	cp.SetGameEndFlag(true)
	cp.SetResult(domain.GameResultWin)
	cp.SetScoop(true)

	result := pp.Output(cp, nil)
	var output controller.ChinesePokerWebOutput
	err := json.Unmarshal([]byte(result), &output)
	require.NoError(t, err)
	assert.Equal(t, "chinesepoker.result.playerScoop", output.MessageCode)
}

func TestChinesePokerWebPresenter_Output_EndPhaseLose(t *testing.T) {
	pp := &ChinesePokerWebPresenter{}
	cp := domain.NewDefaultChinesePoker()
	cp.SetGameEndFlag(true)
	cp.SetResult(domain.GameResultLose)
	cp.SetScoop(false)

	result := pp.Output(cp, nil)
	var output controller.ChinesePokerWebOutput
	err := json.Unmarshal([]byte(result), &output)
	require.NoError(t, err)
	assert.Equal(t, "chinesepoker.result.dealerWins", output.MessageCode)
}

func TestChinesePokerWebPresenter_Output_EndPhaseScoopLose(t *testing.T) {
	pp := &ChinesePokerWebPresenter{}
	cp := domain.NewDefaultChinesePoker()
	cp.SetGameEndFlag(true)
	cp.SetResult(domain.GameResultLose)
	cp.SetScoop(true)

	result := pp.Output(cp, nil)
	var output controller.ChinesePokerWebOutput
	err := json.Unmarshal([]byte(result), &output)
	require.NoError(t, err)
	assert.Equal(t, "chinesepoker.result.dealerScoop", output.MessageCode)
}

func TestChinesePokerWebPresenter_ActionLogOutput(t *testing.T) {
	pp := &ChinesePokerWebPresenter{}
	cp := domain.NewDefaultChinesePoker()
	result := pp.ActionLogOutput(cp)
	assert.Contains(t, result, "[")
}
