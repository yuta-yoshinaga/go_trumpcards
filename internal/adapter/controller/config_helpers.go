//go:build !js || !wasm || casino

package controller

import (
	"errors"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// applyIntIfGte sets *dst to *src if src is non-nil and *src >= minVal.
func applyIntIfGte(dst *int, src *int, minVal int) {
	if src != nil && *src >= minVal {
		*dst = *src
	}
}

// applyBool sets *dst to *src if src is non-nil.
func applyBool(dst *bool, src *bool) {
	if src != nil {
		*dst = *src
	}
}

// applyBettingLimit validates and applies the betting limit,
// clamping the value to the valid range [BettingLimitFixed, BettingLimitNoLimit].
func applyBettingLimit(dst *domain.BettingLimitType, src *int) {
	if src != nil {
		bl := *src
		if bl < int(domain.BettingLimitFixed) {
			bl = int(domain.BettingLimitFixed)
		} else if bl > int(domain.BettingLimitNoLimit) {
			bl = int(domain.BettingLimitNoLimit)
		}
		*dst = domain.BettingLimitType(bl)
	}
}

// applyRebuyConfig applies rebuy-related config fields.
func applyRebuyConfig(rebuyEnabled *bool, rebuyMaxCount, rebuyChips, rebuyPeriodHands *int,
	srcEnabled *bool, srcMaxCount, srcChips, srcPeriodHands *int) {
	applyBool(rebuyEnabled, srcEnabled)
	applyIntIfGte(rebuyMaxCount, srcMaxCount, 1)
	applyIntIfGte(rebuyChips, srcChips, 1)
	applyIntIfGte(rebuyPeriodHands, srcPeriodHands, 1)
}

// applyAddonConfig applies addon-related config fields.
func applyAddonConfig(addonEnabled *bool, addonChips, addonAfterHand *int,
	srcEnabled *bool, srcChips, srcAfterHand *int) {
	applyBool(addonEnabled, srcEnabled)
	applyIntIfGte(addonChips, srcChips, 1)
	applyIntIfGte(addonAfterHand, srcAfterHand, 1)
}

// applyTableSize validates src against isValid and writes it to *dst.
// When src is nil, *dst is left unchanged. When src is non-nil but invalid,
// errMsg is returned wrapped as a controller-style param error.
func applyTableSize(dst *int, src *int, isValid func(int) bool, errMsg string) error {
	if src == nil {
		return nil
	}
	ts := *src
	if !isValid(ts) {
		return errors.New(errMsg)
	}
	*dst = ts
	return nil
}

// validateAndApplyBlinds validates small/big blind values and applies them to dst pointers.
// If only one blind is provided, the other is auto-adjusted.
func validateAndApplyBlinds(sb, bb *int, sbPtr, bbPtr *int, defaultBB int) error {
	sbProvided := sbPtr != nil && *sbPtr >= 1
	bbProvided := bbPtr != nil && *bbPtr >= 1
	if sbProvided {
		*sb = *sbPtr
	}
	if bbProvided {
		*bb = *bbPtr
	}
	if sbProvided && !bbProvided && *sb >= defaultBB {
		*bb = *sb * 2
	} else if bbProvided && !sbProvided && *bb > 1 {
		*sb = *bb / 2
	}
	if *sb >= *bb {
		return errors.New("param error: smallBlind must be less than bigBlind")
	}
	return nil
}
