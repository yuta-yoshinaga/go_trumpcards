import type { TongitsResponse } from '../../../types/card';
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
  2: 'ROUND END',
  3: 'GAME END',
};

/** Format a Tongits game state as terminal text. */
export function formatTongitsState(state: TongitsResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Tongits'));
  lines.push(`round: ${state.roundNumber}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(`discard: ${state.discardTop ? formatCard(state.discardTop) : '[  ]'} | stock: ${state.drawPileCount}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(`${name}: total=${p.cumulativeScore} round=${p.roundScore} cards=${p.cardCount}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.isTongits) lines.push('TONGITS declared on deal!');
  for (const p of state.players) {
    if (p.melds.length > 0) {
      lines.push(`${formatPlayerName(p.id, p.isHuman)} melds:`);
      for (const m of p.melds) lines.push(`  ${formatCardList(m.cards)}`);
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
