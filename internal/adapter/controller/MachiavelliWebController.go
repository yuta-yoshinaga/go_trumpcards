//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MachiavelliCardRefInput は「新しい場」を指定するカード参照（デザイン＋数値）の Web 入力。
type MachiavelliCardRefInput struct {
	Design int `json:"design"`
	Value  int `json:"value"`
}

// MachiavelliWebInput マキャヴェッリ Web インプット
type MachiavelliWebInput struct {
	BaseWebInput
	// TableMelds は play コマンドで提出する「新しい場」全体（メルドごとのカード参照）。
	TableMelds [][]MachiavelliCardRefInput `json:"tableMelds,omitempty"`
	// HandIndices は play コマンドで場に追加する手札インデックス。
	HandIndices []int `json:"handIndices,omitempty"`
	// MeldIdx / HandIndex は layoff コマンド用。
	MeldIdx   *int                  `json:"meldIdx,omitempty"`
	HandIndex *int                  `json:"handIndex,omitempty"`
	Config    *MachiavelliWebConfig `json:"config,omitempty"`
}

// MachiavelliWebConfig マキャヴェッリ Web 設定
type MachiavelliWebConfig struct {
	PlayerCount   *int `json:"playerCount,omitempty"`
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetRounds  *int `json:"targetRounds,omitempty"`
}

// MachiavelliWebOutputMeld テーブル上のメルドのアウトプット
type MachiavelliWebOutputMeld struct {
	Cards []*WebOutputCard `json:"cards"`
	Kind  int              `json:"kind"` // 0=set, 1=run
}

// MachiavelliWebOutputPlayer プレイヤーのアウトプット
type MachiavelliWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
	Deadwood        int              `json:"deadwood"`
}

// MachiavelliWebOutput マキャヴェッリ Web アウトプット
type MachiavelliWebOutput struct {
	Players          []*MachiavelliWebOutputPlayer `json:"players"`
	Table            []*MachiavelliWebOutputMeld   `json:"table"`
	Phase            int                           `json:"phase"`
	RoundNumber      int                           `json:"roundNumber"`
	TargetRounds     int                           `json:"targetRounds"`
	CurrentPlayerIdx int                           `json:"currentPlayerIdx"`
	DealerIdx        int                           `json:"dealerIdx"`
	DrawPileCount    int                           `json:"drawPileCount"`
	GameEndFlag      bool                          `json:"gameEndFlag"`
	WinnerIdx        int                           `json:"winnerIdx"`
	RoundWinnerIdx   int                           `json:"roundWinnerIdx"`
	WebOutputBase
	Config MachiavelliWebOutputConfig `json:"config"`
}

// MachiavelliWebOutputConfig 設定アウトプット
type MachiavelliWebOutputConfig struct {
	PlayerCount   int `json:"playerCount"`
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetRounds  int `json:"targetRounds"`
}

// ToConfig builds a MachiavelliConfig from the nested web config, applying bounds checking.
func (c *MachiavelliWebConfig) ToConfig() domain.MachiavelliConfig {
	cfg := domain.DefaultMachiavelliConfig()
	webutil.ApplyBoundedInt(&cfg.PlayerCount, c.PlayerCount, domain.MachiavelliPlayerCountMin, domain.MachiavelliPlayerCountMax)
	cfg.CpuDifficulty = domain.MachiavelliCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.MachiavelliCpuDifficultyEasy),
		int(domain.MachiavelliCpuDifficultyHard),
		int(cfg.CpuDifficulty),
	))
	webutil.ApplyBoundedInt(&cfg.TargetRounds, c.TargetRounds, 1, 100)
	return cfg
}

// ToConfig builds a MachiavelliConfig from the web input.
func (p MachiavelliWebInput) ToConfig() domain.MachiavelliConfig {
	return configOrDefault(p.Config, (*MachiavelliWebConfig).ToConfig, domain.DefaultMachiavelliConfig())
}

// toDomainRefs converts the web table-meld input into domain card references.
func (p MachiavelliWebInput) toDomainRefs() [][]domain.MachiavelliCardRef {
	refs := make([][]domain.MachiavelliCardRef, len(p.TableMelds))
	for i, meld := range p.TableMelds {
		refs[i] = make([]domain.MachiavelliCardRef, len(meld))
		for j, c := range meld {
			refs[i][j] = domain.MachiavelliCardRef{Design: c.Design, Value: c.Value}
		}
	}
	return refs
}

// MachiavelliWebController マキャヴェッリ Web コントローラー
type MachiavelliWebController = GameWebController[usecase.MachiavelliInteractorIF, MachiavelliWebInput, *MachiavelliWebOutput]

// NewMachiavelliWebController / NewMachiavelliWebControllerWithProvider: 標準／provider 背後の 2 種類のコンストラクタ
var NewMachiavelliWebController, NewMachiavelliWebControllerWithProvider = webControllerPair[usecase.MachiavelliInteractorIF, MachiavelliWebInput, *MachiavelliWebOutput](
	newMachiavelliDefaultOutput, machiavelliDispatch,
)

func newMachiavelliDefaultOutput(msg string) *MachiavelliWebOutput {
	return &MachiavelliWebOutput{
		Players:        make([]*MachiavelliWebOutputPlayer, 0),
		Table:          make([]*MachiavelliWebOutputMeld, 0),
		WinnerIdx:      -1,
		RoundWinnerIdx: -1,
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func machiavelliDispatch(bc *baseController, w http.ResponseWriter, ci usecase.MachiavelliInteractorIF, param MachiavelliWebInput, newDefault func(string) *MachiavelliWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ci.ResetWithConfig(param.ToConfig()))
	case "dr", "draw":
		bc.writePresenterResponse(w, ci.Draw())
	case "p", "play":
		bc.writePresenterResponse(w, ci.Play(param.toDomainRefs(), param.HandIndices))
	case "nm", "newmeld":
		bc.writePresenterResponse(w, ci.NewMeld(param.HandIndices))
	case "lo", "layoff":
		if !requireParam(bc, w, newDefault, param.MeldIdx == nil || param.HandIndex == nil, "param error: meldIdx and handIndex are required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Layoff(*param.MeldIdx, *param.HandIndex))
	case "nr", "nextround":
		bc.writePresenterResponse(w, ci.NextRound())
	default:
		return dispatchLog(param.Command, bc, w, ci.ActionLog)
	}
	return true
}
