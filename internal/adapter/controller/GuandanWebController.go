//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GuandanWebInput 掼蛋 Webインプット
type GuandanWebInput struct {
	BaseWebInput
	// CardIndexes は出す手札のインデックス。
	CardIndexes []int `json:"cardIndexes,omitempty"`
	// CardIndex は還貢で返す 1 枚。
	CardIndex *int              `json:"cardIndex,omitempty"`
	Config    *GuandanWebConfig `json:"config,omitempty"`
}

// GuandanWebConfig 掼蛋 Web設定
type GuandanWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// GuandanWebOutputPlayer 掼蛋 Webアウトプットプレイヤー
type GuandanWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// Team は 0/2 が 0、1/3 が 1。
	Team      int `json:"team"`
	CardCount int `json:"cardCount"`
	// Cards は自分の手札のみ。
	Cards []*WebOutputCard `json:"cards"`
	// FinishedRank は上がった順位 (まだなら 0)。
	FinishedRank  int  `json:"finishedRank"`
	IsCurrentTurn bool `json:"isCurrentTurn"`
}

// GuandanWebOutputCombo 掼蛋 Webアウトプット役
type GuandanWebOutputCombo struct {
	Kind int `json:"kind"`
	Rank int `json:"rank"`
	Size int `json:"size"`
}

// GuandanWebOutputTribute 掼蛋 Webアウトプット進貢
type GuandanWebOutputTribute struct {
	From int            `json:"from"`
	To   int            `json:"to"`
	Card *WebOutputCard `json:"card"`
	// Returned は還貢で返された札 (まだなら null)。
	Returned *WebOutputCard `json:"returned"`
}

// GuandanWebOutputResult 掼蛋 Webアウトプット局結果
type GuandanWebOutputResult struct {
	// Order は上がった順の席。
	Order [domain.GuandanPlayerCnt]int `json:"order"`
	// WinnerTeam は 1 着の側。
	WinnerTeam int `json:"winnerTeam"`
	// Advance は上昇したレベル数。**1 / 2 / 4 のいずれか。**
	Advance int `json:"advance"`
	// FirstSecond は上位独占だったか (このときだけ +4)。
	FirstSecond bool `json:"firstSecond"`
}

// GuandanWebOutput 掼蛋 Webアウトプット
type GuandanWebOutput struct {
	Players          []*GuandanWebOutputPlayer `json:"players"`
	Phase            int                       `json:"phase"`
	HandNumber       int                       `json:"handNumber"`
	CurrentPlayerIdx int                       `json:"currentPlayerIdx"`
	// Level はこの局の基準レベル。**このランクが A の上に割り込む。**
	Level int `json:"level"`
	// TeamLevels は各チームの現在レベル。
	TeamLevels   [domain.GuandanTeamCnt]int `json:"teamLevels"`
	DeclarerTeam int                        `json:"declarerTeam"`
	// LastCombo は場に出ている最後の役。
	LastCombo     *GuandanWebOutputCombo `json:"lastCombo"`
	LastPlayerIdx int                    `json:"lastPlayerIdx"`
	Finished      []int                  `json:"finished"`
	// Tributes はこの局の進貢。**前局の結果が次局の手札を動かす。**
	Tributes []*GuandanWebOutputTribute `json:"tributes"`
	// TributeCancelled は赤ジョーカー保持で貢が取り消されたか。
	TributeCancelled bool                    `json:"tributeCancelled"`
	LastResult       *GuandanWebOutputResult `json:"lastResult"`
	// MinLevel / MaxLevel はレベルの範囲 (2 と A)。
	MinLevel int `json:"minLevel"`
	MaxLevel int `json:"maxLevel"`
	// AdvanceFirstSecond などは上昇量の表 (**1 / 2 / 4**)。
	AdvanceFirstSecond int  `json:"advanceFirstSecond"`
	AdvanceFirstThird  int  `json:"advanceFirstThird"`
	AdvanceFirstFourth int  `json:"advanceFirstFourth"`
	GameEndFlag        bool `json:"gameEndFlag"`
	WinnerTeam         int  `json:"winnerTeam"`
	WebOutputBase
	Config GuandanWebOutputConfig `json:"config"`
}

// GuandanWebOutputConfig 掼蛋設定アウトプット
type GuandanWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig builds a GuandanConfig from the nested web config, applying bounds checking.
func (c *GuandanWebConfig) ToConfig() domain.GuandanConfig {
	cfg := domain.DefaultGuandanConfig()
	cfg.CpuDifficulty = domain.GuandanCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.GuandanCpuDifficultyNormal), int(domain.GuandanCpuDifficultyNormal), int(cfg.CpuDifficulty)))
	return cfg
}

// ToConfig builds a GuandanConfig from the web input.
func (p GuandanWebInput) ToConfig() domain.GuandanConfig {
	return configOrDefault(p.Config, (*GuandanWebConfig).ToConfig, domain.DefaultGuandanConfig())
}

// GuandanWebController 掼蛋 Webコントローラークラス
type GuandanWebController = GameWebController[usecase.GuandanInteractorIF, GuandanWebInput, *GuandanWebOutput]

// NewGuandanWebController and NewGuandanWebControllerWithProvider are
// the standard and provider-backed constructors for GuandanWebController.
var NewGuandanWebController, NewGuandanWebControllerWithProvider = webControllerPair[usecase.GuandanInteractorIF, GuandanWebInput, *GuandanWebOutput](
	newGuandanDefaultOutput, guandanDispatch,
)

func newGuandanDefaultOutput(msg string) *GuandanWebOutput {
	return &GuandanWebOutput{
		Players:       make([]*GuandanWebOutputPlayer, 0),
		Finished:      make([]int, 0),
		Tributes:      make([]*GuandanWebOutputTribute, 0),
		LastPlayerIdx: -1,
		WinnerTeam:    -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func guandanDispatch(bc *baseController, w http.ResponseWriter, gi usecase.GuandanInteractorIF, param GuandanWebInput, newOut func(string) *GuandanWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, gi.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newOut, len(param.CardIndexes) == 0, "param error: cardIndexes is required.") {
			return true
		}
		bc.writePresenterResponse(w, gi.PlayCards(param.CardIndexes))
	case "ps", "pass":
		bc.writePresenterResponse(w, gi.Pass())
	case "t", "tribute":
		if !requireParam(bc, w, newOut, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, gi.ReturnTribute(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, gi.NextHand())
	default:
		return dispatchLog(param.Command, bc, w, gi.ActionLog)
	}
	return true
}
