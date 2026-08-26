import type { RamschResponse } from '../../../types/card';
import {
  formatCard,
  formatCardList,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
} from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  0: 'PLAY',
  1: 'TRICK_END',
  2: 'ROUND_END',
  3: 'GAME_END',
};

/** Format a Ramsch game state as terminal text. */
export function formatRamschState(state: RamschResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Ramsch'));
  lines.push(
    `round: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`,
  );
  // **切り札と得点の向きを毎回書く。** どちらもスカート系のつもりで読むと
  // 逆に取るところで、書いていなければ盤が読めない。
  lines.push('trump: jacks (C > S > H > D)   scoring: most card points LOSES');
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const turnMark = p.id === state.currentPlayerIdx ? ' *' : '';
    lines.push(
      `${name}${turnMark}: cumScore=${p.cumulativeScore} round=${p.roundScore} tricks=${p.trickCount} cardPts=${p.cardPoints}`,
    );
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  hand: ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.currentTrick.length > 0) {
    const trick = state.currentTrick.map((tc) => `P${tc.playerIdx}:${formatCard(tc.card)}`).join('  ');
    lines.push(`trick: ${trick}`);
  }

  // 伏せ札はサーバがラウンド終了まで返さないので、来ていれば出してよい。
  if (state.skat && state.skat.length > 0) {
    lines.push(`skat: ${formatCardList(state.skat)}`);
  }

  if (state.durchmarsch && state.durchmarschIdx >= 0) {
    lines.push(`Durchmarsch! P${state.durchmarschIdx} took every trick — the others lose 120 each.`);
  } else if (state.loserIdx >= 0) {
    const loser = state.players[state.loserIdx];
    lines.push(`P${state.loserIdx} took the most (${loser?.cardPoints ?? 0}) and loses that many.`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) lines.push('Game Over!');

  lines.push(formatSeparator());
  return lines.join('\n');
}
