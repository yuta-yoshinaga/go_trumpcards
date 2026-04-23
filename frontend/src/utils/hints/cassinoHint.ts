import type { CassinoResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Cassino hint heuristic. Recommends the action with the highest expected
 * point gain for the current human hand.
 *
 * Priority order:
 *   1. Take that captures >= 2pt swing (spade / ace / 10♦ / 2♠ / sweep).
 *   2. Take that captures at least 2 cards (most-cards race).
 *   3. Trail the lowest-value non-point card.
 * Never recommends build (builds are positional and depend on future turns).
 */
export function getCassinoHint(state: CassinoResponse): HintResult | null {
  if (!state || state.gameEndFlag) return null;
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0) return null;

  // Detect presence of point cards on the table (spade, ace, big casino 10♦)
  const hasPointOnTable = state.tableCards.some(
    (c) => c.design === 'SPADE' || c.value === 1 || (c.design === 'DIAMOND' && c.value === 10),
  );

  // If any legal take is likely to capture points → strong recommendation
  for (const c of human.cards) {
    if (c.value >= 11 && c.value <= 13) continue; // face cards: rank-match only, low priority
    const sum = state.tableCards.filter((t) => t.value === c.value).reduce((acc, t) => acc + t.value, 0);
    if (sum > 0 && hasPointOnTable) {
      return { targetAction: 'take', reason: 'hint.take.points', confidence: 'strong' };
    }
  }

  // Otherwise if the table has ≥ 2 cards that can be captured, recommend take
  if (state.tableCards.length >= 2) {
    return { targetAction: 'take', reason: 'hint.take.cards', confidence: 'moderate' };
  }

  return { targetAction: 'trail', reason: 'hint.trail.safe', confidence: 'moderate' };
}
