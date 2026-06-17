import type { HandAndFootResponse } from '../../../types/card';
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

/** Format a Hand and Foot game state as terminal text. */
export function formatHandAndFootState(state: HandAndFootResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Hand and Foot'));
  lines.push(`round: ${state.roundNumber}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(
    `stock: ${state.drawPileCount} | discard: ${state.discardTop ? formatCard(state.discardTop) : '[  ]'} (${state.discardPileCount})${state.isFrozen ? ' [FROZEN]' : ''}`,
  );
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const tags: string[] = [];
    tags.push(`team${p.team}`);
    if (p.inFoot) tags.push('In Foot');
    const tagStr = tags.length > 0 ? ` [${tags.join(', ')}]` : '';
    lines.push(
      `${name}: total=${p.cumulativeScore} round=${p.roundScore} hand=${p.cardCount} foot=${p.footCount}${tagStr}`,
    );
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  for (const team of state.teams) {
    const meldCount = team.melds.length;
    if (meldCount === 0 && team.red3Count === 0) continue;
    lines.push(`Team ${team.team}: red3=${team.red3Count}`);
    for (const m of team.melds) {
      const type = m.isCanasta ? '(canasta)' : m.isNatural ? '(natural)' : '';
      lines.push(`  meld: ${formatCardList(m.cards)} ${type}`);
    }
  }
  lines.push('----------');

  if (!state.gameEndFlag) {
    const current = formatPlayerName(state.currentPlayerIdx, state.players[state.currentPlayerIdx]?.isHuman ?? false);
    lines.push(`turn: ${current}`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    lines.push(`Game Over! Winner: Team ${state.winnerTeam}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
