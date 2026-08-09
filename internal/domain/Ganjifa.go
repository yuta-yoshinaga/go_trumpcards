//go:build !js || !wasm || extra

// Package domain ガンジファ (Ganjifa) のドメインモデル。
//
// ムガル帝国期インドの円形カードによるトリックテイキング。**8 スート 96 枚**
// (各スート 1..12) を 3 人に 32 枚ずつ配り、32 トリックを戦う。
//
// このゲームを他と分けているのは **スート群でランクの向きが逆転する** ことに
// 尽きる:
//
//   - **強いスート群** (design 1..4): 数字が大きいほど強い (12 が最強)
//   - **弱いスート群** (design 5..8): 数字が**小さい**ほど強い (1 が最強)
//
// 同じ「3」が、強い群では下から 3 番目、弱い群では上から 3 番目になる。序列を
// 一本の関数で書くと必ずどちらかを取り違えるので、比較は必ず
// [GanjifaCardStrength] を通す。
//
// 標準 52 枚の Card/Deck は design 1..4 しか持たないが、**新しいデッキ型は
// 要らない**。FrenchTarot が切り札を design 5、エクスキューズを design 6 と
// 仮想値で表しているのと同じ手で、5..8 を弱いスート群に割り当てる。描画は
// ADR-0033 の手続き的 CardFace パスがラベルと色で受け持つ。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// GanjifaPlayerCnt プレイヤー数 (人間 1 + CPU 2)。
const GanjifaPlayerCnt = 3

// GanjifaSuitCnt スート数。前半 4 つが強い群、後半 4 つが弱い群。
const GanjifaSuitCnt = 8

// GanjifaRankCnt 各スートの枚数 (1..12)。
const GanjifaRankCnt = 12

// GanjifaDeckSize デッキ枚数 (8 スート × 12 枚)。
const GanjifaDeckSize = GanjifaSuitCnt * GanjifaRankCnt

// GanjifaHandSize 各プレイヤーの手札枚数 (96 / 3)。
const GanjifaHandSize = GanjifaDeckSize / GanjifaPlayerCnt

// GanjifaTrickCount 1 ラウンドのトリック数。
const GanjifaTrickCount = GanjifaHandSize

// GanjifaStrongSuitMax これ以下の design が強いスート群。
//
// **この境界がゲームの中心。**1..4 は数字が大きいほど強く、5..8 は小さいほど
// 強い。境界を動かすと同じ札の意味が反転する。
const GanjifaStrongSuitMax = 4

// GanjifaCardStrength 札の強さを返す。数値が大きいほど強い。
//
// 弱いスート群では順序を反転させる。**生の value で比較してはならない** ——
// 弱い群の 1 は最強札で、12 は最弱札になる。
func GanjifaCardStrength(design, value int) int {
	if design <= GanjifaStrongSuitMax {
		return value
	}
	return GanjifaRankCnt + 1 - value
}

// GanjifaIsStrongSuit 強いスート群か。
func GanjifaIsStrongSuit(design int) bool { return design <= GanjifaStrongSuitMax }

// ganjifaSuitNames はムガル・ガンジファ 8 スートの慣用名 (design 1..8)。
//
// **強い群／弱い群の割り当ては本実装の取り決め**であって史料の断定ではない。
// 史料によってスートの並びは異なるため、ここでは design 1..4 を強い群、
// 5..8 を弱い群として固定し、名前だけを慣用名から採っている。
var ganjifaSuitNames = [GanjifaSuitCnt + 1]string{
	"", "Taj", "Shamsher", "Ashrafi", "Ghulam", "Chang", "Surkh", "Barat", "Qimash",
}

// ganjifaSuitGlyphs は各スートの表示記号 (design 1..8)。
//
// **絵文字は使わない。**CUI は等幅前提で桁を揃えるので、全角に化ける絵文字を
// 入れると 96 枚の手札一覧がずれる。BMP の記号だけを使う。
var ganjifaSuitGlyphs = [GanjifaSuitCnt + 1]string{
	"", "\u265b", "\u2020", "\u25c9", "\u265f", "\u266a", "\u2726", "\u25a4", "\u25a7",
}

// GanjifaSuitName はスートの慣用名を返す。範囲外は空文字。
func GanjifaSuitName(design int) string {
	if design < 1 || design > GanjifaSuitCnt {
		return ""
	}
	return ganjifaSuitNames[design]
}

// GanjifaSuitGlyph はスートの表示記号を返す。範囲外は空文字。
func GanjifaSuitGlyph(design int) string {
	if design < 1 || design > GanjifaSuitCnt {
		return ""
	}
	return ganjifaSuitGlyphs[design]
}

// GanjifaPhase ゲームフェーズ。
type GanjifaPhase int

const (
	// GanjifaPhasePlay トリックプレイ。
	GanjifaPhasePlay GanjifaPhase = iota
	// GanjifaPhaseTrickEnd トリック終了 (結果表示待ち)。
	GanjifaPhaseTrickEnd
	// GanjifaPhaseRoundEnd ラウンド終了。
	GanjifaPhaseRoundEnd
	// GanjifaPhaseGameEnd マッチ終了。
	GanjifaPhaseGameEnd
)

// Ganjifa ガンジファのゲーム本体。
type Ganjifa struct {
	players          []*GanjifaPlayer
	config           GanjifaConfig
	rng              *rand.Rand
	deck             []*Card
	phase            GanjifaPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	trumpSuit        int
	roundTricks      [GanjifaPlayerCnt]int
	playerScores     [GanjifaPlayerCnt]int
	gameEndFlag      bool
	winnerPlayer     int // -1 = 未確定 (同点)
	actionLog        []*ActionLogEntry
}

// ganjifaJSON is the JSON wire format for Ganjifa.
//
// **全フィールドが非公開なので専用のコーデックが要る。**これを省くと
// Cloudflare Worker のセッション復元が空のゲームを返し、リクエストのたびに
// 手札も点も消える。deck / rng は載せない —— 配り終えた後の deck は残り札を
// 持たず、rng は復元後に張り直せばよい。
type ganjifaJSON struct {
	Players          []*GanjifaPlayer      `json:"pl"`
	Config           GanjifaConfig         `json:"cfg"`
	Phase            GanjifaPhase          `json:"ph"`
	RoundNumber      int                   `json:"rn"`
	TrickNumber      int                   `json:"tn"`
	CurrentPlayerIdx int                   `json:"cpi"`
	CurrentTrick     []*TrickCard          `json:"ct"`
	LeadPlayerIdx    int                   `json:"lpi"`
	DealerIdx        int                   `json:"di"`
	TrumpSuit        int                   `json:"ts"`
	RoundTricks      [GanjifaPlayerCnt]int `json:"rt"`
	PlayerScores     [GanjifaPlayerCnt]int `json:"ps"`
	GameEndFlag      bool                  `json:"gef"`
	WinnerPlayer     int                   `json:"wp"`
	ActionLog        []*ActionLogEntry     `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Ganjifa) MarshalJSON() ([]byte, error) {
	return json.Marshal(ganjifaJSON{
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		RoundNumber:      g.roundNumber,
		TrickNumber:      g.trickNumber,
		CurrentPlayerIdx: g.currentPlayerIdx,
		CurrentTrick:     g.currentTrick,
		LeadPlayerIdx:    g.leadPlayerIdx,
		DealerIdx:        g.dealerIdx,
		TrumpSuit:        g.trumpSuit,
		RoundTricks:      g.roundTricks,
		PlayerScores:     g.playerScores,
		GameEndFlag:      g.gameEndFlag,
		WinnerPlayer:     g.winnerPlayer,
		ActionLog:        g.actionLog,
	})
}

// ganjifaMaxSliceLen caps slice sizes during deserialisation.
const ganjifaMaxSliceLen = 5000

// errGanjifaOversized is the single sentinel error for oversized input arrays.
var errGanjifaOversized = errors.New("ganjifa: input array exceeds maximum allowed size")

// errGanjifaInvalidPlayers is returned when restored state lacks exactly GanjifaPlayerCnt players.
var errGanjifaInvalidPlayers = errors.New("ganjifa: invalid player count")

// errGanjifaInvalidTrick is returned when a restored trick card is nil/out of range.
var errGanjifaInvalidTrick = errors.New("ganjifa: invalid trick card")

// errGanjifaInvalidState is returned when a restored index/state field is out of range.
var errGanjifaInvalidState = errors.New("ganjifa: invalid state values in json")

// UnmarshalJSON implements json.Unmarshaler.
//
// **切り札は 1..8 を許す。**標準 52 枚デッキ用のバリデーション (1..4) を
// そのまま持ち込むと、弱い群が切り札のセッションが復元できなくなる。
func (g *Ganjifa) UnmarshalJSON(data []byte) error {
	var j ganjifaJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > ganjifaMaxSliceLen || len(j.CurrentTrick) > ganjifaMaxSliceLen ||
		len(j.ActionLog) > ganjifaMaxSliceLen {
		return errGanjifaOversized
	}
	if len(j.Players) != GanjifaPlayerCnt {
		return errGanjifaInvalidPlayers
	}
	for _, p := range j.Players {
		if p == nil {
			return errGanjifaInvalidPlayers
		}
	}
	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= GanjifaPlayerCnt ||
		j.LeadPlayerIdx < 0 || j.LeadPlayerIdx >= GanjifaPlayerCnt ||
		j.DealerIdx < 0 || j.DealerIdx >= GanjifaPlayerCnt ||
		j.WinnerPlayer < -1 || j.WinnerPlayer >= GanjifaPlayerCnt ||
		j.TrumpSuit < 0 || j.TrumpSuit > GanjifaSuitCnt ||
		j.RoundNumber < 1 ||
		j.TrickNumber < 1 || j.TrickNumber > GanjifaTrickCount ||
		j.Phase < GanjifaPhasePlay || j.Phase > GanjifaPhaseGameEnd {
		return errGanjifaInvalidState
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil || tc.PlayerIdx < 0 || tc.PlayerIdx >= GanjifaPlayerCnt {
			return errGanjifaInvalidTrick
		}
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.roundNumber = j.RoundNumber
	g.trickNumber = j.TrickNumber
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.currentTrick = j.CurrentTrick
	if g.currentTrick == nil {
		g.currentTrick = make([]*TrickCard, 0)
	}
	g.leadPlayerIdx = j.LeadPlayerIdx
	g.dealerIdx = j.DealerIdx
	g.trumpSuit = j.TrumpSuit
	g.roundTricks = j.RoundTricks
	g.playerScores = j.PlayerScores
	g.gameEndFlag = j.GameEndFlag
	g.winnerPlayer = j.WinnerPlayer
	g.actionLog = j.ActionLog
	// **復元したら必ず乱数源を張り直す。**Cloudflare Worker は毎リクエスト KV から
	// 組み直すので SetRand は一度も呼ばれない。rng を nil のままにすると、
	// シャッフル以外で rng を使う経路 (CPU の乱択など) が nil デリファレンスで
	// 落ちる。呼び出し側ごとにガードするのではなく、ここで構造的に潰す。
	g.rng = rand.New(rand.NewSource(rand.Int63()))
	return nil
}

// NewGanjifa コンストラクタ。
func NewGanjifa(players []*GanjifaPlayer, config GanjifaConfig) *Ganjifa {
	return &Ganjifa{
		players: players,
		config:  config,
		// **本番の乱数源はここで入れる。**入れ忘れると shuffle() の nil 分岐に
		// 落ちて山が一切並べ替わらず、毎回同じ 96 枚が同じ席に配られる。
		// 種は rand.Int63() から取る —— time.Now().UnixNano() だと同じ
		// ナノ秒に作った 2 局が同じ配りになりうる。
		rng:          rand.New(rand.NewSource(rand.Int63())),
		winnerPlayer: -1,
	}
}

// NewDefaultGanjifa 標準の 3 人構成 (人間 1, CPU 2) と既定設定で生成する。
func NewDefaultGanjifa() *Ganjifa {
	players := make([]*GanjifaPlayer, GanjifaPlayerCnt)
	players[0] = NewGanjifaPlayer(true)
	for i := 1; i < GanjifaPlayerCnt; i++ {
		players[i] = NewGanjifaPlayer(false)
	}
	return NewGanjifa(players, DefaultGanjifaConfig())
}

// SetRand 乱数生成器を差し替える (テスト用)。
func (g *Ganjifa) SetRand(r *rand.Rand) { g.rng = r }

// buildGanjifaDeck 96 枚デッキを構築する。design 1..8、value 1..12。
func buildGanjifaDeck() []*Card {
	deck := make([]*Card, 0, GanjifaDeckSize)
	for suit := 1; suit <= GanjifaSuitCnt; suit++ {
		for val := 1; val <= GanjifaRankCnt; val++ {
			deck = append(deck, NewCard(suit, val, false))
		}
	}
	return deck
}

// Reset ゲーム初期化。
func (g *Ganjifa) Reset() {
	g.gameEndFlag = false
	g.winnerPlayer = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.playerScores = [GanjifaPlayerCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のラウンドを開始する。規定局数に達していればマッチを終える。
func (g *Ganjifa) NextRound() {
	if g.phase != GanjifaPhaseRoundEnd {
		return
	}
	if g.roundNumber >= g.config.TargetRounds {
		g.finishMatch()
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % GanjifaPlayerCnt
	g.startRound()
}

// startRound 手札を配り、切り札を決めてプレイを始める。
func (g *Ganjifa) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.roundTricks = [GanjifaPlayerCnt]int{}
	for _, p := range g.players {
		p.ResetRound()
	}

	g.deck = buildGanjifaDeck()
	g.shuffle()
	g.deal()
	g.sortAllHands()

	// 切り札はディーラーの最後の札のスート。局ごとに変わる。
	g.trumpSuit = g.chooseTrump()
	g.leadPlayerIdx = (g.dealerIdx + 1) % GanjifaPlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = GanjifaPhasePlay
	g.appendLog(-1, "deal",
		fmt.Sprintf("ラウンド %d 開始 (切り札 %d)", g.roundNumber, g.trumpSuit), nil)
}

// shuffle 山札をシャッフルする。
func (g *Ganjifa) shuffle() {
	// **nil は「並べ替えない」ではなく異常。**コンストラクタが必ず種を入れる
	// ので、ここに来るのはゼロ値の Ganjifa を直接組んだ場合だけ。黙って
	// 素通りさせると、その山は未シャッフルのまま配られる。
	if g.rng == nil {
		g.rng = rand.New(rand.NewSource(rand.Int63()))
	}
	g.rng.Shuffle(len(g.deck), func(i, j int) { g.deck[i], g.deck[j] = g.deck[j], g.deck[i] })
}

// deal 各プレイヤーへ 32 枚ずつ配る。
func (g *Ganjifa) deal() {
	idx := 0
	for i := 0; i < GanjifaHandSize; i++ {
		for j := 0; j < GanjifaPlayerCnt; j++ {
			seat := (g.dealerIdx + 1 + j) % GanjifaPlayerCnt
			if idx < len(g.deck) {
				g.players[seat].AddCard(g.deck[idx])
				idx++
			}
		}
	}
}

// chooseTrump 切り札スートを決める。ディーラーの手札で最も多いスート。
func (g *Ganjifa) chooseTrump() int {
	counts := make([]int, GanjifaSuitCnt+1)
	p := g.players[g.dealerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		counts[p.GetCard(i).GetDesign()]++
	}
	best, bestCnt := 1, -1
	for suit := 1; suit <= GanjifaSuitCnt; suit++ {
		if counts[suit] > bestCnt {
			best, bestCnt = suit, counts[suit]
		}
	}
	return best
}

// sortAllHands 全員の手札をスート・強さ順に整列する。
func (g *Ganjifa) sortAllHands() {
	for _, p := range g.players {
		cards := make([]*Card, 0, p.GetCardsSize())
		for i := 0; i < p.GetCardsSize(); i++ {
			cards = append(cards, p.GetCard(i))
		}
		sort.SliceStable(cards, func(a, b int) bool {
			if cards[a].GetDesign() != cards[b].GetDesign() {
				return cards[a].GetDesign() < cards[b].GetDesign()
			}
			// **群ごとに向きが違うので強さで並べる。**value で並べると
			// 弱い群だけ逆順に見えて、手札が読みにくくなる。
			return GanjifaCardStrength(cards[a].GetDesign(), cards[a].GetValue()) <
				GanjifaCardStrength(cards[b].GetDesign(), cards[b].GetValue())
		})
		p.Reset()
		for _, c := range cards {
			p.AddCard(c)
		}
	}
}

// appendLog 棋譜に 1 行追加する。
func (g *Ganjifa) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: len(g.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// finishMatch マッチを終え、獲得トリック数の累計が最大のプレイヤーを勝者にする。
func (g *Ganjifa) finishMatch() {
	g.gameEndFlag = true
	g.phase = GanjifaPhaseGameEnd
	best := g.playerScores[0]
	winner, tie := 0, false
	for i := 1; i < GanjifaPlayerCnt; i++ {
		switch {
		case g.playerScores[i] > best:
			best, winner, tie = g.playerScores[i], i, false
		case g.playerScores[i] == best:
			tie = true
		}
	}
	// **同点なら勝者なし。**席順で決めると若い席が常に得をする。
	if tie {
		g.winnerPlayer = -1
	} else {
		g.winnerPlayer = winner
	}
	g.appendLog(-1, "gameend", "マッチ終了", nil)
}

// IsHumanTurn 人間のプレイ手番か。
func (g *Ganjifa) IsHumanTurn() bool {
	return g.phase == GanjifaPhasePlay && g.players[g.currentPlayerIdx].GetIsHuman()
}

// PlayerPlay 人間プレイヤーが手札の 1 枚を出す。
func (g *Ganjifa) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != GanjifaPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	if err := g.validatePlay(g.currentPlayerIdx, player.GetCard(cardIndex)); err != nil {
		return err
	}
	g.playCard(g.currentPlayerIdx, player.RemoveCard(cardIndex))
	return nil
}

// CpuPlay 現在の手番が CPU の場合に 1 枚出す。
func (g *Ganjifa) CpuPlay() {
	if g.gameEndFlag || g.phase != GanjifaPhasePlay {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	played := g.players[idx].RemoveCard(g.cpuSelectPlayCard(idx))
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	g.playCard(idx, played)
}

// validatePlay マストフォローを検証する。
func (g *Ganjifa) validatePlay(playerIdx int, card *Card) error {
	if card == nil {
		return NewDomainError(ErrInvalidCard, "カードがありません")
	}
	if len(g.currentTrick) == 0 {
		return nil
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if g.playerHasSuit(playerIdx, leadSuit) && card.GetDesign() != leadSuit {
		return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
	}
	return nil
}

// playerHasSuit プレイヤーが指定スートを持っているか。
func (g *Ganjifa) playerHasSuit(playerIdx, design int) bool {
	return handHasSuit(g.players[playerIdx], design)
}

// GetValidPlayIndices 出せる手札の位置を返す。
func (g *Ganjifa) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	p := g.players[playerIdx]
	var all []int
	for i := 0; i < p.GetCardsSize(); i++ {
		all = append(all, i)
	}
	if len(g.currentTrick) == 0 {
		return all
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	var follow []int
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == leadSuit {
			follow = append(follow, i)
		}
	}
	if len(follow) > 0 {
		return follow
	}
	return all
}

// GetPlayableIndices プレイ可能なカードの位置。
func (g *Ganjifa) GetPlayableIndices(playerIdx int) []int { return g.GetValidPlayIndices(playerIdx) }

// playCard 1 枚を場に出し、トリックが揃えばフェーズを進める。
func (g *Ganjifa) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", "", []*Card{card})
	if len(g.currentTrick) < GanjifaPlayerCnt {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % GanjifaPlayerCnt
		return
	}
	g.phase = GanjifaPhaseTrickEnd
}

// ResolveTrick トリックの勝者を決める。
func (g *Ganjifa) ResolveTrick() {
	if g.phase != GanjifaPhaseTrickEnd {
		return
	}
	winnerIdx := g.trickWinner()
	cards := make([]*Card, len(g.currentTrick))
	for i, tc := range g.currentTrick {
		cards[i] = tc.Card
	}
	g.players[winnerIdx].AddTrick(cards)
	g.roundTricks[winnerIdx]++
	g.appendLog(winnerIdx, "trickwin", fmt.Sprintf("トリック %d を獲得", g.trickNumber), cards)

	g.leadPlayerIdx = winnerIdx
	g.currentPlayerIdx = winnerIdx
	if g.trickNumber >= GanjifaTrickCount {
		g.settleRound()
	}
}

// NextTrick 次のトリックを開始する。
func (g *Ganjifa) NextTrick() {
	if g.phase != GanjifaPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = GanjifaPhasePlay
}

// trickWinner 現在のトリックの勝者を返す。
func (g *Ganjifa) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	bestRank := g.ganjifaRank(g.currentTrick[0].Card)
	for _, tc := range g.currentTrick[1:] {
		if tc.Card.GetDesign() != g.trumpSuit && tc.Card.GetDesign() != leadSuit {
			continue
		}
		if r := g.ganjifaRank(tc.Card); r > bestRank {
			bestRank = r
			winnerIdx = tc.PlayerIdx
		}
	}
	return winnerIdx
}

// ganjifaRank トリック比較用ランク。切り札は非切り札より常に強い。
//
// **強さは必ず GanjifaCardStrength を通す。**弱いスート群では 1 が最強で
// 12 が最弱なので、生の value で比べると群ごとに逆の結果になる。
func (g *Ganjifa) ganjifaRank(card *Card) int {
	base := GanjifaCardStrength(card.GetDesign(), card.GetValue())
	if card.GetDesign() == g.trumpSuit {
		return 100 + base
	}
	return base
}

// cpuSelectPlayCard CPU が出す札を選ぶ。
func (g *Ganjifa) cpuSelectPlayCard(idx int) int {
	valid := g.GetValidPlayIndices(idx)
	if len(valid) == 0 {
		return 0
	}
	if g.config.CpuDifficulty == GanjifaCpuDifficultyEasy {
		return valid[0]
	}
	p := g.players[idx]
	// リードなら一番強い札、追随なら勝てる中で一番弱い札。
	if len(g.currentTrick) == 0 {
		best := valid[0]
		for _, i := range valid[1:] {
			if g.ganjifaRank(p.GetCard(i)) > g.ganjifaRank(p.GetCard(best)) {
				best = i
			}
		}
		return best
	}
	target := g.ganjifaRank(g.currentTrick[0].Card)
	for _, tc := range g.currentTrick[1:] {
		if r := g.ganjifaRank(tc.Card); r > target {
			target = r
		}
	}
	win := -1
	for _, i := range valid {
		r := g.ganjifaRank(p.GetCard(i))
		if r <= target {
			continue
		}
		if win < 0 || r < g.ganjifaRank(p.GetCard(win)) {
			win = i
		}
	}
	if win >= 0 {
		return win
	}
	// 勝てないので一番弱い札を捨てる。
	best := valid[0]
	for _, i := range valid[1:] {
		if g.ganjifaRank(p.GetCard(i)) < g.ganjifaRank(p.GetCard(best)) {
			best = i
		}
	}
	return best
}

// settleRound ラウンドを精算する。獲得トリック数をそのまま持ち点に足す。
func (g *Ganjifa) settleRound() {
	g.phase = GanjifaPhaseRoundEnd
	for i := range g.playerScores {
		g.playerScores[i] += g.roundTricks[i]
	}
	g.appendLog(-1, "settle", fmt.Sprintf("ラウンド %d 終了", g.roundNumber), nil)
}

// ScoreRound ラウンドを締め、規定局数ならマッチを終える。
func (g *Ganjifa) ScoreRound() {
	if g.phase != GanjifaPhaseRoundEnd {
		return
	}
	if g.roundNumber >= g.config.TargetRounds {
		g.finishMatch()
	}
}

// GanjifaHint ヒント情報。
type GanjifaHint struct {
	CardIndices []int  // 推奨カードインデックス
	Reason      string // ヒント理由キー
}

// GetPhase 現在のフェーズ。
func (g *Ganjifa) GetPhase() GanjifaPhase { return g.phase }

// SetPhase フェーズを設定する (テスト用)。
func (g *Ganjifa) SetPhase(p GanjifaPhase) { g.phase = p }

// GetRoundNumber 現在のラウンド番号。
func (g *Ganjifa) GetRoundNumber() int { return g.roundNumber }

// GetTrickNumber 現在のトリック番号。
func (g *Ganjifa) GetTrickNumber() int { return g.trickNumber }

// GetCurrentPlayerIdx 現在の手番。
func (g *Ganjifa) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx 手番を設定する (テスト用)。
func (g *Ganjifa) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 場に出ている札。
func (g *Ganjifa) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// GetLeadPlayerIdx リードした席。
func (g *Ganjifa) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// GetDealerIdx ディーラーの席。
func (g *Ganjifa) GetDealerIdx() int { return g.dealerIdx }

// GetTrumpSuit 切り札スート。
func (g *Ganjifa) GetTrumpSuit() int { return g.trumpSuit }

// GetRoundTricks このラウンドで各席が取ったトリック数。
func (g *Ganjifa) GetRoundTricks() [GanjifaPlayerCnt]int { return g.roundTricks }

// GetPlayerScores 各席の累計トリック数。
func (g *Ganjifa) GetPlayerScores() [GanjifaPlayerCnt]int { return g.playerScores }

// GetGameEndFlag マッチが終了したか。
func (g *Ganjifa) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerPlayer 勝者の席。-1 は未確定または同点。
func (g *Ganjifa) GetWinnerPlayer() int { return g.winnerPlayer }

// GetPlayers 全プレイヤー。
func (g *Ganjifa) GetPlayers() []*GanjifaPlayer { return g.players }

// GetPlayerCnt プレイヤー数。
func (g *Ganjifa) GetPlayerCnt() int { return len(g.players) }

// GetPlayer 指定席のプレイヤー。範囲外は nil。
func (g *Ganjifa) GetPlayer(idx int) *GanjifaPlayer {
	return getPlayer(g.players, idx)
}

// GetConfig 設定。
func (g *Ganjifa) GetConfig() GanjifaConfig { return g.config }

// SetConfig 設定を差し替える。検証は呼び出し側の責務 (共通ヘルパーが先に通す)。
func (g *Ganjifa) SetConfig(c GanjifaConfig) { g.config = c }

// GetActionLog 棋譜。
func (g *Ganjifa) GetActionLog() []*ActionLogEntry {
	return sliceOrEmpty(g.actionLog)
}

// GetHint 人間プレイヤーへのヒント。手番でなければ nil。
func (g *Ganjifa) GetHint() *GanjifaHint {
	human := findHumanIdx(g.players)
	if human < 0 || g.phase != GanjifaPhasePlay || g.currentPlayerIdx != human {
		return nil
	}
	valid := g.GetValidPlayIndices(human)
	if len(valid) == 0 {
		return nil
	}
	idx := g.cpuSelectPlayCard(human)
	return &GanjifaHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
}

// playHintReason ヒント理由キーを判定する。
//
// **弱いスート群の札を勧めるときは、それが分かる理由を返す。**強い群の感覚で
// 「小さい札を出せ」と読むと逆の手になるので、群の別をキーに含める。
func (g *Ganjifa) playHintReason(playerIdx, chosenIdx int) string {
	card := g.players[playerIdx].GetCard(chosenIdx)
	if card == nil {
		return "lead_low"
	}
	weak := !GanjifaIsStrongSuit(card.GetDesign())
	if len(g.currentTrick) == 0 {
		if weak {
			return "lead_weak_suit"
		}
		return "lead_high"
	}
	if card.GetDesign() == g.trumpSuit {
		return "follow_trump"
	}
	if weak {
		return "follow_weak_suit"
	}
	return "follow_win"
}
