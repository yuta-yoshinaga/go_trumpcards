//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PasurWebInput パスールWebインプット
type PasurWebInput struct {
	BaseWebInput
	CardIndex *int            `json:"cardIndex,omitempty"`
	Table     []int           `json:"table,omitempty"`
	Config    *PasurWebConfig `json:"config,omitempty"`
}

// PasurWebConfig パスールWeb設定
type PasurWebConfig struct {
	PlayerCnt *int `json:"playerCnt,omitempty"`
}

// PasurWebOutputPlayer パスールWebアウトプットプレイヤー
type PasurWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// CapturedCount は捕獲した総枚数、Soors はスールの回数。
	CapturedCount int `json:"capturedCount"`
	Soors         int `json:"soors"`
	Score         int `json:"score"`
}

// PasurWebOutputHint ヒント出力
type PasurWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
	// Table は取るべき場札のインデックス（トレールなら空）。
	Table []int `json:"table"`
}

// PasurWebOutput パスールWebアウトプット
type PasurWebOutput struct {
	Players []*PasurWebOutputPlayer `json:"players"`
	Phase   int                     `json:"phase"`
	// Table は場の札。
	Table []*WebOutputCard `json:"table"`
	// CaptureOptions は手札インデックスごとに取れる場札の組み合わせ。
	//
	// **サーバが必ず拒否する操作をクライアントに出させないためにワイヤへ載せる。**
	// 11 の部分集合はページ側で作り直すと必ずズレます。
	CaptureOptions [][][]int `json:"captureOptions"`
	DeckRemaining  int       `json:"deckRemaining"`
	PacksDealt     int       `json:"packsDealt"`
	// LastCaptureIdx は最後に捕獲した席。**場に残った札はここへ行く。**
	LastCaptureIdx   int                 `json:"lastCaptureIdx"`
	CurrentPlayerIdx int                 `json:"currentPlayerIdx"`
	GameEndFlag      bool                `json:"gameEndFlag"`
	Winners          []int               `json:"winners"`
	Hint             *PasurWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config PasurWebOutputConfig `json:"config"`
}

// PasurWebOutputConfig パスール設定アウトプット
type PasurWebOutputConfig struct {
	PlayerCnt int `json:"playerCnt"`
}

// ToConfig builds a PasurConfig from the nested web config, applying bounds checking.
func (c *PasurWebConfig) ToConfig() domain.PasurConfig {
	cfg := DefaultPasurWebConfigValue()
	cfg.PlayerCnt = webutil.BoundedIntPtr(c.PlayerCnt,
		domain.PasurPlayerCntMin, domain.PasurPlayerCntMax, cfg.PlayerCnt)
	return cfg
}

// DefaultPasurWebConfigValue returns the domain default, kept in one place.
func DefaultPasurWebConfigValue() domain.PasurConfig { return domain.DefaultPasurConfig() }

// ToConfig builds a PasurConfig from the web input.
func (p PasurWebInput) ToConfig() domain.PasurConfig {
	return configOrDefault(p.Config, (*PasurWebConfig).ToConfig, domain.DefaultPasurConfig())
}

// PasurWebController パスールWebコントローラークラス
type PasurWebController = GameWebController[usecase.PasurInteractorIF, PasurWebInput, *PasurWebOutput]

// NewPasurWebController and NewPasurWebControllerWithProvider are
// the standard and provider-backed constructors for PasurWebController.
var NewPasurWebController, NewPasurWebControllerWithProvider = webControllerPair[usecase.PasurInteractorIF, PasurWebInput, *PasurWebOutput](
	newPasurDefaultOutput, pasurDispatch,
)

func newPasurDefaultOutput(msg string) *PasurWebOutput {
	return &PasurWebOutput{
		Players:        make([]*PasurWebOutputPlayer, 0),
		Table:          make([]*WebOutputCard, 0),
		CaptureOptions: make([][][]int, 0),
		Winners:        make([]int, 0),
		LastCaptureIdx: -1,
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func pasurDispatch(bc *baseController, w http.ResponseWriter, pi usecase.PasurInteractorIF, param PasurWebInput, newDefault func(string) *PasurWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, pi.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		// **`table` の省略はトレール（場に置く）。** 取れる手があるかどうかは
		// ドメインが判定するので、ここでは指定の有無だけをそのまま渡します。
		bc.writePresenterResponse(w, pi.Play(*param.CardIndex, param.Table))
	case "g", "giveup":
		bc.writePresenterResponse(w, pi.GiveUp())
	default:
		return dispatchHintAndLog(param.Command, bc, w, pi.Hint, pi.ActionLog)
	}
	return true
}
