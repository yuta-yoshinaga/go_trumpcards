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

// unsunKarutaPassThrough は「呼ばれたか」だけを見る素通しのプレゼンター。
type unsunKarutaPassThrough struct{}

func (unsunKarutaPassThrough) Output(_ interfaces.UnsunKarutaGame, lastErr error) string {
	if lastErr != nil {
		return "err:" + lastErr.Error()
	}
	return "ok"
}
func (unsunKarutaPassThrough) HintOutput(_ interfaces.UnsunKarutaGame) string      { return "hint" }
func (unsunKarutaPassThrough) ActionLogOutput(_ interfaces.UnsunKarutaGame) string { return "log" }

func newUnsunKarutaReal() (*usecase.UnsunKarutaInteractor, *domain.UnsunKaruta) {
	g := domain.NewDefaultUnsunKaruta()
	return usecase.NewUnsunKarutaInteractor(g, unsunKarutaPassThrough{}), g
}

func TestNewUnsunKarutaInteractor_NilGuards(t *testing.T) {
	tp := new(presenter.MockUnsunKarutaPresenter)
	assert.PanicsWithValue(t, "UnsunKarutaInteractor: g must not be nil", func() {
		usecase.NewUnsunKarutaInteractor(nil, tp)
	})
	assert.PanicsWithValue(t, "UnsunKarutaInteractor: tp must not be nil", func() {
		usecase.NewUnsunKarutaInteractor(new(interfaces.MockUnsunKarutaGame), nil)
	})
}

// **Reset は人間の手番まで進める。** ここが動かないと、リードが CPU の
// ディールで盤面が最初の手番のまま止まる。
func TestUnsunKarutaInteractor_ResetAdvancesToTheHuman(t *testing.T) {
	hi, g := newUnsunKarutaReal()
	require.Equal(t, "ok", hi.Reset())
	assert.Equal(t, domain.UnsunKarutaPhasePlay, g.GetPhase())
	assert.True(t, g.IsHumanTurn(), "CPU の手番のまま止まっている")
	assert.Len(t, g.GetPlayers(), domain.UnsunKarutaPlayerCnt)
}

func TestUnsunKarutaInteractor_ResetWithConfig(t *testing.T) {
	hi, g := newUnsunKarutaReal()
	require.Equal(t, "ok", hi.ResetWithConfig(domain.UnsunKarutaConfig{
		CpuDifficulty: domain.UnsunKarutaCpuDifficultyEasy,
		TargetDeals:   2,
	}))
	assert.Equal(t, 2, hi.GetConfig().TargetDeals)
	assert.Equal(t, domain.UnsunKarutaCpuDifficultyEasy, g.GetConfig().CpuDifficulty)

	// 範囲外の設定は盤面を作り直さずに弾く。
	out := hi.ResetWithConfig(domain.UnsunKarutaConfig{TargetDeals: domain.UnsunKarutaMaxDeals + 1})
	assert.Contains(t, out, "err:")
	assert.Equal(t, 2, hi.GetConfig().TargetDeals, "弾いた設定が入ってしまっている")
}

// **人間が最後の 1 枚を出したらその場でトリックが決まる。** 解決しないまま
// 返すと、8 枚並んだ盤面で「次のトリック」も押せず固まる。
func TestUnsunKarutaInteractor_PlayResolvesAFullTrick(t *testing.T) {
	hi, g := newUnsunKarutaReal()
	require.Equal(t, "ok", hi.Reset())
	require.Equal(t, domain.UnsunKarutaHandSize, g.GetPlayer(0).GetCardsSize())

	for step := 0; step < 32 && g.GetPhase() == domain.UnsunKarutaPhasePlay; step++ {
		require.True(t, g.IsHumanTurn(), "インタラクターが CPU の手番で返している")
		valid := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
		require.NotEmpty(t, valid)
		require.Equal(t, "ok", hi.Play(valid[0], false))
	}
	require.Equal(t, domain.UnsunKarutaPhaseTrickEnd, g.GetPhase())
	assert.NotEqual(t, -1, g.GetLastTrickWinner(), "トリックが解決していない")
	assert.Len(t, g.GetCurrentTrick(), domain.UnsunKarutaPlayerCnt)
	assert.Equal(t, domain.UnsunKarutaHandSize-1, g.GetPlayer(0).GetCardsSize())
}

// **宣言はリードのときだけ。** リード以外で declare を送ったら弾く。
func TestUnsunKarutaInteractor_DeclareOffLeadIsRejected(t *testing.T) {
	hi, g := newUnsunKarutaReal()
	require.Equal(t, "ok", hi.Reset())
	if len(g.GetCurrentTrick()) == 0 {
		// 人間がリード: 宣言できる。
		valid := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
		require.NotEmpty(t, valid)
		require.Equal(t, "ok", hi.Play(valid[0], true))
		assert.True(t, g.IsMustFollow(), "宣言がフォロー義務を作っていない")
		return
	}
	out := hi.Play(g.GetPlayableIndices(g.GetCurrentPlayerIdx())[0], true)
	assert.Contains(t, out, "err:")
}

// 手番でないときの Play は盤面を触らずに返す。
func TestUnsunKarutaInteractor_PlayOffTurnIsANoop(t *testing.T) {
	hi, g := newUnsunKarutaReal()
	require.Equal(t, "ok", hi.Reset())
	for g.GetPhase() == domain.UnsunKarutaPhasePlay {
		require.Equal(t, "ok", hi.Play(g.GetPlayableIndices(g.GetCurrentPlayerIdx())[0], false))
	}
	require.Equal(t, domain.UnsunKarutaPhaseTrickEnd, g.GetPhase())
	before := g.GetPlayer(0).GetCardsSize()
	assert.Equal(t, "ok", hi.Play(0, false))
	assert.Equal(t, before, g.GetPlayer(0).GetCardsSize(), "手番外の Play が札を減らした")
}

// **1 ディールを最後まで打てる。** インタラクター経由でも CPU が回り、
// トリックが解決し、集計に届く。
func TestUnsunKarutaInteractor_PlaysADealThrough(t *testing.T) {
	hi, g := newUnsunKarutaReal()
	require.Equal(t, "ok", hi.Reset())

	for step := 0; step < 400; step++ {
		switch g.GetPhase() {
		case domain.UnsunKarutaPhasePlay:
			require.Equal(t, "ok", hi.Play(g.GetPlayableIndices(g.GetCurrentPlayerIdx())[0], false))
		case domain.UnsunKarutaPhaseTrickEnd:
			require.Equal(t, "ok", hi.NextTrick())
		case domain.UnsunKarutaPhaseRoundEnd:
			tricks := g.GetTeamTricks()
			// 9 トリックはどちらかの組が必ず取る。
			assert.Equal(t, domain.UnsunKarutaTrickCount, tricks[0]+tricks[1])
			require.Equal(t, "ok", hi.NextRound())
			assert.Equal(t, 2, g.GetRoundNumber())
			return
		case domain.UnsunKarutaPhaseGameEnd:
			t.Fatal("1 ディールで終局している")
		}
	}
	t.Fatal("ディールが終わらない")
}

// **最後のディールの集計で終局する。** NextRound がそのまま次を配ってしまうと
// マッチが終わらない。
func TestUnsunKarutaInteractor_StopsAtTheLastDeal(t *testing.T) {
	hi, g := newUnsunKarutaReal()
	require.Equal(t, "ok", hi.ResetWithConfig(domain.UnsunKarutaConfig{
		CpuDifficulty: domain.UnsunKarutaCpuDifficultyEasy,
		TargetDeals:   1,
	}))

	for step := 0; step < 400 && !g.GetGameEndFlag(); step++ {
		switch g.GetPhase() {
		case domain.UnsunKarutaPhasePlay:
			require.Equal(t, "ok", hi.Play(g.GetPlayableIndices(g.GetCurrentPlayerIdx())[0], false))
		case domain.UnsunKarutaPhaseTrickEnd:
			require.Equal(t, "ok", hi.NextTrick())
		case domain.UnsunKarutaPhaseRoundEnd:
			require.Equal(t, "ok", hi.NextRound())
		}
	}
	require.True(t, g.GetGameEndFlag(), "1 ディールのマッチが終わらない")
	assert.Equal(t, 1, g.GetRoundNumber(), "終わったのに次のディールを配っている")
	// 終局後の操作は盤面を触らない。
	assert.Equal(t, "ok", hi.Play(0, false))
	assert.Equal(t, "ok", hi.NextRound())
	assert.True(t, g.GetGameEndFlag())
}

func TestUnsunKarutaInteractor_HintAndActionLog(t *testing.T) {
	hi, _ := newUnsunKarutaReal()
	require.Equal(t, "ok", hi.Reset())
	assert.Equal(t, "hint", hi.Hint())
	assert.Equal(t, "log", hi.ActionLog())
}

// **保存した盤で指し続けられる。** 非公開フィールドだけの型は MarshalJSON が
// 無いと `{}` になり、復元した卓が空になる。
func TestUnsunKarutaInteractor_SnapshotRestoreKeepsPlaying(t *testing.T) {
	hi, g := newUnsunKarutaReal()
	require.Equal(t, "ok", hi.Reset())
	// 1 トリックぶん進めてから保存する (初期状態は退化していて何も証明しない)。
	for g.GetPhase() == domain.UnsunKarutaPhasePlay {
		require.Equal(t, "ok", hi.Play(g.GetPlayableIndices(g.GetCurrentPlayerIdx())[0], false))
	}
	require.Equal(t, "ok", hi.NextTrick())

	data, err := hi.Snapshot()
	require.NoError(t, err)
	require.Greater(t, len(data), 2, "空の JSON になっている")

	restored, err := usecase.RestoreUnsunKarutaInteractor(data, unsunKarutaPassThrough{})
	require.NoError(t, err)

	// 盤面が本当に戻っている: フェーズ・トリック番号・切り札・手札・取り分。
	rg := restored.Game
	assert.Equal(t, g.GetPhase(), rg.GetPhase())
	assert.Equal(t, g.GetRoundNumber(), rg.GetRoundNumber())
	assert.Equal(t, g.GetTrickNumber(), rg.GetTrickNumber())
	assert.Equal(t, g.GetTrumpSuit(), rg.GetTrumpSuit())
	assert.Equal(t, g.GetCurrentPlayerIdx(), rg.GetCurrentPlayerIdx())
	assert.Equal(t, g.GetTeamTricks(), rg.GetTeamTricks())
	assert.Equal(t, g.GetDealerIdx(), rg.GetDealerIdx())
	require.NotNil(t, rg.TrumpCard())
	assert.Equal(t, g.TrumpCard().GetDesign(), rg.TrumpCard().GetDesign())
	assert.Equal(t, g.TrumpCard().GetValue(), rg.TrumpCard().GetValue())
	for i := 0; i < domain.UnsunKarutaPlayerCnt; i++ {
		require.NotNil(t, rg.GetPlayer(i), "席 %d が空", i)
		assert.Equal(t, g.GetPlayer(i).GetCardsSize(), rg.GetPlayer(i).GetCardsSize(), "席 %d の手札", i)
		assert.Equal(t, g.GetPlayer(i).GetTrickCount(), rg.GetPlayer(i).GetTrickCount(), "席 %d の取り分", i)
	}

	// **復元した盤で指し続けられる。** 合法手は宣言の有無で変わるので、
	// 番号を決め打ちにすると配りによってはフォロー義務違反で弾かれる。
	require.True(t, rg.IsHumanTurn(), "復元した盤が人間の手番でない")
	legal := rg.GetPlayableIndices(rg.GetCurrentPlayerIdx())
	require.NotEmpty(t, legal)
	assert.Equal(t, "ok", restored.Play(legal[0], false))
}

func TestRestoreUnsunKarutaInteractor_RejectsGarbage(t *testing.T) {
	_, err := usecase.RestoreUnsunKarutaInteractor([]byte("{"), unsunKarutaPassThrough{})
	assert.Error(t, err)
}
