import type { MissMilliganResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { MissMilliganPhase } from '../../types/phases';

/** Returns a Miss Milligan frontend hint or null.
 *
 * The backend ranks its own suggestions, and while cards are held aside it
 * deliberately offers nothing but putting them back — dealing and a second
 * waive are both barred until then, so no other move can help. That case is
 * surfaced as its own reason rather than as a generic tableau move. */
export function getMissMilliganHint(state: MissMilliganResponse): HintResult | null {
  if (state.phase !== MissMilliganPhase.PLAYING) return null;
  if (state.isStalemate) return null;
  if (!state.hint) return null;
  if (state.hint.fromZone === 'waived') {
    return { targetAction: 'play.placeWaived', reason: 'hintReason.returnWaived', confidence: 'strong' };
  }
  if (state.hint.toZone === 'foundation') {
    return { targetAction: 'play.foundation', reason: 'hintReason.toFoundation', confidence: 'strong' };
  }
  if (state.hint.fromZone === 'stock') {
    return { targetAction: 'play.deal', reason: 'hintReason.deal', confidence: 'moderate' };
  }
  return { targetAction: 'play.tableau', reason: 'hintReason.toTableau', confidence: 'strong' };
}
