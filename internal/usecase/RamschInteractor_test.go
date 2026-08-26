//go:build test

package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func newRamschInteractor() *usecase.RamschInteractor {
	return usecase.NewRamschInteractor(domain.NewDefaultRamsch(), new(presenter.RamschWebPresenter))
}

// Reset は配って、そのままプレイに入る。入札フェーズが残っていればここで見える。
func TestRamschInteractor_ResetStartsInPlay(t *testing.T) {
	si := newRamschInteractor()
	out := si.Reset()
	assert.NotEmpty(t, out)
	assert.Equal(t, domain.RamschPhasePlay, si.Game.GetPhase())
	assert.True(t, si.Game.IsHumanTurn(), "リセット直後は人間の手番まで進んでいるはず")
}

// 不正な着手はエラーを返し、盤を動かさない。
func TestRamschInteractor_RejectsAnIllegalPlay(t *testing.T) {
	si := newRamschInteractor()
	si.Reset()
	before := si.Game.GetPlayer(0).GetCardsSize()

	out := si.Play(99)
	assert.NotEmpty(t, out)
	assert.Equal(t, before, si.Game.GetPlayer(0).GetCardsSize(), "範囲外の着手で手札が減っている")
}

// **1 ラウンドを最後まで回す。** 各段の配線 (Play → ResolveTrick → NextTrick →
// ScoreRound) が繋がっていないと、どこかで止まる。
func TestRamschInteractor_PlaysARoundThroughToScoring(t *testing.T) {
	si := newRamschInteractor()
	si.Reset()

	for guard := 0; guard < 80; guard++ {
		switch si.Game.GetPhase() {
		case domain.RamschPhasePlay:
			require.True(t, si.Game.IsHumanTurn(), "CPU の手番で止まっている")
			valid := si.Game.GetValidPlayIndices(si.Game.GetCurrentPlayerIdx())
			require.NotEmpty(t, valid)
			si.Play(valid[0])
		case domain.RamschPhaseTrickEnd:
			si.NextTrick()
		case domain.RamschPhaseRoundEnd:
			out := si.NextRound()
			assert.NotEmpty(t, out)
			// 罰点なので、誰かの累計は必ず 0 以下になる。
			worst := 0
			for i := 0; i < si.Game.GetPlayerCnt(); i++ {
				if s := si.Game.GetPlayer(i).GetCumulativeScore(); s < worst {
					worst = s
				}
			}
			assert.Negative(t, worst, "1 ラウンド終えて誰も失点していない")
			return
		default:
			t.Fatalf("unexpected phase %d", si.Game.GetPhase())
		}
	}
	t.Fatal("ラウンドが終わらなかった")
}

// ヒントとアクションログは常に文字列を返す（nil 参照で落ちないこと）。
func TestRamschInteractor_HintAndLog(t *testing.T) {
	si := newRamschInteractor()
	si.Reset()
	assert.NotEmpty(t, si.Hint())
	assert.NotEmpty(t, si.ActionLog())
}

// 設定はそのまま往復する。
func TestRamschInteractor_ResetWithConfig(t *testing.T) {
	si := newRamschInteractor()
	cfg := domain.DefaultRamschConfig()
	cfg.CpuDifficulty = domain.RamschCpuDifficultyHard
	cfg.TargetScore = 250
	si.ResetWithConfig(cfg)

	got := si.GetConfig()
	assert.Equal(t, domain.RamschCpuDifficultyHard, got.CpuDifficulty)
	assert.Equal(t, 250, got.TargetScore)
}

// **KV 往復で盤が保たれること。** Worker はリクエストごとに状態を持たないので、
// ここが欠けると本番だけで壊れる。フィールドを覗くのではなく、
// **復元した盤で実際に指し続けられる**ことを見る。
func TestRamschInteractor_SurvivesAKVRoundTrip(t *testing.T) {
	si := newRamschInteractor()
	// 既定と違う設定にしておく。既定のまま往復させると、設定を丸ごと落としても
	// 「たまたま同じ」になって検査にならない。
	cfg := domain.DefaultRamschConfig()
	cfg.CpuDifficulty = domain.RamschCpuDifficultyHard
	cfg.TargetScore = 321
	si.ResetWithConfig(cfg)

	// **点が付いてから保存する。** 1 手だけ進めて保存すると誰の点も 0 のままで、
	// `cardPoints` をまるごと落としても 0 == 0 で通ってしまう（実際に通った）。
	//
	// 固定回数（2 トリック）では足りない: 7/8/9 だけのトリックは 0 点で、
	// 40 回に 1 回ほど「2 トリック消化しても 0 点」になった。**点が付くまで**
	// 進める ── 120 点あるので、ラウンドを終える前に必ずどこかで付く。
	pointsSoFar := func() int {
		total := 0
		for i := 0; i < si.Game.GetPlayerCnt(); i++ {
			total += si.Game.GetCardPoints(i)
		}
		return total
	}
	for guard := 0; guard < 80 && pointsSoFar() == 0; guard++ {
		switch si.Game.GetPhase() {
		case domain.RamschPhasePlay:
			require.True(t, si.Game.IsHumanTurn())
			v := si.Game.GetValidPlayIndices(si.Game.GetCurrentPlayerIdx())
			require.NotEmpty(t, v)
			si.Play(v[0])
		case domain.RamschPhaseTrickEnd:
			si.NextTrick()
		default:
			t.Fatalf("点が付く前にラウンドが終わった (phase %d)", si.Game.GetPhase())
		}
	}
	require.Positive(t, pointsSoFar(), "点が 1 点も付いていない ── 往復の検査が退化している")
	require.Equal(t, domain.RamschPhasePlay, si.Game.GetPhase(), "プレイ途中で保存できていない")

	// **場に札がある状態で保存する。** ここまでの実プレイだと、保存時に
	// ちょうどトリックの切れ目（人間がリード）で場が空のことがあり、
	// `currentTrick` をまるごと落としても通ってしまう（配りに依存して落ちた）。
	// 場の中身はドメインのヘルパで固定する。
	game, ok := si.Game.(*domain.Ramsch)
	require.True(t, ok)
	game.SetCurrentTrickForTest([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 9, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 12, false)},
	})

	data, err := si.Snapshot()
	require.NoError(t, err)

	restored, err := usecase.RestoreRamschInteractor(data, new(presenter.RamschWebPresenter))
	require.NoError(t, err)

	assert.Equal(t, si.Game.GetPhase(), restored.Game.GetPhase())
	assert.Equal(t, si.Game.GetTrickNumber(), restored.Game.GetTrickNumber())
	assert.Equal(t, si.Game.GetPlayer(0).GetCardsSize(), restored.Game.GetPlayer(0).GetCardsSize())
	for i := 0; i < restored.Game.GetPlayerCnt(); i++ {
		assert.Equal(t, si.Game.GetCardPoints(i), restored.Game.GetCardPoints(i), "player %d の点", i)
	}

	// **場に出ている札も往復すること。** 保存はトリックの途中で起こる
	// （人間の手番 = 誰かが既に出している）。落とすと、復元した盤で
	// 「リードが無い」ことになり、フォローの規則が丸ごと効かなくなる。
	require.NotEmpty(t, si.Game.GetCurrentTrick(), "トリック途中で保存できていない")
	require.Len(t, restored.Game.GetCurrentTrick(), len(si.Game.GetCurrentTrick()),
		"場の札が往復で消えている")
	for i, tc := range si.Game.GetCurrentTrick() {
		got := restored.Game.GetCurrentTrick()[i]
		assert.Equal(t, tc.PlayerIdx, got.PlayerIdx)
		assert.Equal(t, tc.Card.GetDesign(), got.Card.GetDesign())
		assert.Equal(t, tc.Card.GetValue(), got.Card.GetValue())
	}

	// 設定も往復すること。落とすと、復元のたびに難易度と目標点が既定へ戻る。
	assert.Equal(t, domain.RamschCpuDifficultyHard, restored.GetConfig().CpuDifficulty)
	assert.Equal(t, 321, restored.GetConfig().TargetScore)

	// **伏せ札も往復すること。** 最終トリックの獲得者が受け取る 2 枚なので、
	// 落とすと 120 点のうち数点が誰のものにもならない。
	require.Len(t, si.Game.GetSkat(), domain.RamschSkatSize)
	require.Len(t, restored.Game.GetSkat(), domain.RamschSkatSize, "伏せ札が往復で消えている")
	for i, c := range si.Game.GetSkat() {
		assert.Equal(t, c.GetDesign(), restored.Game.GetSkat()[i].GetDesign())
		assert.Equal(t, c.GetValue(), restored.Game.GetSkat()[i].GetValue())
	}

	// 復元した盤で指し続けられること。
	if restored.Game.GetPhase() == domain.RamschPhasePlay && restored.Game.IsHumanTurn() {
		v := restored.Game.GetValidPlayIndices(restored.Game.GetCurrentPlayerIdx())
		require.NotEmpty(t, v, "復元した盤に合法手が無い")
		before := restored.Game.GetPlayer(0).GetCardsSize()
		restored.Play(v[0])
		assert.Less(t, restored.Game.GetPlayer(0).GetCardsSize(), before, "復元後に指せていない")
	}
}
