//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildWinnerResultMessage_Human(t *testing.T) {
	msg := buildWinnerResultMessage(0, true)
	assert.Equal(t, "ゲーム終了！ あなたの勝ち！", msg)
}

func TestBuildWinnerResultMessage_CPU(t *testing.T) {
	msg := buildWinnerResultMessage(2, false)
	assert.Equal(t, "ゲーム終了！ CPU 2の勝ち！", msg)
}

func TestBuildWinnerWebMessage_Human(t *testing.T) {
	resultMsg, code, params := buildWinnerWebMessage("hearts", 0, true)
	assert.Equal(t, "ゲーム終了！ あなたの勝ち！", resultMsg)
	assert.Equal(t, "hearts.result.humanWin", code)
	assert.Nil(t, params)
}

func TestBuildWinnerWebMessage_CPU(t *testing.T) {
	resultMsg, code, params := buildWinnerWebMessage("spades", 2, false)
	assert.Equal(t, "ゲーム終了！ CPU 2の勝ち！", resultMsg)
	assert.Equal(t, "spades.result.cpuWin", code)
	assert.Equal(t, map[string]string{"cpuId": "2"}, params)
}
