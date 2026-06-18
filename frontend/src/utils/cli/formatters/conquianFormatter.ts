import type { ConquianResponse } from '../../../types/card';
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
  1: 'MELD',
  2: 'ROUND END',
  3: 'GAME END',
};

/** Format a Conquian game state as terminal text. */
export function formatConquianState(state: ConquianResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Conquian'));
  lines.push(`round: ${state.roundNumber}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(`discard: ${state.discardTop ? formatCard(state.discardTop) : '[  ]'} | stock: ${state.drawPileCount}`);
  if (state.tookDiscard) lines.push('(took discard — must be melded)');
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(`${name}: wins=${p.wins} cards=${p.cardCount}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
    if (p.melds.length > 0) {
      lines.push('  melds:');
      for (const m of p.melds) {
        lines.push(`    ${formatCardList(m.cards)}`);
      }
    }
  }
  lines.push('----------');

  if (!state.gameEndFlag) {
    const current = formatPlayerName(state.currentPlayerIdx, state.players[state.currentPlayerIdx]?.isHuman ?? false);
    lines.push(`turn: ${current}`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    if (state.winnerIdx < 0) {
      lines.push('Game Over! Draw');
    } else {
      const winner = formatPlayerName(state.winnerIdx, state.players[state.winnerIdx]?.isHuman ?? false);
      lines.push(`Game Over! Winner: ${winner}`);
    }
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
