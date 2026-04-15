import type { Card, LetItRideResponse, MaskedCard } from '../../../types/card';
import { isMaskedCard } from '../../../types/card';
import { formatCard, formatCardList, formatHeader, formatIndexedCards, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = { 1: 'BET', 2: 'FIRST DECISION', 3: 'SECOND DECISION', 4: 'END' };

/** Format community cards with masked cards shown as '??'. */
function formatCommunityCards(cards: (Card | MaskedCard)[]): string {
  return cards.map((c) => (isMaskedCard(c) ? '??' : formatCard(c))).join(', ');
}

/** Format a Let It Ride game state as terminal text. */
export function formatLetitrideState(state: LetItRideResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Let It Ride'));
  lines.push(`chips: ${state.chips}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push('');

  if (state.playerHand.length > 0) {
    lines.push(`Your hand: ${formatIndexedCards(state.playerHand)}`);
  }

  if (state.communityCards.length > 0) {
    if (state.phase === 4) {
      lines.push(`Community: ${formatCardList(state.communityCards as Card[])}`);
    } else {
      lines.push(`Community: ${formatCommunityCards(state.communityCards)}`);
    }
  }

  lines.push('----------');

  if (state.betAmount > 0) lines.push(`bet per spot: ${state.betAmount}`);
  lines.push(`bet1: ${state.bet1Active ? 'active' : 'pulled'}`);
  lines.push(`bet2: ${state.bet2Active ? 'active' : 'pulled'}`);
  lines.push(`bet3: ${state.bet3Active ? 'active' : 'pulled'}`);

  if (state.phase === 4) {
    lines.push(`payout: bet1=${state.bet1Payout} bet2=${state.bet2Payout} bet3=${state.bet3Payout}`);
    lines.push(`total: ${state.totalPayout}`);
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
