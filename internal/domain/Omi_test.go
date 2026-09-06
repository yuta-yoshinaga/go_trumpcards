//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestOmi() *domain.Omi {
	players := []*domain.OmiPlayer{
		domain.NewOmiPlayer(true, 0),  // Player 0: human, team 0
		domain.NewOmiPlayer(false, 1), // Player 1: CPU, team 1
		domain.NewOmiPlayer(false, 0), // Player 2: CPU, team 0
		domain.NewOmiPlayer(false, 1), // Player 3: CPU, team 1
	}
	return domain.NewOmi(domain.NewTrumpCards32(), players, domain.DefaultOmiConfig())
}

func setupOmiHand(e *domain.Omi, playerIdx int, cards []*domain.Card) {
	p := e.GetPlayer(playerIdx)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// --- 1. OmiConfig ---

func TestOmiConfig_Default(t *testing.T) {
	cfg := domain.DefaultOmiConfig()
	assert.Equal(t, domain.OmiCpuDifficultyNormal, cfg.CpuDifficulty)
	assert.Equal(t, 10, cfg.PointLimit)
	assert.NoError(t, cfg.Validate())
}

func TestOmiConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     domain.OmiConfig
		wantErr bool
	}{
		{"valid default", domain.DefaultOmiConfig(), false},
		{"valid easy", domain.OmiConfig{CpuDifficulty: domain.OmiCpuDifficultyEasy, PointLimit: 5}, false},
		{"invalid difficulty", domain.OmiConfig{CpuDifficulty: 5, PointLimit: 10}, true},
		{"invalid point limit", domain.OmiConfig{CpuDifficulty: domain.OmiCpuDifficultyNormal, PointLimit: 0}, true},
		{"negative point limit", domain.OmiConfig{CpuDifficulty: domain.OmiCpuDifficultyNormal, PointLimit: -1}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// --- 2. OmiPlayer ---

func TestOmiPlayer(t *testing.T) {
	p := domain.NewOmiPlayer(true, 0)
	assert.True(t, p.GetIsHuman())
	assert.Equal(t, 0, p.GetTeam())
	assert.Equal(t, 0, p.GetTrickCount())

	p2 := domain.NewOmiPlayer(false, 1)
	assert.False(t, p2.GetIsHuman())
	assert.Equal(t, 1, p2.GetTeam())
}

func TestOmiPlayer_ResetRound(t *testing.T) {
	p := domain.NewOmiPlayer(true, 0)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
	p.AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 10, false)})
	p.SetIsFinished(true)

	p.ResetRound()
	assert.Equal(t, 0, p.GetCardsSize())
	assert.Equal(t, 0, p.GetTrickCount())
	assert.False(t, p.GetIsFinished())
}

func TestOmiPlayer_JSONRoundTrip(t *testing.T) {
	p := domain.NewOmiPlayer(true, 1)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	p.AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 7, false)})

	data, err := json.Marshal(p)
	require.NoError(t, err)

	p2 := new(domain.OmiPlayer)
	err = json.Unmarshal(data, p2)
	require.NoError(t, err)

	assert.Equal(t, p.GetIsHuman(), p2.GetIsHuman())
	assert.Equal(t, p.GetTeam(), p2.GetTeam())
	assert.Equal(t, 1, p2.GetCardsSize())
	assert.Equal(t, 1, p2.GetTrickCount())
}

// --- 3. デッキと配り (2段階配り & 32枚) ---

func TestOmi_ConstantsConsistency(t *testing.T) {
	// 定数どうしの整合性検証:
	// ドメイン定数が Omi の仕様 (32枚デッキ、4人プレイ、各人8枚手札、初期4枚配り、2チーム)
	// と正しく整合していることを確認する。
	// ※ここは定数定義そのものおよび定数間の乗除関係を明示的に担保するためのテスト。
	assert.Equal(t, 4, domain.OmiPlayerCnt, "OmiPlayerCnt constant must be 4")
	assert.Equal(t, 8, domain.OmiHandSize, "OmiHandSize constant must be 8")
	assert.Equal(t, 4, domain.OmiInitialDealSize, "OmiInitialDealSize constant must be 4")
	assert.Equal(t, 2, domain.OmiTeamCnt, "OmiTeamCnt constant must be 2")
	assert.Equal(t, 32, domain.OmiHandSize*domain.OmiPlayerCnt,
		"OmiHandSize * OmiPlayerCnt must strictly equal 32 deck size")
	assert.Equal(t, domain.OmiHandSize, domain.OmiInitialDealSize*2,
		"OmiInitialDealSize * 2 must equal OmiHandSize (two-stage deal of 4 + 4 cards)")
}

func TestOmi_DeckAndDealing_TwoStages(t *testing.T) {
	// 1. デッキと配り: 32枚、各自8枚。4枚配る → 切り札宣言 → 残り4枚の順であること。
	// (宣言時点で全員4枚しか持っていないこと、宣言後に残り4枚が配られて全員8枚になることをassertする)
	game := domain.NewDefaultOmi()
	require.Equal(t, 4, game.GetPlayerCnt(), "Omi is played by strictly 4 players")

	game.Reset()

	// 第1段階直後: フェーズは CallTrump
	assert.Equal(t, domain.OmiPhaseCallTrump, game.GetPhase())
	// 指名者は親の右隣 (dealer=0 なので caller=1)
	assert.Equal(t, 0, game.GetDealerIdx())
	assert.Equal(t, 1, game.GetTrumpCallerIdx())
	assert.Equal(t, 1, game.GetCurrentPlayerIdx())

	// 【自己成就防止のためのリテラル 4 使用】
	// 切り札宣言時点で各自に厳密に 4 枚配られていることは、Omi の仕様 (ルール) そのものである。
	// ここで domain.OmiInitialDealSize 定数を期待値に使うと、定数を誤って8に書き換えるなどの変異が生じた際に
	// 期待値も一緒に 8 に変化してテストが通過してしまう (自己成就)。
	// そのため定数ではなく仕様値のリテラル 4 を直接アサートする。
	firstDealSizes := make([]int, 4)
	for i := 0; i < 4; i++ {
		firstDealSizes[i] = game.GetPlayer(i).GetCardsSize()
		assert.Equal(t, 4, firstDealSizes[i],
			"Player %d must have exactly 4 cards at trump call stage (first deal)", i)
	}

	// 切り札宣言 (caller=1 は CPU)
	game.CpuCallTrump()

	// 第2段階直後: フェーズは Play
	assert.Equal(t, domain.OmiPhasePlay, game.GetPhase())
	assert.NotZero(t, game.GetTrumpSuit())
	assert.Equal(t, 1, game.GetTrickNumber())

	// 【第2段階配り検証: 宣言前 4 枚 → 宣言後 8 枚 (+4 枚)】
	// 全員に残り 4 枚が追加配分され、手札が厳密に 8 枚 (仕様値) となっていること、
	// および第1段階(4枚)から第2段階(8枚)への増分が厳密に +4 枚であることを assert。
	// (宣言前4枚と宣言後8枚の両方を検証することで、「最初から8枚配る」変異も
	// 「宣言後に0枚足す/配りをスキップする」変異も確実に捕まえる)。
	// ここでも domain.OmiHandSize ではなく仕様値リテラル 8 を用いて自己成就を防ぐ。
	totalCards := 0
	for i := 0; i < 4; i++ {
		handSize := game.GetPlayer(i).GetCardsSize()
		assert.Equal(t, 8, handSize,
			"Player %d must have exactly 8 cards after second deal", i)
		assert.Equal(t, 4, handSize-firstDealSizes[i],
			"Player %d must receive exactly 4 additional cards in second deal", i)
		totalCards += handSize
	}
	assert.Equal(t, 32, totalCards, "32 cards must be distributed exactly across 4 players (8x4)")
}

func TestOmi_DealerRotation_And_CallerSeat(t *testing.T) {
	// 指名者はディーラーの右隣 (反時計回りの最初の席: (dealerIdx+1)%4)。
	// ラウンドごとにディーラーが回るので指名者も回る。
	game := domain.NewDefaultOmi()
	game.Reset()

	for r := 0; r < 4; r++ {
		expectedDealer := r % 4
		expectedCaller := (expectedDealer + 1) % 4
		assert.Equal(t, expectedDealer, game.GetDealerIdx(), "round %d dealer", r+1)
		assert.Equal(t, expectedCaller, game.GetTrumpCallerIdx(), "round %d caller", r+1)
		assert.Equal(t, domain.OmiPhaseCallTrump, game.GetPhase())

		// 切り札を決めてラウンド終了まで遷移させる
		if game.IsHumanCallTrumpTurn() {
			err := game.PlayerCallTrump(domain.CardDesignSpade)
			require.NoError(t, err)
		} else {
			game.CpuCallTrump()
		}
		assert.Equal(t, domain.OmiPhasePlay, game.GetPhase())

		// テスト用にラウンド終了状態にして NextRound を呼ぶ
		game.SetPhase(domain.OmiPhaseRoundEnd)
		game.NextRound()
	}
}

// --- 4. ジャック昇格が無いこと (素直なランク順 A > K > Q > J > 10 > 9 > 8 > 7) ---

func TestOmi_NoJackPromotion(t *testing.T) {
	game := domain.NewDefaultOmi()
	game.SetTrumpSuit(domain.CardDesignSpade) // 切り札: Spade

	spadeA := domain.NewCard(domain.CardDesignSpade, 1, false)   // A(14)
	spadeK := domain.NewCard(domain.CardDesignSpade, 13, false)  // K(13)
	spadeQ := domain.NewCard(domain.CardDesignSpade, 12, false)  // Q(12)
	spadeJ := domain.NewCard(domain.CardDesignSpade, 11, false)  // J(11)
	spade10 := domain.NewCard(domain.CardDesignSpade, 10, false) // 10(10)
	spade7 := domain.NewCard(domain.CardDesignSpade, 7, false)   // 7(7)

	rankA := game.CardRankPublic(spadeA)
	rankK := game.CardRankPublic(spadeK)
	rankQ := game.CardRankPublic(spadeQ)
	rankJ := game.CardRankPublic(spadeJ)
	rank10 := game.CardRankPublic(spade10)
	rank7 := game.CardRankPublic(spade7)

	// 切り札スート内でも J は昇格せず、A > K > Q > J > 10 > ... > 7
	assert.Greater(t, rankA, rankK, "Spade A must rank higher than Spade K")
	assert.Greater(t, rankK, rankQ, "Spade K must rank higher than Spade Q")
	assert.Greater(t, rankQ, rankJ, "Spade Q must rank higher than Spade J (No Right Bower!)")
	assert.Greater(t, rankJ, rank10, "Spade J must rank higher than Spade 10")
	assert.Greater(t, rank10, rank7, "Spade 10 must rank higher than Spade 7")

	// 同色スート (Spadeに対するClover) の J が切り札扱いされないこと (No Left Bower!)
	cloverJ := domain.NewCard(domain.CardDesignClover, 11, false)
	rankCloverJ := game.CardRankPublic(cloverJ)

	// Clover J は非切り札なので、最弱の切り札 Spade 7 よりも弱いこと
	assert.Less(t, rankCloverJ, rank7,
		"Clover J must NOT be treated as trump when Spade is trump (No Left Bower!)")
	assert.Equal(t, domain.CardDesignClover, game.EffectiveSuitPublic(cloverJ),
		"Clover J must have suit Clover, never Spade")
}

// --- 5. 得点計算 (5トリック 1点 / 8トリック全取り 2点 / 4-4 引き分け 0点) ---

func TestOmi_ScoreRound_ThreeOutcomes(t *testing.T) {
	// 3. 得点: 5 トリックで 1 点 / 8 トリックで 2 点 / 4-4 で両者 0 点。
	// 3 通りを同じテストで確かめる。
	t.Run("outcome 1: 5-7 tricks gives 1 point", func(t *testing.T) {
		game := newTestOmi()
		game.SetPhase(domain.OmiPhaseRoundEnd)

		// チーム0が 5 トリック、チーム1が 3 トリック
		dummyTrick := []*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)}
		for i := 0; i < 3; i++ {
			game.GetPlayer(0).AddTrick(dummyTrick) // Player 0 (Team 0): 3 tricks
		}
		for i := 0; i < 2; i++ {
			game.GetPlayer(2).AddTrick(dummyTrick) // Player 2 (Team 0): 2 tricks -> Team 0 = 5 tricks
		}
		for i := 0; i < 3; i++ {
			game.GetPlayer(1).AddTrick(dummyTrick) // Player 1 (Team 1): 3 tricks -> Team 1 = 3 tricks
		}

		game.ScoreRound()

		assert.Equal(t, 1, game.GetTeamScore(0), "Team 0 with 5 tricks should score 1 point")
		assert.Equal(t, 0, game.GetTeamScore(1), "Team 1 with 3 tricks should score 0 points")
	})

	t.Run("outcome 2: 8 tricks (Omi) gives 2 points", func(t *testing.T) {
		game := newTestOmi()
		game.SetPhase(domain.OmiPhaseRoundEnd)

		// チーム1が 8 トリック全取り (Omi)
		dummyTrick := []*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)}
		for i := 0; i < 5; i++ {
			game.GetPlayer(1).AddTrick(dummyTrick) // Player 1 (Team 1): 5 tricks
		}
		for i := 0; i < 3; i++ {
			game.GetPlayer(3).AddTrick(dummyTrick) // Player 3 (Team 1): 3 tricks -> Team 1 = 8 tricks
		}

		game.ScoreRound()

		assert.Equal(t, 0, game.GetTeamScore(0), "Team 0 with 0 tricks should score 0 points")
		assert.Equal(t, 2, game.GetTeamScore(1), "Team 1 with 8 tricks (Omi) should score 2 points")
	})

	t.Run("outcome 3: 4-4 draw gives 0 points to both teams", func(t *testing.T) {
		game := newTestOmi()
		game.SetPhase(domain.OmiPhaseRoundEnd)

		// チーム0が 4 トリック、チーム1が 4 トリック
		dummyTrick := []*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)}
		for i := 0; i < 2; i++ {
			game.GetPlayer(0).AddTrick(dummyTrick) // Team 0
			game.GetPlayer(2).AddTrick(dummyTrick) // Team 0 -> 4 tricks
			game.GetPlayer(1).AddTrick(dummyTrick) // Team 1
			game.GetPlayer(3).AddTrick(dummyTrick) // Team 1 -> 4 tricks
		}

		game.ScoreRound()

		assert.Equal(t, 0, game.GetTeamScore(0), "Team 0 with 4 tricks in 4-4 draw should score 0 points")
		assert.Equal(t, 0, game.GetTeamScore(1), "Team 1 with 4 tricks in 4-4 draw should score 0 points")
	})
}

// --- 6. フォロー規則 (マストフォロー / 切り札強制ではない) ---

func TestOmi_FollowRules(t *testing.T) {
	game := newTestOmi()
	game.SetTrumpSuit(domain.CardDesignSpade) // 切り札: Spade
	game.SetPhase(domain.OmiPhasePlay)
	game.SetCurrentPlayerIdx(0) // 人間手番

	// Player 0 手札: Heart 10, Heart K, Diamond 7, Spade 8 (切り札)
	setupOmiHand(game, 0, []*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 10, false),
		domain.NewCard(domain.CardDesignHeart, 13, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
		domain.NewCard(domain.CardDesignSpade, 8, false),
	})

	// トリックのリードカード: Heart 7
	game.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 7, false)},
	})

	// 1. リードスート (Heart) を持っているのに Diamond を出そうとするとエラー (マストフォロー)
	err := game.PlayerPlay(2) // Diamond 7
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)

	// 2. リードスート (Heart) を持っているのに切り札 (Spade) を出そうとしてもエラー
	err = game.PlayerPlay(3) // Spade 8
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)

	// 3. リードスート (Heart) を正しくフォローすれば成功
	err = game.PlayerPlay(0) // Heart 10
	assert.NoError(t, err)

	// 4. リードスートを持っていない場合: 任意に出せる (切り札強制ではない)
	// Player 0 手札から Heart をすべて除去して Diamond 7, Spade 8 のみにする
	setupOmiHand(game, 0, []*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 7, false),
		domain.NewCard(domain.CardDesignSpade, 8, false),
	})
	game.SetCurrentPlayerIdx(0)
	game.SetPhase(domain.OmiPhasePlay)
	game.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 7, false)},
	})

	validIndices := game.GetValidPlayIndices(0)
	assert.ElementsMatch(t, []int{0, 1}, validIndices,
		"When unable to follow lead suit, player can play any card (trump not forced)")

	// 非切り札の Diamond 7 を出してもルール違反にならない
	err = game.PlayerPlay(0)
	assert.NoError(t, err)
}

// --- 7. 切り札指名のバリデーションとパス不可の裁定 ---

func TestOmi_PlayerCallTrump_Validation(t *testing.T) {
	game := newTestOmi()
	game.Reset() // dealer=0, caller=1 (CPU)

	// 人間手番ではないのに人間が呼ぶと ErrNotHumanTurn
	err := game.PlayerCallTrump(domain.CardDesignSpade)
	assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
	assert.False(t, game.IsHumanCallTrumpTurn())

	// 親を 3 に設定して caller を 0 (人間) にする
	game.SetDealerIdx(3)
	game.SetTrumpCallerIdx(0)
	game.SetCurrentPlayerIdx(0)
	game.SetPhase(domain.OmiPhaseCallTrump)

	assert.True(t, game.IsHumanCallTrumpTurn())
	assert.True(t, game.IsHumanTurn())

	// 無効なスート
	err = game.PlayerCallTrump(99)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)

	// 正常なスート指名 (パスは無く、4スートから必ず選ぶ)
	err = game.PlayerCallTrump(domain.CardDesignHeart)
	assert.NoError(t, err)
	assert.Equal(t, domain.CardDesignHeart, game.GetTrumpSuit())
	assert.Equal(t, 0, game.GetTrumpCallerIdx())
	assert.Equal(t, 0, game.GetMakerTeam()) // 席0はチーム0
	assert.Equal(t, domain.OmiPhasePlay, game.GetPhase())

	// プレイフェーズ中にコールを試みると ErrWrongPhase
	err = game.PlayerCallTrump(domain.CardDesignDiamond)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

// --- 8. 統計テスト (2000試行で4項目を検証) ---

func TestOmi_Statistical2000(t *testing.T) {
	// 5. 測って確かめること (統計テストとして恒久化)
	// 2000 回の配りで測り、帯で固定すること。
	//
	// 【各項目の理論的上限と根拠】
	// 1. 全取り (Omi) が起きる割合:
	//    理論的根拠: 8トリック全取りは、宣言チームが圧倒的偏りの切り札・ハイカードを持ち、
	//    相手チームがリードフォローや切り札カットで1トリックも取れない場合に限られる。
	//    通常AI対戦下では極めて稀 (数%程度) であり、どんなに偏っても15%を超えることはない。
	//    一方、0%であれば全取り判定ロジックが死んでいる証拠となるため、
	//    判定帯を「0% より大きく 15% 以下」とする。
	// 2. 切り札を宣言した席の分布:
	//    理論的根拠: ラウンドごとにディーラーが巡回し、指名者は親の右隣 ((dealer+1)%4) に
	//    厳密に巡回する。2000試行で各席の期待値は 25.0% (500回)。
	//    大数の法則および二項分布の5σを考慮しても15%〜35%の安全帯に確実に収まる。
	// 3. 人間 (席 0) が切り札を宣言できたラウンドの割合:
	//    理論的根拠: 席0の期待値は 25.0%。人間が宣言権から締め出されるバグを防ぐため
	//    下限を 15% 以上とする。
	// 4. 1 ラウンドで両チームのトリック数の合計が必ず 8:
	//    理論的根拠: 32枚の手札を4人で8枚ずつ持ち、各トリックに1人1枚出すため、
	//    1ラウンドは厳密に 32 / 4 = 8 トリックで終了する。引き分けトリックは存在しないため、
	//    両チームの獲得トリック数の合計は厳密に 8 となる (数学的恒等式)。

	const trials = 2000

	omiCount := 0
	callerSeats := [4]int{}
	humanCalledCount := 0

	game := domain.NewDefaultOmi()

	for trial := 0; trial < trials; trial++ {
		// ラウンドごとにディーラーを巡回させる (4人プレイヤー)
		game.SetDealerIdx(trial % 4)
		game.SetRoundNumber(trial + 1)
		game.SetPhase(domain.OmiPhaseRoundEnd)
		game.SetGameEndFlag(false)
		game.NextRound()

		caller := game.GetTrumpCallerIdx()
		require.True(t, caller >= 0 && caller < 4)
		callerSeats[caller]++
		if caller == 0 {
			humanCalledCount++
		}

		// 切り札指名
		if game.IsHumanCallTrumpTurn() {
			suit := domain.CardDesignSpade
			_ = game.PlayerCallTrump(suit)
		} else {
			game.CpuCallTrump()
		}

		// 8 トリックをプレイ (仕様上 1 ラウンド厳密に 8 トリック)
		for trick := 1; trick <= 8; trick++ {
			for cardInTrick := 0; cardInTrick < 4; cardInTrick++ {
				if game.IsHumanTurn() {
					valids := game.GetValidPlayIndices(game.GetCurrentPlayerIdx())
					require.NotEmpty(t, valids)
					_ = game.PlayerPlay(valids[0])
				} else {
					game.CpuPlay()
				}
			}
			game.ResolveTrick()
			if trick < 8 {
				game.NextTrick()
			}
		}

		// 4. 1 ラウンドで両チームのトリック数の合計が必ず 8 (数学的恒等式)
		// 仕様値であるリテラル 8 と比較し、定数書き換えによる自己成就を防ぐ
		team0Tricks := game.GetPlayer(0).GetTrickCount() + game.GetPlayer(2).GetTrickCount()
		team1Tricks := game.GetPlayer(1).GetTrickCount() + game.GetPlayer(3).GetTrickCount()
		require.Equal(t, 8, team0Tricks+team1Tricks,
			"Sum of tricks in round must strictly equal 8")

		// 1. 全取り (Omi) のカウント (全 8 トリック獲得)
		if team0Tricks == 8 || team1Tricks == 8 {
			omiCount++
		}

		game.ScoreRound()
	}

	omiRate := float64(omiCount) / float64(trials) * 100.0
	humanRate := float64(humanCalledCount) / float64(trials) * 100.0
	seatRates := [4]float64{}
	for i := 0; i < 4; i++ {
		seatRates[i] = float64(callerSeats[i]) / float64(trials) * 100.0
	}

	t.Logf("=== Omi 2000-Trial Statistics ===")
	t.Logf("1. Omi (Sweep) Rate: %.2f%% (%d / %d)", omiRate, omiCount, trials)
	t.Logf("2. Caller Seat Distribution: Seat0=%.2f%%, Seat1=%.2f%%, Seat2=%.2f%%, Seat3=%.2f%%",
		seatRates[0], seatRates[1], seatRates[2], seatRates[3])
	t.Logf("3. Human (Seat 0) Caller Rate: %.2f%% (%d / %d)", humanRate, humanCalledCount, trials)
	t.Logf("4. Trick Sum Strictly 8 Verified for all %d rounds", trials)

	// 1. 全取り (Omi) が起きる割合: 0% より大きく 15% 以下
	assert.Greater(t, omiRate, 0.0, "Omi rate must be > 0%% (sweep detection alive)")
	assert.LessOrEqual(t, omiRate, 15.0, "Omi rate must be <= 15%% (CPU not trivially swept)")

	// 2. 切り札を宣言した席の分布: どの席も 15% 〜 35%
	for i := 0; i < 4; i++ {
		assert.GreaterOrEqual(t, seatRates[i], 15.0, "Seat %d caller rate must be >= 15%%", i)
		assert.LessOrEqual(t, seatRates[i], 35.0, "Seat %d caller rate must be <= 35%%", i)
	}

	// 3. 人間 (席 0) が切り札を宣言できたラウンドの割合: 15% 以上
	assert.GreaterOrEqual(t, humanRate, 15.0, "Human caller rate must be >= 15%%")
}

// --- 9. JSON 往復 (切り札・宣言者・チーム得点・トリック数) ---

func TestOmi_JSONRoundTrip(t *testing.T) {
	game := domain.NewDefaultOmi()
	game.Reset()

	// 状態をセットアップ
	game.SetTrumpSuit(domain.CardDesignHeart)
	game.SetTrumpCallerIdx(2) // 宣言者: 席2
	game.SetMakerTeam(0)
	game.SetTeamScore(0, 3)
	game.SetTeamScore(1, 1)
	game.SetTrickNumber(4)
	game.SetRoundNumber(2)
	game.SetPhase(domain.OmiPhasePlay)
	game.SetCurrentPlayerIdx(2)

	data, err := json.Marshal(game)
	require.NoError(t, err)

	restored := new(domain.Omi)
	err = json.Unmarshal(data, restored)
	require.NoError(t, err)

	// 検証: 切り札・宣言者・チーム得点・トリック数が保存されていること
	assert.Equal(t, game.GetTrumpSuit(), restored.GetTrumpSuit(), "trump suit mismatch")
	assert.Equal(t, game.GetTrumpCallerIdx(), restored.GetTrumpCallerIdx(), "declarer/caller idx mismatch")
	assert.Equal(t, game.GetTeamScore(0), restored.GetTeamScore(0), "team 0 score mismatch")
	assert.Equal(t, game.GetTeamScore(1), restored.GetTeamScore(1), "team 1 score mismatch")
	assert.Equal(t, game.GetTrickNumber(), restored.GetTrickNumber(), "trick number mismatch")
	assert.Equal(t, game.GetRoundNumber(), restored.GetRoundNumber(), "round number mismatch")
	assert.Equal(t, game.GetMakerTeam(), restored.GetMakerTeam(), "maker team mismatch")
	assert.Equal(t, game.GetPhase(), restored.GetPhase(), "phase mismatch")
}

// --- 10. ゲーム終了判定 (10点先取) ---

func TestOmi_GameEnd_PointLimit(t *testing.T) {
	game := newTestOmi()
	game.SetPhase(domain.OmiPhaseRoundEnd)
	game.SetTeamScore(0, 9)
	game.SetTeamScore(1, 4)

	// チーム0が 5 トリック獲得して +1 点 (合計 10 点到達)
	dummyTrick := []*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)}
	for i := 0; i < 5; i++ {
		game.GetPlayer(0).AddTrick(dummyTrick)
	}
	for i := 0; i < 3; i++ {
		game.GetPlayer(1).AddTrick(dummyTrick)
	}

	game.ScoreRound()

	assert.Equal(t, 10, game.GetTeamScore(0))
	assert.True(t, game.GetGameEndFlag(), "Game should end when a team reaches PointLimit (10)")
	assert.Equal(t, domain.OmiPhaseGameEnd, game.GetPhase())
	assert.Equal(t, 0, game.GetWinnerTeam(), "Team 0 should be the winner")
}

// --- 11. ヒント機能 ---

func TestOmi_GetHint(t *testing.T) {
	game := newTestOmi()
	game.Reset()

	// 1. コールフェーズで人間が指名者
	game.SetDealerIdx(3)
	game.SetTrumpCallerIdx(0)
	game.SetCurrentPlayerIdx(0)
	game.SetPhase(domain.OmiPhaseCallTrump)

	hint := game.GetHint()
	require.NotNil(t, hint)
	require.NotNil(t, hint.Suit)
	assert.True(t, *hint.Suit >= domain.CardDesignSpade && *hint.Suit <= domain.CardDesignDiamond)
	assert.Equal(t, "strategic_call", hint.Reason)

	// 2. プレイフェーズで人間手番
	game.SetPhase(domain.OmiPhasePlay)
	game.SetTrumpSuit(domain.CardDesignSpade)
	setupOmiHand(game, 0, []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false), // Spade Ace
	})

	playHint := game.GetHint()
	require.NotNil(t, playHint)
	require.NotNil(t, playHint.CardIndex)
	assert.Equal(t, 0, *playHint.CardIndex)

	// 非人間手番なら nil
	game.SetCurrentPlayerIdx(1)
	assert.Nil(t, game.GetHint())

	// トリック中 (フォロー、切り札カット、ディスカード) のヒント理由
	game.SetCurrentPlayerIdx(0)
	// a. follow_suit
	game.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 7, false)},
	})
	setupOmiHand(game, 0, []*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 10, false),
	})
	h := game.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "follow_suit", h.Reason)

	// b. trump_cut
	setupOmiHand(game, 0, []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 10, false), // Spade is trump
	})
	h = game.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "trump_cut", h.Reason)

	// c. discard_weak
	setupOmiHand(game, 0, []*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 8, false),
	})
	h = game.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "discard_weak", h.Reason)
}

// --- 12. CPU 難易度別のプレイ (Easy / Hard) ---

func TestOmi_CpuPlay_Difficulties(t *testing.T) {
	t.Run("cpu play easy", func(t *testing.T) {
		cfg := domain.OmiConfig{CpuDifficulty: domain.OmiCpuDifficultyEasy, PointLimit: 10}
		game := domain.NewOmi(domain.NewTrumpCards32(), []*domain.OmiPlayer{
			domain.NewOmiPlayer(false, 0),
			domain.NewOmiPlayer(false, 1),
			domain.NewOmiPlayer(false, 0),
			domain.NewOmiPlayer(false, 1),
		}, cfg)
		game.Reset()
		game.CpuCallTrump()
		assert.Equal(t, domain.OmiPhasePlay, game.GetPhase())

		// CPU easy プレイ実行
		curr := game.GetCurrentPlayerIdx()
		game.CpuPlay()
		assert.Equal(t, 1, len(game.GetCurrentTrick()))
		assert.Equal(t, (curr+1)%4, game.GetCurrentPlayerIdx())
	})

	t.Run("cpu play hard partner awareness and leading", func(t *testing.T) {
		cfg := domain.OmiConfig{CpuDifficulty: domain.OmiCpuDifficultyHard, PointLimit: 10}
		game := domain.NewOmi(domain.NewTrumpCards32(), []*domain.OmiPlayer{
			domain.NewOmiPlayer(false, 0),
			domain.NewOmiPlayer(false, 1),
			domain.NewOmiPlayer(false, 0),
			domain.NewOmiPlayer(false, 1),
		}, cfg)
		game.SetTrumpSuit(domain.CardDesignSpade)
		game.SetPhase(domain.OmiPhasePlay)

		// 1. リード時: 最強カードを出す
		game.SetCurrentPlayerIdx(1)
		game.SetCurrentTrick(nil)
		setupOmiHand(game, 1, []*domain.Card{
			domain.NewCard(domain.CardDesignHeart, 7, false),
			domain.NewCard(domain.CardDesignSpade, 1, false), // Spade Ace (rank 214)
		})
		game.CpuPlay()
		require.Equal(t, 1, len(game.GetCurrentTrick()))
		assert.Equal(t, 1, game.GetCurrentTrick()[0].Card.GetValue()) // Ace

		// 2. パートナーが勝っていて自分が最後番 (3枚出た状態) のとき: 最弱カードを出す
		// Player 3 (Team 1) の番。Trick: Player 1 (Team 1) が Spade Ace を出している。
		game.SetCurrentPlayerIdx(3)
		game.SetCurrentTrick([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
			{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 7, false)},
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignSpade, 8, false)},
		})
		setupOmiHand(game, 3, []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 13, false), // Spade King
			domain.NewCard(domain.CardDesignSpade, 9, false),  // Spade 9 (最弱)
		})
		game.CpuPlay()
		require.Equal(t, 4, len(game.GetCurrentTrick()))
		assert.Equal(t, 9, game.GetCurrentTrick()[3].Card.GetValue()) // 最弱の 9 を出した
	})
}

// --- 13. ゲッター/セッターと境界条件・エラーパス ---

func TestOmi_GettersSettersAndEdgeCases(t *testing.T) {
	game := newTestOmi()
	game.Reset()

	// Config
	cfg := domain.DefaultOmiConfig()
	cfg.PointLimit = 20
	game.SetConfig(cfg)
	assert.Equal(t, 20, game.GetConfig().PointLimit)

	// Setters & Getters
	game.SetLeadPlayerIdx(2)
	assert.Equal(t, 2, game.GetLeadPlayerIdx())

	game.SetWinnerTeam(1)
	assert.Equal(t, 1, game.GetWinnerTeam())

	assert.Equal(t, game.GetTrumpCallerIdx(), game.GetCallerIdx())

	// GetTeamScore 範囲外
	assert.Equal(t, 0, game.GetTeamScore(-1))
	assert.Equal(t, 0, game.GetTeamScore(99))

	// GetPlayer 範囲外
	assert.Nil(t, game.GetPlayer(-1))
	assert.Nil(t, game.GetPlayer(99))

	// IsHumanBidTurn
	assert.Equal(t, game.IsHumanCallTrumpTurn(), game.IsHumanBidTurn())

	// GetCurrentTrick
	assert.Nil(t, game.GetCurrentTrick())

	// ResolveTrick / NextTrick のフェーズガード
	game.SetPhase(domain.OmiPhasePlay)
	game.ResolveTrick() // TrickEnd でないので何もしない
	assert.Equal(t, domain.OmiPhasePlay, game.GetPhase())

	game.NextTrick() // TrickEnd でないので何もしない
	assert.Equal(t, domain.OmiPhasePlay, game.GetPhase())

	// NextRound のフェーズガード
	game.NextRound() // RoundEnd でないので何もしない
	assert.Equal(t, domain.OmiPhasePlay, game.GetPhase())

	// PlayerPlay 不正インデックスガード
	game.SetCurrentPlayerIdx(0)
	assert.ErrorIs(t, game.PlayerPlay(-1), domain.ErrInvalidCard)
	assert.ErrorIs(t, game.PlayerPlay(100), domain.ErrInvalidCard)

	// PlayerPlay / PlayerCallTrump / CpuPlay / CpuCallTrump 終了後ガード
	game.SetGameEndFlag(true)
	assert.ErrorIs(t, game.PlayerPlay(0), domain.ErrGameEnded)
	assert.ErrorIs(t, game.PlayerCallTrump(domain.CardDesignSpade), domain.ErrGameEnded)
	game.CpuPlay()      // no-op
	game.CpuCallTrump() // no-op

	// JSON unmarshal エラーパス
	var badOmi domain.Omi
	assert.Error(t, badOmi.UnmarshalJSON([]byte("invalid json")))

	// 配列サイズ上限チェック
	hugeJSON := `{"ps":[`
	for i := 0; i < 1005; i++ {
		if i > 0 {
			hugeJSON += ","
		}
		hugeJSON += "{}"
	}
	hugeJSON += `]}`
	assert.Error(t, badOmi.UnmarshalJSON([]byte(hugeJSON)))
}
