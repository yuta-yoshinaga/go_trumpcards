import type { TeenDoPaanchResponse } from '../../../types/card';
import { TeenDoPaanchPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [TeenDoPaanchPhase.TRUMP]: 'TRUMP',
  [TeenDoPaanchPhase.PLAY]: 'PLAY',
  [TeenDoPaanchPhase.ROUND_END]: 'ROUND END',
  [TeenDoPaanchPhase.GAME_END]: 'GAME END',
};

/** trumpSuit is a 1-based suit code, as elsewhere in this repo. */
const SUIT_SYMBOLS: Record<number, string> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** Format a 3-2-5 game state as terminal text. */
export function formatTeenDoPaanchState(state: TeenDoPaanchResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('3-2-5'));
  lines.push(
    `round ${state.roundNumber}/${state.config.rounds} | trick ${state.trickNumber + 1}/10 | ${
      PHASE_NAMES[state.phase] ?? state.phase
    }`,
  );
  // **ノルマは宣言ではなく割り当て。** 多く取っても得点にならないことも書く。
  lines.push('targets are assigned (3/2/5), not bid — making your number is all that scores');
  lines.push(
    state.trumpSuit > 0
      ? `trump: ${SUIT_SYMBOLS[state.trumpSuit] ?? '?'}`
      : 'trump: undeclared (the 5-target seat chooses from its first five)',
  );
  // **前ラウンドの札のやり取りは盤面に痕跡が残らない。**
  if (state.lastExchange > 0) {
    lines.push(`${state.lastExchange} cards changed hands for last round's shortfall`);
  }

  lines.push('----------');

  if (state.currentTrick.length > 0) {
    const trick = state.currentTrick
      .map((tc) => `${formatPlayerName(tc.playerIdx, false)}:${formatCard(tc.card)}`)
      .join('  ');
    lines.push(`trick: ${trick}`);
    lines.push('----------');
  }

  state.players.forEach((p) => {
    const marker = p.id === state.currentPlayerIdx && !state.gameEndFlag ? '>' : ' ';
    const role = p.id === state.fivePlayerIdx ? '[trump]' : '';
    lines.push(
      `${marker}${formatPlayerName(p.id, p.isHuman)}${role}: target ${p.target}, took ${p.trickCount} | met ${p.met} | ${p.cardCount} cards`,
    );
  });

  const human = state.players.find((p) => p.isHuman);
  if (human) {
    lines.push('----------');
    const hand = human.cards
      .map((c, i) => `[${i}]${formatCard(c)}${state.validPlays.includes(i) ? '*' : ''}`)
      .join('  ');
    lines.push(`your hand: ${hand || '(empty)'}`);
  }

  if (state.gameEndFlag) {
    lines.push('----------');
    lines.push(
      state.winnerIdx >= 0
        ? `game over — ${formatPlayerName(state.winnerIdx, state.winnerIdx === 0)} wins on targets met`
        : 'game over — tie',
    );
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
