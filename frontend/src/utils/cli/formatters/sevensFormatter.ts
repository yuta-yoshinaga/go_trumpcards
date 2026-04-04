import type { SevensResponse } from '../../../types/card';
import { formatCard, formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

const SUIT_NAMES: Record<number, string> = { 1: 'Spade', 2: 'Clover', 3: 'Heart', 4: 'Diamond' };

/** Format a Sevens game state as terminal text. */
export function formatSevensState(state: SevensResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Sevens'));

  // Board state
  for (let s = 1; s <= 4; s++) {
    const suit = SUIT_NAMES[s] ?? '?';
    const min = state.tableMinVals[s] ?? 0;
    const max = state.tableMaxVals[s] ?? 0;
    lines.push(`  ${suit}: ${min > 0 ? min : '-'} ... 7 ... ${max > 0 ? max : '-'}`);
  }
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    if (p.isFinished) {
      lines.push(`${name} finished (rank: ${p.rank})`);
    } else {
      lines.push(`${name} ${p.cardCount} cards (passes: ${p.passesUsed}/${p.maxPasses})`);
    }
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.cpuActions.length > 0) {
    for (const a of state.cpuActions) {
      const name = formatPlayerName(a.playerIdx, false);
      if (a.playedCard) {
        lines.push(`${name} played ${formatCard(a.playedCard)}`);
      } else {
        lines.push(`${name} passed${a.forcedPass ? ' (forced)' : ''}`);
      }
    }
  }

  if (state.humanAction) {
    const a = state.humanAction;
    if (a.playedCard) {
      lines.push(`You played ${formatCard(a.playedCard)}`);
    } else {
      lines.push(`You passed${a.forcedPass ? ' (forced)' : ''}`);
    }
  }

  if (!state.gameEndFlag) {
    const current = formatPlayerName(state.currentTurn, state.players[state.currentTurn]?.isHuman ?? false);
    lines.push(`turn: ${current}`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) lines.push('Game Over');

  lines.push(formatSeparator());
  return lines.join('\n');
}
