import type { NapoleonsSquareResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { NapoleonsSquarePhase } from '../../types/phases';

/** Returns a Napoleon's Square frontend hint or null.
 *
 * The backend ranks its own suggestions — foundation first, then a tableau move,
 * then "turn the stock" — so this only translates the zone into a reason. A
 * stock hint is deliberately weaker: turning a card is always available and
 * says nothing about the position, so it should not read as strong advice. */
export function getNapoleonsSquareHint(state: NapoleonsSquareResponse): HintResult | null {
  if (state.phase !== NapoleonsSquarePhase.PLAYING) return null;
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
