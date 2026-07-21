package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// PigsTailPlayerCnt ぶたのしっぽデフォルトプレイヤー数
const PigsTailPlayerCnt = 4

// pigsTailShuffleCount シャッフル回数
const pigsTailShuffleCount = 10

// PigsTailCpuAction CPUの1ターン分の行動記録
type PigsTailCpuAction struct {
	DrawPlayerIdx int   // 引いたプレイヤーインデックス
	DrawnCard     *Card // 引いたカード
	PenaltyFlag   bool  // ペナルティが発生したか
	PenaltyCount  int   // 引き取ったカード枚数
	HesitationMs  int   // 迷い時間ディレイ (ミリ秒; 0=無効)
}

// pigsTailCpuActionJSON is the JSON wire format for PigsTailCpuAction.
type pigsTailCpuActionJSON struct {
	DrawPlayerIdx int   `json:"dp"`
	DrawnCard     *Card `json:"dc"`
	PenaltyFlag   bool  `json:"pf"`
	PenaltyCount  int   `json:"pc"`
	HesitationMs  int   `json:"hm"`
}

// MarshalJSON implements json.Marshaler.
func (a *PigsTailCpuAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(pigsTailCpuActionJSON{
		DrawPlayerIdx: a.DrawPlayerIdx,
		DrawnCard:     a.DrawnCard,
		PenaltyFlag:   a.PenaltyFlag,
		PenaltyCount:  a.PenaltyCount,
		HesitationMs:  a.HesitationMs,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *PigsTailCpuAction) UnmarshalJSON(data []byte) error {
	var j pigsTailCpuActionJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	a.DrawPlayerIdx = j.DrawPlayerIdx
	a.DrawnCard = j.DrawnCard
	a.PenaltyFlag = j.PenaltyFlag
	a.PenaltyCount = j.PenaltyCount
	a.HesitationMs = j.HesitationMs
	return nil
}

// PigsTail ぶたのしっぽゲームクラス
type PigsTail struct {
	trumpCards   *TrumpCards          // 山札(円状)
	center       []*Card              // 中央の場札
	players      []*PigsTailPlayer    // プレイヤー
	currentTurn  int                  // 現在の手番プレイヤーインデックス
	gameEndFlag  bool                 // ゲーム終了フラグ
	loserIdx     int                  // 負けたプレイヤーインデックス (手札最多)
	lastDrawCard *Card                // 最後に引いたカード
	lastPenalty  bool                 // 最後のアクションでペナルティが発生したか
	cpuActions   []*PigsTailCpuAction // CPUターンの行動履歴
	humanAction  *PigsTailCpuAction   // 人間プレイヤーの最後の行動記録
	config       PigsTailConfig       // ゲーム設定
	actionLog    []*ActionLogEntry    // 棋譜
}

// NewPigsTail コンストラクタ
func NewPigsTail(trumpCards *TrumpCards, players []*PigsTailPlayer) *PigsTail {
	return &PigsTail{
		trumpCards:   trumpCards,
		center:       make([]*Card, 0),
		players:      players,
		currentTurn:  0,
		gameEndFlag:  false,
		loserIdx:     -1,
		lastDrawCard: nil,
		lastPenalty:  false,
		cpuActions:   nil,
		humanAction:  nil,
		config:       DefaultPigsTailConfig(),
		actionLog:    nil,
	}
}

// buildPigsTailPlayers builds a roster of the given size: 1 human followed by
// (count-1) CPUs. The count is clamped to the supported range so callers can
// pass an unvalidated config value safely.
func buildPigsTailPlayers(count int) []*PigsTailPlayer {
	if count < PigsTailMinPlayers {
		count = PigsTailMinPlayers
	} else if count > PigsTailMaxPlayers {
		count = PigsTailMaxPlayers
	}
	players := make([]*PigsTailPlayer, count)
	players[0] = NewPigsTailPlayer(true)
	for i := 1; i < count; i++ {
		players[i] = NewPigsTailPlayer(false)
	}
	return players
}

// NewDefaultPigsTail returns PigsTail with the standard 4-player setup (1 human, 3 CPU).
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultPigsTail() *PigsTail {
	return NewPigsTail(NewTrumpCards(0), buildPigsTailPlayers(PigsTailPlayerCnt))
}

// SetConfig ゲーム設定をセット
func (pt *PigsTail) SetConfig(config PigsTailConfig) { pt.config = config }

// GetConfig ゲーム設定取得
func (pt *PigsTail) GetConfig() PigsTailConfig { return pt.config }

// Reset ゲーム初期化
func (pt *PigsTail) Reset() {
	pt.gameEndFlag = false
	pt.loserIdx = -1
	pt.currentTurn = 0
	pt.lastDrawCard = nil
	pt.lastPenalty = false
	pt.cpuActions = nil
	pt.humanAction = nil
	pt.center = make([]*Card, 0)
	pt.actionLog = nil

	// 設定の参加人数に合わせてロスター (人間1 + CPU) を再構築する。
	pt.players = buildPigsTailPlayers(pt.config.PlayerCount)

	// プレイ順をランダムにする
	rand.Shuffle(len(pt.players), func(i, j int) {
		pt.players[i], pt.players[j] = pt.players[j], pt.players[i]
	})

	// 52枚のデッキをシャッフルして山札にする (ジョーカーなし)
	pt.trumpCards = NewTrumpCards(0)
	for range pigsTailShuffleCount {
		pt.trumpCards.Shuffle()
	}
}

// GetCircleCount 山札の残り枚数を取得する
func (pt *PigsTail) GetCircleCount() int {
	return pt.trumpCards.GetRemainingCount()
}

// GetCenter 中央の場札を取得する
func (pt *PigsTail) GetCenter() []*Card {
	return pt.center
}

// GetCenterTopCard 場札の一番上のカードを取得する (空ならnil)
func (pt *PigsTail) GetCenterTopCard() *Card {
	if len(pt.center) == 0 {
		return nil
	}
	return pt.center[len(pt.center)-1]
}

// drawAndPlace 山札から1枚引いて場札に出し、スートマッチ判定を行う
func (pt *PigsTail) drawAndPlace(playerIdx int) (*Card, bool) {
	card := pt.trumpCards.DrawCard()
	if card == nil {
		return nil, false
	}

	// スートマッチ判定: 場札のトップと同じスートならペナルティ
	topCard := pt.GetCenterTopCard()
	penalty := topCard != nil && card.GetDesign() == topCard.GetDesign()

	if penalty {
		// ペナルティ: 場札を全て手札に引き取る
		pt.center = append(pt.center, card)
		for _, c := range pt.center {
			pt.players[playerIdx].AddCard(c)
		}
		pt.appendLog(playerIdx, "penalty",
			fmt.Sprintf("drew %s, matched suit — took %d cards", cardShortStr(card), len(pt.center)),
			[]*Card{card})
		pt.center = make([]*Card, 0)
	} else {
		// 場札にカードを追加
		pt.center = append(pt.center, card)
		pt.appendLog(playerIdx, "draw",
			fmt.Sprintf("drew %s", cardShortStr(card)),
			[]*Card{card})
	}

	pt.lastDrawCard = card
	pt.lastPenalty = penalty

	return card, penalty
}

// checkGameEnd 山札がなくなったらゲーム終了、手札最多プレイヤーが負け
func (pt *PigsTail) checkGameEnd() bool {
	if pt.trumpCards.GetRemainingCount() > 0 {
		return false
	}

	pt.gameEndFlag = true

	// 手札最多のプレイヤーが負け (同数なら最初に見つかった方)
	maxCards := -1
	for i, p := range pt.players {
		if p.GetCardsSize() > maxCards {
			maxCards = p.GetCardsSize()
			pt.loserIdx = i
		}
	}

	return true
}

// advanceTurn 手番を次のプレイヤーへ進める
func (pt *PigsTail) advanceTurn() {
	if pt.gameEndFlag {
		return
	}
	pt.currentTurn = (pt.currentTurn + 1) % len(pt.players)
}

// PlayerAction 人間プレイヤーのアクション (山札から1枚引く)
func (pt *PigsTail) PlayerAction(_ int) error {
	if pt.gameEndFlag {
		return ErrGameEnded
	}
	if !pt.players[pt.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}

	// 人間のターン開始時にCPU行動履歴をリセット
	pt.cpuActions = nil

	card, penalty := pt.drawAndPlace(pt.currentTurn)
	penaltyCount := 0
	if penalty {
		penaltyCount = pt.players[pt.currentTurn].GetCardsSize()
	}

	pt.humanAction = &PigsTailCpuAction{
		DrawPlayerIdx: pt.currentTurn,
		DrawnCard:     card,
		PenaltyFlag:   penalty,
		PenaltyCount:  penaltyCount,
	}

	if !pt.checkGameEnd() {
		pt.advanceTurn()
	}
	return nil
}

// CpuAction CPUプレイヤーが1ターン実行
func (pt *PigsTail) CpuAction() error {
	if pt.gameEndFlag {
		return ErrGameEnded
	}
	if pt.players[pt.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}

	playerIdx := pt.currentTurn
	card, penalty := pt.drawAndPlace(playerIdx)
	penaltyCount := 0
	if penalty {
		penaltyCount = pt.players[playerIdx].GetCardsSize()
	}

	action := &PigsTailCpuAction{
		DrawPlayerIdx: playerIdx,
		DrawnCard:     card,
		PenaltyFlag:   penalty,
		PenaltyCount:  penaltyCount,
	}
	if pt.config.CpuHesitationEnabled {
		if penalty {
			action.HesitationMs = hesitationMediumMin + rand.Intn(hesitationMediumMax-hesitationMediumMin+1)
		} else {
			action.HesitationMs = hesitationFastMin + rand.Intn(hesitationFastMax-hesitationFastMin+1)
		}
	}
	pt.cpuActions = append(pt.cpuActions, action)

	if !pt.checkGameEnd() {
		pt.advanceTurn()
	}
	return nil
}

// GetGameEndFlag ゲーム終了フラグを取得する
func (pt *PigsTail) GetGameEndFlag() bool { return pt.gameEndFlag }

// IsHumanTurn 現在の手番が人間かを返す
func (pt *PigsTail) IsHumanTurn() bool {
	return !pt.gameEndFlag && pt.players[pt.currentTurn].GetIsHuman()
}

// GetPlayerCnt プレイヤー数を取得する
func (pt *PigsTail) GetPlayerCnt() int { return len(pt.players) }

// GetPlayer 指定インデックスのプレイヤーを取得する
func (pt *PigsTail) GetPlayer(i int) *PigsTailPlayer {
	if i < 0 || i >= len(pt.players) {
		return nil
	}
	return pt.players[i]
}

// GetCurrentTurn 現在の手番プレイヤーインデックスを取得する
func (pt *PigsTail) GetCurrentTurn() int { return pt.currentTurn }

// GetLoserIdx 負けプレイヤーインデックスを取得する
func (pt *PigsTail) GetLoserIdx() int { return pt.loserIdx }

// GetLastDrawCard 最後に引いたカードを取得する
func (pt *PigsTail) GetLastDrawCard() *Card { return pt.lastDrawCard }

// GetLastPenalty 最後のアクションでペナルティが発生したかを取得する
func (pt *PigsTail) GetLastPenalty() bool { return pt.lastPenalty }

// GetCpuActions CPUターンの行動履歴を取得する
func (pt *PigsTail) GetCpuActions() []*PigsTailCpuAction { return pt.cpuActions }

// GetHumanAction 人間の最後の行動記録を取得する
func (pt *PigsTail) GetHumanAction() *PigsTailCpuAction { return pt.humanAction }

// GetActionLog 棋譜を取得する
func (pt *PigsTail) GetActionLog() []*ActionLogEntry { return pt.actionLog }

// appendLog 棋譜にエントリを追加する
func (pt *PigsTail) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	pt.actionLog = append(pt.actionLog, &ActionLogEntry{
		TurnNumber: len(pt.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// cardShortStr カードの短い文字列表現を返す
func cardShortStr(c *Card) string {
	if c == nil {
		return "?"
	}
	designs := map[int]string{
		CardDesignSpade:   "S",
		CardDesignHeart:   "H",
		CardDesignDiamond: "D",
		CardDesignClover:  "C",
	}
	d, ok := designs[c.GetDesign()]
	if !ok {
		d = "?"
	}
	return fmt.Sprintf("%s%d", d, c.GetValue())
}

// --- JSON Marshaling ---

// pigsTailJSON is the JSON wire format for PigsTail.
type pigsTailJSON struct {
	TrumpCards   *TrumpCards          `json:"tc"`
	Center       []*Card              `json:"ce"`
	Players      []*PigsTailPlayer    `json:"pl"`
	CurrentTurn  int                  `json:"ct"`
	GameEndFlag  bool                 `json:"ge"`
	LoserIdx     int                  `json:"li"`
	LastDrawCard *Card                `json:"lc"`
	LastPenalty  bool                 `json:"lp"`
	CpuActions   []*PigsTailCpuAction `json:"ca"`
	HumanAction  *PigsTailCpuAction   `json:"ha"`
	Config       PigsTailConfig       `json:"cf"`
	ActionLog    []*ActionLogEntry    `json:"al"`
}

// pigsTailMaxSliceLen caps slice sizes during deserialisation.
const pigsTailMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (pt *PigsTail) MarshalJSON() ([]byte, error) {
	return json.Marshal(pigsTailJSON{
		TrumpCards:   pt.trumpCards,
		Center:       pt.center,
		Players:      pt.players,
		CurrentTurn:  pt.currentTurn,
		GameEndFlag:  pt.gameEndFlag,
		LoserIdx:     pt.loserIdx,
		LastDrawCard: pt.lastDrawCard,
		LastPenalty:  pt.lastPenalty,
		CpuActions:   pt.cpuActions,
		HumanAction:  pt.humanAction,
		Config:       pt.config,
		ActionLog:    pt.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (pt *PigsTail) UnmarshalJSON(data []byte) error {
	var j pigsTailJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > pigsTailMaxSliceLen || len(j.Center) > pigsTailMaxSliceLen ||
		len(j.CpuActions) > pigsTailMaxSliceLen || len(j.ActionLog) > pigsTailMaxSliceLen {
		return fmt.Errorf("pigsTail: input array exceeds maximum allowed size")
	}
	pt.trumpCards = j.TrumpCards
	if pt.trumpCards == nil {
		pt.trumpCards = NewTrumpCards(0)
	}
	pt.center = j.Center
	if pt.center == nil {
		pt.center = make([]*Card, 0)
	}
	pt.players = j.Players
	if pt.players == nil {
		pt.players = make([]*PigsTailPlayer, 0)
	}
	pt.currentTurn = j.CurrentTurn
	pt.gameEndFlag = j.GameEndFlag
	pt.loserIdx = j.LoserIdx
	pt.lastDrawCard = j.LastDrawCard
	pt.lastPenalty = j.LastPenalty
	pt.cpuActions = j.CpuActions
	pt.humanAction = j.HumanAction
	pt.config = j.Config
	pt.actionLog = j.ActionLog
	if pt.actionLog == nil {
		pt.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
