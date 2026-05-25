import type { Card, MaskedCard, RussianPokerResponse } from '../../../types/card';
import { isMaskedCard } from '../../../types/card';
import { formatCard, formatCardList, formatHeader, formatIndexedCards, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  1: 'BET',
  2: 'ACTION',
  3: 'SELECT',
  4: 'POST-ACTION',
  5: 'FORCE QUALIFY',
  6: 'END',
};

/** Format the dealer's partial hand during the pre-end phases (1 face-up + hidden). */
function formatDealerPartialHand(dealerHand: (Card | MaskedCard)[]): string {
  return dealerHand.map((c) => (isMaskedCard(c) ? '??' : formatCard(c))).join(', ');
}

/** Format a Russian Poker game state as terminal text. */
export function formatRussianpokerState(state: RussianPokerResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Russian Poker'));
  lines.push(`chips: ${state.chips}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push('');

  if (state.playerHand.length > 0) {
    lines.push(`Your hand: ${formatIndexedCards(state.playerHand)}`);
  }

  if (state.phase === 6 && state.dealerHand.length > 0) {
    lines.push(`Dealer: ${formatCardList(state.dealerHand as Card[])}`);
    lines.push(`Dealer qualified: ${state.dealerQualified ? 'yes' : 'no'}`);
  } else if (state.dealerHand.length > 0) {
    lines.push(`Dealer: ${formatDealerPartialHand(state.dealerHand)}`);
  }
  lines.push('----------');

  if (state.anteBet > 0) lines.push(`ante: ${state.anteBet}`);
  if (state.exchangeCount > 0) lines.push(`exchanged: ${state.exchangeCount} (fee: ${state.exchangeFee})`);
  if (state.bought6th) lines.push(`bought 6th card (fee: ${state.buy6thFee})`);
  if (state.playBet > 0) lines.push(`play bet: ${state.playBet}`);
  if (state.forceExchanged) lines.push(`force exchange (fee: ${state.forceExchangeFee})`);

  if (state.phase === 6) {
    lines.push(`payout: ante=${state.antePayout} play=${state.playPayout}`);
    lines.push(`total: ${state.totalPayout}`);
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
