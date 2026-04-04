import type { MemoryResponse } from '../../../types/card';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

/** Format a Memory game state as terminal text. */
export function formatMemoryState(state: MemoryResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Memory'));
  lines.push(`turn: #${state.turnNumber}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(`${name}: ${p.pairCount} pairs`);
  }
  lines.push('----------');

  // Board grid (assume square-ish)
  const size = Math.ceil(Math.sqrt(state.board.length));
  for (let row = 0; row < size; row++) {
    const cells: string[] = [];
    for (let col = 0; col < size; col++) {
      const idx = row * size + col;
      if (idx >= state.board.length) {
        cells.push('    ');
        continue;
      }
      const bc = state.board[idx];
      if (bc.taken) {
        cells.push('    ');
      } else if (bc.faceUp && bc.card) {
        cells.push(formatCard(bc.card).padEnd(4));
      } else {
        cells.push(`[${String(idx).padStart(2)}]`);
      }
    }
    lines.push(cells.join(' '));
  }
  lines.push('----------');

  if (state.lastMatchResult) lines.push('Match!');

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
