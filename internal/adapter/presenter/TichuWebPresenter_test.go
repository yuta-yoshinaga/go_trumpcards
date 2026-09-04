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
)

func newTichuAllCpu() *domain.Tichu {
	players := []*domain.TichuPlayer{
		domain.NewTichuPlayer(false),
		domain.NewTichuPlayer(false),
		domain.NewTichuPlayer(false),
		domain.NewTichuPlayer(false),
	}
	return domain.NewTichu(domain.NewTrumpCards(domain.TichuJokerCount), players, domain.DefaultTichuConfig())
}

func TestTichuWebPresenter_DeclarePhase(t *testing.T) {
	tg := newTichuAllCpu()
	tg.Reset()

	p := new(presenter.TichuWebPresenter)
	var resp controller.TichuWebOutput
	require.NoError(t, json.Unmarshal([]byte(p.Output(tg, nil)), &resp))
	assert.Equal(t, "declare", resp.Phase)
	assert.Len(t, resp.Players, domain.TichuPlayerCnt)
	assert.False(t, resp.GameEndFlag)
	// teams alternate
	assert.Equal(t, 0, resp.Players[0].Team)
	assert.Equal(t, 1, resp.Players[1].Team)
}

func TestTichuWebPresenter_PlayAndEnd(t *testing.T) {
	tg := newTichuAllCpu()
	tg.Reset()
	for tg.GetPhase() == domain.TichuPhaseDeclare {
		tg.CpuPlay()
	}
	p := new(presenter.TichuWebPresenter)
	var play controller.TichuWebOutput
	require.NoError(t, json.Unmarshal([]byte(p.Output(tg, nil)), &play))
	assert.Equal(t, "play", play.Phase)

	for !tg.GetGameEndFlag() {
		tg.CpuPlay()
	}
	var end controller.TichuWebOutput
	require.NoError(t, json.Unmarshal([]byte(p.Output(tg, nil)), &end))
	assert.Equal(t, "end", end.Phase)
	assert.True(t, end.GameEndFlag)
	assert.NotEmpty(t, end.Message)
}

func TestTichuWebPresenter_Error(t *testing.T) {
	tg := newTichuAllCpu()
	tg.Reset()
	p := new(presenter.TichuWebPresenter)
	var resp controller.TichuWebOutput
	require.NoError(t, json.Unmarshal([]byte(p.Output(tg, errors.New("boom"))), &resp))
	assert.Equal(t, "boom", resp.Message)
}

func TestTichuWebPresenter_ActionLog(t *testing.T) {
	tg := newTichuAllCpu()
	tg.Reset()
	p := new(presenter.TichuWebPresenter)
	out := p.ActionLogOutput(tg)
	assert.NotEmpty(t, out)
}

func TestTichuWebPresenter_DogLead(t *testing.T) {
	p := new(presenter.TichuWebPresenter)

	// **席の名前を Go 側で埋めない。**フロントは `messageCode` を自分のロケールで
	// 組み直すので、名前を渡すと英語の文の中に「あなた」が残る。誰が人間かで文が
	// 変わるのでコードを 3 つに分けてある。
	t.Run("the human playing the dog names no seat but its own code", func(t *testing.T) {
		tg := domain.NewDefaultTichu()
		tg.Reset()
		tg.SetDogLeadPassedForTest(true, 0) // Human (P0) played dog

		var resp controller.TichuWebOutput
		require.NoError(t, json.Unmarshal([]byte(p.Output(tg, nil)), &resp))
		assert.Equal(t, "tichu.dogLeadPassedByYou", resp.MessageCode)
		assert.Equal(t, "2", resp.MessageParams["to"])
		assert.NotContains(t, resp.MessageParams["to"], "CPU", "params carry the seat, not a rendered name")
		assert.Contains(t, resp.Message, "犬:")
		assert.NotContains(t, resp.Message, "{{")
	})

	t.Run("a CPU passing the lead to the human gets the to-you code", func(t *testing.T) {
		tg := domain.NewDefaultTichu()
		tg.Reset()
		tg.SetDogLeadPassedForTest(true, 2) // CPU 2's partner is the human at seat 0

		var resp controller.TichuWebOutput
		require.NoError(t, json.Unmarshal([]byte(p.Output(tg, nil)), &resp))
		assert.Equal(t, "tichu.dogLeadPassedToYou", resp.MessageCode)
		assert.Equal(t, "2", resp.MessageParams["from"])
		assert.NotContains(t, resp.Message, "{{")
	})

	t.Run("a dog between two CPUs names both seats", func(t *testing.T) {
		tg := domain.NewDefaultTichu()
		tg.Reset()
		tg.SetDogLeadPassedForTest(true, 1)

		var resp controller.TichuWebOutput
		require.NoError(t, json.Unmarshal([]byte(p.Output(tg, nil)), &resp))
		assert.Equal(t, "tichu.dogLeadPassedBetweenCpus", resp.MessageCode)
		assert.Equal(t, "1", resp.MessageParams["from"])
		assert.Equal(t, "3", resp.MessageParams["to"])
		assert.NotContains(t, resp.Message, "{{")
	})

	t.Run("says nothing when no dog was played", func(t *testing.T) {
		tg := domain.NewDefaultTichu()
		tg.Reset()
		tg.SetDogLeadPassedForTest(false, 0)

		var resp controller.TichuWebOutput
		require.NoError(t, json.Unmarshal([]byte(p.Output(tg, nil)), &resp))
		assert.NotContains(t, resp.MessageCode, "dogLeadPassed")
	})

	t.Run("game end takes priority over dogLeadPassed", func(t *testing.T) {
		tg := newTichuAllCpu()
		tg.Reset()
		for !tg.GetGameEndFlag() {
			tg.CpuPlay()
		}
		tg.SetDogLeadPassedForTest(true, 1)

		var resp controller.TichuWebOutput
		require.NoError(t, json.Unmarshal([]byte(p.Output(tg, nil)), &resp))
		assert.Equal(t, "tichu.result.summary", resp.MessageCode)
	})

	// **人間が犬を出した瞬間の告知が、CPU の手番を回した後も残っていること。**
	// `TichuInteractor.Play` は人間の 1 手のあと CPU を回しきってから描画するので、
	// CPU 側でフラグを消すと、この告知は**一度も画面に出ない** (#6431)。
	// 単体で `SetDogLeadPassedForTest` を叩くテストは、この経路を一度も通らない。
	t.Run("survives the CPU turns that run in the same request", func(t *testing.T) {
		tg := domain.NewDefaultTichu()
		tg.Reset()
		tichuGiveHumanTheDog(t, tg)

		require.NoError(t, tg.PlayerPlay([]int{tg.GetPlayer(0).GetCardsSize() - 1}))
		// インタラクタと同じように、人間の手番に戻るまで CPU を回す。
		for i := 0; i < 16 && !tg.IsHumanTurn() && !tg.GetGameEndFlag(); i++ {
			tg.CpuPlay()
		}

		var resp controller.TichuWebOutput
		require.NoError(t, json.Unmarshal([]byte(p.Output(tg, nil)), &resp))
		assert.Equal(t, "tichu.dogLeadPassedByYou", resp.MessageCode)
	})

	t.Run("error takes priority over dogLeadPassed", func(t *testing.T) {
		tg := domain.NewDefaultTichu()
		tg.Reset()
		tg.SetDogLeadPassedForTest(true, 0)

		var resp controller.TichuWebOutput
		require.NoError(t, json.Unmarshal([]byte(p.Output(tg, errors.New("invalid"))), &resp))
		assert.Equal(t, "invalid", resp.Message)
		assert.Empty(t, resp.MessageCode)
	})
}

// tichuGiveHumanTheDog puts the Dog on top of the human's hand and hands them the turn.
func tichuGiveHumanTheDog(t *testing.T, tg *domain.Tichu) {
	t.Helper()
	tg.SetPhaseForTest(domain.TichuPhasePlay)
	tg.SetCurrentTurnForTest(0)
	tg.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignJoker, domain.TichuDog, false))
}
