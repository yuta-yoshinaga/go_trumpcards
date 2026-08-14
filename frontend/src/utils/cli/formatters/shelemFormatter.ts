import type { ShelemPlayer, ShelemResponse } from '../../../types/card';
import { ShelemPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [ShelemPhase.BID]: 'BID',
  [ShelemPhase.DISCARD]: 'DISCARD',
  [ShelemPhase.PLAY]: 'PLAY',
  [ShelemPhase.ROUND_END]: 'ROUND END',
  [ShelemPhase.GAME_END]: 'GAME END',
};

/** Tricks per round (twelve cards each). */
const TRICKS_PER_ROUND = 12;

/** trumpSuit is a 1-based suit code, as elsewhere in this repo. */
const SUIT_SYMBOLS: Record<number, string> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** A seat's standing in the bidding. */
function roleStr(p: ShelemPlayer, declarer: boolean): string {
  if (declarer && p.declaredShelem) return 'declared Shelem';
  if (declarer) return `won ${p.bid}`;
  if (p.passed) return 'passed';
  return p.bid >= 0 ? `bid ${p.bid}` : 'bidding';
}

/** Format a Shelem game state as terminal text. */
export function formatShelemState(state: ShelemResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Shelem'));
  lines.push(
    `round ${state.roundNumber} | trick ${state.trickNumber + 1}/${TRICKS_PER_ROUND} | first to ${
      state.config.target
    } | ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  // **点になるのは A/10/5 だけ。** 盤面から読めないので常時出す。
  lines.push('point cards: A and 10 are 10, the 5 is 5 — a round holds exactly 100');
  lines.push(`total: yours=${state.scores[0] ?? 0} theirs=${state.scores[1] ?? 0}`);

  if (state.declarerIdx >= 0) {
    const who = formatPlayerName(state.declarerIdx, state.declarerIdx === 0);
    lines.push(
      state.shelemBid
        ? `contract: Shelem — every trick (${who})`
        : `contract: ${state.contract} (${who}; ${state.roundPoints[state.declarerIdx % 2] ?? 0} so far)`,
    );
  } else {
    lines.push(`contract: undecided (bid at least ${state.minBid})`);
  }
  if (state.trumpSuit > 0) lines.push(`trump: ${SUIT_SYMBOLS[state.trumpSuit] ?? '?'}`);
  if (state.widowSize > 0) lines.push(`widow: ${state.widowSize} cards face down`);

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
    lines.push(
      `${marker}${formatPlayerName(p.id, p.isHuman)}[T${p.team}]: ${roleStr(p, p.id === state.declarerIdx)} | ${p.trickCount} tricks | ${p.cardCount} cards`,
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
      state.winnerTeam >= 0
        ? `game over — team ${state.winnerTeam} wins (${state.scores[0] ?? 0} - ${state.scores[1] ?? 0})`
        : `game over — tie (${state.scores[0] ?? 0} - ${state.scores[1] ?? 0})`,
    );
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
