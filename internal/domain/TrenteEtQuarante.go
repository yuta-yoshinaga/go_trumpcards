//go:build !js || !wasm || extra

// Package domain トラント・エ・カラント (Trente et Quarante / Rouge et Noir / トラント・エ・カラント)
// のドメインモデル。
//
// Trente et Quarante はモンテカルロなどで親しまれるフランス発祥のカジノ・バンキング
// ゲーム。プレイヤーはカードのプレイ判断を一切行わず、ディーラーが 2 列 (黒 = Noir /
// 赤 = Rouge) を配る前に 1 種類のベットを選ぶだけの、最も単純なバンキングゲームである。
//
// # デッキ
//
// 6 デッキ 312 枚のシュー。ジョーカーは使わない。カード値は A=1、2〜10 は数字通り、
// J/Q/K=10。色は ♥♦ = 赤 (Rouge)、♠♣ = 黒 (Noir)。
//
// # 1 ラウンドの流れ
//
//  1. プレイヤーは配札前に Noir / Rouge / Couleur / Inverse のいずれか 1 種と、チップから
//     ステークを選んで賭ける。
//  2. ディーラーは黒 (Noir) 列を 1 枚ずつ表向きに配り、列の合計が 31 以上になるまで続ける
//     (合計は必ず 31〜40 に収まる)。続いて赤 (Rouge) 列を同様に配る。
//  3. 合計が小さい (31 に近い) 列が勝つ。
//  4. Couleur は「最初に配られた札 (= Noir 列の 1 枚目) の色」が勝ち列の色 (Noir 列 = 黒、
//     Rouge 列 = 赤) と一致すれば勝ち。Inverse は異なれば勝ち。
//  5. 両列が同点ならプッシュ (ステーク返却)。ただし 31 での同点は "Refait" となり、
//     胴元がステークの半額を取る (プレイヤーは半額を失う)。
//  6. 勝ちはイーブンマネー (1:1) 配当。
//
// # マルチラウンド
//
// プレイヤーはチップを保持し、ラウンドをまたいで賭け続ける (Casino War と同様)。シューの
// 残りが少なくなったら配札前に自動でシャッフルし直す。
//
// 本実装は extra ワーカーから到達可能なよう、シュー生成 (NewTrumpCardsWithDecks) と
// バンキング/解決ロジックをすべてインラインで持つ。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// トラント・エ・カラントの定数
const (
	// TrenteEtQuaranteDeckCount はシューを構成するデッキ数。
	TrenteEtQuaranteDeckCount = 6
	// TrenteEtQuaranteTotalCards はシューの総枚数 (52 × 6)。
	TrenteEtQuaranteTotalCards = CardCnt * TrenteEtQuaranteDeckCount
	// TrenteEtQuaranteTarget は各列を配り続ける下限合計 (これ以上で停止, 31〜40)。
	TrenteEtQuaranteTarget = 31
	// TrenteEtQuaranteDefaultChips はデフォルトの初期チップ。
	TrenteEtQuaranteDefaultChips = 1000
	// TrenteEtQuaranteMinBet は最低ベット額 (兼ベット単位)。
	TrenteEtQuaranteMinBet = 10
	// TrenteEtQuaranteMaxBet は最大ベット額。
	TrenteEtQuaranteMaxBet = 10000
	// TrenteEtQuaranteReshuffleThreshold はこの枚数未満でシューをシャッフルし直す閾値。
	TrenteEtQuaranteReshuffleThreshold = 52
	// TrenteEtQuaranteMaxRowLen は 1 列の配札枚数の防御的上限 (デシリアライズ検証用)。
	TrenteEtQuaranteMaxRowLen = 40
)

// TrenteEtQuarantePhase はゲームフェーズ。
type TrenteEtQuarantePhase int

// Trente et Quarante のフェーズ定数。ワイヤー値はフロントエンドの enum
// (BET=1, DEALT=2, END=3) と一致させる。本実装は配札と解決が原子的なため中間の
// DEALT=2 は出力せず、BET(1) と END(3) の 2 状態のみを取る。
const (
	// TrenteEtQuarantePhaseBet ベット受付中 (配札前)。ワイヤー値 0 (フロントの BET)。
	TrenteEtQuarantePhaseBet TrenteEtQuarantePhase = 0
	// TrenteEtQuarantePhaseResult ラウンド解決済み (結果表示; 次ラウンド待ち)。ワイヤー値 1 (フロントの RESULT)。
	TrenteEtQuarantePhaseResult TrenteEtQuarantePhase = 1
)

// 勝ち列の識別子
const (
	// TrenteEtQuaranteRowNone 勝ち列なし (プッシュ / Refait)
	TrenteEtQuaranteRowNone = -1
	// TrenteEtQuaranteRowNoir 黒 (Noir) 列の勝ち
	TrenteEtQuaranteRowNoir = 0
	// TrenteEtQuaranteRowRouge 赤 (Rouge) 列の勝ち
	TrenteEtQuaranteRowRouge = 1
)

// TrenteEtQuaranteResult はラウンド結果 (勝ち=1 / 引き分け(プッシュ)=0 / 負け=-1)。
// GameResult は共有ファイル internal/domain/game_result.go に移動したので到達可能に
// なったが、この型名は JSON ペイロードに出るため統合していない（#4462）。値は
// GameResult と同一。
type TrenteEtQuaranteResult int

const (
	// TrenteEtQuaranteResultLose 負け
	TrenteEtQuaranteResultLose TrenteEtQuaranteResult = -1
	// TrenteEtQuaranteResultDraw 引き分け (プッシュ)
	TrenteEtQuaranteResultDraw TrenteEtQuaranteResult = 0
	// TrenteEtQuaranteResultWin 勝ち
	TrenteEtQuaranteResultWin TrenteEtQuaranteResult = 1

	// trenteEtQuaranteMaxSliceLen はデシリアライズ時のスライス長の上限。
	trenteEtQuaranteMaxSliceLen = 1000
)

// デシリアライズ検証用のセンチネルエラー
var (
	errTrenteEtQuaranteInvalidStats  = errors.New("trenteetquarante: invalid player stats")
	errTrenteEtQuaranteNegativeChips = errors.New("trenteetquarante: negative chips")
)

// TrenteEtQuaranteHint はヒント情報 (バンキングゲームのため戦略はなく、教育的な助言)。
type TrenteEtQuaranteHint struct {
	Bet    TrenteEtQuaranteBet // 推奨ベット種別
	Reason string              // ヒント理由キー
}

// trenteEtQuaranteState はゲーム進行状態。
type trenteEtQuaranteState struct {
	phase        TrenteEtQuarantePhase
	currentBet   TrenteEtQuaranteBet
	stake        int
	noirRow      []*Card
	rougeRow     []*Card
	noirTotal    int
	rougeTotal   int
	winningRow   int  // TrenteEtQuaranteRowNone/Noir/Rouge
	firstCardRed bool // 最初に配られた札 (Noir 列 1 枚目) が赤か
	refait       bool // 31 での同点 (胴元が半額)
	result       TrenteEtQuaranteResult
	payout       int // このラウンドでチップに戻された総額 (負け=0, プッシュ=stake, 勝ち=stake*2, Refait=stake/2)
	roundNumber  int
	gameEndFlag  bool
	scored       bool // ラウンド結果を確定済みか (二重確定防止)
	actionLog    []*ActionLogEntry
}

// TrenteEtQuarante はトラント・エ・カラントの状態を保持する集約ルート。
type TrenteEtQuarante struct {
	trumpCards *TrumpCards
	player     *TrenteEtQuarantePlayer
	config     TrenteEtQuaranteConfig
	state      trenteEtQuaranteState
}

// NewTrenteEtQuarante はコンストラクタ。
func NewTrenteEtQuarante(trumpCards *TrumpCards, player *TrenteEtQuarantePlayer, config TrenteEtQuaranteConfig) *TrenteEtQuarante {
	g := &TrenteEtQuarante{
		trumpCards: trumpCards,
		player:     player,
		config:     config,
		state: trenteEtQuaranteState{
			phase:      TrenteEtQuarantePhaseBet,
			currentBet: config.DefaultBet,
			winningRow: TrenteEtQuaranteRowNone,
			actionLog:  make([]*ActionLogEntry, 0),
		},
	}
	return g
}

// NewDefaultTrenteEtQuarante は標準構成 (312 枚シュー + 初期チップ) を生成する。
// CUI / Web / Worker 構築の単一情報源。
func NewDefaultTrenteEtQuarante() *TrenteEtQuarante {
	g := NewTrenteEtQuarante(
		newTrenteEtQuaranteShoe(),
		NewTrenteEtQuarantePlayer(TrenteEtQuaranteDefaultChips),
		DefaultTrenteEtQuaranteConfig(),
	)
	g.trumpCards.Shuffle()
	return g
}

// newTrenteEtQuaranteShoe は 6 デッキ 312 枚のシューを生成する。NewTrumpCardsWithDecks は
// ビルドタグ無しの TrumpCards.go にあり extra ワーカーからも到達可能。
func newTrenteEtQuaranteShoe() *TrumpCards {
	return NewTrumpCardsWithDecks(TrenteEtQuaranteDeckCount, 0)
}

// --- ゲーム進行 ---

// Reset は新しいゲームを開始する。チップは最低ベット額を満たしていれば保持し、
// 満たさなければデフォルトに戻す。シューは新規に生成してシャッフルする。
func (g *TrenteEtQuarante) Reset() {
	if g.player == nil {
		g.player = NewTrenteEtQuarantePlayer(TrenteEtQuaranteDefaultChips)
	}
	if g.player.GetChips() < TrenteEtQuaranteMinBet {
		g.player.SetChips(TrenteEtQuaranteDefaultChips)
	}
	g.player.ResetStats()
	g.trumpCards = newTrenteEtQuaranteShoe()
	g.trumpCards.Shuffle()
	g.state = trenteEtQuaranteState{
		phase:      TrenteEtQuarantePhaseBet,
		currentBet: g.config.DefaultBet,
		winningRow: TrenteEtQuaranteRowNone,
		actionLog:  make([]*ActionLogEntry, 0),
	}
}

// NextRound は同じシュー・チップで次のラウンドを始める。ラウンド状態のみをクリアし、
// シューの残りが少なければ配札前にシャッフルし直す。
func (g *TrenteEtQuarante) NextRound() {
	round := g.state.roundNumber
	log := g.state.actionLog
	g.state = trenteEtQuaranteState{
		phase:       TrenteEtQuarantePhaseBet,
		currentBet:  g.config.DefaultBet,
		winningRow:  TrenteEtQuaranteRowNone,
		roundNumber: round,
		actionLog:   log,
	}
}

// ensureShoe はシューの残りが閾値未満なら新規シューを生成してシャッフルする。
func (g *TrenteEtQuarante) ensureShoe() {
	if g.trumpCards == nil || g.trumpCards.GetRemainingCount() < TrenteEtQuaranteReshuffleThreshold {
		g.trumpCards = newTrenteEtQuaranteShoe()
		g.trumpCards.Shuffle()
		g.appendLog(-1, "shuffle", "shoe reshuffled", nil)
	}
}

// PlaceBet はベット種別とステークを賭け、両列を配って即座に解決する。
func (g *TrenteEtQuarante) PlaceBet(bet TrenteEtQuaranteBet, stake int) error {
	if g.state.phase != TrenteEtQuarantePhaseBet {
		return NewDomainError(ErrWrongPhase, "bet is only allowed during the bet phase")
	}
	if !TrenteEtQuaranteBetValid(bet) {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("invalid bet type %d", bet))
	}
	if stake < TrenteEtQuaranteMinBet || stake%TrenteEtQuaranteMinBet != 0 || stake > TrenteEtQuaranteMaxBet {
		return NewDomainError(ErrInvalidAmount, "invalid stake amount")
	}
	if !g.player.SubtractChips(stake) {
		return NewDomainError(ErrInsufficientChips, "insufficient chips")
	}
	g.state.currentBet = bet
	g.state.stake = stake
	g.appendLog(0, "bet",
		fmt.Sprintf("%s stake=%d", TrenteEtQuaranteBetNames[bet], stake), nil)

	g.ensureShoe()
	g.dealRows()
	g.resolve()
	return nil
}

// dealRows は黒 (Noir) 列・赤 (Rouge) 列を順に配る。各列は合計が 31 以上になるまで
// 1 枚ずつ引く。最初に配られた札 (Noir 列 1 枚目) の色を記録する。
func (g *TrenteEtQuarante) dealRows() {
	g.state.noirRow, g.state.noirTotal = g.dealOneRow()
	g.state.rougeRow, g.state.rougeTotal = g.dealOneRow()
	if len(g.state.noirRow) > 0 {
		g.state.firstCardRed = trenteEtQuaranteIsRed(g.state.noirRow[0])
	}
	g.appendLog(-1, "deal",
		fmt.Sprintf("Noir=%d (%d cards), Rouge=%d (%d cards)",
			g.state.noirTotal, len(g.state.noirRow), g.state.rougeTotal, len(g.state.rougeRow)),
		nil)
}

// dealOneRow は合計が TrenteEtQuaranteTarget 以上になるまで札を引き、列と合計を返す。
func (g *TrenteEtQuarante) dealOneRow() ([]*Card, int) {
	row := make([]*Card, 0, 8)
	total := 0
	for total < TrenteEtQuaranteTarget {
		card := g.trumpCards.DrawCard()
		if card == nil {
			// シュー枯渇 (通常は ensureShoe で防止)。安全のため打ち切る。
			g.trumpCards = newTrenteEtQuaranteShoe()
			g.trumpCards.Shuffle()
			continue
		}
		row = append(row, card)
		total += trenteEtQuaranteCardValue(card)
	}
	return row, total
}

// resolve は両列の合計を比較し、ベット結果と配当を確定する。
func (g *TrenteEtQuarante) resolve() {
	switch {
	case g.state.noirTotal < g.state.rougeTotal:
		g.state.winningRow = TrenteEtQuaranteRowNoir
	case g.state.rougeTotal < g.state.noirTotal:
		g.state.winningRow = TrenteEtQuaranteRowRouge
	default:
		// 同点。31 での同点は Refait (胴元が半額)、それ以外はプッシュ。
		g.state.winningRow = TrenteEtQuaranteRowNone
		g.state.refait = g.state.noirTotal == TrenteEtQuaranteTarget
	}

	switch {
	case g.state.refait:
		// Refait: プレイヤーはステークの半額を失う (半額を返却)。
		g.state.result = TrenteEtQuaranteResultLose
		g.state.payout = g.state.stake / 2
		g.player.AddChips(g.state.payout)
		g.player.RecordRound(false)
		g.appendLog(-1, "refait",
			fmt.Sprintf("Refait at 31 — half stake to house (refund=%d)", g.state.payout), nil)
	case g.state.winningRow == TrenteEtQuaranteRowNone:
		// プッシュ: ステーク全額返却。
		g.state.result = TrenteEtQuaranteResultDraw
		g.state.payout = g.state.stake
		g.player.AddChips(g.state.payout)
		g.player.RecordRound(false)
		g.appendLog(-1, "push", "tie — stake returned", nil)
	default:
		won := trenteEtQuaranteBetWins(g.state.currentBet, g.state.winningRow, g.state.firstCardRed)
		if won {
			g.state.result = TrenteEtQuaranteResultWin
			g.state.payout = g.state.stake * 2
			g.player.AddChips(g.state.payout)
		} else {
			g.state.result = TrenteEtQuaranteResultLose
			g.state.payout = 0
		}
		g.player.RecordRound(won)
		g.appendLog(-1, "result",
			fmt.Sprintf("%s row wins; bet=%s -> %s (payout=%d)",
				trenteEtQuaranteRowName(g.state.winningRow),
				TrenteEtQuaranteBetNames[g.state.currentBet],
				trenteEtQuaranteResultName(g.state.result), g.state.payout), nil)
	}

	g.state.scored = true
	g.state.gameEndFlag = true
	g.state.phase = TrenteEtQuarantePhaseResult
	g.state.roundNumber++
}

// --- 解決ロジック (インライン) ---

// trenteEtQuaranteCardValue はカードの点数を返す (A=1, 2〜10 = 数字, J/Q/K=10)。
func trenteEtQuaranteCardValue(c *Card) int {
	if c == nil {
		return 0
	}
	v := c.GetValue()
	if v > 10 {
		return 10
	}
	return v
}

// trenteEtQuaranteIsRed は札が赤 (♥♦) かどうかを返す。
func trenteEtQuaranteIsRed(c *Card) bool {
	if c == nil {
		return false
	}
	d := c.GetDesign()
	return d == CardDesignHeart || d == CardDesignDiamond
}

// trenteEtQuaranteBetWins はベット種別・勝ち列・最初の札の色から勝敗を判定する。
// winningRow は Noir/Rouge のいずれか (プッシュ/Refait では呼ばない)。
func trenteEtQuaranteBetWins(bet TrenteEtQuaranteBet, winningRow int, firstCardRed bool) bool {
	// 勝ち列の色: Rouge 列 (=1) が赤、Noir 列 (=0) が黒。
	winColorRed := winningRow == TrenteEtQuaranteRowRouge
	switch bet {
	case TrenteEtQuaranteBetNoir:
		return winningRow == TrenteEtQuaranteRowNoir
	case TrenteEtQuaranteBetRouge:
		return winningRow == TrenteEtQuaranteRowRouge
	case TrenteEtQuaranteBetCouleur:
		return firstCardRed == winColorRed
	case TrenteEtQuaranteBetInverse:
		return firstCardRed != winColorRed
	default:
		return false
	}
}

// trenteEtQuaranteRowName は勝ち列の名称を返す。
func trenteEtQuaranteRowName(row int) string {
	switch row {
	case TrenteEtQuaranteRowNoir:
		return "Noir"
	case TrenteEtQuaranteRowRouge:
		return "Rouge"
	default:
		return "None"
	}
}

// trenteEtQuaranteResultName は結果の名称を返す。
func trenteEtQuaranteResultName(r TrenteEtQuaranteResult) string {
	switch r {
	case TrenteEtQuaranteResultWin:
		return "Win"
	case TrenteEtQuaranteResultLose:
		return "Lose"
	default:
		return "Push"
	}
}

// --- Hint ---

// GetHint はベット受付中に教育的な助言を返す。バンキングゲームのため戦略的優劣はなく、
// すべてのベットはイーブンマネーだが Refait (31 の同点) が胴元のエッジになる点を示す。
func (g *TrenteEtQuarante) GetHint() *TrenteEtQuaranteHint {
	if g.state.phase != TrenteEtQuarantePhaseBet {
		return nil
	}
	return &TrenteEtQuaranteHint{
		Bet:    g.config.DefaultBet,
		Reason: "even_odds",
	}
}

// --- ヘルパー ---

func (g *TrenteEtQuarante) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.state.actionLog = append(g.state.actionLog, &ActionLogEntry{
		TurnNumber: len(g.state.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- 状態アクセサ ---

// ResolveRowsForTest は所定の列を注入してラウンドを解決する (テスト用)。乱数配札を
// 迂回して勝敗解決・Refait ロジックを決定的に検証するためのショートカット。
func (g *TrenteEtQuarante) ResolveRowsForTest(bet TrenteEtQuaranteBet, stake int, noir, rouge []*Card) {
	g.state.currentBet = bet
	g.state.stake = stake
	g.state.noirRow = noir
	g.state.rougeRow = rouge
	g.state.noirTotal = 0
	for _, c := range noir {
		g.state.noirTotal += trenteEtQuaranteCardValue(c)
	}
	g.state.rougeTotal = 0
	for _, c := range rouge {
		g.state.rougeTotal += trenteEtQuaranteCardValue(c)
	}
	if len(noir) > 0 {
		g.state.firstCardRed = trenteEtQuaranteIsRed(noir[0])
	}
	g.resolve()
}

// GetPhase は現在のフェーズを返す。
func (g *TrenteEtQuarante) GetPhase() TrenteEtQuarantePhase { return g.state.phase }

// SetPhase はフェーズを設定する (テスト用)。
func (g *TrenteEtQuarante) SetPhase(p TrenteEtQuarantePhase) { g.state.phase = p }

// GetGameEndFlag はラウンド終了フラグを返す。
func (g *TrenteEtQuarante) GetGameEndFlag() bool { return g.state.gameEndFlag }

// GetCurrentBet は現在のベット種別を返す。
func (g *TrenteEtQuarante) GetCurrentBet() TrenteEtQuaranteBet { return g.state.currentBet }

// GetStake は現在のステークを返す。
func (g *TrenteEtQuarante) GetStake() int { return g.state.stake }

// GetNoirRow は黒 (Noir) 列の札を返す。
func (g *TrenteEtQuarante) GetNoirRow() []*Card { return g.state.noirRow }

// GetRougeRow は赤 (Rouge) 列の札を返す。
func (g *TrenteEtQuarante) GetRougeRow() []*Card { return g.state.rougeRow }

// GetNoirTotal は黒 (Noir) 列の合計を返す。
func (g *TrenteEtQuarante) GetNoirTotal() int { return g.state.noirTotal }

// GetRougeTotal は赤 (Rouge) 列の合計を返す。
func (g *TrenteEtQuarante) GetRougeTotal() int { return g.state.rougeTotal }

// GetWinningRow は勝ち列 (None/Noir/Rouge) を返す。
func (g *TrenteEtQuarante) GetWinningRow() int { return g.state.winningRow }

// GetFirstCardRed は最初に配られた札が赤かどうかを返す。
func (g *TrenteEtQuarante) GetFirstCardRed() bool { return g.state.firstCardRed }

// GetRefait は Refait (31 の同点) だったかどうかを返す。
func (g *TrenteEtQuarante) GetRefait() bool { return g.state.refait }

// GetResult は現在のベットに対する勝敗結果を返す。
func (g *TrenteEtQuarante) GetResult() TrenteEtQuaranteResult { return g.state.result }

// GetPayout はこのラウンドでチップに戻された総額を返す。
func (g *TrenteEtQuarante) GetPayout() int { return g.state.payout }

// GetChips は保有チップ数を返す。
func (g *TrenteEtQuarante) GetChips() int {
	if g.player == nil {
		return 0
	}
	return g.player.GetChips()
}

// SetChips は保有チップ数を設定する (テスト用)。
func (g *TrenteEtQuarante) SetChips(chips int) {
	if g.player != nil {
		g.player.SetChips(chips)
	}
}

// GetRoundNumber はこれまでに解決したラウンド数を返す。
func (g *TrenteEtQuarante) GetRoundNumber() int { return g.state.roundNumber }

// GetRemainingDeck はシューの残り枚数を返す。
func (g *TrenteEtQuarante) GetRemainingDeck() int {
	if g.trumpCards == nil {
		return 0
	}
	return g.trumpCards.GetRemainingCount()
}

// GetPlayer はプレイヤーを返す。
func (g *TrenteEtQuarante) GetPlayer() *TrenteEtQuarantePlayer { return g.player }

// GetConfig はローカルルール設定を返す。
func (g *TrenteEtQuarante) GetConfig() TrenteEtQuaranteConfig { return g.config }

// SetConfig はローカルルール設定を変更する。
func (g *TrenteEtQuarante) SetConfig(cfg TrenteEtQuaranteConfig) { g.config = cfg }

// GetActionLog は棋譜を返す。
func (g *TrenteEtQuarante) GetActionLog() []*ActionLogEntry { return g.state.actionLog }

// --- JSON Serialization ---

// trenteEtQuaranteJSON is the JSON wire format for TrenteEtQuarante.
type trenteEtQuaranteJSON struct {
	TrumpCards   *TrumpCards             `json:"tc"`
	Player       *TrenteEtQuarantePlayer `json:"pl"`
	Config       TrenteEtQuaranteConfig  `json:"cf"`
	Phase        TrenteEtQuarantePhase   `json:"ph"`
	CurrentBet   TrenteEtQuaranteBet     `json:"cb"`
	Stake        int                     `json:"st"`
	NoirRow      []*Card                 `json:"nr"`
	RougeRow     []*Card                 `json:"rr"`
	NoirTotal    int                     `json:"nt"`
	RougeTotal   int                     `json:"rt"`
	WinningRow   int                     `json:"wr"`
	FirstCardRed bool                    `json:"fc"`
	Refait       bool                    `json:"rf"`
	Result       TrenteEtQuaranteResult  `json:"re"`
	Payout       int                     `json:"po"`
	RoundNumber  int                     `json:"rn"`
	GameEndFlag  bool                    `json:"ge"`
	Scored       bool                    `json:"sc"`
	ActionLog    []*ActionLogEntry       `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *TrenteEtQuarante) MarshalJSON() ([]byte, error) {
	return json.Marshal(trenteEtQuaranteJSON{
		TrumpCards:   g.trumpCards,
		Player:       g.player,
		Config:       g.config,
		Phase:        g.state.phase,
		CurrentBet:   g.state.currentBet,
		Stake:        g.state.stake,
		NoirRow:      g.state.noirRow,
		RougeRow:     g.state.rougeRow,
		NoirTotal:    g.state.noirTotal,
		RougeTotal:   g.state.rougeTotal,
		WinningRow:   g.state.winningRow,
		FirstCardRed: g.state.firstCardRed,
		Refait:       g.state.refait,
		Result:       g.state.result,
		Payout:       g.state.payout,
		RoundNumber:  g.state.roundNumber,
		GameEndFlag:  g.state.gameEndFlag,
		Scored:       g.state.scored,
		ActionLog:    g.state.actionLog,
	})
}

// trenteEtQuaranteValidPhase は有効なフェーズかどうか。
func trenteEtQuaranteValidPhase(p TrenteEtQuarantePhase) bool {
	return p == TrenteEtQuarantePhaseBet || p == TrenteEtQuarantePhaseResult
}

// trenteEtQuaranteValidateCards は復元したカードスライスに nil が無いか検証する。
func trenteEtQuaranteValidateCards(cards []*Card) error {
	for _, c := range cards {
		if c == nil {
			return fmt.Errorf("trenteetquarante: nil card in state")
		}
	}
	return nil
}

// UnmarshalJSON implements json.Unmarshaler. 不正な永続化データを拒否するための
// バリデーションを行う。
func (g *TrenteEtQuarante) UnmarshalJSON(data []byte) error {
	var j trenteEtQuaranteJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.NoirRow) > TrenteEtQuaranteMaxRowLen || len(j.RougeRow) > TrenteEtQuaranteMaxRowLen ||
		len(j.ActionLog) > trenteEtQuaranteMaxSliceLen {
		return fmt.Errorf("trenteetquarante: input array exceeds maximum allowed size")
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("trenteetquarante: invalid config: %w", err)
	}
	if !trenteEtQuaranteValidPhase(j.Phase) {
		return fmt.Errorf("trenteetquarante: invalid phase %d", j.Phase)
	}
	if !TrenteEtQuaranteBetValid(j.CurrentBet) {
		return fmt.Errorf("trenteetquarante: invalid bet %d", j.CurrentBet)
	}
	if j.Stake < 0 || j.Stake > TrenteEtQuaranteMaxBet {
		return fmt.Errorf("trenteetquarante: stake out of range")
	}
	if j.NoirTotal < 0 || j.RougeTotal < 0 || j.Payout < 0 || j.RoundNumber < 0 {
		return fmt.Errorf("trenteetquarante: negative numeric state")
	}
	if j.WinningRow < TrenteEtQuaranteRowNone || j.WinningRow > TrenteEtQuaranteRowRouge {
		return fmt.Errorf("trenteetquarante: winning row out of range")
	}
	if j.Result < TrenteEtQuaranteResultLose || j.Result > TrenteEtQuaranteResultWin {
		return fmt.Errorf("trenteetquarante: result out of range")
	}
	if err := trenteEtQuaranteValidateCards(j.NoirRow); err != nil {
		return err
	}
	if err := trenteEtQuaranteValidateCards(j.RougeRow); err != nil {
		return err
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = newTrenteEtQuaranteShoe()
	}
	g.player = j.Player
	if g.player == nil {
		g.player = NewTrenteEtQuarantePlayer(TrenteEtQuaranteDefaultChips)
	}
	g.config = j.Config
	g.state = trenteEtQuaranteState{
		phase:        j.Phase,
		currentBet:   j.CurrentBet,
		stake:        j.Stake,
		noirRow:      j.NoirRow,
		rougeRow:     j.RougeRow,
		noirTotal:    j.NoirTotal,
		rougeTotal:   j.RougeTotal,
		winningRow:   j.WinningRow,
		firstCardRed: j.FirstCardRed,
		refait:       j.Refait,
		result:       j.Result,
		payout:       j.Payout,
		roundNumber:  j.RoundNumber,
		gameEndFlag:  j.GameEndFlag,
		scored:       j.Scored,
		actionLog:    j.ActionLog,
	}
	if g.state.noirRow == nil {
		g.state.noirRow = make([]*Card, 0)
	}
	if g.state.rougeRow == nil {
		g.state.rougeRow = make([]*Card, 0)
	}
	if g.state.actionLog == nil {
		g.state.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
