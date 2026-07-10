import type { TrenteEtQuaranteResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { TrenteEtQuaranteBetType, TrenteEtQuarantePhase } from '../../types/phases';

/**
 * Returns a Trente et Quarante hint during the bet phase.
 *
 * All four bets are even money with the same tiny house edge (the Refait on 31
 * is the only edge, ~1.1% with insurance). Noir is the traditional default and
 * is as good as any other bet, so recommend Noir when the player has not yet
 * placed a wager.
 */
export function getTrenteEtQuaranteHint(state: TrenteEtQuaranteResponse): HintResult | null {
  if (state.phase !== TrenteEtQuarantePhase.BET) return null;
  if (state.currentBet === TrenteEtQuaranteBetType.NOIR) return null;
  return { targetAction: 'bet', reason: 'hint.evenMoney', confidence: 'moderate' };
}
