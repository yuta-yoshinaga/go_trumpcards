//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MaoWebInput マオWebインプット
type MaoWebInput struct {
	BaseWebInput
	CardIndex *int          `json:"cardIndex,omitempty"`
	Suit      *int          `json:"suit,omitempty"`
	Word      *string       `json:"word,omitempty"`
	Config    *MaoWebConfig `json:"config,omitempty"`
}

// MaoWebConfig マオWeb設定
type MaoWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// MaoWebOutputPlayer マオWebアウトプットプレイヤー
type MaoWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
	HasDeclared     bool             `json:"hasDeclared"`
}

// MaoWebOutput マオWebアウトプット。
// 隠しルール (hiddenRule) は意図的に含めない。クライアントには
// awaitingWord / hintUnlocked / ruleHint / rulePenalty のみを公開する。
type MaoWebOutput struct {
	Players          []*MaoWebOutputPlayer `json:"players"`
	Phase            int                   `json:"phase"`
	RoundNumber      int                   `json:"roundNumber"`
	CurrentPlayerIdx int                   `json:"currentPlayerIdx"`
	DiscardTop       *WebOutputCard        `json:"discardTop"`
	DrawPileCount    int                   `json:"drawPileCount"`
	ChosenSuit       int                   `json:"chosenSuit"`
	PenaltyDrawCount int                   `json:"penaltyDrawCount"`
	Direction        int                   `json:"direction"`
	GameEndFlag      bool                  `json:"gameEndFlag"`
	WinnerIdx        int                   `json:"winnerIdx"`
	AwaitingWord     bool                  `json:"awaitingWord"`
	CorrectCount     int                   `json:"correctCount"`
	HintUnlocked     bool                  `json:"hintUnlocked"`
	RuleHint         string                `json:"ruleHint"`
	// RuleHintCode は RuleHint の i18n キー (`mao.` を除いた部分)。
	// フロントはこちらを翻訳する (#4917)。
	RuleHintCode string `json:"ruleHintCode,omitempty"`
	RulePenalty  bool   `json:"rulePenalty"`
	WebOutputBase
	Config MaoWebOutputConfig `json:"config"`
}

// MaoWebOutputConfig マオ設定アウトプット
type MaoWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// ToConfig builds a MaoConfig from the nested web config, applying bounds checking.
func (c *MaoWebConfig) ToConfig() domain.MaoConfig {
	cfg := domain.DefaultMaoConfig()
	cfg.CpuDifficulty = domain.MaoCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.MaoCpuDifficultyEasy), int(domain.MaoCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 1000)
	return cfg
}

// ToConfig builds a MaoConfig from the web input.
func (p MaoWebInput) ToConfig() domain.MaoConfig {
	return configOrDefault(p.Config, (*MaoWebConfig).ToConfig, domain.DefaultMaoConfig())
}

// MaoWebController マオWebコントローラークラス
type MaoWebController = GameWebController[usecase.MaoInteractorIF, MaoWebInput, *MaoWebOutput]

// NewMaoWebController and NewMaoWebControllerWithProvider are
// the standard and provider-backed constructors for MaoWebController.
var NewMaoWebController, NewMaoWebControllerWithProvider = webControllerPair[usecase.MaoInteractorIF, MaoWebInput, *MaoWebOutput](
	newMaoDefaultOutput, maoDispatch,
)

func newMaoDefaultOutput(msg string) *MaoWebOutput {
	return &MaoWebOutput{
		Players:       make([]*MaoWebOutputPlayer, 0),
		WinnerIdx:     -1,
		Direction:     1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func maoDispatch(bc *baseController, w http.ResponseWriter, ci usecase.MaoInteractorIF, param MaoWebInput, newDefault func(string) *MaoWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ci.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Play(*param.CardIndex))
	case "d", "draw":
		bc.writePresenterResponse(w, ci.Draw())
	case "s", "suit":
		if !requireParam(bc, w, newDefault, param.Suit == nil, "param error: suit is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.ChooseSuit(*param.Suit))
	case "dc", "declare":
		bc.writePresenterResponse(w, ci.Declare())
	case "sk", "skipdeclare":
		bc.writePresenterResponse(w, ci.SkipDeclare())
	case "dw", "declareword":
		if !requireParam(bc, w, newDefault, param.Word == nil, "param error: word is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.DeclareWord(*param.Word))
	case "nr", "nextround":
		bc.writePresenterResponse(w, ci.NextRound())
	default:
		return dispatchLog(param.Command, bc, w, ci.ActionLog)
	}
	return true
}
