//go:build test

package usecase_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// piedmonteseTarotPassThrough は「呼ばれたか」だけを見る素通しのプレゼンター。
type piedmonteseTarotPassThrough struct{}

func (piedmonteseTarotPassThrough) Output(_ interfaces.PiedmonteseTarotGame, lastErr error) string {
	if lastErr != nil {
		return "err:" + lastErr.Error()
	}
	return "ok"
}
func (piedmonteseTarotPassThrough) HintOutput(_ interfaces.PiedmonteseTarotGame) string {
	return "hint"
}
func (piedmonteseTarotPassThrough) ActionLogOutput(_ interfaces.PiedmonteseTarotGame) string {
	return "log"
}

func newPiedmonteseTarotReal() (*usecase.PiedmonteseTarotInteractor, *domain.PiedmonteseTarot) {
	g := domain.NewDefaultPiedmonteseTarot()
	return usecase.NewPiedmonteseTarotInteractor(g, piedmonteseTarotPassThrough{}), g
}

func TestNewPiedmonteseTarotInteractor_NilGuards(t *testing.T) {
	tp := new(presenter.MockPiedmonteseTarotPresenter)
	assert.PanicsWithValue(t, "PiedmonteseTarotInteractor: g must not be nil", func() {
		usecase.NewPiedmonteseTarotInteractor(nil, tp)
	})
	assert.PanicsWithValue(t, "PiedmonteseTarotInteractor: tp must not be nil", func() {
		usecase.NewPiedmonteseTarotInteractor(new(interfaces.MockPiedmonteseTarotGame), nil)
	})
}

// **CPU の親は自動でスカルトする。** ここが動かないと、CPU が親のディールは
// 人間が何も打てないまま止まる。
func TestPiedmonteseTarotInteractor_ResetAdvancesPastACpuDealer(t *testing.T) {
	hi, g := newPiedmonteseTarotReal()
	require.Equal(t, "ok", hi.Reset())
	// 席 0 が親なので人間のスカルト待ちになる。
	require.True(t, g.IsHumanScartoTurn())

	// 親を CPU に移して配り直すと、スカルト待ちのまま止まる。
	g.SetDealerForTest(1)
	require.Equal(t, domain.PiedmonteseTarotPhaseScarto, g.GetPhase())
	require.False(t, g.IsHumanScartoTurn())

	// **インタラクターを 1 回通せば CPU の親が捨てる。** ここが動かないと、
	// CPU が親のディールは人間が何も打てないまま止まる。
	require.Equal(t, "ok", hi.NextTrick())
	assert.NotEqual(t, domain.PiedmonteseTarotPhaseScarto, g.GetPhase(),
		"CPU の親がスカルトしていない")
	assert.True(t, g.IsHumanTurn() || g.GetPhase() != domain.PiedmonteseTarotPhasePlay,
		"人間の手番まで進んでいない")
}

// **1 ディールを最後まで打てる。** インタラクター経由でも CPU が回り、
// トリックが解決し、精算に届く。
func TestPiedmonteseTarotInteractor_PlaysADealThrough(t *testing.T) {
	hi, g := newPiedmonteseTarotReal()
	require.Equal(t, "ok", hi.Reset())
	require.Equal(t, "ok", hi.Discard(domain.PiedmonteseTarotCpuScartoForTest(g, g.GetDealerIdx())))

	for step := 0; step < 400; step++ {
		switch g.GetPhase() {
		case domain.PiedmonteseTarotPhasePlay:
			valid := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
			require.NotEmpty(t, valid)
			require.Equal(t, "ok", hi.Play(valid[0]))
		case domain.PiedmonteseTarotPhaseTrickEnd:
			require.Equal(t, "ok", hi.NextTrick())
		default:
			require.Contains(t, []domain.PiedmonteseTarotPhase{
				domain.PiedmonteseTarotPhaseRoundEnd, domain.PiedmonteseTarotPhaseGameEnd,
			}, g.GetPhase())
			total := 0
			for _, v := range g.CapturedThirds() {
				total += v
			}
			assert.Equal(t, domain.PiedmonteseTarotTotalThirds, total, "札の取り分が合わない")
			return
		}
	}
	require.FailNow(t, "ディールが終わらない")
}

func TestPiedmonteseTarotInteractor_HintAndLog(t *testing.T) {
	hi, _ := newPiedmonteseTarotReal()
	require.Equal(t, "ok", hi.Reset())
	assert.Equal(t, "hint", hi.Hint())
	assert.Equal(t, "log", hi.ActionLog())
}

// **配れない設定は断る。** 5 人卓には配り方が無い。
func TestPiedmonteseTarotInteractor_ResetWithConfigRejectsAnUndealtTable(t *testing.T) {
	hi, g := newPiedmonteseTarotReal()
	cfg := domain.DefaultPiedmonteseTarotConfig()
	cfg.Seats = 5
	assert.Contains(t, hi.ResetWithConfig(cfg), "err:")
	assert.Equal(t, domain.PiedmonteseTarotDefaultSeats, g.GetConfig().Seats, "卓が作り替えられている")

	cfg.Seats = 3
	assert.Equal(t, "ok", hi.ResetWithConfig(cfg))
	assert.Equal(t, 3, g.GetPlayerCnt(), "席数を変えたのに席が作り直されていない")
	assert.Equal(t, 25, g.HandSize())
}

func TestPiedmonteseTarotInteractor_SnapshotRoundTrip(t *testing.T) {
	hi, g := newPiedmonteseTarotReal()
	require.Equal(t, "ok", hi.Reset())
	require.Equal(t, "ok", hi.Discard(domain.PiedmonteseTarotCpuScartoForTest(g, g.GetDealerIdx())))

	data, err := hi.Snapshot()
	require.NoError(t, err)
	var probe map[string]any
	require.NoError(t, json.Unmarshal(data, &probe))

	restored, err := usecase.RestorePiedmonteseTarotInteractor(data, piedmonteseTarotPassThrough{})
	require.NoError(t, err)
	assert.Equal(t, hi.GetConfig(), restored.GetConfig())
	// **復元した卓で打ち続けられる。**
	assert.Contains(t, []string{"ok", "err:" + domain.ErrWrongPhase.Error()},
		restored.NextTrick())
}

func TestPiedmonteseTarotInteractor_RestoreRejectsGarbage(t *testing.T) {
	_, err := usecase.RestorePiedmonteseTarotInteractor([]byte(`{`), piedmonteseTarotPassThrough{})
	assert.Error(t, err)
}
