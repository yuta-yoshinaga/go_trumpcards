//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ShengJiWebInput 升级 Webインプット
type ShengJiWebInput struct {
	BaseWebInput
	// Suit は亮牌するスート。**0 はパスという意味を持つ**ので、
	// 省略と区別するためにポインタで受ける。
	Suit *int `json:"suit,omitempty"`
	// CardIndexes は埋め戻す 8 枚、または出す手。
	CardIndexes []int             `json:"cardIndexes,omitempty"`
	Config      *ShengJiWebConfig `json:"config,omitempty"`
}

// ShengJiWebConfig 升级 Web設定
type ShengJiWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// ShengJiWebOutputPlayer 升级 Webアウトプットプレイヤー
type ShengJiWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// Team は 0/2 が 0、1/3 が 1。
	Team      int `json:"team"`
	CardCount int `json:"cardCount"`
	// Cards は自分の手札のみ。
	Cards []*WebOutputCard `json:"cards"`
	// IsDeclarer は宣言側か。
	IsDeclarer    bool `json:"isDeclarer"`
	IsCurrentTurn bool `json:"isCurrentTurn"`
}

// ShengJiWebOutputCombo 升级 Webアウトプット手の形
type ShengJiWebOutputCombo struct {
	Kind int `json:"kind"`
	Rank int `json:"rank"`
	Size int `json:"size"`
	// Trump は切札群の手か。
	Trump bool `json:"trump"`
	Suit  int  `json:"suit"`
}

// ShengJiWebOutputPlay 升级 Webアウトプット 1 人ぶんの手
type ShengJiWebOutputPlay struct {
	Seat  int              `json:"seat"`
	Cards []*WebOutputCard `json:"cards"`
}

// ShengJiWebOutputResult 升级 Webアウトプット局結果
type ShengJiWebOutputResult struct {
	DeclarerTeam int `json:"declarerTeam"`
	// DefenderPoints は守備側が集めた点。**点を集めるのは守備側。**
	DefenderPoints int `json:"defenderPoints"`
	// KittyPoints は底牌から守備側に入った点 (倍率適用後)。
	KittyPoints     int  `json:"kittyPoints"`
	KittyMultiplier int  `json:"kittyMultiplier"`
	DeclarerHeld    bool `json:"declarerHeld"`
	Advance         int  `json:"advance"`
	AdvancingTeam   int  `json:"advancingTeam"`
}

// ShengJiWebOutputDeclaration 升级 Webアウトプット亮牌
type ShengJiWebOutputDeclaration struct {
	Seat int `json:"seat"`
	Suit int `json:"suit"`
	// Strength は 1 が単張、2 が対子。**強い宣言だけが上書きできる。**
	Strength int `json:"strength"`
}

// ShengJiWebOutput 升级 Webアウトプット
type ShengJiWebOutput struct {
	Players          []*ShengJiWebOutputPlayer `json:"players"`
	Phase            int                       `json:"phase"`
	HandNumber       int                       `json:"handNumber"`
	CurrentPlayerIdx int                       `json:"currentPlayerIdx"`
	// Level はこの局の基準レベル。**このランクの札は全スートが切札になる。**
	Level int `json:"level"`
	// TeamLevels は各チームの現在レベル。
	TeamLevels   [domain.ShengJiTeamCnt]int `json:"teamLevels"`
	DeclarerTeam int                        `json:"declarerTeam"`
	// TrumpSuit は切札スート (0 は無主)。
	TrumpSuit   int                          `json:"trumpSuit"`
	Declaration *ShengJiWebOutputDeclaration `json:"declaration"`
	// DeclarableSuits は人間がいま亮牌できるスートと強さ。
	DeclarableSuits map[string]int `json:"declarableSuits"`
	// KittySize は底牌の枚数。**中身は終局まで送らない。**
	KittySize int              `json:"kittySize"`
	Kitty     []*WebOutputCard `json:"kitty"`
	// Trick はいまのトリックに出された手 (リード順)。
	Trick       []*ShengJiWebOutputPlay `json:"trick"`
	TrickLeader int                     `json:"trickLeader"`
	LeadCombo   *ShengJiWebOutputCombo  `json:"leadCombo"`
	// TeamPoints は各チームがこの局に集めた点。
	TeamPoints      [domain.ShengJiTeamCnt]int `json:"teamPoints"`
	TrickCount      int                        `json:"trickCount"`
	LastTrickWinner int                        `json:"lastTrickWinner"`
	LastResult      *ShengJiWebOutputResult    `json:"lastResult"`
	MinLevel        int                        `json:"minLevel"`
	MaxLevel        int                        `json:"maxLevel"`
	KittySizeMax    int                        `json:"kittySizeMax"`
	// TotalPoints は 1 局の総得点 (200)、DefenderTarget は守備側の目標 (80)。
	TotalPoints    int  `json:"totalPoints"`
	DefenderTarget int  `json:"defenderTarget"`
	AdvanceStep    int  `json:"advanceStep"`
	GameEndFlag    bool `json:"gameEndFlag"`
	WinnerTeam     int  `json:"winnerTeam"`
	WebOutputBase
	Config ShengJiWebOutputConfig `json:"config"`
}

// ShengJiWebOutputConfig 升级設定アウトプット
type ShengJiWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig builds a ShengJiConfig from the nested web config, applying bounds checking.
func (c *ShengJiWebConfig) ToConfig() domain.ShengJiConfig {
	cfg := domain.DefaultShengJiConfig()
	cfg.CpuDifficulty = domain.ShengJiCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.ShengJiCpuDifficultyNormal), int(domain.ShengJiCpuDifficultyNormal), int(cfg.CpuDifficulty)))
	return cfg
}

// ToConfig builds a ShengJiConfig from the web input.
func (p ShengJiWebInput) ToConfig() domain.ShengJiConfig {
	return configOrDefault(p.Config, (*ShengJiWebConfig).ToConfig, domain.DefaultShengJiConfig())
}

// ShengJiWebController 升级 Webコントローラークラス
type ShengJiWebController = GameWebController[usecase.ShengJiInteractorIF, ShengJiWebInput, *ShengJiWebOutput]

// NewShengJiWebController and NewShengJiWebControllerWithProvider are
// the standard and provider-backed constructors for ShengJiWebController.
var NewShengJiWebController, NewShengJiWebControllerWithProvider = webControllerPair[usecase.ShengJiInteractorIF, ShengJiWebInput, *ShengJiWebOutput](
	newShengJiDefaultOutput, shengJiDispatch,
)

func newShengJiDefaultOutput(msg string) *ShengJiWebOutput {
	return &ShengJiWebOutput{
		Players:         make([]*ShengJiWebOutputPlayer, 0),
		Trick:           make([]*ShengJiWebOutputPlay, 0),
		DeclarableSuits: map[string]int{},
		LastTrickWinner: -1,
		WinnerTeam:      -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func shengJiDispatch(bc *baseController, w http.ResponseWriter, si usecase.ShengJiInteractorIF, param ShengJiWebInput, newOut func(string) *ShengJiWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, si.ResetWithConfig(param.ToConfig()))
	case "d", "declare":
		// **0 はパス。**省略と区別が要る。
		if !requireParam(bc, w, newOut, param.Suit == nil, "param error: suit is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.Declare(*param.Suit))
	case "b", "bury":
		if !requireParam(bc, w, newOut, len(param.CardIndexes) == 0, "param error: cardIndexes is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.BuryKitty(param.CardIndexes))
	case "p", "play":
		if !requireParam(bc, w, newOut, len(param.CardIndexes) == 0, "param error: cardIndexes is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.Play(param.CardIndexes))
	case "n", "next":
		bc.writePresenterResponse(w, si.NextHand())
	default:
		return dispatchLog(param.Command, bc, w, si.ActionLog)
	}
	return true
}
