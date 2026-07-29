package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
)

// NertzPhase Nertz ゲームフェーズ
type NertzPhase int

// Nertz のフェーズ定数
const (
	// NertzPhaseIdle 未開始 (Reset 前)
	NertzPhaseIdle NertzPhase = iota
	// NertzPhasePlaying プレイ中
	NertzPhasePlaying
	// NertzPhaseRoundEnd ラウンド終了 (誰かがナッツパイルを空にした)
	NertzPhaseRoundEnd
	// NertzPhaseGameEnd ゲーム終了 (目標スコア到達)
	NertzPhaseGameEnd
)

// NertzZone カード移動元/先のゾーン種別
type NertzZone int

// Nertz のゾーン定数
const (
	// NertzZoneNertz ナッツパイル
	NertzZoneNertz NertzZone = iota
	// NertzZoneTableau タブロー (リバー)
	NertzZoneTableau
	// NertzZoneWaste ウェイスト
	NertzZoneWaste
	// NertzZoneStock ストック
	NertzZoneStock
	// NertzZoneFoundation ファウンデーション
	NertzZoneFoundation
)

// NertzFoundationsPerPlayer プレイヤー1人あたり用意するファウンデーション枠数
// (1デッキ52枚あたり最大4本のスート別ファウンデーションが完成する)
const NertzFoundationsPerPlayer = 4

// nertzMaxSliceLen JSON 復元時のスライス上限
const nertzMaxSliceLen = 10000

// Nertz 多人数リアルタイム対戦型クロンダイクゲーム本体。
// 注: 当ドメインは純粋に状態のみを保持する。リアルタイムは usecase 層で
// "tick" を駆動して表現する (ADR-0031 を参照)。
type Nertz struct {
	decks       []*TrumpCards
	players     []*NertzPlayer
	foundations []*NertzFoundation
	config      NertzConfig
	phase       NertzPhase
	roundNo     int
	winnerIdx   int // ラウンド勝者 (Nertzコール); -1 = 未確定
	matchWinner int // マッチ勝者; -1 = 未確定
	moveCount   int
	actionLog   []*ActionLogEntry
	history     []*nertzSnapshot // 人間の単一ステップ Undo 用
}

// nertzSnapshot Undo 用スナップショット
type nertzSnapshot struct {
	playersJSON     []byte
	foundationsJSON []byte
	phase           NertzPhase
	moveCount       int
	winnerIdx       int
}

// NewNertz Nertz を作成する。プレイヤー数は cfg.PlayerCount に従う。
func NewNertz(cfg NertzConfig) *Nertz {
	g := &Nertz{
		config:      cfg,
		phase:       NertzPhaseIdle,
		roundNo:     0,
		winnerIdx:   -1,
		matchWinner: -1,
	}
	return g
}

// NewDefaultNertz 既定設定 (4プレイヤー / DrawCount=3 / 目標100点) の Nertz を返す。
func NewDefaultNertz() *Nertz {
	return NewNertz(DefaultNertzConfig())
}

// Reset 現在の設定で新しいマッチを開始する (累積スコアもリセット)。
func (g *Nertz) Reset() {
	if err := g.config.Validate(); err != nil {
		g.config = DefaultNertzConfig()
	}
	g.roundNo = 1
	g.matchWinner = -1
	g.players = make([]*NertzPlayer, g.config.PlayerCount)
	g.decks = make([]*TrumpCards, g.config.PlayerCount)
	for i := 0; i < g.config.PlayerCount; i++ {
		g.players[i] = NewNertzPlayer(g.defaultPlayerName(i), i != 0, i)
		g.decks[i] = NewTrumpCards(0)
	}
	g.startRound()
}

// ResetWithConfig 設定を適用して Reset する。
func (g *Nertz) ResetWithConfig(cfg NertzConfig) {
	if err := cfg.Validate(); err != nil {
		cfg = DefaultNertzConfig()
	}
	g.config = cfg
	g.Reset()
}

// NextRound 次ラウンドを開始する (累積スコアは保持)。
// マッチ勝者が確定済みなら何もしない。
func (g *Nertz) NextRound() {
	if g.matchWinner >= 0 {
		return
	}
	g.roundNo++
	g.startRound()
}

// startRound ラウンド盤面を初期化する (累積スコアは保持)。
func (g *Nertz) startRound() {
	g.phase = NertzPhasePlaying
	g.winnerIdx = -1
	g.moveCount = 0
	g.actionLog = nil
	g.history = nil

	// 各プレイヤーに対応デッキで配る
	for i, p := range g.players {
		p.ResetRoundPiles()
		deck := g.decks[i]
		deck.Shuffle()
		// ナッツパイル: 13枚 (末尾 = トップ)
		for k := 0; k < NertzPileSize; k++ {
			p.PushNertz(deck.DrawCard())
		}
		// タブロー: 4列に1枚ずつ (全て表向き)
		for c := 0; c < NertzTableauCnt; c++ {
			p.PushTableau(c, &NertzTableauCard{Card: deck.DrawCard(), FaceUp: true})
		}
		// 残り 35 枚 = ストック (DrawCard 順を逆にして末尾=Top にする)
		for deck.GetRemainingCount() > 0 {
			p.PushStock(deck.DrawCard())
		}
	}

	// ファウンデーション: プレイヤー数 × 4 枠を空で確保
	g.foundations = make([]*NertzFoundation, NertzFoundationsPerPlayer*g.config.PlayerCount)
	for i := range g.foundations {
		g.foundations[i] = NewNertzFoundation()
	}
}

func (g *Nertz) defaultPlayerName(idx int) string {
	if idx == 0 {
		return "You"
	}
	return fmt.Sprintf("CPU%d", idx)
}

// --- Public actions ---

// DrawStock 指定プレイヤーがストックから DrawCount 枚をウェイストにめくる。
// ストックが空ならウェイストをリサイクルしてストックに戻す。
func (g *Nertz) DrawStock(playerIdx int) error {
	if err := g.assertPlayable(); err != nil {
		return err
	}
	p, err := g.player(playerIdx)
	if err != nil {
		return err
	}
	if p.StockSize() == 0 {
		if p.WasteSize() == 0 {
			return errors.New("nertz: stock and waste are both empty")
		}
		g.takeSnapshot(playerIdx)
		p.RecycleWasteToStock()
		g.appendLog(playerIdx, "recycle", "ウェイストをストックに戻しました", nil)
		return nil
	}
	g.takeSnapshot(playerIdx)
	count := g.config.DrawCount
	drawn := make([]*Card, 0, count)
	for i := 0; i < count && p.StockSize() > 0; i++ {
		c := p.PopStock()
		p.PushWaste(c)
		drawn = append(drawn, c)
	}
	g.moveCount++
	g.appendLog(playerIdx, "draw", "ストックからカードを引きました", drawn)
	return nil
}

// MoveNertzToFoundation ナッツパイルのトップを foundationIdx へ送る。
func (g *Nertz) MoveNertzToFoundation(playerIdx, foundationIdx int) error {
	if err := g.assertPlayable(); err != nil {
		return err
	}
	p, err := g.player(playerIdx)
	if err != nil {
		return err
	}
	f, err := g.foundation(foundationIdx)
	if err != nil {
		return err
	}
	top := p.NertzTop()
	if top == nil {
		return errors.New("nertz: nertz pile is empty")
	}
	if !f.CanAccept(top) {
		return errors.New("nertz: cannot place card on foundation")
	}
	g.takeSnapshot(playerIdx)
	c := p.PopNertz()
	if err := f.Push(c, p.GetDeckIdx()); err != nil {
		return err
	}
	g.moveCount++
	g.appendLog(playerIdx, "moveNF", fmt.Sprintf("プレイヤー%dがナッツ→ファウンデーション%d", playerIdx, foundationIdx), []*Card{c})
	g.checkRoundEndForPlayer(playerIdx)
	return nil
}

// MoveNertzToTableau ナッツパイルのトップをタブロー toCol に置く。
func (g *Nertz) MoveNertzToTableau(playerIdx, toCol int) error {
	if err := g.assertPlayable(); err != nil {
		return err
	}
	p, err := g.player(playerIdx)
	if err != nil {
		return err
	}
	if toCol < 0 || toCol >= NertzTableauCnt {
		return errors.New("nertz: invalid tableau column")
	}
	top := p.NertzTop()
	if top == nil {
		return errors.New("nertz: nertz pile is empty")
	}
	if !canPlaceOnNertzTableau(top, p.TableauTop(toCol)) {
		return errors.New("nertz: cannot place card on tableau")
	}
	g.takeSnapshot(playerIdx)
	c := p.PopNertz()
	p.PushTableau(toCol, &NertzTableauCard{Card: c, FaceUp: true})
	g.moveCount++
	g.appendLog(playerIdx, "moveNT", fmt.Sprintf("プレイヤー%dがナッツ→タブロー%d", playerIdx, toCol), []*Card{c})
	g.checkRoundEndForPlayer(playerIdx)
	return nil
}

// MoveWasteToFoundation ウェイストのトップを foundationIdx へ送る。
func (g *Nertz) MoveWasteToFoundation(playerIdx, foundationIdx int) error {
	if err := g.assertPlayable(); err != nil {
		return err
	}
	p, err := g.player(playerIdx)
	if err != nil {
		return err
	}
	f, err := g.foundation(foundationIdx)
	if err != nil {
		return err
	}
	top := p.WasteTop()
	if top == nil {
		return errors.New("nertz: waste is empty")
	}
	if !f.CanAccept(top) {
		return errors.New("nertz: cannot place card on foundation")
	}
	g.takeSnapshot(playerIdx)
	c := p.PopWaste()
	if err := f.Push(c, p.GetDeckIdx()); err != nil {
		return err
	}
	g.moveCount++
	g.appendLog(playerIdx, "moveWF", fmt.Sprintf("プレイヤー%dがウェイスト→ファウンデーション%d", playerIdx, foundationIdx), []*Card{c})
	return nil
}

// MoveWasteToTableau ウェイストのトップをタブロー toCol に置く。
func (g *Nertz) MoveWasteToTableau(playerIdx, toCol int) error {
	if err := g.assertPlayable(); err != nil {
		return err
	}
	p, err := g.player(playerIdx)
	if err != nil {
		return err
	}
	if toCol < 0 || toCol >= NertzTableauCnt {
		return errors.New("nertz: invalid tableau column")
	}
	top := p.WasteTop()
	if top == nil {
		return errors.New("nertz: waste is empty")
	}
	if !canPlaceOnNertzTableau(top, p.TableauTop(toCol)) {
		return errors.New("nertz: cannot place card on tableau")
	}
	g.takeSnapshot(playerIdx)
	c := p.PopWaste()
	p.PushTableau(toCol, &NertzTableauCard{Card: c, FaceUp: true})
	g.moveCount++
	g.appendLog(playerIdx, "moveWT", fmt.Sprintf("プレイヤー%dがウェイスト→タブロー%d", playerIdx, toCol), []*Card{c})
	return nil
}

// MoveTableauToFoundation タブロー fromCol の最下段カード (= トップ) を foundationIdx へ送る。
// 列が空になった場合はナッツパイルから 1 枚自動補充する。
func (g *Nertz) MoveTableauToFoundation(playerIdx, fromCol, foundationIdx int) error {
	if err := g.assertPlayable(); err != nil {
		return err
	}
	p, err := g.player(playerIdx)
	if err != nil {
		return err
	}
	if fromCol < 0 || fromCol >= NertzTableauCnt {
		return errors.New("nertz: invalid from column")
	}
	if p.TableauSize(fromCol) == 0 {
		return errors.New("nertz: tableau column is empty")
	}
	f, err := g.foundation(foundationIdx)
	if err != nil {
		return err
	}
	top := p.TableauTop(fromCol)
	if !f.CanAccept(top) {
		return errors.New("nertz: cannot place card on foundation")
	}
	g.takeSnapshot(playerIdx)
	tail := p.TakeTableauTail(fromCol, p.TableauSize(fromCol)-1)
	if err := f.Push(tail[0].Card, p.GetDeckIdx()); err != nil {
		return err
	}
	g.autoFillFromNertz(p, fromCol)
	g.moveCount++
	g.appendLog(playerIdx, "moveTF", fmt.Sprintf("プレイヤー%dがタブロー%d→ファウンデーション%d", playerIdx, fromCol, foundationIdx), []*Card{tail[0].Card})
	g.checkRoundEndForPlayer(playerIdx)
	return nil
}

// MoveTableauToTableau タブロー fromCol の fromIdx 以降を toCol に移す。
// 移動後に fromCol が空になった場合はナッツパイルから 1 枚自動補充する。
func (g *Nertz) MoveTableauToTableau(playerIdx, fromCol, fromIdx, toCol int) error {
	if err := g.assertPlayable(); err != nil {
		return err
	}
	p, err := g.player(playerIdx)
	if err != nil {
		return err
	}
	if fromCol < 0 || fromCol >= NertzTableauCnt || toCol < 0 || toCol >= NertzTableauCnt {
		return errors.New("nertz: invalid tableau column")
	}
	if fromCol == toCol {
		return errors.New("nertz: from and to columns are the same")
	}
	col := p.GetTableauColumn(fromCol)
	if fromIdx < 0 || fromIdx >= len(col) {
		return errors.New("nertz: invalid card index")
	}
	bottom := col[fromIdx].Card
	if !canPlaceOnNertzTableau(bottom, p.TableauTop(toCol)) {
		return errors.New("nertz: cannot place card on tableau")
	}
	g.takeSnapshot(playerIdx)
	tail := p.TakeTableauTail(fromCol, fromIdx)
	moved := make([]*Card, len(tail))
	for i, tc := range tail {
		p.PushTableau(toCol, tc)
		moved[i] = tc.Card
	}
	g.autoFillFromNertz(p, fromCol)
	g.moveCount++
	g.appendLog(playerIdx, "moveTT", fmt.Sprintf("プレイヤー%dがタブロー%d→タブロー%d", playerIdx, fromCol, toCol), moved)
	return nil
}

// autoFillFromNertz fromCol が空でかつナッツパイルに残りがあれば 1 枚補充する。
func (g *Nertz) autoFillFromNertz(p *NertzPlayer, fromCol int) {
	if p.TableauSize(fromCol) > 0 {
		return
	}
	if c := p.PopNertz(); c != nil {
		p.PushTableau(fromCol, &NertzTableauCard{Card: c, FaceUp: true})
	}
}

// checkRoundEndForPlayer playerIdx のナッツパイルが空ならラウンド終了処理を行う。
func (g *Nertz) checkRoundEndForPlayer(playerIdx int) {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return
	}
	if g.players[playerIdx].NertzSize() == 0 {
		g.phase = NertzPhaseRoundEnd
		g.winnerIdx = playerIdx
		g.applyRoundScoring()
		g.appendLog(playerIdx, "nertz", fmt.Sprintf("プレイヤー%dがナッツパイルを出し切った", playerIdx), nil)
		if g.checkMatchEnd() {
			g.phase = NertzPhaseGameEnd
		}
	}
}

// applyRoundScoring ラウンド終了時の得点を各プレイヤーに反映する。
// +1 / foundation contribution, -2 / nertz pile remaining card
func (g *Nertz) applyRoundScoring() {
	for _, f := range g.foundations {
		for _, deckIdx := range f.GetContributors() {
			if deckIdx >= 0 && deckIdx < len(g.players) {
				g.players[deckIdx].AddScore(1)
			}
		}
	}
	for _, p := range g.players {
		p.AddScore(-2 * p.NertzSize())
	}
}

// checkMatchEnd 目標スコアに到達したプレイヤーがいればマッチ勝者として記録する。
// 同時到達は最大スコアを優先 (タイは最小 idx)。
func (g *Nertz) checkMatchEnd() bool {
	winner, top := -1, g.config.TargetScore-1
	for i, p := range g.players {
		if p.GetScore() > top {
			winner = i
			top = p.GetScore()
		}
	}
	if winner >= 0 {
		g.matchWinner = winner
		return true
	}
	return false
}

// --- Tick / CPU AI ---

// NertzAction CPU/人間アクションを構造化した記述子。
// プレゼンテーション層がアニメーションキューに使う。
type NertzAction struct {
	PlayerIdx  int
	ActionType string // "moveNF", "moveNT", "moveWF", "moveWT", "moveTF", "moveTT", "draw", "recycle"
	FromZone   NertzZone
	FromCol    int // tableau col (otherwise -1)
	FromIdx    int // tableau substack start idx (otherwise -1)
	ToZone     NertzZone
	ToCol      int // tableau col or foundation idx
	Cards      []*Card
}

// Tick CPU プレイヤーを1ラウンドぶん進める (各 CPU の手数は CpuTickMoves 上限)。
// 適用されたアクション一覧を返す。フェーズが Playing でないときは何もしない。
//
// CPU の処理順は毎 tick 無作為化する。固定順だと低 index の CPU が共有
// ファウンデーションへの送り込みで一貫した先行優位を持ってしまうため
// (PR #1528 レビュー指摘)。
//
// 各 CPU の手選択は CpuDifficulty に従って分岐する (FindCpuMove 参照):
// Easy/Normal は first-match を高速反復、Hard は先読みスコアで質の高い
// 手を選ぶ (issue #1562)。
func (g *Nertz) Tick() []*NertzAction {
	if g.phase != NertzPhasePlaying {
		return nil
	}
	budget := g.config.ResolvedCpuTickMoves()
	out := make([]*NertzAction, 0, budget*len(g.players))
	cpuOrder := make([]int, 0, len(g.players))
	for i, p := range g.players {
		if p.GetIsCpu() {
			cpuOrder = append(cpuOrder, i)
		}
	}
	rand.Shuffle(len(cpuOrder), func(i, j int) { cpuOrder[i], cpuOrder[j] = cpuOrder[j], cpuOrder[i] })
	for _, i := range cpuOrder {
		for k := 0; k < budget; k++ {
			if g.phase != NertzPhasePlaying {
				return out
			}
			move := g.FindCpuMove(i)
			if move == nil {
				break
			}
			if err := g.applyAction(move); err != nil {
				break
			}
			out = append(out, move)
		}
	}
	return out
}

// FindCpuMove playerIdx の CPU が次に行うべき手 (アクション記述子) を返す。
// 該当する手がなければ nil。
//
// 優先度 (Easy/Normal)::
//  1. ナッツパイル → ファウンデーション
//  2. ウェイスト → ファウンデーション
//  3. タブロー → ファウンデーション
//  4. ナッツパイル → タブロー (空列以外を優先)
//  5. ウェイスト → タブロー
//  6. タブロー間移動 (列単位; 空 toCol への進歩のない移動はスキップ)
//  7. ストックから引く (リサイクル含む)
//
// Hard 難易度では `findCpuMoveHard` に分岐し、全ての合法手を列挙してから
// 「ナッツパイル削減を最大化する」スコア評価で最適手を選ぶ (issue #1562)。
// Easy/Normal はこのまま first-match で素早く返す — その方が tick budget
// 内で多くの手を消費できるので「軽快なミス」の表現に向いている。
//
// step 6 の fromIdx は現状 0 固定で、部分スタック移動は実装していない。
func (g *Nertz) FindCpuMove(playerIdx int) *NertzAction {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	if g.config.CpuDifficulty == NertzCpuDifficultyHard {
		return g.findCpuMoveHard(playerIdx)
	}
	return g.findCpuMoveFast(playerIdx)
}

// findCpuMoveFast は Easy/Normal 用の first-match ヒューリスティック。
// 旧来の FindCpuMove と完全に同一の挙動 — 最初に見つかった合法手を返す。
func (g *Nertz) findCpuMoveFast(playerIdx int) *NertzAction {
	p := g.players[playerIdx]

	// 1. Nertz top → any foundation
	if top := p.NertzTop(); top != nil {
		for fi, f := range g.foundations {
			if f.CanAccept(top) {
				return &NertzAction{
					PlayerIdx:  playerIdx,
					ActionType: "moveNF",
					FromZone:   NertzZoneNertz, FromCol: -1, FromIdx: -1,
					ToZone: NertzZoneFoundation, ToCol: fi,
					Cards: []*Card{top},
				}
			}
		}
	}

	// 2. Waste top → any foundation
	if top := p.WasteTop(); top != nil {
		for fi, f := range g.foundations {
			if f.CanAccept(top) {
				return &NertzAction{
					PlayerIdx:  playerIdx,
					ActionType: "moveWF",
					FromZone:   NertzZoneWaste, FromCol: -1, FromIdx: -1,
					ToZone: NertzZoneFoundation, ToCol: fi,
					Cards: []*Card{top},
				}
			}
		}
	}

	// 3. Tableau bottom → any foundation
	for col := 0; col < NertzTableauCnt; col++ {
		top := p.TableauTop(col)
		if top == nil {
			continue
		}
		for fi, f := range g.foundations {
			if f.CanAccept(top) {
				return &NertzAction{
					PlayerIdx:  playerIdx,
					ActionType: "moveTF",
					FromZone:   NertzZoneTableau, FromCol: col, FromIdx: p.TableauSize(col) - 1,
					ToZone: NertzZoneFoundation, ToCol: fi,
					Cards: []*Card{top},
				}
			}
		}
	}

	// 4. Nertz top → tableau (non-empty preferred)
	if top := p.NertzTop(); top != nil {
		for _, preferEmpty := range []bool{false, true} {
			for col := 0; col < NertzTableauCnt; col++ {
				colEmpty := p.TableauSize(col) == 0
				if colEmpty != preferEmpty {
					continue
				}
				if canPlaceOnNertzTableau(top, p.TableauTop(col)) {
					return &NertzAction{
						PlayerIdx:  playerIdx,
						ActionType: "moveNT",
						FromZone:   NertzZoneNertz, FromCol: -1, FromIdx: -1,
						ToZone: NertzZoneTableau, ToCol: col,
						Cards: []*Card{top},
					}
				}
			}
		}
	}

	// 5. Waste top → tableau
	if top := p.WasteTop(); top != nil {
		for col := 0; col < NertzTableauCnt; col++ {
			if canPlaceOnNertzTableau(top, p.TableauTop(col)) {
				return &NertzAction{
					PlayerIdx:  playerIdx,
					ActionType: "moveWT",
					FromZone:   NertzZoneWaste, FromCol: -1, FromIdx: -1,
					ToZone: NertzZoneTableau, ToCol: col,
					Cards: []*Card{top},
				}
			}
		}
	}

	// 6. Tableau substack → another tableau col.
	// Skip whole-column moves into an empty destination when the Nertz pile
	// can no longer refill the source column. Otherwise the CPU can swap a
	// stack between two empty columns each tick forever (no card revealed,
	// no nertz reduction) — see PR #1528 review.
	for fromCol := 0; fromCol < NertzTableauCnt; fromCol++ {
		col := p.GetTableauColumn(fromCol)
		if len(col) == 0 {
			continue
		}
		bottom := col[0].Card
		for toCol := 0; toCol < NertzTableauCnt; toCol++ {
			if toCol == fromCol {
				continue
			}
			if !canPlaceOnNertzTableau(bottom, p.TableauTop(toCol)) {
				continue
			}
			if p.TableauTop(toCol) == nil && p.NertzSize() == 0 {
				continue
			}
			cards := make([]*Card, len(col))
			for i, tc := range col {
				cards[i] = tc.Card
			}
			return &NertzAction{
				PlayerIdx:  playerIdx,
				ActionType: "moveTT",
				FromZone:   NertzZoneTableau, FromCol: fromCol, FromIdx: 0,
				ToZone: NertzZoneTableau, ToCol: toCol,
				Cards: cards,
			}
		}
	}

	// 7. Draw from stock (or recycle)
	if p.StockSize() > 0 || p.WasteSize() > 0 {
		return &NertzAction{
			PlayerIdx:  playerIdx,
			ActionType: "draw",
			FromZone:   NertzZoneStock, FromCol: -1, FromIdx: -1,
			ToZone: NertzZoneWaste, ToCol: -1,
		}
	}
	return nil
}

// findCpuMoveHard は Hard 難易度用の先読みスコアアルゴリズム (issue #1562)。
// 全ての合法手を列挙し、scoreNertzMove で評価して最高得点の手を返す。
// 同点の場合は first-match と同じ列挙順で前のものを優先する (決定的に) 。
//
// 旧来の Hard は Easy/Normal と同じヒューリスティックを高速回転させるだけ
// で、強さの差は単に「速い」というストレスフルな UX に偏っていた。本実装
// は per-tick budget は据え置きにしたまま、選ばれる手の質で差別化する。
func (g *Nertz) findCpuMoveHard(playerIdx int) *NertzAction {
	candidates := g.enumerateCpuMoves(playerIdx)
	if len(candidates) == 0 {
		return nil
	}
	bestScore := scoreNertzMoveMin
	var best *NertzAction
	for _, m := range candidates {
		s := g.scoreNertzMove(playerIdx, m)
		if s > bestScore {
			bestScore = s
			best = m
		}
	}
	return best
}

// scoreNertzMoveMin は scoreNertzMove が決して返さない値の番兵。findCpuMoveHard
// の比較で「初回は必ず置き換わる」性質を保つ。
const scoreNertzMoveMin = -1 << 30

// enumerateCpuMoves は playerIdx の合法手を全て列挙する。findCpuMoveFast の
// 1〜7 と同じカテゴリ順に追加するので、同点比較の決定性が保たれる。draw は
// 他に手が無かったときだけ追加する (常に並べると Tick budget を浪費する)。
func (g *Nertz) enumerateCpuMoves(playerIdx int) []*NertzAction {
	p := g.players[playerIdx]
	moves := make([]*NertzAction, 0, 16)

	// 1. Nertz top → any foundation
	if top := p.NertzTop(); top != nil {
		for fi, f := range g.foundations {
			if f.CanAccept(top) {
				moves = append(moves, &NertzAction{
					PlayerIdx:  playerIdx,
					ActionType: "moveNF",
					FromZone:   NertzZoneNertz, FromCol: -1, FromIdx: -1,
					ToZone: NertzZoneFoundation, ToCol: fi,
					Cards: []*Card{top},
				})
			}
		}
	}

	// 2. Waste top → any foundation
	if top := p.WasteTop(); top != nil {
		for fi, f := range g.foundations {
			if f.CanAccept(top) {
				moves = append(moves, &NertzAction{
					PlayerIdx:  playerIdx,
					ActionType: "moveWF",
					FromZone:   NertzZoneWaste, FromCol: -1, FromIdx: -1,
					ToZone: NertzZoneFoundation, ToCol: fi,
					Cards: []*Card{top},
				})
			}
		}
	}

	// 3. Tableau bottom → any foundation
	for col := range NertzTableauCnt {
		top := p.TableauTop(col)
		if top == nil {
			continue
		}
		for fi, f := range g.foundations {
			if f.CanAccept(top) {
				moves = append(moves, &NertzAction{
					PlayerIdx:  playerIdx,
					ActionType: "moveTF",
					FromZone:   NertzZoneTableau, FromCol: col, FromIdx: p.TableauSize(col) - 1,
					ToZone: NertzZoneFoundation, ToCol: fi,
					Cards: []*Card{top},
				})
			}
		}
	}

	// 4. Nertz top → tableau (both empty and non-empty considered;
	//    scoreNertzMove discriminates between them).
	if top := p.NertzTop(); top != nil {
		for col := range NertzTableauCnt {
			if canPlaceOnNertzTableau(top, p.TableauTop(col)) {
				moves = append(moves, &NertzAction{
					PlayerIdx:  playerIdx,
					ActionType: "moveNT",
					FromZone:   NertzZoneNertz, FromCol: -1, FromIdx: -1,
					ToZone: NertzZoneTableau, ToCol: col,
					Cards: []*Card{top},
				})
			}
		}
	}

	// 5. Waste top → tableau
	if top := p.WasteTop(); top != nil {
		for col := range NertzTableauCnt {
			if canPlaceOnNertzTableau(top, p.TableauTop(col)) {
				moves = append(moves, &NertzAction{
					PlayerIdx:  playerIdx,
					ActionType: "moveWT",
					FromZone:   NertzZoneWaste, FromCol: -1, FromIdx: -1,
					ToZone: NertzZoneTableau, ToCol: col,
					Cards: []*Card{top},
				})
			}
		}
	}

	// 6. Tableau substack → another tableau col. Apply the same "skip useless
	//    swap into empty when nertz pile is empty" guard as findCpuMoveFast.
	//    Two perf notes (PR #1587 review):
	//    - Read p.tableau directly to skip the defensive copy in
	//      GetTableauColumn — same package, same trust boundary.
	//    - Build the shared `cards` slice once per fromCol; every valid toCol
	//      can reuse it because moveTT moves the entire column verbatim.
	for fromCol := range NertzTableauCnt {
		col := p.tableau[fromCol]
		if len(col) == 0 {
			continue
		}
		bottom := col[0].Card
		var cards []*Card // lazily built on first valid destination
		for toCol := range NertzTableauCnt {
			if toCol == fromCol {
				continue
			}
			if !canPlaceOnNertzTableau(bottom, p.TableauTop(toCol)) {
				continue
			}
			if p.TableauTop(toCol) == nil && p.NertzSize() == 0 {
				continue
			}
			if cards == nil {
				cards = make([]*Card, len(col))
				for i, tc := range col {
					cards[i] = tc.Card
				}
			}
			moves = append(moves, &NertzAction{
				PlayerIdx:  playerIdx,
				ActionType: "moveTT",
				FromZone:   NertzZoneTableau, FromCol: fromCol, FromIdx: 0,
				ToZone: NertzZoneTableau, ToCol: toCol,
				Cards: cards,
			})
		}
	}

	// 7. Draw — only as last resort. Adding it unconditionally would dilute
	//    the score comparison (draw is always legal when stock or waste has
	//    cards) and the Hard CPU would constantly pick "draw" over a +5 TT
	//    reorganization since draw is +1.
	if len(moves) == 0 && (p.StockSize() > 0 || p.WasteSize() > 0) {
		moves = append(moves, &NertzAction{
			PlayerIdx:  playerIdx,
			ActionType: "draw",
			FromZone:   NertzZoneStock, FromCol: -1, FromIdx: -1,
			ToZone: NertzZoneWaste, ToCol: -1,
		})
	}
	return moves
}

// scoreNertzMove rates a candidate move on its strategic value for a Hard
// CPU. The headline lever is "reduce the Nertz pile" — emptying it is the
// only path to scoring (and ending the round), so anything that pulls a
// card off the Nertz pile beats anything that doesn't, all else equal.
// Foundation work is the next most valuable since each foundation card is
// worth 1 point at round end. Tableau reorganization is filler — useful
// when nothing else is available, but not chased over real progress.
//
// The base score categories below intentionally leave gaps so secondary
// adjustments (avoiding empty-column traps when the Nertz pile is empty,
// preferring lower foundation indices, etc.) can fine-tune without
// crossing category boundaries. See issue #1562.
func (g *Nertz) scoreNertzMove(playerIdx int, m *NertzAction) int {
	if m == nil {
		return scoreNertzMoveMin
	}
	p := g.players[playerIdx]
	score := 0
	switch m.ActionType {
	case "moveNF":
		// Nertz → Foundation: best possible move. Pile -1 AND foundation +1.
		// Two birds; this is what Nertz is fundamentally about.
		score = 100
	case "moveNT":
		// Nertz → Tableau: pile -1 even though no foundation progress. Still
		// extremely valuable because pile reduction is the win condition.
		score = 50
	case "moveWF":
		// Waste → Foundation: foundation +1, no pile effect.
		score = 30
	case "moveTF":
		// Tableau → Foundation: foundation +1, frees a tableau slot. Slightly
		// less valuable than waste-to-foundation because tableau columns
		// generally have downstream value as Nertz/Waste landing pads.
		score = 20
	case "moveWT":
		// Waste → Tableau: defensive; clears waste so the next stock cycle
		// reveals a fresh card.
		score = 10
	case "moveTT":
		// Tableau reorganization. Sometimes unlocks a TF down the line, but
		// without lookahead we can't tell — keep low priority.
		score = 5
	case "draw":
		// Last resort.
		score = 1
	}

	// Adjustment 1: when the Nertz pile is empty, the only way to refill an
	// emptied tableau column is via the waste-stock cycle (the actual rule
	// allows "anything fits in an empty column", but the player cannot
	// generate cards on demand). Penalize moves that empty a column with
	// no clear refill plan so the CPU doesn't strand its waste pile.
	if p.NertzSize() == 0 {
		switch m.ActionType {
		case "moveTF", "moveTT":
			if m.FromCol >= 0 && p.TableauSize(m.FromCol) == len(m.Cards) {
				score -= 8
			}
		}
	}

	// Adjustment 2: when filling an empty tableau column from nertz or
	// waste, a small penalty preserves the empty column's flexibility —
	// it can accept any card, so spending it on a known card wastes
	// optionality. Mirrors the "non-empty preferred" heuristic in
	// findCpuMoveFast (PR #1587 review). Magnitude (-3) keeps the
	// adjusted scores inside their categories: moveNT 50 → 47, still
	// well above moveWF 30; moveWT 10 → 7, still above moveTT 5.
	switch m.ActionType {
	case "moveNT", "moveWT":
		if m.ToZone == NertzZoneTableau && p.TableauTop(m.ToCol) == nil {
			score -= 3
		}
	}

	// Adjustment 3: tableau-to-foundation favors smaller foundation indices
	// (deterministic tiebreak across runs). enumerateCpuMoves already iterates
	// foundations in index order, so the Hard CPU and the Easy/Normal CPU
	// produce the same move when ties survive — the test suite stays stable.
	//
	// Modulo NertzFoundationsPerPlayer caps the swing at 0..3 regardless of
	// player count: in N-player games there are NertzFoundationsPerPlayer*N
	// foundations (16 in 4-player, 24 in 6-player), and an unbounded
	// `score -= m.ToCol` would cross category boundaries (e.g. moveTF base
	// 20 minus index 11 = 9, less than moveWT base 10 — wrong winner). The
	// modulo gives a per-suit-position tiebreak that stays inside the moveTF
	// band (PR #1587 review).
	if m.ToZone == NertzZoneFoundation {
		score -= m.ToCol % NertzFoundationsPerPlayer
	}

	return score
}

// NertzHint 人間プレイヤー (idx=0) への次の推奨手。
type NertzHint struct {
	FromZone  string // "nertz" | "waste" | "tableau"
	FromCol   int    // tableau col, otherwise -1
	CardIndex int    // tableau substack start, otherwise -1
	ToZone    string // "tableau" | "foundation"
	ToCol     int    // tableau col or foundation idx
}

// GetHint プレイヤー 0 (人間) の次の推奨手を返す。なければ nil。
func (g *Nertz) GetHint() *NertzHint {
	if g.phase != NertzPhasePlaying {
		return nil
	}
	move := g.FindCpuMove(0)
	if move == nil {
		return nil
	}
	from, fromCol, idx := nertzZoneStringForHint(move.FromZone, move.FromCol, move.FromIdx)
	to, _, _ := nertzZoneStringForHint(move.ToZone, move.ToCol, -1)
	if to == "" {
		return nil
	}
	return &NertzHint{
		FromZone:  from,
		FromCol:   fromCol,
		CardIndex: idx,
		ToZone:    to,
		ToCol:     move.ToCol,
	}
}

func nertzZoneStringForHint(z NertzZone, col, idx int) (string, int, int) {
	switch z {
	case NertzZoneNertz:
		return "nertz", -1, -1
	case NertzZoneWaste:
		return "waste", -1, -1
	case NertzZoneTableau:
		return "tableau", col, idx
	case NertzZoneFoundation:
		return "foundation", col, -1
	default:
		return "", -1, -1
	}
}

// applyAction NertzAction を実行する (Tick / 外部 API 共用)。
func (g *Nertz) applyAction(a *NertzAction) error {
	switch a.ActionType {
	case "moveNF":
		return g.MoveNertzToFoundation(a.PlayerIdx, a.ToCol)
	case "moveNT":
		return g.MoveNertzToTableau(a.PlayerIdx, a.ToCol)
	case "moveWF":
		return g.MoveWasteToFoundation(a.PlayerIdx, a.ToCol)
	case "moveWT":
		return g.MoveWasteToTableau(a.PlayerIdx, a.ToCol)
	case "moveTF":
		return g.MoveTableauToFoundation(a.PlayerIdx, a.FromCol, a.ToCol)
	case "moveTT":
		return g.MoveTableauToTableau(a.PlayerIdx, a.FromCol, a.FromIdx, a.ToCol)
	case "draw":
		return g.DrawStock(a.PlayerIdx)
	default:
		return fmt.Errorf("nertz: unknown action %q", a.ActionType)
	}
}

// --- Helpers ---

// canPlaceOnNertzTableau Klondike と同じ降順交互色の判定。
// destTop が nil (= 空列) の場合は任意のカードを受け入れる。
// (Klondike の K-only ルールは Nertz では採用しない —
// 空列は補充ループ/任意カードでの再利用が想定されるため。)
func canPlaceOnNertzTableau(card, destTop *Card) bool {
	if card == nil {
		return false
	}
	if destTop == nil {
		return true
	}
	if !isAlternateColor(card, destTop) {
		return false
	}
	return card.GetValue() == destTop.GetValue()-1
}

func (g *Nertz) assertPlayable() error {
	if g.phase != NertzPhasePlaying {
		return errors.New("nertz: game is not in playing phase")
	}
	return nil
}

func (g *Nertz) player(idx int) (*NertzPlayer, error) {
	if idx < 0 || idx >= len(g.players) {
		return nil, fmt.Errorf("nertz: invalid player index %d", idx)
	}
	return g.players[idx], nil
}

func (g *Nertz) foundation(idx int) (*NertzFoundation, error) {
	if idx < 0 || idx >= len(g.foundations) {
		return nil, fmt.Errorf("nertz: invalid foundation index %d", idx)
	}
	return g.foundations[idx], nil
}

func (g *Nertz) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: g.moveCount,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- Snapshot / Undo ---

// takeSnapshot 現在の人間プレイヤーの直前状態をスナップショット保存する。
// 人間プレイヤー (playerIdx == 0) のアクションのみ Undo の対象とする。
// CPU の手は Undo 不能 (リアルタイム性を保つため)。
//
// Undo は単一ステップのみサポートするため、履歴は最大1件に制限する。
// (各スナップショットは players + foundations の JSON であり、長時間の
// セッションで append し続けると 1 セッションあたり数 MB に到達する。
// PR #1528 レビュー指摘。)
func (g *Nertz) takeSnapshot(playerIdx int) {
	if playerIdx != 0 {
		return
	}
	pj, err := json.Marshal(g.players)
	if err != nil {
		return
	}
	fj, err := json.Marshal(g.foundations)
	if err != nil {
		return
	}
	g.history = []*nertzSnapshot{{
		playersJSON:     pj,
		foundationsJSON: fj,
		phase:           g.phase,
		moveCount:       g.moveCount,
		winnerIdx:       g.winnerIdx,
	}}
}

// CanUndo Undo 可能かどうか
func (g *Nertz) CanUndo() bool {
	return len(g.history) > 0 && g.phase == NertzPhasePlaying
}

// Undo 直前の人間アクションを取り消す。
// CPU の手は Undo の対象ではないため、復元前後で他プレイヤーの盤面が
// 戻ってしまう点に注意 (利用者には人間視点の単発巻き戻しとして説明)。
func (g *Nertz) Undo() error {
	if !g.CanUndo() {
		return errors.New("nertz: cannot undo")
	}
	snap := g.history[len(g.history)-1]
	g.history = g.history[:len(g.history)-1]
	var players []*NertzPlayer
	if err := json.Unmarshal(snap.playersJSON, &players); err != nil {
		return fmt.Errorf("nertz: failed to restore players: %w", err)
	}
	var founds []*NertzFoundation
	if err := json.Unmarshal(snap.foundationsJSON, &founds); err != nil {
		return fmt.Errorf("nertz: failed to restore foundations: %w", err)
	}
	g.players = players
	g.foundations = founds
	g.phase = snap.phase
	g.moveCount = snap.moveCount
	g.winnerIdx = snap.winnerIdx
	return nil
}

// --- State getters / setters ---

// GetPhase 現在のフェーズ
func (g *Nertz) GetPhase() NertzPhase { return g.phase }

// SetPhase テスト用フェーズ設定
func (g *Nertz) SetPhase(p NertzPhase) { g.phase = p }

// GetRoundNo ラウンド番号 (1始まり)
func (g *Nertz) GetRoundNo() int { return g.roundNo }

// GetWinnerIdx ラウンド勝者 (-1 = 未確定)
func (g *Nertz) GetWinnerIdx() int { return g.winnerIdx }

// GetMatchWinner マッチ勝者 (-1 = 未確定)
func (g *Nertz) GetMatchWinner() int { return g.matchWinner }

// GetConfig 設定取得
func (g *Nertz) GetConfig() NertzConfig { return g.config }

// GetPlayers プレイヤースナップショット
func (g *Nertz) GetPlayers() []*NertzPlayer { return g.players }

// GetFoundations ファウンデーションスナップショット
func (g *Nertz) GetFoundations() []*NertzFoundation { return g.foundations }

// SetFoundations テスト/復元用
func (g *Nertz) SetFoundations(fs []*NertzFoundation) { g.foundations = fs }

// GetActionLog 棋譜
func (g *Nertz) GetActionLog() []*ActionLogEntry { return g.actionLog }

// GetGameEndFlag returns true once the game has left the playing phase.
func (g *Nertz) GetGameEndFlag() bool { return g.phase != NertzPhasePlaying }

// GetMoveCount 手数
func (g *Nertz) GetMoveCount() int { return g.moveCount }

// --- JSON ---

// nertzJSON is the JSON wire format for Nertz.
type nertzJSON struct {
	Decks       []*TrumpCards      `json:"dk"`
	Players     []*NertzPlayer     `json:"pl"`
	Foundations []*NertzFoundation `json:"fd"`
	Config      NertzConfig        `json:"cf"`
	Phase       NertzPhase         `json:"ph"`
	RoundNo     int                `json:"rn"`
	WinnerIdx   int                `json:"wi"`
	MatchWinner int                `json:"mw"`
	MoveCount   int                `json:"mc"`
	ActionLog   []*ActionLogEntry  `json:"al"`
	// History must round-trip: the Cloudflare Worker is stateless per request
	// and rebuilds the game from KV every call, so an unpersisted undo stack
	// means Undo/UndoN/UndoToEscape silently never work in production (#4478).
	History []*nertzSnapshot `json:"hi,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (g *Nertz) MarshalJSON() ([]byte, error) {
	return json.Marshal(nertzJSON{
		Decks:       g.decks,
		Players:     g.players,
		Foundations: g.foundations,
		Config:      g.config,
		Phase:       g.phase,
		RoundNo:     g.roundNo,
		WinnerIdx:   g.winnerIdx,
		MatchWinner: g.matchWinner,
		MoveCount:   g.moveCount,
		ActionLog:   g.actionLog,
		History:     g.history,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *Nertz) UnmarshalJSON(data []byte) error {
	var j nertzJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > nertzMaxSliceLen || len(j.Foundations) > nertzMaxSliceLen ||
		len(j.ActionLog) > nertzMaxSliceLen {
		return errors.New("nertz: input array exceeds maximum allowed size")
	}
	g.decks = j.Decks
	g.players = j.Players
	g.foundations = j.Foundations
	g.config = j.Config
	g.phase = j.Phase
	g.roundNo = j.RoundNo
	g.winnerIdx = j.WinnerIdx
	g.matchWinner = j.MatchWinner
	g.moveCount = j.MoveCount
	if len(j.History) > nertzMaxSliceLen {
		return errors.New("nertz: history exceeds maximum allowed size")
	}
	g.actionLog = j.ActionLog
	g.history = j.History
	return nil
}

// nertzSnapshotMaxBytes bounds one embedded snapshot document. A full table is
// a few KB, so 1 MiB is far above any legitimate value and exists only to stop
// a hostile KV entry from being expanded.
const nertzSnapshotMaxBytes = 1 << 20

// nertzSnapshotJSON is the wire format for a single undo snapshot.
// nertzSnapshot uses unexported fields, so marshalling it directly would emit
// `[{},{}]` -- the undo depth would survive but every snapshot would be blank,
// and Undo would wipe the board instead of rewinding it (#4478).
//
// The two document fields ride as json.RawMessage rather than []byte: a []byte
// would be re-encoded as base64 on every save, which costs a third more bytes
// per snapshot in KV for no benefit.
type nertzSnapshotJSON struct {
	Players     json.RawMessage `json:"pl"`
	Foundations json.RawMessage `json:"fd"`
	Phase       NertzPhase      `json:"ph"`
	MoveCount   int             `json:"mc"`
	WinnerIdx   int             `json:"wi"`
}

// MarshalJSON implements json.Marshaler for nertzSnapshot.
func (s *nertzSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(nertzSnapshotJSON{
		Players:     s.playersJSON,
		Foundations: s.foundationsJSON,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		WinnerIdx:   s.winnerIdx,
	})
}

// UnmarshalJSON implements json.Unmarshaler for nertzSnapshot.
func (s *nertzSnapshot) UnmarshalJSON(data []byte) error {
	var j nertzSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > nertzSnapshotMaxBytes || len(j.Foundations) > nertzSnapshotMaxBytes {
		return errors.New("nertz: snapshot document exceeds maximum allowed size")
	}
	s.playersJSON = j.Players
	s.foundationsJSON = j.Foundations
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.winnerIdx = j.WinnerIdx
	return nil
}
