import type { SevenBridgeResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { SevenBridgePhase } from '../../types/phases';

/** Returns a Seven Bridge frontend hint for the human player. */
export function getSevenbridgeHint(state: SevenBridgeResponse): HintResult | null {
  if (!state) return null;
  if (state.gameEndFlag) return null;
  if (state.phase !== SevenBridgePhase.DRAW && state.phase !== SevenBridgePhase.PLAY) return null;

  const human = state.players?.find((p) => p.isHuman);
  if (!human) return null;

  if (state.phase === SevenBridgePhase.DRAW && state.discardTop) {
    const top = state.discardTop;
    // Pon opportunity: 2+ of same rank in hand
    const sameRank = human.cards.filter((c) => c.value === top.value).length;
    if (sameRank >= 2) {
      return { targetAction: 'pon', reason: 'frontendHint.sevenbridgePon', confidence: 'strong' };
    }
    // Chi opportunity: 2 cards in same suit forming a run with top
    const sameSuit = human.cards
      .filter((c) => c.design === top.design)
      .map((c) => c.value)
      .sort((a, b) => a - b);
    const hasChi =
      (sameSuit.includes(top.value - 2) && sameSuit.includes(top.value - 1)) ||
      (sameSuit.includes(top.value - 1) && sameSuit.includes(top.value + 1)) ||
      (sameSuit.includes(top.value + 1) && sameSuit.includes(top.value + 2));
    if (hasChi) {
      return { targetAction: 'chi', reason: 'frontendHint.sevenbridgeChi', confidence: 'moderate' };
    }
  }

  if (state.phase === SevenBridgePhase.PLAY) {
    // Encourage melding when 3+ of same rank are in hand
    const ranks = new Map<number, number>();
    for (const c of human.cards) {
      ranks.set(c.value, (ranks.get(c.value) ?? 0) + 1);
    }
    for (const count of ranks.values()) {
      if (count >= 3) {
        return { targetAction: 'meld', reason: 'frontendHint.sevenbridgeMeld', confidence: 'strong' };
      }
    }

    const suitedValues = new Map<string, Set<number>>();
    for (const card of human.cards) {
      const values = suitedValues.get(card.design) ?? new Set<number>();
      values.add(card.value);
      suitedValues.set(card.design, values);
    }
    for (const values of suitedValues.values()) {
      const sorted = [...values].sort((a, b) => a - b);
      let runLength = 1;
      for (let i = 1; i < sorted.length; i++) {
        runLength = sorted[i] === sorted[i - 1] + 1 ? runLength + 1 : 1;
        if (runLength >= 3) {
          return { targetAction: 'meld', reason: 'frontendHint.sevenbridgeMeld', confidence: 'strong' };
        }
      }
    }
  }

  return null;
}
