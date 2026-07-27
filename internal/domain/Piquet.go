//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sort"
)

// PiquetPlayerCnt Piquetは2人用ゲーム
const PiquetPlayerCnt = 2

// PiquetHandSize 各プレイヤーの初期手札枚数
const PiquetHandSize = 12

// PiquetTalonSize タロン (山札残り) の枚数 = 8
const PiquetTalonSize = 8

// PiquetElderTalonSize Elder が交換可能な上位5枚
const PiquetElderTalonSize = 5

// PiquetYoungerTalonSize Younger が交換可能な下位3枚
const PiquetYoungerTalonSize = 3

// PiquetTricksPerRound 1ラウンドのトリック数
const PiquetTricksPerRound = 12

// ボーナス点
const (
	// PiquetCarteBlancheBonus 絵札なし宣言ボーナス
	PiquetCarteBlancheBonus = 10
	// PiquetRepiqueBonus 宣言だけで30点に達したときのボーナス
	PiquetRepiqueBonus = 60
	// PiquetPiqueBonus Elderが宣言+プレイで30点に達したときのボーナス
	PiquetPiqueBonus = 30
	// PiquetCardsBonus トリックの過半数獲得ボーナス
	PiquetCardsBonus = 10
	// PiquetCapotBonus 全トリック獲得ボーナス (cardsの代わり)
	PiquetCapotBonus = 40
	// PiquetTrickLeadPoint リードした側がそのトリックを獲得した場合の得点
	PiquetTrickLeadPoint = 1
	// PiquetTrickWonPoint リードしなかった側がトリックを獲得した場合の得点
	PiquetTrickWonPoint = 1
	// PiquetLastTrickBonus 最終トリックボーナス
	PiquetLastTrickBonus = 1
	// PiquetRubiconScore 100点ルビコン閾値
	PiquetRubiconScore = 100
	// PiquetWinnerBonus 通常のパルティ勝者ボーナス
	PiquetWinnerBonus = 100
)

// 宣言種別の場合分け
const (
	// PiquetMinSequenceLength シーケンスとして有効な最小枚数
	PiquetMinSequenceLength = 3
	// PiquetMinSetLength セットとして有効な最小枚数
	PiquetMinSetLength = 3
	// PiquetSetMinRank セット対象は10以上 (10,J,Q,K,A)
	PiquetSetMinRank = 10
)

// PiquetPhase ゲームフェーズ
type PiquetPhase int

// Piquetフェーズ定数
const (
	// PiquetPhaseExchange 交換フェーズ (Elder→Youngerの順)
	PiquetPhaseExchange PiquetPhase = iota
	// PiquetPhaseDeclaration 宣言フェーズ (Point→Sequence→Setの順)
	PiquetPhaseDeclaration
	// PiquetPhasePlay トリックプレイフェーズ
	PiquetPhasePlay
	// PiquetPhaseScore ラウンド終了 (次ディール待ち)
	PiquetPhaseScore
	// PiquetPhaseGameEnd パルティ終了
	PiquetPhaseGameEnd
)

// PiquetDeclarationKind 宣言の種類
type PiquetDeclarationKind int

// 宣言種別定数
const (
	// PiquetDeclKindPoint Point宣言 (同一スートのカード枚数)
	PiquetDeclKindPoint PiquetDeclarationKind = iota
	// PiquetDeclKindSequence Sequence宣言 (同一スート連続)
	PiquetDeclKindSequence
	// PiquetDeclKindSet Set宣言 (同位札の三枚/四枚)
	PiquetDeclKindSet
)

// PiquetExchangeTurn 現在の交換手番
type PiquetExchangeTurn int

// 交換手番定数
const (
	// PiquetExchangeTurnElder Elderの交換待ち
	PiquetExchangeTurnElder PiquetExchangeTurn = iota
	// PiquetExchangeTurnYounger Youngerの交換待ち
	PiquetExchangeTurnYounger
	// PiquetExchangeTurnDone 両者の交換完了
	PiquetExchangeTurnDone
)

// PiquetClaim 宣言の中身 (Point/Sequence/Set 共通)
//
//	Length:   Point=枚数、Sequence=連続枚数、Set=3 or 4
//	TopRank:  Sequence=最上位カード値、Set=ランク (10,11,12,13,1)
//	PipTotal: Point の比較に使うピップ合計 (A=11, K/Q/J=10, 10=10, 9=9, ...)
//	Suit:     Point/Sequence のスート
//	Cards:    宣言の根拠となるカード
type PiquetClaim struct {
	Length   int     `json:"l"`
	TopRank  int     `json:"r"`
	PipTotal int     `json:"p"`
	Suit     int     `json:"s"`
	Cards    []*Card `json:"c"`
}

// PiquetDeclarationResult 1宣言の比較結果
//
//	Winner: 0=Elder, 1=Younger, -1=タイ(無得点)
//	Score:  勝者に加算された得点 (タイ時は0)
//	ScoredBy: 得点者 (Winner と同じ、タイなら-1)
//	Sets: Set宣言で勝者が持つ全てのセット (trio/quatorze 加算用)
type PiquetDeclarationResult struct {
	Kind         PiquetDeclarationKind `json:"k"`
	ElderClaim   *PiquetClaim          `json:"ec,omitempty"`
	YoungerClaim *PiquetClaim          `json:"yc,omitempty"`
	Winner       int                   `json:"w"`
	Score        int                   `json:"sc"`
	ScoredBy     int                   `json:"sb"`
	Sets         []*PiquetClaim        `json:"sets,omitempty"`
}

// PiquetHint ヒント情報
type PiquetHint struct {
	// CardIndex プレイ時の推奨カードインデックス
	CardIndex *int
	// DiscardIndices 交換時の推奨破棄インデックス
	DiscardIndices []int
	// Reason 推奨理由 (i18nキー)
	Reason string
}

// Piquet ゲームクラス
type Piquet struct {
	trumpCards  *TrumpCards
	players     [PiquetPlayerCnt]*PiquetPlayer
	config      PiquetConfig
	phase       PiquetPhase
	dealNumber  int
	elderIdx    int
	gameEndFlag bool
	winnerIdx   int

	// 交換フェーズ
	talon                []*Card
	elderExchanged       []*Card // Elder が取った上位5枚 (実際に交換した数だけ)
	youngerExchanged     []*Card // Younger が取った下位3枚 (実際に交換した数だけ)
	elderExchangedCnt    int
	youngerExchangedCnt  int
	exchangeTurn         PiquetExchangeTurn
	elderRevealedTalon   []*Card // Elder が交換しなかった上位5枚 (公開済み)
	youngerRevealedTalon []*Card

	// 宣言フェーズ
	declStage   PiquetDeclarationKind
	declResults []*PiquetDeclarationResult

	// プレイフェーズ
	currentTrick     []*TrickCard
	currentPlayerIdx int
	trickNumber      int
	leadPlayerIdx    int
	tricksWon        [PiquetPlayerCnt]int

	// pique/repique 追跡
	firstScorerIdx       int  // -1までは無得点
	elderReached30InPlay bool // pique 検出用 (Elder専用)

	// ボーナス内訳 (デバッグ用)
	carteBlanche [PiquetPlayerCnt]bool

	// アクションログ
	actionLog []*ActionLogEntry
}

// NewPiquet コンストラクタ
func NewPiquet(trumpCards *TrumpCards, players []*PiquetPlayer, config PiquetConfig) *Piquet {
	p := &Piquet{
		trumpCards: trumpCards,
		config:     config,
		winnerIdx:  -1,
	}
	for i := 0; i < PiquetPlayerCnt && i < len(players); i++ {
		p.players[i] = players[i]
	}
	return p
}

// NewDefaultPiquet 既定の2人プレイヤー (Human + CPU) でコンストラクト。
// 第1ディールの Elder は Human (idx=0)。
func NewDefaultPiquet() *Piquet {
	players := []*PiquetPlayer{
		NewPiquetPlayer(true),
		NewPiquetPlayer(false),
	}
	return NewPiquet(NewTrumpCardsBelote(), players, DefaultPiquetConfig())
}

// Reset パルティを最初から開始する
func (p *Piquet) Reset() {
	p.dealNumber = 1
	p.elderIdx = 0 // 第1ディールの Elder は idx=0 (human)
	p.gameEndFlag = false
	p.winnerIdx = -1
	for _, pl := range p.players {
		pl.ResetMatch()
	}
	p.actionLog = nil
	p.startDeal()
}

// NextDeal 次のディールに進む。PiquetPhaseScore でなければ no-op。
func (p *Piquet) NextDeal() {
	if p.phase != PiquetPhaseScore {
		return
	}
	if p.dealNumber >= p.config.DealsPerPartie {
		p.finishPartie()
		return
	}
	p.dealNumber++
	p.elderIdx = (p.elderIdx + 1) % PiquetPlayerCnt
	p.startDeal()
}

// startDeal 1ディールの初期化と配札・カルトブランシュ自動判定
func (p *Piquet) startDeal() {
	for _, pl := range p.players {
		pl.ResetRound()
	}
	p.elderExchanged = nil
	p.youngerExchanged = nil
	p.elderExchangedCnt = 0
	p.youngerExchangedCnt = 0
	p.elderRevealedTalon = nil
	p.youngerRevealedTalon = nil
	p.exchangeTurn = PiquetExchangeTurnElder
	p.declStage = PiquetDeclKindPoint
	p.declResults = nil
	p.currentTrick = nil
	p.currentPlayerIdx = p.elderIdx
	p.trickNumber = 0
	p.leadPlayerIdx = p.elderIdx
	p.tricksWon = [PiquetPlayerCnt]int{}
	p.firstScorerIdx = -1
	p.elderReached30InPlay = false
	p.carteBlanche = [PiquetPlayerCnt]bool{}

	// 32枚デッキをシャッフル → 12枚ずつ配 → 残り8枚をタロン
	p.trumpCards.Shuffle()
	for range PiquetHandSize {
		for _, pl := range p.players {
			c := p.trumpCards.DrawCard()
			if c != nil {
				pl.AddCard(c)
			}
		}
	}
	p.talon = make([]*Card, 0, PiquetTalonSize)
	for range PiquetTalonSize {
		c := p.trumpCards.DrawCard()
		if c != nil {
			p.talon = append(p.talon, c)
		}
	}

	// カルトブランシュ自動判定 (絵札なし)
	for i := range PiquetPlayerCnt {
		if hasNoCourtCards(p.players[i]) {
			p.carteBlanche[i] = true
			p.players[i].AddDeclScore(PiquetCarteBlancheBonus)
			p.recordFirstScorer(i)
			p.appendLog(i, "carteblanche", "カルトブランシュ +10", nil)
		}
	}

	p.phase = PiquetPhaseExchange
	p.appendLog(p.elderIdx, "deal", fmt.Sprintf("ディール%d開始 (Elder=%d)", p.dealNumber, p.elderIdx), nil)
}

// finishPartie パルティ終了 (ルビコン判定)
func (p *Piquet) finishPartie() {
	p.phase = PiquetPhaseGameEnd
	p.gameEndFlag = true
	// 通算スコア比較
	s0 := p.players[0].GetMatchScore()
	s1 := p.players[1].GetMatchScore()
	switch {
	case s0 > s1:
		p.applyPartieResult(0, s0, s1)
	case s1 > s0:
		p.applyPartieResult(1, s1, s0)
	default:
		// 引き分け
		p.winnerIdx = -1
	}
	p.appendLog(p.winnerIdx, "partie", "パルティ終了", nil)
}

// applyPartieResult ルビコン判定の結果スコアに変換する。
// 古典ルール:
//   - 通常: 勝者 += 100 + (勝者 - 敗者)
//   - 敗者が100点未満 (ルビコン): 勝者 += 100 + (勝者 + 敗者)
func (p *Piquet) applyPartieResult(winner, winScore, loseScore int) {
	p.winnerIdx = winner
	bonus := PiquetWinnerBonus
	if loseScore < PiquetRubiconScore {
		bonus += winScore + loseScore
	} else {
		bonus += winScore - loseScore
	}
	p.players[winner].AddMatchScore(bonus)
}

// hasNoCourtCards 手札に J/Q/K が一切ないか
func hasNoCourtCards(pl *PiquetPlayer) bool {
	for i := 0; i < pl.GetCardsSize(); i++ {
		v := pl.GetCard(i).GetValue()
		if v == 11 || v == 12 || v == 13 {
			return false
		}
	}
	return true
}

// recordFirstScorer 初めて得点したプレイヤーを記録する (pique/repique 判定用)
func (p *Piquet) recordFirstScorer(idx int) {
	if p.firstScorerIdx == -1 {
		p.firstScorerIdx = idx
	}
}

// ───── Getters ─────

// GetPhase 現在のフェーズ
func (p *Piquet) GetPhase() PiquetPhase { return p.phase }

// GetDealNumber 現在のディール番号 (1始まり)
func (p *Piquet) GetDealNumber() int { return p.dealNumber }

// GetDealsPerPartie 設定されたディール数
func (p *Piquet) GetDealsPerPartie() int { return p.config.DealsPerPartie }

// GetElderIdx 現ディールの Elder インデックス (0 or 1)
func (p *Piquet) GetElderIdx() int { return p.elderIdx }

// GetYoungerIdx 現ディールの Younger インデックス
func (p *Piquet) GetYoungerIdx() int { return (p.elderIdx + 1) % PiquetPlayerCnt }

// GetPlayer プレイヤーを取得
func (p *Piquet) GetPlayer(idx int) *PiquetPlayer {
	if idx < 0 || idx >= PiquetPlayerCnt {
		return nil
	}
	return p.players[idx]
}

// GetPlayers 全プレイヤーを取得
func (p *Piquet) GetPlayers() []*PiquetPlayer { return p.players[:] }

// GetTalon タロン (公開しない8枚) を返す
func (p *Piquet) GetTalon() []*Card { return p.talon }

// GetElderTalon Elder の交換対象 (上位5枚) を返す
func (p *Piquet) GetElderTalon() []*Card {
	if len(p.talon) < PiquetElderTalonSize {
		return nil
	}
	return p.talon[:PiquetElderTalonSize]
}

// GetYoungerTalon Younger の交換対象 (下位3枚) を返す
func (p *Piquet) GetYoungerTalon() []*Card {
	if len(p.talon) < PiquetTalonSize {
		return nil
	}
	return p.talon[PiquetElderTalonSize:PiquetTalonSize]
}

// GetExchangeTurn 現在の交換手番
func (p *Piquet) GetExchangeTurn() PiquetExchangeTurn { return p.exchangeTurn }

// GetElderExchangedCnt Elder の交換枚数 (1..5)
func (p *Piquet) GetElderExchangedCnt() int { return p.elderExchangedCnt }

// GetYoungerExchangedCnt Younger の交換枚数 (0..3)
func (p *Piquet) GetYoungerExchangedCnt() int { return p.youngerExchangedCnt }

// GetCarteBlanche 指定プレイヤーがカルトブランシュ達成か
func (p *Piquet) GetCarteBlanche(idx int) bool {
	if idx < 0 || idx >= PiquetPlayerCnt {
		return false
	}
	return p.carteBlanche[idx]
}

// GetDeclStage 現在の宣言ステージ
func (p *Piquet) GetDeclStage() PiquetDeclarationKind { return p.declStage }

// GetDeclResults これまでの宣言結果
func (p *Piquet) GetDeclResults() []*PiquetDeclarationResult { return p.declResults }

// GetCurrentTrick 現在のトリック
func (p *Piquet) GetCurrentTrick() []*TrickCard { return p.currentTrick }

// GetCurrentPlayerIdx 現在の手番プレイヤー
func (p *Piquet) GetCurrentPlayerIdx() int { return p.currentPlayerIdx }

// GetTrickNumber 0始まりのトリック番号 (0..11)
func (p *Piquet) GetTrickNumber() int { return p.trickNumber }

// GetLeadPlayerIdx 現トリックのリード
func (p *Piquet) GetLeadPlayerIdx() int { return p.leadPlayerIdx }

// GetTricksWon プレイヤーごとの獲得トリック数
func (p *Piquet) GetTricksWon(idx int) int {
	if idx < 0 || idx >= PiquetPlayerCnt {
		return 0
	}
	return p.tricksWon[idx]
}

// GetGameEndFlag パルティ終了フラグ
func (p *Piquet) GetGameEndFlag() bool { return p.gameEndFlag }

// GetWinnerIdx パルティ勝者 (-1=未確定 or 引き分け)
func (p *Piquet) GetWinnerIdx() int { return p.winnerIdx }

// GetActionLog アクションログ
func (p *Piquet) GetActionLog() []*ActionLogEntry { return p.actionLog }

// GetConfig 設定値
func (p *Piquet) GetConfig() PiquetConfig { return p.config }

// SetConfig 設定を上書きする
func (p *Piquet) SetConfig(cfg PiquetConfig) { p.config = cfg }

// IsHumanTurn 現在の手番が人間か
func (p *Piquet) IsHumanTurn() bool {
	return p.players[p.currentPlayerIdx].GetIsHuman()
}

// ───── Exchange Phase ─────

// ExchangeElder Elder の交換 (1..5 枚)
//
//	discardIndices: 手札のうち捨てるカードのインデックス
//	Elder は上位5枚の talon から discardIndices と同じ枚数だけ取る
func (p *Piquet) ExchangeElder(discardIndices []int) error {
	if p.phase != PiquetPhaseExchange {
		return errors.New("not in exchange phase")
	}
	if p.exchangeTurn != PiquetExchangeTurnElder {
		return errors.New("not elder's exchange turn")
	}
	n := len(discardIndices)
	if n < 1 || n > PiquetElderTalonSize {
		return fmt.Errorf("elder must exchange 1..%d cards", PiquetElderTalonSize)
	}
	if err := validateUniqueRange(discardIndices, PiquetHandSize); err != nil {
		return err
	}
	elder := p.players[p.elderIdx]
	discarded := elder.RemoveCards(discardIndices)
	if len(discarded) != n {
		return errors.New("invalid discard indices")
	}
	// Elder は talon の上位5枚から先頭 n 枚を取る
	taken := p.talon[:n]
	for _, c := range taken {
		elder.AddCard(c)
	}
	// 残った上位5枚は Elder が見られる
	p.elderRevealedTalon = append(p.elderRevealedTalon, p.talon[n:PiquetElderTalonSize]...)
	// talon を「Younger用3枚 + 残り上位 (Elder revealed) + Younger revealed」に整理
	rest := p.talon[PiquetElderTalonSize:]
	p.talon = rest // youngerTalon (3枚) のみが残る
	p.elderExchanged = taken
	p.elderExchangedCnt = n
	p.exchangeTurn = PiquetExchangeTurnYounger
	p.appendLog(p.elderIdx, "exchange", fmt.Sprintf("Elder交換 %d枚", n), discarded)
	return nil
}

// ExchangeYounger Younger の交換 (0..3 枚)
func (p *Piquet) ExchangeYounger(discardIndices []int) error {
	if p.phase != PiquetPhaseExchange {
		return errors.New("not in exchange phase")
	}
	if p.exchangeTurn != PiquetExchangeTurnYounger {
		return errors.New("not younger's exchange turn")
	}
	n := len(discardIndices)
	if n < 0 || n > PiquetYoungerTalonSize {
		return fmt.Errorf("younger must exchange 0..%d cards", PiquetYoungerTalonSize)
	}
	if err := validateUniqueRange(discardIndices, PiquetHandSize); err != nil {
		return err
	}
	younger := p.players[p.GetYoungerIdx()]
	if n > 0 {
		discarded := younger.RemoveCards(discardIndices)
		if len(discarded) != n {
			return errors.New("invalid discard indices")
		}
		taken := p.talon[:n]
		for _, c := range taken {
			younger.AddCard(c)
		}
		p.youngerExchanged = taken
		// 残った Younger talon は Younger 用 reveal
		p.youngerRevealedTalon = append(p.youngerRevealedTalon, p.talon[n:PiquetYoungerTalonSize]...)
		p.talon = nil
		p.appendLog(p.GetYoungerIdx(), "exchange", fmt.Sprintf("Younger交換 %d枚", n), discarded)
	} else {
		// 0枚交換: Younger は3枚全部 reveal できる (Elderにも公開される伝統)
		p.youngerRevealedTalon = append(p.youngerRevealedTalon, p.talon[:PiquetYoungerTalonSize]...)
		p.talon = nil
		p.appendLog(p.GetYoungerIdx(), "exchange", "Younger交換 0枚", nil)
	}
	p.youngerExchangedCnt = n
	p.exchangeTurn = PiquetExchangeTurnDone
	p.phase = PiquetPhaseDeclaration
	p.declStage = PiquetDeclKindPoint
	return nil
}

// validateUniqueRange インデックスが範囲内かつ重複なし
func validateUniqueRange(indices []int, max int) error {
	seen := make(map[int]bool)
	for _, idx := range indices {
		if idx < 0 || idx >= max {
			return fmt.Errorf("index %d out of range [0,%d)", idx, max)
		}
		if seen[idx] {
			return fmt.Errorf("duplicate index %d", idx)
		}
		seen[idx] = true
	}
	return nil
}

// GetElderRevealedTalon Elder が見られる talon カード
func (p *Piquet) GetElderRevealedTalon() []*Card { return p.elderRevealedTalon }

// GetYoungerRevealedTalon Younger が見られる talon カード
func (p *Piquet) GetYoungerRevealedTalon() []*Card { return p.youngerRevealedTalon }

// ───── Declaration Phase ─────

// ResolveDeclaration 現ステージ (Point→Sequence→Set) の宣言を比較する。
// 結果を declResults に追加し、次ステージへ進める。
// 3ステージ完了時に PiquetPhasePlay へ遷移する。
func (p *Piquet) ResolveDeclaration() (*PiquetDeclarationResult, error) {
	if p.phase != PiquetPhaseDeclaration {
		return nil, errors.New("not in declaration phase")
	}
	elder := p.players[p.elderIdx]
	younger := p.players[p.GetYoungerIdx()]

	var result *PiquetDeclarationResult
	switch p.declStage {
	case PiquetDeclKindPoint:
		result = compareClaims(PiquetDeclKindPoint, bestPoint(elder), bestPoint(younger), p.elderIdx)
	case PiquetDeclKindSequence:
		result = compareClaims(PiquetDeclKindSequence, bestSequence(elder), bestSequence(younger), p.elderIdx)
	case PiquetDeclKindSet:
		eSets := allSets(elder)
		ySets := allSets(younger)
		result = compareSets(eSets, ySets, p.elderIdx)
	default:
		return nil, errors.New("unknown declaration stage")
	}

	// 得点を反映
	if result.Winner >= 0 && result.Score > 0 {
		p.players[result.ScoredBy].AddDeclScore(result.Score)
		p.recordFirstScorer(result.ScoredBy)
	}
	p.declResults = append(p.declResults, result)
	p.appendLog(result.ScoredBy, "declare", fmt.Sprintf("宣言結果 kind=%d winner=%d score=%d", result.Kind, result.Winner, result.Score), nil)

	// 次ステージ or プレイ遷移
	switch p.declStage {
	case PiquetDeclKindPoint:
		p.declStage = PiquetDeclKindSequence
	case PiquetDeclKindSequence:
		p.declStage = PiquetDeclKindSet
	case PiquetDeclKindSet:
		// 全宣言完了 → repique 判定
		p.checkRepique()
		p.phase = PiquetPhasePlay
		p.currentPlayerIdx = p.elderIdx
		p.leadPlayerIdx = p.elderIdx
	}
	return result, nil
}

// checkRepique 宣言だけで30点に達した側に +60 を付与する。
// 条件: 相手はこの時点で0点である必要がある (カルトブランシュ含む)。
func (p *Piquet) checkRepique() {
	for i := range PiquetPlayerCnt {
		opp := (i + 1) % PiquetPlayerCnt
		if p.players[i].GetDeclScore() >= 30 && p.players[opp].GetDeclScore() == 0 {
			p.players[i].AddBonusScore(PiquetRepiqueBonus)
			p.appendLog(i, "repique", "ルピーク +60", nil)
		}
	}
}

// ───── Play Phase ─────

// GetLegalPlayIndices 指定プレイヤーがプレイ可能なカードのインデックス
// ルール: フォロースートが可能なら必須。
func (p *Piquet) GetLegalPlayIndices(playerIdx int) []int {
	if p.phase != PiquetPhasePlay {
		return nil
	}
	pl := p.players[playerIdx]
	n := pl.GetCardsSize()
	if n == 0 {
		return nil
	}
	// リードがなければ全て合法
	if len(p.currentTrick) == 0 {
		out := make([]int, n)
		for i := range out {
			out[i] = i
		}
		return out
	}
	leadSuit := p.currentTrick[0].Card.GetDesign()
	var followIdx []int
	for i := range n {
		if pl.GetCard(i).GetDesign() == leadSuit {
			followIdx = append(followIdx, i)
		}
	}
	if len(followIdx) > 0 {
		return followIdx
	}
	out := make([]int, n)
	for i := range out {
		out[i] = i
	}
	return out
}

// PlayCard 現在の手番プレイヤーがカードをプレイ
func (p *Piquet) PlayCard(cardIdx int) error {
	if p.phase != PiquetPhasePlay {
		return errors.New("not in play phase")
	}
	playerIdx := p.currentPlayerIdx
	legal := p.GetLegalPlayIndices(playerIdx)
	if !slices.Contains(legal, cardIdx) {
		return errors.New("illegal play")
	}
	pl := p.players[playerIdx]
	card := pl.RemoveCard(cardIdx)
	p.currentTrick = append(p.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	p.appendLog(playerIdx, "play", "カードプレイ", []*Card{card})

	// リード側が出した瞬間は得点
	if playerIdx == p.leadPlayerIdx {
		p.players[playerIdx].AddTrickScore(PiquetTrickLeadPoint)
		p.recordFirstScorer(playerIdx)
		p.checkPique(playerIdx)
	}

	if len(p.currentTrick) == PiquetPlayerCnt {
		p.resolveTrick()
	} else {
		p.currentPlayerIdx = (p.currentPlayerIdx + 1) % PiquetPlayerCnt
	}
	return nil
}

// resolveTrick トリック解決
func (p *Piquet) resolveTrick() {
	leadSuit := p.currentTrick[0].Card.GetDesign()
	bestIdx := 0
	for i := 1; i < len(p.currentTrick); i++ {
		c := p.currentTrick[i].Card
		if c.GetDesign() != leadSuit {
			continue
		}
		if piquetCardRank(c.GetValue()) > piquetCardRank(p.currentTrick[bestIdx].Card.GetValue()) {
			bestIdx = i
		}
	}
	winner := p.currentTrick[bestIdx].PlayerIdx

	// 取得点: リーダー側が勝った場合は既に +1 加算済み (PlayCardのリード加点)
	// リーダー以外が取った場合は +1
	if winner != p.leadPlayerIdx {
		p.players[winner].AddTrickScore(PiquetTrickWonPoint)
		p.recordFirstScorer(winner)
		p.checkPique(winner)
	}

	// トリック保管
	trickCards := make([]*Card, 0, PiquetPlayerCnt)
	for _, tc := range p.currentTrick {
		trickCards = append(trickCards, tc.Card)
	}
	p.players[winner].AddTrick(trickCards)
	p.tricksWon[winner]++

	// 最終トリックボーナス
	if p.trickNumber == PiquetTricksPerRound-1 {
		p.players[winner].AddTrickScore(PiquetLastTrickBonus)
		p.recordFirstScorer(winner)
		p.checkPique(winner)
	}

	p.currentTrick = nil
	p.trickNumber++
	if p.trickNumber >= PiquetTricksPerRound {
		// プレイ終了 → スコアリング
		p.endRoundScoring()
		return
	}
	p.leadPlayerIdx = winner
	p.currentPlayerIdx = winner
}

// checkPique Elder が宣言+プレイで30点に達した瞬間に +30 を与える。
// 条件: 相手はその時点で0点 (まだ得点していない)。
// Elder専用ルール。
func (p *Piquet) checkPique(scorer int) {
	if scorer != p.elderIdx {
		return
	}
	if p.elderReached30InPlay {
		return
	}
	elderTotal := p.players[p.elderIdx].GetDeclScore() + p.players[p.elderIdx].GetTrickScore()
	youngerTotal := p.players[p.GetYoungerIdx()].GetDeclScore() + p.players[p.GetYoungerIdx()].GetTrickScore()
	if elderTotal >= 30 && youngerTotal == 0 {
		p.players[p.elderIdx].AddBonusScore(PiquetPiqueBonus)
		p.elderReached30InPlay = true
		p.appendLog(p.elderIdx, "pique", "ピーク +30", nil)
	}
}

// endRoundScoring プレイ完了後のラウンドボーナス処理 (cards / capot)
func (p *Piquet) endRoundScoring() {
	t0 := p.tricksWon[0]
	t1 := p.tricksWon[1]
	switch {
	case t0 == PiquetTricksPerRound:
		p.players[0].AddBonusScore(PiquetCapotBonus)
		p.appendLog(0, "capot", "カポー +40", nil)
	case t1 == PiquetTricksPerRound:
		p.players[1].AddBonusScore(PiquetCapotBonus)
		p.appendLog(1, "capot", "カポー +40", nil)
	case t0 > t1:
		p.players[0].AddBonusScore(PiquetCardsBonus)
		p.appendLog(0, "cards", "カード +10", nil)
	case t1 > t0:
		p.players[1].AddBonusScore(PiquetCardsBonus)
		p.appendLog(1, "cards", "カード +10", nil)
		// 引き分けなら誰にも cards ボーナスなし (歴史的にはディーラーに付与する説もあるがここでは無し)
	}

	// ラウンドスコアを通算スコアに加算
	for i := range PiquetPlayerCnt {
		p.players[i].AddMatchScore(p.players[i].GetRoundScore())
	}
	p.phase = PiquetPhaseScore
	p.appendLog(-1, "round", fmt.Sprintf("ディール%d 終了", p.dealNumber), nil)

	// パルティ終了判定
	if p.dealNumber >= p.config.DealsPerPartie {
		p.finishPartie()
	}
}

// ───── CPU ─────

// CpuPlay CPU の手番処理 (フェーズに応じて自動)
func (p *Piquet) CpuPlay() {
	if p.phase == PiquetPhaseExchange {
		switch p.exchangeTurn {
		case PiquetExchangeTurnElder:
			if !p.players[p.elderIdx].GetIsHuman() {
				_ = p.ExchangeElder(p.cpuChooseElderDiscards())
			}
		case PiquetExchangeTurnYounger:
			if !p.players[p.GetYoungerIdx()].GetIsHuman() {
				_ = p.ExchangeYounger(p.cpuChooseYoungerDiscards())
			}
		}
		return
	}
	if p.phase == PiquetPhasePlay {
		if p.players[p.currentPlayerIdx].GetIsHuman() {
			return
		}
		_ = p.PlayCard(p.cpuChoosePlay(p.currentPlayerIdx))
	}
}

// cpuChooseElderDiscards Elder CPU は弱いカード5枚を捨てて全交換する。
func (p *Piquet) cpuChooseElderDiscards() []int {
	return p.cpuLowestDiscardIndices(p.elderIdx, PiquetElderTalonSize)
}

// cpuChooseYoungerDiscards Younger CPU は弱いカード3枚を捨てて全交換する。
func (p *Piquet) cpuChooseYoungerDiscards() []int {
	return p.cpuLowestDiscardIndices(p.GetYoungerIdx(), PiquetYoungerTalonSize)
}

// cpuLowestDiscardIndices 弱いランク順に n 枚のインデックスを返す。
func (p *Piquet) cpuLowestDiscardIndices(playerIdx, n int) []int {
	pl := p.players[playerIdx]
	type cardRef struct {
		idx, rank int
	}
	refs := make([]cardRef, pl.GetCardsSize())
	for i := 0; i < pl.GetCardsSize(); i++ {
		refs[i] = cardRef{i, piquetCardRank(pl.GetCard(i).GetValue())}
	}
	sort.Slice(refs, func(i, j int) bool { return refs[i].rank < refs[j].rank })
	out := make([]int, 0, n)
	for i := 0; i < n && i < len(refs); i++ {
		out = append(out, refs[i].idx)
	}
	sort.Ints(out)
	return out
}

// cpuChoosePlay CPU は合法なカードのうち最も低いものをプレイする (基本AI)。
func (p *Piquet) cpuChoosePlay(playerIdx int) int {
	legal := p.GetLegalPlayIndices(playerIdx)
	if len(legal) == 0 {
		return -1
	}
	pl := p.players[playerIdx]
	bestIdx := legal[0]
	bestRank := piquetCardRank(pl.GetCard(bestIdx).GetValue())
	for _, idx := range legal[1:] {
		r := piquetCardRank(pl.GetCard(idx).GetValue())
		if r < bestRank {
			bestRank = r
			bestIdx = idx
		}
	}
	return bestIdx
}

// ───── Hint ─────

// GetHint プレイ/交換時の推奨を返す
func (p *Piquet) GetHint(playerIdx int) *PiquetHint {
	switch p.phase {
	case PiquetPhaseExchange:
		if p.exchangeTurn == PiquetExchangeTurnElder && playerIdx == p.elderIdx {
			return &PiquetHint{DiscardIndices: p.cpuChooseElderDiscards(), Reason: "lowest"}
		}
		if p.exchangeTurn == PiquetExchangeTurnYounger && playerIdx == p.GetYoungerIdx() {
			return &PiquetHint{DiscardIndices: p.cpuChooseYoungerDiscards(), Reason: "lowest"}
		}
	case PiquetPhasePlay:
		if playerIdx == p.currentPlayerIdx {
			idx := p.cpuChoosePlay(playerIdx)
			if idx >= 0 {
				return &PiquetHint{CardIndex: &idx, Reason: "lowest"}
			}
		}
	}
	return nil
}

// ───── Log ─────

func (p *Piquet) appendLog(playerIdx int, action, detail string, cards []*Card) {
	entry := &ActionLogEntry{
		ActionType: action,
		Detail:     detail,
		Cards:      cards,
		PlayerIdx:  playerIdx,
		TurnNumber: len(p.actionLog),
	}
	p.actionLog = append(p.actionLog, entry)
}

// ───── Card rank helper ─────

// piquetCardRank ピケのランク (A=8, K=7, Q=6, J=5, 10=4, 9=3, 8=2, 7=1)
func piquetCardRank(value int) int {
	switch value {
	case 1:
		return 8
	case 13:
		return 7
	case 12:
		return 6
	case 11:
		return 5
	case 10:
		return 4
	case 9:
		return 3
	case 8:
		return 2
	case 7:
		return 1
	}
	return 0
}

// piquetPipValue ピケのピップ値 (A=11, K=Q=J=10, 10=10, 9=9, 8=8, 7=7)
func piquetPipValue(value int) int {
	switch value {
	case 1:
		return 11
	case 13, 12, 11:
		return 10
	default:
		return value
	}
}

// ───── Hand evaluation helpers ─────

// bestPoint 手札から最高の Point (同一スート枚数) を返す
func bestPoint(pl *PiquetPlayer) *PiquetClaim {
	suits := [5][]int{}
	for i := 0; i < pl.GetCardsSize(); i++ {
		c := pl.GetCard(i)
		suits[c.GetDesign()] = append(suits[c.GetDesign()], i)
	}
	var best *PiquetClaim
	for s := 1; s <= 4; s++ {
		idxs := suits[s]
		if len(idxs) == 0 {
			continue
		}
		pip := 0
		cards := make([]*Card, 0, len(idxs))
		for _, idx := range idxs {
			c := pl.GetCard(idx)
			pip += piquetPipValue(c.GetValue())
			cards = append(cards, c)
		}
		claim := &PiquetClaim{
			Length:   len(idxs),
			PipTotal: pip,
			Suit:     s,
			Cards:    cards,
		}
		if best == nil ||
			claim.Length > best.Length ||
			(claim.Length == best.Length && claim.PipTotal > best.PipTotal) {
			best = claim
		}
	}
	return best
}

// bestSequence 手札から最高の Sequence (同一スート連続3枚以上) を返す
func bestSequence(pl *PiquetPlayer) *PiquetClaim {
	bySuit := [5][]*Card{}
	for i := 0; i < pl.GetCardsSize(); i++ {
		c := pl.GetCard(i)
		bySuit[c.GetDesign()] = append(bySuit[c.GetDesign()], c)
	}
	var best *PiquetClaim
	for s := 1; s <= 4; s++ {
		cards := bySuit[s]
		if len(cards) < PiquetMinSequenceLength {
			continue
		}
		// ランク昇順にソート
		sort.SliceStable(cards, func(i, j int) bool {
			return piquetCardRank(cards[i].GetValue()) < piquetCardRank(cards[j].GetValue())
		})
		// 連続ランの最長を探す
		i := 0
		for i < len(cards) {
			j := i + 1
			for j < len(cards) && piquetCardRank(cards[j].GetValue()) == piquetCardRank(cards[j-1].GetValue())+1 {
				j++
			}
			runLen := j - i
			if runLen >= PiquetMinSequenceLength {
				top := cards[j-1]
				run := append([]*Card{}, cards[i:j]...)
				claim := &PiquetClaim{
					Length:  runLen,
					TopRank: piquetCardRank(top.GetValue()),
					Suit:    s,
					Cards:   run,
				}
				if best == nil ||
					claim.Length > best.Length ||
					(claim.Length == best.Length && claim.TopRank > best.TopRank) {
					best = claim
				}
			}
			i = j
		}
	}
	return best
}

// allSets 手札から trio/quatorze の全セットを返す (10以上のランクのみ)
func allSets(pl *PiquetPlayer) []*PiquetClaim {
	byRank := map[int][]*Card{}
	for i := 0; i < pl.GetCardsSize(); i++ {
		c := pl.GetCard(i)
		if c.GetValue() < PiquetSetMinRank && c.GetValue() != 1 {
			continue
		}
		byRank[c.GetValue()] = append(byRank[c.GetValue()], c)
	}
	var out []*PiquetClaim
	// ランク順 (A,K,Q,J,10)
	rankOrder := []int{1, 13, 12, 11, 10}
	for _, r := range rankOrder {
		cards := byRank[r]
		if len(cards) >= 3 {
			out = append(out, &PiquetClaim{
				Length:  len(cards),
				TopRank: piquetCardRank(r),
				Cards:   cards,
			})
		}
	}
	return out
}

// compareClaims Point/Sequence 用の比較
func compareClaims(kind PiquetDeclarationKind, elder, younger *PiquetClaim, elderIdx int) *PiquetDeclarationResult {
	result := &PiquetDeclarationResult{
		Kind:         kind,
		ElderClaim:   elder,
		YoungerClaim: younger,
		Winner:       -1,
		ScoredBy:     -1,
	}
	winner := compareClaimStrength(kind, elder, younger)
	switch winner {
	case 1:
		// Elder勝ち
		result.Winner = elderIdx
		result.ScoredBy = elderIdx
		result.Score = claimScore(kind, elder)
	case -1:
		// Younger勝ち
		result.Winner = (elderIdx + 1) % PiquetPlayerCnt
		result.ScoredBy = result.Winner
		result.Score = claimScore(kind, younger)
	}
	return result
}

// compareClaimStrength 1=Elder勝ち, -1=Younger勝ち, 0=タイ
func compareClaimStrength(kind PiquetDeclarationKind, elder, younger *PiquetClaim) int {
	if elder == nil && younger == nil {
		return 0
	}
	if elder == nil {
		return -1
	}
	if younger == nil {
		return 1
	}
	if elder.Length > younger.Length {
		return 1
	}
	if elder.Length < younger.Length {
		return -1
	}
	// 同枚数 → タイブレーク
	switch kind {
	case PiquetDeclKindPoint:
		if elder.PipTotal > younger.PipTotal {
			return 1
		}
		if elder.PipTotal < younger.PipTotal {
			return -1
		}
		return 0
	case PiquetDeclKindSequence:
		if elder.TopRank > younger.TopRank {
			return 1
		}
		if elder.TopRank < younger.TopRank {
			return -1
		}
		return 0
	}
	return 0
}

// claimScore 単一 claim のスコア
func claimScore(kind PiquetDeclarationKind, claim *PiquetClaim) int {
	if claim == nil {
		return 0
	}
	switch kind {
	case PiquetDeclKindPoint:
		return claim.Length
	case PiquetDeclKindSequence:
		switch claim.Length {
		case 3:
			return 3
		case 4:
			return 4
		case 5:
			return 15
		case 6:
			return 16
		case 7:
			return 17
		default:
			if claim.Length >= 8 {
				return 18
			}
			return 0
		}
	case PiquetDeclKindSet:
		if claim.Length == 4 {
			return 14
		}
		if claim.Length == 3 {
			return 3
		}
	}
	return 0
}

// compareSets Set用の比較。最強のセットで勝者を決め、勝者の全セットを合算スコア化する。
func compareSets(elderSets, youngerSets []*PiquetClaim, elderIdx int) *PiquetDeclarationResult {
	result := &PiquetDeclarationResult{
		Kind:     PiquetDeclKindSet,
		Winner:   -1,
		ScoredBy: -1,
	}
	var elderBest, youngerBest *PiquetClaim
	if len(elderSets) > 0 {
		elderBest = elderSets[0]
		for _, s := range elderSets {
			if isStrongerSet(s, elderBest) {
				elderBest = s
			}
		}
	}
	if len(youngerSets) > 0 {
		youngerBest = youngerSets[0]
		for _, s := range youngerSets {
			if isStrongerSet(s, youngerBest) {
				youngerBest = s
			}
		}
	}
	result.ElderClaim = elderBest
	result.YoungerClaim = youngerBest

	winner := compareSetStrength(elderBest, youngerBest)
	switch winner {
	case 1:
		result.Winner = elderIdx
		result.ScoredBy = elderIdx
		for _, s := range elderSets {
			result.Score += claimScore(PiquetDeclKindSet, s)
			result.Sets = append(result.Sets, s)
		}
	case -1:
		result.Winner = (elderIdx + 1) % PiquetPlayerCnt
		result.ScoredBy = result.Winner
		for _, s := range youngerSets {
			result.Score += claimScore(PiquetDeclKindSet, s)
			result.Sets = append(result.Sets, s)
		}
	}
	return result
}

// isStrongerSet a が b より強いセットか (4枚 > 3枚, 同枚数なら TopRank で比較)
func isStrongerSet(a, b *PiquetClaim) bool {
	if a.Length > b.Length {
		return true
	}
	if a.Length < b.Length {
		return false
	}
	return a.TopRank > b.TopRank
}

// compareSetStrength 1=Elder, -1=Younger, 0=どちらも勝てない
func compareSetStrength(elder, younger *PiquetClaim) int {
	if elder == nil && younger == nil {
		return 0
	}
	if elder == nil {
		return -1
	}
	if younger == nil {
		return 1
	}
	// 4枚 > 3枚
	if elder.Length > younger.Length {
		return 1
	}
	if elder.Length < younger.Length {
		return -1
	}
	// 同枚数 → TopRank
	if elder.TopRank > younger.TopRank {
		return 1
	}
	if elder.TopRank < younger.TopRank {
		return -1
	}
	return 0 // 完全タイ
}

// ───── JSON ─────

// piquetJSON wire format
type piquetJSON struct {
	TrumpCards           *TrumpCards                    `json:"tc"`
	Players              [PiquetPlayerCnt]*PiquetPlayer `json:"pl"`
	Config               PiquetConfig                   `json:"cfg"`
	Phase                PiquetPhase                    `json:"ph"`
	DealNumber           int                            `json:"dn"`
	ElderIdx             int                            `json:"ei"`
	GameEndFlag          bool                           `json:"ge"`
	WinnerIdx            int                            `json:"wi"`
	Talon                []*Card                        `json:"tl"`
	ElderExchanged       []*Card                        `json:"ee"`
	YoungerExchanged     []*Card                        `json:"ye"`
	ElderExchangedCnt    int                            `json:"ec"`
	YoungerExchangedCnt  int                            `json:"yc"`
	ExchangeTurn         PiquetExchangeTurn             `json:"et"`
	ElderRevealedTalon   []*Card                        `json:"er"`
	YoungerRevealedTalon []*Card                        `json:"yr"`
	DeclStage            PiquetDeclarationKind          `json:"ds"`
	DeclResults          []*PiquetDeclarationResult     `json:"dr"`
	CurrentTrick         []*TrickCard                   `json:"ct"`
	CurrentPlayerIdx     int                            `json:"cp"`
	TrickNumber          int                            `json:"tn"`
	LeadPlayerIdx        int                            `json:"lp"`
	TricksWon            [PiquetPlayerCnt]int           `json:"tw"`
	FirstScorerIdx       int                            `json:"fs"`
	ElderReached30InPlay bool                           `json:"er30"`
	CarteBlanche         [PiquetPlayerCnt]bool          `json:"cb"`
	ActionLog            []*ActionLogEntry              `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (p *Piquet) MarshalJSON() ([]byte, error) {
	return json.Marshal(piquetJSON{
		TrumpCards:           p.trumpCards,
		Players:              p.players,
		Config:               p.config,
		Phase:                p.phase,
		DealNumber:           p.dealNumber,
		ElderIdx:             p.elderIdx,
		GameEndFlag:          p.gameEndFlag,
		WinnerIdx:            p.winnerIdx,
		Talon:                p.talon,
		ElderExchanged:       p.elderExchanged,
		YoungerExchanged:     p.youngerExchanged,
		ElderExchangedCnt:    p.elderExchangedCnt,
		YoungerExchangedCnt:  p.youngerExchangedCnt,
		ExchangeTurn:         p.exchangeTurn,
		ElderRevealedTalon:   p.elderRevealedTalon,
		YoungerRevealedTalon: p.youngerRevealedTalon,
		DeclStage:            p.declStage,
		DeclResults:          p.declResults,
		CurrentTrick:         p.currentTrick,
		CurrentPlayerIdx:     p.currentPlayerIdx,
		TrickNumber:          p.trickNumber,
		LeadPlayerIdx:        p.leadPlayerIdx,
		TricksWon:            p.tricksWon,
		FirstScorerIdx:       p.firstScorerIdx,
		ElderReached30InPlay: p.elderReached30InPlay,
		CarteBlanche:         p.carteBlanche,
		ActionLog:            p.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *Piquet) UnmarshalJSON(data []byte) error {
	var j piquetJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	p.trumpCards = j.TrumpCards
	p.players = j.Players
	p.config = j.Config
	p.phase = j.Phase
	p.dealNumber = j.DealNumber
	p.elderIdx = j.ElderIdx
	p.gameEndFlag = j.GameEndFlag
	p.winnerIdx = j.WinnerIdx
	p.talon = j.Talon
	p.elderExchanged = j.ElderExchanged
	p.youngerExchanged = j.YoungerExchanged
	p.elderExchangedCnt = j.ElderExchangedCnt
	p.youngerExchangedCnt = j.YoungerExchangedCnt
	p.exchangeTurn = j.ExchangeTurn
	p.elderRevealedTalon = j.ElderRevealedTalon
	p.youngerRevealedTalon = j.YoungerRevealedTalon
	p.declStage = j.DeclStage
	p.declResults = j.DeclResults
	p.currentTrick = j.CurrentTrick
	p.currentPlayerIdx = j.CurrentPlayerIdx
	p.trickNumber = j.TrickNumber
	p.leadPlayerIdx = j.LeadPlayerIdx
	p.tricksWon = j.TricksWon
	p.firstScorerIdx = j.FirstScorerIdx
	p.elderReached30InPlay = j.ElderReached30InPlay
	p.carteBlanche = j.CarteBlanche
	p.actionLog = j.ActionLog
	return nil
}
