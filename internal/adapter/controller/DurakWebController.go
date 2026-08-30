package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DurakWebConfig ドゥラーク設定 (入力・出力共用)
type DurakWebConfig struct {
	PlayerCount     int  `json:"playerCount"`
	CpuDifficulty   int  `json:"cpuDifficulty"`
	TransferEnabled bool `json:"transferEnabled"`
}

// ToConfig converts DurakWebConfig to domain.DurakConfig.
func (c DurakWebConfig) ToConfig() domain.DurakConfig {
	return domain.DurakConfig{
		PlayerCount:     c.PlayerCount,
		CpuDifficulty:   domain.DurakCpuDifficulty(c.CpuDifficulty),
		TransferEnabled: c.TransferEnabled,
	}
}

// DurakWebInput ドゥラークWebインプット
type DurakWebInput struct {
	BaseWebInput
	CardIdx   *int            `json:"cardIdx"`   // 攻撃/防御カードインデックス
	AttackIdx *int            `json:"attackIdx"` // 防御対象の攻撃カードインデックス
	Config    *DurakWebConfig `json:"config"`    // リセット時の設�� (省略可)
	SortMode  *int            `json:"sortMode"`  // ソートモード (省略可)
}

// DurakWebOutputPlayer ドゥラークWebアウトプットプレイヤー
type DurakWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	IsFinished bool             `json:"isFinished"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
}

// DurakWebOutputTablePair テーブルペア出力
type DurakWebOutputTablePair struct {
	Attack  *WebOutputCard `json:"attack"`
	Defense *WebOutputCard `json:"defense"`
}

// DurakWebOutputAction ドゥラークのプレイヤー行動記録
type DurakWebOutputAction struct {
	PlayerIdx  int            `json:"playerIdx"`
	ActionType int            `json:"actionType"` // 0=attack, 1=defend, 2=pass, 3=take, 4=transfer
	Card       *WebOutputCard `json:"card"`       // 出したカード (nil = パス/テイク)
	AttackIdx  int            `json:"attackIdx"`  // 防御時: 対象攻撃カード (-1 = 該当なし)
}

// DurakWebOutput ドゥラークWebアウトプット
type DurakWebOutput struct {
	Players     []*DurakWebOutputPlayer    `json:"players"`
	CurrentTurn int                        `json:"currentTurn"`
	Phase       int                        `json:"phase"`
	AttackerIdx int                        `json:"attackerIdx"`
	DefenderIdx int                        `json:"defenderIdx"`
	TablePairs  []*DurakWebOutputTablePair `json:"tablePairs"`
	TrumpSuit   string                     `json:"trumpSuit"`
	TrumpCard   *WebOutputCard             `json:"trumpCard"`
	StockCount  int                        `json:"stockCount"`
	LoserIdx    int                        `json:"loserIdx"`
	GameEndFlag bool                       `json:"gameEndFlag"`
	Config      DurakWebConfig             `json:"config"`
	CpuActions  []*DurakWebOutputAction    `json:"cpuActions"`
	HumanAction *DurakWebOutputAction      `json:"humanAction"`
	BoutNumber  int                        `json:"boutNumber"`
	SortMode    int                        `json:"sortMode"`
	WebOutputBase
}

// DurakWebController ドゥラークWebコントローラークラス
type DurakWebController = GameWebController[usecase.DurakInteractorIF, DurakWebInput, *DurakWebOutput]

// NewDurakWebController and NewDurakWebControllerWithProvider are
// the standard and provider-backed constructors for DurakWebController.
var NewDurakWebController, NewDurakWebControllerWithProvider = webControllerPair[usecase.DurakInteractorIF, DurakWebInput, *DurakWebOutput](
	newDurakDefaultOutput, durakDispatch,
)

func newDurakDefaultOutput(msg string) *DurakWebOutput {
	return &DurakWebOutput{
		Players:       make([]*DurakWebOutputPlayer, 0),
		TablePairs:    make([]*DurakWebOutputTablePair, 0),
		CpuActions:    make([]*DurakWebOutputAction, 0),
		LoserIdx:      -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func durakDispatch(bc *baseController, w http.ResponseWriter, di usecase.DurakInteractorIF, param DurakWebInput, _ func(string) *DurakWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.Config != nil {
			bc.writePresenterResponse(w, di.ResetWithConfig(param.Config.ToConfig()))
		} else {
			bc.writePresenterResponse(w, di.Reset())
		}
	case "a", "attack":
		idx := 0
		if param.CardIdx != nil {
			idx = *param.CardIdx
		}
		bc.writePresenterResponse(w, di.Attack(idx))
	case "d", "defend":
		atkIdx := 0
		handIdx := 0
		if param.AttackIdx != nil {
			atkIdx = *param.AttackIdx
		}
		if param.CardIdx != nil {
			handIdx = *param.CardIdx
		}
		bc.writePresenterResponse(w, di.Defend(atkIdx, handIdx))
	case "p", "pass":
		bc.writePresenterResponse(w, di.Pass())
	case "t", "take":
		bc.writePresenterResponse(w, di.TakeCards())
	case "tr", "transfer":
		handIdx := 0
		if param.CardIdx != nil {
			handIdx = *param.CardIdx
		}
		bc.writePresenterResponse(w, di.Transfer(handIdx))
	case "sort":
		mode := domain.DurakSortBySuit
		if param.SortMode != nil {
			if m := *param.SortMode; m >= int(domain.DurakSortBySuit) && m <= int(domain.DurakSortByValue) {
				mode = domain.DurakSortMode(m)
			}
		}
		bc.writePresenterResponse(w, di.Sort(mode))
	default:
		return dispatchHintAndLog(param.Command, bc, w, di.Hint, di.ActionLog)
	}
	return true
}
