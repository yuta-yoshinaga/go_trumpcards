//go:build !js || !wasm || casino

package domain

// checkAndTransitionAddon はアドオン判定を行い、人間にアドオン決定を促す場合は true を返す。
func (s *SevenCardStud) checkAndTransitionAddon() bool {
	if s.config.AddonEnabled && s.handCount == s.config.AddonAfterHand {
		needHumanAddon := false
		for i, p := range s.players {
			if !s.addonUsed[i] {
				if p.GetIsHuman() {
					needHumanAddon = true
				} else {
					p.AddChips(s.config.AddonChips)
					s.addonUsed[i] = true
				}
			}
		}
		if needHumanAddon {
			s.phase = SevenCardStudPhaseRebuy
			s.rebuyPhaseType = SevenCardStudRebuyPhaseAddon
			return true
		}
	}
	return false
}

// Rebuy 人間プレイヤーがリバイを実行する
func (s *SevenCardStud) Rebuy() error {
	if s.phase != SevenCardStudPhaseRebuy || s.rebuyPhaseType != SevenCardStudRebuyPhaseRebuy {
		return NewDomainError(ErrWrongPhase, "Rebuy is not available now.")
	}
	for i, p := range s.players {
		if p.GetIsHuman() && p.GetChips() <= 0 && s.rebuyCounts[i] < s.config.RebuyMaxCount {
			p.AddChips(s.config.RebuyChips)
			s.rebuyCounts[i]++
			s.appendLog(i, "rebuy", "rebuy", nil)
			break
		}
	}
	s.rebuyPhaseType = SevenCardStudRebuyPhaseNone
	if s.checkAndTransitionAddon() {
		return nil
	}
	return s.continueReset()
}

// SkipRebuy 人間プレイヤーがリバイを辞退する
func (s *SevenCardStud) SkipRebuy() error {
	if s.phase != SevenCardStudPhaseRebuy || s.rebuyPhaseType != SevenCardStudRebuyPhaseRebuy {
		return NewDomainError(ErrWrongPhase, "Rebuy is not available now.")
	}
	s.rebuyPhaseType = SevenCardStudRebuyPhaseNone
	for _, p := range s.players {
		if p.GetIsHuman() && p.GetChips() <= 0 {
			s.phase = SevenCardStudPhaseEnd
			s.gameEndFlag = true
			return nil
		}
	}
	if s.checkAndTransitionAddon() {
		return nil
	}
	return s.continueReset()
}

// Addon 人間プレイヤーがアドオンを実行する
func (s *SevenCardStud) Addon() error {
	if s.phase != SevenCardStudPhaseRebuy || s.rebuyPhaseType != SevenCardStudRebuyPhaseAddon {
		return NewDomainError(ErrWrongPhase, "Addon is not available now.")
	}
	for i, p := range s.players {
		if p.GetIsHuman() && !s.addonUsed[i] {
			p.AddChips(s.config.AddonChips)
			s.addonUsed[i] = true
			break
		}
	}
	s.rebuyPhaseType = SevenCardStudRebuyPhaseNone
	return s.continueReset()
}

// SkipAddon 人間プレイヤーがアドオンを辞退する
func (s *SevenCardStud) SkipAddon() error {
	if s.phase != SevenCardStudPhaseRebuy || s.rebuyPhaseType != SevenCardStudRebuyPhaseAddon {
		return NewDomainError(ErrWrongPhase, "Addon is not available now.")
	}
	s.rebuyPhaseType = SevenCardStudRebuyPhaseNone
	return s.continueReset()
}

// IsRebuyAvailable 人間プレイヤーがリバイ可能かどうか
func (s *SevenCardStud) IsRebuyAvailable() bool {
	if !s.config.RebuyEnabled || s.handCount > s.config.RebuyPeriodHands {
		return false
	}
	for i, p := range s.players {
		if p.GetIsHuman() && p.GetChips() <= 0 && s.rebuyCounts[i] < s.config.RebuyMaxCount {
			return true
		}
	}
	return false
}

// IsAddonAvailable 人間プレイヤーがアドオン可能かどうか
func (s *SevenCardStud) IsAddonAvailable() bool {
	if !s.config.AddonEnabled || s.handCount != s.config.AddonAfterHand {
		return false
	}
	for i, p := range s.players {
		if p.GetIsHuman() && !s.addonUsed[i] {
			return true
		}
	}
	return false
}

// GetRebuyCounts プレイヤーごとのリバイ回数取得
func (s *SevenCardStud) GetRebuyCounts() []int {
	return copyOf(s.rebuyCounts)
}

// GetAddonUsed プレイヤーごとのアドオン使用フラグ取得
func (s *SevenCardStud) GetAddonUsed() []bool {
	return copyOf(s.addonUsed)
}

// GetRebuyPhaseType リバイフェーズ種別取得
func (s *SevenCardStud) GetRebuyPhaseType() int { return s.rebuyPhaseType }
