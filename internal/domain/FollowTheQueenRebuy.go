//go:build !js || !wasm || casino

package domain

// checkAndTransitionAddon はアドオン判定を行い、人間にアドオン決定を促す場合は true を返す。
func (s *FollowTheQueen) checkAndTransitionAddon() bool {
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
			s.phase = FollowTheQueenPhaseRebuy
			s.rebuyPhaseType = FollowTheQueenRebuyPhaseAddon
			return true
		}
	}
	return false
}

// Rebuy 人間プレイヤーがリバイを実行する
func (s *FollowTheQueen) Rebuy() error {
	if s.phase != FollowTheQueenPhaseRebuy || s.rebuyPhaseType != FollowTheQueenRebuyPhaseRebuy {
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
	s.rebuyPhaseType = FollowTheQueenRebuyPhaseNone
	if s.checkAndTransitionAddon() {
		return nil
	}
	return s.continueReset()
}

// SkipRebuy 人間プレイヤーがリバイを辞退する
func (s *FollowTheQueen) SkipRebuy() error {
	if s.phase != FollowTheQueenPhaseRebuy || s.rebuyPhaseType != FollowTheQueenRebuyPhaseRebuy {
		return NewDomainError(ErrWrongPhase, "Rebuy is not available now.")
	}
	s.rebuyPhaseType = FollowTheQueenRebuyPhaseNone
	for _, p := range s.players {
		if p.GetIsHuman() && p.GetChips() <= 0 {
			s.phase = FollowTheQueenPhaseEnd
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
func (s *FollowTheQueen) Addon() error {
	if s.phase != FollowTheQueenPhaseRebuy || s.rebuyPhaseType != FollowTheQueenRebuyPhaseAddon {
		return NewDomainError(ErrWrongPhase, "Addon is not available now.")
	}
	for i, p := range s.players {
		if p.GetIsHuman() && !s.addonUsed[i] {
			p.AddChips(s.config.AddonChips)
			s.addonUsed[i] = true
			break
		}
	}
	s.rebuyPhaseType = FollowTheQueenRebuyPhaseNone
	return s.continueReset()
}

// SkipAddon 人間プレイヤーがアドオンを辞退する
func (s *FollowTheQueen) SkipAddon() error {
	if s.phase != FollowTheQueenPhaseRebuy || s.rebuyPhaseType != FollowTheQueenRebuyPhaseAddon {
		return NewDomainError(ErrWrongPhase, "Addon is not available now.")
	}
	s.rebuyPhaseType = FollowTheQueenRebuyPhaseNone
	return s.continueReset()
}

// IsRebuyAvailable 人間プレイヤーがリバイ可能かどうか
func (s *FollowTheQueen) IsRebuyAvailable() bool {
	return rebuyAvailable(s.config.RebuyEnabled, s.handCount, s.config.RebuyPeriodHands, s.players, s.rebuyCounts, s.config.RebuyMaxCount)
}

// IsAddonAvailable 人間プレイヤーがアドオン可能かどうか
func (s *FollowTheQueen) IsAddonAvailable() bool {
	return addonAvailable(s.config.AddonEnabled, s.handCount, s.config.AddonAfterHand, s.players, s.addonUsed)
}

// GetRebuyCounts プレイヤーごとのリバイ回数取得
func (s *FollowTheQueen) GetRebuyCounts() []int {
	return copyOf(s.rebuyCounts)
}

// GetAddonUsed プレイヤーごとのアドオン使用フラグ取得
func (s *FollowTheQueen) GetAddonUsed() []bool {
	return copyOf(s.addonUsed)
}

// GetRebuyPhaseType リバイフェーズ種別取得
func (s *FollowTheQueen) GetRebuyPhaseType() int { return s.rebuyPhaseType }
