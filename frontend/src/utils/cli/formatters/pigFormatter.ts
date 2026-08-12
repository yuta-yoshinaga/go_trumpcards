import type { PigResponse } from '../../../types/card';
import { PigPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [PigPhase.PASS]: 'PASS',
  [PigPhase.SIGNAL]: 'SIGNAL',
  [PigPhase.ROUND_END]: 'ROUND END',
  [PigPhase.GAME_END]: 'GAME END',
};

/** Format a Pig game state as terminal text. */
export function formatPigState(state: PigResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Pig'));
  lines.push(
    `round ${state.roundNumber} | deck ${state.deckSize} | ${state.passCount} passes | ${
      PHASE_NAMES[state.phase] ?? state.phase
    }`,
  );
  // **取り合うものが無いのが規則そのもの。** 遅れることだけが負け。
  lines.push('four of a kind is signalled in silence — the last player to notice takes a letter of P-I-G');

  lines.push('----------');

  state.players.forEach((p) => {
    const marker = p.id === state.currentPlayerIdx && state.phase === PigPhase.PASS ? '>' : ' ';
    const role = p.eliminated
      ? '[out]'
      : p.noticedOrder > 0
        ? `[noticed ${p.noticedOrder}]`
        : p.hasChosenPass
          ? '[has chosen]'
          : '';
    lines.push(
      `${marker}${formatPlayerName(p.id, p.isHuman)}${role}: ${p.cardCount} cards, letters [${p.letterWord || '-'}]`,
    );
  });

  const human = state.players.find((p) => p.isHuman);
  if (human) {
    lines.push('----------');
    const hand = human.cards.map((c, i) => `[${i}]${formatCard(c)}`).join('  ');
    lines.push(`your hand: ${hand || '(empty)'}`);
  }

  if (!state.gameEndFlag) {
    if (state.phase === PigPhase.SIGNAL) {
      lines.push(
        human?.hasSignalled
          ? `you signalled — ${state.noticedCnt} have noticed so far`
          : 'somebody put a hand to their nose — signal now, the last to react takes a letter',
      );
    } else if (state.phase === PigPhase.ROUND_END && state.roundLoserIdx >= 0) {
      const loser = state.players[state.roundLoserIdx];
      lines.push(
        `${formatPlayerName(state.roundLoserIdx, state.roundLoserIdx === 0)} was last to notice — letters [${loser?.letterWord ?? ''}]`,
      );
      lines.push('next — deal the next round');
    } else if (human?.eliminated) {
      // **脱落しても局は続く。** 打てない理由を名乗らないと固まったように見える。
      lines.push('you are out — the remaining players will finish it between them');
    } else if (human?.hasChosenPass) {
      lines.push('you have chosen — waiting for everyone else');
    }
  }

  if (state.gameEndFlag) {
    lines.push('----------');
    lines.push(`game over — ${formatPlayerName(state.winnerIdx, state.winnerIdx === 0)} was the last one standing`);
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
