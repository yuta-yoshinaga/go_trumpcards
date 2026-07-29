import type { WindmillResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { WindmillPhase } from '../../types/phases';

/** Returns a Windmill frontend hint or null.
 *
 * The backend already ranks its own suggestions, and it only offers the corner
 * pull-back once ordinary moves are gone. That case is surfaced as its own
 * reason because it dismantles a finished corner -- the player should know the
 * suggestion costs something rather than reading it as a routine move. */
export function getWindmillHint(state: WindmillResponse): HintResult | null {
  if (state.phase !== WindmillPhase.PLAYING) return null;
  if (state.isStalemate) return null;
  if (!state.hint) return null;
  if (state.hint.fromZone === 'corner') {
    return { targetAction: 'play.pullBack', reason: 'hintReason.pullBack', confidence: 'moderate' };
  }
  if (state.hint.fromZone === 'stock') {
    return { targetAction: 'play.draw', reason: 'hintReason.draw', confidence: 'moderate' };
  }
  if (state.hint.toZone === 'center') {
    return { targetAction: 'play.center', reason: 'hintReason.toCenter', confidence: 'strong' };
  }
  return { targetAction: 'play.corner', reason: 'hintReason.toCorner', confidence: 'strong' };
}
