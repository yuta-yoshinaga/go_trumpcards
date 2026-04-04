import type { DoubtResponse } from '../../../types/card';
import { formatCardList, formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

const VALUE_NAMES: Record<number, string> = { 1: 'A', 11: 'J', 12: 'Q', 13: 'K' };

function valueName(v: number): string {
  return VALUE_NAMES[v] ?? String(v);
}

/** Format a Doubt game state as terminal text. */
export function formatDoubtState(state: DoubtResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Doubt'));
  lines.push(`table: ${state.tableCardCount} cards`);
  lines.push('');

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

  if (state.lastAction) {
    const name = formatPlayerName(
      state.lastAction.playerIdx,
      state.players[state.lastAction.playerIdx]?.isHuman ?? false,
    );
    lines.push(`${name} played ${state.lastAction.cardCount} as "${valueName(state.lastAction.claimedValue)}"`);
  }

  if (state.lastDoubtResult) {
    const r = state.lastDoubtResult;
    const doubter = formatPlayerName(r.doubterIdx, state.players[r.doubterIdx]?.isHuman ?? false);
    lines.push(`${doubter} doubted: ${r.wasLying ? 'CAUGHT!' : 'WRONG!'}`);
    if (r.revealedCards.length > 0) lines.push(`  revealed: ${formatCardList(r.revealedCards)}`);
  }

  if (state.phase === 1) lines.push('Doubt window open!');

  if (!state.gameEndFlag) {
    const current = formatPlayerName(state.currentTurn, state.players[state.currentTurn]?.isHuman ?? false);
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
