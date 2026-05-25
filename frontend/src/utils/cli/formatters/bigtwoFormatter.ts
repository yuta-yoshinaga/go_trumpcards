import type { BigTwoResponse } from '../../../types/card';
import { formatCardList, formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

/** Format a Big Two game state as terminal text. */
export function formatBigTwoState(state: BigTwoResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Big Two'));

  // Players
  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    if (p.isFinished) {
      lines.push(`${name} finished (rank: ${p.rank})`);
    } else {
      lines.push(`${name} ${p.cardCount} cards`);
    }
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }

  lines.push('----------');

  // Table
  if (state.tableCards.length > 0) {
    const lastPlayer = formatPlayerName(
      state.lastPlayPlayerIdx,
      state.players[state.lastPlayPlayerIdx]?.isHuman ?? false,
    );
    lines.push(`table: ${formatCardList(state.tableCards)} (by ${lastPlayer})`);
  } else {
    lines.push('table: empty (anyone can play)');
  }

  // CPU actions
  if (state.cpuActions.length > 0) {
    for (const a of state.cpuActions) {
      const name = formatPlayerName(a.playerIdx, false);
      if (!a.playedCards || a.playedCards.length === 0) {
        lines.push(`${name} passed`);
      } else {
        lines.push(`${name} played ${formatCardList(a.playedCards)}`);
      }
    }
  }

  // Human action
  if (state.humanAction) {
    const a = state.humanAction;
    if (!a.playedCards || a.playedCards.length === 0) {
      lines.push('You passed');
    } else {
      lines.push(`You played ${formatCardList(a.playedCards)}`);
    }
  }

  // Turn info
  if (!state.gameEndFlag) {
    const current = formatPlayerName(state.currentTurn, state.players[state.currentTurn]?.isHuman ?? false);
    lines.push(`turn: ${current}`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) lines.push('Game Over');

  lines.push(formatSeparator());
  return lines.join('\n');
}
