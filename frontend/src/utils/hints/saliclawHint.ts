import type { SalicLawResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { SalicLawPhase } from '../../types/phases';

/**
 * Returns a Salic Law frontend hint or null.
 *
 * There is no waste and the stock is not a move source: a `stock` hint always
 * means "deal another card". Mapping it through the move wording would name a
 * destination column that does not exist.
 */
export function getSalicLawHint(state: SalicLawResponse): HintResult | null {
  if (state.phase !== SalicLawPhase.PLAYING) return null;
  if (state.isStalemate) return null;
  if (!state.hint) return null;
  if (state.hint.fromZone === 'stock') {
    return { targetAction: 'play.draw', reason: 'hintReason.draw', confidence: 'moderate' };
  }
  if (state.hint.toZone === 'foundation') {
    return { targetAction: 'play.foundation', reason: 'hintReason.toFoundation', confidence: 'strong' };
  }
  return { targetAction: 'play.tableau', reason: 'hintReason.toTableau', confidence: 'strong' };
}
