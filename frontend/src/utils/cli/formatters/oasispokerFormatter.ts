import type { Card, MaskedCard, OasisPokerResponse } from '../../../types/card';
import { isMaskedCard } from '../../../types/card';
import { formatCard, formatCardList, formatHeader, formatIndexedCards, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = { 1: 'BET', 2: 'EXCHANGE', 3: 'ACTION', 4: 'END' };

/** Format the dealer's partial hand during the pre-end phases (1 face-up + hidden). */
function formatDealerPartialHand(dealerHand: (Card | MaskedCard)[]): string {
  return dealerHand.map((c) => (isMaskedCard(c) ? '??' : formatCard(c))).join(', ');
}

/** Format an Oasis Poker game state as terminal text. */
export function formatOasispokerState(state: OasisPokerResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Oasis Poker'));
  lines.push(`chips: ${state.chips}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push('');

  if (state.playerHand.length > 0) {
    lines.push(`Your hand: ${formatIndexedCards(state.playerHand)}`);
  }

  if (state.phase === 4 && state.dealerHand.length > 0) {
    lines.push(`Dealer: ${formatCardList(state.dealerHand as Card[])}`);
    lines.push(`Dealer qualified: ${state.dealerQualified ? 'yes' : 'no'}`);
  } else if (state.dealerHand.length > 0) {
    lines.push(`Dealer: ${formatDealerPartialHand(state.dealerHand)}`);
  }
  lines.push('----------');

  if (state.anteBet > 0) lines.push(`ante: ${state.anteBet}`);
  if (state.jackpotBet > 0) lines.push(`jackpot: ${state.jackpotBet}`);
  if (state.exchangeCount > 0) lines.push(`exchanged: ${state.exchangeCount} (fee: ${state.exchangeFee})`);
  if (state.playBet > 0) lines.push(`play bet: ${state.playBet}`);

  if (state.phase === 4) {
    lines.push(`payout: ante=${state.antePayout} play=${state.playPayout} jackpot=${state.jackpotPayout}`);
    lines.push(`total: ${state.totalPayout}`);
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
