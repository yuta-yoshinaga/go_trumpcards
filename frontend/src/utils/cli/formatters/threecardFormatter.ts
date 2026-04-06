import type { ThreeCardResponse } from '../../../types/card';
import { formatCardList, formatHeader, formatIndexedCards, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = { 1: 'BET', 2: 'ACTION', 3: 'END' };

/** Format a Three Card Poker game state as terminal text. */
export function formatThreecardState(state: ThreeCardResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Three Card Poker'));
  lines.push(`chips: ${state.chips}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push('');

  if (state.playerHand.length > 0) {
    lines.push(`Your hand: ${formatIndexedCards(state.playerHand)}`);
  }

  if (state.phase === 3 && state.dealerHand.length > 0) {
    lines.push(`Dealer: ${formatCardList(state.dealerHand)}`);
    lines.push(`Dealer qualified: ${state.dealerQualified ? 'yes' : 'no'}`);
  }
  lines.push('----------');

  if (state.anteBet > 0) lines.push(`ante: ${state.anteBet}`);
  if (state.pairPlusBet > 0) lines.push(`pair plus: ${state.pairPlusBet}`);
  if (state.playBet > 0) lines.push(`play bet: ${state.playBet}`);

  if (state.phase === 3) {
    lines.push(
      `payout: ante=${state.antePayout} play=${state.playPayout} bonus=${state.anteBonusPayout} pp=${state.pairPlusPayout}`,
    );
    lines.push(`total: ${state.totalPayout}`);
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
