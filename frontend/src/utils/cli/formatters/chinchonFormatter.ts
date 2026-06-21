import type { ChinchonResponse } from '../../../types/card';
import {
  formatCard,
  formatCardList,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
} from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  0: 'DRAW',
  1: 'DISCARD',
  2: 'LAYOFF',
  3: 'ROUND END',
  4: 'GAME END',
};

/** Format a Chinchón game state as terminal text. */
export function formatChinchonState(state: ChinchonResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Chinchon'));
  lines.push(`round: ${state.roundNumber}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(`discard: ${state.discardTop ? formatCard(state.discardTop) : '[  ]'} | stock: ${state.drawPileCount}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const out = p.eliminated ? ' [OUT]' : '';
    lines.push(`${name}: total=${p.cumulativeScore} round=${p.roundScore} cards=${p.cardCount}${out}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.knockerIdx >= 0) {
    const knocker = formatPlayerName(state.knockerIdx, state.players[state.knockerIdx]?.isHuman ?? false);
    lines.push(`${knocker} knocked!`);
    if (state.knockerMelds.length > 0) {
      lines.push('melds:');
      for (const m of state.knockerMelds) {
        lines.push(`  ${formatCardList(m.cards)}`);
      }
    }
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    const winner = formatPlayerName(state.winnerIdx, state.players[state.winnerIdx]?.isHuman ?? false);
    lines.push(`Game Over! Winner: ${winner}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
