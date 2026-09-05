//go:build test

package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func TestBinokelReachability_HumanWinsBid(t *testing.T) {
	game := domain.NewDefaultBinokel()
	ppMock := new(presenter.MockBinokelPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return(`{"ok":true}`)

	pi := usecase.NewBinokelInteractor(game, ppMock)
	_ = pi.Reset()

	require.Equal(t, domain.BinokelPhaseBid, game.GetPhase())
	require.True(t, game.IsHumanBidTurn())

	// 人間が 500 点でビッド (CPU は降りる)
	_ = pi.Bid(500)

	// 落札者が人間(0)になり、Dabb フェーズに遷移
	require.Equal(t, 0, game.GetHighestBidder())
	require.Equal(t, 500, game.GetHighestBid())
	require.Equal(t, domain.BinokelPhaseDabb, game.GetPhase())
	require.True(t, game.IsHumanDabbTurn())
	assert.Len(t, game.GetDabb(), domain.BinokelDabbSize)
	assert.Equal(t, domain.BinokelHandSize+domain.BinokelDabbSize, game.GetPlayer(0).GetCardsSize())

	// 人間が Dabb へ 3 枚捨てる
	_ = pi.DiscardToDabb([]int{0, 1, 2})

	// トランプ宣言フェーズへ遷移し、手札が 15 枚に戻る
	require.Equal(t, domain.BinokelPhaseTrump, game.GetPhase())
	require.True(t, game.IsHumanTurn())
	assert.Equal(t, domain.BinokelHandSize, game.GetPlayer(0).GetCardsSize())
	assert.Len(t, game.GetDabbDiscarded(), domain.BinokelDabbSize)

	// 人間がトランプを宣言 (スペード)
	_ = pi.CallTrump(domain.CardDesignSpade)

	// メルドフェーズへ遷移
	require.Equal(t, domain.BinokelPhaseMeld, game.GetPhase())
	require.Equal(t, domain.CardDesignSpade, game.GetTrumpSuit())
	assert.GreaterOrEqual(t, game.GetPlayer(0).GetMeldScore(), 0)

	// メルド確認
	_ = pi.ConfirmMelds()

	// プレイフェーズへ遷移 (落札者がリード)
	require.Equal(t, domain.BinokelPhasePlay, game.GetPhase())
	require.Equal(t, 0, game.GetLeadPlayerIdx())
	require.Equal(t, 0, game.GetCurrentPlayerIdx())
	require.True(t, game.IsHumanTurn())

	// プレイ可能なカードを 1 枚出す
	validIndices := game.GetValidPlayIndices(0)
	require.NotEmpty(t, validIndices)
	_ = pi.Play(validIndices[0])

	// 人間がプレイした後、CPU ターンも自動実行される (人間は1枚消費、CPUも1枚以上消費)
	assert.Equal(t, domain.BinokelHandSize-1, game.GetPlayer(0).GetCardsSize())
	assert.LessOrEqual(t, game.GetPlayer(1).GetCardsSize(), domain.BinokelHandSize-1)
	assert.LessOrEqual(t, game.GetPlayer(2).GetCardsSize(), domain.BinokelHandSize-1)
}

func TestBinokelReachability_CpuWinsBid(t *testing.T) {
	ppMock := new(presenter.MockBinokelPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return(`{"ok":true}`)

	var (
		game *domain.Binokel
		pi   *usecase.BinokelInteractor
	)

	// CPU がビッドを行う配りが出るまで配り直す
	cpuWonBid := false
	for attempt := 0; attempt < 50; attempt++ {
		game = domain.NewDefaultBinokel()
		pi = usecase.NewBinokelInteractor(game, ppMock)
		_ = pi.Reset()

		if game.GetHighestBid() > 0 && game.GetHighestBidder() != 0 {
			cpuWonBid = true
			break
		}
	}
	require.True(t, cpuWonBid, "CPU should have placed a bid in 50 attempts")

	// 人間はパスする → CPU が落札
	_ = pi.Pass()

	// CPU が落札した場合、runCpuBids が自動で Dabb 捨てとトランプ宣言を完了して Meld フェーズに進む
	require.NotEqual(t, 0, game.GetHighestBidder())
	require.Equal(t, domain.BinokelPhaseMeld, game.GetPhase())
	require.Len(t, game.GetDabbDiscarded(), domain.BinokelDabbSize)
	require.NotZero(t, game.GetTrumpSuit())

	// 人間がメルドを確認
	_ = pi.ConfirmMelds()

	// プレイフェーズへ遷移 (CPU 落札者がリードし、人間のターンまで自動進行)
	require.Equal(t, domain.BinokelPhasePlay, game.GetPhase())
	require.True(t, game.IsHumanTurn())

	// 人間がカードを 1 枚プレイ
	validIndices := game.GetValidPlayIndices(0)
	require.NotEmpty(t, validIndices)
	_ = pi.Play(validIndices[0])

	assert.Equal(t, domain.BinokelHandSize-1, game.GetPlayer(0).GetCardsSize())
}
