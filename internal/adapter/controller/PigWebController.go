//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PigWebInput ピッグWebインプット
type PigWebInput struct {
	BaseWebInput
	CardIndex *int          `json:"cardIndex,omitempty"`
	Config    *PigWebConfig `json:"config,omitempty"`
}

// PigWebConfig ピッグWeb設定
type PigWebConfig struct {
	PlayerCnt     *int `json:"playerCnt,omitempty"`
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// PigWebOutputPlayer ピッグWebアウトプットプレイヤー
type PigWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// Letters は溜まった文字数。3 で脱落。
	Letters int `json:"letters"`
	// LetterWord は溜まった文字そのもの ("P" / "PI" / "PIG")。
	LetterWord string `json:"letterWord"`
	Eliminated bool   `json:"eliminated"`
	// HasSignalled はこのラウンドで合図に気づいたか。
	HasSignalled bool `json:"hasSignalled"`
	// NoticedOrder は気づいた順 (1 始まり、0 = まだ)。
	NoticedOrder int `json:"noticedOrder"`
	// HasChosenPass は渡す札を選び終えたか。**同時に渡すので順番待ちが出ます。**
	HasChosenPass bool `json:"hasChosenPass"`
}

// PigWebOutputHint ヒント出力
type PigWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
}

// PigWebOutput ピッグWebアウトプット
type PigWebOutput struct {
	Players    []*PigWebOutputPlayer `json:"players"`
	Phase      int                   `json:"phase"`
	ValidPlays []int                 `json:"validPlays"`
	// SignallerIdx は最初に合図した席 (-1 = 合図なし)。
	SignallerIdx     int               `json:"signallerIdx"`
	NoticedCnt       int               `json:"noticedCnt"`
	RoundLoserIdx    int               `json:"roundLoserIdx"`
	RoundNumber      int               `json:"roundNumber"`
	PassCount        int               `json:"passCount"`
	DeckSize         int               `json:"deckSize"`
	CurrentPlayerIdx int               `json:"currentPlayerIdx"`
	GameEndFlag      bool              `json:"gameEndFlag"`
	WinnerIdx        int               `json:"winnerIdx"`
	Hint             *PigWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config PigWebOutputConfig `json:"config"`
}

// PigWebOutputConfig ピッグ設定アウトプット
type PigWebOutputConfig struct {
	PlayerCnt     int `json:"playerCnt"`
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig builds a PigConfig from the nested web config, applying bounds checking.
func (c *PigWebConfig) ToConfig() domain.PigConfig {
	cfg := DefaultPigConfigForWeb()
	cfg.PlayerCnt = webutil.BoundedIntPtr(c.PlayerCnt,
		domain.PigPlayerCntMin, domain.PigPlayerCntMax, cfg.PlayerCnt)
	cfg.CpuDifficulty = domain.PigCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.PigCpuEasy), int(domain.PigCpuHard), int(cfg.CpuDifficulty)))
	return cfg
}

// DefaultPigConfigForWeb は Web 側の既定設定を返す。
func DefaultPigConfigForWeb() domain.PigConfig { return domain.DefaultPigConfig() }

// ToConfig builds a PigConfig from the web input.
func (p PigWebInput) ToConfig() domain.PigConfig {
	return configOrDefault(p.Config, (*PigWebConfig).ToConfig, domain.DefaultPigConfig())
}

// PigWebController ピッグWebコントローラークラス
type PigWebController = GameWebController[usecase.PigInteractorIF, PigWebInput, *PigWebOutput]

// NewPigWebController and NewPigWebControllerWithProvider are the standard and
// provider-backed constructors for PigWebController.
var NewPigWebController, NewPigWebControllerWithProvider = webControllerPair[usecase.PigInteractorIF, PigWebInput, *PigWebOutput](
	newPigDefaultOutput, pigDispatch,
)

func newPigDefaultOutput(msg string) *PigWebOutput {
	return &PigWebOutput{
		Players:       make([]*PigWebOutputPlayer, 0),
		ValidPlays:    make([]int, 0),
		SignallerIdx:  -1,
		RoundLoserIdx: -1,
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func pigDispatch(bc *baseController, w http.ResponseWriter, pi usecase.PigInteractorIF, param PigWebInput, newDefault func(string) *PigWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, pi.ResetWithConfig(param.ToConfig()))
	case "p", "pass":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, pi.Pass(*param.CardIndex))
	case "s", "signal":
		bc.writePresenterResponse(w, pi.Signal())
	case "n", "next":
		bc.writePresenterResponse(w, pi.NextRound())
	case "g", "giveup":
		bc.writePresenterResponse(w, pi.GiveUp())
	default:
		return dispatchHintAndLog(param.Command, bc, w, pi.Hint, pi.ActionLog)
	}
	return true
}
