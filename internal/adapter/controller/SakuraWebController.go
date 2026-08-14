//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SakuraWebConfig はさくら (肥後花) の Web 設定。
type SakuraWebConfig struct {
	Seats  *int `json:"seats,omitempty"`
	Rounds *int `json:"rounds,omitempty"`
}

// ToConfig は SakuraWebConfig を domain.SakuraConfig に変換する (境界チェック付き)。
func (c *SakuraWebConfig) ToConfig() domain.SakuraConfig {
	cfg := domain.DefaultSakuraConfig()
	cfg.Seats = webutil.BoundedIntPtr(c.Seats, domain.SakuraMinSeats, domain.SakuraMaxSeats, cfg.Seats)
	cfg.Rounds = webutil.BoundedIntPtr(c.Rounds, domain.SakuraMinRounds, domain.SakuraMaxRounds, cfg.Rounds)
	return cfg
}

// SakuraWebInput はさくら Web インプット。
type SakuraWebInput struct {
	BaseWebInput
	CardIndex  *int             `json:"cardIndex,omitempty"`
	FieldIndex *int             `json:"fieldIndex,omitempty"`
	Config     *SakuraWebConfig `json:"config,omitempty"`
}

// ToConfig は Web インプットから domain.SakuraConfig を構築する。
func (p SakuraWebInput) ToConfig() domain.SakuraConfig {
	return configOrDefault(p.Config, (*SakuraWebConfig).ToConfig, domain.DefaultSakuraConfig())
}

// SakuraWebOutputBonus は成立した追加役 1 件。
type SakuraWebOutputBonus struct {
	Key    string `json:"key"`
	Points int    `json:"points"`
}

// SakuraWebOutputPlayer はさくら Web アウトプットプレイヤー。
type SakuraWebOutputPlayer struct {
	ID          int                     `json:"id"`
	Name        string                  `json:"name"`
	IsHuman     bool                    `json:"isHuman"`
	CardCount   int                     `json:"cardCount"`
	Cards       []*WebOutputCard        `json:"cards"`
	Taken       []*WebOutputCard        `json:"taken"`
	TakenCount  int                     `json:"takenCount"`
	CardPoints  int                     `json:"cardPoints"`
	Bonuses     []*SakuraWebOutputBonus `json:"bonuses"`
	BonusPoints int                     `json:"bonusPoints"`
	TotalPoints int                     `json:"totalPoints"`
	Score       int                     `json:"score"`
	RoundScore  int                     `json:"roundScore"`
	RoundWins   int                     `json:"roundWins"`
}

// SakuraWebOutputSeatResult は 1 席のラウンド結果。
type SakuraWebOutputSeatResult struct {
	CardPoints  int                     `json:"cardPoints"`
	Bonuses     []*SakuraWebOutputBonus `json:"bonuses"`
	BonusPoints int                     `json:"bonusPoints"`
	Total       int                     `json:"total"`
}

// SakuraWebOutputRoundResult は 1 ラウンドの結果。
type SakuraWebOutputRoundResult struct {
	Round  int                          `json:"round"`
	Winner int                          `json:"winner"`
	Seats  []*SakuraWebOutputSeatResult `json:"seats"`
}

// SakuraWebOutputHint はヒント出力。
type SakuraWebOutputHint struct {
	CardIndex  int    `json:"cardIndex"`
	FieldIndex int    `json:"fieldIndex"`
	Reason     string `json:"reason"`
}

// SakuraWebConfigOutput は設定アウトプット。
type SakuraWebConfigOutput struct {
	Seats  int `json:"seats"`
	Rounds int `json:"rounds"`
}

// SakuraWebOutput はさくら Web アウトプット。
type SakuraWebOutput struct {
	Players     []*SakuraWebOutputPlayer `json:"players"`
	Phase       int                      `json:"phase"`
	Round       int                      `json:"round"`
	TotalRounds int                      `json:"totalRounds"`
	CurrentTurn int                      `json:"currentTurn"`
	Dealer      int                      `json:"dealer"`
	FieldCards  []*WebOutputCard         `json:"fieldCards"`
	StockCount  int                      `json:"stockCount"`
	// CaptureOptions は手札インデックスごとに合わせられる場札インデックス。
	CaptureOptions map[int][]int `json:"captureOptions"`
	// ChoiceOptions は「取る札を選ぶ必要がある」手札だけの一覧。
	//
	// **合わせられる = 選ばせる、ではない。** 場に同月が 3 枚あるときは 4 枚
	// まとめて取るので、どれを押しても結果は変わらない ── 画面が枚数から
	// 判定し直すと「選べと言われたのに選択が効かない」表示になる。
	ChoiceOptions map[int][]int               `json:"choiceOptions"`
	Winner        int                         `json:"winner"`
	GameEndFlag   bool                        `json:"gameEndFlag"`
	IsHumanTurn   bool                        `json:"isHumanTurn"`
	LastResult    *SakuraWebOutputRoundResult `json:"lastResult"`
	Hint          *SakuraWebOutputHint        `json:"hint,omitempty"`
	WebOutputBase
	Config SakuraWebConfigOutput `json:"config"`
}

// SakuraWebController はさくら Web コントローラークラス。
type SakuraWebController = GameWebController[usecase.SakuraInteractorIF, SakuraWebInput, *SakuraWebOutput]

// NewSakuraWebController, NewSakuraWebControllerWithProvider are the standard and
// provider-backed constructors for SakuraWebController.
var NewSakuraWebController, NewSakuraWebControllerWithProvider = webControllerPair[usecase.SakuraInteractorIF, SakuraWebInput, *SakuraWebOutput](
	newSakuraDefaultOutput, sakuraDispatch,
)

func newSakuraDefaultOutput(msg string) *SakuraWebOutput {
	return &SakuraWebOutput{
		Players:        make([]*SakuraWebOutputPlayer, 0),
		FieldCards:     make([]*WebOutputCard, 0),
		CaptureOptions: make(map[int][]int),
		ChoiceOptions:  make(map[int][]int),
		Winner:         -1,
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func sakuraDispatch(bc *baseController, w http.ResponseWriter, si usecase.SakuraInteractorIF, param SakuraWebInput, newDefault func(string) *SakuraWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, si.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		fieldIdx := -1
		if param.FieldIndex != nil {
			fieldIdx = *param.FieldIndex
		}
		bc.writePresenterResponse(w, si.Play(*param.CardIndex, fieldIdx))
	case "nr", "nextround", "n", "next":
		bc.writePresenterResponse(w, si.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, si.Hint, si.ActionLog)
	}
	return true
}
