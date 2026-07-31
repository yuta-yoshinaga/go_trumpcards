//go:build !js || !wasm || extra3

// Package domain — キッレ (Kille / Cambio) のドメインモデル。
//
// スウェーデンのカックー系。**専用 42 枚**(21 種 × 2 枚、単一スート)。
//
// # issue #4383 の仕様案との相違
//
// 42 枚専用デッキという点は合っているが、**絵札の効果が 2 つとも逆**で、
// 脱落の仕組みも違う。
//
//   - issue は「**Harlekin（道化）は交換拒否可能**」とするが、拒否しない。
//     Harlequin は**表向きに交換され**、配られた／山から引いたものは**最強**、
//     **交換されたものは最弱**になる。向きで強さが反転する札である
//   - issue は「**Kuckuck（カッコウ）は交換強制**」とするが、**逆**。
//     «No one swaps the Cuckoo!» — 交換されず、**その場でラウンドが終わって
//     全員が公開する**
//   - issue が触れていない絵札が 3 種ある。**Hussar** は仕掛けた側を脱落させ、
//     **Pig** は交換を巻き戻して**元の持ち主**を脱落させ、**Cavalier / Inn** は
//     交換者を次の人へ回す
//   - issue は「ライフ（通常3）」とするが、**ライフ制ではない**。掛け金で参加し、
//     負ければ脱落。**買い戻しは 3 回まで**で、1 口 → ポット半分 → ポット全額
//   - issue は「最下位ランクの手札を持つ者が…」とするが、**Harlequin を除いた**
//     最下位が負け。加えて Hussar / Pig で落とされた人も一緒に脱落する
//   - 3〜6 人ではなく **3〜12 人**
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// KillePlayerCnt はプレイヤー数 (4 人。原典は 3〜12 人)。
const KillePlayerCnt = 4

// KillePhase はゲームフェーズ。
type KillePhase int

// Kille のフェーズ定数
const (
	// KillePhaseExchange 交換ラウンド
	KillePhaseExchange KillePhase = iota
	// KillePhaseShowdown 公開・精算済み
	KillePhaseShowdown
	// KillePhaseGameEnd 決着
	KillePhaseGameEnd
)

// Kille の絵札効果の識別子。脱落理由として記録する。
const (
	// KilleKnockHussar Hussar に返り討ちにされた
	KilleKnockHussar = "hussar"
	// KilleKnockPig Pig に噛まれた
	KilleKnockPig = "pig"
	// KilleKnockLowest 最弱で負けた
	KilleKnockLowest = ""
)

// KilleEvent は交換で起きた出来事の記録。
type KilleEvent struct {
	// Kind は "swap" / "satisfied" / "cuckoo" / "hussar" / "pig" / "pass" /
	// "stock" のいずれか。
	Kind string
	// Actor は仕掛けた席。
	Actor int
	// Target は仕掛けられた席 (-1: 山札との交換)。
	Target int
}

// Kille はキッレのゲームクラス。
type Kille struct {
	players []*KillePlayer
	config  KilleConfig
	phase   KillePhase

	stock []*Card
	pot   int

	currentIdx int
	dealerIdx  int
	roundNo    int

	events []*KilleEvent
	// loserIdxs は直近ラウンドで脱落した席。
	loserIdxs []int

	gameEndFlag bool
	winnerIdx   int
	actionLog   []*ActionLogEntry
}

// NewKille はコンストラクタ。
func NewKille(players []*KillePlayer, config KilleConfig) *Kille {
	return &Kille{players: players, config: config, winnerIdx: -1}
}

// NewDefaultKille は標準の 4 人セットアップを返す。
func NewDefaultKille() *Kille {
	players := make([]*KillePlayer, 0, KillePlayerCnt)
	players = append(players, NewKillePlayer(true))
	for range KillePlayerCnt - 1 {
		players = append(players, NewKillePlayer(false))
	}
	return NewKille(players, DefaultKilleConfig())
}

// Reset はゲーム全体を初期化する。
func (k *Kille) Reset() {
	k.pot = 0
	k.dealerIdx = 0
	k.roundNo = 0
	k.gameEndFlag = false
	k.winnerIdx = -1
	k.actionLog = nil
	for _, p := range k.players {
		p.AddChips(-p.GetChips())
		p.reentries = 0
	}
	k.dealRound()
}

// dealRound は 1 ラウンドを配る。**各自 1 枚だけ。**
func (k *Kille) dealRound() {
	k.events = nil
	k.loserIdxs = nil
	for _, p := range k.players {
		p.ResetRound()
	}

	// 全員が掛け金を出す。
	for _, p := range k.players {
		p.AddChips(-k.config.Stake)
		k.pot += k.config.Stake
	}

	deck := newKilleDeck()
	killeShuffle(deck)
	for i, p := range k.players {
		p.AddCard(deck[i])
	}
	k.stock = append([]*Card(nil), deck[len(k.players):]...)

	// **ディーラーは最後。**先手はその左隣。
	k.currentIdx = (k.dealerIdx + 1) % len(k.players)
	k.phase = KillePhaseExchange
	k.addLog(-1, "deal", fmt.Sprintf("one card each, pot is %d", k.pot), nil)
}

// killeShuffle は Fisher-Yates。
func killeShuffle(cards []*Card) {
	for i := len(cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}
}

// Satisfied は「交換しない」と宣言して手番を終える。
func (k *Kille) Satisfied(player int) error {
	if err := k.checkTurn(player); err != nil {
		return err
	}
	k.GetPlayer(player).SetSatisfied(true)
	k.events = append(k.events, &KilleEvent{Kind: "satisfied", Actor: player, Target: -1})
	k.addLog(player, "satisfied", "is satisfied", nil)
	k.advance()
	return nil
}

// Exchange は左隣と手札を交換しようとする。
//
// **相手の札次第で結果が変わる。**Cuckoo なら即公開、Hussar なら仕掛けた側が
// 脱落、Pig なら巻き戻して元の持ち主が脱落、Cavalier / Inn なら次の人へ回す。
func (k *Kille) Exchange(player int) error {
	if err := k.checkTurn(player); err != nil {
		return err
	}
	// **ディーラーは山札と交換する。**隣に回せる相手がいない位置なので。
	if player == k.dealerIdx {
		return k.exchangeWithStock(player)
	}
	return k.exchangeWithNeighbour(player, k.nextSeat(player), 0)
}

// exchangeWithNeighbour は target と交換しようとする。hops は Cavalier / Inn で
// 回された回数で、一周したら打ち切る。
func (k *Kille) exchangeWithNeighbour(player, target, hops int) error {
	if hops >= len(k.players) {
		// 全員が Cavalier / Inn だった。交換できずに手番を終える。
		k.addLog(player, "pass", "nobody could be swapped with", nil)
		k.advance()
		return nil
	}
	// **自分自身は交換相手にならない。**Cavalier / Inn で一周して戻ってくると
	// ここに来る。素通しすると自分と自分を入れ替えて手札が消える。
	if target == player {
		k.addLog(player, "pass", "the pass came full circle", nil)
		k.advance()
		return nil
	}
	tp := k.GetPlayer(target)
	if tp == nil || tp.IsOut() {
		return k.exchangeWithNeighbour(player, k.nextSeat(target), hops+1)
	}

	switch KilleRankOf(tp.GetCard(0)) {
	case KilleCuckoo:
		// **«No one swaps the Cuckoo!»** その場でラウンドが終わる。
		k.events = append(k.events, &KilleEvent{Kind: "cuckoo", Actor: player, Target: target})
		k.addLog(target, "cuckoo", "nobody swaps the Cuckoo; the round ends", nil)
		k.finishRound()
		return nil
	case KilleHussar:
		// **«Hussar strikes!»** 仕掛けた側が落ちる。
		k.GetPlayer(player).SetOut(KilleKnockHussar)
		k.events = append(k.events, &KilleEvent{Kind: "hussar", Actor: player, Target: target})
		k.addLog(target, "hussar", "the Hussar strikes down the challenger", nil)
		k.advance()
		return nil
	case KillePig:
		// **«Pig bites back!»** 交換は取り消され、その札に関わる過去の交換も
		// すべて巻き戻り、**元の持ち主**が落ちる。
		k.revertSwapsInvolving(target)
		tp.SetOut(KilleKnockPig)
		k.events = append(k.events, &KilleEvent{Kind: "pig", Actor: player, Target: target})
		k.addLog(target, "pig", "the Pig bites back; its holder is out", nil)
		k.advance()
		return nil
	case KilleCavalier, KilleInn:
		// **«Pass the Cavalier/Inn!»** 次の人に回す。
		k.events = append(k.events, &KilleEvent{Kind: "pass", Actor: player, Target: target})
		k.addLog(target, "pass", "pass along to the next player", nil)
		return k.exchangeWithNeighbour(player, k.nextSeat(target), hops+1)
	}

	k.swap(player, target)
	k.events = append(k.events, &KilleEvent{Kind: "swap", Actor: player, Target: target})
	k.addLog(player, "swap", "swaps with the next player", nil)
	k.advance()
	return nil
}

// exchangeWithStock はディーラーが山札と交換する。
func (k *Kille) exchangeWithStock(player int) error {
	if len(k.stock) == 0 {
		return fmt.Errorf("the stock is empty")
	}
	p := k.GetPlayer(player)
	drawn := k.stock[0]
	k.stock = k.stock[1:]
	old := p.RemoveCard(0)
	p.AddCard(drawn)
	// **山から引いた Harlequin は最強のまま。**交換で渡ってきたものだけが弱い。
	p.SetHarlequinSwapped(false)
	k.stock = append(k.stock, old)

	k.events = append(k.events, &KilleEvent{Kind: "stock", Actor: player, Target: -1})
	k.addLog(player, "stock", "exchanges with the stock", []*Card{drawn})
	k.advance()
	return nil
}

// swap は 2 席の手札を入れ替える。
func (k *Kille) swap(a, b int) {
	pa, pb := k.GetPlayer(a), k.GetPlayer(b)
	ca, cb := pa.RemoveCard(0), pb.RemoveCard(0)
	pa.AddCard(cb)
	pb.AddCard(ca)
	// **交換で渡った Harlequin は最弱になる。**受け取った側に印を付ける。
	pa.SetHarlequinSwapped(KilleRankOf(cb) == KilleHarlequin)
	pb.SetHarlequinSwapped(KilleRankOf(ca) == KilleHarlequin)
}

// revertSwapsInvolving は seat の札に関わった交換を、記録を逆順にたどって
// 巻き戻す。**Pig の «bites back» はここまでやる。**
func (k *Kille) revertSwapsInvolving(seat int) {
	for i := len(k.events) - 1; i >= 0; i-- {
		e := k.events[i]
		if e.Kind != "swap" {
			continue
		}
		if e.Actor != seat && e.Target != seat {
			continue
		}
		k.swap(e.Actor, e.Target)
		// 巻き戻した交換は無かったことにする。
		k.events = append(k.events[:i], k.events[i+1:]...)
		// 連鎖して動いた札を追うため、相手側も見る。
		if e.Actor == seat {
			seat = e.Target
		} else {
			seat = e.Actor
		}
	}
}

// checkTurn は動ける状態かを確かめる。
func (k *Kille) checkTurn(player int) error {
	if k.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if k.phase != KillePhaseExchange {
		return fmt.Errorf("the exchange round is not in progress")
	}
	if player != k.currentIdx {
		return fmt.Errorf("it is not player %d's turn", player)
	}
	if k.GetPlayer(player).IsOut() {
		return fmt.Errorf("player %d is out of this round", player)
	}
	return nil
}

// nextSeat は idx の次の席を返す。
func (k *Kille) nextSeat(idx int) int { return (idx + 1) % len(k.players) }

// advance は次の生存者へ手番を回す。ディーラーまで回り切ったら公開する。
func (k *Kille) advance() {
	if k.currentIdx == k.dealerIdx {
		k.finishRound()
		return
	}
	for i := 1; i <= len(k.players); i++ {
		n := (k.currentIdx + i) % len(k.players)
		if !k.players[n].IsOut() {
			k.currentIdx = n
			return
		}
		if n == k.dealerIdx {
			break
		}
	}
	k.finishRound()
}

// KilleStrength は席の実効的な強さを返す。
//
// **Harlequin は向きで反転する。**交換で渡ってきたものは最弱として扱うので、
// 素の種の値ではなく 0 を返す。
func (k *Kille) KilleStrength(seat int) int {
	p := k.GetPlayer(seat)
	if p == nil || p.GetCardsSize() == 0 {
		return 0
	}
	r := KilleRankOf(p.GetCard(0))
	if r == KilleHarlequin && p.IsHarlequinSwapped() {
		return 0
	}
	return int(r)
}

// finishRound は公開して精算する。
func (k *Kille) finishRound() {
	// **Harlequin を除いた**最弱が負け。交換で渡った Harlequin は
	// KilleStrength が 0 を返すので自然に最弱側へ落ちる。
	lowest, lowestSeat := 1<<30, -1
	for i, p := range k.players {
		if p.IsOut() {
			continue
		}
		if s := k.KilleStrength(i); s < lowest {
			lowest, lowestSeat = s, i
		}
	}
	if lowestSeat >= 0 {
		k.players[lowestSeat].SetOut(KilleKnockLowest)
	}

	k.loserIdxs = nil
	for i, p := range k.players {
		if p.IsOut() {
			k.loserIdxs = append(k.loserIdxs, i)
		}
	}
	k.addLog(-1, "showdown", fmt.Sprintf("%d player(s) go out", len(k.loserIdxs)), nil)

	k.roundNo++
	k.phase = KillePhaseShowdown
	k.checkGameEnd()
}

// KilleReentryCost は次に買い戻すのに要る額を返す。
//
// **1 回目は 1 口、2 回目はポットの半分、3 回目はポット全額。**4 回目は無い。
func (k *Kille) KilleReentryCost(seat int) int {
	p := k.GetPlayer(seat)
	if p == nil || !p.CanReenter() {
		return 0
	}
	switch p.GetReentries() {
	case 0:
		return k.config.Stake
	case 1:
		return k.pot / 2
	default:
		return k.pot
	}
}

// Reenter は脱落した席が買い戻す。
func (k *Kille) Reenter(seat int) error {
	if k.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if k.phase != KillePhaseShowdown {
		return fmt.Errorf("the round is still in progress")
	}
	p := k.GetPlayer(seat)
	if p == nil || !p.IsOut() {
		return fmt.Errorf("player %d is not out", seat)
	}
	if !p.CanReenter() {
		return fmt.Errorf("player %d has already bought back %d times", seat, KilleMaxReentries)
	}
	cost := k.KilleReentryCost(seat)
	p.AddChips(-cost)
	k.pot += cost
	p.AddReentry()
	p.out = false
	p.knockedBy = ""
	k.addLog(seat, "reenter", fmt.Sprintf("buys back in for %d", cost), nil)
	return nil
}

// checkGameEnd は残り 1 人になっていれば決着させる。
func (k *Kille) checkGameEnd() {
	alive, last := 0, -1
	for i, p := range k.players {
		// 買い戻せる人はまだ終わっていない。
		if !p.IsOut() || p.CanReenter() {
			alive++
			last = i
		}
	}
	if alive > 1 {
		return
	}
	k.winnerIdx = last
	if last >= 0 {
		k.players[last].AddChips(k.pot)
		k.pot = 0
	}
	k.gameEndFlag = true
	k.phase = KillePhaseGameEnd
	k.addLog(last, "game_end", "is the last one standing", nil)
}

// NextRound は次のラウンドを配る。
func (k *Kille) NextRound() error {
	if k.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if k.phase != KillePhaseShowdown {
		return fmt.Errorf("the round is still in progress")
	}
	// 買い戻さなかった脱落者はそのまま退場。
	for _, p := range k.players {
		if p.IsOut() && !p.CanReenter() {
			p.SetIsFinished(true)
		}
	}
	k.dealerIdx = (k.dealerIdx + 1) % len(k.players)
	k.dealRound()
	return nil
}

// ---- CPU ----

// KilleCpuAction は CPU が選んだ手。
type KilleCpuAction struct {
	// Type は "exchange" / "satisfied"。
	Type string
}

// KilleCpuDecide は idx の CPU が取る手を決める。
//
// 弱い札なら交換を仕掛け、強い札なら満足する。閾値は数札の真ん中あたり。
func (k *Kille) KilleCpuDecide(idx int) KilleCpuAction {
	p := k.GetPlayer(idx)
	if p == nil || p.GetCardsSize() == 0 {
		return KilleCpuAction{Type: "satisfied"}
	}
	if k.KilleStrength(idx) <= int(KilleNum6) {
		return KilleCpuAction{Type: "exchange"}
	}
	return KilleCpuAction{Type: "satisfied"}
}

// ---- 公開アクセサ ----

// GetPlayers は全プレイヤーを返す。
func (k *Kille) GetPlayers() []*KillePlayer { return k.players }

// GetPlayer は idx のプレイヤーを返す。
func (k *Kille) GetPlayer(idx int) *KillePlayer {
	if idx < 0 || idx >= len(k.players) {
		return nil
	}
	return k.players[idx]
}

// GetPhase は現在のフェーズを返す。
func (k *Kille) GetPhase() KillePhase { return k.phase }

// GetCurrentPlayerIdx は手番のプレイヤー添字を返す。
func (k *Kille) GetCurrentPlayerIdx() int { return k.currentIdx }

// GetDealerIdx はディーラーの席を返す。
func (k *Kille) GetDealerIdx() int { return k.dealerIdx }

// GetPot は場の掛け金を返す。
func (k *Kille) GetPot() int { return k.pot }

// GetStockCount は山札の残り枚数を返す。
func (k *Kille) GetStockCount() int { return len(k.stock) }

// GetEvents は交換で起きた出来事を返す。
func (k *Kille) GetEvents() []*KilleEvent { return k.events }

// GetLoserIdxs は直近ラウンドで脱落した席を返す。
func (k *Kille) GetLoserIdxs() []int { return k.loserIdxs }

// GetRoundNumber は完了したラウンド数を返す。
func (k *Kille) GetRoundNumber() int { return k.roundNo }

// GetGameEndFlag は決着しているかを返す。
func (k *Kille) GetGameEndFlag() bool { return k.gameEndFlag }

// GetWinnerIdx は勝者の添字を返す (-1: 未決着)。
func (k *Kille) GetWinnerIdx() int { return k.winnerIdx }

// GetConfig はゲーム設定を返す。
func (k *Kille) GetConfig() KilleConfig { return k.config }

// SetConfig はゲーム設定をセットする。
func (k *Kille) SetConfig(c KilleConfig) { k.config = c }

// GetActionLog は棋譜を返す。
func (k *Kille) GetActionLog() []*ActionLogEntry { return k.actionLog }

// SetPhaseForTest はテスト用にフェーズを差し替える。
func (k *Kille) SetPhaseForTest(p KillePhase) { k.phase = p }

// SetCurrentPlayerForTest はテスト用に手番を差し替える。
func (k *Kille) SetCurrentPlayerForTest(idx int) { k.currentIdx = idx }

// SetDealerForTest はテスト用にディーラーを差し替える。
func (k *Kille) SetDealerForTest(idx int) { k.dealerIdx = idx }

// SetPotForTest はテスト用にポットを差し替える。
func (k *Kille) SetPotForTest(n int) { k.pot = n }

// SetHandForTest はテスト用に席の手札を 1 枚に差し替える。
func (k *Kille) SetHandForTest(seat int, r KilleRank) {
	p := k.GetPlayer(seat)
	p.Reset()
	p.AddCard(NewKilleCard(r))
	p.SetHarlequinSwapped(false)
}

// SetStockForTest はテスト用に山札を差し替える。
func (k *Kille) SetStockForTest(cards []*Card) { k.stock = cards }

// addLog は棋譜に 1 件追加する。
func (k *Kille) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	k.actionLog = append(k.actionLog, &ActionLogEntry{
		TurnNumber: len(k.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// killeJSON is the JSON wire format for Kille.
type killeJSON struct {
	Players   []*KillePlayer    `json:"pl"`
	Config    KilleConfig       `json:"cfg"`
	Phase     KillePhase        `json:"ph"`
	Stock     []*Card           `json:"st"`
	Pot       int               `json:"pt"`
	Current   int               `json:"cur"`
	Dealer    int               `json:"dl"`
	RoundNo   int               `json:"rn"`
	Events    []*KilleEvent     `json:"ev"`
	Losers    []int             `json:"ls"`
	GameEnd   bool              `json:"ge"`
	WinnerIdx int               `json:"wi"`
	ActionLog []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (k *Kille) MarshalJSON() ([]byte, error) {
	return json.Marshal(killeJSON{
		Players: k.players, Config: k.config, Phase: k.phase, Stock: k.stock,
		Pot: k.pot, Current: k.currentIdx, Dealer: k.dealerIdx, RoundNo: k.roundNo,
		Events: k.events, Losers: k.loserIdxs, GameEnd: k.gameEndFlag,
		WinnerIdx: k.winnerIdx, ActionLog: k.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// KV から戻る生バイト列は信用できないので、席数に合わせて詰め直し、設定を
// 検証する。**ポットと買い戻し回数は進行そのもの**なので、そのまま復元する。
func (k *Kille) UnmarshalJSON(data []byte) error {
	var raw killeJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Players) != KillePlayerCnt {
		return fmt.Errorf("expected %d players, got %d", KillePlayerCnt, len(raw.Players))
	}
	if err := raw.Config.Validate(); err != nil {
		return err
	}
	if raw.Phase < KillePhaseExchange || raw.Phase > KillePhaseGameEnd {
		return fmt.Errorf("unknown phase: %d", raw.Phase)
	}

	k.players = raw.Players
	k.config = raw.Config
	k.phase = raw.Phase
	k.stock = raw.Stock
	k.pot = raw.Pot
	if k.pot < 0 {
		k.pot = 0
	}
	k.roundNo = raw.RoundNo
	k.gameEndFlag = raw.GameEnd
	k.actionLog = raw.ActionLog

	k.currentIdx = clampKilleIdx(raw.Current, len(k.players))
	k.dealerIdx = clampKilleIdx(raw.Dealer, len(k.players))
	k.winnerIdx = raw.WinnerIdx
	if k.winnerIdx < -1 || k.winnerIdx >= len(k.players) {
		k.winnerIdx = -1
	}

	k.events = make([]*KilleEvent, 0, len(raw.Events))
	for _, e := range raw.Events {
		if e == nil || e.Actor < 0 || e.Actor >= len(k.players) {
			continue
		}
		// Target は -1 (山札) を許す。
		if e.Target < -1 || e.Target >= len(k.players) {
			continue
		}
		k.events = append(k.events, e)
	}

	k.loserIdxs = make([]int, 0, len(raw.Losers))
	for _, i := range raw.Losers {
		if i >= 0 && i < len(k.players) {
			k.loserIdxs = append(k.loserIdxs, i)
		}
	}
	return nil
}

// clampKilleIdx は席番号を 0..n-1 に収める。
func clampKilleIdx(idx, n int) int {
	if idx < 0 || idx >= n {
		return 0
	}
	return idx
}
