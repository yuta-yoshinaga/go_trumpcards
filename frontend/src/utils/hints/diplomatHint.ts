import type { DiplomatResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { DiplomatPhase } from '../../types/phases';

/**
 * Returns a Diplomat frontend hint or null.
 *
 * The stock is only ever a draw here: an empty column is filled from another
 * column or the waste, so there is no fill-a-gap-from-the-deck move to flag.
 */
export function getDiplomatHint(state: DiplomatResponse): HintResult | null {
  if (state.phase !== DiplomatPhase.PLAYING) return null;
  if (state.isStalemate) return null;
  if (!state.hint) return null;
  if (state.hint.toZone === 'foundation') {
    return { targetAction: 'play.foundation', reason: 'hintReason.toFoundation', confidence: 'strong' };
  }
  if (state.hint.fromZone === 'stock') {
    return { targetAction: 'play.draw', reason: 'hintReason.draw', confidence: 'moderate' };
  }
  return { targetAction: 'play.tableau', reason: 'hintReason.toTableau', confidence: 'strong' };
}
