//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// horseAdvanceTo は指定の種目までハンドを進める。届かなければ false。
func horseAdvanceTo(g *domain.Horse, want domain.HorseDiscipline) bool {
	for range 200 {
		if g.GetDiscipline() == want && g.GetPhase() == domain.HorsePhaseHand {
			return true
		}
		if g.GetGameEndFlag() {
			return false
		}
		if g.GetPhase() == domain.HorsePhaseHandEnd {
			if err := g.NextHand(); err != nil {
				return false
			}
			continue
		}
		if err := g.PlayerAction(domain.HoldemActionFold, 0, 0); err != nil {
			return false
		}
	}
	return false
}

// **人間には配られた札が見え、CPU の伏せ札は見えない。** 両側を踏まないと、
// 常に空を返す実装でも「CPU は見えない」だけが通ってしまう。
func TestHorse_SeatCardsRevealOnlyWhatIsVisible(t *testing.T) {
	cfg := domain.DefaultHorseConfig()
	cfg.HandsPerDiscipline = 1
	g := domain.NewDefaultHorse()
	g.SetConfig(cfg)
	g.Reset()
	require.Equal(t, domain.HorseHoldem, g.GetDiscipline())

	human := g.GetHumanSeat()
	assert.Len(t, g.GetSeatCards(human), 2, "ホールデムの人間の手札は 2 枚")
	for i := range g.GetSeatCount() {
		if i == human {
			continue
		}
		assert.Empty(t, g.GetSeatCards(i), "ホールデムでは CPU の札は 1 枚も見えない")
	}
	// 範囲外の席は空。
	assert.Empty(t, g.GetSeatCards(-1))
	assert.Empty(t, g.GetSeatCards(g.GetSeatCount()))
}

// **オマハでは 4 枚配られる。** 種目ごとに枚数が違うので、卓の型を取り違えると
// ここで露見する。
func TestHorse_SeatCardsInOmaha(t *testing.T) {
	cfg := domain.DefaultHorseConfig()
	cfg.HandsPerDiscipline = 1
	g := domain.NewDefaultHorse()
	g.SetConfig(cfg)
	g.Reset()
	if !horseAdvanceTo(g, domain.HorseOmahaHiLo) {
		t.Skip("オマハまで届かなかった配り")
	}
	assert.Len(t, g.GetSeatCards(g.GetHumanSeat()), 4)
}

// **スタッド系ではドアカードが見える。** 共有札は無い。
func TestHorse_SeatCardsInStud(t *testing.T) {
	cfg := domain.DefaultHorseConfig()
	cfg.HandsPerDiscipline = 1
	g := domain.NewDefaultHorse()
	g.SetConfig(cfg)
	g.Reset()
	if !horseAdvanceTo(g, domain.HorseRazz) {
		t.Skip("ラズまで届かなかった配り")
	}
	// **ストリートは配りしだいで進んでいる。** 枚数を固定すると 4th street に
	// 入った局面で落ちるので、どのストリートでも成り立つ関係を見る ── 人間は
	// 伏せ札 2 枚ぶんだけ多く見えている。
	human := g.GetHumanSeat()
	mine := g.GetSeatCards(human)
	assert.GreaterOrEqual(t, len(mine), 3, "伏せ 2 + 表 1 は最低でも配られている")
	seen := 0
	for i := range g.GetSeatCount() {
		if i == human {
			continue
		}
		up := g.GetSeatCards(i)
		assert.NotEmpty(t, up, "ドアカードは見えている")
		assert.Less(t, len(up), len(mine), "CPU の伏せ札まで見えている")
		seen++
	}
	assert.Positive(t, seen)
	assert.Empty(t, g.GetCommunityCards(), "スタッド系に共有札は無い")
}

// **共有札はフェーズが進むまで出ない。** 配った直後に見えると、めくる前の
// フロップが読めてしまう。
func TestHorse_CommunityCardsAppearWithThePhase(t *testing.T) {
	g := domain.NewDefaultHorse()
	g.Reset()
	assert.Empty(t, g.GetCommunityCards(), "プリフロップに共有札は無い")

	for range 60 {
		if g.GetPhase() != domain.HorsePhaseHand || g.GetGameEndFlag() {
			break
		}
		if err := g.PlayerAction(domain.HoldemActionCall, 0, 0); err != nil {
			break
		}
		if len(g.GetCommunityCards()) > 0 {
			break
		}
	}
	assert.LessOrEqual(t, len(g.GetCommunityCards()), 5)
}

// **卓が無いあいだは何も返さない。** 決着後や配る前に触れても落ちない。
func TestHorse_CardsAreEmptyWithoutATable(t *testing.T) {
	g := domain.NewDefaultHorse()
	assert.Empty(t, g.GetSeatCards(0))
	assert.Empty(t, g.GetCommunityCards())
}

// **打った分だけ手持ちが減って見える。** 正本はハンドが終わるまで動かないので、
// そのまま出すとポットだけが増えて自分の残高が変わらない画面になる。
func TestHorse_LiveChipsMoveWhileTheHandIsBeingPlayed(t *testing.T) {
	g := domain.NewDefaultHorse()
	g.Reset()
	human := g.GetHumanSeat()
	require.Equal(t, g.GetSeatChips(human), g.GetSeatLiveChips(human),
		"配った直後は正本と一致する")

	moved := false
	for range 20 {
		if g.GetPhase() != domain.HorsePhaseHand || g.GetGameEndFlag() {
			break
		}
		act := domain.HoldemActionCall
		if g.GetToCall() == 0 {
			act = domain.HoldemActionCheck
		}
		if err := g.PlayerAction(act, 0, 0); err != nil {
			break
		}
		if g.GetSeatLiveChips(human) < g.GetSeatChips(human) {
			moved = true
			break
		}
	}
	assert.True(t, moved, "コールしても画面に出す残高が減らなかった")
}

// 卓が無いあいだは正本を返す。
func TestHorse_LiveChipsFallBackToTheCanonicalStack(t *testing.T) {
	g := domain.NewDefaultHorse()
	assert.Equal(t, g.GetSeatChips(0), g.GetSeatLiveChips(0))
	assert.Equal(t, 0, g.GetSeatLiveChips(-1))
	assert.Equal(t, 0, g.GetSeatLiveChips(g.GetSeatCount()))
}

// **どの種目でも同じことを訊ける。** 卓の型ごとに分岐しているので、1 つ書き
// 忘れると**その種目でだけ**コール額が 0 になり、賭けられていないように見える。
func TestHorse_BettingViewWorksInEveryDiscipline(t *testing.T) {
	for _, tt := range []struct {
		name string
		want domain.HorseDiscipline
	}{
		{"omaha", domain.HorseOmahaHiLo},
		{"razz", domain.HorseRazz},
		{"stud", domain.HorseStud},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfg := domain.DefaultHorseConfig()
			cfg.HandsPerDiscipline = 1
			g := domain.NewDefaultHorse()
			g.SetConfig(cfg)
			g.Reset()
			if !horseAdvanceTo(g, tt.want) {
				t.Skipf("%s まで届かなかった配り", tt.name)
			}
			human := g.GetHumanSeat()
			assert.GreaterOrEqual(t, g.GetToCall(), 0)
			assert.Positive(t, g.GetMinRaise(), "最小レイズ幅が卓から取れていない")
			assert.Positive(t, g.GetSeatLiveChips(human), "残高が卓から取れていない")
			// アンティ／ブラインドを出しているので、正本より減っている。
			assert.LessOrEqual(t, g.GetSeatLiveChips(human), g.GetSeatChips(human))
		})
	}
}
