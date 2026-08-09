//go:build !js || !wasm || casino

package domain

// checkAndTransitionAddon はアドオン判定を行い、人間にアドオン決定を促す場合は true を返す。
// CPU は自動でアドオンを実行する。
func (h *Holdem) checkAndTransitionAddon() bool {
	if h.config.AddonEnabled && h.handCount == h.config.AddonAfterHand {
		needHumanAddon := false
		for i, p := range h.players {
			if !h.addonUsed[i] {
				if p.GetIsHuman() {
					needHumanAddon = true
				} else {
					p.AddChips(h.config.AddonChips)
					h.addonUsed[i] = true
				}
			}
		}
		if needHumanAddon {
			h.phase = HoldemPhaseRebuy
			h.rebuyPhaseType = HoldemRebuyPhaseAddon
			return true
		}
	}
	return false
}

// Rebuy 人間プレイヤーがリバイを実行する
func (h *Holdem) Rebuy() error {
	if h.phase != HoldemPhaseRebuy || h.rebuyPhaseType != HoldemRebuyPhaseRebuy {
		return NewDomainError(ErrWrongPhase, "Rebuy is not available now.")
	}
	for i, p := range h.players {
		if p.GetIsHuman() && p.GetChips() <= 0 && h.rebuyCounts[i] < h.config.RebuyMaxCount {
			p.AddChips(h.config.RebuyChips)
			h.rebuyCounts[i]++
			h.appendLog(i, "rebuy", "rebuy", nil)
			break
		}
	}
	h.rebuyPhaseType = HoldemRebuyPhaseNone
	if h.checkAndTransitionAddon() {
		return nil
	}
	return h.continueReset()
}

// SkipRebuy 人間プレイヤーがリバイを辞退する
func (h *Holdem) SkipRebuy() error {
	if h.phase != HoldemPhaseRebuy || h.rebuyPhaseType != HoldemRebuyPhaseRebuy {
		return NewDomainError(ErrWrongPhase, "Rebuy is not available now.")
	}
	h.rebuyPhaseType = HoldemRebuyPhaseNone
	// バスト中の人間がリバイを辞退 → ゲーム終了
	for _, p := range h.players {
		if p.GetIsHuman() && p.GetChips() <= 0 {
			h.phase = HoldemPhaseEnd
			h.gameEndFlag = true
			return nil
		}
	}
	// 人間にチップが残っている場合 (通常ありえないが安全策)
	if h.checkAndTransitionAddon() {
		return nil
	}
	return h.continueReset()
}

// Addon 人間プレイヤーがアドオンを実行する
func (h *Holdem) Addon() error {
	if h.phase != HoldemPhaseRebuy || h.rebuyPhaseType != HoldemRebuyPhaseAddon {
		return NewDomainError(ErrWrongPhase, "Addon is not available now.")
	}
	for i, p := range h.players {
		if p.GetIsHuman() && !h.addonUsed[i] {
			p.AddChips(h.config.AddonChips)
			h.addonUsed[i] = true
			break
		}
	}
	h.rebuyPhaseType = HoldemRebuyPhaseNone
	return h.continueReset()
}

// SkipAddon 人間プレイヤーがアドオンを辞退する
func (h *Holdem) SkipAddon() error {
	if h.phase != HoldemPhaseRebuy || h.rebuyPhaseType != HoldemRebuyPhaseAddon {
		return NewDomainError(ErrWrongPhase, "Addon is not available now.")
	}
	h.rebuyPhaseType = HoldemRebuyPhaseNone
	return h.continueReset()
}

// IsRebuyAvailable 人間プレイヤーがリバイ可能かどうか
func (h *Holdem) IsRebuyAvailable() bool {
	return rebuyAvailable(h.config.RebuyEnabled, h.handCount, h.config.RebuyPeriodHands, h.players, h.rebuyCounts, h.config.RebuyMaxCount)
}

// IsAddonAvailable 人間プレイヤーがアドオン可能かどうか
func (h *Holdem) IsAddonAvailable() bool {
	return addonAvailable(h.config.AddonEnabled, h.handCount, h.config.AddonAfterHand, h.players, h.addonUsed)
}

// GetRebuyCounts プレイヤーごとのリバイ回数取得
func (h *Holdem) GetRebuyCounts() []int {
	return copyOf(h.rebuyCounts)
}

// GetAddonUsed プレイヤーごとのアドオン使用フラグ取得
func (h *Holdem) GetAddonUsed() []bool {
	return copyOf(h.addonUsed)
}

// GetRebuyPhaseType リバイフェーズ種別取得
func (h *Holdem) GetRebuyPhaseType() int { return h.rebuyPhaseType }
