import type { GrandfathersClockResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { GrandfathersClockPhase } from '../../types/phases';

/** Returns a Grandfather's Clock frontend hint or null.
 *
 * Every card is face-up and there is no stock, so the whole game is ordering
 * the moves you can already see. The backend ranks its own suggestions — a
 * clock face first, then a tableau shuffle — and this only translates the zone
 * into a reason. */
export function getGrandfathersClockHint(state: GrandfathersClockResponse): HintResult | null {
  if (state.phase !== GrandfathersClockPhase.PLAYING) return null;
  if (state.isStalemate) return null;
  if (!state.hint) return null;
  if (state.hint.toZone === 'foundation') {
    return { targetAction: 'play.foundation', reason: 'hintReason.toFoundation', confidence: 'strong' };
  }
  return { targetAction: 'play.tableau', reason: 'hintReason.toTableau', confidence: 'strong' };
}
