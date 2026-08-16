//go:build !js || !wasm || casino

package controller

import (
	"encoding/json"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PineappleWebInput パイナ���プルポーカーWebインプット
type PineappleWebInput struct {
	BaseWebInput
	PokerCommonInput
	PokerBlindsInput
	Amount      int             `json:"amount,omitempty"`
	HumanPlayMs int             `json:"humanPlayMs,omitempty"`
	CardIdx     *int            `json:"cardIdx,omitempty"`
	CardIdxs    []int           `json:"cardIdxs,omitempty"`
	CpuMetaAI   bool            `json:"cpuMetaAI,omitempty"`
	Profile     json.RawMessage `json:"profile,omitempty"`
}

// PineappleWebOutput パイナップルポーカーWebアウトプット
// HoldemWebOutput を埋め込み、Pineapple 固有フィールドのみ追加する。
type PineappleWebOutput struct {
	HoldemWebOutput
	IsDiscardPhase   bool   `json:"isDiscardPhase"`
	DiscardDone      []bool `json:"discardDone"`
	InitialDealCount int    `json:"initialDealCount"`
	// LiveBestHand は人間の暫定ベスト役のキー ("straightFlush" など)。ショーダウン前のベッティング
	// 中だけ埋まり、それ以外は空。
	//
	// Omaha の Web は同じ表示を TypeScript 側で組み直しているが、こちらは
	// サーバが答える。役の探索をフロントにもう 1 つ持つと、ドメインを直した
	// ときに片方だけ古くなる (#5601 で Agnes から同じ複製を消したばかり)。
	LiveBestHand string `json:"liveBestHand"`
}

// ToConfig builds a PineappleConfig from the web input.
func (p PineappleWebInput) ToConfig() (domain.PineappleConfig, error) {
	cfg := domain.DefaultPineappleConfig()
	if err := validateAndApplyBlinds(&cfg.SmallBlind, &cfg.BigBlind, p.SmallBlind, p.BigBlind, cfg.BigBlind); err != nil {
		return domain.PineappleConfig{}, err
	}
	applyBool(&cfg.TournamentMode, p.TournamentMode)
	applyIntIfGte(&cfg.BlindLevelHands, p.BlindLevelHands, 1)
	applyIntIfGte(&cfg.BlindMultiplier, p.BlindMultiplier, 101)
	applyBettingLimit(&cfg.BettingLimit, p.BettingLimit)
	if err := applyTableSize(&cfg.TableSize, p.TableSize, domain.IsValidHoldemTableSize, "param error: tableSize must be 4, 6, or 9"); err != nil {
		return domain.PineappleConfig{}, err
	}
	applyRebuyConfig(&cfg.RebuyEnabled, &cfg.RebuyMaxCount, &cfg.RebuyChips, &cfg.RebuyPeriodHands,
		p.RebuyEnabled, p.RebuyMaxCount, p.RebuyChips, p.RebuyPeriodHands)
	applyAddonConfig(&cfg.AddonEnabled, &cfg.AddonChips, &cfg.AddonAfterHand,
		p.AddonEnabled, p.AddonChips, p.AddonAfterHand)
	cfg.CpuMetaAI = p.CpuMetaAI
	return cfg, nil
}

// PineappleWebController パイナップルポーカーWebコントローラークラス
type PineappleWebController = GameWebController[usecase.PineappleInteractorIF, PineappleWebInput, *PineappleWebOutput]

// NewPineappleWebController and NewPineappleWebControllerWithProvider are
// the standard and provider-backed constructors for PineappleWebController.
var NewPineappleWebController, NewPineappleWebControllerWithProvider = webControllerPair[usecase.PineappleInteractorIF, PineappleWebInput, *PineappleWebOutput](
	newPineappleDefaultOutput, pineappleDispatch,
)

func newPineappleDefaultOutput(msg string) *PineappleWebOutput {
	return &PineappleWebOutput{
		HoldemWebOutput: HoldemWebOutput{
			Players:        make([]*HoldemWebOutputPlayer, 0),
			CommunityCards: make([]*WebOutputCard, 0),
			SidePots:       make([]*HoldemWebOutputSidePot, 0),
			RoundResults:   make([]*HoldemWebOutputResult, 0),
			CpuActions:     make([]*HoldemWebOutputCpuAction, 0),
			WebOutputBase:  WebOutputBase{Message: msg},
		},
	}
}

func pineappleDispatch(bc *baseController, w http.ResponseWriter, pgi usecase.PineappleInteractorIF, param PineappleWebInput, newDefault func(string) *PineappleWebOutput) bool {
	if dispatchPokerAction(bc, w, pgi, param.Command, param.Amount, param.HumanPlayMs) {
		return true
	}
	switch param.Command {
	case "r", "reset":
		cfg, err := param.ToConfig()
		if err != nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault(err.Error()))
			return true
		}
		bc.writePresenterResponse(w, pgi.ResetWithConfig(cfg, param.Profile))
	case "d", "discard":
		// 複数枚指定 (Irish Poker の2枚捨て) を優先。なければ単一 cardIdx。
		if len(param.CardIdxs) > 0 {
			bc.writePresenterResponse(w, pgi.DiscardMany(param.CardIdxs))
			return true
		}
		if !requireParam(bc, w, newDefault, param.CardIdx == nil, "param error: cardIdx is required for discard") {
			return true
		}
		bc.writePresenterResponse(w, pgi.Discard(*param.CardIdx))
	default:
		return dispatchLog(param.Command, bc, w, pgi.ActionLog)
	}
	return true
}
