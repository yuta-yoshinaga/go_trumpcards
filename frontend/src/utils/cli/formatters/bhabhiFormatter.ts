import type { BhabhiResponse } from '../../../types/card';
import { BhabhiPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [BhabhiPhase.PLAY]: 'PLAY',
  [BhabhiPhase.GAME_END]: 'GAME END',
};

/** leadSuit is a 1-based suit code, as elsewhere in this repo. */
const SUIT_SYMBOLS: Record<number, string> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** Format a Bhabhi game state as terminal text. */
export function formatBhabhiState(state: BhabhiResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Bhabhi'));
  lines.push(
    `trick ${state.trickNumber + 1} | ${state.players.length} players | ${state.aliveCount} still in | ${
      PHASE_NAMES[state.phase] ?? state.phase
    }`,
  );
  // **勝者ではなく敗者を決めるゲーム。** 目的を毎回書く。
  lines.push('the last player still holding cards is the Bhabhi (the loser)');
  lines.push(
    state.leadSuit > 0
      ? `led: ${SUIT_SYMBOLS[state.leadSuit] ?? '?'} (${state.pile.length} on the table — fail to follow and you take them all)`
      : 'the table is empty — lead whatever you like',
  );

  lines.push('----------');

  if (state.pile.length > 0) {
    const pile = state.pile.map((tc) => `${formatPlayerName(tc.playerIdx, false)}:${formatCard(tc.card)}`).join('  ');
    lines.push(`pile: ${pile}`);
    lines.push('----------');
  }

  state.players.forEach((p) => {
    const marker = p.id === state.currentPlayerIdx && !state.gameEndFlag ? '>' : ' ';
    // **順位は上がった順であって強さではない。**
    const status = p.rank > 0 ? `out, ${p.rank} to finish` : `${p.cardCount} cards`;
    lines.push(`${marker}${formatPlayerName(p.id, p.isHuman)}: ${status} | ${p.pickups} pickups`);
  });

  const human = state.players.find((p) => p.isHuman);
  if (human) {
    lines.push('----------');
    const hand = human.cards
      .map((c, i) => `[${i}]${formatCard(c)}${state.validPlays.includes(i) ? '*' : ''}`)
      .join('  ');
    lines.push(`your hand: ${hand || '(empty)'}`);
  }

  // **直前の引き取りは盤面に痕跡が残らない。** 何枚どこへ行ったか言う。
  if (state.lastPickupIdx >= 0 && !state.gameEndFlag) {
    lines.push(
      `${formatPlayerName(state.lastPickupIdx, state.lastPickupIdx === 0)} just picked up ${state.lastPickupSize} cards`,
    );
  }

  if (state.gameEndFlag) {
    lines.push('----------');
    const who = state.bhabhiIdx >= 0 ? formatPlayerName(state.bhabhiIdx, state.bhabhiIdx === 0) : '?';
    lines.push(
      state.stalemate
        ? `deadlocked after ${state.trickNumber} tricks — ${who} holds the most cards and is the Bhabhi`
        : `game over — ${who} is the Bhabhi`,
    );
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
