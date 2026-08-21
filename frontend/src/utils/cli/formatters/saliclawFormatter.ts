import type { SalicLawResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

/** Format a SalicLaw game state as terminal text. */
export function formatSalicLawState(state: SalicLawResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Salic Law'));

  const fnd = state.foundation.map((pile) => (pile.length > 0 ? formatCard(pile[pile.length - 1]) : '[  ]'));
  lines.push(`foundations: ${fnd.join(' | ')}`);

  // 捨て札は無い。配った札はそのまま列に乗る。
  lines.push(`stock: ${state.stockCount}  open columns: ${state.openPiles}/${state.tableau.length}`);
  lines.push(`queens out of play: ${state.queens.map(formatCard).join(' ')}`);
  lines.push('----------');

  for (let pile = 0; pile < state.tableau.length; pile++) {
    const cards = state.tableau[pile];
    if (cards.length === 0) {
      // まだ K が出ていない列。置き先ではないので「空き」とは書かない。
      lines.push(`t${pile}: [not open] (a king opens it)`);
      continue;
    }
    const body = cards.map((c, i) => `[${i}]${formatCard(c)}`).join(' ');
    // K だけの列がこのゲーム唯一の置き場所なので、目印を付ける。
    lines.push(cards.length === 1 ? `t${pile}: ${body} (king only)` : `t${pile}: ${body}`);
  }
  lines.push('----------');

  lines.push(`moves: ${state.moveCount}  undo:${state.canUndo ? 'yes' : 'no'}`);

  if (state.hint && isRequestedHint(state)) {
    if (state.hint.fromZone === 'stock') {
      // 「配れ」は移動ではない。移動の体裁に落とすと行き先の無い -1 が出る。
      lines.push('HINT: deal another card from the stock');
    } else {
      const from = `t${state.hint.fromIdx}`;
      const to = state.hint.toIdx >= 0 ? `${state.hint.toZone}${state.hint.toIdx}` : state.hint.toZone;
      lines.push(`HINT: ${from} → ${to}`);
    }
  }
  if (state.isStalemate) lines.push('Stalemate - no more moves possible');
  if (state.message) lines.push(state.message);
  if (state.phase === 1) lines.push('Congratulations! You win!');

  lines.push(formatSeparator());
  return lines.join('\n');
}
