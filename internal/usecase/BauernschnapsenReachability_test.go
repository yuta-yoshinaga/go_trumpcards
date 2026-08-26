//go:build test

package usecase_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// bauernschnapsenOut は Web プレゼンタが返す JSON のうち、この試験で見る分だけ。
type bauernschnapsenOut struct {
	Phase            int   `json:"phase"`
	CurrentPlayerIdx int   `json:"currentPlayerIdx"`
	Contract         int   `json:"contract"`
	DeclarerIdx      int   `json:"declarerIdx"`
	TrumpSuit        int   `json:"trumpSuit"`
	ValidPlayIndices []int `json:"validPlayIndices"`
	GameEndFlag      bool  `json:"gameEndFlag"`
}

func parseBauernschnapsenOut(t *testing.T, s string) bauernschnapsenOut {
	t.Helper()
	var o bauernschnapsenOut
	require.NoError(t, json.Unmarshal([]byte(s), &o))
	return o
}

// TestBauernschnapsenInteractor_PlayableFromReset は**呼び出し元から**盤面を回す。
//
// ドメインの単体試験は DeclareContract を直接叩いていたので全部緑だったが、
// インタラクタは契約フェーズを進める術を持っておらず、Reset 直後の
// 契約フェーズから誰も抜け出せなかった。つまり CUI も Web も最初の手番で
// 固まったまま出荷されるところだった。この試験はその経路を通す。
func TestBauernschnapsenInteractor_PlayableFromReset(t *testing.T) {
	g := domain.NewDefaultBauernschnapsen()
	gi := usecase.NewBauernschnapsenInteractor(g, &presenter.BauernschnapsenWebPresenter{})

	out := parseBauernschnapsenOut(t, gi.Reset())

	// Reset は CPU の宣言を進め、人間の宣言待ちで止まる —— もしくは
	// 人間より後ろの席しか残っていなければプレイフェーズまで進む。
	require.Contains(t,
		[]int{int(domain.BauernschnapsenPhaseContract), int(domain.BauernschnapsenPhasePlay)},
		out.Phase, "Reset 後に契約フェーズかプレイフェーズのどちらかにいること")

	if out.Phase == int(domain.BauernschnapsenPhaseContract) {
		assert.Equal(t, 0, out.CurrentPlayerIdx, "止まるのは人間の手番でだけ")
		out = parseBauernschnapsenOut(t,
			gi.DeclareContract(domain.BauernschnapsenContractNone, domain.BauernschnapsenNoTrump))
	}

	// 契約が確定し、プレイフェーズに入っていること。
	require.Equal(t, int(domain.BauernschnapsenPhasePlay), out.Phase,
		"宣言のあとはプレイフェーズへ進むこと")
	assert.NotEqual(t, int(domain.BauernschnapsenContractNone), out.Contract,
		"誰も宣言しなくても既定の契約が入ること")
	assert.GreaterOrEqual(t, out.DeclarerIdx, 0, "宣言者が確定していること")
	if domain.BauernschnapsenContract(out.Contract) != domain.BauernschnapsenContractBettel {
		assert.NotEqual(t, domain.BauernschnapsenNoTrump, out.TrumpSuit,
			"Bettel 以外は切り札が決まっていること")
	}

	// **1 枚も出せずに終わっていないこと。** 人間の手番まで来ているなら
	// 出せる札が必ずあり、実際に出せる。
	require.Equal(t, 0, out.CurrentPlayerIdx, "人間の手番で止まっていること")
	require.NotEmpty(t, out.ValidPlayIndices, "出せる札があること")
	out = parseBauernschnapsenOut(t, gi.Play(out.ValidPlayIndices[0]))
	assert.NotEqual(t, int(domain.BauernschnapsenPhaseContract), out.Phase,
		"プレイ後に契約フェーズへ戻らないこと")
}

// TestBauernschnapsenInteractor_RoundCompletes は 1 ラウンドを最後まで回す。
//
// 契約フェーズはラウンドごとに始まるので、2 ラウンド目以降で固まらないことも
// ここで見る (NextRound も契約を進めなければならない)。
func TestBauernschnapsenInteractor_RoundCompletes(t *testing.T) {
	g := domain.NewDefaultBauernschnapsen()
	gi := usecase.NewBauernschnapsenInteractor(g, &presenter.BauernschnapsenWebPresenter{})
	out := parseBauernschnapsenOut(t, gi.Reset())

	rounds, plays := 0, 0
	// 上限は「20 枚 ÷ 4 席 = 5 トリック」× 数ラウンド分の十分な回数。
	for i := 0; i < 400 && !out.GameEndFlag && rounds < 2; i++ {
		switch domain.BauernschnapsenPhase(out.Phase) {
		case domain.BauernschnapsenPhaseContract:
			out = parseBauernschnapsenOut(t,
				gi.DeclareContract(domain.BauernschnapsenContractNone, domain.BauernschnapsenNoTrump))
		case domain.BauernschnapsenPhasePlay:
			require.NotEmpty(t, out.ValidPlayIndices, "人間の手番なら出せる札があること")
			out = parseBauernschnapsenOut(t, gi.Play(out.ValidPlayIndices[0]))
			plays++
		case domain.BauernschnapsenPhaseTrickEnd:
			out = parseBauernschnapsenOut(t, gi.NextTrick())
		case domain.BauernschnapsenPhaseRoundEnd:
			out = parseBauernschnapsenOut(t, gi.NextRound())
			rounds++
		default:
			t.Fatalf("進めないフェーズ %d で止まった", out.Phase)
		}
	}

	// **どれだけ働いたかを主張する。** 0 手で緑になる試験は何も見ていない。
	assert.GreaterOrEqual(t, rounds, 2, "2 ラウンド進むこと (2 回目の契約フェーズも抜けられること)")
	assert.GreaterOrEqual(t, plays, 10, "人間が 1 ラウンド 5 枚 × 2 ラウンド出せていること")
}
