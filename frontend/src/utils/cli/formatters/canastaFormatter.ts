import type { CanastaResponse } from '../../../types/card';
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
  2: 'DISCARD',
  3: 'ROUND END',
  4: 'GAME END',
};

/** Format a Canasta game state as terminal text. */
export function formatCanastaState(state: CanastaResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Canasta'));
  lines.push(`round: ${state.roundNumber}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(
    `stock: ${state.drawPileCount} | discard: ${state.discardTop ? formatCard(state.discardTop) : '[  ]'} (${state.discardPileCount})${state.isFrozen ? ' [FROZEN]' : ''}`,
  );
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const tags: string[] = [];
    if (p.hasCanasta) tags.push('Canasta');
    if (p.hasInitMeld) tags.push('Init Meld');
    const tagStr = tags.length > 0 ? ` [${tags.join(', ')}]` : '';
    lines.push(
      `${name}: total=${p.cumulativeScore} round=${p.roundScore} cards=${p.cardCount} red3=${p.red3Count}${tagStr}`,
    );
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
    if (p.melds.length > 0) {
      for (const m of p.melds) {
        const type = m.isCanasta ? '(canasta)' : m.isNatural ? '(natural)' : '';
        lines.push(`  meld: ${formatCardList(m.cards)} ${type}`);
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
    const winner = formatPlayerName(state.winnerIdx, state.players[state.winnerIdx]?.isHuman ?? false);
    lines.push(`Game Over! Winner: ${winner}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
