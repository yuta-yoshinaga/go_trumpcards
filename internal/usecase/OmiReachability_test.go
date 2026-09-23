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

// playOneRoundPlaysHumanAndAdvancesTricks は1ラウンド分のトリック（8トリック）を
// 人間が1枚ずつプレイし、トリック終了を進めてラウンド終了まで完走させるヘルパー。
// 人間がプレイした枚数を返す。
func playOneRoundPlaysHumanAndAdvancesTricks(t *testing.T, omi *domain.Omi, interactor *usecase.OmiInteractor) int {
	t.Helper()
	humanPlays := 0

	for i := 0; i < 200 && omi.GetPhase() != domain.OmiPhaseRoundEnd && !omi.GetGameEndFlag(); i++ {
		switch omi.GetPhase() {
		case domain.OmiPhaseCallTrump:
			if omi.IsHumanCallTrumpTurn() {
				interactor.CallTrump(domain.CardDesignSpade)
			}
		case domain.OmiPhasePlay:
			require.True(t, omi.IsHumanTurn(), "must be human turn when in play phase")
			validPlays := omi.GetValidPlayIndices(0)
			require.NotEmpty(t, validPlays, "human must have valid plays")
			interactor.Play(validPlays[0])
			humanPlays++
		case domain.OmiPhaseTrickEnd:
			interactor.NextTrick()
		default:
			t.Fatalf("unexpected phase: %v", omi.GetPhase())
		}
	}

	return humanPlays
}

func TestOmiReachability_CpuCallerRound(t *testing.T) {
	// ディーラー0（人間）→ 指名者1（CPU）の配り
	// Reset() から始めて、CPUが自動宣言し、第2段階配布後、第1トリックで人間の手番まで到達し、
	// 人間が各トリック1枚ずつプレイして全8トリック完走・ラウンド終了まで進むことを検証する。
	omi := domain.NewDefaultOmi()
	cfg := omi.GetConfig()
	cfg.PointLimit = 100 // 途中でゲーム終了しないように十分な上限を設定
	omi.SetConfig(cfg)

	interactor := usecase.NewOmiInteractor(omi, new(presenter.OmiCuiPresenter))
	interactor.Reset()

	// ディーラーは席0、切り札指名者は席1 (CPU)
	require.Equal(t, 0, omi.GetDealerIdx())
	require.Equal(t, 1, omi.GetTrumpCallerIdx())

	// CPUが切り札宣言を完了し、第2段階配布も終わってプレイフェーズに入っていること
	require.Equal(t, domain.OmiPhasePlay, omi.GetPhase())
	require.NotZero(t, omi.GetTrumpSuit(), "trump suit must be declared")

	// 第1トリックのリード(席1)から席2, 席3と自動プレイが進み、人間(席0)の手番で止まっていること
	require.True(t, omi.IsHumanTurn())
	require.Len(t, omi.GetCurrentTrick(), 3, "seats 1, 2, 3 should have played before seat 0")
	require.Equal(t, 8, omi.GetPlayer(0).GetCardsSize(), "human hand should have 8 cards after deal 2")

	// 8トリック完走させる
	humanPlays := playOneRoundPlaysHumanAndAdvancesTricks(t, omi, interactor)

	// アサーション: 人間が8枚プレイしたこと、8トリックが完了してラウンド終了フェーズに到達したこと
	assert.Equal(t, 8, humanPlays, "human must play exactly 8 cards")
	assert.Equal(t, domain.OmiPhaseRoundEnd, omi.GetPhase(), "must reach round end phase")

	totalTricks := 0
	for i := 0; i < omi.GetPlayerCnt(); i++ {
		totalTricks += omi.GetPlayer(i).GetTrickCount()
	}
	assert.Equal(t, 8, totalTricks, "total tricks won across all players must be exactly 8")
}

func TestOmiReachability_HumanCallerRound(t *testing.T) {
	// ディーラー3（CPU）→ 指名者0（人間）の配り
	// ラウンドを回してディーラーを3にする (またはReset後にNextRoundを3回進める)
	omi := domain.NewDefaultOmi()
	cfg := omi.GetConfig()
	cfg.PointLimit = 100
	omi.SetConfig(cfg)

	interactor := usecase.NewOmiInteractor(omi, new(presenter.OmiCuiPresenter))
	interactor.Reset()

	// ラウンドを進めてディーラーが席3になるまで回す
	for omi.GetDealerIdx() != 3 {
		// ラウンド終了状態にするために各ラウンドを完走
		humanPlays := playOneRoundPlaysHumanAndAdvancesTricks(t, omi, interactor)
		require.Equal(t, 8, humanPlays)
		require.Equal(t, domain.OmiPhaseRoundEnd, omi.GetPhase())

		// 次のラウンドへ
		interactor.NextRound()
	}

	// ディーラーが席3、指名者が席0 (人間！) であることを確認
	require.Equal(t, 3, omi.GetDealerIdx())
	require.Equal(t, 0, omi.GetTrumpCallerIdx())

	// 人間が指名者のため、CallTrumpフェーズで停止していること
	require.Equal(t, domain.OmiPhaseCallTrump, omi.GetPhase(), "must be in CallTrump phase waiting for human")
	require.True(t, omi.IsHumanCallTrumpTurn(), "must be human's call turn")
	require.Equal(t, 4, omi.GetPlayer(0).GetCardsSize(), "human hand should have 4 cards before trump call")

	// 人間が切り札を宣言する (例: スペード)
	interactor.CallTrump(domain.CardDesignSpade)

	// 切り札確定後、第2段階配布が行われ、手札が8枚になり、プレイフェーズへ
	require.Equal(t, domain.CardDesignSpade, omi.GetTrumpSuit())
	require.Equal(t, domain.OmiPhasePlay, omi.GetPhase())
	require.Equal(t, 8, omi.GetPlayer(0).GetCardsSize(), "human hand should have 8 cards after deal 2")

	// 人間が指名者のため、第1トリックのリードは人間(席0)
	require.Equal(t, 0, omi.GetCurrentPlayerIdx())
	require.True(t, omi.IsHumanTurn())
	require.Empty(t, omi.GetCurrentTrick(), "current trick must be empty at lead")

	// 8トリック完走させる
	humanPlays := playOneRoundPlaysHumanAndAdvancesTricks(t, omi, interactor)

	// アサーション: 人間が8枚プレイしたこと、8トリック完了したこと
	assert.Equal(t, 8, humanPlays, "human must play exactly 8 cards")
	assert.Equal(t, domain.OmiPhaseRoundEnd, omi.GetPhase(), "must reach round end phase")

	totalTricks := 0
	for i := 0; i < omi.GetPlayerCnt(); i++ {
		totalTricks += omi.GetPlayer(i).GetTrickCount()
	}
	assert.Equal(t, 8, totalTricks, "total tricks won across all players must be exactly 8")
}
