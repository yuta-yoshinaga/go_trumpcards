import type { NinetyNineResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

/** Maps numeric trump suit to a display label. */
const SUIT_LABEL: Readonly<Record<number, string>> = {
  1: 'SPADE',
  2: 'CLOVER',
  3: 'HEART',
  4: 'DIAMOND',
};

/** Format a Ninety-Nine game state as terminal text. */
export function formatNinetynineState(state: NinetyNineResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Ninety-Nine'));
  lines.push(`deal: ${state.dealNumber}  trick: ${state.trickNumber}  hand size: ${state.handSize}`);
  lines.push(`trump: ${SUIT_LABEL[state.trumpSuit] ?? '-'}  target: ${state.targetScore}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(
      `${name}: total=${p.cumulativeScore} round=${p.roundScore} bid=${p.bid} tricks=${p.trickCount} buried=${p.buriedCount}`,
    );
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.currentTrick.length > 0) {
    const parts = state.currentTrick.map((tc) => {
      const name = formatPlayerName(tc.playerIdx, state.players[tc.playerIdx]?.isHuman ?? false);
      return `${name}=${formatCard(tc.card)}`;
    });
    lines.push(`trick: ${parts.join(', ')}`);
  }

  if (state.phase === 0) {
    lines.push('Bury phase: bury 3 cards (suit-sum = bid)');
  }

  if (state.hint && isRequestedHint(state)) {
    if (state.hint.buryIndices !== undefined) {
      lines.push(`HINT: bury [${state.hint.buryIndices.join(', ')}] (${state.hint.reason})`);
    }
    if (state.hint.cardIndex !== undefined) lines.push(`HINT: play [${state.hint.cardIndex}] (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    const winner = formatPlayerName(state.winnerIdx, state.players[state.winnerIdx]?.isHuman ?? false);
    lines.push(`Game Over! Winner: ${winner}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
