//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
	"sort"
)

// ミシシッピ・スタッドフェーズ定数
const (
	MississippiStudPhaseAnte     = 1 // アンティ（ベット）フェーズ
	MississippiStudPhaseThirdSt  = 2 // 3rd Street: 2 枚のホールカード公開後
	MississippiStudPhaseFourthSt = 3 // 4th Street: 1 枚目コミュニティ公開後
	MississippiStudPhaseFifthSt  = 4 // 5th Street: 2 枚目コミュニティ公開後
	MississippiStudPhaseEnd      = 5 // 終了フェーズ
)

// ミシシッピ・スタッドデフォルト値
const (
	MississippiStudDefaultChips  = 1000  // 初期チップ
	MississippiStudMinBet        = 10    // 最低アンティ
	MississippiStudMaxBet        = 10000 // 最大アンティ
	MississippiStudHoleCardCnt   = 2     // ホールカード枚数
	MississippiStudCommunityCnt  = 3     // コミュニティカード枚数
	MississippiStudStreetCnt     = 3     // ストリート数
	MississippiStudMinMultiplier = 1     // 最低ストリートベット倍率
	MississippiStudMaxMultiplier = 3     // 最大ストリートベット倍率
)

// ミシシッピ・スタッド配当倍率 (勝利時 = 投入額 + 投入額 * 倍率 を獲得)
const (
	MississippiStudPayLoss          = 0   // 6 未満ペア / ペアなし: 全ベット没収
	MississippiStudPayPush          = -1  // 6〜10 のペア: ベット返却のみ
	MississippiStudPayHighPair      = 1   // J 以上のペア
	MississippiStudPayTwoPair       = 2   // ツーペア
	MississippiStudPayThreeOfAKind  = 3   // スリーカード
	MississippiStudPayStraight      = 4   // ストレート
	MississippiStudPayFlush         = 6   // フラッシュ
	MississippiStudPayFullHouse     = 10  // フルハウス
	MississippiStudPayFourOfAKind   = 40  // フォーカード
	MississippiStudPayStraightFlush = 100 // ストレートフラッシュ
	MississippiStudPayRoyalFlush    = 500 // ロイヤルフラッシュ
)

// mississippiStudMinRoundCost は Reset 時にチップ不足とみなす閾値倍率。
// アンティ (1x) + 3 ストリート最低各 1x = 4x のチップが必要。
const mississippiStudMinRoundCost = 4

// MississippiStud ミシシッピ・スタッド本体
type MississippiStud struct {
	trumpCards        *TrumpCards
	playerHand        []*Card // ホールカード 2 枚
	communityCards    []*Card // コミュニティカード 3 枚 (公開状態は communityRevealed)
	communityRevealed [MississippiStudCommunityCnt]bool
	chips             ChipHolder
	anteAmount        int                           // 1 ユニットあたりのアンティ額
	streetMultipliers [MississippiStudStreetCnt]int // 0=未ベット, 1/2/3=ストリートベット倍率
	folded            bool                          // フォールドフラグ
	phase             int                           // 現在のフェーズ
	gameEndFlag       bool                          // ゲーム終了フラグ
	result            GameResult                    // ゲーム結果
	handRank          int                           // 最終ハンドランク
	payoutMultiplier  int                           // 適用された配当倍率 (-1=プッシュ, 0=ロス)
	antePayout        int                           // アンティ部分の配当
	streetPayouts     [MississippiStudStreetCnt]int // ストリートベット部分の配当
	totalPayout       int                           // 合計配当
	actionLog         []*ActionLogEntry             // 棋譜
}

// NewMississippiStud コンストラクタ
func NewMississippiStud(trumpCards *TrumpCards) *MississippiStud {
	trumpCards.Shuffle()
	return &MississippiStud{
		trumpCards: trumpCards,
		phase:      MississippiStudPhaseAnte,
	}
}

// NewDefaultMississippiStud デフォルト設定でミシシッピ・スタッドを生成する
func NewDefaultMississippiStud() *MississippiStud {
	m := NewMississippiStud(NewTrumpCards(0))
	m.chips.SetChips(MississippiStudDefaultChips)
	return m
}

// Reset ゲーム初期化
func (m *MississippiStud) Reset() {
	m.gameEndFlag = false
	m.phase = MississippiStudPhaseAnte
	m.playerHand = nil
	m.communityCards = nil
	m.communityRevealed = [MississippiStudCommunityCnt]bool{}
	m.anteAmount = 0
	m.streetMultipliers = [MississippiStudStreetCnt]int{}
	m.folded = false
	m.result = 0
	m.handRank = 0
	m.payoutMultiplier = 0
	m.antePayout = 0
	m.streetPayouts = [MississippiStudStreetCnt]int{}
	m.totalPayout = 0
	m.actionLog = nil
	if m.chips.GetChips() < MississippiStudMinBet*mississippiStudMinRoundCost {
		m.chips.SetChips(MississippiStudDefaultChips)
	}
	m.trumpCards = NewTrumpCards(0)
	m.trumpCards.Shuffle()
}

// Bet ベット (アンティ) を置きカードを配る。
func (m *MississippiStud) Bet(amount int) error {
	if m.phase != MississippiStudPhaseAnte {
		return NewDomainError(ErrWrongPhase, "Bet is only allowed during the ante phase.")
	}
	if amount < MississippiStudMinBet || amount%MississippiStudMinBet != 0 || amount > MississippiStudMaxBet {
		return NewDomainError(ErrInvalidAmount, "Invalid bet amount.")
	}
	if !m.chips.SubtractChips(amount) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}
	m.anteAmount = amount
	m.appendLog(0, "ante", fmt.Sprintf("ante=%d", amount), nil)
	m.deal()
	m.phase = MississippiStudPhaseThirdSt
	return nil
}

// Play 1x/2x/3x ストリートベットを置きフェーズを進める。
// 5th Street での Play は最終解決を行う。
func (m *MississippiStud) Play(multiplier int) error {
	streetIdx, err := m.streetIndexForPlay()
	if err != nil {
		return err
	}
	if multiplier < MississippiStudMinMultiplier || multiplier > MississippiStudMaxMultiplier {
		return NewDomainError(ErrInvalidAmount, "Multiplier must be 1, 2, or 3.")
	}
	cost := m.anteAmount * multiplier
	if !m.chips.SubtractChips(cost) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}
	m.streetMultipliers[streetIdx] = multiplier
	m.communityRevealed[streetIdx] = true
	m.appendLog(0, "play", fmt.Sprintf("street=%d x%d (-%d chips)", streetIdx+3, multiplier, cost), nil)
	switch m.phase {
	case MississippiStudPhaseThirdSt:
		m.phase = MississippiStudPhaseFourthSt
	case MississippiStudPhaseFourthSt:
		m.phase = MississippiStudPhaseFifthSt
	case MississippiStudPhaseFifthSt:
		m.resolve()
	}
	return nil
}

// Fold 現時点までのベットをすべて没収しゲームを終了する。
func (m *MississippiStud) Fold() error {
	if _, err := m.streetIndexForPlay(); err != nil {
		return err
	}
	m.folded = true
	m.appendLog(0, "fold", "player folds", nil)
	m.resolve()
	return nil
}

// streetIndexForPlay 現フェーズで操作可能なストリートインデックスを返す。
func (m *MississippiStud) streetIndexForPlay() (int, error) {
	switch m.phase {
	case MississippiStudPhaseThirdSt:
		return 0, nil
	case MississippiStudPhaseFourthSt:
		return 1, nil
	case MississippiStudPhaseFifthSt:
		return 2, nil
	default:
		return -1, NewDomainError(ErrWrongPhase, "Play/Fold is only allowed during decision phases.")
	}
}

// deal ホールカード 2 枚とコミュニティカード 3 枚を配る (コミュニティは全て伏せ)。
func (m *MississippiStud) deal() {
	m.playerHand = make([]*Card, 0, MississippiStudHoleCardCnt)
	m.communityCards = make([]*Card, 0, MississippiStudCommunityCnt)
	for range MississippiStudHoleCardCnt {
		m.playerHand = append(m.playerHand, m.trumpCards.DrawCard())
	}
	for range MississippiStudCommunityCnt {
		m.communityCards = append(m.communityCards, m.trumpCards.DrawCard())
	}
	m.appendLog(-1, "deal", "dealt 2 hole + 3 community cards", nil)
}

// resolve ゲーム解決。フォールド or 5th Street プレイ後に呼ばれる。
func (m *MississippiStud) resolve() {
	m.gameEndFlag = true
	m.phase = MississippiStudPhaseEnd

	if m.folded {
		m.result = GameResultLose
		m.payoutMultiplier = MississippiStudPayLoss
		m.appendLog(-1, "result", fmt.Sprintf("folded; wager=%d lost", m.GetTotalBet()), nil)
		return
	}

	for i := range m.communityRevealed {
		m.communityRevealed[i] = true
	}
	fullHand := make([]*Card, 0, MississippiStudHoleCardCnt+MississippiStudCommunityCnt)
	fullHand = append(fullHand, m.playerHand...)
	fullHand = append(fullHand, m.communityCards...)
	m.handRank = evalFiveCardHand(fullHand)
	m.payoutMultiplier = mississippiStudPayoutMultiplier(m.handRank, fullHand)

	switch {
	case m.payoutMultiplier > 0:
		m.result = GameResultWin
		m.antePayout = m.anteAmount + m.anteAmount*m.payoutMultiplier
		for i, mult := range m.streetMultipliers {
			bet := m.anteAmount * mult
			m.streetPayouts[i] = bet + bet*m.payoutMultiplier
		}
	case m.payoutMultiplier == MississippiStudPayPush:
		m.result = GameResultDraw
		m.antePayout = m.anteAmount
		for i, mult := range m.streetMultipliers {
			m.streetPayouts[i] = m.anteAmount * mult
		}
	default:
		m.result = GameResultLose
	}

	m.totalPayout = m.antePayout
	for _, p := range m.streetPayouts {
		m.totalPayout += p
	}
	if m.totalPayout > 0 {
		m.chips.AddChips(m.totalPayout)
	}

	resultStr := "player loses"
	switch m.result {
	case GameResultWin:
		resultStr = "player wins"
	case GameResultDraw:
		resultStr = "push"
	}
	m.appendLog(-1, "result", fmt.Sprintf("%s rank=%d total=%d", resultStr, m.handRank, m.totalPayout), nil)
}

// mississippiStudPayoutMultiplier は最終 5 枚役と内訳から配当倍率を返す。
// -1 = プッシュ (6〜10 ペア), 0 = 没収 (6 未満ペア / ペアなし), 1+ = 勝利倍率。
func mississippiStudPayoutMultiplier(rank int, hand []*Card) int {
	switch rank {
	case PokerHandRoyalFlush:
		return MississippiStudPayRoyalFlush
	case PokerHandStraightFlush:
		return MississippiStudPayStraightFlush
	case PokerHandFourOfAKind:
		return MississippiStudPayFourOfAKind
	case PokerHandFullHouse:
		return MississippiStudPayFullHouse
	case PokerHandFlush:
		return MississippiStudPayFlush
	case PokerHandStraight:
		return MississippiStudPayStraight
	case PokerHandThreeOfAKind:
		return MississippiStudPayThreeOfAKind
	case PokerHandTwoPair:
		return MississippiStudPayTwoPair
	case PokerHandOnePair:
		return mississippiStudPairTier(hand)
	default:
		return MississippiStudPayLoss
	}
}

// MississippiStudMadeHand は現在の手役と、それが配当表に載るかどうか。
type MississippiStudMadeHand struct {
	// Rank は既知のカードから作れる最良の役 (PokerHand*)。
	Rank int
	// PaytableEligible は配当表の対象か。6以上のペアかエースのペア、
	// またはワンペアより上の役。
	PaytableEligible bool
}

// 推奨アクション (#4710)。
const (
	// MSRecommendPlay3x 3倍を置く。
	MSRecommendPlay3x = "play3x"
	// MSRecommendPlay1x 1倍を置く。
	MSRecommendPlay1x = "play1x"
	// MSRecommendFold 降りる。
	MSRecommendFold = "fold"
)

// knownCards はホールカードと**公開済みの**コミュニティカードを返す。
//
// **未公開のカードを混ぜない。**混ぜると、プレイヤーがまだ見ていない札を
// 根拠にした助言になる。
func (m *MississippiStud) knownCards() []*Card {
	cards := append([]*Card{}, m.playerHand...)
	for i, c := range m.communityCards {
		if i < len(m.communityRevealed) && m.communityRevealed[i] {
			cards = append(cards, c)
		}
	}
	return cards
}

// GetCurrentMadeHand は既知のカードからできている役を返す。2枚未満、または
// ハイカードどまりのときは nil。
//
// **Web は ms-made-hand に役と配当対象かを常時出しているのに、CUI には
// どちらも無かった (#4710)。**
func (m *MississippiStud) GetCurrentMadeHand() *MississippiStudMadeHand {
	cards := m.knownCards()
	if len(cards) < 2 {
		return nil
	}
	rank := mississippiStudBestRank(cards)
	if rank <= PokerHandHighCard {
		return nil
	}
	if rank > PokerHandOnePair {
		return &MississippiStudMadeHand{Rank: rank, PaytableEligible: true}
	}
	return &MississippiStudMadeHand{
		Rank:             rank,
		PaytableEligible: mississippiStudPairTier(cards) != MississippiStudPayLoss,
	}
}

// mississippiStudBestRank は既知のカードから作れる最良の役を返す。
// 5枚に満たないときは同ランクの重なりだけで判定する (フラッシュや
// ストレートは5枚そろわないと成立しない)。
func mississippiStudBestRank(cards []*Card) int {
	if len(cards) >= 5 {
		best := PokerHandHighCard
		for _, combo := range combinations(cards, 5) {
			if r := evalFiveCardHand(combo); r > best {
				best = r
			}
		}
		return best
	}
	counts := map[int]int{}
	for _, c := range cards {
		counts[c.GetValue()]++
	}
	maxCount, pairs := 0, 0
	for _, n := range counts {
		if n > maxCount {
			maxCount = n
		}
		if n >= 2 {
			pairs++
		}
	}
	switch {
	case maxCount >= 4:
		return PokerHandFourOfAKind
	case maxCount == 3:
		return PokerHandThreeOfAKind
	case pairs >= 2:
		return PokerHandTwoPair
	case maxCount == 2:
		return PokerHandOnePair
	default:
		return PokerHandHighCard
	}
}

// RecommendBet は現在のストリートでの推奨アクションを返す。判断のいらない
// フェーズでは空文字。
//
// **判定はフロントの getMississippiStudHint と同じ規則。**配当対象の役が
// できていれば 3x、まともなドローがあれば 1x、それ以外は降りる。ずれると
// 同じ局面で CUI と Web が違う倍率を指す。
func (m *MississippiStud) RecommendBet() string {
	switch m.phase {
	case MississippiStudPhaseThirdSt, MississippiStudPhaseFourthSt, MississippiStudPhaseFifthSt:
	default:
		return ""
	}
	cards := m.knownCards()
	if len(cards) == 0 {
		return ""
	}
	if made := m.GetCurrentMadeHand(); made != nil && made.PaytableEligible {
		return MSRecommendPlay3x
	}
	if mississippiStudHasReasonableDraw(cards, m.phase) {
		return MSRecommendPlay1x
	}
	return MSRecommendFold
}

// mississippiStudHasReasonableDraw はフラッシュ/ストレートのドロー、または
// 3rd street でのハイカード2枚を「まだ賭ける価値がある」と見なす。
func mississippiStudHasReasonableDraw(cards []*Card, phase int) bool {
	slotsLeft := (MississippiStudHoleCardCnt + MississippiStudCommunityCnt) - len(cards)
	if mississippiStudHasFlushDraw(cards, slotsLeft) || mississippiStudHasStraightDraw(cards, slotsLeft) {
		return true
	}
	if phase != MississippiStudPhaseThirdSt {
		return false
	}
	high := 0
	for _, c := range cards {
		if v := c.GetValue(); v == 1 || v >= 10 {
			high++
		}
	}
	return high >= 2
}

// mississippiStudHasFlushDraw は残り枚数でフラッシュが間に合うかを返す。
func mississippiStudHasFlushDraw(cards []*Card, slotsLeft int) bool {
	bySuit := map[int]int{}
	for _, c := range cards {
		bySuit[c.GetDesign()]++
	}
	for _, n := range bySuit {
		if n >= 3 && n+slotsLeft >= (MississippiStudHoleCardCnt+MississippiStudCommunityCnt) {
			return true
		}
	}
	return false
}

// mississippiStudHasStraightDraw は残り枚数でストレートが間に合うかを返す。
// **エースは 1 と 14 の両方で数える。**A-2-3-4-5 も 10-J-Q-K-A も成立する。
func mississippiStudHasStraightDraw(cards []*Card, slotsLeft int) bool {
	set := map[int]bool{}
	for _, c := range cards {
		v := c.GetValue()
		set[v] = true
		if v == 1 {
			set[14] = true
		}
	}
	values := make([]int, 0, len(set))
	for v := range set {
		values = append(values, v)
	}
	sort.Ints(values)
	run := 1
	for i := 1; i < len(values); i++ {
		if values[i] == values[i-1]+1 {
			run++
		} else {
			run = 1
		}
		if run+slotsLeft >= (MississippiStudHoleCardCnt + MississippiStudCommunityCnt) {
			return true
		}
	}
	return false
}

// mississippiStudPairTier はペア構成のティアを返す。
func mississippiStudPairTier(hand []*Card) int {
	counts := make(map[int]int)
	for _, c := range hand {
		counts[c.GetValue()]++
	}
	for val, cnt := range counts {
		if cnt < 2 {
			continue
		}
		// J(11), Q(12), K(13), A(1) のペア → 配当
		if val >= 11 || val == 1 {
			return MississippiStudPayHighPair
		}
		// 6〜10 のペア → プッシュ
		if val >= 6 && val <= 10 {
			return MississippiStudPayPush
		}
	}
	return MississippiStudPayLoss
}

// appendLog 棋譜にエントリを追加する
func (m *MississippiStud) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	m.actionLog = append(m.actionLog, &ActionLogEntry{
		TurnNumber: len(m.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- Getters ---

// GetPlayerHand ホールカードを取得する。
func (m *MississippiStud) GetPlayerHand() []*Card { return m.playerHand }

// GetCommunityCards コミュニティカードを取得する (公開状態は GetCommunityRevealed で参照)。
func (m *MississippiStud) GetCommunityCards() []*Card { return m.communityCards }

// GetCommunityRevealed コミュニティカードそれぞれの公開状態を返す。
func (m *MississippiStud) GetCommunityRevealed() [MississippiStudCommunityCnt]bool {
	return m.communityRevealed
}

// GetPhase 現在のフェーズを取得する。
func (m *MississippiStud) GetPhase() int { return m.phase }

// GetGameEndFlag ゲーム終了フラグを取得する。
func (m *MississippiStud) GetGameEndFlag() bool { return m.gameEndFlag }

// GetAnteAmount アンティ額を取得する。
func (m *MississippiStud) GetAnteAmount() int { return m.anteAmount }

// GetStreetMultipliers ストリートベット倍率を取得する。
func (m *MississippiStud) GetStreetMultipliers() [MississippiStudStreetCnt]int {
	return m.streetMultipliers
}

// GetFolded フォールドフラグを取得する。
func (m *MississippiStud) GetFolded() bool { return m.folded }

// GetTotalBet 現在までに投じたチップ総量を返す (アンティ + 各ストリートのベット)。
func (m *MississippiStud) GetTotalBet() int {
	total := m.anteAmount
	for _, mult := range m.streetMultipliers {
		total += m.anteAmount * mult
	}
	return total
}

// GetResult ゲーム結果を取得する。
func (m *MississippiStud) GetResult() GameResult { return m.result }

// GetHandRank 最終ハンドランクを取得する。
func (m *MississippiStud) GetHandRank() int { return m.handRank }

// GetPayoutMultiplier 適用された配当倍率を取得する (-1=プッシュ, 0=ロス, 1+=勝利)。
func (m *MississippiStud) GetPayoutMultiplier() int { return m.payoutMultiplier }

// GetAntePayout アンティ部分の配当を取得する。
func (m *MississippiStud) GetAntePayout() int { return m.antePayout }

// GetStreetPayouts ストリート部分の配当を取得する。
func (m *MississippiStud) GetStreetPayouts() [MississippiStudStreetCnt]int {
	return m.streetPayouts
}

// GetTotalPayout 合計配当を取得する。
func (m *MississippiStud) GetTotalPayout() int { return m.totalPayout }

// GetChips チップ残高を取得する。
func (m *MississippiStud) GetChips() int { return m.chips.GetChips() }

// GetActionLog 棋譜を取得する。
func (m *MississippiStud) GetActionLog() []*ActionLogEntry { return m.actionLog }

// --- Test helpers ---

// SetPhase フェーズ設定（テスト用）。
func (m *MississippiStud) SetPhase(phase int) { m.phase = phase }

// SetPlayerHand プレイヤーハンド設定（テスト用）。
func (m *MississippiStud) SetPlayerHand(cards []*Card) { m.playerHand = cards }

// SetCommunityCards コミュニティカード設定（テスト用）。
func (m *MississippiStud) SetCommunityCards(cards []*Card) { m.communityCards = cards }

// SetCommunityRevealed コミュニティ公開状態設定（テスト用）。
func (m *MississippiStud) SetCommunityRevealed(revealed [MississippiStudCommunityCnt]bool) {
	m.communityRevealed = revealed
}

// SetAnteAmount アンティ額設定（テスト用）。
func (m *MississippiStud) SetAnteAmount(amount int) { m.anteAmount = amount }

// SetStreetMultipliers ストリートベット倍率設定（テスト用）。
func (m *MississippiStud) SetStreetMultipliers(mults [MississippiStudStreetCnt]int) {
	m.streetMultipliers = mults
}

// SetFolded フォールド設定（テスト用）。
func (m *MississippiStud) SetFolded(folded bool) { m.folded = folded }

// SetChips チップ設定（テスト用）。
func (m *MississippiStud) SetChips(chips int) { m.chips.SetChips(chips) }

// mississippiStudJSON は JSON ワイヤーフォーマット。
type mississippiStudJSON struct {
	TrumpCards        *TrumpCards                       `json:"tc"`
	PlayerHand        []*Card                           `json:"ph"`
	CommunityCards    []*Card                           `json:"cc"`
	CommunityRevealed [MississippiStudCommunityCnt]bool `json:"cr"`
	Chips             *ChipHolder                       `json:"ch"`
	AnteAmount        int                               `json:"ba"`
	StreetMultipliers [MississippiStudStreetCnt]int     `json:"sm"`
	Folded            bool                              `json:"fd"`
	Phase             int                               `json:"ps"`
	GameEndFlag       bool                              `json:"ge"`
	Result            GameResult                        `json:"rs"`
	HandRank          int                               `json:"hr"`
	PayoutMultiplier  int                               `json:"pm"`
	AntePayout        int                               `json:"ap"`
	StreetPayouts     [MississippiStudStreetCnt]int     `json:"sp"`
	TotalPayout       int                               `json:"tp"`
	ActionLog         []*ActionLogEntry                 `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (m *MississippiStud) MarshalJSON() ([]byte, error) {
	return json.Marshal(mississippiStudJSON{
		TrumpCards:        m.trumpCards,
		PlayerHand:        m.playerHand,
		CommunityCards:    m.communityCards,
		CommunityRevealed: m.communityRevealed,
		Chips:             &m.chips,
		AnteAmount:        m.anteAmount,
		StreetMultipliers: m.streetMultipliers,
		Folded:            m.folded,
		Phase:             m.phase,
		GameEndFlag:       m.gameEndFlag,
		Result:            m.result,
		HandRank:          m.handRank,
		PayoutMultiplier:  m.payoutMultiplier,
		AntePayout:        m.antePayout,
		StreetPayouts:     m.streetPayouts,
		TotalPayout:       m.totalPayout,
		ActionLog:         m.actionLog,
	})
}

// mississippiStudMaxSliceLen はデシリアライズ時のスライス長上限。
const mississippiStudMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (m *MississippiStud) UnmarshalJSON(data []byte) error {
	var j mississippiStudJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.PlayerHand) > mississippiStudMaxSliceLen ||
		len(j.CommunityCards) > mississippiStudMaxSliceLen ||
		len(j.ActionLog) > mississippiStudMaxSliceLen {
		return fmt.Errorf("mississippistud: input array exceeds maximum allowed size")
	}

	m.trumpCards = j.TrumpCards
	if m.trumpCards == nil {
		m.trumpCards = NewTrumpCards(0)
	}
	m.playerHand = j.PlayerHand
	if m.playerHand == nil {
		m.playerHand = make([]*Card, 0)
	}
	m.communityCards = j.CommunityCards
	if m.communityCards == nil {
		m.communityCards = make([]*Card, 0)
	}
	m.communityRevealed = j.CommunityRevealed
	if j.Chips != nil {
		m.chips = *j.Chips
	}
	m.anteAmount = j.AnteAmount
	m.streetMultipliers = j.StreetMultipliers
	m.folded = j.Folded
	m.phase = j.Phase
	m.gameEndFlag = j.GameEndFlag
	m.result = j.Result
	m.handRank = j.HandRank
	m.payoutMultiplier = j.PayoutMultiplier
	m.antePayout = j.AntePayout
	m.streetPayouts = j.StreetPayouts
	m.totalPayout = j.TotalPayout
	m.actionLog = j.ActionLog
	if m.actionLog == nil {
		m.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
