import type { ContractRummyResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { aceHighAdjacent, heaviestSpare, isMaterial } from './rummyHintShape';

/** フェーズ番号 (sync: internal/domain/ContractRummy.go)。 */
const PHASE_DRAW = 0;
const PHASE_PLAY = 1;

/**
 * Returns a frontend {@link HintResult} for Contract Rummy, or null when no
 * suggestion is available.
 *
 * A player's `melds` is what they have laid down, not what their hand could
 * make, so this uses the same shallow connects-with-something test as the other
 * rummies rather than proving a meld.
 *
 * What is specific here is the contract. Until it is met the whole contract has
 * to go down in one turn, so the heavy cards are material and throwing them is
 * wrong — the same inversion Kalooki has, but driven by a per-round slot list
 * rather than a point threshold, which is why the hint names the slots left.
 */
export function getContractRummyHint(state: ContractRummyResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0 || state.currentPlayerIdx !== human.id) return null;

  if (state.phase === PHASE_DRAW) {
    const top = state.discardTop;
    return top && isMaterial(top, human.cards, aceHighAdjacent)
      ? { targetAction: 'takeDiscard', reason: 'frontendHint.contractrummyTakeDiscard', confidence: 'moderate' }
      : { targetAction: 'drawStock', reason: 'frontendHint.contractrummyDrawStock', confidence: 'moderate' };
  }

  if (state.phase !== PHASE_PLAY) return null;

  // **契約を満たすまでは崩さない。**一度に全部そろえる必要があるので、
  // 重い札こそ材料になる。
  if (!human.contractMet && state.contractSlots.length > 0) {
    return { targetAction: 'meld', reason: 'frontendHint.contractrummyMeetContract', confidence: 'moderate' };
  }

  const idx = heaviestSpare(human.cards, aceHighAdjacent);
  return { targetAction: `card-${idx}`, reason: 'frontendHint.contractrummyDiscardHeavy', confidence: 'moderate' };
}
