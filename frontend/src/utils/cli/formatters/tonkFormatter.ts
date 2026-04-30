import type { TonkResponse } from '../../../types/card';
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

/** Format a Tonk game state as terminal text. */
export function formatTonkState(state: TonkResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Tonk'));
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

  if (state.knockerIdx >= 0) {
    const knocker = formatPlayerName(state.knockerIdx, state.players[state.knockerIdx]?.isHuman ?? false);
    if (state.isTonk) {
      lines.push(`${knocker} declared TONK on deal!`);
    } else {
      lines.push(`${knocker} knocked!${state.isUndercut ? ' (UNDERCUT!)' : ''}`);
    }
    if (state.knockerMelds.length > 0) {
      lines.push('melds:');
      for (const m of state.knockerMelds) {
        lines.push(`  ${formatCardList(m.cards)}`);
      }
    }
    if (state.knockerDeadwood.length > 0) {
      lines.push(`deadwood: ${formatCardList(state.knockerDeadwood)}`);
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
