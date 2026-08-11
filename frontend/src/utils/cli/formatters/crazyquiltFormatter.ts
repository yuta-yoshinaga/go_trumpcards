import type { CrazyQuiltResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

/** Format a CrazyQuilt game state as terminal text. */
export function formatCrazyQuiltState(state: CrazyQuiltResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Crazy Quilt'));

  // 昇順/降順を出さないと、その本が次に何を欲しがっているか読めない。
  const fnd = state.foundation.map((pile, i) => {
    const dir = state.foundationAscending?.[i] ? '↑' : '↓';
    return `${dir}${pile.length > 0 ? formatCard(pile[pile.length - 1]) : '[  ]'}`;
  });
  lines.push(`foundations: ${fnd.join(' | ')}`);

  const wasteTop = state.waste.length > 0 ? formatCard(state.waste[state.waste.length - 1]) : '[  ]';
  lines.push(`stock: ${state.stockCount}  redeals: ${state.redealsLeft}  waste: ${wasteTop}`);
  lines.push('----------');

  // 8 行に分けて描き、**取れる札には * を付ける**。短辺の露出は向きに依存
  // するので、印が無いと盤面を見ても何が取れるのか分からない。
  for (let row = 0; row < 8; row++) {
    const cells: string[] = [];
    for (let col = 0; col < 8; col++) {
      const idx = row * 8 + col;
      const card = state.quilt[idx];
      if (!card) {
        cells.push('  . ');
        continue;
      }
      cells.push(`${state.available[idx] ? '*' : ' '}${formatCard(card)}`);
    }
    lines.push(cells.join(' '));
  }
  lines.push('* = takeable (a short side is exposed)');
  lines.push('----------');

  lines.push(`moves: ${state.moveCount}  undo:${state.canUndo ? 'yes' : 'no'}`);

  if (state.hint && isRequestedHint(state)) {
    const from = state.hint.fromIdx >= 0 ? `t${state.hint.fromIdx}` : state.hint.fromZone;
    const to = state.hint.toIdx >= 0 ? `${state.hint.toZone}${state.hint.toIdx}` : state.hint.toZone;
    lines.push(`HINT: ${from} → ${to}`);
  }
  if (state.isStalemate) lines.push('Stalemate - no more moves possible');
  if (state.message) lines.push(state.message);
  if (state.phase === 1) lines.push('Congratulations! You win!');

  lines.push(formatSeparator());
  return lines.join('\n');
}
