//go:build test

package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// quodlibetPassThrough は「呼ばれたか」だけを見る素通しのプレゼンター。
type quodlibetPassThrough struct{}

func (quodlibetPassThrough) Output(_ interfaces.QuodlibetGame, lastErr error) string {
	if lastErr != nil {
		return "err:" + lastErr.Error()
	}
	return "ok"
}
func (quodlibetPassThrough) HintOutput(_ interfaces.QuodlibetGame) string      { return "hint" }
func (quodlibetPassThrough) ActionLogOutput(_ interfaces.QuodlibetGame) string { return "log" }

func newQuodlibetReal() (*usecase.QuodlibetInteractor, *domain.Quodlibet) {
	g := domain.NewDefaultQuodlibet()
	return usecase.NewQuodlibetInteractor(g, quodlibetPassThrough{}), g
}

func TestNewQuodlibetInteractor_NilGuards(t *testing.T) {
	qp := new(presenter.MockQuodlibetPresenter)
	assert.PanicsWithValue(t, "QuodlibetInteractor: g must not be nil", func() {
		usecase.NewQuodlibetInteractor(nil, qp)
	})
	assert.PanicsWithValue(t, "QuodlibetInteractor: qp must not be nil", func() {
		usecase.NewQuodlibetInteractor(new(interfaces.MockQuodlibetGame), nil)
	})
}

// 席 0 が第 1 ディールの親なので、Reset は人間のコントラクト選択で止まる。
func TestQuodlibetInteractor_ResetStopsAtTheHumanContractChoice(t *testing.T) {
	qi, g := newQuodlibetReal()
	require.Equal(t, "ok", qi.Reset())
	assert.Equal(t, domain.QuodlibetPhaseSelectContract, g.GetPhase())
	assert.True(t, g.IsHumanTurn())
	assert.Len(t, g.GetAvailableContracts(), domain.QuodlibetContractsPerRound)
}

// **CPU が親の輪でも盤面は進む。** 選ばせずに抜けると、選択フェーズのまま
// 誰も何もできない盤面で止まる。
func TestQuodlibetInteractor_CpuDealerChoosesWithoutBeingAsked(t *testing.T) {
	qi, g := newQuodlibetReal()
	require.Equal(t, "ok", qi.Reset())
	require.Equal(t, "ok", qi.SelectContract(domain.QuodlibetMinus))
	quodlibetFinishDeal(t, qi, g)
	require.Equal(t, "ok", qi.NextDeal())

	require.False(t, g.GetPlayer(g.GetDealerIdx()).GetIsHuman(), "第 2 ディールの親は CPU")
	assert.NotEqual(t, domain.QuodlibetPhaseSelectContract, g.GetPhase(),
		"CPU の親が種目を選んでいない")
	assert.GreaterOrEqual(t, g.GetCurrentContract(), 0)
}

func TestQuodlibetInteractor_ResetWithConfig(t *testing.T) {
	qi, g := newQuodlibetReal()
	require.Equal(t, "ok", qi.ResetWithConfig(domain.QuodlibetConfig{
		CpuDifficulty:      domain.QuodlibetCpuDifficultyEasy,
		AutoSelectContract: true,
	}))
	assert.Equal(t, domain.QuodlibetCpuDifficultyEasy, qi.GetConfig().CpuDifficulty)
	// 自動選択なら人間にも訊かず、そのままプレイに入る。
	assert.Equal(t, domain.QuodlibetPhasePlay, g.GetPhase())

	out := qi.ResetWithConfig(domain.QuodlibetConfig{CpuDifficulty: 9})
	assert.Contains(t, out, "err:")
	assert.Equal(t, domain.QuodlibetCpuDifficultyEasy, qi.GetConfig().CpuDifficulty,
		"弾いた設定が入ってしまっている")
}

// **選べるのはこの輪の残りだけ。** 別の輪の種目は弾く。
func TestQuodlibetInteractor_RejectsAContractFromAnotherWheel(t *testing.T) {
	qi, g := newQuodlibetReal()
	require.Equal(t, "ok", qi.Reset())
	assert.Contains(t, qi.SelectContract(domain.QuodlibetSnack), "err:")
	assert.Equal(t, domain.QuodlibetPhaseSelectContract, g.GetPhase(), "弾いたのに進んでいる")
	assert.Equal(t, -1, g.GetCurrentContract())
}

// 手番でない Play は盤面を触らずに返す。
func TestQuodlibetInteractor_PlayOffTurnIsANoop(t *testing.T) {
	qi, g := newQuodlibetReal()
	require.Equal(t, "ok", qi.Reset())
	// 選択フェーズでは IsHumanTurn は真だが、出せる札は無い。
	before := g.GetPlayer(0).GetCardsSize()
	qi.Play(0)
	assert.Equal(t, before, g.GetPlayer(0).GetCardsSize(), "選択フェーズで札が減った")
}

// **1 ディールを最後まで打てる。** インタラクター経由でも CPU が回り、
// トリックが解決し、集計に届く。
func TestQuodlibetInteractor_PlaysADealThrough(t *testing.T) {
	qi, g := newQuodlibetReal()
	require.Equal(t, "ok", qi.Reset())
	require.Equal(t, "ok", qi.SelectContract(domain.QuodlibetMinus))
	quodlibetFinishDeal(t, qi, g)

	assert.Equal(t, domain.QuodlibetPhaseDealEnd, g.GetPhase())
	detail := g.GetLastDealDetail()
	require.NotNil(t, detail)
	assert.Equal(t, domain.QuodlibetMinus, detail.Contract)
	// マイナスは 8 トリックぶんの罰点が誰かに付く。
	total := 0
	for i := 0; i < domain.QuodlibetPlayerCnt; i++ {
		total += detail.Points[i]
	}
	assert.Positive(t, total, "誰も罰点を負っていない")
}

// **12 ディールで終局する。** 13 個目を配ってしまうとマッチが終わらない。
func TestQuodlibetInteractor_StopsAfterTwelveDeals(t *testing.T) {
	qi, g := newQuodlibetReal()
	require.Equal(t, "ok", qi.ResetWithConfig(domain.QuodlibetConfig{
		CpuDifficulty:      domain.QuodlibetCpuDifficultyEasy,
		AutoSelectContract: true,
	}))

	for deal := 0; deal < domain.QuodlibetTotalDeals+2 && !g.GetGameEndFlag(); deal++ {
		quodlibetFinishDeal(t, qi, g)
		require.Equal(t, "ok", qi.NextDeal())
	}
	require.True(t, g.GetGameEndFlag(), "12 ディールで終局しない")
	assert.Len(t, g.GetDealHistory(), domain.QuodlibetTotalDeals)
	// 終局後の操作は盤面を触らない。
	assert.Equal(t, "ok", qi.NextDeal())
	assert.Equal(t, "ok", qi.Play(0))
	assert.Contains(t, qi.SelectContract(0), "ok")
	assert.True(t, g.GetGameEndFlag())
}

func TestQuodlibetInteractor_HintAndActionLog(t *testing.T) {
	qi, _ := newQuodlibetReal()
	require.Equal(t, "ok", qi.Reset())
	assert.Equal(t, "hint", qi.Hint())
	assert.Equal(t, "log", qi.ActionLog())
}

// **保存した盤で指し続けられる。** 非公開フィールドだけの型は MarshalJSON が
// 無いと `{}` になり、復元した卓が空になる。
func TestQuodlibetInteractor_SnapshotRestoreKeepsPlaying(t *testing.T) {
	qi, g := newQuodlibetReal()
	require.Equal(t, "ok", qi.Reset())
	require.Equal(t, "ok", qi.SelectContract(domain.QuodlibetMinus))
	// 初期状態は退化しているので、1 手進めてから保存する。
	if g.GetPhase() == domain.QuodlibetPhasePlay && g.IsHumanTurn() {
		require.Equal(t, "ok", qi.Play(g.GetPlayableIndices(g.GetCurrentTurn())[0]))
	}

	data, err := qi.Snapshot()
	require.NoError(t, err)
	require.Greater(t, len(data), 2, "空の JSON になっている")

	restored, err := usecase.RestoreQuodlibetInteractor(data, quodlibetPassThrough{})
	require.NoError(t, err)
	rg := restored.Game
	assert.Equal(t, g.GetPhase(), rg.GetPhase())
	assert.Equal(t, g.GetCurrentContract(), rg.GetCurrentContract())
	assert.Equal(t, g.GetDealNumber(), rg.GetDealNumber())
	assert.Equal(t, g.GetDealerIdx(), rg.GetDealerIdx())
	assert.Equal(t, g.GetCurrentTurn(), rg.GetCurrentTurn())
	assert.Equal(t, g.GetUsedContracts(), rg.GetUsedContracts())
	for i := 0; i < domain.QuodlibetPlayerCnt; i++ {
		require.NotNil(t, rg.GetPlayer(i), "席 %d が空", i)
		assert.Equal(t, g.GetPlayer(i).GetCardsSize(), rg.GetPlayer(i).GetCardsSize(), "席 %d の手札", i)
		assert.Equal(t, g.GetPlayer(i).GetPenalty(), rg.GetPlayer(i).GetPenalty(), "席 %d の罰点", i)
	}

	// 復元した盤で最後まで打てる。
	for deal := 0; deal < domain.QuodlibetTotalDeals+2 && !rg.GetGameEndFlag(); deal++ {
		quodlibetFinishDeal(t, restored, rg)
		require.Equal(t, "ok", restored.NextDeal())
	}
	assert.True(t, rg.GetGameEndFlag(), "復元した盤で終局に届かない")
}

func TestRestoreQuodlibetInteractor_RejectsGarbage(t *testing.T) {
	_, err := usecase.RestoreQuodlibetInteractor([]byte("{"), quodlibetPassThrough{})
	assert.Error(t, err)
}

// quodlibetFinishDeal は現在のディールを DealEnd まで打つ。
func quodlibetFinishDeal(t *testing.T, qi *usecase.QuodlibetInteractor, g interfaces.QuodlibetGame) {
	t.Helper()
	for step := 0; step < 600; step++ {
		switch g.GetPhase() {
		case domain.QuodlibetPhaseSelectContract:
			avail := g.GetAvailableContracts()
			require.NotEmpty(t, avail)
			require.Equal(t, "ok", qi.SelectContract(avail[0]))
		case domain.QuodlibetPhasePlay:
			idx := -1
			if valid := g.GetPlayableIndices(g.GetCurrentTurn()); len(valid) > 0 {
				idx = valid[0]
			}
			require.Equal(t, "ok", qi.Play(idx))
		default:
			return
		}
	}
	t.Fatal("ディールが終わらない")
}
