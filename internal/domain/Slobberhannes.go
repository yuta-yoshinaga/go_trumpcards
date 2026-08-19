//go:build !js || !wasm || classic

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// SlobberhannesPhase スロバーハンネスのゲームフェーズ
type SlobberhannesPhase int

// Slobberhannesのフェーズ定数
const (
	// SlobberhannesPhasePlay プレイ中
	SlobberhannesPhasePlay SlobberhannesPhase = iota
	// SlobberhannesPhaseRoundEnd ラウンド終了（得点確定、次ラウンド待ち）
	SlobberhannesPhaseRoundEnd
	// SlobberhannesPhaseGameEnd ゲーム終了
	SlobberhannesPhaseGameEnd
)

// SlobberhannesPlayerCnt プレイヤー数（4 人固定）
const SlobberhannesPlayerCnt = 4

// SlobberhannesHandSize 各プレイヤーの手札枚数
const SlobberhannesHandSize = 8

// SlobberhannesTricksPerRound 1 ラウンドのトリック数
const SlobberhannesTricksPerRound = SlobberhannesHandSize

// SlobberhannesPenalty 罰点。最初のトリック・最後のトリック・♣Q の 3 つ。
const SlobberhannesPenalty = -1

// SlobberhannesCleanBonus 3 つとも回避したときのボーナス
const SlobberhannesCleanBonus = 1

// SlobberhannesQueenSuit 罰点になるクイーンのスート（クラブ）
const SlobberhannesQueenSuit = CardDesignClover

// SlobberhannesQueenValue クイーンの値
const SlobberhannesQueenValue = 12

// slobberhannesMaxSliceLen caps slice sizes during deserialisation.
const slobberhannesMaxSliceLen = 1000

// Slobberhannes スロバーハンネス ゲームクラス。
//
// ドイツ・低地地方の**回避型**トリックテイキング。ピケット 32 枚
// (A,7,8,9,10,J,Q,K × 4 スート) を 4 人に 8 枚ずつ配り、8 トリックを戦う。
//
// **切り札は無い。** リードされたスートの最強札がそのトリックを取る。issue
// #5233 は「切り札の有無を規定した上で」と判断を委ねているが、実際の
// スロバーハンネスに切り札は無いのでそれに従う。切り札を足すと、回避したい
// トリックを確実に押し付けられてしまい、ゲームの緊張が消える。
//
// 罰点は 3 つで、いずれも -1:
//
//   - **最初の**トリックを取る
//   - **最後の**トリックを取る
//   - **♣Q を含む**トリックを取る
//
// 3 つとも回避したプレイヤーには +1。つまり 1 ラウンドの取りうる得点は
// +1 / 0 / -1 / -2 / -3 で、**合計が最も大きいプレイヤーが勝つ**。
//
// issue #5233 は「合計得点が最も少ない（＝失点が少ない）プレイヤーが勝利」と
// 書いているが、これは矛盾している: 罰が -1 でボーナスが +1 なら、合計が
// 大きいほうが優秀になる。罰点を正の「失点」として数える流儀と、符号つきで
// 数える流儀が混ざったもの。ここでは issue が明示している -1 / +1 の符号を
// 採用し、**最大値が勝ち**とする。
//
// Hearts との違いは、罰の対象が**カードだけでなくトリックの位置**にある点。
// 「取ってはいけない札」ではなく「取ってはいけないタイミング」があるので、
// 最初と最後の 2 トリックだけは持ち札の強弱と無関係に危険になる。
type Slobberhannes struct {
	trumpCards *TrumpCards
	players    []*SlobberhannesPlayer
	config     SlobberhannesConfig

	phase       SlobberhannesPhase
	roundNumber int
	trickNumber int
	// currentTrick 現在のトリックに出された札
	currentTrick     []*TrickCard
	currentPlayerIdx int
	leadPlayerIdx    int
	dealerIdx        int

	gameEndFlag bool
	winnerIdx   int

	actionLogBase
}

// NewSlobberhannes コンストラクタ
func NewSlobberhannes(trumpCards *TrumpCards, players []*SlobberhannesPlayer, config SlobberhannesConfig) *Slobberhannes {
	return &Slobberhannes{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		winnerIdx:  -1,
	}
}

// NewDefaultSlobberhannes 既定構成（人間 1 + CPU 3）のコンストラクタ
func NewDefaultSlobberhannes() *Slobberhannes {
	players := make([]*SlobberhannesPlayer, 0, SlobberhannesPlayerCnt)
	for i := range SlobberhannesPlayerCnt {
		players = append(players, NewSlobberhannesPlayer(i == 0))
	}
	return NewSlobberhannes(NewTrumpCardsBelote(), players, DefaultSlobberhannesConfig())
}

// Reset ゲーム全体を初期化する
func (s *Slobberhannes) Reset() {
	s.phase = SlobberhannesPhasePlay
	s.roundNumber = 1
	s.dealerIdx = 0
	s.gameEndFlag = false
	s.winnerIdx = -1
	s.actionLog = nil
	for _, p := range s.players {
		p.ResetGame()
	}
	s.dealRound()
}

// dealRound 1 ラウンド分を配る。手番はディーラーの左隣から。
func (s *Slobberhannes) dealRound() {
	s.trickNumber = 0
	s.currentTrick = nil
	for _, p := range s.players {
		p.ResetRound()
	}

	// **ピケット 32 枚を使う。** Skat の newSkatDeck は 32 枚だが中身が
	// ♠1-13/♣1-13/♥1-6/♦なし で、デッキとして成立していない (#5296)。
	s.trumpCards = NewTrumpCardsBelote()
	s.trumpCards.Shuffle()
	for range SlobberhannesHandSize {
		for i := range SlobberhannesPlayerCnt {
			idx := (s.dealerIdx + 1 + i) % SlobberhannesPlayerCnt
			if c := s.trumpCards.DrawCard(); c != nil {
				s.players[idx].AddCard(c)
			}
		}
	}
	s.leadPlayerIdx = (s.dealerIdx + 1) % SlobberhannesPlayerCnt
	s.currentPlayerIdx = s.leadPlayerIdx
	s.sortAllHands()
	s.appendLog(-1, "deal", fmt.Sprintf("ラウンド%d を開始", s.roundNumber), nil)
}

// sortAllHands 手札をスートごと・強さ順に並べる
func (s *Slobberhannes) sortAllHands() {
	for _, p := range s.players {
		s.sortHand(p)
	}
}

// sortHand 1 人分の手札を並べ替える
func (s *Slobberhannes) sortHand(p *SlobberhannesPlayer) {
	sortPlayerHand(p, func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return slobberhannesRank(ci) < slobberhannesRank(cj)
	})
}

// PlayerPlay 人間プレイヤーが手札の cardIndex を出す
func (s *Slobberhannes) PlayerPlay(cardIndex int) error {
	if s.gameEndFlag {
		return errors.New("game has ended")
	}
	if s.phase != SlobberhannesPhasePlay {
		return errors.New("round has ended")
	}
	if s.currentPlayerIdx != 0 {
		return errors.New("not your turn")
	}
	return s.play(0, cardIndex)
}

// CpuPlay CPU が 1 枚出す
func (s *Slobberhannes) CpuPlay() {
	if s.gameEndFlag || s.phase != SlobberhannesPhasePlay || s.currentPlayerIdx == 0 {
		return
	}
	_ = s.play(s.currentPlayerIdx, s.chooseCpuCard(s.currentPlayerIdx))
}

// play 指定プレイヤーが 1 枚出す
func (s *Slobberhannes) play(playerIdx, cardIndex int) error {
	p := s.players[playerIdx]
	if cardIndex < 0 || cardIndex >= p.GetCardsSize() {
		return fmt.Errorf("invalid card index: %d", cardIndex)
	}
	card := p.GetCard(cardIndex)
	if !s.canPlay(playerIdx, card) {
		return errors.New("must follow suit")
	}
	p.RemoveCard(cardIndex)
	s.currentTrick = append(s.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	s.appendLog(playerIdx, "play", cardStr(card), []*Card{card})

	if len(s.currentTrick) < SlobberhannesPlayerCnt {
		s.currentPlayerIdx = (playerIdx + 1) % SlobberhannesPlayerCnt
		return nil
	}
	s.resolveTrick()
	return nil
}

// canPlay フォロー義務を満たすか。リードなら何でも出せる。
func (s *Slobberhannes) canPlay(playerIdx int, card *Card) bool {
	if len(s.currentTrick) == 0 {
		return true
	}
	leadSuit := s.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return true
	}
	p := s.players[playerIdx]
	for i := range p.GetCardsSize() {
		if p.GetCard(i).GetDesign() == leadSuit {
			return false
		}
	}
	return true
}

// GetValidPlayIndices 出せる手札のインデックスを返す
func (s *Slobberhannes) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(s.players) {
		return nil
	}
	p := s.players[playerIdx]
	valid := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		if s.canPlay(playerIdx, p.GetCard(i)) {
			valid = append(valid, i)
		}
	}
	return valid
}

// resolveTrick トリックを解決し、罰点の対象なら記録する
func (s *Slobberhannes) resolveTrick() {
	winner := s.trickWinner()
	cards := make([]*Card, 0, len(s.currentTrick))
	hasQueen := false
	for _, tc := range s.currentTrick {
		cards = append(cards, tc.Card)
		if SlobberhannesIsPenaltyQueen(tc.Card) {
			hasQueen = true
		}
	}
	s.players[winner].AddTrick(cards)

	// **位置による罰**。最初と最後のトリックは、中身に関係なく危険。
	if s.trickNumber == 0 {
		s.players[winner].tookFirstTrick = true
		s.appendLog(winner, "penalty", "最初のトリックを取った", nil)
	}
	if s.trickNumber == SlobberhannesTricksPerRound-1 {
		s.players[winner].tookLastTrick = true
		s.appendLog(winner, "penalty", "最後のトリックを取った", nil)
	}
	if hasQueen {
		s.players[winner].tookQueen = true
		s.appendLog(winner, "penalty", "クラブのクイーンを取った", nil)
	}

	s.trickNumber++
	s.currentTrick = nil
	s.leadPlayerIdx = winner
	s.currentPlayerIdx = winner

	if s.trickNumber >= SlobberhannesTricksPerRound {
		s.finishRound()
	}
}

// finishRound ラウンドの得点を確定させる
func (s *Slobberhannes) finishRound() {
	for i, p := range s.players {
		n := p.PenaltyCount()
		if n == 0 {
			p.AddScore(SlobberhannesCleanBonus)
			s.appendLog(i, "score", "全回避ボーナス +1", nil)
			continue
		}
		p.AddScore(SlobberhannesPenalty * n)
		s.appendLog(i, "score", fmt.Sprintf("罰点 %d", SlobberhannesPenalty*n), nil)
	}

	if s.roundNumber >= s.config.Rounds {
		s.finishGame()
		return
	}
	s.phase = SlobberhannesPhaseRoundEnd
}

// NextRound 次のラウンドを開始する
func (s *Slobberhannes) NextRound() {
	if s.gameEndFlag || s.phase != SlobberhannesPhaseRoundEnd {
		return
	}
	s.roundNumber++
	s.dealerIdx = (s.dealerIdx + 1) % SlobberhannesPlayerCnt
	s.phase = SlobberhannesPhasePlay
	s.dealRound()
}

// finishGame 勝者を決めて終了する
func (s *Slobberhannes) finishGame() {
	s.phase = SlobberhannesPhaseGameEnd
	s.gameEndFlag = true

	// **合計が最大のプレイヤーが勝ち。** 罰が負・ボーナスが正なので、
	// 「失点が少ない」と「得点が大きい」は同じことを指す。
	best := s.players[0].GetScore()
	bestIdx := 0
	tied := false
	for i := 1; i < len(s.players); i++ {
		switch score := s.players[i].GetScore(); {
		case score > best:
			best, bestIdx, tied = score, i, false
		case score == best:
			tied = true
		}
	}
	if tied {
		s.winnerIdx = -1
		s.appendLog(-1, "result", "同点で決着つかず", nil)
		return
	}
	s.winnerIdx = bestIdx
	s.appendLog(bestIdx, "result", fmt.Sprintf("勝者（%d点）", best), nil)
}

// trickWinner 現在のトリックの勝者。切り札が無いので、リードのスートの最強札。
func (s *Slobberhannes) trickWinner() int {
	if len(s.currentTrick) == 0 {
		return s.leadPlayerIdx
	}
	leadSuit := s.currentTrick[0].Card.GetDesign()
	bestIdx := s.currentTrick[0].PlayerIdx
	best := s.currentTrick[0].Card
	for _, tc := range s.currentTrick[1:] {
		if tc.Card.GetDesign() != leadSuit {
			continue
		}
		if slobberhannesRank(tc.Card) > slobberhannesRank(best) {
			best, bestIdx = tc.Card, tc.PlayerIdx
		}
	}
	return bestIdx
}

// slobberhannesRank 札の強さ。A が最強、7 が最弱。
func slobberhannesRank(c *Card) int {
	if c == nil {
		return 0
	}
	if v := c.GetValue(); v == 1 {
		return CardValueMax + 1 // A は K より強い
	}
	return c.GetValue()
}

// SlobberhannesIsPenaltyQueen は罰点札 (♣Q) かを返す。
//
// **位置ではなく中身に効く唯一の罰点。**最初/最後のトリックは番号で警告
// できるが、♣Q は「今場に出ているか」がリスクの本体なので、画面側にも
// 同じ判定が要る (#5745)。
func SlobberhannesIsPenaltyQueen(c *Card) bool {
	return c != nil && c.GetDesign() == SlobberhannesQueenSuit && c.GetValue() == SlobberhannesQueenValue
}

// chooseCpuCard CPU の手を選ぶ。**取りたくない**ゲームなので、基本は
// 「そのトリックを取らない一番高い札」を探し、無ければ最弱を捨てる。
func (s *Slobberhannes) chooseCpuCard(playerIdx int) int {
	valid := s.GetValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	p := s.players[playerIdx]

	// リードのとき: ♣Q を持っていれば、それを引きずり出される前に安全に
	// 処分したいが、リードで出すと自分が取る可能性が高い。ここでは最弱を
	// リードして様子を見る。
	if len(s.currentTrick) == 0 {
		return s.pickLowest(p, valid)
	}

	loseIdx, loseRank := -1, -1
	minIdx, minRank := valid[0], slobberhannesRank(p.GetCard(valid[0]))
	leadSuit := s.currentTrick[0].Card.GetDesign()
	bestSoFar := s.currentBestRank()

	for _, i := range valid {
		c := p.GetCard(i)
		r := slobberhannesRank(c)
		if r < minRank {
			minIdx, minRank = i, r
		}
		// 「取らずに済む」札のうち、一番高いものを捨てるのが得。
		followsLead := c.GetDesign() == leadSuit
		if (!followsLead || r < bestSoFar) && r > loseRank {
			loseIdx, loseRank = i, r
		}
	}

	// ♣Q を持っていて、確実に取らずに捨てられるなら今が好機。
	if qIdx, ok := s.findPenaltyQueen(p, valid); ok && !s.wouldWin(p.GetCard(qIdx)) {
		return qIdx
	}
	// 罰点対象でなくても取りたくないゲームなので、取らずに済むなら常にそうする。
	if loseIdx >= 0 {
		return loseIdx
	}
	return minIdx
}

// currentBestRank 現在のトリックでリードのスートの最強ランク
func (s *Slobberhannes) currentBestRank() int {
	if len(s.currentTrick) == 0 {
		return -1
	}
	leadSuit := s.currentTrick[0].Card.GetDesign()
	best := -1
	for _, tc := range s.currentTrick {
		if tc.Card.GetDesign() == leadSuit && slobberhannesRank(tc.Card) > best {
			best = slobberhannesRank(tc.Card)
		}
	}
	return best
}

// wouldWin その札を今出したらトリックを取ってしまうか
func (s *Slobberhannes) wouldWin(c *Card) bool {
	if c == nil || len(s.currentTrick) == 0 {
		return true
	}
	if c.GetDesign() != s.currentTrick[0].Card.GetDesign() {
		return false
	}
	return slobberhannesRank(c) > s.currentBestRank()
}

// findPenaltyQueen 合法手のなかの ♣Q を探す
func (s *Slobberhannes) findPenaltyQueen(p *SlobberhannesPlayer, valid []int) (int, bool) {
	for _, i := range valid {
		if SlobberhannesIsPenaltyQueen(p.GetCard(i)) {
			return i, true
		}
	}
	return 0, false
}

// pickLowest 合法手のなかで一番弱い札
func (s *Slobberhannes) pickLowest(p *SlobberhannesPlayer, valid []int) int {
	bestIdx, bestRank := valid[0], slobberhannesRank(p.GetCard(valid[0]))
	for _, i := range valid[1:] {
		if r := slobberhannesRank(p.GetCard(i)); r < bestRank {
			bestIdx, bestRank = i, r
		}
	}
	return bestIdx
}

// --- Getters ---

// GetPhase 現在のフェーズ
func (s *Slobberhannes) GetPhase() SlobberhannesPhase { return s.phase }

// GetConfig 現在の設定
func (s *Slobberhannes) GetConfig() SlobberhannesConfig { return s.config }

// SetConfig 設定を差し替える
func (s *Slobberhannes) SetConfig(c SlobberhannesConfig) { s.config = c }

// GetRoundNumber 現在のラウンド番号（1 起点）
func (s *Slobberhannes) GetRoundNumber() int { return s.roundNumber }

// GetTrickNumber 現在のトリック番号（0 起点）
func (s *Slobberhannes) GetTrickNumber() int { return s.trickNumber }

// GetCurrentTrick 現在のトリック
func (s *Slobberhannes) GetCurrentTrick() []*TrickCard { return s.currentTrick }

// GetCurrentPlayerIdx 現在の手番
func (s *Slobberhannes) GetCurrentPlayerIdx() int { return s.currentPlayerIdx }

// GetLeadPlayerIdx リードプレイヤー
func (s *Slobberhannes) GetLeadPlayerIdx() int { return s.leadPlayerIdx }

// GetDealerIdx ディーラー
func (s *Slobberhannes) GetDealerIdx() int { return s.dealerIdx }

// GetPlayerCnt プレイヤー数
func (s *Slobberhannes) GetPlayerCnt() int { return len(s.players) }

// GetPlayer 指定インデックスのプレイヤー
func (s *Slobberhannes) GetPlayer(i int) *SlobberhannesPlayer {
	if i < 0 || i >= len(s.players) {
		return nil
	}
	return s.players[i]
}

// GetGameEndFlag ゲーム終了フラグ
func (s *Slobberhannes) GetGameEndFlag() bool { return s.gameEndFlag }

// GetWinnerIdx 勝者（-1: 未確定または同点）
func (s *Slobberhannes) GetWinnerIdx() int { return s.winnerIdx }

// IsHumanTurn 人間の手番か
func (s *Slobberhannes) IsHumanTurn() bool {
	return !s.gameEndFlag && s.phase == SlobberhannesPhasePlay && s.currentPlayerIdx == 0
}

// SlobberhannesHint ヒント情報
type SlobberhannesHint struct {
	// CardIndex 推奨する手札のインデックス
	CardIndex *int
	// Reason ヒント理由キー
	Reason string
}

// GetHint 人間プレイヤーへの推奨手を返す。手番でなければ nil。
//
// 回避型なので狙いは常に「取らないこと」だが、**そのトリックが罰点対象か
// どうかで切迫度が変わる**。最初・最後・♣Q 入りのトリックは何としても
// 避ける必要があり、それ以外は高い札を安全に処分する好機になる。
func (s *Slobberhannes) GetHint() *SlobberhannesHint {
	if !s.IsHumanTurn() || s.players[0].GetCardsSize() == 0 {
		return nil
	}
	idx := s.chooseCpuCard(0)
	return &SlobberhannesHint{CardIndex: &idx, Reason: s.hintReason()}
}

// hintReason 現在の狙いを表す理由キーを返す
func (s *Slobberhannes) hintReason() string {
	if len(s.currentTrick) == 0 {
		return "slobberhannesLeadLow"
	}
	if s.trickIsDangerous() {
		return "slobberhannesAvoid"
	}
	return "slobberhannesDump"
}

// trickIsDangerous 現在のトリックが罰点対象か（最初・最後・♣Q 入り）
func (s *Slobberhannes) trickIsDangerous() bool {
	if s.trickNumber == 0 || s.trickNumber == SlobberhannesTricksPerRound-1 {
		return true
	}
	for _, tc := range s.currentTrick {
		if SlobberhannesIsPenaltyQueen(tc.Card) {
			return true
		}
	}
	return false
}

// GiveUp 投了する
func (s *Slobberhannes) GiveUp() {
	if s.gameEndFlag {
		return
	}
	s.phase = SlobberhannesPhaseGameEnd
	s.gameEndFlag = true
	s.winnerIdx = -1
	s.appendLog(0, "giveup", "ギブアップしました", nil)
}

// appendLog 棋譜エントリを追加
func (s *Slobberhannes) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	s.appendLogAt(s.trickNumber, playerIdx, actionType, detail, cards)
}

// slobberhannesJSON is the KV snapshot format for Slobberhannes.
type slobberhannesJSON struct {
	TrumpCards       *TrumpCards            `json:"tc"`
	Players          []*SlobberhannesPlayer `json:"pl"`
	Config           SlobberhannesConfig    `json:"cf"`
	Phase            SlobberhannesPhase     `json:"ph"`
	RoundNumber      int                    `json:"rn"`
	TrickNumber      int                    `json:"tn"`
	CurrentTrick     []*TrickCard           `json:"ct"`
	CurrentPlayerIdx int                    `json:"cp"`
	LeadPlayerIdx    int                    `json:"lp"`
	DealerIdx        int                    `json:"di"`
	GameEndFlag      bool                   `json:"ge"`
	WinnerIdx        int                    `json:"wi"`
	ActionLog        []*ActionLogEntry      `json:"al"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (s *Slobberhannes) MarshalJSON() ([]byte, error) {
	return json.Marshal(&slobberhannesJSON{
		TrumpCards:       s.trumpCards,
		Players:          s.players,
		Config:           s.config,
		Phase:            s.phase,
		RoundNumber:      s.roundNumber,
		TrickNumber:      s.trickNumber,
		CurrentTrick:     s.currentTrick,
		CurrentPlayerIdx: s.currentPlayerIdx,
		LeadPlayerIdx:    s.leadPlayerIdx,
		DealerIdx:        s.dealerIdx,
		GameEndFlag:      s.gameEndFlag,
		WinnerIdx:        s.winnerIdx,
		ActionLog:        s.actionLog,
	})
}

// UnmarshalJSON KV スナップショットからの復元。KV には以前のバージョンが書いた
// 任意のバイト列が入りうるので、壊れた状態でゲームを開始させないよう値域を検証する。
func (s *Slobberhannes) UnmarshalJSON(data []byte) error {
	var j slobberhannesJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Phase < SlobberhannesPhasePlay || j.Phase > SlobberhannesPhaseGameEnd {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.TrickNumber < 0 || j.TrickNumber > SlobberhannesTricksPerRound {
		return fmt.Errorf("invalid trick number: %d", j.TrickNumber)
	}
	if j.RoundNumber < 1 || j.RoundNumber > SlobberhannesRoundsMax {
		return fmt.Errorf("invalid round number: %d", j.RoundNumber)
	}
	if len(j.ActionLog) > slobberhannesMaxSliceLen {
		return errors.New("slobberhannes: input array exceeds maximum allowed size")
	}
	if len(j.CurrentTrick) > SlobberhannesPlayerCnt {
		return fmt.Errorf("current trick holds %d cards", len(j.CurrentTrick))
	}
	for name, idx := range map[string]int{
		"current player": j.CurrentPlayerIdx,
		"lead player":    j.LeadPlayerIdx,
		"dealer":         j.DealerIdx,
	} {
		if idx < 0 || idx >= SlobberhannesPlayerCnt {
			return fmt.Errorf("invalid %s: %d", name, idx)
		}
	}
	if j.WinnerIdx < -1 || j.WinnerIdx >= SlobberhannesPlayerCnt {
		return fmt.Errorf("invalid winner: %d", j.WinnerIdx)
	}
	if j.TrumpCards != nil {
		s.trumpCards = j.TrumpCards
	}
	if len(j.Players) == SlobberhannesPlayerCnt {
		s.players = j.Players
	}
	s.config = j.Config
	s.phase = j.Phase
	s.roundNumber = j.RoundNumber
	s.trickNumber = j.TrickNumber
	s.currentTrick = j.CurrentTrick
	s.currentPlayerIdx = j.CurrentPlayerIdx
	s.leadPlayerIdx = j.LeadPlayerIdx
	s.dealerIdx = j.DealerIdx
	s.gameEndFlag = j.GameEndFlag
	s.winnerIdx = j.WinnerIdx
	s.actionLog = j.ActionLog
	return nil
}
