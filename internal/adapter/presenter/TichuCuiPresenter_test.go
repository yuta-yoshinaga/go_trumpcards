package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestTichuCuiPresenter_Output(t *testing.T) {
	tg := newTichuAllCpu()
	tg.Reset()
	p := new(presenter.TichuCuiPresenter)

	declare := p.Output(tg, nil)
	assert.NotEmpty(t, declare)

	for tg.GetPhase() == domain.TichuPhaseDeclare {
		tg.CpuPlay()
	}
	play := p.Output(tg, nil)
	assert.Contains(t, play, "----------")

	for !tg.GetGameEndFlag() {
		tg.CpuPlay()
	}
	end := p.Output(tg, nil)
	assert.NotEmpty(t, end)

	withErr := p.Output(tg, errors.New("boom"))
	assert.True(t, strings.Contains(withErr, "boom"))
}

func TestTichuCuiPresenter_ActionLog(t *testing.T) {
	tg := newTichuAllCpu()
	tg.Reset()
	p := new(presenter.TichuCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(tg))
}

// **得点差もボムの使用状況も終局まで分からなかった。**Web は常時スコアバーに
// 出している (#4888)。
func TestTichuCuiPresenter_ShowsRunningScoreAndBombs(t *testing.T) {
	tg := newTichuAllCpu()
	tg.Reset()
	p := new(presenter.TichuCuiPresenter)

	// 宣言フェーズから既にスコアが出る。
	declare := p.Output(tg, nil)
	assert.Contains(t, declare, "チームA (P0/P2):")
	assert.Contains(t, declare, "チームB (P1/P3):")
	// まだボムは使われていないので行は出さない。
	assert.NotContains(t, declare, "ボム使用")

	for tg.GetPhase() == domain.TichuPhaseDeclare {
		tg.CpuPlay()
	}
	assert.Contains(t, p.Output(tg, nil), "チームA (P0/P2):")

	// ボムが使われたら回数が出る。
	tg.SetBombCountForTest(2)
	assert.Contains(t, p.Output(tg, nil), "ボム使用: 2回")

	// ワンツーが成立したらその旨も出る。
	tg.SetIsOneTwoForTest(true)
	assert.Contains(t, p.Output(tg, nil), "ワンツー成立")

	// **終局時に二重に出さない。**下の gameEnd ブロックが出す。
	for !tg.GetGameEndFlag() {
		tg.CpuPlay()
	}
	end := p.Output(tg, nil)
	assert.Equal(t, 1, strings.Count(end, "チームA (P0/P2):"))
	assert.Contains(t, end, "ディール終了")
}
