import type { KalookiResponse } from '../../../types/card';
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
  2: 'ROUND END',
  3: 'GAME END',
};

/** Format a Kalooki game state as terminal text. */
export function formatKalookiState(state: KalookiResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Kalooki'));
  lines.push(`phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(`opening: ${state.openingThreshold}`);
  lines.push(`discard: ${state.discardTop ? formatCard(state.discardTop) : '[  ]'} | stock: ${state.drawPileCount}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const opened = (p.hasOpened ?? false) ? 'opened' : 'not opened';
    lines.push(`${name}: score=${p.cumulativeScore} (+${p.roundScore}) cards=${p.cardCount} [${opened}]`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
    if (p.melds.length > 0) {
      lines.push('  melds:');
      // `layoff <playerIdx> <meldIdx> <cardIdx>` は狙うメルドの番号を要求するのに、
      // 場は札を並べるだけで番号がどこにも出ていなかった (#6462)。コマンドが取る
      // 0 始まりの添字をそのまま出す ── 手札の `[0]` 表記と同じ形。
      p.melds.forEach((m, mi) => {
        lines.push(`    [${mi}] ${formatCardList(m.cards)}`);
      });
    }
  }
  lines.push('----------');

  if (!state.gameEndFlag) {
    const current = formatPlayerName(state.currentPlayerIdx, state.players[state.currentPlayerIdx]?.isHuman ?? false);
    lines.push(`turn: ${current}`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    if (state.winnerIdx < 0) {
      lines.push('Game Over! Draw');
    } else {
      const winner = formatPlayerName(state.winnerIdx, state.players[state.winnerIdx]?.isHuman ?? false);
      lines.push(`Game Over! Winner: ${winner}`);
    }
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
