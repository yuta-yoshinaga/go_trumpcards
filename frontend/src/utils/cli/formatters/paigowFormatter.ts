import type { PaiGowResponse } from '../../../types/card';
import { formatCardList, formatHeader, formatIndexedCards, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = { 1: 'BET', 2: 'SET_HANDS', 3: 'END' };

/** Format a Pai Gow Poker game state as terminal text. */
export function formatPaigowState(state: PaiGowResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Pai Gow Poker'));
  lines.push(`chips: ${state.chips}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push('');

  if (state.playerCards.length > 0 && state.phase === 2) {
    lines.push(`Your cards: ${formatIndexedCards(state.playerCards)}`);
  }

  if (state.phase === 3) {
    if (state.playerHighHand.length > 0) lines.push(`Player high: ${formatCardList(state.playerHighHand)}`);
    if (state.playerLowHand.length > 0) lines.push(`Player low: ${formatCardList(state.playerLowHand)}`);
    if (state.dealerHighHand.length > 0) lines.push(`Dealer high: ${formatCardList(state.dealerHighHand)}`);
    if (state.dealerLowHand.length > 0) lines.push(`Dealer low: ${formatCardList(state.dealerLowHand)}`);
  }
  lines.push('----------');

  if (state.bet > 0) lines.push(`bet: ${state.bet}`);

  if (state.phase === 3) {
    lines.push(`payout: ${state.payout}  commission: ${state.commission}`);
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
