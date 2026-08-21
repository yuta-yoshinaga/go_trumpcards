import type { PerseveranceResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { PerseverancePhase } from '../../types/phases';

/** Returns a Perseverance frontend hint or null.
 *
 * Perseverance is a demanding solitaire — Kings sit at the bottom of their
 * column and the tableau builds down IN SUIT, so destinations are scarce. Two
 * redeals soften that, but they are finite. The strongest tactical advice is
 * "always send a card to the foundation if it is legal" because freeing tableau
 * column tops unlocks future moves. The backend already supplies the next-move
 * hint via `state.hint`, so we surface it here. */
export function getPerseveranceHint(state: PerseveranceResponse): HintResult | null {
  if (state.phase !== PerseverancePhase.PLAYING) return null;
  if (state.isStalemate) return null;
  if (!state.hint) return null;
  const target = state.hint.toZone === 'foundation' ? 'play.foundation' : 'play.tableau';
  const reason = state.hint.toZone === 'foundation' ? 'hintReason.toFoundation' : 'hintReason.toTableau';
  return { targetAction: target, reason, confidence: 'strong' };
}
