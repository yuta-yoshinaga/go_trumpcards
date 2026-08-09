//go:build !js || !wasm || extra3

// Package domain ブラ (Bura / Бура) のドメインモデル。
//
// ロシアの ace-ten 系トリックテイキングゲーム。36 枚のショートデッキを使い、
// 手札は常に 3 枚に補充される。カード点は A=11, 10=10, K=4, Q=3, J=2 で
// 合計 120 点、31 点を集めた側が勝つ。
//
// 実装にあたり、issue #4402 の記載と実際のルール (pagat.com) が食い違って
// いた点は実ルール側を採用している。特に以下の 3 つはゲーム性そのものに
// 関わる:
//
//   - **フォロー義務はない。** 応じる側は任意のカードを出してよく、勝てない
//     カードを捨てて強い札を温存するのがこのゲームの駆け引きの中心である。
//     「同スート上位か切札で応じる」のは *勝つための条件* であって *出せる
//     条件* ではない。
//   - **31 点は自己申告制。** 到達しただけでは勝たない。宣言して初めて勝ち、
//     足りていなければ負ける。自動判定にすると宣言漏れと誤申告という
//     このゲーム最大の緊張が消える。
//   - **山札が尽きて誰も宣言しなければ引き分け。** 点数で決着させると
//     31 点先取という目標が飾りになる。
//
// リードは 1〜3 枚の同一スートをまとめて出せる。これを受けるには同じ枚数を
// 出し、リードの各カードをそれぞれ別のカードで上回る必要がある。この判定は
// 枚数が高々 3 なので全順列を試す完全探索で行う (buraBeatsCombination)。
// 貪欲法でも解けるが、「切札は万能・リードスートは閾値」という二種類の
// 被覆が混ざるため正しさの議論が必要になる。3! = 6 通りを試す方が安い。
package domain

import (
	"encoding/json"
	"fmt"
)

// BuraPlayerCnt ブラのプレイヤー数 (v1は2人固定)
const BuraPlayerCnt = 2

// BuraHandSize 各プレイヤーの手札枚数 (山札がある間は毎トリック後に補充される)
const BuraHandSize = 3

// BuraWinThreshold 勝利に必要なカード点。これ以上を集めて宣言すると勝ち。
const BuraWinThreshold = 31

// BuraTotalPoints デッキ全体の合計点 (4 スート × 30 点)
const BuraTotalPoints = 120

// BuraPhase ゲームフェーズ
type BuraPhase int

// Buraのフェーズ定数
const (
	// BuraPhasePlay トリックプレイフェーズ
	BuraPhasePlay BuraPhase = iota
	// BuraPhaseGameEnd ゲーム終了フェーズ
	BuraPhaseGameEnd
)

// BuraCombination 手札が成立させている即勝ち役。
type BuraCombination int

// Buraの即勝ち役定数
const (
	// BuraCombinationNone 役なし
	BuraCombinationNone BuraCombination = iota
	// BuraCombinationBura 切札3枚。役の名前がそのままゲーム名になっている。
	BuraCombinationBura
	// BuraCombinationMoscow エース3枚 (モスクワ)
	BuraCombinationMoscow
	// BuraCombinationLittleMoscow 6が3枚、ただし切札の6を含むこと (小モスクワ)
	BuraCombinationLittleMoscow
	// BuraCombinationMolodka 切札以外の同一スート3枚 (モロトカ)
	BuraCombinationMolodka
)

// buraCardPoints カード点。A=11, 10=10, K=4, Q=3, J=2, それ以外 0。
var buraCardPoints = map[int]int{
	1:  11, // A
	10: 10,
	13: 4, // K
	12: 3, // Q
	11: 2, // J
}

// buraRankOrder スート内の強さ。値が大きいほど強い。
// A>10>K>Q>J>9>8>7>6 を 1-base で表す。10 が K/Q/J より強いのが
// ace-ten 系の特徴で、カード値 (10 < J=11) の順に並べると逆になる。
var buraRankOrder = map[int]int{
	6:  1,
	7:  2,
	8:  3,
	9:  4,
	11: 5, // J
	12: 6, // Q
	13: 7, // K
	10: 8,
	1:  9, // A
}

// BuraCardPoints カードの点数を返す。nil は 0 点。
func BuraCardPoints(c *Card) int {
	if c == nil {
		return 0
	}
	return buraCardPoints[c.GetValue()]
}

// BuraRankOrder カードのスート内順位を返す (大きいほど強い)。nil は 0。
func BuraRankOrder(c *Card) int {
	if c == nil {
		return 0
	}
	return buraRankOrder[c.GetValue()]
}

// buraCardBeats a が b を上回るかを返す。
// 切札は非切札を常に上回り、同スートなら順位で比較する。
// 異なる非切札スート同士は上回れない (捨て札にしかならない)。
func buraCardBeats(a, b *Card, trumpSuit int) bool {
	if a == nil || b == nil {
		return false
	}
	aTrump := a.GetDesign() == trumpSuit
	bTrump := b.GetDesign() == trumpSuit
	switch {
	case aTrump && !bTrump:
		return true
	case !aTrump && bTrump:
		return false
	case a.GetDesign() != b.GetDesign():
		return false
	default:
		return BuraRankOrder(a) > BuraRankOrder(b)
	}
}

// buraBeatsCombination resp が lead を完全に受けきるかを返す。
//
// 枚数が一致し、かつ resp の各カードを lead の各カードへ 1 対 1 に割り当てて
// すべて上回れる組み合わせが存在するときに true。割り当ては全順列を試す。
// 手札が 3 枚なので最大でも 6 通りしかなく、貪欲法の正しさを論じるより安い。
func buraBeatsCombination(resp, lead []*Card, trumpSuit int) bool {
	if len(resp) != len(lead) || len(lead) == 0 {
		return false
	}
	used := make([]bool, len(resp))
	var search func(i int) bool
	search = func(i int) bool {
		if i == len(lead) {
			return true
		}
		for j := range resp {
			if used[j] || !buraCardBeats(resp[j], lead[i], trumpSuit) {
				continue
			}
			used[j] = true
			if search(i + 1) {
				return true
			}
			used[j] = false
		}
		return false
	}
	return search(0)
}

// buraValidateLead リードとして出せる組み合わせかを検証する。
// 1〜BuraHandSize 枚の、すべて同一スートのカードのみ許される。
func buraValidateLead(cards []*Card) error {
	if len(cards) == 0 {
		return fmt.Errorf("must lead at least one card")
	}
	if len(cards) > BuraHandSize {
		return fmt.Errorf("cannot lead more than %d cards", BuraHandSize)
	}
	suit := -1
	for _, c := range cards {
		if c == nil {
			return fmt.Errorf("cannot lead a missing card")
		}
		if suit == -1 {
			suit = c.GetDesign()
			continue
		}
		if c.GetDesign() != suit {
			return fmt.Errorf("a lead must be a single suit")
		}
	}
	return nil
}

// BuraDetectCombination 手札が成立させている即勝ち役を返す。
// 3 枚ちょうどのときのみ成立しうる。判定順は bura > Moscow > little Moscow >
// molodka で、切札 3 枚は molodka ではなく bura として扱う。
func BuraDetectCombination(hand []*Card, trumpSuit int) BuraCombination {
	if len(hand) != BuraHandSize {
		return BuraCombinationNone
	}
	for _, c := range hand {
		if c == nil {
			return BuraCombinationNone
		}
	}

	trumps, aces, sixes, trumpSix := 0, 0, 0, false
	sameSuit := true
	for _, c := range hand {
		if c.GetDesign() == trumpSuit {
			trumps++
			if c.GetValue() == 6 {
				trumpSix = true
			}
		}
		if c.GetValue() == 1 {
			aces++
		}
		if c.GetValue() == 6 {
			sixes++
		}
		if c.GetDesign() != hand[0].GetDesign() {
			sameSuit = false
		}
	}

	switch {
	case trumps == BuraHandSize:
		return BuraCombinationBura
	case aces == BuraHandSize:
		return BuraCombinationMoscow
	case sixes == BuraHandSize && trumpSix:
		return BuraCombinationLittleMoscow
	case sameSuit:
		return BuraCombinationMolodka
	default:
		return BuraCombinationNone
	}
}

// Bura ブラのゲームクラス
type Bura struct {
	trumpCards       *TrumpCards
	players          []*BuraPlayer
	config           BuraConfig
	phase            BuraPhase
	stock            []*Card
	trumpCard        *Card // 山札の底に表向きで置かれる切札指示カード
	trumpSuit        int
	trickNumber      int
	currentPlayerIdx int
	leadPlayerIdx    int
	currentLead      []*Card // リードされたカード (トリック解決時に空へ戻る)
	currentResponse  []*Card // 応じたカード
	trickCards       []*Card // 現トリックの場札すべて
	playerPoints     []int
	gameEndFlag      bool
	winnerIdx        int // -1: 未確定または引き分け
	drawFlag         bool
	actionLogBase
}

// NewBura コンストラクタ
func NewBura(trumpCards *TrumpCards, players []*BuraPlayer, config BuraConfig) *Bura {
	return &Bura{
		trumpCards:   trumpCards,
		players:      players,
		config:       config,
		winnerIdx:    -1,
		playerPoints: make([]int, len(players)),
	}
}

// NewDefaultBura 標準の 2 人対戦セットアップを返す。
// 人間プレイヤー (idx 0) と CPU (idx 1) の組み合わせ。
func NewDefaultBura() *Bura {
	players := []*BuraPlayer{
		NewBuraPlayer(true),
		NewBuraPlayer(false),
	}
	return NewBura(NewTrumpCardsShortDeck(), players, DefaultBuraConfig())
}

// Reset ゲーム初期化
func (b *Bura) Reset() {
	b.phase = BuraPhasePlay
	b.gameEndFlag = false
	b.drawFlag = false
	b.winnerIdx = -1
	b.trickNumber = 0
	b.currentLead = nil
	b.currentResponse = nil
	b.trickCards = nil
	b.playerPoints = make([]int, len(b.players))
	b.actionLog = nil
	b.stock = nil
	b.trumpCard = nil

	for _, p := range b.players {
		p.ResetGame()
	}

	b.trumpCards.Shuffle()

	for range BuraHandSize {
		for _, p := range b.players {
			if c := b.trumpCards.DrawCard(); c != nil {
				p.AddCard(c)
			}
		}
	}

	b.stock = make([]*Card, 0)
	for {
		c := b.trumpCards.DrawCard()
		if c == nil {
			break
		}
		b.stock = append(b.stock, c)
	}

	// 山札の底のカードが切札を決める。表向きに置かれ、最後に引かれる。
	if len(b.stock) > 0 {
		b.trumpCard = b.stock[len(b.stock)-1]
		b.stock = b.stock[:len(b.stock)-1]
		b.trumpSuit = b.trumpCard.GetDesign()
	} else {
		b.trumpSuit = CardDesignSpade
	}

	b.leadPlayerIdx = 0
	b.currentPlayerIdx = 0
	b.addLog(-1, "deal", fmt.Sprintf("trump suit is %d", b.trumpSuit), []*Card{b.trumpCard})
}

// PlayCards idx のプレイヤーが手札の indices を出す。
// リード側は同一スートを 1〜3 枚、応じる側はリードと同じ枚数を出す。
func (b *Bura) PlayCards(idx int, indices []int) error {
	if b.gameEndFlag {
		return fmt.Errorf("the round is over")
	}
	if idx != b.currentPlayerIdx {
		return fmt.Errorf("it is not player %d's turn", idx)
	}
	p := b.GetPlayer(idx)
	if p == nil {
		return fmt.Errorf("no such player: %d", idx)
	}
	cards, err := b.peekCards(p, indices)
	if err != nil {
		return err
	}

	if len(b.currentLead) == 0 {
		if err := buraValidateLead(cards); err != nil {
			return err
		}
		b.currentLead = p.RemoveCards(indices)
		b.trickCards = append(b.trickCards, b.currentLead...)
		b.currentPlayerIdx = b.nextPlayer(idx)
		return nil
	}

	// 応じる側。枚数だけが制約で、勝てないカードを捨てるのも合法。
	if len(cards) != len(b.currentLead) {
		return fmt.Errorf("must play exactly %d card(s)", len(b.currentLead))
	}
	b.currentResponse = p.RemoveCards(indices)
	b.trickCards = append(b.trickCards, b.currentResponse...)
	b.resolveTrick()
	return nil
}

// peekCards indices が手札の有効な相異なる添字かを検証し、カードを返す。
// この時点では手札からは取り除かない (検証に失敗しても状態を壊さないため)。
func (b *Bura) peekCards(p *BuraPlayer, indices []int) ([]*Card, error) {
	if len(indices) == 0 {
		return nil, fmt.Errorf("must play at least one card")
	}
	seen := make(map[int]bool, len(indices))
	cards := make([]*Card, 0, len(indices))
	for _, i := range indices {
		if i < 0 || i >= p.GetCardsSize() {
			return nil, fmt.Errorf("card index %d out of range", i)
		}
		if seen[i] {
			return nil, fmt.Errorf("card index %d played twice", i)
		}
		seen[i] = true
		c := p.GetCard(i)
		if c == nil {
			return nil, fmt.Errorf("card index %d is empty", i)
		}
		cards = append(cards, c)
	}
	return cards, nil
}

// resolveTrick 現在のトリックを解決し、点を勝者へ加算して手札を補充する。
func (b *Bura) resolveTrick() {
	winner := b.leadPlayerIdx
	if buraBeatsCombination(b.currentResponse, b.currentLead, b.trumpSuit) {
		winner = b.nextPlayer(b.leadPlayerIdx)
	}

	taken := b.trickCards
	points := 0
	for _, c := range taken {
		points += BuraCardPoints(c)
	}
	b.playerPoints[winner] += points
	b.addLog(winner, "trick", fmt.Sprintf("wins trick %d (%d pt)", b.trickNumber+1, points), taken)

	b.trickNumber++
	b.currentLead = nil
	b.currentResponse = nil
	b.trickCards = nil

	b.replenish(winner)

	b.leadPlayerIdx = winner
	b.currentPlayerIdx = winner

	// 誰も手札を持たなくなったら、宣言のないまま流局。実ルールでは
	// 全員が改めてステークを出して同じ親が配り直す。
	//
	// 勝者の手札だけを見ないこと。2 人・均等補充では両者同時に尽きるので
	// 今は等価だが、それは条件ではなく偶然の一致でしかない。
	if b.allHandsEmpty() {
		b.drawFlag = true
		b.gameEndFlag = true
		b.phase = BuraPhaseGameEnd
		b.winnerIdx = -1
		b.currentPlayerIdx = -1
		b.addLog(-1, "draw", "stock exhausted with no claim -- the round is a draw", nil)
	}
}

// replenish トリックの勝者から順に、山札が続く限り手札を 3 枚へ戻す。
func (b *Bura) replenish(winner int) {
	for range BuraHandSize {
		for i := range b.players {
			idx := (winner + i) % len(b.players)
			p := b.players[idx]
			if p.GetCardsSize() >= BuraHandSize {
				continue
			}
			c := b.drawFromStock()
			if c == nil {
				return
			}
			p.AddCard(c)
		}
	}
}

// drawFromStock 山札から 1 枚引く。山札が尽きたら最後に切札指示カードを渡す。
func (b *Bura) drawFromStock() *Card {
	if len(b.stock) > 0 {
		c := b.stock[0]
		b.stock = b.stock[1:]
		return c
	}
	if b.trumpCard != nil {
		c := b.trumpCard
		b.trumpCard = nil
		return c
	}
	return nil
}

// Claim idx のプレイヤーが 31 点到達を宣言する。
// 到達していれば勝ち、足りなければその場で相手の勝ちとなる。
func (b *Bura) Claim(idx int) error {
	if b.gameEndFlag {
		return fmt.Errorf("the round is over")
	}
	if idx < 0 || idx >= len(b.players) {
		return fmt.Errorf("no such player: %d", idx)
	}
	b.gameEndFlag = true
	b.phase = BuraPhaseGameEnd
	b.currentPlayerIdx = -1

	if b.playerPoints[idx] >= BuraWinThreshold {
		b.winnerIdx = idx
		b.addLog(idx, "claim", fmt.Sprintf("claims %d points and wins", b.playerPoints[idx]), nil)
		return nil
	}
	b.winnerIdx = b.nextPlayer(idx)
	b.addLog(idx, "claim", fmt.Sprintf("claims with only %d points and forfeits", b.playerPoints[idx]), nil)
	return nil
}

// DeclareCombination idx のプレイヤーが手札の即勝ち役を宣言する。
// 役が成立していればその場で勝ち、していなければ何も起きずエラーを返す。
func (b *Bura) DeclareCombination(idx int) error {
	if b.gameEndFlag {
		return fmt.Errorf("the round is over")
	}
	p := b.GetPlayer(idx)
	if p == nil {
		return fmt.Errorf("no such player: %d", idx)
	}
	hand := make([]*Card, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		hand = append(hand, p.GetCard(i))
	}
	combo := BuraDetectCombination(hand, b.trumpSuit)
	if combo == BuraCombinationNone {
		return fmt.Errorf("no winning combination in hand")
	}
	b.gameEndFlag = true
	b.phase = BuraPhaseGameEnd
	b.winnerIdx = idx
	b.currentPlayerIdx = -1
	b.addLog(idx, "declare", fmt.Sprintf("declares combination %d and wins", combo), hand)
	return nil
}

// nextPlayer idx の次のプレイヤー添字を返す。
func (b *Bura) nextPlayer(idx int) int {
	return (idx + 1) % len(b.players)
}

// addLog アクションログへ 1 行追加する。
func (b *Bura) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	b.appendLog(playerIdx, actionType, detail, cards)
}

// ---- 公開アクセサ ----

// GetPlayers 全プレイヤーを返す。
func (b *Bura) GetPlayers() []*BuraPlayer { return b.players }

// GetPlayer idx のプレイヤーを返す。範囲外は nil。
func (b *Bura) GetPlayer(idx int) *BuraPlayer {
	return getPlayer(b.players, idx)
}

// GetStock 山札 (切札指示カードを含まない) を返す。
func (b *Bura) GetStock() []*Card { return b.stock }

// GetTrumpCard 表向きの切札指示カードを返す。山札から引かれた後は nil。
func (b *Bura) GetTrumpCard() *Card { return b.trumpCard }

// GetTrumpSuit 切札スートを返す。
func (b *Bura) GetTrumpSuit() int { return b.trumpSuit }

// GetPhase 現在のフェーズを返す。
func (b *Bura) GetPhase() BuraPhase { return b.phase }

// GetCurrentPlayerIdx 手番のプレイヤー添字を返す。終局後は -1。
func (b *Bura) GetCurrentPlayerIdx() int { return b.currentPlayerIdx }

// GetLeadPlayerIdx 現トリックのリード側プレイヤー添字を返す。
func (b *Bura) GetLeadPlayerIdx() int { return b.leadPlayerIdx }

// GetCurrentLead リードされたカードを返す。トリック解決後は空。
func (b *Bura) GetCurrentLead() []*Card { return b.currentLead }

// GetTrickNumber 完了したトリック数を返す。
func (b *Bura) GetTrickNumber() int { return b.trickNumber }

// GetPlayerPoints idx のプレイヤーの獲得点を返す。
func (b *Bura) GetPlayerPoints(idx int) int {
	if idx < 0 || idx >= len(b.playerPoints) {
		return 0
	}
	return b.playerPoints[idx]
}

// SetPlayerPoints idx のプレイヤーの獲得点を設定する (テスト用)。
func (b *Bura) SetPlayerPoints(idx, points int) {
	if idx < 0 || idx >= len(b.playerPoints) {
		return
	}
	b.playerPoints[idx] = points
}

// GetGameEndFlag ゲームが終了しているかを返す。
func (b *Bura) GetGameEndFlag() bool { return b.gameEndFlag }

// GetWinnerIdx 勝者の添字を返す。未確定または引き分けは -1。
func (b *Bura) GetWinnerIdx() int { return b.winnerIdx }

// IsDraw 宣言のないまま流局したかを返す。
func (b *Bura) IsDraw() bool { return b.drawFlag }

// GetConfig ゲーム設定を返す。
func (b *Bura) GetConfig() BuraConfig { return b.config }

// SetConfig ゲーム設定を差し替える。
func (b *Bura) SetConfig(c BuraConfig) { b.config = c }

// ---- JSON ----

// buraJSON is the KV wire format. Workers rebuild the game from this on every
// request, so anything omitted here is lost between calls.
type buraJSON struct {
	Players          []*BuraPlayer     `json:"pl"`
	Config           BuraConfig        `json:"cf"`
	Phase            BuraPhase         `json:"ph"`
	Stock            []*Card           `json:"st"`
	TrumpCard        *Card             `json:"tc"`
	TrumpSuit        int               `json:"ts"`
	TrickNumber      int               `json:"tn"`
	CurrentPlayerIdx int               `json:"cp"`
	LeadPlayerIdx    int               `json:"lp"`
	CurrentLead      []*Card           `json:"cl"`
	CurrentResponse  []*Card           `json:"cr"`
	TrickCards       []*Card           `json:"tk"`
	PlayerPoints     []int             `json:"pp"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	DrawFlag         bool              `json:"df"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (b *Bura) MarshalJSON() ([]byte, error) {
	return json.Marshal(buraJSON{
		Players:          b.players,
		Config:           b.config,
		Phase:            b.phase,
		Stock:            b.stock,
		TrumpCard:        b.trumpCard,
		TrumpSuit:        b.trumpSuit,
		TrickNumber:      b.trickNumber,
		CurrentPlayerIdx: b.currentPlayerIdx,
		LeadPlayerIdx:    b.leadPlayerIdx,
		CurrentLead:      b.currentLead,
		CurrentResponse:  b.currentResponse,
		TrickCards:       b.trickCards,
		PlayerPoints:     b.playerPoints,
		GameEndFlag:      b.gameEndFlag,
		WinnerIdx:        b.winnerIdx,
		DrawFlag:         b.drawFlag,
		ActionLog:        b.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// Workers are stateless, so this runs on untrusted KV bytes on every request.
// Indices are clamped rather than trusted: a corrupted currentPlayerIdx would
// otherwise index straight past the players slice.
func (b *Bura) UnmarshalJSON(data []byte) error {
	var j buraJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) == 0 {
		return fmt.Errorf("bura: no players in snapshot")
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("bura: %w", err)
	}
	b.players = j.Players
	b.config = j.Config
	b.phase = j.Phase
	b.stock = j.Stock
	b.trumpCard = j.TrumpCard
	b.trumpSuit = j.TrumpSuit
	b.trickNumber = j.TrickNumber
	b.currentLead = j.CurrentLead
	b.currentResponse = j.CurrentResponse
	b.trickCards = j.TrickCards
	b.gameEndFlag = j.GameEndFlag
	b.drawFlag = j.DrawFlag
	b.actionLog = j.ActionLog

	n := len(b.players)
	b.currentPlayerIdx = clampPlayerIdx(j.CurrentPlayerIdx, n)
	b.leadPlayerIdx = clampPlayerIdx(j.LeadPlayerIdx, n)
	b.winnerIdx = clampPlayerIdx(j.WinnerIdx, n)

	b.playerPoints = make([]int, n)
	copy(b.playerPoints, j.PlayerPoints)

	if b.trumpCards == nil {
		b.trumpCards = NewTrumpCardsShortDeck()
	}
	return nil
}

// clampPlayerIdx 範囲外のプレイヤー添字を -1 (未確定) に丸める。
func clampPlayerIdx(idx, n int) int {
	if idx < 0 || idx >= n {
		return -1
	}
	return idx
}

// allHandsEmpty 全プレイヤーの手札が尽きているかを返す。
func (b *Bura) allHandsEmpty() bool {
	return allHandsEmpty(b.players)
}

// ---- CPU ----

// BuraCpuAction CPU が選んだ手。
type BuraCpuAction struct {
	// Declare が true なら即勝ち役を宣言する。
	Declare bool
	// Claim が true なら 31 点到達を宣言する。
	Claim bool
	// Indices は出すカードの手札添字 (Declare/Claim が false のとき有効)。
	Indices []int
}

// BuraCpuDecide idx の CPU が取る手を決める。
//
// 優先順位は 即勝ち役 > 31点宣言 > カードプレイ。宣言は誤ると即負けなので、
// 到達を確認できたときだけ行う (Claim の実装は誤申告を罰する)。
func (b *Bura) BuraCpuDecide(idx int) BuraCpuAction {
	p := b.GetPlayer(idx)
	if p == nil || p.GetCardsSize() == 0 {
		return BuraCpuAction{}
	}
	hand := b.handOf(idx)

	if BuraDetectCombination(hand, b.trumpSuit) != BuraCombinationNone {
		return BuraCpuAction{Declare: true}
	}
	if b.playerPoints[idx] >= BuraWinThreshold {
		return BuraCpuAction{Claim: true}
	}
	if len(b.currentLead) == 0 {
		return BuraCpuAction{Indices: b.cpuChooseLead(hand)}
	}
	return BuraCpuAction{Indices: b.cpuChooseResponse(hand)}
}

// handOf idx の手札をスライスで返す。
func (b *Bura) handOf(idx int) []*Card {
	p := b.GetPlayer(idx)
	if p == nil {
		return nil
	}
	hand := make([]*Card, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		hand = append(hand, p.GetCard(i))
	}
	return hand
}

// cpuChooseLead リードするカードの添字を選ぶ。
//
// 同一スートをまとめて出せると一度に大量の点を取れるので、非切札で最も
// 点の高いスート群をまとめて出す。まとまりがなければ最も安いカード 1 枚を
// 捨ててこちらの強い札を温存する。
func (b *Bura) cpuChooseLead(hand []*Card) []int {
	bySuit := map[int][]int{}
	for i, c := range hand {
		if c == nil {
			continue
		}
		bySuit[c.GetDesign()] = append(bySuit[c.GetDesign()], i)
	}

	bestIdx, bestPoints := []int(nil), -1
	for suit, indices := range bySuit {
		if len(indices) < 2 || suit == b.trumpSuit {
			continue
		}
		points := 0
		for _, i := range indices {
			points += BuraCardPoints(hand[i])
		}
		if points > bestPoints {
			bestPoints, bestIdx = points, indices
		}
	}
	if bestIdx != nil {
		return bestIdx
	}

	cheapest, cheapestRank := 0, -1
	for i, c := range hand {
		if c == nil {
			continue
		}
		rank := BuraRankOrder(c)
		if c.GetDesign() == b.trumpSuit {
			rank += 100 // 切札は温存する
		}
		if cheapestRank == -1 || rank < cheapestRank {
			cheapest, cheapestRank = i, rank
		}
	}
	return []int{cheapest}
}

// cpuChooseResponse リードに応じるカードの添字を選ぶ。
//
// 受けきれる組み合わせがあれば、そのうち最も安いものを選ぶ。受けきれない
// ときは最も点の低いカードを捨てる (フォロー義務はないので何を出してもよい)。
func (b *Bura) cpuChooseResponse(hand []*Card) []int {
	n := len(b.currentLead)
	if n == 0 || len(hand) < n {
		return nil
	}

	best, bestCost := []int(nil), 0
	for _, combo := range buraCombinations(len(hand), n) {
		cards := make([]*Card, 0, n)
		for _, i := range combo {
			cards = append(cards, hand[i])
		}
		cost := 0
		for _, c := range cards {
			cost += BuraCardPoints(c) + BuraRankOrder(c)
			if c.GetDesign() == b.trumpSuit {
				cost += 20 // 切札を使うのは高くつく
			}
		}
		if !buraBeatsCombination(cards, b.currentLead, b.trumpSuit) {
			continue
		}
		if best == nil || cost < bestCost {
			best, bestCost = combo, cost
		}
	}
	if best != nil {
		return best
	}

	// 受けきれない。最も安い n 枚を捨てる。
	discard, discardCost := []int(nil), 0
	for _, combo := range buraCombinations(len(hand), n) {
		cost := 0
		for _, i := range combo {
			cost += BuraCardPoints(hand[i])*10 + BuraRankOrder(hand[i])
		}
		if discard == nil || cost < discardCost {
			discard, discardCost = combo, cost
		}
	}
	return discard
}

// buraCombinations 0..size-1 から k 個を選ぶ全組み合わせを返す。
// 手札は最大 3 枚なので候補は高々 3 通りしかない。
func buraCombinations(size, k int) [][]int {
	var out [][]int
	cur := make([]int, 0, k)
	var rec func(start int)
	rec = func(start int) {
		if len(cur) == k {
			pick := make([]int, k)
			copy(pick, cur)
			out = append(out, pick)
			return
		}
		for i := start; i < size; i++ {
			cur = append(cur, i)
			rec(i + 1)
			cur = cur[:len(cur)-1]
		}
	}
	rec(0)
	return out
}

// SetTrumpSuit 切札スートを設定する (テスト用)。
func (b *Bura) SetTrumpSuit(suit int) { b.trumpSuit = suit }

// SetCurrentPlayerIdx 手番のプレイヤーを設定する (テスト用)。
func (b *Bura) SetCurrentPlayerIdx(idx int) { b.currentPlayerIdx = idx }

// SetCurrentLead リードされたカードを設定する (テスト用)。
func (b *Bura) SetCurrentLead(cards []*Card) { b.currentLead = cards }
