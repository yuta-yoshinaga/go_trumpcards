import type { SevenTwentySevenResponse } from '../../../types/card';
import { formatHeader, formatIndexedCards, formatPlayerName, formatSeparator, isRequestedHint } from '../formatterBase';

const PHASE_NAMES = ['Draw', 'Result'];

/** Format a SevenTwentySeven game state as terminal text. */
export function formatSevenTwentySevenState(state: SevenTwentySevenResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('SevenTwentySeven'));
  lines.push(
    `round: ${state.roundNumber}  draw: ${state.drawRound}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(`pot: ${state.pot}  ante: ${state.ante}`);
  // **2 つの目標を毎回書く。** 7 と 27 のどちらに寄せるかがこのゲームそのもの。
  lines.push('targets: 7 and 27 (over = out on that side). faces = 0.5, ace = 1 or 11');
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const status = p.out
      ? 'OUT'
      : p.wonLow && p.wonHigh
        ? 'SCOOP'
        : p.wonLow
          ? 'won low'
          : p.wonHigh
            ? 'won high'
            : p.standing
              ? 'stood pat'
              : 'drawing';
    const scores = p.lowScore || p.highScore ? `  ${p.lowScore || '?'} / ${p.highScore || '?'}` : '';
    lines.push(`${name}: chips=${p.chips} bet=${p.roundBet} [${status}]${scores}`);
    if (p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.hint && isRequestedHint(state)) {
    const call = state.hint.draw ? 'take a card' : 'stand pat';
    lines.push(`HINT: ${call} (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.matchWinnerIdx >= 0) {
    lines.push(
      `Game Over! Winner: ${formatPlayerName(state.matchWinnerIdx, state.players[state.matchWinnerIdx]?.isHuman ?? false)}`,
    );
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
