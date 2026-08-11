import type { MendikotResponse } from '../../../types/card';
import { MendikotPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [MendikotPhase.PLAY]: 'PLAY',
  [MendikotPhase.HAND_END]: 'HAND END',
  [MendikotPhase.GAME_END]: 'GAME END',
};

/** trumpSuit is a 1-based suit code, as elsewhere in this repo. */
const SUIT_SYMBOLS: Record<number, string> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** How the previous hand was decided, and what it was worth. */
const HAND_END_TEXT: Record<string, (team: number) => string> = {
  tens: (t) => `hand over — team ${t} took more tens (+1)`,
  tricks: (t) => `hand over — tens split two apiece, so team ${t} wins on tricks (+1)`,
  mendikot: (t) => `hand over — team ${t} took all four tens: Mendikot, +2`,
  whitewash: (t) => `hand over — team ${t} took every trick: Whitewash, +3`,
};

/** Format a Mendikot game state as terminal text. */
export function formatMendikotState(state: MendikotResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Mendikot'));
  lines.push(
    `hand ${state.handNumber} | first to ${state.config.target} hands | ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  // **勝敗を決めるのは 10 の枚数。** 盤面から読めないので常に出す。
  lines.push(
    `tens: yours=${state.teamTens[0] ?? 0} theirs=${state.teamTens[1] ?? 0} (of ${state.tensInDeck}; three takes the hand)`,
  );
  lines.push(`tricks: yours=${state.teamTricks[0] ?? 0} theirs=${state.teamTricks[1] ?? 0} (decides a 2-2 split)`);
  lines.push(`hand points: yours=${state.scores[0] ?? 0} theirs=${state.scores[1] ?? 0}`);
  lines.push(
    state.trumpSuit > 0
      ? `trump: ${SUIT_SYMBOLS[state.trumpSuit] ?? '?'} (set by the first player who could not follow)`
      : 'trump: undecided (the first player who cannot follow sets it with the card they play)',
  );

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
    const role = p.id === state.trumpChooserIdx ? '[set trump]' : '';
    lines.push(
      `${marker}${formatPlayerName(p.id, p.isHuman)}[T${p.team}]${role}: ${p.tens} tens | ${p.trickCount} tricks | ${p.cardCount} cards`,
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

  // **決まり方で 1/2/3 点と変わる。** どれだったか言わないと得点が飛んで見える。
  if (state.phase === MendikotPhase.HAND_END && state.lastHandWinner >= 0) {
    lines.push('----------');
    const text = HAND_END_TEXT[state.lastHandKind];
    lines.push(text ? text(state.lastHandWinner) : `hand over — team ${state.lastHandWinner} takes it`);
  }

  if (state.gameEndFlag) {
    lines.push('----------');
    lines.push(
      state.winnerTeam >= 0
        ? `game over — team ${state.winnerTeam} wins (${state.scores[0] ?? 0} - ${state.scores[1] ?? 0})`
        : `game over — tie (${state.scores[0] ?? 0} - ${state.scores[1] ?? 0})`,
    );
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
