package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// GoFishPlayerCnt Go Fishプレイヤー数
const GoFishPlayerCnt = 4

// GoFishInitialHandSize 初期手札枚数 (4人以上は5枚)
const GoFishInitialHandSize = 5

// GoFishBookSize ブック完成に必要な同ランク枚数
const GoFishBookSize = 4

// GoFishTotalBooks ブック総数 (A-K の13種)
const GoFishTotalBooks = 13

// goFishShuffleCount シャッフル回数
const goFishShuffleCount = 10

// GoFishPhase ゲームフェーズ
type GoFishPhase int

// GoFishのフェーズ定数
const (
	GoFishPhasePlay    GoFishPhase = iota // プレイ中
	GoFishPhaseGameEnd                    // ゲーム終了
)

// GoFish固有エラー
var (
	ErrGoFishNotYourTurn   = errors.New("not your turn")
	ErrGoFishInvalidTarget = errors.New("invalid target player")
	ErrGoFishInvalidRank   = errors.New("you do not hold that rank")
	ErrGoFishGameEnded     = errors.New("game has ended")
	ErrGoFishAskSelf       = errors.New("cannot ask yourself")
)

// GoFishCpuAction CPUの1回の要求記録
type GoFishCpuAction struct {
	AskPlayerIdx  int   // 要求したプレイヤーインデックス
	AskTargetIdx  int   // 要求された相手のインデックス
	AskRank       int   // 要求したランク
	Success       bool  // 相手がカードを持っていたか
	CardsReceived int   // 受け取ったカード枚数
	DrawnCard     *Card // Go Fishで引いたカード (nil=要求成功)
	BookFormed    bool  // ブックが完成したか
	BookRank      int   // 完成したブックのランク (BookFormed=falseなら0)
}

// goFishMemoryEntry CPU記憶エントリ
type goFishMemoryEntry struct {
	askerIdx  int  // 要求者
	targetIdx int  // 要求先
	rank      int  // 要求ランク
	hadCards  bool // 相手が持っていたか
	turnSeen  int  // 記憶したターン
}

// GetTurnSeen MemoryEntry インタフェース
func (e goFishMemoryEntry) GetTurnSeen() int { return e.turnSeen }

// goFishMemoryEntryJSON is the wire format for goFishMemoryEntry. The struct
// fields are unexported, so without an explicit marshaller the entries would
// silently round-trip as zero values and the Hard CPU would lose its
// ask-history after a session restore.
type goFishMemoryEntryJSON struct {
	AskerIdx  int  `json:"a"`
	TargetIdx int  `json:"g"`
	Rank      int  `json:"r"`
	HadCards  bool `json:"h"`
	TurnSeen  int  `json:"t"`
}

// MarshalJSON implements json.Marshaler.
func (e goFishMemoryEntry) MarshalJSON() ([]byte, error) {
	return json.Marshal(goFishMemoryEntryJSON{
		AskerIdx:  e.askerIdx,
		TargetIdx: e.targetIdx,
		Rank:      e.rank,
		HadCards:  e.hadCards,
		TurnSeen:  e.turnSeen,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (e *goFishMemoryEntry) UnmarshalJSON(data []byte) error {
	var j goFishMemoryEntryJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	e.askerIdx = j.AskerIdx
	e.targetIdx = j.TargetIdx
	e.rank = j.Rank
	e.hadCards = j.HadCards
	e.turnSeen = j.TurnSeen
	return nil
}

// GoFish Go Fishゲームクラス
type GoFish struct {
	trumpCards  *TrumpCards
	players     []*GoFishPlayer
	config      GoFishConfig
	phase       GoFishPhase
	currentTurn int
	gameEndFlag bool
	winnerIdx   int
	turnNumber  int

	// 最後のアクション状態 (プレゼンタ用)
	lastAskPlayerIdx  int
	lastAskTargetIdx  int
	lastAskRank       int
	lastAskSuccess    bool
	lastCardsReceived []*Card
	lastDrawnCard     *Card
	lastBookFormed    bool
	lastBookRank      int

	cpuActions  []*GoFishCpuAction
	humanAction *GoFishCpuAction
	actionLog   []*ActionLogEntry

	// CPU記憶 (Hard難易度)
	cpuMemories []goFishMemoryEntry
}

// NewGoFish コンストラクタ
func NewGoFish(trumpCards *TrumpCards, players []*GoFishPlayer) *GoFish {
	return &GoFish{
		trumpCards:       trumpCards,
		players:          players,
		winnerIdx:        -1,
		lastAskPlayerIdx: -1,
		lastAskTargetIdx: -1,
		cpuActions:       make([]*GoFishCpuAction, 0),
		actionLog:        make([]*ActionLogEntry, 0),
		cpuMemories:      make([]goFishMemoryEntry, 0),
	}
}

// NewDefaultGoFish returns GoFish with the standard 4-player setup (1 human, 3 CPU).
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultGoFish() *GoFish {
	players := []*GoFishPlayer{
		NewGoFishPlayer(true),
		NewGoFishPlayer(false),
		NewGoFishPlayer(false),
		NewGoFishPlayer(false),
	}
	return NewGoFish(NewTrumpCards(0), players)
}

// Reset ゲームを初期化する
func (g *GoFish) Reset() {
	// プレイヤーリセット
	resetPlayers(g.players, func(p *GoFishPlayer) {
		p.ResetBooks()
	})

	// デッキシャッフル
	g.trumpCards = NewTrumpCards(0) // ジョーカーなし
	for range goFishShuffleCount {
		g.trumpCards.Shuffle()
	}

	// 配牌 (各プレイヤーに5枚)
	for range GoFishInitialHandSize {
		for _, p := range g.players {
			card := g.trumpCards.DrawCard()
			if card != nil {
				p.AddCard(card)
			}
		}
	}

	// 配牌後のブックチェック
	for i := range g.players {
		g.checkAndFormBooks(i)
	}

	// 状態初期化
	g.phase = GoFishPhasePlay
	g.currentTurn = 0
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.turnNumber = 1
	g.lastAskPlayerIdx = -1
	g.lastAskTargetIdx = -1
	g.lastAskRank = 0
	g.lastAskSuccess = false
	g.lastCardsReceived = nil
	g.lastDrawnCard = nil
	g.lastBookFormed = false
	g.lastBookRank = 0
	g.cpuActions = make([]*GoFishCpuAction, 0)
	g.humanAction = nil
	g.actionLog = make([]*ActionLogEntry, 0)
	g.cpuMemories = make([]goFishMemoryEntry, 0)

	g.checkGameEnd()
}

// SetConfig ゲーム設定をセットする
func (g *GoFish) SetConfig(config GoFishConfig) { g.config = config }

// GetConfig ゲーム設定を取得する
func (g *GoFish) GetConfig() GoFishConfig { return g.config }

// PlayerAsk 人間プレイヤーが相手にランクを要求する
func (g *GoFish) PlayerAsk(targetIdx, rank int) error {
	if g.gameEndFlag {
		return ErrGoFishGameEnded
	}
	if !g.IsHumanTurn() {
		return ErrGoFishNotYourTurn
	}
	if targetIdx == g.currentTurn {
		return ErrGoFishAskSelf
	}
	if targetIdx < 0 || targetIdx >= len(g.players) {
		return ErrGoFishInvalidTarget
	}
	if g.players[targetIdx].GetCardsSize() == 0 {
		return ErrGoFishInvalidTarget
	}
	asker := g.players[g.currentTurn]
	if !asker.HasRank(rank) {
		return ErrGoFishInvalidRank
	}

	g.cpuActions = make([]*GoFishCpuAction, 0)
	action := g.executeAsk(g.currentTurn, targetIdx, rank)
	g.humanAction = action
	return nil
}

// CpuAsk CPUプレイヤーが1回要求を実行する
func (g *GoFish) CpuAsk() error {
	if g.gameEndFlag {
		return ErrGoFishGameEnded
	}
	if g.IsHumanTurn() {
		return ErrGoFishNotYourTurn
	}
	cpuIdx := g.currentTurn
	cpu := g.players[cpuIdx]

	// 手札がなければ山札から引く
	if cpu.GetCardsSize() == 0 {
		card := g.trumpCards.DrawCard()
		if card != nil {
			cpu.AddCard(card)
		} else {
			g.advanceTurn()
			return nil
		}
	}

	targetIdx, rank := g.cpuSelectAsk(cpuIdx)
	if targetIdx < 0 {
		// 有効な相手がいない場合はターンを進める
		g.advanceTurn()
		return nil
	}

	action := g.executeAsk(cpuIdx, targetIdx, rank)
	g.cpuActions = append(g.cpuActions, action)
	return nil
}

// executeAsk 要求の実行 (人間/CPU共通)
func (g *GoFish) executeAsk(askerIdx, targetIdx, rank int) *GoFishCpuAction {
	asker := g.players[askerIdx]
	target := g.players[targetIdx]

	action := &GoFishCpuAction{
		AskPlayerIdx: askerIdx,
		AskTargetIdx: targetIdx,
		AskRank:      rank,
	}

	// 最後のアクション状態を更新
	g.lastAskPlayerIdx = askerIdx
	g.lastAskTargetIdx = targetIdx
	g.lastAskRank = rank
	g.lastBookFormed = false
	g.lastBookRank = 0
	g.lastDrawnCard = nil

	// CPU記憶に追加
	hadCards := target.HasRank(rank)
	g.cpuMemories = append(g.cpuMemories, goFishMemoryEntry{
		askerIdx:  askerIdx,
		targetIdx: targetIdx,
		rank:      rank,
		hadCards:  hadCards,
		turnSeen:  g.turnNumber,
	})

	if hadCards {
		// 相手がカードを持っている → 全て受け取る
		cards := target.RemoveAllOfRank(rank)
		for _, c := range cards {
			asker.AddCard(c)
		}
		action.Success = true
		action.CardsReceived = len(cards)
		g.lastAskSuccess = true
		g.lastCardsReceived = cards

		// 棋譜
		g.actionLog = append(g.actionLog, &ActionLogEntry{
			TurnNumber: g.turnNumber,
			PlayerIdx:  askerIdx,
			ActionType: "ask_hit",
			Detail:     fmt.Sprintf("P%d asked P%d for rank %d → got %d card(s)", askerIdx, targetIdx, rank, len(cards)),
			Cards:      cards,
		})
	} else {
		// Go Fish! 山札から1枚引く
		action.Success = false
		action.CardsReceived = 0
		g.lastAskSuccess = false
		g.lastCardsReceived = nil

		drawnCard := g.trumpCards.DrawCard()
		if drawnCard != nil {
			asker.AddCard(drawnCard)
			action.DrawnCard = drawnCard
			g.lastDrawnCard = drawnCard
		}

		// 棋譜
		g.actionLog = append(g.actionLog, &ActionLogEntry{
			TurnNumber: g.turnNumber,
			PlayerIdx:  askerIdx,
			ActionType: "ask_miss",
			Detail:     fmt.Sprintf("P%d asked P%d for rank %d → Go Fish!", askerIdx, targetIdx, rank),
		})
	}

	// ブックチェック
	bookFormed, bookRank := g.checkAndFormBooks(askerIdx)
	if bookFormed {
		action.BookFormed = true
		action.BookRank = bookRank
		g.lastBookFormed = true
		g.lastBookRank = bookRank
	}

	// 手札がなくなった場合、山札から補充
	if asker.GetCardsSize() == 0 && g.trumpCards.GetRemainingCount() > 0 {
		card := g.trumpCards.DrawCard()
		if card != nil {
			asker.AddCard(card)
		}
	}

	// ターン管理
	if hadCards {
		// 成功 → もう一度ターン (手札があれば)
		if asker.GetCardsSize() == 0 {
			g.advanceTurn()
		}
		// 手札がある場合は currentTurn を変えない
	} else if action.DrawnCard != nil && action.DrawnCard.GetValue() == rank {
		// Go Fishで引いたカードが要求ランクと一致 → もう一度ターン
		if asker.GetCardsSize() == 0 {
			g.advanceTurn()
		}
	} else {
		// ターン終了
		g.advanceTurn()
	}

	g.checkGameEnd()
	return action
}

// checkAndFormBooks 手札からブックを検出して形成する。ブックが形成された場合 true とランクを返す。
func (g *GoFish) checkAndFormBooks(playerIdx int) (bool, int) {
	p := g.players[playerIdx]
	formed := false
	bookRank := 0
	for {
		found := false
		for _, rank := range p.GetDistinctRanks() {
			if p.CountRank(rank) >= GoFishBookSize {
				cards := p.RemoveAllOfRank(rank)
				p.AddBook(cards)
				formed = true
				bookRank = rank
				found = true

				g.actionLog = append(g.actionLog, &ActionLogEntry{
					TurnNumber: g.turnNumber,
					PlayerIdx:  playerIdx,
					ActionType: "book",
					Detail:     fmt.Sprintf("P%d completed book of rank %d", playerIdx, rank),
					Cards:      cards,
				})
				break
			}
		}
		if !found {
			break
		}
	}
	return formed, bookRank
}

// advanceTurn ターンを進める
func (g *GoFish) advanceTurn() {
	g.turnNumber++
	n := len(g.players)
	for i := 1; i <= n; i++ {
		next := (g.currentTurn + i) % n
		p := g.players[next]
		// 手札がある、もしくは山札から引ける
		if p.GetCardsSize() > 0 || g.trumpCards.GetRemainingCount() > 0 {
			g.currentTurn = next
			return
		}
	}
	// 全員手札なし & 山札なし → ゲーム終了
	g.currentTurn = (g.currentTurn + 1) % n
}

// checkGameEnd ゲーム終了判定
func (g *GoFish) checkGameEnd() {
	// 全13ブック完成
	totalBooks := 0
	for _, p := range g.players {
		totalBooks += p.GetBookCount()
	}
	if totalBooks >= GoFishTotalBooks {
		g.endGame()
		return
	}

	// 全員手札なし & 山札なし
	allEmpty := true
	for _, p := range g.players {
		if p.GetCardsSize() > 0 {
			allEmpty = false
			break
		}
	}
	if allEmpty && g.trumpCards.GetRemainingCount() == 0 {
		g.endGame()
		return
	}
}

// endGame ゲーム終了処理
func (g *GoFish) endGame() {
	g.gameEndFlag = true
	g.phase = GoFishPhaseGameEnd

	// 勝者判定 (最多ブック)
	maxBooks := -1
	g.winnerIdx = 0
	for i, p := range g.players {
		if p.GetBookCount() > maxBooks {
			maxBooks = p.GetBookCount()
			g.winnerIdx = i
		}
	}
}

// cpuSelectAsk CPUが要求先と要求ランクを選択する
func (g *GoFish) cpuSelectAsk(cpuIdx int) (targetIdx, rank int) {
	cpu := g.players[cpuIdx]
	ranks := cpu.GetDistinctRanks()
	if len(ranks) == 0 {
		return -1, 0
	}

	// 有効な相手 (手札がある他プレイヤー)
	opponents := make([]int, 0)
	for i, p := range g.players {
		if i != cpuIdx && p.GetCardsSize() > 0 {
			opponents = append(opponents, i)
		}
	}
	if len(opponents) == 0 {
		return -1, 0
	}

	switch g.config.CpuDifficulty {
	case GoFishCpuDifficultyEasy:
		return g.cpuAskEasy(opponents, ranks)
	case GoFishCpuDifficultyHard:
		return g.cpuAskHard(cpuIdx, opponents, ranks)
	default:
		return g.cpuAskNormal(cpuIdx, opponents, ranks)
	}
}

// cpuAskEasy ランダムに相手とランクを選択
func (g *GoFish) cpuAskEasy(opponents []int, ranks []int) (int, int) {
	return opponents[rand.Intn(len(opponents))], ranks[rand.Intn(len(ranks))]
}

// cpuAskNormal 手札ベース + 直近の記憶で選択
func (g *GoFish) cpuAskNormal(cpuIdx int, opponents []int, ranks []int) (int, int) {
	// 直近の記憶から、相手が特定のランクを持っている可能性が高いペアを探す
	const memoryWindow = 5
	for i := len(g.cpuMemories) - 1; i >= 0 && i >= len(g.cpuMemories)-memoryWindow; i-- {
		mem := g.cpuMemories[i]
		if mem.askerIdx == cpuIdx {
			continue // 自分が聞いた記憶はスキップ
		}
		// 相手がこのランクを要求した = 持っている可能性がある
		for _, oppIdx := range opponents {
			if mem.askerIdx == oppIdx && g.players[cpuIdx].HasRank(mem.rank) {
				return oppIdx, mem.rank
			}
		}
	}

	// 記憶にヒットしなければ、最も枚数が多いランクをランダムな相手に要求
	cpu := g.players[cpuIdx]
	bestRank := ranks[0]
	bestCount := 0
	for _, r := range ranks {
		cnt := cpu.CountRank(r)
		if cnt > bestCount {
			bestCount = cnt
			bestRank = r
		}
	}
	return opponents[rand.Intn(len(opponents))], bestRank
}

// cpuAskHard 全記憶を使い最適な相手とランクを推測
func (g *GoFish) cpuAskHard(cpuIdx int, opponents []int, ranks []int) (int, int) {
	// 各相手×各ランクのスコアを計算
	type candidate struct {
		targetIdx int
		rank      int
		score     float64
	}
	candidates := make([]candidate, 0)

	for _, oppIdx := range opponents {
		for _, rank := range ranks {
			score := g.cpuCalcHardScore(cpuIdx, oppIdx, rank)
			candidates = append(candidates, candidate{oppIdx, rank, score})
		}
	}

	if len(candidates) == 0 {
		return g.cpuAskEasy(opponents, ranks)
	}

	// 最高スコアの候補を選択
	best := candidates[0]
	for _, c := range candidates[1:] {
		if c.score > best.score {
			best = c
		}
	}

	// 同スコアの候補がある場合はランダム
	tied := make([]candidate, 0)
	for _, c := range candidates {
		if c.score == best.score {
			tied = append(tied, c)
		}
	}
	chosen := tied[rand.Intn(len(tied))]
	return chosen.targetIdx, chosen.rank
}

// cpuCalcHardScore 高難易度CPU: 相手がランクを持っている確率スコアを計算
func (g *GoFish) cpuCalcHardScore(cpuIdx, oppIdx, rank int) float64 {
	score := 0.0

	for _, mem := range g.cpuMemories {
		if mem.rank != rank {
			continue
		}
		// 相手がこのランクを要求した = 持っていた
		if mem.askerIdx == oppIdx {
			score += 2.0
		}
		// 誰かがこの相手にこのランクを要求して持っていた
		if mem.targetIdx == oppIdx && mem.hadCards {
			score += 3.0
		}
		// 誰かがこの相手にこのランクを要求して持っていなかった
		if mem.targetIdx == oppIdx && !mem.hadCards {
			score -= 2.0
		}
		// このランクのブックが他で完成していたら0
	}

	// 自分がこのランクを多く持っているほどブック完成に近い
	cpu := g.players[cpuIdx]
	score += float64(cpu.CountRank(rank)) * 0.5

	return score
}

// IsHumanTurn 現在の手番が人間かを返す
func (g *GoFish) IsHumanTurn() bool {
	return g.players[g.currentTurn].GetIsHuman()
}

// GetGameEndFlag ゲーム終了フラグを取得する
func (g *GoFish) GetGameEndFlag() bool { return g.gameEndFlag }

// GetPhase ゲームフェーズを取得する
func (g *GoFish) GetPhase() GoFishPhase { return g.phase }

// GetCurrentTurn 現在の手番プレイヤーインデックスを取得する
func (g *GoFish) GetCurrentTurn() int { return g.currentTurn }

// GetWinnerIdx 勝者プレイヤーインデックスを取得する
func (g *GoFish) GetWinnerIdx() int { return g.winnerIdx }

// GetTurnNumber 現在のターン番号を取得する
func (g *GoFish) GetTurnNumber() int { return g.turnNumber }

// GetPlayerCnt プレイヤー数を取得する
func (g *GoFish) GetPlayerCnt() int { return len(g.players) }

// GetPlayer 指定インデックスのプレイヤーを取得する
func (g *GoFish) GetPlayer(i int) *GoFishPlayer {
	return getPlayer(g.players, i)
}

// GetDeckRemaining 山札の残り枚数を取得する
func (g *GoFish) GetDeckRemaining() int { return g.trumpCards.GetRemainingCount() }

// GetLastAskPlayerIdx 最後に要求したプレイヤーのインデックスを取得する
func (g *GoFish) GetLastAskPlayerIdx() int { return g.lastAskPlayerIdx }

// GetLastAskTargetIdx 最後に要求された相手のインデックスを取得する
func (g *GoFish) GetLastAskTargetIdx() int { return g.lastAskTargetIdx }

// GetLastAskRank 最後に要求されたランクを取得する
func (g *GoFish) GetLastAskRank() int { return g.lastAskRank }

// GetLastAskSuccess 最後の要求が成功したかを返す
func (g *GoFish) GetLastAskSuccess() bool { return g.lastAskSuccess }

// GetLastCardsReceived 最後に受け取ったカードを取得する
func (g *GoFish) GetLastCardsReceived() []*Card { return g.lastCardsReceived }

// GetLastDrawnCard 最後にGo Fishで引いたカードを取得する
func (g *GoFish) GetLastDrawnCard() *Card { return g.lastDrawnCard }

// GetLastBookFormed 最後のアクションでブックが完成したかを返す
func (g *GoFish) GetLastBookFormed() bool { return g.lastBookFormed }

// GetLastBookRank 最後に完成したブックのランクを取得する
func (g *GoFish) GetLastBookRank() int { return g.lastBookRank }

// GetCpuActions CPUターンの行動履歴を取得する
func (g *GoFish) GetCpuActions() []*GoFishCpuAction { return g.cpuActions }

// GetKnownRanks returns, per player index, the sorted ranks that player is
// publicly known to hold from past asks — a player must hold a card of the rank
// they ask for — excluding ranks they have since booked. Derived from the
// shared ask memory so the CUI hint matches what the CPU can already deduce.
func (g *GoFish) GetKnownRanks() map[int][]int {
	asked := make(map[int]map[int]bool, len(g.players))
	for _, e := range g.cpuMemories {
		if asked[e.askerIdx] == nil {
			asked[e.askerIdx] = make(map[int]bool)
		}
		asked[e.askerIdx][e.rank] = true
	}
	out := make(map[int][]int, len(g.players))
	for i, p := range g.players {
		if p == nil {
			out[i] = nil
			continue
		}
		booked := make(map[int]bool)
		for _, book := range p.GetBooks() {
			if len(book) > 0 {
				booked[book[0].GetValue()] = true
			}
		}
		ranks := make([]int, 0, len(asked[i]))
		for r := range asked[i] {
			if !booked[r] {
				ranks = append(ranks, r)
			}
		}
		sort.Ints(ranks)
		out[i] = ranks
	}
	return out
}

// GetHumanAction 人間の最後の行動記録を取得する
func (g *GoFish) GetHumanAction() *GoFishCpuAction { return g.humanAction }

// GetActionLog 棋譜を取得する
func (g *GoFish) GetActionLog() []*ActionLogEntry { return g.actionLog }

// goFishMaxSliceLen caps slice sizes during deserialisation, matching
// ADR-0028's defensive policy for KV-restored session blobs.
const goFishMaxSliceLen = 1000

// goFishJSON is the JSON wire format for GoFish.
type goFishJSON struct {
	TrumpCards        *TrumpCards         `json:"tc"`
	Players           []*GoFishPlayer     `json:"ps"`
	Config            GoFishConfig        `json:"cf"`
	Phase             GoFishPhase         `json:"ph"`
	CurrentTurn       int                 `json:"ct"`
	GameEndFlag       bool                `json:"ge"`
	WinnerIdx         int                 `json:"wi"`
	TurnNumber        int                 `json:"tn"`
	LastAskPlayerIdx  int                 `json:"lap"`
	LastAskTargetIdx  int                 `json:"lat"`
	LastAskRank       int                 `json:"lar"`
	LastAskSuccess    bool                `json:"las"`
	LastCardsReceived []*Card             `json:"lcr"`
	LastDrawnCard     *Card               `json:"ldc"`
	LastBookFormed    bool                `json:"lbf"`
	LastBookRank      int                 `json:"lbr"`
	CpuActions        []*GoFishCpuAction  `json:"ca"`
	HumanAction       *GoFishCpuAction    `json:"ha"`
	ActionLog         []*ActionLogEntry   `json:"al"`
	CpuMemories       []goFishMemoryEntry `json:"cm"`
}

// MarshalJSON implements json.Marshaler.
func (g *GoFish) MarshalJSON() ([]byte, error) {
	return json.Marshal(goFishJSON{
		TrumpCards:        g.trumpCards,
		Players:           g.players,
		Config:            g.config,
		Phase:             g.phase,
		CurrentTurn:       g.currentTurn,
		GameEndFlag:       g.gameEndFlag,
		WinnerIdx:         g.winnerIdx,
		TurnNumber:        g.turnNumber,
		LastAskPlayerIdx:  g.lastAskPlayerIdx,
		LastAskTargetIdx:  g.lastAskTargetIdx,
		LastAskRank:       g.lastAskRank,
		LastAskSuccess:    g.lastAskSuccess,
		LastCardsReceived: g.lastCardsReceived,
		LastDrawnCard:     g.lastDrawnCard,
		LastBookFormed:    g.lastBookFormed,
		LastBookRank:      g.lastBookRank,
		CpuActions:        g.cpuActions,
		HumanAction:       g.humanAction,
		ActionLog:         g.actionLog,
		CpuMemories:       g.cpuMemories,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *GoFish) UnmarshalJSON(data []byte) error {
	var j goFishJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > goFishMaxSliceLen ||
		len(j.LastCardsReceived) > goFishMaxSliceLen ||
		len(j.CpuActions) > goFishMaxSliceLen ||
		len(j.ActionLog) > goFishMaxSliceLen ||
		len(j.CpuMemories) > goFishMaxSliceLen {
		return fmt.Errorf("goFish: input array exceeds maximum allowed size")
	}
	g.trumpCards = j.TrumpCards
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.currentTurn = j.CurrentTurn
	g.gameEndFlag = j.GameEndFlag
	g.winnerIdx = j.WinnerIdx
	g.turnNumber = j.TurnNumber
	g.lastAskPlayerIdx = j.LastAskPlayerIdx
	g.lastAskTargetIdx = j.LastAskTargetIdx
	g.lastAskRank = j.LastAskRank
	g.lastAskSuccess = j.LastAskSuccess
	g.lastCardsReceived = j.LastCardsReceived
	g.lastDrawnCard = j.LastDrawnCard
	g.lastBookFormed = j.LastBookFormed
	g.lastBookRank = j.LastBookRank
	g.cpuActions = j.CpuActions
	g.humanAction = j.HumanAction
	g.actionLog = j.ActionLog
	g.cpuMemories = j.CpuMemories
	if g.cpuActions == nil {
		g.cpuActions = make([]*GoFishCpuAction, 0)
	}
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	if g.cpuMemories == nil {
		g.cpuMemories = make([]goFishMemoryEntry, 0)
	}
	return nil
}
