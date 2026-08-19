import type { RollingStoneResponse } from '../../../types/card';
import { RollingStonePhase } from '../../../types/phases';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const SUIT_SYMBOLS: Record<number, string> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

const PHASE_NAMES: Record<number, string> = {
  [RollingStonePhase.PLAY]: 'PLAY',
  [RollingStonePhase.GAME_END]: 'GAME END',
};

/** Format a Rolling Stone game state as terminal text. */
export function formatRollingStoneState(state: RollingStoneResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Rolling Stone'));
  lines.push(
    `trick ${state.trickNumber + 1} | deck ${state.deckSize} (${state.deckSize - state.discarded} still in play) | ${
      PHASE_NAMES[state.phase] ?? state.phase
    }`,
  );
  // **勝利条件が逆さまなのが規則そのもの。** 毎回書く。
  lines.push('winning a trick scores nothing — run out of cards first, and failing to follow hands you the trick');

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
    const role = p.finishedAt > 0 ? `[out in ${p.finishedAt}]` : p.id === state.lastPickupIdx ? '[just picked up]' : '';
    lines.push(`${marker}${formatPlayerName(p.id, p.isHuman)}${role}: ${p.cardCount} cards, picked up ${p.pickups}x`);
  });

  const human = state.players.find((p) => p.isHuman);
  if (human) {
    lines.push('----------');
    const hand = human.cards
      .map((c, i) => `[${i}]${formatCard(c)}${state.validPlays.includes(i) ? '*' : ''}`)
      .join('  ');
    lines.push(`your hand: ${hand || '(empty)'}`);
  }

  // **出せる札が無いことははっきり言う。** 黙っていると打てない理由が分からない。
  if (state.mustPickUp) {
    lines.push(
      `you cannot follow ${SUIT_SYMBOLS[state.leadSuit] ?? '?'} — pickup takes the ${state.currentTrick.length} cards on the table`,
    );
  }

  if (state.gameEndFlag) {
    lines.push('----------');
    const winner = state.players[state.winnerIdx];
    if (winner && winner.cardCount > 0) {
      // **上限で切った局は「上がった」わけではない。**
      lines.push(
        `game over — nobody ran out; ${formatPlayerName(state.winnerIdx, state.winnerIdx === 0)} wins with the fewest (${winner.cardCount})`,
      );
    } else {
      lines.push(`game over — ${formatPlayerName(state.winnerIdx, state.winnerIdx === 0)} ran out first`);
    }
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
