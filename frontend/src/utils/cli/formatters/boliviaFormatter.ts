import type { BoliviaMeldData, BoliviaResponse } from '../../../types/card';
import { BOLIVIA_MELD_KIND } from '../../../types/games/bolivia';
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

/** Describe a meld's kind/completion state for terminal display. */
function meldLabel(m: BoliviaMeldData): string {
  // **3 種類ある。** `isBolivia` が真になるのはワイルドのメルド (kind 2) だけで、
  // シーケンス (kind 1) では決して真にならない ── 2 分岐のままだと完成した
  // ボリビアが `(set)` と表示される (レビュー指摘)。
  if (m.kind === BOLIVIA_MELD_KIND.ESCALERA) {
    return m.isEscalera ? '(escalera)' : '(sequence)';
  }
  if (m.kind === BOLIVIA_MELD_KIND.WILD) {
    return m.isBolivia ? '(bolivia)' : '(wild)';
  }
  if (m.isCanasta) return '(canasta)';
  return m.isNatural ? '(natural)' : '(set)';
}

/** Format a Bolivia game state as terminal text. */
export function formatBoliviaState(state: BoliviaResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Bolivia'));
  lines.push(`round: ${state.roundNumber}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(
    `stock: ${state.drawPileCount} | discard: ${state.discardTop ? formatCard(state.discardTop) : '[  ]'} (${state.discardPileCount})${state.isFrozen ? ' [FROZEN]' : ''}`,
  );
  const t0 = state.teamScores[0] ?? 0;
  const t1 = state.teamScores[1] ?? 0;
  lines.push(`team 0: ${t0} | team 1: ${t1}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const tags: string[] = [];
    if (p.hasCanasta) tags.push('Canasta');
    // **エスカレラの印を落とさない。** 上がりを止めているのはこちらなので、
    // 誰が持っているかが席の行に出ていないと、卓の状況が読めない。
    if (p.hasEscalera) tags.push('Escalera');
    if (p.hasBolivia) tags.push('Bolivia');
    if (p.hasInitMeld) tags.push('Init Meld');
    const tagStr = tags.length > 0 ? ` [${tags.join(', ')}]` : '';
    lines.push(
      `${name} (team ${p.team}): total=${p.cumulativeScore} round=${p.roundScore} cards=${p.cardCount} red3=${p.red3Count}${tagStr}`,
    );
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
    if (p.melds.length > 0) {
      for (const m of p.melds) {
        lines.push(`  meld: ${formatCardList(m.cards)} ${meldLabel(m)}`);
      }
    }
  }
  lines.push('----------');

  if (!state.gameEndFlag) {
    const current = formatPlayerName(state.currentPlayerIdx, state.players[state.currentPlayerIdx]?.isHuman ?? false);
    lines.push(`turn: ${current}`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    lines.push(`Game Over! Winning team: ${state.winnerIdx}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
