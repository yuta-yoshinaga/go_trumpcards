import type { HighCardFlushResponse } from '../../../types/card';
import { HighCardFlushPhase } from '../../../types/phases';
import { formatCardList, formatHeader, formatIndexedCards, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [HighCardFlushPhase.BET]: 'BET',
  [HighCardFlushPhase.ACTION]: 'ACTION',
  [HighCardFlushPhase.END]: 'END',
};

/** Format a High Card Flush game state as terminal text. */
export function formatHighcardflushState(state: HighCardFlushResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('High Card Flush'));
  lines.push(`chips: ${state.chips}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push('');

  if (state.playerHand.length > 0) {
    lines.push(`Your hand: ${formatIndexedCards(state.playerHand)}`);
    lines.push(`Your longest flush: ${state.playerFlushLen}`);
  }

  if (state.dealerHand.length > 0) {
    if (state.phase === HighCardFlushPhase.END) {
      lines.push(`Dealer: ${formatCardList(state.dealerHand)}`);
      lines.push(`Dealer longest flush: ${state.dealerFlushLen}${state.dealerQualified ? '' : ' (not qualified)'}`);
    } else {
      lines.push('Dealer: (hidden)');
    }
  }

  lines.push('----------');
  if (state.anteBet > 0) lines.push(`ante: ${state.anteBet}`);
  if (state.raiseBet > 0) lines.push(`raise: ${state.raiseBet}`);
  if (state.flushBonusBet > 0) lines.push(`flush bonus bet: ${state.flushBonusBet}`);
  if (state.straightFlushBet > 0) lines.push(`straight flush bet: ${state.straightFlushBet}`);

  if (state.phase === HighCardFlushPhase.END) {
    lines.push(
      `payout: ante=${state.antePayout} raise=${state.raisePayout} ` +
        `fb=${state.flushBonusPayout} sf=${state.straightFlushPayout}`,
    );
    lines.push(`total: ${state.totalPayout}`);
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
