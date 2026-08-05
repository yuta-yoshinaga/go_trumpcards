//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"fmt"
	"sort"
)

// ファロのフェーズ定数（フロントエンドが数値フェーズを使用するため int）。
const (
	// FaroPhaseBetting はベットフェーズ（プレイヤーがレイアウトにチップを置ける）。
	FaroPhaseBetting = 1
	// FaroPhaseTurn はターン進行フェーズ（2枚ずつめくって決着）。
	FaroPhaseTurn = 2
	// FaroPhaseCall はコール（最後の3枚の順序予想）フェーズ。
	FaroPhaseCall = 3
	// FaroPhaseRoundEnd はディール終了フェーズ（NextRound 待ち）。
	FaroPhaseRoundEnd = 4
	// FaroPhaseGameEnd はゲーム終了フェーズ（チップ切れ）。
	FaroPhaseGameEnd = 5
)

// ファロのルール定数。
const (
	// FaroDeckSize は使用デッキ枚数（標準52枚）。
	FaroDeckSize = 52
	// FaroSodaBurn はディール開始時に焼く「ソーダ」枚数。
	FaroSodaBurn = 1
	// FaroTurnCards は1ターンでめくる枚数（敗北札+勝利札）。
	FaroTurnCards = 2
	// FaroTurnsPerDeal は1ディールのターン数。1(soda)+24*2(=48)+3(call)=52。
	FaroTurnsPerDeal = 24
	// FaroCallCards はコールで予想する残り枚数。
	FaroCallCards = 3
	// FaroCallPayoutMultiplier はコール成功時の配当倍率（4:1）。
	FaroCallPayoutMultiplier = 4
	// FaroMinRank / FaroMaxRank はレイアウトのランク範囲（A=1..K=13）。
	FaroMinRank = 1
	FaroMaxRank = 13
	// faroMaxSliceLen はデシリアライズ時のスライス/マップ長の上限。
	faroMaxSliceLen = 1000
)

// FaroBet は1ランクへのベット（金額とカッパー指定）を表す。
// Copper が true の場合、そのランクが「敗北」する側に賭けたことを意味する。
type FaroBet struct {
	Amount int  `json:"am"`
	Copper bool `json:"cp"`
}

// FaroTurnResult は1ターンの結果（敗北札・勝利札と純損益）を表す。
type FaroTurnResult struct {
	LosingCard  *Card `json:"lc"`
	WinningCard *Card `json:"wc"`
	Split       bool  `json:"sp"`
	Net         int   `json:"nt"` // そのターンのプレイヤー純損益（正=利得）。
}

// Faro はファロ（19世紀アメリカの銀行賭博ゲーム）の本体。
//
// ルール（忠実だが有限・決定的）:
//   - 単一の人間プレイヤー対バンク。52枚デッキ。
//   - プレイヤーは13ランク（A..K）のレイアウトにチップを置く。カッパーを付けると
//     「そのランクが敗北する」方に賭ける。
//   - ディール = シャッフル → ソーダ1枚を焼く → 2枚1組のターンを24回（48枚）→
//     残り3枚はコール用。合計 1 + 48 + 3 = 52 枚。
//   - 各ターン: 1枚目=敗北札（バンカー側、そのランクの素ベットは没収・カッパーは1:1配当）、
//     2枚目=勝利札（プレイヤー側、そのランクの素ベットは1:1配当・カッパーは没収）。
//   - スプリット（1枚目と2枚目が同ランク）: バンカーがそのランクのベットの半額を取る
//     （切り捨て）。残った半額はレイアウトに残る。
//   - ベットは解決またはクリアされるまでレイアウト上に残り、毎ターン再利用される。
//   - コール: 残り3枚のとき、順序を正しく当てると4:1（元金は返却）。
//   - ディール終了（コール後）または24ターン消化でラウンド終了。チップ切れでゲーム終了。
type Faro struct {
	config      FaroConfig
	trumpCards  *TrumpCards
	chips       ChipHolder
	bets        map[int]*FaroBet // rank(1..13) -> bet
	soda        *Card
	turnsPlayed int
	lastTurn    *FaroTurnResult
	callOrder   []int // プレイヤーが宣言したコール順（ランク）。
	callCards   []*Card
	callWon     bool
	phase       int
	gameEndFlag bool
	totalPayout int // 直近のラウンドのプレイヤー純損益（正=利得）。
	actionLog   []*ActionLogEntry
}

// NewFaro はトランプデッキを受け取りファロを生成する。
func NewFaro(trumpCards *TrumpCards) *Faro {
	return NewFaroWithConfig(trumpCards, DefaultFaroConfig())
}

// NewFaroWithConfig は設定付きでファロを生成する。
func NewFaroWithConfig(trumpCards *TrumpCards, config FaroConfig) *Faro {
	if err := config.Validate(); err != nil {
		config = DefaultFaroConfig()
	}
	f := &Faro{
		config:     config,
		trumpCards: trumpCards,
		bets:       make(map[int]*FaroBet),
		phase:      FaroPhaseBetting,
	}
	f.chips.SetChips(config.StartChips)
	f.trumpCards.Shuffle()
	return f
}

// NewDefaultFaro はデフォルト設定のファロを生成するファクトリ関数。
func NewDefaultFaro() *Faro {
	return NewFaroWithConfig(NewTrumpCards(0), DefaultFaroConfig())
}

// Reset はゲームを初期化する（チップが最低ベット未満なら開始チップに戻す）。
func (f *Faro) Reset() {
	if f.chips.GetChips() < f.config.MinBet {
		f.chips.SetChips(f.config.StartChips)
	}
	f.startDeal()
	f.gameEndFlag = false
	f.phase = FaroPhaseBetting
}

// NextRound は次のディールを開始する。チップ切れならゲーム終了にする。
func (f *Faro) NextRound() {
	if f.chips.GetChips() < f.config.MinBet {
		f.gameEndFlag = true
		f.phase = FaroPhaseGameEnd
		f.appendLog(-1, "gameEnd", "player out of chips", nil)
		return
	}
	f.startDeal()
	f.phase = FaroPhaseBetting
}

// startDeal は新しいデッキを用意し、ソーダを焼いてベットフェーズへ移る。
func (f *Faro) startDeal() {
	f.trumpCards = NewTrumpCards(0)
	for range 10 {
		f.trumpCards.Shuffle()
	}
	f.bets = make(map[int]*FaroBet)
	f.turnsPlayed = 0
	f.lastTurn = nil
	f.callOrder = nil
	f.callCards = nil
	f.callWon = false
	f.totalPayout = 0
	f.actionLog = nil
	f.soda = f.trumpCards.DrawCard() // ソーダ1枚を焼く。
	f.appendLog(-1, "soda", "burned soda card", []*Card{f.soda})
}

// totalBet はレイアウト上の全ベット金額の合計を返す。
func (f *Faro) totalBet() int {
	total := 0
	for _, b := range f.bets {
		total += b.Amount
	}
	return total
}

// PlayerPlaceBet はランク rank（1..13）に金額 amount のベットを置く（既存ベットに上書き）。
// copper が true の場合は「そのランクが敗北する」方に賭ける。
func (f *Faro) PlayerPlaceBet(rank, amount int, copper bool) error {
	if f.phase != FaroPhaseBetting && f.phase != FaroPhaseTurn {
		return NewDomainError(ErrWrongPhase, "Betting is only allowed during the betting or turn phase.")
	}
	if rank < FaroMinRank || rank > FaroMaxRank {
		return NewDomainError(ErrInvalidPlay, "Bet rank must be between 1 (A) and 13 (K).")
	}
	if amount < f.config.MinBet || amount%f.config.MinBet != 0 || amount > f.config.MaxBet {
		return NewDomainError(ErrInvalidAmount, "Invalid bet amount.")
	}
	// 既存ベットを返金してから新しい金額を差し引く（上書きセマンティクス）。
	prev := 0
	if b, ok := f.bets[rank]; ok {
		prev = b.Amount
	}
	// 必要な追加分のみ検証。
	delta := amount - prev
	if delta > 0 && !f.chips.SubtractChips(delta) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}
	if delta < 0 {
		f.chips.AddChips(-delta)
	}
	f.bets[rank] = &FaroBet{Amount: amount, Copper: copper}
	f.appendLog(0, "bet", fmt.Sprintf("rank=%d amount=%d copper=%t", rank, amount, copper), nil)
	return nil
}

// PlayerClearBet はランク rank のベットを取り消し、チップを返金する。
func (f *Faro) PlayerClearBet(rank int) error {
	if f.phase != FaroPhaseBetting && f.phase != FaroPhaseTurn {
		return NewDomainError(ErrWrongPhase, "Clearing a bet is only allowed during the betting or turn phase.")
	}
	if b, ok := f.bets[rank]; ok {
		f.chips.AddChips(b.Amount)
		delete(f.bets, rank)
		f.appendLog(0, "clearBet", fmt.Sprintf("rank=%d", rank), nil)
	}
	return nil
}

// PlayerClearAll はすべてのベットを取り消し、チップを返金する。
func (f *Faro) PlayerClearAll() error {
	if f.phase != FaroPhaseBetting && f.phase != FaroPhaseTurn {
		return NewDomainError(ErrWrongPhase, "Clearing bets is only allowed during the betting or turn phase.")
	}
	for rank, b := range f.bets {
		f.chips.AddChips(b.Amount)
		delete(f.bets, rank)
	}
	f.appendLog(0, "clearAll", "cleared all bets", nil)
	return nil
}

// PlayerDealTurn は2枚（敗北札・勝利札）をめくり、レイアウト上のベットを解決する。
func (f *Faro) PlayerDealTurn() error {
	if f.phase != FaroPhaseBetting && f.phase != FaroPhaseTurn {
		return NewDomainError(ErrWrongPhase, "Dealing a turn is only allowed during the betting or turn phase.")
	}
	if f.turnsPlayed >= FaroTurnsPerDeal {
		return NewDomainError(ErrWrongPhase, "All turns have been dealt; the call is now available.")
	}
	if f.trumpCards.GetRemainingCount() < FaroTurnCards+FaroCallCards {
		return NewDomainError(ErrDeckExhausted, "Not enough cards remain to deal a turn.")
	}
	losing := f.trumpCards.DrawCard()
	winning := f.trumpCards.DrawCard()
	if losing == nil || winning == nil {
		return NewDomainError(ErrDeckExhausted, "Deck exhausted.")
	}
	f.turnsPlayed++
	result := &FaroTurnResult{LosingCard: losing, WinningCard: winning}
	losingRank := rankOf1to13(losing)
	winningRank := rankOf1to13(winning)
	result.Split = losingRank == winningRank
	result.Net = f.resolveTurn(losingRank, winningRank, result.Split)
	f.lastTurn = result
	f.totalPayout += result.Net
	f.appendLog(-1, "turn", fmt.Sprintf("losing=%d winning=%d split=%t net=%d", losingRank, winningRank, result.Split, result.Net), []*Card{losing, winning})

	if f.turnsPlayed >= FaroTurnsPerDeal {
		f.phase = FaroPhaseCall
		f.drawCallCards()
		f.appendLog(-1, "callReady", "three cards remain; call is available", nil)
	} else {
		f.phase = FaroPhaseTurn
	}
	return nil
}

// resolveTurn は1ターン分のベット解決を行い、プレイヤーの純損益を返す。
func (f *Faro) resolveTurn(losingRank, winningRank int, split bool) int {
	net := 0
	if split {
		// スプリット: バンカーがそのランクのベットの半額を取る（切り捨て）。残りは残留。
		if b, ok := f.bets[losingRank]; ok {
			half := b.Amount / 2
			b.Amount -= half
			net -= half
			if b.Amount == 0 {
				delete(f.bets, losingRank)
			}
		}
		return net
	}
	// 敗北札の解決。
	if b, ok := f.bets[losingRank]; ok {
		if b.Copper {
			// カッパー（敗北予想）が的中 → 1:1 配当。元金もレイアウトに残らず精算して返す。
			f.chips.AddChips(b.Amount * 2)
			net += b.Amount
		} else {
			// 素ベットは敗北 → 没収。
			net -= b.Amount
		}
		delete(f.bets, losingRank)
	}
	// 勝利札の解決。
	if b, ok := f.bets[winningRank]; ok {
		if b.Copper {
			// カッパーは勝利札では負け → 没収。
			net -= b.Amount
		} else {
			// 素ベットは勝利 → 1:1 配当。
			f.chips.AddChips(b.Amount * 2)
			net += b.Amount
		}
		delete(f.bets, winningRank)
	}
	return net
}

// drawCallCards はコール対象の最後の3枚を順序どおりにドローして保持する（冪等）。
func (f *Faro) drawCallCards() {
	if len(f.callCards) == FaroCallCards {
		return
	}
	f.callCards = make([]*Card, 0, FaroCallCards)
	for range FaroCallCards {
		c := f.trumpCards.DrawCard()
		if c == nil {
			break
		}
		f.callCards = append(f.callCards, c)
	}
}

// PlayerCall は最後の3枚の順序（ランクの並び）を予想する。
// 正しければ 4:1（元金返却）でラウンド終了する。誤りなら没収してラウンド終了する。
// order が空（長さ0）の場合はコールを見送ってラウンドを終了する。
func (f *Faro) PlayerCall(order []int) error {
	if f.phase != FaroPhaseCall {
		return NewDomainError(ErrWrongPhase, "Call is only allowed when three cards remain.")
	}
	f.drawCallCards()
	if len(order) == 0 {
		// コールを見送る。
		f.callWon = false
		f.callOrder = nil
		f.phase = FaroPhaseRoundEnd
		f.appendLog(0, "call", "skipped call", nil)
		return nil
	}
	if len(order) != FaroCallCards {
		return NewDomainError(ErrInvalidIndices, "Call order must list exactly three ranks.")
	}
	for _, r := range order {
		if r < FaroMinRank || r > FaroMaxRank {
			return NewDomainError(ErrInvalidIndices, "Call ranks must be between 1 (A) and 13 (K).")
		}
	}
	// コールにはベットを伴う。レイアウトの合計を賭けとして用いる。
	stake := f.totalBet()
	if stake < f.config.MinBet {
		stake = f.config.MinBet
	}
	if !f.chips.SubtractChips(stake) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips for the call.")
	}
	f.callOrder = append([]int(nil), order...)
	actual := make([]int, len(f.callCards))
	for i, c := range f.callCards {
		actual[i] = rankOf1to13(c)
	}
	f.callWon = ranksEqual(order, actual)
	if f.callWon {
		payout := stake + stake*FaroCallPayoutMultiplier
		f.chips.AddChips(payout)
		f.totalPayout += stake * FaroCallPayoutMultiplier
		f.appendLog(0, "call", fmt.Sprintf("call won payout=%d", payout), f.callCards)
	} else {
		f.totalPayout -= stake
		f.appendLog(0, "call", "call lost", f.callCards)
	}
	f.phase = FaroPhaseRoundEnd
	return nil
}

// rankOf1to13 はカード値を 1..13 のランクに正規化する（A=1, J=11, Q=12, K=13）。
func rankOf1to13(c *Card) int {
	if c == nil {
		return 0
	}
	return c.GetValue()
}

// ranksEqual は2つのランク列が等しいかを判定する。
func ranksEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// appendLog は棋譜にエントリを追加する。
func (f *Faro) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	f.actionLog = append(f.actionLog, &ActionLogEntry{
		TurnNumber: len(f.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- Getters ---

// GetPhase は現在のフェーズを返す。
func (f *Faro) GetPhase() int { return f.phase }

// GetGameEndFlag はゲーム終了フラグを返す。
func (f *Faro) GetGameEndFlag() bool { return f.gameEndFlag }

// GetChips は現在のチップ数を返す。
func (f *Faro) GetChips() int { return f.chips.GetChips() }

// GetRemainingByRank は各ランクの残り枚数を返す (index 1..13 が A..K)。
//
// **未配の山札から直接数える。**公開済みカードを別に蓄えると、リセットの
// 取りこぼしや二重計上でケースキーパーが嘘をつく (#4894)。ソーダも配った
// 時点で山札から抜けているので、自動的に除かれる。
func (f *Faro) GetRemainingByRank() [FaroMaxRank + 1]int {
	var out [FaroMaxRank + 1]int
	if f.trumpCards == nil {
		return out
	}
	// 同じ package なので山札を直接見る。TrumpCards に新しい公開 API は要らない。
	for _, c := range f.trumpCards.deck {
		if c == nil || c.GetDraw() {
			continue
		}
		v := c.GetValue()
		if v >= FaroMinRank && v <= FaroMaxRank {
			out[v]++
		}
	}
	return out
}

// GetTurnsPlayed は消化済みターン数を返す。
func (f *Faro) GetTurnsPlayed() int { return f.turnsPlayed }

// GetTurnsTotal は1ディールあたりの総ターン数を返す。
func (f *Faro) GetTurnsTotal() int { return FaroTurnsPerDeal }

// GetRemainingCount はデッキ残枚数を返す。
func (f *Faro) GetRemainingCount() int { return f.trumpCards.GetRemainingCount() }

// GetSoda は焼かれたソーダ札を返す。
func (f *Faro) GetSoda() *Card { return f.soda }

// GetLastTurn は直近ターンの結果を返す（未プレイなら nil）。
func (f *Faro) GetLastTurn() *FaroTurnResult { return f.lastTurn }

// GetCallCards はコール対象の残り3枚を返す。
func (f *Faro) GetCallCards() []*Card { return f.callCards }

// GetCallOrder はプレイヤーが宣言したコール順を返す。
func (f *Faro) GetCallOrder() []int { return f.callOrder }

// GetCallWon はコールが成功したかを返す。
func (f *Faro) GetCallWon() bool { return f.callWon }

// GetTotalPayout は直近ラウンドのプレイヤー純損益を返す（正=利得）。
func (f *Faro) GetTotalPayout() int { return f.totalPayout }

// GetBets はレイアウト上の全ベットを rank->FaroBet のマップで返す。
func (f *Faro) GetBets() map[int]*FaroBet { return f.bets }

// GetBetRanks はベットされているランクを昇順で返す（決定的出力用）。
func (f *Faro) GetBetRanks() []int {
	ranks := make([]int, 0, len(f.bets))
	for r := range f.bets {
		ranks = append(ranks, r)
	}
	sort.Ints(ranks)
	return ranks
}

// GetConfig は設定を返す。
func (f *Faro) GetConfig() FaroConfig { return f.config }

// GetActionLog は棋譜を返す。
func (f *Faro) GetActionLog() []*ActionLogEntry { return f.actionLog }

// --- Test helpers ---

// SetPhase テスト用。
func (f *Faro) SetPhase(phase int) { f.phase = phase }

// SetChips テスト用。
func (f *Faro) SetChips(chips int) { f.chips.SetChips(chips) }

// SetBet テスト用。指定ランクにベットを直接セットする。
func (f *Faro) SetBet(rank, amount int, copper bool) {
	f.bets[rank] = &FaroBet{Amount: amount, Copper: copper}
}

// SetTurnsPlayed テスト用。
func (f *Faro) SetTurnsPlayed(n int) { f.turnsPlayed = n }

// faroJSON は Faro の JSON ワイヤーフォーマット。
type faroJSON struct {
	Config      FaroConfig        `json:"cf"`
	TrumpCards  *TrumpCards       `json:"tc"`
	Chips       *ChipHolder       `json:"ch"`
	Bets        map[int]*FaroBet  `json:"bt"`
	Soda        *Card             `json:"so"`
	TurnsPlayed int               `json:"tn"`
	LastTurn    *FaroTurnResult   `json:"lt"`
	CallOrder   []int             `json:"co"`
	CallCards   []*Card           `json:"cc"`
	CallWon     bool              `json:"cw"`
	Phase       int               `json:"ps"`
	GameEndFlag bool              `json:"ge"`
	TotalPayout int               `json:"tp"`
	ActionLog   []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (f *Faro) MarshalJSON() ([]byte, error) {
	return json.Marshal(faroJSON{
		Config:      f.config,
		TrumpCards:  f.trumpCards,
		Chips:       &f.chips,
		Bets:        f.bets,
		Soda:        f.soda,
		TurnsPlayed: f.turnsPlayed,
		LastTurn:    f.lastTurn,
		CallOrder:   f.callOrder,
		CallCards:   f.callCards,
		CallWon:     f.callWon,
		Phase:       f.phase,
		GameEndFlag: f.gameEndFlag,
		TotalPayout: f.totalPayout,
		ActionLog:   f.actionLog,
	})
}

// errFaroInput はデシリアライズ時の共通入力検証エラー。
var errFaroInput = fmt.Errorf("faro: invalid serialized state")

// UnmarshalJSON implements json.Unmarshaler.
func (f *Faro) UnmarshalJSON(data []byte) error {
	var j faroJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		j.Config = DefaultFaroConfig()
	}
	if j.Phase < FaroPhaseBetting || j.Phase > FaroPhaseGameEnd {
		return errFaroInput
	}
	if j.TurnsPlayed < 0 || j.TurnsPlayed > FaroTurnsPerDeal {
		return errFaroInput
	}
	if len(j.Bets) > faroMaxSliceLen || len(j.CallOrder) > faroMaxSliceLen ||
		len(j.CallCards) > faroMaxSliceLen || len(j.ActionLog) > faroMaxSliceLen {
		return errFaroInput
	}
	for rank, b := range j.Bets {
		if rank < FaroMinRank || rank > FaroMaxRank {
			return errFaroInput
		}
		if b == nil || b.Amount < 0 {
			return errFaroInput
		}
	}
	for _, r := range j.CallOrder {
		if r < FaroMinRank || r > FaroMaxRank {
			return errFaroInput
		}
	}
	f.config = j.Config
	f.trumpCards = j.TrumpCards
	if f.trumpCards == nil {
		f.trumpCards = NewTrumpCards(0)
	}
	if j.Chips != nil {
		f.chips = *j.Chips
	}
	f.bets = j.Bets
	if f.bets == nil {
		f.bets = make(map[int]*FaroBet)
	}
	f.soda = j.Soda
	f.turnsPlayed = j.TurnsPlayed
	f.lastTurn = j.LastTurn
	f.callOrder = j.CallOrder
	f.callCards = j.CallCards
	f.callWon = j.CallWon
	f.phase = j.Phase
	f.gameEndFlag = j.GameEndFlag
	f.totalPayout = j.TotalPayout
	f.actionLog = j.ActionLog
	if f.actionLog == nil {
		f.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
