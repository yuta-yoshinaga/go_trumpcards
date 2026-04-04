import type { VideoPokerResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = { 1: 'BET', 2: 'DRAW', 3: 'RESULT' };

/** Format a Video Poker game state as terminal text. */
export function formatVideopokerState(state: VideoPokerResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader(state.variantName || 'Video Poker'));
  lines.push(`chips: ${state.chips}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  if (state.betAmount > 0) lines.push(`bet: ${state.betAmount}`);
  lines.push('');

  if (state.hand.length > 0) {
    const cards = state.hand.map((c, i) => {
      const held = state.heldIndices[i] ? '*' : ' ';
      return `${held}[${i}]${formatCard(c)}`;
    });
    lines.push(cards.join('  '));
  }
  lines.push('----------');

  if (state.phase === 3) {
    if (state.handName) lines.push(`hand: ${state.handName}`);
    lines.push(`payout: ${state.payout}`);
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
