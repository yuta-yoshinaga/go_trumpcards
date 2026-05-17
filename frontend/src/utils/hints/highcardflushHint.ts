import type { HighCardFlushResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { HighCardFlushPhase } from '../../types/phases';

/**
 * Returns a hint for High Card Flush during the Action phase.
 *
 * Strategy (commonly accepted "Basic" High Card Flush strategy):
 *   - 6+ card flush → raise 3x
 *   - 5-card flush  → raise 2x
 *   - 4-card flush  → raise 1x
 *   - 3-card flush with high card Queen or better → raise 1x
 *   - otherwise → fold
 */
export function getHighCardFlushHint(state: HighCardFlushResponse): HintResult | null {
  if (state.phase !== HighCardFlushPhase.ACTION) return null;
  const len = state.playerFlushLen;
  if (len >= 6) {
    return { targetAction: 'raise3x', reason: 'hintReason.raise3x', confidence: 'strong' };
  }
  if (len === 5) {
    return { targetAction: 'raise2x', reason: 'hintReason.raise2x', confidence: 'strong' };
  }
  if (len === 4) {
    return { targetAction: 'raise1x', reason: 'hintReason.raise1xHighFlush', confidence: 'strong' };
  }
  if (len === 3) {
    // Find highest card in any 3-of-a-kind suit. Compute manually here so the
    // hint is independent of the server hint payload.
    const bySuit: Record<string, number[]> = {};
    for (const c of state.playerHand) {
      const v = c.value === 1 ? 14 : c.value;
      const existing = bySuit[c.design] ?? [];
      existing.push(v);
      bySuit[c.design] = existing;
    }
    let best = 0;
    for (const vals of Object.values(bySuit)) {
      if (vals.length >= 3) {
        const top = Math.max(...vals);
        if (top > best) best = top;
      }
    }
    if (best >= 12) {
      return { targetAction: 'raise1x', reason: 'hintReason.raise1xMarginal', confidence: 'moderate' };
    }
  }
  return { targetAction: 'fold', reason: 'hintReason.fold', confidence: 'moderate' };
}
