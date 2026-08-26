//go:build !js || !wasm || extra4

package domain

import (
	"encoding/json"
	"fmt"
)

// カジノウォーフェーズ定数
const (
	CasinoWarPhaseBet          = 1 // ベットフェーズ
	CasinoWarPhaseInitialDealt = 2 // 初手2枚配布済み（ResolveInitial 待ち）
	CasinoWarPhaseTieDecision  = 3 // タイ時の Surrender / War 選択フェーズ
	CasinoWarPhaseWarDealt     = 4 // ウォー追加カード配布済み
	CasinoWarPhaseEnd          = 5 // 終了フェーズ
)

// カジノウォーデフォルト値
const (
	CasinoWarDefaultChips = 1000  // デフォルトチップ
	CasinoWarMinBet       = 10    // 最低ベット額
	CasinoWarMaxBet       = 10000 // 最大ベット額
	CasinoWarBurnCount    = 3     // ウォー時の焼き札枚数
)

// CasinoWar カジノウォーゲーム本体
type CasinoWar struct {
	trumpCards    *TrumpCards
	playerCard    *Card
	dealerCard    *Card
	playerWarCard *Card
	dealerWarCard *Card
	burnCards     []*Card
	chips         ChipHolder
	ante          int
	warBet        int
	phase         int
	gameEndFlag   bool
	result        GameResult
	totalPayout   int
	actionLogBase
}

// NewCasinoWar コンストラクタ
func NewCasinoWar(trumpCards *TrumpCards) *CasinoWar {
	trumpCards.Shuffle()
	return &CasinoWar{
		trumpCards: trumpCards,
		phase:      CasinoWarPhaseBet,
	}
}

// NewDefaultCasinoWar デフォルト設定のカジノウォーを生成するファクトリ関数
func NewDefaultCasinoWar() *CasinoWar {
	cw := NewCasinoWar(NewTrumpCards(0))
	cw.chips.SetChips(CasinoWarDefaultChips)
	return cw
}

// Reset ゲーム初期化
func (cw *CasinoWar) Reset() {
	cw.gameEndFlag = false
	cw.phase = CasinoWarPhaseBet
	cw.playerCard = nil
	cw.dealerCard = nil
	cw.playerWarCard = nil
	cw.dealerWarCard = nil
	cw.burnCards = nil
	cw.ante = 0
	cw.warBet = 0
	cw.result = 0
	cw.totalPayout = 0
	cw.actionLog = nil
	if cw.chips.GetChips() < CasinoWarMinBet {
		cw.chips.SetChips(CasinoWarDefaultChips)
	}
	cw.trumpCards = NewTrumpCards(0)
	cw.trumpCards.Shuffle()
}

// Bet アンテをベットしプレイヤー／ディーラーに 1 枚ずつ配る
func (cw *CasinoWar) Bet(amount int) error {
	if cw.phase != CasinoWarPhaseBet {
		return NewDomainError(ErrWrongPhase, "Bet is only allowed during the bet phase.")
	}
	if amount < CasinoWarMinBet || amount%CasinoWarMinBet != 0 || amount > CasinoWarMaxBet {
		return NewDomainError(ErrInvalidAmount, "Invalid bet amount.")
	}
	if !cw.chips.SubtractChips(amount) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}
	cw.ante = amount
	cw.appendLog(0, "bet", fmt.Sprintf("ante=%d", amount), nil)

	cw.dealInitial()
	cw.phase = CasinoWarPhaseInitialDealt
	return nil
}

// dealInitial プレイヤーとディーラーに 1 枚ずつ配る
func (cw *CasinoWar) dealInitial() {
	if cw.playerCard == nil {
		cw.playerCard = cw.trumpCards.DrawCard()
	}
	if cw.dealerCard == nil {
		cw.dealerCard = cw.trumpCards.DrawCard()
	}
	cw.appendLog(-1, "deal", "dealt initial cards", []*Card{cw.playerCard, cw.dealerCard})
}

// ResolveInitial 初手 2 枚を比較しフェーズ遷移
func (cw *CasinoWar) ResolveInitial() {
	if cw.playerCard == nil || cw.dealerCard == nil {
		return
	}
	pr, dr := rankOf(cw.playerCard), rankOf(cw.dealerCard)
	switch {
	case pr > dr:
		cw.result = GameResultWin
		cw.totalPayout = cw.ante * 2
		cw.chips.AddChips(cw.totalPayout)
		cw.gameEndFlag = true
		cw.phase = CasinoWarPhaseEnd
		cw.appendLog(-1, "result", "player wins", nil)
	case pr < dr:
		cw.result = GameResultLose
		cw.totalPayout = 0
		cw.gameEndFlag = true
		cw.phase = CasinoWarPhaseEnd
		cw.appendLog(-1, "result", "player loses", nil)
	default:
		cw.phase = CasinoWarPhaseTieDecision
		cw.appendLog(-1, "tie", "ranks tied — surrender or war", nil)
	}
}

// Surrender タイ時の降参。アンテの半額を取り戻して終了する
func (cw *CasinoWar) Surrender() error {
	if cw.phase != CasinoWarPhaseTieDecision {
		return NewDomainError(ErrWrongPhase, "Surrender is only allowed during the tie decision phase.")
	}
	// Bet enforces ante % CasinoWarMinBet == 0; CasinoWarMinBet (10) is even,
	// so this integer division is exact. If CasinoWarMinBet ever becomes odd,
	// the casino convention is to round down, which is what / does here.
	refund := cw.ante / 2
	cw.totalPayout = refund
	if refund > 0 {
		cw.chips.AddChips(refund)
	}
	cw.result = GameResultLose
	cw.gameEndFlag = true
	cw.phase = CasinoWarPhaseEnd
	cw.appendLog(0, "surrender", fmt.Sprintf("refund=%d", refund), nil)
	return nil
}

// GoToWar タイ時のウォー宣言。同額をベットし焼き札 3 枚を挟んで再勝負する
func (cw *CasinoWar) GoToWar() error {
	if cw.phase != CasinoWarPhaseTieDecision {
		return NewDomainError(ErrWrongPhase, "War is only allowed during the tie decision phase.")
	}
	if !cw.chips.SubtractChips(cw.ante) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips for war bet.")
	}
	cw.warBet = cw.ante
	cw.appendLog(0, "war", fmt.Sprintf("warBet=%d", cw.warBet), nil)
	cw.dealWar()
	cw.ResolveWar()
	return nil
}

// dealWar 焼き札 3 枚＋プレイヤー／ディーラー 1 枚ずつを配る
func (cw *CasinoWar) dealWar() {
	if len(cw.burnCards) < CasinoWarBurnCount {
		cw.burnCards = make([]*Card, 0, CasinoWarBurnCount)
		for range CasinoWarBurnCount {
			cw.burnCards = append(cw.burnCards, cw.trumpCards.DrawCard())
		}
	}
	if cw.playerWarCard == nil {
		cw.playerWarCard = cw.trumpCards.DrawCard()
	}
	if cw.dealerWarCard == nil {
		cw.dealerWarCard = cw.trumpCards.DrawCard()
	}
	cw.appendLog(-1, "deal", "burn 3 + war cards", []*Card{cw.playerWarCard, cw.dealerWarCard})
	cw.phase = CasinoWarPhaseWarDealt
}

// ResolveWar ウォー後のカードを評価しペイアウトする
func (cw *CasinoWar) ResolveWar() {
	if cw.playerWarCard == nil || cw.dealerWarCard == nil || cw.gameEndFlag {
		return
	}
	pr, dr := rankOf(cw.playerWarCard), rankOf(cw.dealerWarCard)
	if pr >= dr {
		cw.result = GameResultWin
		// アンテはプッシュ（返金）、ウォー bet は 1:1 で支払い
		cw.totalPayout = cw.ante + cw.warBet*2
		cw.chips.AddChips(cw.totalPayout)
		var detail string
		if pr == dr {
			detail = "war tie counted as player win"
		} else {
			detail = "player wins war"
		}
		cw.appendLog(-1, "result", detail, nil)
	} else {
		cw.result = GameResultLose
		cw.totalPayout = 0
		cw.appendLog(-1, "result", "player loses war", nil)
	}
	cw.gameEndFlag = true
	cw.phase = CasinoWarPhaseEnd
}

// --- Getters ---

// GetPlayerCard プレイヤーの初手カード
func (cw *CasinoWar) GetPlayerCard() *Card { return cw.playerCard }

// GetDealerCard ディーラーの初手カード
func (cw *CasinoWar) GetDealerCard() *Card { return cw.dealerCard }

// GetPlayerWarCard プレイヤーのウォーカード
func (cw *CasinoWar) GetPlayerWarCard() *Card { return cw.playerWarCard }

// GetDealerWarCard ディーラーのウォーカード
func (cw *CasinoWar) GetDealerWarCard() *Card { return cw.dealerWarCard }

// GetBurnCards 焼き札 3 枚
func (cw *CasinoWar) GetBurnCards() []*Card { return cw.burnCards }

// GetPhase フェーズ
func (cw *CasinoWar) GetPhase() int { return cw.phase }

// GetGameEndFlag 終了フラグ
func (cw *CasinoWar) GetGameEndFlag() bool { return cw.gameEndFlag }

// GetAnte アンテ額
func (cw *CasinoWar) GetAnte() int { return cw.ante }

// GetWarBet ウォーベット額
func (cw *CasinoWar) GetWarBet() int { return cw.warBet }

// GetResult 結果
func (cw *CasinoWar) GetResult() GameResult { return cw.result }

// GetTotalPayout 合計配当
func (cw *CasinoWar) GetTotalPayout() int { return cw.totalPayout }

// GetChips チップ
func (cw *CasinoWar) GetChips() int { return cw.chips.GetChips() }

// --- Test helpers ---

// SetPhase テスト用
func (cw *CasinoWar) SetPhase(phase int) { cw.phase = phase }

// SetPlayerCard テスト用
func (cw *CasinoWar) SetPlayerCard(c *Card) { cw.playerCard = c }

// SetDealerCard テスト用
func (cw *CasinoWar) SetDealerCard(c *Card) { cw.dealerCard = c }

// SetPlayerWarCard テスト用
func (cw *CasinoWar) SetPlayerWarCard(c *Card) { cw.playerWarCard = c }

// SetDealerWarCard テスト用
func (cw *CasinoWar) SetDealerWarCard(c *Card) { cw.dealerWarCard = c }

// SetChips テスト用
func (cw *CasinoWar) SetChips(chips int) { cw.chips.SetChips(chips) }

// casinoWarJSON は CasinoWar の JSON ワイヤーフォーマット
type casinoWarJSON struct {
	TrumpCards    *TrumpCards       `json:"tc"`
	PlayerCard    *Card             `json:"pc"`
	DealerCard    *Card             `json:"dc"`
	PlayerWarCard *Card             `json:"pw"`
	DealerWarCard *Card             `json:"dw"`
	BurnCards     []*Card           `json:"bc"`
	Chips         *ChipHolder       `json:"ch"`
	Ante          int               `json:"an"`
	WarBet        int               `json:"wb"`
	Phase         int               `json:"ps"`
	GameEndFlag   bool              `json:"ge"`
	Result        GameResult        `json:"gr"`
	TotalPayout   int               `json:"tp"`
	ActionLog     []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (cw *CasinoWar) MarshalJSON() ([]byte, error) {
	return json.Marshal(casinoWarJSON{
		TrumpCards:    cw.trumpCards,
		PlayerCard:    cw.playerCard,
		DealerCard:    cw.dealerCard,
		PlayerWarCard: cw.playerWarCard,
		DealerWarCard: cw.dealerWarCard,
		BurnCards:     cw.burnCards,
		Chips:         &cw.chips,
		Ante:          cw.ante,
		WarBet:        cw.warBet,
		Phase:         cw.phase,
		GameEndFlag:   cw.gameEndFlag,
		Result:        cw.result,
		TotalPayout:   cw.totalPayout,
		ActionLog:     cw.actionLog,
	})
}

// casinoWarMaxSliceLen はデシリアライズ時のスライス長の上限
const casinoWarMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (cw *CasinoWar) UnmarshalJSON(data []byte) error {
	var j casinoWarJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.BurnCards) > casinoWarMaxSliceLen || len(j.ActionLog) > casinoWarMaxSliceLen {
		return fmt.Errorf("casinowar: input array exceeds maximum allowed size")
	}
	cw.trumpCards = j.TrumpCards
	if cw.trumpCards == nil {
		cw.trumpCards = NewTrumpCards(0)
	}
	cw.playerCard = j.PlayerCard
	cw.dealerCard = j.DealerCard
	cw.playerWarCard = j.PlayerWarCard
	cw.dealerWarCard = j.DealerWarCard
	cw.burnCards = j.BurnCards
	if cw.burnCards == nil {
		cw.burnCards = make([]*Card, 0)
	}
	if j.Chips != nil {
		cw.chips = *j.Chips
	}
	cw.ante = j.Ante
	cw.warBet = j.WarBet
	cw.phase = j.Phase
	cw.gameEndFlag = j.GameEndFlag
	cw.result = j.Result
	cw.totalPayout = j.TotalPayout
	cw.actionLog = j.ActionLog
	if cw.actionLog == nil {
		cw.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
