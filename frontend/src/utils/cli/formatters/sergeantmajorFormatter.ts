import type { SergeantMajorResponse } from '../../../types/card';
import { SergeantMajorPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [SergeantMajorPhase.TRUMP]: 'TRUMP',
  [SergeantMajorPhase.DISCARD]: 'DISCARD',
  [SergeantMajorPhase.PLAY]: 'PLAY',
  [SergeantMajorPhase.ROUND_END]: 'ROUND END',
  [SergeantMajorPhase.GAME_END]: 'GAME END',
};

/** trumpSuit is a 1-based suit code, as elsewhere in this repo. */
const SUIT_SYMBOLS: Record<number, string> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** Format a Sergeant Major game state as terminal text. */
export function formatSergeantMajorState(state: SergeantMajorResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Sergeant Major (8-5-3)'));
  lines.push(
    `round ${state.roundNumber}/${state.config.rounds} | trick ${state.trickNumber + 1}/16 | ${
      PHASE_NAMES[state.phase] ?? state.phase
    }`,
  );
  // **ノルマは席順で決まる。** 規則そのものを毎回書く。
  lines.push('targets follow the seats: dealer 8, next 5, next 3 — nobody bids');
  lines.push(
    state.trumpSuit > 0
      ? `trump: ${SUIT_SYMBOLS[state.trumpSuit] ?? '?'} (declared by ${formatPlayerName(
          state.dealerIdx,
          state.dealerIdx === 0,
        )})`
      : `trump: undeclared (the dealer chooses, and takes the ${state.kittySize}-card kitty)`,
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
    const role = p.id === state.dealerIdx ? '[dealer]' : '';
    lines.push(
      `${marker}${formatPlayerName(p.id, p.isHuman)}${role}: target ${p.target}, took ${p.trickCount} | total ${p.score} | ${p.cardCount} cards`,
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
        ? `game over — ${formatPlayerName(state.winnerIdx, state.winnerIdx === 0)} wins on points`
        : 'game over — tie',
    );
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
