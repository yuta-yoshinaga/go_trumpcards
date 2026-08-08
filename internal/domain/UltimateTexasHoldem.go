//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
)

// アルティメット・テキサスホールデムフェーズ定数
const (
	UltimateTexasHoldemPhaseBet     = 1 // ベット（アンテ＋ブラインド＋オプションのトリップス）
	UltimateTexasHoldemPhasePreFlop = 2 // ホールカード公開後（Play 3x/4x or Check）
	UltimateTexasHoldemPhaseFlop    = 3 // フロップ公開後（Play 2x or Check）
	UltimateTexasHoldemPhaseRiver   = 4 // ターン＋リバー公開後（Play 1x or Fold）
	UltimateTexasHoldemPhaseEnd     = 5 // 終了
)

// アルティメット・テキサスホールデムデフォルト値
const (
	UltimateTexasHoldemDefaultChips = 1000  // デフォルトチップ
	UltimateTexasHoldemMinBet       = 10    // 最低アンテ
	UltimateTexasHoldemMaxBet       = 10000 // 最大アンテ
	UltimateTexasHoldemHoleCards    = 2     // ホールカード枚数
	UltimateTexasHoldemBoardCards   = 5     // コミュニティカード枚数
)

// プレイベットの倍率（フェーズ別に許可される値）
const (
	UltimateTexasHoldemPlayPreFlop4x = 4 // プリフロップ 4×アンテ
	UltimateTexasHoldemPlayPreFlop3x = 3 // プリフロップ 3×アンテ
	UltimateTexasHoldemPlayFlop2x    = 2 // フロップ後 2×アンテ
	UltimateTexasHoldemPlayRiver1x   = 1 // リバー後 1×アンテ
)

// Blind paytable（プレイヤー勝利かつストレート以上のときのみ支払う。
// ストレート未満で勝利／引き分けはプッシュ、敗北はブラインド没収）
const (
	UltimateTexasHoldemBlindPayRoyalFlush    = 500 // ロイヤルフラッシュ 500:1
	UltimateTexasHoldemBlindPayStraightFlush = 50  // ストレートフラッシュ 50:1
	UltimateTexasHoldemBlindPayFourOfAKind   = 10  // フォーカード 10:1
	UltimateTexasHoldemBlindPayFullHouse     = 3   // フルハウス 3:1
	UltimateTexasHoldemBlindPayStraight      = 1   // ストレート 1:1
	// フラッシュは 3:2 倍（最低ベット10／増分10で整数演算が安全。blindBet*3/2 で計算）
)

// Trips（オプションのサイドベット）paytable（ディーラーの手・勝敗に関係なく評価）
const (
	UltimateTexasHoldemTripsPayRoyalFlush    = 50 // ロイヤルフラッシュ 50:1
	UltimateTexasHoldemTripsPayStraightFlush = 40 // ストレートフラッシュ 40:1
	UltimateTexasHoldemTripsPayFourOfAKind   = 30 // フォーカード 30:1
	UltimateTexasHoldemTripsPayFullHouse     = 8  // フルハウス 8:1
	UltimateTexasHoldemTripsPayFlush         = 7  // フラッシュ 7:1
	UltimateTexasHoldemTripsPayStraight      = 4  // ストレート 4:1
	UltimateTexasHoldemTripsPayThreeOfAKind  = 3  // スリーカード 3:1
)

// UltimateTexasHoldem アルティメット・テキサスホールデムクラス
type UltimateTexasHoldem struct {
	trumpCards      *TrumpCards // トランプカード
	playerHand      []*Card     // プレイヤーホールカード
	dealerHand      []*Card     // ディーラーホールカード
	community       []*Card     // コミュニティカード
	chips           ChipHolder  // チップ
	anteBet         int         // アンテベット額
	blindBet        int         // ブラインドベット額（常に anteBet と同額）
	tripsBet        int         // トリップス（任意のサイドベット）額
	playBet         int         // プレイベット額（0/1×/2×/3×/4× アンテ）
	folded          bool        // リバーでフォールドしたかどうか
	phase           int         // 現在のフェーズ
	gameEndFlag     bool        // ゲーム終了フラグ
	result          GameResult  // ショーダウン結果（フォールド時は Lose）
	dealerQualified bool        // ディーラークオリファイ（ペア以上）
	antePayout      int         // アンテ配当（返却額込み）
	blindPayout     int         // ブラインド配当（返却額込み）
	playPayout      int         // プレイベット配当（返却額込み）
	tripsPayout     int         // トリップス配当（返却額込み）
	playerHandRank  int         // プレイヤー最良5枚ランク
	dealerHandRank  int         // ディーラー最良5枚ランク
	playerBest      []*Card     // プレイヤー最良5枚
	dealerBest      []*Card     // ディーラー最良5枚
	actionLogBase
}

// NewUltimateTexasHoldem コンストラクタ
func NewUltimateTexasHoldem(trumpCards *TrumpCards) *UltimateTexasHoldem {
	trumpCards.Shuffle()
	return &UltimateTexasHoldem{
		trumpCards: trumpCards,
		phase:      UltimateTexasHoldemPhaseBet,
	}
}

// NewDefaultUltimateTexasHoldem デフォルト設定でゲームを生成するファクトリ関数
func NewDefaultUltimateTexasHoldem() *UltimateTexasHoldem {
	u := NewUltimateTexasHoldem(NewTrumpCards(0))
	u.chips.SetChips(UltimateTexasHoldemDefaultChips)
	return u
}

// Reset ゲーム初期化（チップは引き継ぐ。残高がアンテ＋ブラインドに足りない場合のみ初期化）
func (u *UltimateTexasHoldem) Reset() {
	u.gameEndFlag = false
	u.phase = UltimateTexasHoldemPhaseBet
	u.playerHand = nil
	u.dealerHand = nil
	u.community = nil
	u.anteBet = 0
	u.blindBet = 0
	u.tripsBet = 0
	u.playBet = 0
	u.folded = false
	u.result = 0
	u.dealerQualified = false
	u.antePayout = 0
	u.blindPayout = 0
	u.playPayout = 0
	u.tripsPayout = 0
	u.playerHandRank = 0
	u.dealerHandRank = 0
	u.playerBest = nil
	u.dealerBest = nil
	u.actionLog = nil
	if u.chips.GetChips() < UltimateTexasHoldemMinBet*2 {
		u.chips.SetChips(UltimateTexasHoldemDefaultChips)
	}
	// Reset re-creates and re-shuffles the deck. Ten shuffles (vs one in the
	// constructor) is a deliberate paranoia step to reduce correlation between
	// successive rounds — match the TexasHoldemBonus / CaribbeanStud convention.
	u.trumpCards = NewTrumpCards(0)
	for range 10 {
		u.trumpCards.Shuffle()
	}
}

// Bet アンテ＋同額のブラインド＋オプションのトリップスをベットし、ホールカードを配る。
func (u *UltimateTexasHoldem) Bet(ante, trips int) error {
	if u.phase != UltimateTexasHoldemPhaseBet {
		return NewDomainError(ErrWrongPhase, "Bet is only allowed during the bet phase.")
	}
	if ante < UltimateTexasHoldemMinBet || ante%UltimateTexasHoldemMinBet != 0 || ante > UltimateTexasHoldemMaxBet {
		return NewDomainError(ErrInvalidAmount, "Invalid ante amount.")
	}
	if trips < 0 {
		return NewDomainError(ErrInvalidAmount, "Trips bet must not be negative.")
	}
	if trips > 0 && (trips < UltimateTexasHoldemMinBet || trips%UltimateTexasHoldemMinBet != 0 || trips > UltimateTexasHoldemMaxBet) {
		return NewDomainError(ErrInvalidAmount, "Invalid trips bet amount.")
	}
	totalCost := ante*2 + trips // ante + blind(=ante) + trips
	if !u.chips.SubtractChips(totalCost) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}
	u.anteBet = ante
	u.blindBet = ante
	u.tripsBet = trips
	u.appendLog(0, "bet", fmt.Sprintf("ante=%d blind=%d trips=%d", ante, u.blindBet, trips), nil)

	u.dealHole()
	u.phase = UltimateTexasHoldemPhasePreFlop
	return nil
}

// Play フェーズに応じたプレイベットを置き、必要なら残りのコミュニティカードを公開してショーダウンに進む。
// 許可される multiplier:
//
//	プリフロップ: 3 または 4
//	フロップ:     2
//	リバー:       1
func (u *UltimateTexasHoldem) Play(multiplier int) error {
	switch u.phase {
	case UltimateTexasHoldemPhasePreFlop:
		if multiplier != UltimateTexasHoldemPlayPreFlop3x && multiplier != UltimateTexasHoldemPlayPreFlop4x {
			return NewDomainError(ErrInvalidAmount, "Pre-flop play multiplier must be 3 or 4.")
		}
		bet := u.anteBet * multiplier
		if !u.chips.SubtractChips(bet) {
			return NewDomainError(ErrInsufficientChips, "Insufficient chips for play bet.")
		}
		u.playBet = bet
		u.appendLog(0, "play", fmt.Sprintf("preflop play=%dx (bet=%d)", multiplier, bet), nil)
		u.dealFlop()
		u.dealTurn()
		u.dealRiver()
		u.resolve()
		return nil
	case UltimateTexasHoldemPhaseFlop:
		if multiplier != UltimateTexasHoldemPlayFlop2x {
			return NewDomainError(ErrInvalidAmount, "Flop play multiplier must be 2.")
		}
		bet := u.anteBet * multiplier
		if !u.chips.SubtractChips(bet) {
			return NewDomainError(ErrInsufficientChips, "Insufficient chips for play bet.")
		}
		u.playBet = bet
		u.appendLog(0, "play", fmt.Sprintf("flop play=2x (bet=%d)", bet), nil)
		u.dealTurn()
		u.dealRiver()
		u.resolve()
		return nil
	case UltimateTexasHoldemPhaseRiver:
		if multiplier != UltimateTexasHoldemPlayRiver1x {
			return NewDomainError(ErrInvalidAmount, "River play multiplier must be 1.")
		}
		bet := u.anteBet * multiplier
		if !u.chips.SubtractChips(bet) {
			return NewDomainError(ErrInsufficientChips, "Insufficient chips for play bet.")
		}
		u.playBet = bet
		u.appendLog(0, "play", fmt.Sprintf("river play=1x (bet=%d)", bet), nil)
		u.resolve()
		return nil
	default:
		return NewDomainError(ErrWrongPhase, "Play is not allowed in the current phase.")
	}
}

// Check プリフロップまたはフロップ後にチェック（ベットせず次フェーズへ）。
// フロップ後でチェックするとターン＋リバーが同時に公開される（リバーフェーズへ進む）。
func (u *UltimateTexasHoldem) Check() error {
	switch u.phase {
	case UltimateTexasHoldemPhasePreFlop:
		u.appendLog(0, "check", "preflop check", nil)
		u.dealFlop()
		u.phase = UltimateTexasHoldemPhaseFlop
		return nil
	case UltimateTexasHoldemPhaseFlop:
		u.appendLog(0, "check", "flop check", nil)
		u.dealTurn()
		u.dealRiver()
		u.phase = UltimateTexasHoldemPhaseRiver
		return nil
	default:
		return NewDomainError(ErrWrongPhase, "Check is only allowed during the pre-flop or flop phase.")
	}
}

// Fold リバー後のみ可。アンテ＋ブラインドを失う。トリップスは別途評価。
func (u *UltimateTexasHoldem) Fold() error {
	if u.phase != UltimateTexasHoldemPhaseRiver {
		return NewDomainError(ErrWrongPhase, "Fold is only allowed during the river phase.")
	}
	u.appendLog(0, "fold", "player folds", nil)

	u.folded = true
	u.result = GameResultLose
	// 棋譜とトリップス評価のためにプレイヤーランクを確定する。
	playerAll := append([]*Card{}, u.playerHand...)
	playerAll = append(playerAll, u.community...)
	u.playerHandRank, u.playerBest = evalBestFromSeven(playerAll)
	dealerAll := append([]*Card{}, u.dealerHand...)
	dealerAll = append(dealerAll, u.community...)
	u.dealerHandRank, u.dealerBest = evalBestFromSeven(dealerAll)
	u.dealerQualified = u.checkDealerQualifies()

	u.evaluateTrips()
	if u.tripsPayout > 0 {
		u.chips.AddChips(u.tripsPayout)
	}

	u.gameEndFlag = true
	u.phase = UltimateTexasHoldemPhaseEnd
	u.appendLog(-1, "result", "player folded", nil)
	return nil
}

// dealHole 各2枚のホールカードを配る。
func (u *UltimateTexasHoldem) dealHole() {
	u.playerHand = make([]*Card, 0, UltimateTexasHoldemHoleCards)
	u.dealerHand = make([]*Card, 0, UltimateTexasHoldemHoleCards)
	for range UltimateTexasHoldemHoleCards {
		u.playerHand = append(u.playerHand, u.trumpCards.DrawCard())
		u.dealerHand = append(u.dealerHand, u.trumpCards.DrawCard())
	}
	u.appendLog(-1, "deal", "dealt 2 hole cards each", nil)
}

// dealFlop 3枚のフロップを配る。
func (u *UltimateTexasHoldem) dealFlop() {
	if u.community == nil {
		u.community = make([]*Card, 0, UltimateTexasHoldemBoardCards)
	}
	for range 3 {
		u.community = append(u.community, u.trumpCards.DrawCard())
	}
	u.updatePlayerCurrentRank()
	u.appendLog(-1, "flop", "flop dealt", nil)
}

// dealTurn ターン（4枚目）を配る。
func (u *UltimateTexasHoldem) dealTurn() {
	u.community = append(u.community, u.trumpCards.DrawCard())
	u.updatePlayerCurrentRank()
	u.appendLog(-1, "turn", "turn dealt", nil)
}

// dealRiver リバー（5枚目）を配る。
func (u *UltimateTexasHoldem) dealRiver() {
	u.community = append(u.community, u.trumpCards.DrawCard())
	u.appendLog(-1, "river", "river dealt", nil)
}

// updatePlayerCurrentRank プレイヤーの現時点での最良ハンドランクを更新する（フロントエンドのヒント表示用）。
func (u *UltimateTexasHoldem) updatePlayerCurrentRank() {
	all := append([]*Card{}, u.playerHand...)
	all = append(all, u.community...)
	u.playerHandRank, _ = evalBestFromSeven(all)
}

// 推奨アクション。CUI のヒント表示 (#4709) と Web のボタン強調が同じ判断を
// 出すよう、判定はここ1か所に置く。
const (
	// UTHRecommendPlay4x プリフロップで 4x を置く。
	UTHRecommendPlay4x = "play4x"
	// UTHRecommendPlay3x プリフロップで 3x を置く。
	UTHRecommendPlay3x = "play3x"
	// UTHRecommendPlay2x フロップで 2x を置く。
	UTHRecommendPlay2x = "play2x"
	// UTHRecommendPlay1x リバーで 1x を置く。
	UTHRecommendPlay1x = "play1x"
	// UTHRecommendCheck チェックして次のストリートを見る。
	UTHRecommendCheck = "check"
	// UTHRecommendFold リバーで降りる。
	UTHRecommendFold = "fold"
)

// RecommendPlay は現在のフェーズでの推奨アクションを返す。判断のいらない
// フェーズでは空文字。
//
// **倍率まで返す。**Web はプリフロップの強さで 4x / 3x ボタンを光らせるのに、
// CUI には 4x/3x/2x/1x/check/fold を選ぶ材料が何も無かった (#4709)。
//
// 判定はフロントの utHoldemPreflopStrength と同じ規則 (strong→4x、
// moderate→3x、weak→check)。フロップ以降はワンペア以上かどうかで決める。
func (u *UltimateTexasHoldem) RecommendPlay() string {
	switch u.phase {
	case UltimateTexasHoldemPhasePreFlop:
		switch uthPreflopStrength(u.playerHand) {
		case uthStrengthStrong:
			return UTHRecommendPlay4x
		case uthStrengthModerate:
			return UTHRecommendPlay3x
		default:
			return UTHRecommendCheck
		}
	case UltimateTexasHoldemPhaseFlop:
		if u.currentBestRank() >= PokerHandOnePair {
			return UTHRecommendPlay2x
		}
		return UTHRecommendCheck
	case UltimateTexasHoldemPhaseRiver:
		if u.currentBestRank() >= PokerHandOnePair {
			return UTHRecommendPlay1x
		}
		return UTHRecommendFold
	default:
		return ""
	}
}

// currentBestRank はホールカードとボードから今の最良ランクを評価する。
//
// **保存済みの playerHandRank は読まない。**あれは配札のたびに更新される
// フィールドで、いつ更新されたかに助言が依存してしまう。
func (u *UltimateTexasHoldem) currentBestRank() int {
	all := append([]*Card{}, u.playerHand...)
	all = append(all, u.community...)
	if len(all) < 5 {
		return PokerHandHighCard
	}
	rank, _ := evalBestFromSeven(all)
	return rank
}

// uthStrength はプリフロップの手の強さ。
type uthStrength int

const (
	uthStrengthWeak uthStrength = iota
	uthStrengthModerate
	uthStrengthStrong
)

// uthPreflopStrength はホールカード2枚の強さを返す。
//
// **フロントの utHoldemPreflopStrength と同じ規則。**ずれると同じ手札で
// CUI が 3x、Web のボタン強調が 4x を指すことになる。エースは 14 として
// 数える (value は 1)。
func uthPreflopStrength(hand []*Card) uthStrength {
	if len(hand) < 2 {
		return uthStrengthWeak
	}
	rank := func(c *Card) int {
		if v := c.GetValue(); v == 1 {
			return 14
		}
		return c.GetValue()
	}
	ra, rb := rank(hand[0]), rank(hand[1])
	hi, lo := ra, rb
	if lo > hi {
		hi, lo = lo, hi
	}
	suited := hand[0].GetDesign() == hand[1].GetDesign()

	switch {
	case ra == rb: // any pair
		return uthStrengthStrong
	case hi == 14: // any ace
		return uthStrengthStrong
	case hi == 13 && suited: // suited king
		return uthStrengthStrong
	case hi == 13 && lo >= 11: // K-Q / K-J offsuit
		return uthStrengthStrong
	case hi == 12 && lo == 11 && suited: // Q-J suited
		return uthStrengthStrong
	case hi == 13: // K-x offsuit
		return uthStrengthModerate
	case hi == 12 && lo == 11: // Q-J offsuit
		return uthStrengthModerate
	case suited && hi-lo <= 2 && lo >= 6: // suited connector mid+
		return uthStrengthModerate
	case suited && hi >= 12: // Q-x suited
		return uthStrengthModerate
	default:
		return uthStrengthWeak
	}
}

// resolve ショーダウン処理。
func (u *UltimateTexasHoldem) resolve() {
	playerAll := append([]*Card{}, u.playerHand...)
	playerAll = append(playerAll, u.community...)
	dealerAll := append([]*Card{}, u.dealerHand...)
	dealerAll = append(dealerAll, u.community...)

	u.playerHandRank, u.playerBest = evalBestFromSeven(playerAll)
	u.dealerHandRank, u.dealerBest = evalBestFromSeven(dealerAll)
	u.dealerQualified = u.checkDealerQualifies()

	cmp := u.compareBest()
	switch {
	case cmp > 0:
		u.result = GameResultWin
	case cmp < 0:
		u.result = GameResultLose
	default:
		u.result = GameResultDraw
	}

	u.calculatePayouts()
	u.evaluateTrips()

	totalPayout := u.antePayout + u.blindPayout + u.playPayout + u.tripsPayout
	if totalPayout > 0 {
		u.chips.AddChips(totalPayout)
	}

	u.gameEndFlag = true
	u.phase = UltimateTexasHoldemPhaseEnd

	var resultStr string
	switch u.result {
	case GameResultWin:
		resultStr = "player wins"
	case GameResultDraw:
		resultStr = "push"
	default:
		resultStr = "dealer wins"
	}
	u.appendLog(-1, "result", resultStr, nil)
}

// compareBest 最良5枚の比較
func (u *UltimateTexasHoldem) compareBest() int {
	if u.playerHandRank > u.dealerHandRank {
		return 1
	}
	if u.playerHandRank < u.dealerHandRank {
		return -1
	}
	return compareHighCardsSlice(u.playerBest, u.dealerBest)
}

// checkDealerQualifies ディーラークオリファイ条件: ワンペア以上
func (u *UltimateTexasHoldem) checkDealerQualifies() bool {
	return u.dealerHandRank >= PokerHandOnePair
}

// calculatePayouts アンテ／ブラインド／プレイの配当を計算する。
func (u *UltimateTexasHoldem) calculatePayouts() {
	// ── アンテ ──
	// ディーラー未クオリファイ時は常にプッシュ。クオリファイ時は勝敗で決まる。
	if !u.dealerQualified {
		u.antePayout = u.anteBet
	} else {
		switch u.result {
		case GameResultWin:
			u.antePayout = u.anteBet * 2
		case GameResultDraw:
			u.antePayout = u.anteBet
		case GameResultLose:
			u.antePayout = 0
		}
	}

	// ── プレイ ──
	// クオリファイ有無に関わらず通常通り（勝てば1:1、引き分けでプッシュ、負けで没収）
	switch u.result {
	case GameResultWin:
		u.playPayout = u.playBet * 2
	case GameResultDraw:
		u.playPayout = u.playBet
	case GameResultLose:
		u.playPayout = 0
	}

	// ── ブラインド ──
	// 勝利: ストレート以上ならペイテーブル、未満ならプッシュ。
	// 引き分け: プッシュ。負け: 没収。
	switch u.result {
	case GameResultWin:
		u.blindPayout = u.blindBet + u.blindProfit()
	case GameResultDraw:
		u.blindPayout = u.blindBet
	case GameResultLose:
		u.blindPayout = 0
	}
}

// blindProfit ブラインド配当（返却分を除く純利益）。ストレート未満は0（プッシュ扱い）。
func (u *UltimateTexasHoldem) blindProfit() int {
	switch u.playerHandRank {
	case PokerHandRoyalFlush:
		return u.blindBet * UltimateTexasHoldemBlindPayRoyalFlush
	case PokerHandStraightFlush:
		return u.blindBet * UltimateTexasHoldemBlindPayStraightFlush
	case PokerHandFourOfAKind:
		return u.blindBet * UltimateTexasHoldemBlindPayFourOfAKind
	case PokerHandFullHouse:
		return u.blindBet * UltimateTexasHoldemBlindPayFullHouse
	case PokerHandFlush:
		return u.blindBet * 3 / 2 // 3:2（最低ベット10/増分10なので整数演算で正確）
	case PokerHandStraight:
		return u.blindBet * UltimateTexasHoldemBlindPayStraight
	default:
		return 0
	}
}

// evaluateTrips トリップスサイドベットの評価（ディーラーの手や勝敗、フォールドに関わらず評価）
func (u *UltimateTexasHoldem) evaluateTrips() {
	if u.tripsBet <= 0 {
		return
	}
	var mult int
	switch u.playerHandRank {
	case PokerHandRoyalFlush:
		mult = UltimateTexasHoldemTripsPayRoyalFlush
	case PokerHandStraightFlush:
		mult = UltimateTexasHoldemTripsPayStraightFlush
	case PokerHandFourOfAKind:
		mult = UltimateTexasHoldemTripsPayFourOfAKind
	case PokerHandFullHouse:
		mult = UltimateTexasHoldemTripsPayFullHouse
	case PokerHandFlush:
		mult = UltimateTexasHoldemTripsPayFlush
	case PokerHandStraight:
		mult = UltimateTexasHoldemTripsPayStraight
	case PokerHandThreeOfAKind:
		mult = UltimateTexasHoldemTripsPayThreeOfAKind
	default:
		return
	}
	u.tripsPayout = u.tripsBet + u.tripsBet*mult
}

// --- Getters ---

// GetPlayerHand プレイヤーホールカード取得
func (u *UltimateTexasHoldem) GetPlayerHand() []*Card { return u.playerHand }

// GetDealerHand ディーラーホールカード取得
func (u *UltimateTexasHoldem) GetDealerHand() []*Card { return u.dealerHand }

// GetCommunity コミュニティカード取得
func (u *UltimateTexasHoldem) GetCommunity() []*Card { return u.community }

// GetPhase 現在のフェーズ
func (u *UltimateTexasHoldem) GetPhase() int { return u.phase }

// GetGameEndFlag ゲーム終了フラグ
func (u *UltimateTexasHoldem) GetGameEndFlag() bool { return u.gameEndFlag }

// GetAnteBet アンテベット額
func (u *UltimateTexasHoldem) GetAnteBet() int { return u.anteBet }

// GetBlindBet ブラインドベット額
func (u *UltimateTexasHoldem) GetBlindBet() int { return u.blindBet }

// GetTripsBet トリップス（サイドベット）額
func (u *UltimateTexasHoldem) GetTripsBet() int { return u.tripsBet }

// GetPlayBet プレイベット額
func (u *UltimateTexasHoldem) GetPlayBet() int { return u.playBet }

// GetFolded リバーでフォールドしたか
func (u *UltimateTexasHoldem) GetFolded() bool { return u.folded }

// GetResult ゲーム結果
func (u *UltimateTexasHoldem) GetResult() GameResult { return u.result }

// GetDealerQualified ディーラークオリファイ状態
func (u *UltimateTexasHoldem) GetDealerQualified() bool { return u.dealerQualified }

// GetAntePayout アンテ配当（返却含む合計）
func (u *UltimateTexasHoldem) GetAntePayout() int { return u.antePayout }

// GetBlindPayout ブラインド配当（返却含む合計）
func (u *UltimateTexasHoldem) GetBlindPayout() int { return u.blindPayout }

// GetPlayPayout プレイベット配当（返却含む合計）
func (u *UltimateTexasHoldem) GetPlayPayout() int { return u.playPayout }

// GetTripsPayout トリップス配当（返却含む合計）
func (u *UltimateTexasHoldem) GetTripsPayout() int { return u.tripsPayout }

// GetTotalPayout 合計配当
func (u *UltimateTexasHoldem) GetTotalPayout() int {
	return u.antePayout + u.blindPayout + u.playPayout + u.tripsPayout
}

// GetPlayerHandRank プレイヤーハンドランク
func (u *UltimateTexasHoldem) GetPlayerHandRank() int { return u.playerHandRank }

// GetDealerHandRank ディーラーハンドランク
func (u *UltimateTexasHoldem) GetDealerHandRank() int { return u.dealerHandRank }

// GetPlayerBest プレイヤーの最良5枚
func (u *UltimateTexasHoldem) GetPlayerBest() []*Card { return u.playerBest }

// GetDealerBest ディーラーの最良5枚
func (u *UltimateTexasHoldem) GetDealerBest() []*Card { return u.dealerBest }

// GetChips チップ
func (u *UltimateTexasHoldem) GetChips() int { return u.chips.GetChips() }

// --- Test helpers ---

// SetPhase フェーズ設定（テスト用）
func (u *UltimateTexasHoldem) SetPhase(phase int) { u.phase = phase }

// SetPlayerHand プレイヤーホールカード設定（テスト用）
func (u *UltimateTexasHoldem) SetPlayerHand(cards []*Card) { u.playerHand = cards }

// SetDealerHand ディーラーホールカード設定（テスト用）
func (u *UltimateTexasHoldem) SetDealerHand(cards []*Card) { u.dealerHand = cards }

// SetCommunity コミュニティカード設定（テスト用）
func (u *UltimateTexasHoldem) SetCommunity(cards []*Card) { u.community = cards }

// SetAnteBet アンテベット額設定（テスト用）
func (u *UltimateTexasHoldem) SetAnteBet(amount int) { u.anteBet = amount }

// SetBlindBet ブラインドベット額設定（テスト用）
func (u *UltimateTexasHoldem) SetBlindBet(amount int) { u.blindBet = amount }

// SetTripsBet トリップスベット額設定（テスト用）
func (u *UltimateTexasHoldem) SetTripsBet(amount int) { u.tripsBet = amount }

// SetPlayBet プレイベット額設定（テスト用）
func (u *UltimateTexasHoldem) SetPlayBet(amount int) { u.playBet = amount }

// SetChips チップ設定（テスト用）
func (u *UltimateTexasHoldem) SetChips(chips int) { u.chips.SetChips(chips) }

// ultimateTexasHoldemJSON は UltimateTexasHoldem の JSON ワイヤーフォーマット
type ultimateTexasHoldemJSON struct {
	TrumpCards      *TrumpCards       `json:"tc"`
	PlayerHand      []*Card           `json:"ph"`
	DealerHand      []*Card           `json:"dh"`
	Community       []*Card           `json:"cm"`
	Chips           *ChipHolder       `json:"ch"`
	AnteBet         int               `json:"ab"`
	BlindBet        int               `json:"bb"`
	TripsBet        int               `json:"tb"`
	PlayBet         int               `json:"pl"`
	Folded          bool              `json:"fd"`
	Phase           int               `json:"ps"`
	GameEndFlag     bool              `json:"ge"`
	Result          GameResult        `json:"rs"`
	DealerQualified bool              `json:"dq"`
	AntePayout      int               `json:"ap"`
	BlindPayout     int               `json:"bp"`
	PlayPayout      int               `json:"pp"`
	TripsPayout     int               `json:"tp"`
	PlayerHandRank  int               `json:"pr"`
	DealerHandRank  int               `json:"dr"`
	PlayerBest      []*Card           `json:"pB"`
	DealerBest      []*Card           `json:"dB"`
	ActionLog       []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (u *UltimateTexasHoldem) MarshalJSON() ([]byte, error) {
	return json.Marshal(ultimateTexasHoldemJSON{
		TrumpCards:      u.trumpCards,
		PlayerHand:      u.playerHand,
		DealerHand:      u.dealerHand,
		Community:       u.community,
		Chips:           &u.chips,
		AnteBet:         u.anteBet,
		BlindBet:        u.blindBet,
		TripsBet:        u.tripsBet,
		PlayBet:         u.playBet,
		Folded:          u.folded,
		Phase:           u.phase,
		GameEndFlag:     u.gameEndFlag,
		Result:          u.result,
		DealerQualified: u.dealerQualified,
		AntePayout:      u.antePayout,
		BlindPayout:     u.blindPayout,
		PlayPayout:      u.playPayout,
		TripsPayout:     u.tripsPayout,
		PlayerHandRank:  u.playerHandRank,
		DealerHandRank:  u.dealerHandRank,
		PlayerBest:      u.playerBest,
		DealerBest:      u.dealerBest,
		ActionLog:       u.actionLog,
	})
}

// ultimateTexasHoldemMaxSliceLen caps slice sizes during deserialisation.
const ultimateTexasHoldemMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (u *UltimateTexasHoldem) UnmarshalJSON(data []byte) error {
	var j ultimateTexasHoldemJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.PlayerHand) > ultimateTexasHoldemMaxSliceLen ||
		len(j.DealerHand) > ultimateTexasHoldemMaxSliceLen ||
		len(j.Community) > ultimateTexasHoldemMaxSliceLen ||
		len(j.PlayerBest) > ultimateTexasHoldemMaxSliceLen ||
		len(j.DealerBest) > ultimateTexasHoldemMaxSliceLen ||
		len(j.ActionLog) > ultimateTexasHoldemMaxSliceLen {
		return fmt.Errorf("ultimatetexasholdem: input array exceeds maximum allowed size")
	}

	u.trumpCards = j.TrumpCards
	if u.trumpCards == nil {
		u.trumpCards = NewTrumpCards(0)
	}
	u.playerHand = sliceOrEmpty(j.PlayerHand)
	u.dealerHand = sliceOrEmpty(j.DealerHand)
	u.community = sliceOrEmpty(j.Community)
	if j.Chips != nil {
		u.chips = *j.Chips
	}
	u.anteBet = j.AnteBet
	u.blindBet = j.BlindBet
	u.tripsBet = j.TripsBet
	u.playBet = j.PlayBet
	u.folded = j.Folded
	u.phase = j.Phase
	u.gameEndFlag = j.GameEndFlag
	u.result = j.Result
	u.dealerQualified = j.DealerQualified
	u.antePayout = j.AntePayout
	u.blindPayout = j.BlindPayout
	u.playPayout = j.PlayPayout
	u.tripsPayout = j.TripsPayout
	u.playerHandRank = j.PlayerHandRank
	u.dealerHandRank = j.DealerHandRank
	u.playerBest = sliceOrEmpty(j.PlayerBest)
	u.dealerBest = sliceOrEmpty(j.DealerBest)
	u.actionLog = j.ActionLog
	if u.actionLog == nil {
		u.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
