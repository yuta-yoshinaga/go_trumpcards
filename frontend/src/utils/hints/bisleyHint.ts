import type { BisleyResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { BisleyPhase } from '../../types/phases';

/** Returns a Bisley frontend hint or null.
 *
 * Every card is face-up from the deal, so there is no information to uncover —
 * the whole game is deciding which of the two foundation directions to feed.
 * The backend already prefers a foundation move over a tableau shuffle, so we
 * surface `state.hint` unchanged and only translate the zone into a reason. */
export function getBisleyHint(state: BisleyResponse): HintResult | null {
  if (state.phase !== BisleyPhase.PLAYING) return null;
  if (state.isStalemate) return null;
  if (!state.hint) return null;
  switch (state.hint.toZone) {
    case 'ace':
      return { targetAction: 'play.aceFoundation', reason: 'hintReason.toAce', confidence: 'strong' };
    case 'king':
      return { targetAction: 'play.kingFoundation', reason: 'hintReason.toKing', confidence: 'strong' };
    default:
      return { targetAction: 'play.tableau', reason: 'hintReason.toTableau', confidence: 'strong' };
  }
}
