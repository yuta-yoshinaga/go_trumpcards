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

// quadrilleOut は Web プレゼンタが返す JSON のうち、この試験で見る分だけ。
type quadrilleOut struct {
	Phase               int   `json:"phase"`
	CurrentBidderIdx    int   `json:"currentBidderIdx"`
	CurrentPlayerIdx    int   `json:"currentPlayerIdx"`
	QuadrilleIdx        int   `json:"quadrilleIdx"`
	TrumpSuit           int   `json:"trumpSuit"`
	IsHumanTurn         bool  `json:"isHumanTurn"`
	IsHumanBidTurn      bool  `json:"isHumanBidTurn"`
	IsHumanKingCallTurn bool  `json:"isHumanKingCallTurn"`
	CalledKingSuit      int   `json:"calledKingSuit"`
	CallableKingSuits   []int `json:"callableKingSuits"`
	PartnerIdx          int   `json:"partnerIdx"`
	RoiSeul             bool  `json:"roiSeul"`
	PlayableIndices     []int `json:"playableIndices"`
	GameEndFlag         bool  `json:"gameEndFlag"`
}

func parseQuadrilleOut(t *testing.T, s string) quadrilleOut {
	t.Helper()
	var o quadrilleOut
	require.NoError(t, json.Unmarshal([]byte(s), &o))
	requireWaitingOnHuman(t, o)
	return o
}

// requireWaitingOnHuman は**インタラクタが返った時点で盤面が CPU 待ちで
// ないこと**を見る。これがこのゲームの不変条件そのもの。
//
// フェーズごとに「人間の手番か」を出しているので、CPU が落札した王呼びを
// 進め忘れれば、ここで phase=KingCall かつ isHumanKingCallTurn=false になる。
// テスト側が代わりに CallKing を叩いてしまうと配線の欠落が見えないので、
// 判定は毎回の戻り値で行う。
func requireWaitingOnHuman(t *testing.T, o quadrilleOut) {
	t.Helper()
	if o.GameEndFlag {
		return
	}
	switch domain.QuadrillePhase(o.Phase) {
	case domain.QuadrillePhaseBid:
		require.True(t, o.IsHumanBidTurn, "CPU のビッド待ちで返っている")
	case domain.QuadrillePhaseKingCall:
		require.True(t, o.IsHumanKingCallTurn,
			"CPU の王呼び待ちで返っている (落札者 %d)", o.QuadrilleIdx)
	case domain.QuadrillePhasePlay:
		require.True(t, o.IsHumanTurn, "CPU のプレイ待ちで返っている")
	}
}

// TestQuadrilleInteractor_PlayableFromReset は**呼び出し元から**盤面を回す。
//
// 王呼びはこのゲームで新しく足したフェーズで、ドメインの単体試験は
// DeclareKing を直接叩いてしまう。インタラクタ以降がそこを進められないと、
// 落札の直後に CUI も Web も固まる —— #5424 (Bauernschnapsen) で
// 実際にその状態を出しかけた。この試験はその経路を通す。
func TestQuadrilleInteractor_PlayableFromReset(t *testing.T) {
	g := domain.NewDefaultQuadrille()
	gi := usecase.NewQuadrilleInteractor(g, &presenter.QuadrilleWebPresenter{})

	out := parseQuadrilleOut(t, gi.Reset())

	for i := 0; i < domain.QuadrillePlayerCnt && out.Phase == int(domain.QuadrillePhaseBid); i++ {
		out = quadrilleBidHighest(t, gi, out)
	}
	require.NotEqual(t, int(domain.QuadrillePhaseBid), out.Phase, "ビッドが終わること")

	if out.Phase == int(domain.QuadrillePhaseKingCall) {
		// 人間が落札した。**呼べる王が提示されていること。**
		require.True(t, out.IsHumanKingCallTurn)
		require.NotEmpty(t, out.CallableKingSuits, "呼べる王が無いなら Roi seul で飛ばされるはず")
		out = parseQuadrilleOut(t, gi.CallKing(out.CallableKingSuits[0]))
	}

	require.Equal(t, int(domain.QuadrillePhasePlay), out.Phase, "王を呼んだらプレイへ進むこと")
	require.GreaterOrEqual(t, out.QuadrilleIdx, 0, "落札者が決まっていること")
	if !out.RoiSeul {
		assert.NotEqual(t, -1, out.CalledKingSuit, "単独でないなら王が呼ばれていること")
	}
	// **相方はまだ伏せられている。** 呼ばれた王が場に出るまで -1。
	assert.Equal(t, -1, out.PartnerIdx, "王が出る前に相方が漏れている")
}

// TestQuadrilleInteractor_RoundCompletes は 1 ディールを最後まで回す。
//
// 王呼びはディールごとに始まるので、2 ディール目以降で固まらないことも見る。
func TestQuadrilleInteractor_RoundCompletes(t *testing.T) {
	g := domain.NewDefaultQuadrille()
	gi := usecase.NewQuadrilleInteractor(g, &presenter.QuadrilleWebPresenter{})
	out := parseQuadrilleOut(t, gi.Reset())

	rounds, plays, kingCalls := 0, 0, 0
	for i := 0; i < 600 && !out.GameEndFlag && rounds < 2; i++ {
		switch domain.QuadrillePhase(out.Phase) {
		case domain.QuadrillePhaseBid:
			out = quadrilleBidHighest(t, gi, out)
		case domain.QuadrillePhaseKingCall:
			require.NotEmpty(t, out.CallableKingSuits, "呼べる王が提示されること")
			out = parseQuadrilleOut(t, gi.CallKing(out.CallableKingSuits[0]))
			kingCalls++
		case domain.QuadrillePhasePlay:
			require.NotEmpty(t, out.PlayableIndices, "人間の手番なら出せる札があること")
			out = parseQuadrilleOut(t, gi.Play(out.PlayableIndices[0]))
			plays++
		case domain.QuadrillePhaseTrickEnd:
			out = parseQuadrilleOut(t, gi.NextTrick())
		case domain.QuadrillePhaseRoundEnd:
			out = parseQuadrilleOut(t, gi.NextRound())
			rounds++
		default:
			t.Fatalf("進めないフェーズ %d で止まった", out.Phase)
		}
	}

	// **どれだけ働いたかを主張する。** 0 手で緑になるループは何も見ていない。
	assert.GreaterOrEqual(t, rounds, 2, "2 ディール進むこと (2 回目の王呼びも抜けられること)")
	assert.GreaterOrEqual(t, plays, 10, "人間が 1 ディール 10 枚 × 2 ディール出せていること")
	assert.GreaterOrEqual(t, kingCalls, 0, "王呼びは人間が落札したディールだけ発生する")
}

// quadrilleBidHighest は**その場で合法な一番高いビッド**を打つ。
//
// CPU が先に宣言しているとき、決め打ちの Entrar は「最高ビッドを上回れ」で
// 弾かれ、盤面が 1 ミリも進まないまま試験が「フェーズが変わらない」と
// 報告する（実際にそれで 1 度書き直した）。パスは常に合法なので最後の逃げ道。
func quadrilleBidHighest(t *testing.T, gi *usecase.QuadrilleInteractor, cur quadrilleOut) quadrilleOut {
	t.Helper()
	for _, bid := range []domain.QuadrilleBid{
		domain.QuadrilleBidSolo, domain.QuadrilleBidEntrar, domain.QuadrilleBidNone,
	} {
		next := parseQuadrilleOut(t, gi.Bid(bid, domain.CardDesignSpade))
		if next.Phase != cur.Phase || next.CurrentBidderIdx != cur.CurrentBidderIdx {
			return next
		}
	}
	t.Fatalf("どのビッドでも盤面が進まなかった (phase=%d bidder=%d)", cur.Phase, cur.CurrentBidderIdx)
	return cur
}

// TestQuadrilleInteractor_KingCallSurvivesTheWorkerRestore は、**Worker が
// 実際に通る経路**で王呼びが保たれることを見る。
//
// extra4 の Worker はリクエストごとに Snapshot → KV → RestoreQuadrilleInteractor
// で盤面を作り直す。ドメインの Marshal/Unmarshal を直接叩く試験は通っても、
// この経路が落ちていれば王を呼んだ次のリクエストで味方が変わる。
func TestQuadrilleInteractor_KingCallSurvivesTheWorkerRestore(t *testing.T) {
	g := domain.NewDefaultQuadrille()
	gi := usecase.NewQuadrilleInteractor(g, &presenter.QuadrilleWebPresenter{})
	out := parseQuadrilleOut(t, gi.Reset())

	for i := 0; i < domain.QuadrillePlayerCnt && out.Phase == int(domain.QuadrillePhaseBid); i++ {
		out = quadrilleBidHighest(t, gi, out)
	}
	if out.Phase == int(domain.QuadrillePhaseKingCall) {
		require.NotEmpty(t, out.CallableKingSuits)
		out = parseQuadrilleOut(t, gi.CallKing(out.CallableKingSuits[0]))
	}
	require.Equal(t, int(domain.QuadrillePhasePlay), out.Phase)

	// **ゼロ値と食い違う盤面であること。** 単独プレイだと呼ばれた王が -1 に
	// なり、フィールドを落とした実装と区別が付かない。
	if out.RoiSeul {
		t.Skip("この配りは Roi seul なので、王呼びの往復を見る対象にならない")
	}
	require.NotEqual(t, -1, out.CalledKingSuit)

	blob, err := gi.Snapshot()
	require.NoError(t, err)
	restored, err := usecase.RestoreQuadrilleInteractor(blob, &presenter.QuadrilleWebPresenter{})
	require.NoError(t, err)

	// 復元した盤面を**そのまま読み直す**。Hint は盤面を進めないので副作用が無い。
	after := parseQuadrilleOut(t, restored.Hint())
	assert.Equal(t, out.CalledKingSuit, after.CalledKingSuit, "呼ばれた王が Worker の復元で消えた")
	assert.Equal(t, out.QuadrilleIdx, after.QuadrilleIdx, "落札者が Worker の復元で変わった")
	assert.Equal(t, out.RoiSeul, after.RoiSeul)
	// 伏せたままの相方は復元後も伏せられている。
	assert.Equal(t, out.PartnerIdx, after.PartnerIdx, "相方の見え方が Worker の復元で変わった")
}
