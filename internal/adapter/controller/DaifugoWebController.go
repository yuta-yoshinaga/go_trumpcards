package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"net/http"
)

// DaifugoWebConfig ローカルルール設定 (入力・出力共用)
type DaifugoWebConfig struct {
	JokerCount                int  `json:"jokerCount"`
	EightCutEnabled           bool `json:"eightCutEnabled"`
	SuitLockMode              int  `json:"suitLockMode"`
	ElevenBackEnabled         bool `json:"elevenBackEnabled"`
	SequenceEnabled           bool `json:"sequenceEnabled"`
	CardExchangeEnabled       bool `json:"cardExchangeEnabled"`
	BlindExchangeEnabled      bool `json:"blindExchangeEnabled"`
	FiveSkipEnabled           bool `json:"fiveSkipEnabled"`
	FiveSkipCount             int  `json:"fiveSkipCount"`
	SevenPassEnabled          bool `json:"sevenPassEnabled"`
	TenDiscardEnabled         bool `json:"tenDiscardEnabled"`
	SpadeThreeEnabled         bool `json:"spadeThreeEnabled"`
	CapitalFallEnabled        bool `json:"capitalFallEnabled"`
	NineReverseEnabled        bool `json:"nineReverseEnabled"`
	CoupDetatEnabled          bool `json:"coupDetatEnabled"`
	NumberLockEnabled         bool `json:"numberLockEnabled"`
	SandstormEnabled          bool `json:"sandstormEnabled"`
	EmperorEnabled            bool `json:"emperorEnabled"`
	SequenceRevolutionEnabled bool `json:"sequenceRevolutionEnabled"`
	SequenceLockEnabled       bool `json:"sequenceLockEnabled"`
	IllegalFinishEnabled      bool `json:"illegalFinishEnabled"`
	QueenBomberEnabled        bool `json:"queenBomberEnabled"`
	CpuDifficulty             int  `json:"cpuDifficulty"`
}

// ToConfig converts DaifugoWebConfig to domain.DaifugoConfig.
func (c DaifugoWebConfig) ToConfig() domain.DaifugoConfig {
	return domain.DaifugoConfig{
		JokerCount:                c.JokerCount,
		EightCutEnabled:           c.EightCutEnabled,
		SuitLockMode:              domain.DaifugoSuitLockMode(c.SuitLockMode),
		ElevenBackEnabled:         c.ElevenBackEnabled,
		SequenceEnabled:           c.SequenceEnabled,
		CardExchangeEnabled:       c.CardExchangeEnabled,
		BlindExchangeEnabled:      c.BlindExchangeEnabled,
		FiveSkipEnabled:           c.FiveSkipEnabled,
		FiveSkipCount:             c.FiveSkipCount,
		SevenPassEnabled:          c.SevenPassEnabled,
		TenDiscardEnabled:         c.TenDiscardEnabled,
		SpadeThreeEnabled:         c.SpadeThreeEnabled,
		CapitalFallEnabled:        c.CapitalFallEnabled,
		NineReverseEnabled:        c.NineReverseEnabled,
		CoupDetatEnabled:          c.CoupDetatEnabled,
		NumberLockEnabled:         c.NumberLockEnabled,
		SandstormEnabled:          c.SandstormEnabled,
		EmperorEnabled:            c.EmperorEnabled,
		SequenceRevolutionEnabled: c.SequenceRevolutionEnabled,
		SequenceLockEnabled:       c.SequenceLockEnabled,
		IllegalFinishEnabled:      c.IllegalFinishEnabled,
		QueenBomberEnabled:        c.QueenBomberEnabled,
		CpuDifficulty:             domain.DaifugoCpuDifficulty(c.CpuDifficulty),
	}
}

// DaifugoWebInput 大富豪Webインプット
type DaifugoWebInput struct {
	BaseWebInput
	Indices  []int             `json:"indices"`  // 出すカードのインデックス。play コマンド用。空の場合はパス。
	Config   *DaifugoWebConfig `json:"config"`   // リセット時のローカルルール設定 (省略可)
	SortMode *int              `json:"sortMode"` // ソートモード (sort コマンド用、省略可)
}

// DaifugoWebOutputPlayer 大富豪Webアウトプットプレイヤー
type DaifugoWebOutputPlayer struct {
	ID                   int              `json:"id"`
	IsHuman              bool             `json:"isHuman"`
	IsFinished           bool             `json:"isFinished"`
	Rank                 int              `json:"rank"`
	CardCount            int              `json:"cardCount"`
	Cards                []*WebOutputCard `json:"cards"`
	IllegalFinishPenalty bool             `json:"illegalFinishPenalty"`
}

// DaifugoWebOutputAction 大富豪のプレイヤー行動記録
type DaifugoWebOutputAction struct {
	PlayerIdx   int              `json:"playerIdx"`
	PlayedCards []*WebOutputCard `json:"playedCards"` // nil = パス
}

// DaifugoWebOutputExchangeAction カード交換記録
type DaifugoWebOutputExchangeAction struct {
	FromPlayerIdx int              `json:"fromPlayerIdx"`
	ToPlayerIdx   int              `json:"toPlayerIdx"`
	Cards         []*WebOutputCard `json:"cards"`
}

// DaifugoWebOutput 大富豪Webアウトプット
type DaifugoWebOutput struct {
	Players             []*DaifugoWebOutputPlayer         `json:"players"`
	CurrentTurn         int                               `json:"currentTurn"`
	TableCards          []*WebOutputCard                  `json:"tableCards"`
	LastPlayPlayerIdx   int                               `json:"lastPlayPlayerIdx"`
	GameEndFlag         bool                              `json:"gameEndFlag"`
	RevolutionActive    bool                              `json:"revolutionActive"`
	ElevenBackActive    bool                              `json:"elevenBackActive"`
	SuitLocked          bool                              `json:"suitLocked"`
	LockedSuit          string                            `json:"lockedSuit"`
	TableIsSequence     bool                              `json:"tableIsSequence"`
	Config              DaifugoWebConfig                  `json:"config"`
	ExchangeActions     []*DaifugoWebOutputExchangeAction `json:"exchangeActions"`
	CpuActions          []*DaifugoWebOutputAction         `json:"cpuActions"`
	HumanAction         *DaifugoWebOutputAction           `json:"humanAction"`
	PendingAction       string                            `json:"pendingAction"`       // "none"|"sevenPass"|"tenDiscard"|"queenBomber"
	PendingActionTarget int                               `json:"pendingActionTarget"` // -1 if none
	ReverseDirection    bool                              `json:"reverseDirection"`
	NumberLocked        bool                              `json:"numberLocked"`
	SequenceLocked      bool                              `json:"sequenceLocked"`
	SortMode            int                               `json:"sortMode"`
	// PlayableCardIndices は人間の手札のうち、いま出せる組み合わせに1つでも
	// 含まれるインデックス。CUI の `*` 印と同じ domain.GetPlayableCardIndices の
	// 結果をそのまま送る (#5477)。
	//
	// **判定できないときは空配列ではなく null。**空配列は「1枚も出せない」と
	// 読めてしまい、実際には出せる手札を全部「出せない」と表示することになる。
	PlayableCardIndices []int `json:"playableCardIndices"`
	WebOutputBase
}

// DaifugoWebController 大富豪Webコントローラークラス
type DaifugoWebController = GameWebController[usecase.DaifugoInteractorIF, DaifugoWebInput, *DaifugoWebOutput]

// NewDaifugoWebController and NewDaifugoWebControllerWithProvider are
// the standard and provider-backed constructors for DaifugoWebController.
var NewDaifugoWebController, NewDaifugoWebControllerWithProvider = webControllerPair[usecase.DaifugoInteractorIF, DaifugoWebInput, *DaifugoWebOutput](
	newDaifugoDefaultOutput, daifugoDispatch,
)

func newDaifugoDefaultOutput(msg string) *DaifugoWebOutput {
	return &DaifugoWebOutput{
		Players:             make([]*DaifugoWebOutputPlayer, 0),
		TableCards:          make([]*WebOutputCard, 0),
		CpuActions:          make([]*DaifugoWebOutputAction, 0),
		ExchangeActions:     make([]*DaifugoWebOutputExchangeAction, 0),
		WebOutputBase:       WebOutputBase{Message: msg},
		PendingAction:       "none",
		PendingActionTarget: -1,
	}
}

func daifugoDispatch(bc *baseController, w http.ResponseWriter, dgi usecase.DaifugoInteractorIF, param DaifugoWebInput, _ func(string) *DaifugoWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.Config != nil {
			bc.writePresenterResponse(w, dgi.ResetWithConfig(param.Config.ToConfig()))
		} else {
			bc.writePresenterResponse(w, dgi.Reset())
		}
	case "p", "play":
		indices := param.Indices
		if indices == nil {
			indices = []int{}
		}
		bc.writePresenterResponse(w, dgi.Play(indices))
	case "sort":
		mode := domain.DaifugoSortByStrength
		if param.SortMode != nil {
			if m := *param.SortMode; m >= int(domain.DaifugoSortByStrength) && m <= int(domain.DaifugoSortByNumber) {
				mode = domain.DaifugoSortMode(m)
			}
		}
		bc.writePresenterResponse(w, dgi.Sort(mode))
	default:
		return dispatchLog(param.Command, bc, w, dgi.ActionLog)
	}
	return true
}
