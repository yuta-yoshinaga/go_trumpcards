//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
)

// カリビアン・ドロー・ポーカーフェーズ定数
//
// **クローン元の Caribbean Stud には無いドローフェーズが挟まる。** 完全な
// スタッドなら配られた 5 枚をそのままコールするか降りるかだけだが、こちらは
// 勝負を決める前に手札を作り直せる。
const (
	CaribbeanDrawPhaseBet    = 1 // ベットフェーズ
	CaribbeanDrawPhaseDraw   = 2 // ドローフェーズ（最大2枚まで交換）
	CaribbeanDrawPhaseAction = 3 // アクションフェーズ（Call/Fold選択）
	CaribbeanDrawPhaseEnd    = 4 // 終了フェーズ
)

// ドロー（交換）の制約。
const (
	// CaribbeanDrawMaxExchange は 1 ラウンドに交換できる最大枚数。
	CaribbeanDrawMaxExchange = 2
	// CaribbeanDrawQualifyPairValue はディーラーがクオリファイする最低ペア。
	// **8 のペア。** クローン元の「ペア以上、または A-K ハイ」より厳しい。
	CaribbeanDrawQualifyPairValue = 8
	// CaribbeanDrawExchangeCostRatio は交換手数料をアンテの何倍にするか。
	// **交換はタダではない** —— 手数料が無ければ常に交換するのが最適になり、
	// 「引くかどうか」という唯一の判断が消える。
	CaribbeanDrawExchangeCostRatio = 1
)

// カリビアン・ドロー・ポーカーデフォルト値
const (
	CaribbeanDrawDefaultChips = 1000  // デフォルトチップ
	CaribbeanDrawMinBet       = 10    // 最低ベット額
	CaribbeanDrawMaxBet       = 10000 // 最大ベット額
	CaribbeanDrawHandSize     = 5     // ハンドサイズ
)

// プレイベット配当倍率（コール時）。
//
// **クローン元より薄い。** 最大2枚引ける卓では役が完成しやすく、スタッドの
// 配当表をそのまま置くとプレイヤー有利に振り切れる。実在の Caribbean Draw も
// 同じ理由で上位役を絞っている。
const (
	CaribbeanDrawPayRoyalFlush    = 50 // ロイヤルフラッシュ 50:1
	CaribbeanDrawPayStraightFlush = 20 // ストレートフラッシュ 20:1
	CaribbeanDrawPayFourOfAKind   = 10 // フォーカード 10:1
	CaribbeanDrawPayFullHouse     = 5  // フルハウス 5:1
	CaribbeanDrawPayFlush         = 4  // フラッシュ 4:1
	CaribbeanDrawPayStraight      = 3  // ストレート 3:1
	CaribbeanDrawPayThreeOfAKind  = 2  // スリーカード 2:1
	CaribbeanDrawPayTwoPair       = 1  // ツーペア 1:1
	CaribbeanDrawPayPair          = 1  // ワンペア以下 1:1
)

// プログレッシブジャックポット（サイドベット）配当倍率。
//
// **交換後の手で判定する。** 引いて完成させたロイヤルフラッシュも本物として
// 払う —— サイドベットが賭けているのは「配られた手」ではなく「作った手」。
// 倍率はクローン元より控えめで、ドローぶんの完成しやすさを吸収する。
const (
	CaribbeanDrawJackpotRoyalFlush    = 10000 // ロイヤルフラッシュ
	CaribbeanDrawJackpotStraightFlush = 2000  // ストレートフラッシュ
	CaribbeanDrawJackpotFourOfAKind   = 200   // フォーカード
	CaribbeanDrawJackpotFullHouse     = 50    // フルハウス
	CaribbeanDrawJackpotFlush         = 25    // フラッシュ
)

// CaribbeanDraw カリビアン・ドロー・ポーカークラス
type CaribbeanDraw struct {
	trumpCards      *TrumpCards // トランプカード
	playerHand      []*Card     // プレイヤーハンド
	dealerHand      []*Card     // ディーラーハンド
	chips           ChipHolder  // チップ
	anteBet         int         // アンテベット額
	jackpotBet      int         // ジャックポットサイドベット額
	playBet         int         // プレイ（コール）ベット額
	phase           int         // 現在のフェーズ
	gameEndFlag     bool        // ゲーム終了フラグ
	result          GameResult  // ゲーム結果
	antePayout      int         // アンテ配当
	playPayout      int         // プレイ配当
	jackpotPayout   int         // ジャックポット配当
	drawCost        int         // ドロー手数料（結果画面に残す）
	dealerQualified bool        // ディーラークオリファイフラグ
	playerHandRank  int         // プレイヤーハンドランク
	dealerHandRank  int         // ディーラーハンドランク
	actionLogBase
}

// NewCaribbeanDraw コンストラクタ
func NewCaribbeanDraw(trumpCards *TrumpCards) *CaribbeanDraw {
	trumpCards.Shuffle()
	return &CaribbeanDraw{
		trumpCards: trumpCards,
		phase:      CaribbeanDrawPhaseBet,
	}
}

// NewDefaultCaribbeanDraw デフォルト設定のカリビアン・ドロー・ポーカーを生成するファクトリ関数
func NewDefaultCaribbeanDraw() *CaribbeanDraw {
	cs := NewCaribbeanDraw(NewTrumpCards(0))
	cs.chips.SetChips(CaribbeanDrawDefaultChips)
	return cs
}

// Reset ゲーム初期化
func (cs *CaribbeanDraw) Reset() {
	cs.gameEndFlag = false
	cs.phase = CaribbeanDrawPhaseBet
	cs.drawCost = 0
	cs.playerHand = nil
	cs.dealerHand = nil
	cs.anteBet = 0
	cs.jackpotBet = 0
	cs.playBet = 0
	cs.result = 0
	cs.antePayout = 0
	cs.playPayout = 0
	cs.jackpotPayout = 0
	cs.dealerQualified = false
	cs.playerHandRank = 0
	cs.dealerHandRank = 0
	cs.actionLog = nil
	if cs.chips.GetChips() < CaribbeanDrawMinBet {
		cs.chips.SetChips(CaribbeanDrawDefaultChips)
	}
	cs.trumpCards = NewTrumpCards(0)
	for range 10 {
		cs.trumpCards.Shuffle()
	}
}

// Bet アンテベット＆カード配布。jackpot に正の値を渡すとジャックポットサイドベットを追加する。
func (cs *CaribbeanDraw) Bet(ante, jackpot int) error {
	if cs.phase != CaribbeanDrawPhaseBet {
		return NewDomainError(ErrWrongPhase, "Bet is only allowed during the bet phase.")
	}
	if ante < CaribbeanDrawMinBet || ante%CaribbeanDrawMinBet != 0 || ante > CaribbeanDrawMaxBet {
		return NewDomainError(ErrInvalidAmount, "Invalid ante amount.")
	}
	if jackpot < 0 {
		return NewDomainError(ErrInvalidAmount, "Jackpot bet must not be negative.")
	}
	if jackpot > 0 && (jackpot < CaribbeanDrawMinBet || jackpot%CaribbeanDrawMinBet != 0 || jackpot > CaribbeanDrawMaxBet) {
		return NewDomainError(ErrInvalidAmount, "Invalid jackpot bet amount.")
	}
	totalCost := ante + jackpot
	if !cs.chips.SubtractChips(totalCost) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}
	cs.anteBet = ante
	cs.jackpotBet = jackpot
	cs.appendLog(0, "bet", fmt.Sprintf("ante=%d jackpot=%d", ante, jackpot), nil)

	cs.deal()
	cs.phase = CaribbeanDrawPhaseDraw
	return nil
}

// Draw は手札のうち最大2枚を交換する。空の indices は「交換しない」。
//
// **手数料はアンテと同額。** タダなら常に引くのが最適になり、このゲームが
// クローン元に足している唯一の判断が消える。降りる前提で引かないことにも
// 意味が出るよう、料金はドローを選んだときにだけ引く。
func (cs *CaribbeanDraw) Draw(indices []int) error {
	if cs.phase != CaribbeanDrawPhaseDraw {
		return NewDomainError(ErrWrongPhase, "Draw is only allowed during the draw phase.")
	}
	if len(indices) > CaribbeanDrawMaxExchange {
		return NewDomainError(ErrInvalidPlay,
			fmt.Sprintf("At most %d cards may be exchanged.", CaribbeanDrawMaxExchange))
	}
	seen := make(map[int]bool, len(indices))
	for _, i := range indices {
		if i < 0 || i >= len(cs.playerHand) {
			return NewDomainError(ErrInvalidPlay, "Card index out of range.")
		}
		if seen[i] {
			return NewDomainError(ErrInvalidPlay, "The same card cannot be exchanged twice.")
		}
		seen[i] = true
	}

	if len(indices) > 0 {
		cost := cs.anteBet * CaribbeanDrawExchangeCostRatio
		if !cs.chips.SubtractChips(cost) {
			return NewDomainError(ErrInsufficientChips, "Insufficient chips for the draw.")
		}
		cs.drawCost = cost
		// **引いた札は必ず置き換える。** 先に全部抜いてから配ると添字がずれる
		// ので、その場で差し替える。
		for _, i := range indices {
			// **DrawCard は山が尽きると nil を返す。** そのまま手札へ入れると
			// 役の評価が nil を踏んで落ちる。通常の 1 ラウンドは 52 枚中 12 枚
			// しか使わないので起きないが、KV から戻した卓の deckDrawCnt 次第で
			// 到達しうる。引けなければその札は替えない。
			if c := cs.trumpCards.DrawCard(); c != nil {
				cs.playerHand[i] = c
			}
		}
		// 引いた後の手で役を取り直す。ここを忘れると、交換して完成させた役が
		// 画面に出ないまま勝負することになる。
		cs.playerHandRank = evalFiveCardHand(cs.playerHand)
		cs.appendLog(0, "draw", fmt.Sprintf("exchanged %d card(s) for %d", len(indices), cost), nil)
	} else {
		cs.appendLog(0, "draw", "stands pat", nil)
	}

	cs.phase = CaribbeanDrawPhaseAction
	return nil
}

// Play コール（アンテの2倍のプレイベットを置いて勝負）
func (cs *CaribbeanDraw) Play() error {
	if cs.phase != CaribbeanDrawPhaseAction {
		return NewDomainError(ErrWrongPhase, "Play is only allowed during the action phase.")
	}
	playBet := cs.anteBet * 2
	if !cs.chips.SubtractChips(playBet) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips for play bet.")
	}
	cs.playBet = playBet
	cs.appendLog(0, "play", fmt.Sprintf("play bet=%d", cs.playBet), nil)

	cs.resolve()
	return nil
}

// Fold フォールド（アンテ没収。ジャックポットは別途評価）
func (cs *CaribbeanDraw) Fold() error {
	if cs.phase != CaribbeanDrawPhaseAction {
		return NewDomainError(ErrWrongPhase, "Fold is only allowed during the action phase.")
	}
	cs.appendLog(0, "fold", "player folds", nil)

	cs.result = GameResultLose
	cs.playerHandRank = evalFiveCardHand(cs.playerHand)
	cs.dealerHandRank = evalFiveCardHand(cs.dealerHand)
	// **降りても資格の有無は計算する。** 配当には効かない (勝負していないので
	// calculatePayouts は走らない) が、結果画面はディーラーの手を開いて資格を
	// 書く。ここを飛ばすと K のペアを持つディーラーに「クオリファイせず」と出る。
	cs.dealerQualified = cs.checkDealerQualifies()

	cs.evaluateJackpot()
	if cs.jackpotPayout > 0 {
		cs.chips.AddChips(cs.jackpotPayout)
	}

	cs.gameEndFlag = true
	cs.phase = CaribbeanDrawPhaseEnd
	cs.appendLog(-1, "result", "player folded", nil)
	return nil
}

// deal 5枚ずつ配る
func (cs *CaribbeanDraw) deal() {
	cs.playerHand = make([]*Card, 0, CaribbeanDrawHandSize)
	cs.dealerHand = make([]*Card, 0, CaribbeanDrawHandSize)
	for range CaribbeanDrawHandSize {
		cs.playerHand = append(cs.playerHand, cs.trumpCards.DrawCard())
		cs.dealerHand = append(cs.dealerHand, cs.trumpCards.DrawCard())
	}
	// **自分の役はこの時点で決まる。** 見えている 5 枚の評価なので隠す理由が
	// 無く、むしろ「どれを捨てるか」を決める唯一の材料。resolve まで 0 の
	// ままだと、フラッシュを配られても画面には "High Card" と出る。
	cs.playerHandRank = evalFiveCardHand(cs.playerHand)
	cs.appendLog(-1, "deal", "dealt 5 cards each", nil)
}

// resolve ゲーム解決（Play後の処理）
func (cs *CaribbeanDraw) resolve() {
	cs.playerHandRank = evalFiveCardHand(cs.playerHand)
	cs.dealerHandRank = evalFiveCardHand(cs.dealerHand)
	cs.dealerQualified = cs.checkDealerQualifies()

	switch cmp := cs.compareHands(); {
	case !cs.dealerQualified:
		// **勝負が成立していないので役は比べない。** アンテが 1:1、プレイは
		// 返却なので、3×ante 賭けて 4×ante 戻る —— 手の善し悪しに関わらず
		// 取り分は必ずアンテぶん増える。素の比較で result を決めると、
		// 儲けた局面に「ディーラーの勝ちです」と赤字で出すことになり、
		// 文言・勝敗演出・実際の残高が三者三様にずれる。
		cs.result = GameResultWin
	case cmp > 0:
		cs.result = GameResultWin
	case cmp < 0:
		cs.result = GameResultLose
	default:
		cs.result = GameResultDraw
	}

	cs.calculatePayouts()
	cs.evaluateJackpot()

	totalPayout := cs.antePayout + cs.playPayout + cs.jackpotPayout
	if totalPayout > 0 {
		cs.chips.AddChips(totalPayout)
	}

	cs.gameEndFlag = true
	cs.phase = CaribbeanDrawPhaseEnd

	var resultStr string
	switch cs.result {
	case GameResultWin:
		resultStr = "player wins"
	case GameResultDraw:
		resultStr = "push"
	default:
		resultStr = "dealer wins"
	}
	cs.appendLog(-1, "result", resultStr, nil)
}

// compareHands プレイヤーとディーラーのハンドを比較する
func (cs *CaribbeanDraw) compareHands() int {
	if cs.playerHandRank > cs.dealerHandRank {
		return 1
	}
	if cs.playerHandRank < cs.dealerHandRank {
		return -1
	}
	return compareHighCardsSlice(cs.playerHand, cs.dealerHand)
}

// checkDealerQualifies ディーラークオリファイ条件: **8 のペア以上**。
//
// クローン元 (Caribbean Stud) の「ペア以上、または A-K ハイ」より厳しい。
// 共有ヘルパ `dealerQualifies` はスタッドの条件そのものなので使えない ——
// ペアはペアでも 7 以下のペアと A-K ハイは、ここではクオリファイしない。
func (cs *CaribbeanDraw) checkDealerQualifies() bool {
	if cs.dealerHandRank > PokerHandOnePair {
		return true // ツーペア以上は無条件で足りる
	}
	// ペアなら 8 以上。ハイカードにはペアが無いので 0 が返り、自動的に落ちる
	// —— **A-K ハイでもクオリファイしない**のがクローン元との違い。
	return caribbeanDrawPairValue(cs.dealerHand) >= CaribbeanDrawQualifyPairValue
}

// caribbeanDrawPairValue は手札に含まれるペアのランクを返す。ペアが無ければ 0。
//
// **A は 14 として扱う。** 1 のまま比べると A のペアが 8 のペアに負け、最強の
// ペアがクオリファイしないという逆転が起きる。
func caribbeanDrawPairValue(hand []*Card) int {
	freq := make(map[int]int, len(hand))
	for _, c := range hand {
		if c == nil {
			continue
		}
		v := c.GetValue()
		if v == 1 {
			v = 14
		}
		freq[v]++
	}
	best := 0
	for v, n := range freq {
		if n >= 2 && v > best {
			best = v
		}
	}
	return best
}

// calculatePayouts アンテ／プレイの配当計算
func (cs *CaribbeanDraw) calculatePayouts() {
	if !cs.dealerQualified {
		// ディーラー未クオリファイ: アンテ1:1、プレイベットはプッシュ（返却のみ）
		cs.antePayout = cs.anteBet * 2
		cs.playPayout = cs.playBet
		return
	}
	switch cs.result {
	case GameResultWin:
		cs.antePayout = cs.anteBet * 2
		multiplier := cs.playMultiplier()
		cs.playPayout = cs.playBet + cs.playBet*multiplier
	case GameResultDraw:
		cs.antePayout = cs.anteBet
		cs.playPayout = cs.playBet
	case GameResultLose:
		cs.antePayout = 0
		cs.playPayout = 0
	}
}

// playMultiplier プレイベット配当倍率（プレイヤーハンドランクに基づく）
func (cs *CaribbeanDraw) playMultiplier() int {
	switch cs.playerHandRank {
	case PokerHandRoyalFlush:
		return CaribbeanDrawPayRoyalFlush
	case PokerHandStraightFlush:
		return CaribbeanDrawPayStraightFlush
	case PokerHandFourOfAKind:
		return CaribbeanDrawPayFourOfAKind
	case PokerHandFullHouse:
		return CaribbeanDrawPayFullHouse
	case PokerHandFlush:
		return CaribbeanDrawPayFlush
	case PokerHandStraight:
		return CaribbeanDrawPayStraight
	case PokerHandThreeOfAKind:
		return CaribbeanDrawPayThreeOfAKind
	case PokerHandTwoPair:
		return CaribbeanDrawPayTwoPair
	default:
		return CaribbeanDrawPayPair
	}
}

// evaluateJackpot ジャックポットサイドベット評価（独立）
func (cs *CaribbeanDraw) evaluateJackpot() {
	if cs.jackpotBet <= 0 {
		return
	}
	switch cs.playerHandRank {
	case PokerHandRoyalFlush:
		cs.jackpotPayout = cs.jackpotBet * CaribbeanDrawJackpotRoyalFlush
	case PokerHandStraightFlush:
		cs.jackpotPayout = cs.jackpotBet * CaribbeanDrawJackpotStraightFlush
	case PokerHandFourOfAKind:
		cs.jackpotPayout = cs.jackpotBet * CaribbeanDrawJackpotFourOfAKind
	case PokerHandFullHouse:
		cs.jackpotPayout = cs.jackpotBet * CaribbeanDrawJackpotFullHouse
	case PokerHandFlush:
		cs.jackpotPayout = cs.jackpotBet * CaribbeanDrawJackpotFlush
	}
}

// --- Getters ---

// GetPlayerHand プレイヤーハンド取得
func (cs *CaribbeanDraw) GetPlayerHand() []*Card { return cs.playerHand }

// GetDealerHand ディーラーハンド取得
func (cs *CaribbeanDraw) GetDealerHand() []*Card { return cs.dealerHand }

// GetPhase 現在のフェーズ
func (cs *CaribbeanDraw) GetPhase() int { return cs.phase }

// GetGameEndFlag ゲーム終了フラグ
func (cs *CaribbeanDraw) GetGameEndFlag() bool { return cs.gameEndFlag }

// GetAnteBet アンテベット額
func (cs *CaribbeanDraw) GetAnteBet() int { return cs.anteBet }

// GetJackpotBet ジャックポットベット額
func (cs *CaribbeanDraw) GetJackpotBet() int { return cs.jackpotBet }

// GetPlayBet プレイベット額
func (cs *CaribbeanDraw) GetPlayBet() int { return cs.playBet }

// GetResult ゲーム結果
func (cs *CaribbeanDraw) GetResult() GameResult { return cs.result }

// GetAntePayout アンテ配当
func (cs *CaribbeanDraw) GetAntePayout() int { return cs.antePayout }

// GetPlayPayout プレイ配当
func (cs *CaribbeanDraw) GetPlayPayout() int { return cs.playPayout }

// GetJackpotPayout ジャックポット配当
func (cs *CaribbeanDraw) GetJackpotPayout() int { return cs.jackpotPayout }

// GetDrawCost はこのラウンドで払った交換手数料。引かなければ 0。
func (cs *CaribbeanDraw) GetDrawCost() int { return cs.drawCost }

// GetTotalPayout 合計配当
func (cs *CaribbeanDraw) GetTotalPayout() int {
	return cs.antePayout + cs.playPayout + cs.jackpotPayout
}

// GetDealerQualified ディーラークオリファイ
func (cs *CaribbeanDraw) GetDealerQualified() bool { return cs.dealerQualified }

// GetPlayerHandRank プレイヤーハンドランク
func (cs *CaribbeanDraw) GetPlayerHandRank() int { return cs.playerHandRank }

// GetDealerHandRank ディーラーハンドランク
func (cs *CaribbeanDraw) GetDealerHandRank() int { return cs.dealerHandRank }

// GetChips チップ
func (cs *CaribbeanDraw) GetChips() int { return cs.chips.GetChips() }

// --- Test helpers ---

// SetPhase フェーズ設定（テスト用）
func (cs *CaribbeanDraw) SetPhase(phase int) { cs.phase = phase }

// TrumpCardsForTest は山札を返す（テスト用）。ドローで引かれる札を仕込むのに使う。
func (cs *CaribbeanDraw) TrumpCardsForTest() *TrumpCards { return cs.trumpCards }

// SetPlayerHand プレイヤーハンド設定（テスト用）
func (cs *CaribbeanDraw) SetPlayerHand(cards []*Card) { cs.playerHand = cards }

// SetDealerHand ディーラーハンド設定（テスト用）
func (cs *CaribbeanDraw) SetDealerHand(cards []*Card) { cs.dealerHand = cards }

// SetAnteBet アンテベット額設定（テスト用）
func (cs *CaribbeanDraw) SetAnteBet(amount int) { cs.anteBet = amount }

// SetJackpotBet ジャックポットベット額設定（テスト用）
func (cs *CaribbeanDraw) SetJackpotBet(amount int) { cs.jackpotBet = amount }

// SetPlayBet プレイベット額設定（テスト用）
func (cs *CaribbeanDraw) SetPlayBet(amount int) { cs.playBet = amount }

// SetChips チップ設定（テスト用）
func (cs *CaribbeanDraw) SetChips(chips int) { cs.chips.SetChips(chips) }

// caribbeanDrawJSON は CaribbeanDraw の JSON ワイヤーフォーマット
type caribbeanDrawJSON struct {
	TrumpCards      *TrumpCards       `json:"tc"`
	PlayerHand      []*Card           `json:"ph"`
	DealerHand      []*Card           `json:"dh"`
	Chips           *ChipHolder       `json:"ch"`
	AnteBet         int               `json:"ab"`
	JackpotBet      int               `json:"jb"`
	PlayBet         int               `json:"pb"`
	Phase           int               `json:"ps"`
	GameEndFlag     bool              `json:"ge"`
	Result          GameResult        `json:"rs"`
	AntePayout      int               `json:"ap"`
	PlayPayout      int               `json:"plp"`
	JackpotPayout   int               `json:"jp"`
	DrawCost        int               `json:"dc"`
	DealerQualified bool              `json:"dq"`
	PlayerHandRank  int               `json:"pr"`
	DealerHandRank  int               `json:"dr"`
	ActionLog       []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (cs *CaribbeanDraw) MarshalJSON() ([]byte, error) {
	return json.Marshal(caribbeanDrawJSON{
		TrumpCards:      cs.trumpCards,
		PlayerHand:      cs.playerHand,
		DealerHand:      cs.dealerHand,
		Chips:           &cs.chips,
		AnteBet:         cs.anteBet,
		JackpotBet:      cs.jackpotBet,
		PlayBet:         cs.playBet,
		Phase:           cs.phase,
		GameEndFlag:     cs.gameEndFlag,
		Result:          cs.result,
		AntePayout:      cs.antePayout,
		PlayPayout:      cs.playPayout,
		JackpotPayout:   cs.jackpotPayout,
		DrawCost:        cs.drawCost,
		DealerQualified: cs.dealerQualified,
		PlayerHandRank:  cs.playerHandRank,
		DealerHandRank:  cs.dealerHandRank,
		ActionLog:       cs.actionLog,
	})
}

// caribbeanDrawMaxSliceLen caps slice sizes during deserialisation.
const caribbeanDrawMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (cs *CaribbeanDraw) UnmarshalJSON(data []byte) error {
	var j caribbeanDrawJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.PlayerHand) > caribbeanDrawMaxSliceLen || len(j.DealerHand) > caribbeanDrawMaxSliceLen ||
		len(j.ActionLog) > caribbeanDrawMaxSliceLen {
		return fmt.Errorf("caribbeandraw: input array exceeds maximum allowed size")
	}

	cs.trumpCards = j.TrumpCards
	if cs.trumpCards == nil {
		cs.trumpCards = NewTrumpCards(0)
	}
	cs.playerHand = j.PlayerHand
	if cs.playerHand == nil {
		cs.playerHand = make([]*Card, 0)
	}
	cs.dealerHand = j.DealerHand
	if cs.dealerHand == nil {
		cs.dealerHand = make([]*Card, 0)
	}
	if j.Chips != nil {
		cs.chips = *j.Chips
	}
	cs.anteBet = j.AnteBet
	cs.jackpotBet = j.JackpotBet
	cs.playBet = j.PlayBet
	cs.phase = j.Phase
	cs.gameEndFlag = j.GameEndFlag
	cs.result = j.Result
	cs.antePayout = j.AntePayout
	cs.playPayout = j.PlayPayout
	cs.jackpotPayout = j.JackpotPayout
	cs.drawCost = j.DrawCost
	cs.dealerQualified = j.DealerQualified
	cs.playerHandRank = j.PlayerHandRank
	cs.dealerHandRank = j.DealerHandRank
	cs.actionLog = j.ActionLog
	if cs.actionLog == nil {
		cs.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
