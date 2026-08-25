//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"fmt"
)

// 席。**バンカー 1 + 左右 2 つの子。**
const (
	// BaccaratBanquePlayerCnt は席数。
	BaccaratBanquePlayerCnt = 3
	// BaccaratBanqueBankerIdx はバンカーの席 (人間)。
	BaccaratBanqueBankerIdx = 0
	// BaccaratBanqueRightIdx は右のタブロー。
	BaccaratBanqueRightIdx = 1
	// BaccaratBanqueLeftIdx は左のタブロー。
	BaccaratBanqueLeftIdx = 2
	// BaccaratBanqueReshuffleFloor はこれを下回るとシューを組み直す。
	BaccaratBanqueReshuffleFloor = 12
)

// フェーズ。
const (
	// BaccaratBanquePhasePunters は子が 3 枚目を決めている状態。
	BaccaratBanquePhasePunters = "punters"
	// BaccaratBanquePhaseBanker は人間 (バンカー) が 3 枚目を決める状態。
	BaccaratBanquePhaseBanker = "banker"
	// BaccaratBanquePhaseResult は 1 クーの結果を見せている状態。
	BaccaratBanquePhaseResult = "result"
	// BaccaratBanquePhaseGameEnd は終局。
	BaccaratBanquePhaseGameEnd = "gameEnd"
)

// BaccaratBanqueSideResult は 1 つのタブローとの決着。
type BaccaratBanqueSideResult struct {
	// SeatIdx はタブローの席。
	SeatIdx int
	// Outcome は勝敗の識別子。
	Outcome string
	// Bet は張られていた額。
	Bet int
	// Delta はバンカーから見た増減。
	Delta int
}

// BaccaratBanqueCoupResult は 1 クーの結果。
type BaccaratBanqueCoupResult struct {
	// BankerTotal はバンカーの合計。
	BankerTotal int
	// Sides は左右それぞれの決着。
	Sides []BaccaratBanqueSideResult
	// BankerDelta はバンカーの増減の合計。
	BankerDelta int
	// BankerNatural はバンカーがナチュラルだったか。
	BankerNatural bool
}

// BaccaratBanque はバカラ・バンクの状態を保持する集約ルート。
type BaccaratBanque struct {
	shoe    []*Card
	drawIdx int
	players []*BaccaratBanquePlayer
	config  BaccaratBanqueConfig
	phase   string
	// coupNumber は何回目のクーか。
	coupNumber int
	// pendingIdx は次に 3 枚目を決める子の席 (-1 = 子は済んだ)。
	pendingIdx int
	// bankHeld はこのバンクで何クー続けているか。
	//
	// **1 回負けてもバンクは動かない。** シューを配り切るか、自分から退くか、
	// 資金が尽きるまで同じ席が持つ ── #5462 の要件 6 はここを取り違えている。
	bankHeld int
	// retired はバンカーが自分から退いたか。
	retired     bool
	lastResult  *BaccaratBanqueCoupResult
	gameEndFlag bool
	// winnerIdx は終局時の勝者 (バンカーが残っていれば 0)。
	winnerIdx int
	actionLogBase
}

// NewBaccaratBanque はコンストラクタ。
func NewBaccaratBanque(players []*BaccaratBanquePlayer, config BaccaratBanqueConfig) *BaccaratBanque {
	return &BaccaratBanque{
		players:    players,
		config:     config,
		phase:      BaccaratBanquePhasePunters,
		pendingIdx: BaccaratBanqueRightIdx,
		winnerIdx:  -1,
	}
}

// NewDefaultBaccaratBanque は既定の設定で生成する。
func NewDefaultBaccaratBanque() *BaccaratBanque {
	cfg := DefaultBaccaratBanqueConfig()
	return NewBaccaratBanque(newBaccaratBanquePlayers(cfg.StartChips), cfg)
}

// newBaccaratBanquePlayers は席 0 をバンカー (人間) にして 3 席を作る。
func newBaccaratBanquePlayers(chips int) []*BaccaratBanquePlayer {
	return []*BaccaratBanquePlayer{
		NewBaccaratBanquePlayer(true, chips),
		NewBaccaratBanquePlayer(false, chips),
		NewBaccaratBanquePlayer(false, chips),
	}
}

// Reset はゲームを最初から始める。
func (b *BaccaratBanque) Reset() {
	b.players = newBaccaratBanquePlayers(b.config.StartChips)
	b.coupNumber = 0
	b.bankHeld = 0
	b.retired = false
	b.gameEndFlag = false
	b.winnerIdx = -1
	b.lastResult = nil
	b.actionLog = make([]*ActionLogEntry, 0)
	b.buildShoe()
	b.startCoup()
}

// buildShoe は 3 組を混ぜたシューを用意する。
func (b *BaccaratBanque) buildShoe() {
	b.shoe = NewBaccaratBanqueShoe()
	b.drawIdx = 0
	b.appendLog(-1, "shoe", fmt.Sprintf("new shoe of %d cards", len(b.shoe)), nil)
}

// NextCoup は次のクーを始める。
func (b *BaccaratBanque) NextCoup() {
	if b.gameEndFlag || b.phase != BaccaratBanquePhaseResult {
		return
	}
	b.startCoup()
}

// startCoup は左右の子とバンカーに 2 枚ずつ配る。
func (b *BaccaratBanque) startCoup() {
	// **シューを配り切ったらバンクは終わり。** そこが 1 つの区切り。
	if b.remaining() < BaccaratBanqueReshuffleFloor {
		b.endBank("shoeExhausted")
		if b.gameEndFlag {
			return
		}
		b.buildShoe()
	}
	for _, p := range b.players {
		p.ResetCoup()
	}
	b.coupNumber++
	b.bankHeld++
	// 子は決まった額を張る。
	for _, idx := range []int{BaccaratBanqueRightIdx, BaccaratBanqueLeftIdx} {
		bet := b.config.BetAmount
		if c := b.players[idx].GetChips(); c < bet {
			bet = c
		}
		b.players[idx].SetBet(bet)
	}
	// **右・左・バンカーの順に 1 枚ずつ、それを 2 周。**
	for round := 0; round < 2; round++ {
		for _, idx := range []int{BaccaratBanqueRightIdx, BaccaratBanqueLeftIdx, BaccaratBanqueBankerIdx} {
			if c := b.draw(); c != nil {
				b.players[idx].AddCard(c)
			}
		}
	}
	b.appendLog(-1, "deal", fmt.Sprintf("coup %d dealt", b.coupNumber), nil)
	b.phase = BaccaratBanquePhasePunters
	b.pendingIdx = BaccaratBanqueRightIdx
	b.resolvePunters()
}

// draw はシューから 1 枚引く。
func (b *BaccaratBanque) draw() *Card {
	if b.drawIdx >= len(b.shoe) {
		return nil
	}
	c := b.shoe[b.drawIdx]
	b.drawIdx++
	return c
}

// remaining はシューの残り枚数を返す。
func (b *BaccaratBanque) remaining() int { return len(b.shoe) - b.drawIdx }

// resolvePunters は左右の子の 3 枚目を決める。
//
// **裁量があるのは合計 5 のときだけ。** 0-4 は必ず引き、6-7 は必ず止まる ──
// どちらも CPU が持つので、ここで一気に片付けてバンカーの番へ渡す。
func (b *BaccaratBanque) resolvePunters() {
	for _, idx := range []int{BaccaratBanqueRightIdx, BaccaratBanqueLeftIdx} {
		p := b.players[idx]
		switch BaccaratBanquePunterRule(p.GetHand()) {
		case BaccaratBanqueDrawMust:
			b.dealThird(idx)
		case BaccaratBanqueDrawFree:
			if b.punterTakesOnFive(idx) {
				b.dealThird(idx)
			}
		}
	}
	b.pendingIdx = -1
	b.phase = BaccaratBanquePhaseBanker
	// **バンカーがナチュラルなら引く余地は無い。** そのまま決着させる。
	if BaccaratBanqueIsNatural(b.players[BaccaratBanqueBankerIdx].GetHand()) {
		b.settle()
	}
}

// dealThird は 3 枚目を配る。
func (b *BaccaratBanque) dealThird(idx int) {
	c := b.draw()
	if c == nil {
		return
	}
	b.players[idx].AddCard(c)
	b.players[idx].SetDrawn(true)
	b.appendLog(idx, "draw", fmt.Sprintf("seat %d draws a third card", idx), []*Card{c})
}

// BankerDraw は人間 (バンカー) が 3 枚目を引くかを決める。
func (b *BaccaratBanque) BankerDraw(draw bool) error {
	if b.gameEndFlag {
		return NewDomainErrorCode(ErrGameEnded, "baccaratbanque.errGameEnded", nil)
	}
	if b.phase != BaccaratBanquePhaseBanker {
		return NewDomainErrorCode(ErrWrongPhase, "baccaratbanque.errNotBankerPhase", nil)
	}
	banker := b.players[BaccaratBanqueBankerIdx]
	if BaccaratBanqueIsNatural(banker.GetHand()) {
		return NewDomainErrorCode(ErrInvalidPlay, "baccaratbanque.errNaturalCannotDraw", nil)
	}
	if draw {
		b.dealThird(BaccaratBanqueBankerIdx)
	}
	b.settle()
	return nil
}

// settle は左右それぞれと決着し、チップを動かす。
func (b *BaccaratBanque) settle() {
	banker := b.players[BaccaratBanqueBankerIdx]
	bankerTotal := banker.GetTotal()
	res := &BaccaratBanqueCoupResult{
		BankerTotal:   bankerTotal,
		BankerNatural: BaccaratBanqueIsNatural(banker.GetHand()),
	}

	// **左右は別勘定。** 片方に勝ってもう片方に負けることがある。
	for _, idx := range []int{BaccaratBanqueRightIdx, BaccaratBanqueLeftIdx} {
		p := b.players[idx]
		outcome := BaccaratBanqueCompare(bankerTotal, p.GetTotal())
		delta := 0
		switch outcome {
		case BaccaratBanqueOutcomeBankerWin:
			delta = p.GetBet()
		case BaccaratBanqueOutcomePunterWin:
			delta = -p.GetBet()
		}
		banker.AddChips(delta)
		p.AddChips(-delta)
		res.BankerDelta += delta
		res.Sides = append(res.Sides, BaccaratBanqueSideResult{
			SeatIdx: idx, Outcome: outcome, Bet: p.GetBet(), Delta: delta,
		})
	}

	b.lastResult = res
	b.appendLog(BaccaratBanqueBankerIdx, "settle",
		fmt.Sprintf("coup %d: banker %d, delta %+d", b.coupNumber, bankerTotal, res.BankerDelta), nil)
	b.phase = BaccaratBanquePhaseResult

	// **1 回負けてもバンクは動かない。** 資金が尽きたときだけ終わる。
	if banker.GetChips() <= 0 {
		b.endBank("bankrupt")
	}
}

// Retire はバンカーが自分からバンクを降りる。
//
// **降りるのは自分の意思か、資金切れか、シューを配り切ったときだけ。**
// 1 回負けたから、では降りない。
func (b *BaccaratBanque) Retire() error {
	if b.gameEndFlag {
		return NewDomainErrorCode(ErrGameEnded, "baccaratbanque.errGameEnded", nil)
	}
	if b.phase != BaccaratBanquePhaseResult {
		return NewDomainErrorCode(ErrWrongPhase, "baccaratbanque.errNotResultPhase", nil)
	}
	b.retired = true
	b.endBank("retired")
	return nil
}

// endBank はバンクを畳んで終局する。
func (b *BaccaratBanque) endBank(reason string) {
	if b.gameEndFlag {
		return
	}
	b.gameEndFlag = true
	b.phase = BaccaratBanquePhaseGameEnd
	// **勝ち負けは元手と比べて決める。** 増えていればバンカーの勝ち。
	if b.players[BaccaratBanqueBankerIdx].GetChips() > b.config.StartChips {
		b.winnerIdx = BaccaratBanqueBankerIdx
	}
	b.appendLog(-1, "bankEnd",
		fmt.Sprintf("bank ends after %d coup(s): %s", b.bankHeld, reason), nil)
}

// IsHumanTurn は人間 (バンカー) の判断待ちかを返す。
func (b *BaccaratBanque) IsHumanTurn() bool {
	return !b.gameEndFlag && b.phase == BaccaratBanquePhaseBanker
}

// GetConfig はゲーム設定を返す。
func (b *BaccaratBanque) GetConfig() BaccaratBanqueConfig { return b.config }

// SetConfig はゲーム設定を差し替える。
func (b *BaccaratBanque) SetConfig(cfg BaccaratBanqueConfig) { b.config = cfg }

// GetGameEndFlag は終局フラグを返す。
func (b *BaccaratBanque) GetGameEndFlag() bool { return b.gameEndFlag }

// GetPhase は現在のフェーズを返す。
func (b *BaccaratBanque) GetPhase() string { return b.phase }

// GetCoupNumber は何回目のクーかを返す。
func (b *BaccaratBanque) GetCoupNumber() int { return b.coupNumber }

// GetBankHeld はこのバンクで続けたクー数を返す。
func (b *BaccaratBanque) GetBankHeld() int { return b.bankHeld }

// GetShoeRemaining はシューの残り枚数を返す。
func (b *BaccaratBanque) GetShoeRemaining() int { return b.remaining() }

// IsRetired はバンカーが自分から降りたかを返す。
func (b *BaccaratBanque) IsRetired() bool { return b.retired }

// GetPlayerCnt は席数を返す。
func (b *BaccaratBanque) GetPlayerCnt() int { return len(b.players) }

// GetPlayer は指定席を返す。
func (b *BaccaratBanque) GetPlayer(i int) *BaccaratBanquePlayer {
	if i < 0 || i >= len(b.players) {
		return nil
	}
	return b.players[i]
}

// GetLastResult は直前のクーの結果を返す。
func (b *BaccaratBanque) GetLastResult() *BaccaratBanqueCoupResult { return b.lastResult }

// GetWinnerIdx は勝者の席を返す (-1 = バンカーの負け越し)。
func (b *BaccaratBanque) GetWinnerIdx() int { return b.winnerIdx }

// SetShoeForTest はシューを差し替える (テスト用)。
func (b *BaccaratBanque) SetShoeForTest(cards []*Card) {
	b.shoe = cards
	b.drawIdx = 0
}

// SetPhaseForTest はフェーズを差し替える (テスト用)。
func (b *BaccaratBanque) SetPhaseForTest(p string) { b.phase = p }

// SettleForTest は決着させる (テスト用)。
func (b *BaccaratBanque) SettleForTest() { b.settle() }

// baccaratBanqueJSON is the JSON wire format for BaccaratBanque.
type baccaratBanqueJSON struct {
	Shoe        []*Card                   `json:"sh"`
	DrawIdx     int                       `json:"di"`
	Players     []*BaccaratBanquePlayer   `json:"pl"`
	Config      BaccaratBanqueConfig      `json:"cf"`
	Phase       string                    `json:"ph"`
	CoupNumber  int                       `json:"cn"`
	PendingIdx  int                       `json:"pi"`
	BankHeld    int                       `json:"bh"`
	Retired     bool                      `json:"rt"`
	LastResult  *BaccaratBanqueCoupResult `json:"lr"`
	GameEndFlag bool                      `json:"ge"`
	WinnerIdx   int                       `json:"wi"`
	ActionLog   []*ActionLogEntry         `json:"al"`
}

// MarshalJSON implements json.Marshaler.
//
// **非公開フィールドだけの型は MarshalJSON が無いと `{}` になる。** シューの
// 位置が消えると、復元した盤では配り切りが来ずバンクが永遠に続く。
func (b *BaccaratBanque) MarshalJSON() ([]byte, error) {
	return json.Marshal(baccaratBanqueJSON{
		Shoe: b.shoe, DrawIdx: b.drawIdx, Players: b.players, Config: b.config,
		Phase: b.phase, CoupNumber: b.coupNumber, PendingIdx: b.pendingIdx,
		BankHeld: b.bankHeld, Retired: b.retired, LastResult: b.lastResult,
		GameEndFlag: b.gameEndFlag, WinnerIdx: b.winnerIdx, ActionLog: b.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *BaccaratBanque) UnmarshalJSON(data []byte) error {
	var j baccaratBanqueJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	b.players = j.Players
	if len(b.players) != BaccaratBanquePlayerCnt {
		return fmt.Errorf("baccaratbanque: expected %d seats, got %d",
			BaccaratBanquePlayerCnt, len(b.players))
	}
	b.shoe, b.drawIdx, b.config = j.Shoe, j.DrawIdx, j.Config
	b.phase, b.coupNumber, b.pendingIdx = j.Phase, j.CoupNumber, j.PendingIdx
	b.bankHeld, b.retired = j.BankHeld, j.Retired
	b.lastResult, b.gameEndFlag, b.winnerIdx = j.LastResult, j.GameEndFlag, j.WinnerIdx
	b.actionLog = j.ActionLog
	if b.actionLog == nil {
		b.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
