import type { Card, CaribbeanStudResponse } from '../../../types/card';
import { formatCard, formatCardList, formatHeader, formatIndexedCards, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = { 1: 'BET', 2: 'ACTION', 3: 'END' };

/** Returns true if the card is masked (face-down, not yet revealed). */
function isMaskedCard(card: Card): boolean {
  return (card as { design: string }).design === '';
}

/** Format the dealer's partial hand during the action phase (1 face-up + hidden). */
function formatDealerActionHand(dealerHand: Card[]): string {
  return dealerHand.map((c) => (isMaskedCard(c) ? '??' : formatCard(c))).join(', ');
}

/** Format a Caribbean Stud Poker game state as terminal text. */
export function formatCaribbeanstudState(state: CaribbeanStudResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Caribbean Stud Poker'));
  lines.push(`chips: ${state.chips}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push('');

  if (state.playerHand.length > 0) {
    lines.push(`Your hand: ${formatIndexedCards(state.playerHand)}`);
  }

  if (state.phase === 2 && state.dealerHand.length > 0) {
    lines.push(`Dealer: ${formatDealerActionHand(state.dealerHand)}`);
  }

  if (state.phase === 3 && state.dealerHand.length > 0) {
    lines.push(`Dealer: ${formatCardList(state.dealerHand)}`);
    lines.push(`Dealer qualified: ${state.dealerQualified ? 'yes' : 'no'}`);
  }
  lines.push('----------');

  if (state.anteBet > 0) lines.push(`ante: ${state.anteBet}`);
  if (state.jackpotBet > 0) lines.push(`jackpot: ${state.jackpotBet}`);
  if (state.playBet > 0) lines.push(`play bet: ${state.playBet}`);

  if (state.phase === 3) {
    lines.push(`payout: ante=${state.antePayout} play=${state.playPayout} jackpot=${state.jackpotPayout}`);
    lines.push(`total: ${state.totalPayout}`);
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
