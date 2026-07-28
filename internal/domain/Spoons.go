//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// SpoonsPhase はゲームフェーズ。
type SpoonsPhase int

// スプーンのフェーズ定数
const (
	// SpoonsPhasePass 通常のカード回し (パス) 進行中
	SpoonsPhasePass SpoonsPhase = 0
	// SpoonsPhaseGrab 誰かがフォーオブアカインドを揃えてスプーン争奪が始まった状態
	SpoonsPhaseGrab SpoonsPhase = 1
	// SpoonsPhaseRoundEnd ラウンド終了 (スプーンが出尽くし、負けた者に文字が付いた)
	SpoonsPhaseRoundEnd SpoonsPhase = 2
	// SpoonsPhaseGameEnd ゲーム終了 (生存者が 1 人になった)
	SpoonsPhaseGameEnd SpoonsPhase = 3
)

// SpoonsPhaseMin / SpoonsPhaseMax はフェーズ列挙の範囲 (復元時の検証用)。
const (
	SpoonsPhaseMin = SpoonsPhasePass
	SpoonsPhaseMax = SpoonsPhaseGameEnd
)

// Spoons はスプーン (Spoons / スプーン) のゲームクラス。
//
// アメリカのパーティ系スピードゲーム。4 人 (人間 1 + CPU 3) で 52 枚デッキを
// 用い、各プレイヤーは 4 枚を持つ。テーブル中央には人数 - 1 本のスプーンがある。
//
// # パス回しのモデル化 (リアルタイム→ターン状態)
//
// 本来は全員同時にカードを回すが、サーバ側ではターン状態として決定的に
// モデル化する。回しの「フィーダー」(feederIdx, 配り手) がドローパイルから
// 新しいカードを 1 枚ずつ供給し、各プレイヤーは順番 (時計回り) に、流れてくる
// カード (passedCard) を受け取って 5 枚にし、不要な 1 枚を次のプレイヤーへ渡す。
// PlayerPass(cardIndex) で人間が渡す札を選び、CPU は CpuPlay で自動的に渡す。
//
// # フォーオブアカインド → グラブ
//
// パスの結果、手札 4 枚が同じランクになったプレイヤーは即座にスプーンを 1 本
// 掴む (grab)。最初の 1 本が掴まれた瞬間にグラブウィンドウ (SpoonsPhaseGrab) が
// 開き、残り全員が残りのスプーンを奪い合う。人間は PlayerGrabSpoon、CPU は
// CpuPlay で難易度に応じて掴む/掴み損ねる。スプーンが尽きると、取れなかった
// プレイヤーが文字を 1 つ受け取る。
//
// # 文字と脱落
//
// 各プレイヤーは S-P-O-O-N-S の 6 文字を溜めると脱落する。脱落すると以降の
// 配札・パス・グラブから除外され、スプーン本数は (生存者数 - 1) に減る。
// 生存者が 1 人になるとそのプレイヤーが勝者となる。
//
// # 停止保証
//
// フルCPU対戦でも必ず終了する: 毎ラウンド最低 1 文字が付与されるため、文字は
// 単調増加し、いずれ脱落者が出て生存者は減少する。加えてラウンド数上限
// (SpoonsMaxRounds) のガードを設けている。
type Spoons struct {
	trumpCards       *TrumpCards
	players          [SpoonsPlayerCnt]*SpoonsPlayer
	config           SpoonsConfig
	phase            SpoonsPhase
	drawPile         []*Card
	passedCard       *Card // 次のプレイヤーへ流れている札 (nil = フィーダーがドロー)
	feederIdx        int   // 新しい札をドローパイルから供給する配り手
	currentPlayerIdx int   // 現在パスする番のプレイヤー
	spoonsRemaining  int   // テーブル上に残るスプーン本数
	grabWindowOpen   bool  // グラブウィンドウが開いているか
	firstGrabberIdx  int   // 最初にスプーンを掴んだプレイヤー (-1=未)
	roundLoserIdx    int   // 直近ラウンドで文字が付いたプレイヤー (-1=未)
	passCount        int   // 当該ラウンドのパス回数 (停止保証のフェイルセーフ用)
	roundNumber      int
	gameEndFlag      bool
	winnerIdx        int
	actionLog        []*ActionLogEntry

	// rng CPU グラブ抽選用 (テストで差し替え可能)
	rng *rand.Rand
}

// NewSpoons はコンストラクタ。
func NewSpoons(trumpCards *TrumpCards, players []*SpoonsPlayer, config SpoonsConfig) *Spoons {
	g := &Spoons{
		trumpCards:      trumpCards,
		config:          config,
		winnerIdx:       -1,
		firstGrabberIdx: -1,
		roundLoserIdx:   -1,
		rng:             rand.New(rand.NewSource(rand.Int63())),
	}
	for i := 0; i < SpoonsPlayerCnt && i < len(players); i++ {
		g.players[i] = players[i]
	}
	return g
}

// NewDefaultSpoons は標準セットアップ (人間 1 + CPU 3) の Spoons を返す。
func NewDefaultSpoons() *Spoons {
	players := []*SpoonsPlayer{
		NewSpoonsPlayer(true),
		NewSpoonsPlayer(false),
		NewSpoonsPlayer(false),
		NewSpoonsPlayer(false),
	}
	return NewSpoons(NewTrumpCards(0), players, DefaultSpoonsConfig())
}

// SetRand はテスト用に乱数源を差し替える。
func (g *Spoons) SetRand(r *rand.Rand) {
	if r != nil {
		g.rng = r
	}
}

// Reset はゲーム全体を初期化して新しいゲームを開始する。
func (g *Spoons) Reset() {
	for _, p := range g.players {
		if p == nil {
			continue
		}
		p.SetLetters(0)
		p.SetEliminated(false)
	}
	g.roundNumber = 0
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.actionLog = nil
	g.dealRound()
}

// GetConfig は設定を返す。
func (g *Spoons) GetConfig() SpoonsConfig { return g.config }

// SetConfig は設定を更新する。
func (g *Spoons) SetConfig(cfg SpoonsConfig) { g.config = cfg }

// ResetWithConfig は設定を更新してゲームを初期化する。
func (g *Spoons) ResetWithConfig(cfg SpoonsConfig) {
	g.config = cfg
	g.Reset()
}

// NextRound は次のラウンドを開始する。ゲーム終了済みなら何もしない。
func (g *Spoons) NextRound() {
	if g.gameEndFlag {
		return
	}
	g.dealRound()
}

// dealRound は生存者へ手札を配り、新しいラウンドのパスフェーズを開始する。
func (g *Spoons) dealRound() {
	g.roundNumber++
	g.phase = SpoonsPhasePass
	g.passedCard = nil
	g.grabWindowOpen = false
	g.firstGrabberIdx = -1
	g.roundLoserIdx = -1
	g.passCount = 0

	active := g.activeIndices()
	g.spoonsRemaining = len(active) - 1
	if g.spoonsRemaining < 0 {
		g.spoonsRemaining = 0
	}

	for _, p := range g.players {
		if p == nil {
			continue
		}
		p.Reset()
		p.SetHasSpoon(false)
	}

	g.trumpCards.Shuffle()
	g.drawPile = g.drawAll()

	// 生存者へ SpoonsHandSize 枚ずつ配る。
	for _, idx := range active {
		if g.players[idx] == nil {
			continue
		}
		for n := 0; n < SpoonsHandSize; n++ {
			c := g.popDraw()
			if c != nil {
				g.players[idx].AddCard(c)
			}
		}
	}

	// フィーダー (配り手) と最初の手番は最初の生存者。
	if len(active) > 0 {
		g.feederIdx = active[0]
		g.currentPlayerIdx = active[0]
	}
	g.appendLog(-1, "deal", fmt.Sprintf("round %d dealt, %d spoons", g.roundNumber, g.spoonsRemaining), nil)
}

// drawAll はトランプから全カードを取り出してスライスで返す。
func (g *Spoons) drawAll() []*Card {
	cards := make([]*Card, 0, CardCnt)
	for {
		c := g.trumpCards.DrawCard()
		if c == nil {
			break
		}
		cards = append(cards, c)
	}
	return cards
}

// popDraw はドローパイルの先頭 1 枚を取り出す (空なら nil)。
func (g *Spoons) popDraw() *Card {
	if len(g.drawPile) == 0 {
		return nil
	}
	c := g.drawPile[0]
	g.drawPile = g.drawPile[1:]
	return c
}

// --- ゲッター ---

// GetPhase は現在のフェーズを返す。
func (g *Spoons) GetPhase() SpoonsPhase { return g.phase }

// GetGameEndFlag はゲーム終了フラグを返す。
func (g *Spoons) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx は勝者インデックスを返す (-1=未確定)。
func (g *Spoons) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt はプレイヤー数を返す。
func (g *Spoons) GetPlayerCnt() int { return SpoonsPlayerCnt }

// GetPlayer は指定インデックスのプレイヤーを返す (範囲外は nil)。
func (g *Spoons) GetPlayer(i int) *SpoonsPlayer {
	if i < 0 || i >= SpoonsPlayerCnt {
		return nil
	}
	return g.players[i]
}

// GetSpoonsRemaining はテーブル上に残るスプーン本数を返す。
func (g *Spoons) GetSpoonsRemaining() int { return g.spoonsRemaining }

// GetCurrentPlayerIdx は現在パスする番のプレイヤーを返す。
func (g *Spoons) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// GetFeederIdx は配り手プレイヤーを返す。
func (g *Spoons) GetFeederIdx() int { return g.feederIdx }

// GetDrawPileSize はドローパイルの残枚数を返す。
func (g *Spoons) GetDrawPileSize() int { return len(g.drawPile) }

// GetPassedCard は現在流れているカードを返す (nil=フィーダーがドロー予定)。
func (g *Spoons) GetPassedCard() *Card { return g.passedCard }

// IsGrabWindowOpen はグラブウィンドウが開いているかを返す。
func (g *Spoons) IsGrabWindowOpen() bool { return g.grabWindowOpen }

// GetFirstGrabberIdx は最初にスプーンを掴んだプレイヤーを返す (-1=未)。
func (g *Spoons) GetFirstGrabberIdx() int { return g.firstGrabberIdx }

// GetRoundLoserIdx は直近ラウンドで文字が付いたプレイヤーを返す (-1=未)。
func (g *Spoons) GetRoundLoserIdx() int { return g.roundLoserIdx }

// GetRoundNumber は現在のラウンド番号を返す。
func (g *Spoons) GetRoundNumber() int { return g.roundNumber }

// GetActionLog は棋譜を返す。
func (g *Spoons) GetActionLog() []*ActionLogEntry { return g.actionLog }

// IsHumanTurn は現在パス/グラブの手番が人間かを返す。
// パスフェーズでは currentPlayerIdx、グラブフェーズでは人間が未グラブなら true。
func (g *Spoons) IsHumanTurn() bool {
	if g.gameEndFlag {
		return false
	}
	human := g.GetPlayer(0)
	if human == nil || !human.GetIsHuman() || human.GetEliminated() {
		return false
	}
	switch g.phase {
	case SpoonsPhasePass:
		return g.currentPlayerIdx == 0
	case SpoonsPhaseGrab:
		return !human.GetHasSpoon() && g.spoonsRemaining > 0
	default:
		return false
	}
}

// --- アクション ---

// PlayerPass は人間プレイヤーが手札の 1 枚を次のプレイヤーへ渡す。
//
// cardIndex は受け取り後 (= 5 枚状態) の手札インデックス。パスフェーズで人間の
// 手番でないときや不正なインデックスのときは ErrInvalidPlay を返す。
func (g *Spoons) PlayerPass(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SpoonsPhasePass {
		return ErrWrongPhase
	}
	if g.currentPlayerIdx != 0 {
		return ErrInvalidPlay
	}
	g.receiveIncoming(0)
	p := g.players[0]
	if p == nil {
		return ErrInvalidPlay
	}
	if cardIndex < 0 || cardIndex >= p.GetCardsSize() {
		return ErrInvalidCard
	}
	g.doPass(0, cardIndex)
	return nil
}

// PlayerGrabSpoon は人間プレイヤーがスプーンを掴む。
//
// グラブウィンドウが開いていないときや既に掴んでいるときは ErrInvalidPlay。
func (g *Spoons) PlayerGrabSpoon() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SpoonsPhaseGrab {
		return ErrWrongPhase
	}
	human := g.players[0]
	if human == nil {
		return ErrInvalidPlay
	}
	if human.GetEliminated() || human.GetHasSpoon() || g.spoonsRemaining <= 0 {
		return ErrInvalidPlay
	}
	g.grabSpoon(0)
	g.maybeCloseGrabWindow()
	return nil
}

// CpuPlay は CPU の手番を 1 ステップ進める。
//
// パスフェーズでは現在の CPU 手番がパスを行い、グラブフェーズでは未グラブの
// CPU が難易度に応じてスプーンを掴む (または掴み損ねる) 試行を 1 体ずつ行う。
// 人間の操作が必要な状態 (人間の手番) では何もしない。
func (g *Spoons) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	switch g.phase {
	case SpoonsPhasePass:
		g.cpuPass()
	case SpoonsPhaseGrab:
		g.cpuGrab()
	}
}

// cpuPass は現在の CPU 手番のパスを実行する。人間の手番なら何もしない。
func (g *Spoons) cpuPass() {
	idx := g.currentPlayerIdx
	p := g.GetPlayer(idx)
	if p == nil || p.GetEliminated() || p.GetIsHuman() {
		return
	}
	g.receiveIncoming(idx)
	g.doPass(idx, g.cpuChoosePass(idx))
}

// cpuChoosePass は CPU が渡す札のインデックスを選ぶ。
//
// 最も枚数の少ないランク (= 揃いにくいランク) のカードを 1 枚捨てる単純戦略。
func (g *Spoons) cpuChoosePass(idx int) int {
	p := g.players[idx]
	counts := map[int]int{}
	for i := 0; i < p.GetCardsSize(); i++ {
		if c := p.GetCard(i); c != nil {
			counts[c.GetValue()]++
		}
	}
	worst := 0
	worstCount := SpoonsMaxLetters + p.GetCardsSize() + 1
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c == nil {
			continue
		}
		if counts[c.GetValue()] < worstCount {
			worstCount = counts[c.GetValue()]
			worst = i
		}
	}
	return worst
}

// cpuGrab はグラブウィンドウ中の CPU 取得を進める。
//
// まず取りこぼしを許容した 1 周、続いて確実に掴む 1 周を行うことで、人間が
// グラブを選ばない (またはできない) 場合でもラウンドが必ず前進する。
func (g *Spoons) cpuGrab() {
	g.cpuRace(true)  // 取りこぼしあり (人間に猶予)
	g.cpuRace(false) // 確実に前進
	g.maybeCloseGrabWindow()
}

// receiveIncoming はプレイヤー idx に流れてくる札を渡し、手札を 5 枚にする。
// フィーダー自身の番で passedCard が nil の場合はドローパイルから供給する。
func (g *Spoons) receiveIncoming(idx int) {
	p := g.players[idx]
	if g.passedCard == nil {
		// 回しの起点: フィーダーがドローパイルから新しい札を 1 枚引く。
		g.passedCard = g.popDraw()
	}
	if g.passedCard != nil {
		p.AddCard(g.passedCard)
		g.passedCard = nil
	}
}

// doPass はプレイヤー idx が手札の cardIndex を次の生存プレイヤーへ渡し、
// フォーオブアカインド判定とウィンドウ判定を行ったうえで手番を進める。
func (g *Spoons) doPass(idx, cardIndex int) {
	p := g.players[idx]
	passed := p.RemoveCard(cardIndex)
	g.passedCard = passed
	cardArg := []*Card(nil)
	if passed != nil {
		cardArg = []*Card{passed}
	}
	g.appendLog(idx, "pass", "pass a card to the left", cardArg)
	g.passCount++

	// フォーオブアカインド: 揃ったら即グラブしウィンドウを開く。
	if p.HasFourOfAKind() && g.phase == SpoonsPhasePass && g.spoonsRemaining > 0 {
		g.openGrabWindow(idx)
		return
	}

	// 停止保証のフェイルセーフ: パス回数が上限を超えたら、揃わなくても現在の
	// プレイヤーがグラブを開始したものとみなしてウィンドウを開く。
	if g.passCount >= SpoonsMaxPassesPerRound && g.spoonsRemaining > 0 {
		g.openGrabWindow(idx)
		return
	}

	g.advancePass()
}

// advancePass は次の生存プレイヤーへ手番を移す。フィーダーへ戻ると
// passedCard を捨ててドローパイルから新規供給するサイクルになる。
func (g *Spoons) advancePass() {
	next := g.nextActive(g.currentPlayerIdx)
	if next == g.feederIdx {
		// 1 周完了: 流れてきた札はフィーダーが「捨て札」にし、ドローパイルへ戻す
		// (尽きないように)。ドローパイルが空ならその札を再投入する。
		if g.passedCard != nil {
			g.drawPile = append(g.drawPile, g.passedCard)
			g.passedCard = nil
		}
	}
	g.currentPlayerIdx = next
}

// openGrabWindow はプレイヤー idx がフォーオブアカインドを揃えた瞬間に
// スプーンを掴ませ、グラブウィンドウを開く。
//
// 続いて他の CPU が難易度別の取りこぼし確率で「レース」を 1 周行う (人間より
// 先に掴むことがある)。これにより人間が掴むときには競争相手として CPU が既に
// 何本か取っている場合があり、人間も負ける (= 文字を取る) 可能性が生まれる。
// 人間は SpoonsPhaseGrab が続く限り PlayerGrabSpoon で参加できる。
func (g *Spoons) openGrabWindow(idx int) {
	g.phase = SpoonsPhaseGrab
	g.grabWindowOpen = true
	g.firstGrabberIdx = idx
	g.appendLog(idx, "four", "four of a kind! grab a spoon", nil)
	g.grabSpoon(idx)
	g.cpuRace(true)
	g.maybeCloseGrabWindow()
}

// cpuRace は未グラブの CPU 全員にスプーン取得を 1 周試行させる。
// missAllowed が true のとき各 CPU は難易度別の取りこぼし確率で見送ることが
// あり、人間に取得の猶予を残す。false のときは確実に掴み、フェイルセーフとして
// ウィンドウを必ず前進させる。
func (g *Spoons) cpuRace(missAllowed bool) {
	for _, idx := range g.activeIndices() {
		if g.spoonsRemaining <= 0 {
			return
		}
		p := g.players[idx]
		if p.GetIsHuman() || p.GetHasSpoon() {
			continue
		}
		if missAllowed && g.rng.Float64() < g.config.GrabMissChance() {
			continue
		}
		g.grabSpoon(idx)
	}
}

// grabSpoon はプレイヤー idx にスプーンを 1 本掴ませる。
func (g *Spoons) grabSpoon(idx int) {
	if g.spoonsRemaining <= 0 {
		return
	}
	p := g.players[idx]
	if p.GetHasSpoon() {
		return
	}
	p.SetHasSpoon(true)
	g.spoonsRemaining--
	g.appendLog(idx, "grab", "grabbed a spoon", nil)
}

// maybeCloseGrabWindow はスプーンが尽きたらラウンドを締めて文字を付与する。
func (g *Spoons) maybeCloseGrabWindow() {
	if g.phase != SpoonsPhaseGrab || g.spoonsRemaining > 0 {
		return
	}
	g.endRound()
}

// endRound はスプーンを取れなかった生存プレイヤーに文字を付け、脱落・勝敗を
// 判定してラウンドを終了する。
func (g *Spoons) endRound() {
	g.grabWindowOpen = false

	loser := -1
	for _, idx := range g.activeIndices() {
		p := g.players[idx]
		if !p.GetHasSpoon() {
			loser = idx
			eliminated := p.AddLetter()
			detail := fmt.Sprintf("missed spoon, +1 letter (now %d)", p.GetLetters())
			if eliminated {
				detail = "missed spoon, eliminated (SPOONS)"
			}
			g.appendLog(idx, "letter", detail, nil)
		}
	}
	g.roundLoserIdx = loser
	g.phase = SpoonsPhaseRoundEnd

	g.checkGameEnd()
}

// checkGameEnd は生存者が 1 人以下、またはラウンド上限に達したらゲームを締める。
func (g *Spoons) checkGameEnd() {
	active := g.activeIndices()
	if len(active) == 1 {
		g.endGame(active[0])
		return
	}
	if len(active) == 0 {
		g.endGame(-1)
		return
	}
	if g.roundNumber >= SpoonsMaxRounds {
		// 停止保証のフェイルセーフ: 最も文字数の少ない生存者を勝者にする。
		best := active[0]
		for _, idx := range active {
			if g.players[idx].GetLetters() < g.players[best].GetLetters() {
				best = idx
			}
		}
		g.endGame(best)
	}
}

// endGame はゲームを終了し勝者を確定する。
func (g *Spoons) endGame(winner int) {
	g.gameEndFlag = true
	g.phase = SpoonsPhaseGameEnd
	g.winnerIdx = winner
	g.grabWindowOpen = false
	if winner >= 0 && winner < SpoonsPlayerCnt && g.players[winner] != nil {
		g.players[winner].SetIsFinished(true)
	}
	g.appendLog(winner, "gameEnd", "game over", nil)
}

// --- 内部ヘルパ ---

// activeIndices は脱落していないプレイヤーのインデックスを昇順で返す。
func (g *Spoons) activeIndices() []int {
	out := make([]int, 0, SpoonsPlayerCnt)
	for i := 0; i < SpoonsPlayerCnt; i++ {
		if g.players[i] != nil && !g.players[i].GetEliminated() {
			out = append(out, i)
		}
	}
	return out
}

// nextActive は idx の次の生存プレイヤーインデックス (時計回り) を返す。
// 生存者が idx 自身しかいない場合は idx を返す。
func (g *Spoons) nextActive(idx int) int {
	for step := 1; step <= SpoonsPlayerCnt; step++ {
		cand := (idx + step) % SpoonsPlayerCnt
		if g.players[cand] != nil && !g.players[cand].GetEliminated() {
			return cand
		}
	}
	return idx
}

// appendLog は棋譜にエントリを追加する。
func (g *Spoons) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: len(g.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- JSON ---

// spoonsJSON is the JSON wire format for Spoons.
type spoonsJSON struct {
	TrumpCards       *TrumpCards                    `json:"tc"`
	Players          [SpoonsPlayerCnt]*SpoonsPlayer `json:"ps"`
	Config           SpoonsConfig                   `json:"cf"`
	Phase            SpoonsPhase                    `json:"ph"`
	DrawPile         []*Card                        `json:"dp"`
	PassedCard       *Card                          `json:"pc"`
	FeederIdx        int                            `json:"fi"`
	CurrentPlayerIdx int                            `json:"ci"`
	SpoonsRemaining  int                            `json:"sr"`
	GrabWindowOpen   bool                           `json:"gw"`
	FirstGrabberIdx  int                            `json:"fg"`
	RoundLoserIdx    int                            `json:"rl"`
	PassCount        int                            `json:"pn"`
	RoundNumber      int                            `json:"rn"`
	GameEndFlag      bool                           `json:"ge"`
	WinnerIdx        int                            `json:"wi"`
	ActionLog        []*ActionLogEntry              `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Spoons) MarshalJSON() ([]byte, error) {
	return json.Marshal(spoonsJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		DrawPile:         g.drawPile,
		PassedCard:       g.passedCard,
		FeederIdx:        g.feederIdx,
		CurrentPlayerIdx: g.currentPlayerIdx,
		SpoonsRemaining:  g.spoonsRemaining,
		GrabWindowOpen:   g.grabWindowOpen,
		FirstGrabberIdx:  g.firstGrabberIdx,
		RoundLoserIdx:    g.roundLoserIdx,
		PassCount:        g.passCount,
		RoundNumber:      g.roundNumber,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		ActionLog:        g.actionLog,
	})
}

// errSpoonsInvalidState は復元データが不正なときの番兵エラー。
var errSpoonsInvalidState = fmt.Errorf("invalid spoons state")

// UnmarshalJSON implements json.Unmarshaler with defensive validation.
func (g *Spoons) UnmarshalJSON(data []byte) error {
	var j spoonsJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if j.Phase < SpoonsPhaseMin || j.Phase > SpoonsPhaseMax {
		return errSpoonsInvalidState
	}
	for i := 0; i < SpoonsPlayerCnt; i++ {
		if j.Players[i] == nil {
			return errSpoonsInvalidState
		}
	}
	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= SpoonsPlayerCnt {
		return errSpoonsInvalidState
	}
	if j.FeederIdx < 0 || j.FeederIdx >= SpoonsPlayerCnt {
		return errSpoonsInvalidState
	}
	if j.WinnerIdx < -1 || j.WinnerIdx >= SpoonsPlayerCnt {
		return errSpoonsInvalidState
	}
	if j.SpoonsRemaining < 0 || j.SpoonsRemaining > SpoonsPlayerCnt {
		return errSpoonsInvalidState
	}
	if len(j.ActionLog) > SpoonsMaxRounds*SpoonsPlayerCnt*SpoonsPlayerCnt {
		return errSpoonsInvalidState
	}

	g.trumpCards = j.TrumpCards
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.drawPile = j.DrawPile
	g.passedCard = j.PassedCard
	g.feederIdx = j.FeederIdx
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.spoonsRemaining = j.SpoonsRemaining
	g.grabWindowOpen = j.GrabWindowOpen
	g.firstGrabberIdx = j.FirstGrabberIdx
	g.roundLoserIdx = j.RoundLoserIdx
	g.passCount = j.PassCount
	g.roundNumber = j.RoundNumber
	g.gameEndFlag = j.GameEndFlag
	g.winnerIdx = j.WinnerIdx
	g.actionLog = j.ActionLog
	if g.rng == nil {
		g.rng = rand.New(rand.NewSource(rand.Int63()))
	}
	return nil
}
