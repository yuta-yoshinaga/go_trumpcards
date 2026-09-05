//go:build test

package usecase_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func TestMarjapussiReachability_FullRound(t *testing.T) {
	g := domain.NewDefaultMarjapussi()
	tp := new(presenter.MockMarjapussiPresenter)
	tp.On("Output", mock.Anything, mock.Anything).Return(marjapussiMockOutput)

	ti := usecase.NewMarjapussiInteractor(g, tp)
	ti.Reset()

	const maxSteps = 200
	humanPlays := 0

	for step := 0; step < maxSteps && g.GetPhase() != domain.MarjapussiPhaseRoundEnd && !g.GetGameEndFlag(); step++ {
		switch g.GetPhase() {
		case domain.MarjapussiPhasePlay:
			if g.IsHumanTurn() {
				valid := g.GetPlayableIndices(0)
				require.NotEmpty(t, valid, "human must have playable cards during their turn")
				ti.Play(valid[0])
				humanPlays++
			}
		case domain.MarjapussiPhaseTrickEnd:
			ti.NextTrick()
		}
	}

	assert.Equal(t, domain.MarjapussiPhaseRoundEnd, g.GetPhase(), "game must reach RoundEnd")
	assert.Equal(t, domain.MarjapussiHandSize, humanPlays, "human must have played exactly all 8 cards")

	totalTricks := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		totalTricks += g.GetPlayer(i).GetTrickCount()
	}
	assert.Equal(t, domain.MarjapussiTrickCount, totalTricks, "all 8 tricks must be held by players")
	assert.Equal(t, domain.MarjapussiTrickCount, g.GetTrickNumber(), "8 tricks must be completed")

	cardPts := g.GetRoundCardPoints()
	marriagePts := g.GetRoundMarriage()
	assert.Equal(t, 120, cardPts[0]+cardPts[1], "total card points (including pussi) must be 120")

	// NextRound scores the round and starts next round.
	ti.NextRound()
	scores := g.GetTeamScores()
	assert.Equal(t, cardPts[0]+marriagePts[0], scores[0])
	assert.Equal(t, cardPts[1]+marriagePts[1], scores[1])
}

// findDealWithHumanMarriage 人間 (席 0) が結婚を宣言可能な配りを探索し、
// 人間リードの状態で返す。
func findDealWithHumanMarriage(t *testing.T) (*domain.Marjapussi, int, int, int) {
	t.Helper()
	for trial := 0; trial < 1000; trial++ {
		g := domain.NewDefaultMarjapussi()
		g.Reset()
		opts := g.GetMarriageOptions(0)
		if len(opts) > 0 {
			suit := opts[0].Suit
			pts := opts[0].Points
			human := g.GetPlayer(0)
			leadIdx := -1
			for i := 0; i < human.GetCardsSize(); i++ {
				c := human.GetCard(i)
				if c.GetDesign() == suit && (c.GetValue() == 13 || c.GetValue() == 12) {
					leadIdx = i
					break
				}
			}
			if leadIdx >= 0 {
				g.SetLeadPlayerIdx(0)
				g.SetCurrentPlayerIdx(0)
				g.SetCurrentTrick(nil)
				return g, suit, pts, leadIdx
			}
		}
	}
	t.Fatal("must find a deal with marriage available for human")
	return nil, 0, 0, -1
}

func TestMarjapussiReachability_Marriage(t *testing.T) {
	// 1. ドメインの直接呼び出し:
	// ti.Play (インタラクタ) を呼ぶと人間の 1 手の後に CPU の手番まで自動で進み、
	// トリックを取った CPU が後続トリックで結婚を宣言して切り札を上書きする場合がある。
	// そのため「人間が結婚札をリードした瞬間に切り札が更新されること」は、
	// CPU が介入しない domain.PlayerPlay の直後で確定的に検証する。
	domainGame, dSuit, dPts, dLeadIdx := findDealWithHumanMarriage(t)
	require.NoError(t, domainGame.PlayerPlay(dLeadIdx))
	assert.Equal(t, dSuit, domainGame.GetTrumpSuit(), "domain.PlayerPlay 直後: trump suit must be updated to marriage suit")
	assert.Equal(t, dPts, domainGame.GetRoundMarriage()[0], "domain.PlayerPlay 直後: team 0 must receive marriage points")

	// 2. インタラクタ経由の到達可能性:
	// ti.Play はインタラクタ経由の呼び出しであるため、人間のプレイ後に CPU の手番が自動で進む。
	// Marjapussi では、リード時に K+Q を持っていれば CPU も結婚を宣言でき、そのたびに切り札が変わる。
	// さらに、味方 (席 2) が後続トリックで結婚を宣言した場合はチーム 0 の結婚点が加算される。
	// つまり ti.Play 後の GetTrumpSuit() は人間の宣言したスートであり続けるとは限らず、
	// チーム 0 の結婚点も foundPts より大きくなりうる。
	// したがって、インタラクタ経由では後続の CPU プレイに上書きされない事実を検証する:
	// (a) アクションログに席 0 の結婚宣言 (foundSuit) が記録されていること (ログは追記のため消えない)
	// (b) チーム 0 の結婚点が少なくとも人間の宣言分 (foundPts 以上: >=) 獲得されていること
	tp := new(presenter.MockMarjapussiPresenter)
	tp.On("Output", mock.Anything, mock.Anything).Return(marjapussiMockOutput)

	foundGame, foundSuit, foundPts, leadCardIdx := findDealWithHumanMarriage(t)
	ti := usecase.NewMarjapussiInteractor(foundGame, tp)
	ti.Play(leadCardIdx)

	// 味方 (席 2) も後続で結婚を宣言できるため、foundPts との一致 (==) ではなく
	// foundPts 以上 (>=) であることを確認する。
	marriagePts := foundGame.GetRoundMarriage()
	assert.GreaterOrEqual(t, marriagePts[0], foundPts, "team 0 must receive at least the initial marriage points")

	suitNames := map[int]string{
		domain.CardDesignSpade:   "Spades",
		domain.CardDesignClover:  "Clubs",
		domain.CardDesignHeart:   "Hearts",
		domain.CardDesignDiamond: "Diamonds",
	}
	expectedSuitName := suitNames[foundSuit]
	hasMarriageLog := false
	for _, log := range foundGame.GetActionLog() {
		if log.PlayerIdx == 0 && log.ActionType == "marriage" && strings.Contains(log.Detail, expectedSuitName) {
			hasMarriageLog = true
			break
		}
	}
	assert.True(t, hasMarriageLog, "action log must record human marriage for foundSuit")
}

func TestMarjapussiReachability_CpuOverwritesTrumpByMarriage(t *testing.T) {
	// CPU が後続トリックで結婚を宣言して切り札が上書きされる経路を確定的に検証する。
	// これはゲームの正当なルール仕様であり、欠陥ではないことを示す。
	//
	// 進行手順:
	// 1. 人間 (席 0) がスペード K+Q を持ち、スペード K をリードして結婚宣言 (+20点、切り札=スペード)
	// 2. CPU 1 (席 1) がスペード A でトリック 1 を勝つ
	// 3. トリック 2 のリード手番となった CPU 1 はハート K+Q を持っており、ハート K をリードして結婚宣言
	// 4. 切り札がスペードからハートへ上書きされ、チーム 1 にも結婚点 (+20点) が加算される
	g := domain.NewDefaultMarjapussi()
	g.Reset()
	g.SetPhase(domain.MarjapussiPhasePlay)
	g.SetLeadPlayerIdx(0)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	g.SetTrumpSuit(0)

	// 各プレイヤーの手札を確定的に構築
	p0 := g.GetPlayer(0)
	p0.Reset()
	p0.AddCard(domain.NewCard(domain.CardDesignSpade, 13, false)) // K (トリック 1 結婚リード)
	p0.AddCard(domain.NewCard(domain.CardDesignSpade, 12, false)) // Q
	p0.AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))

	p1 := g.GetPlayer(1)
	p1.Reset()
	p1.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))  // A (トリック 1 を勝つ)
	p1.AddCard(domain.NewCard(domain.CardDesignHeart, 13, false)) // K (トリック 2 で結婚リード)
	p1.AddCard(domain.NewCard(domain.CardDesignHeart, 12, false)) // Q

	p2 := g.GetPlayer(2)
	p2.Reset()
	p2.AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	p2.AddCard(domain.NewCard(domain.CardDesignClover, 6, false))
	p2.AddCard(domain.NewCard(domain.CardDesignClover, 7, false))

	p3 := g.GetPlayer(3)
	p3.Reset()
	p3.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
	p3.AddCard(domain.NewCard(domain.CardDesignDiamond, 8, false))
	p3.AddCard(domain.NewCard(domain.CardDesignDiamond, 9, false))

	tp := new(presenter.MockMarjapussiPresenter)
	tp.On("Output", mock.Anything, mock.Anything).Return(marjapussiMockOutput)

	ti := usecase.NewMarjapussiInteractor(g, tp)
	// 人間がスペード K (index 0) をプレイ
	ti.Play(0)

	// トリック 1 で人間がスペード結婚宣言後、席 1 がスペード A で勝ち、
	// トリック 2 で席 1 (CPU) がハート結婚を宣言して切り札をハートに上書きしたことを確認。
	assert.Equal(t, domain.CardDesignHeart, g.GetTrumpSuit(), "CPU の後続結婚宣言により切り札がハートに上書きされること")

	marriagePts := g.GetRoundMarriage()
	assert.Equal(t, 20, marriagePts[0], "チーム 0 には人間のスペード結婚点 (20点) が記録されていること")
	assert.Equal(t, 20, marriagePts[1], "チーム 1 には CPU 1 のハート結婚点 (20点) が記録されていること")

	humanMarriageFound := false
	cpuMarriageFound := false
	for _, log := range g.GetActionLog() {
		if log.ActionType == "marriage" {
			if log.PlayerIdx == 0 && strings.Contains(log.Detail, "Spades") {
				humanMarriageFound = true
			}
			if log.PlayerIdx == 1 && strings.Contains(log.Detail, "Hearts") {
				cpuMarriageFound = true
			}
		}
	}
	assert.True(t, humanMarriageFound, "人間のスペード結婚宣言がログに残っていること")
	assert.True(t, cpuMarriageFound, "CPU 1 のハート結婚宣言がログに残っていること")
}
