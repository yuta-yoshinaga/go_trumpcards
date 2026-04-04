import type { OldMaidResponse } from '../../../types/card';
import { formatCard, formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

/** Format an Old Maid game state as terminal text. */
export function formatOldmaidState(state: OldMaidResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Old Maid'));

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    if (p.isFinished) {
      lines.push(`${name} finished`);
    } else {
      lines.push(`${name} ${p.cardCount} cards`);
    }
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.cpuActions.length > 0) {
    for (const a of state.cpuActions) {
      const name = formatPlayerName(a.drawPlayerIdx, false);
      const from = formatPlayerName(a.drawFromIdx, state.players[a.drawFromIdx]?.isHuman ?? false);
      const card = a.drawnCard ? formatCard(a.drawnCard) : '?';
      lines.push(`${name} drew ${card} from ${from} (${a.discardedPairs} pairs discarded)`);
    }
  }

  if (state.humanAction) {
    const a = state.humanAction;
    const from = formatPlayerName(a.drawFromIdx, false);
    const card = a.drawnCard ? formatCard(a.drawnCard) : '?';
    lines.push(`You drew ${card} from ${from} (${a.discardedPairs} pairs discarded)`);
  }

  if (!state.gameEndFlag) {
    const current = formatPlayerName(state.currentTurn, state.players[state.currentTurn]?.isHuman ?? false);
    const target = formatPlayerName(state.nextDrawTargetIdx, state.players[state.nextDrawTargetIdx]?.isHuman ?? false);
    lines.push(`turn: ${current} draws from ${target}`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) lines.push('Game Over');

  lines.push(formatSeparator());
  return lines.join('\n');
}
