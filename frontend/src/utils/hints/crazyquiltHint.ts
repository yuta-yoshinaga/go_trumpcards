import type { CrazyQuiltResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { CrazyQuiltPhase } from '../../types/phases';

/** Returns a CrazyQuilt frontend hint or null.
 *
 * Filling a gap straight from the stock is called out separately: it is the one
 * move that spends a stock card without turning it, and with a single pass
 * through the deck that choice is worth flagging. */
export function getCrazyQuiltHint(state: CrazyQuiltResponse): HintResult | null {
  if (state.phase !== CrazyQuiltPhase.PLAYING) return null;
  if (state.isStalemate) return null;
  if (!state.hint) return null;
  if (state.hint.toZone === 'foundation') {
    return { targetAction: 'play.foundation', reason: 'hintReason.toFoundation', confidence: 'strong' };
  }
  if (state.hint.fromZone === 'stock') {
    if (state.hint.toZone === 'tableau') {
      return { targetAction: 'play.fillGap', reason: 'hintReason.fillGap', confidence: 'strong' };
    }
    return { targetAction: 'play.draw', reason: 'hintReason.draw', confidence: 'moderate' };
  }
  return { targetAction: 'play.tableau', reason: 'hintReason.toTableau', confidence: 'strong' };
}
