import type { BakersDozenResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { BakersDozenPhase } from '../../types/phases';

/** Returns a Baker's Dozen frontend hint or null.
 *
 * Baker's Dozen is a difficult solitaire — Kings cannot be moved out of the way
 * and there is no second deal. The strongest tactical advice is "always send a
 * card to the foundation if it is legal" because freeing tableau column tops
 * unlocks future moves. The backend already supplies the next-move hint via
 * `state.hint`, so we surface it here. */
export function getBakersdozenHint(state: BakersDozenResponse): HintResult | null {
  if (state.phase !== BakersDozenPhase.PLAYING) return null;
  if (state.isStalemate) return null;
  if (!state.hint) return null;
  const target = state.hint.toZone === 'foundation' ? 'play.foundation' : 'play.tableau';
  const reason = state.hint.toZone === 'foundation' ? 'hintReason.toFoundation' : 'hintReason.toTableau';
  return { targetAction: target, reason, confidence: 'strong' };
}
