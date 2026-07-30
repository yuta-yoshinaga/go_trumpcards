import type { Card, LaughAndLieDownPlayer, LaughAndLieDownResponse } from '../../../types/card';
import { formatCard, formatHeader } from '../formatterBase';

function formatList(cards: Card[], indexed: boolean): string {
  if (cards.length === 0) return '-';
  return cards.map((c, i) => (indexed ? `${i}:${formatCard(c)}` : formatCard(c))).join(' ');
}

/** Render one seat. A hidden seat arrives with a count and no hand cards. */
function formatSeat(p: LaughAndLieDownPlayer, dealerIdx: number): string {
  const who = p.isHuman ? 'you' : `cpu${p.id.toString()}`;
  const role = p.id === dealerIdx ? ' (dealer)' : '';
  const hand = p.hidden ? `${p.cardCount.toString()} cards` : formatList(p.cards, true);
  const down = p.laidDown ? ' -- laid down' : '';
  return `${who}${role}: ${p.wonCount.toString()} won${down}\n  hand: ${hand}`;
}

/** Format a Laugh and Lie Down game state as terminal text. */
export function formatLaughAndLieDownState(state: LaughAndLieDownResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Laugh and Lie Down'));
  lines.push(`pot ${state.pot.toString()} · dealer seat ${state.dealerIdx.toString()}`);
  // 「1 枚か 3 枚」と「取れなければ手札を場に置いて降りる」がこのゲームで最も
  // 間違えやすい 2 点なので、毎フレーム出す。
  lines.push('capture one OR three of a rank / cannot capture -> your whole hand joins the table');
  // 場は伏せた山ではないので全部出す。どのランクが何枚残っているかが見えて
  // いないと 3 枚取りの判断ができない。
  lines.push(`table: ${formatList(state.layout, false)}`);

  for (const p of state.players) {
    lines.push(formatSeat(p, state.dealerIdx));
  }

  if (state.threeTakeIndices.length > 0) {
    lines.push(`three-card takes available: ${state.threeTakeIndices.join(' ')}`);
  }

  if (state.gameEndFlag) {
    if (state.lastInIdx >= 0) {
      lines.push(`last in: seat ${state.lastInIdx.toString()}`);
    }
    for (const p of state.players) {
      const who = p.isHuman ? 'you' : `cpu${p.id.toString()}`;
      lines.push(`${who}: ${p.wonCount.toString()} won -> ${p.score.toString()}`);
    }
  }

  return lines.join('\n');
}
