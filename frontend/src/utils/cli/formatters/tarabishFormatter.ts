import type { TarabishPlayer, TarabishResponse } from '../../../types/card';
import { TarabishPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [TarabishPhase.BID]: 'BID',
  [TarabishPhase.PLAY]: 'PLAY',
  [TarabishPhase.ROUND_END]: 'ROUND END',
  [TarabishPhase.GAME_END]: 'GAME END',
};

/** Tricks per round (nine cards each). */
const TRICKS_PER_ROUND = 9;

/** trumpSuit is a 1-based suit code, as elsewhere in this repo. */
const SUIT_SYMBOLS: Record<number, string> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** A seat's meld, spelled out rather than as a bare number. */
function meldStr(p: TarabishPlayer): string {
  if (p.meldPoints === 0) return 'no meld';
  const parts = [p.runLen > 0 && `run of ${p.runLen}`, p.hasBella && 'bella'].filter(Boolean);
  return `${parts.join('+')}=${p.meldPoints}`;
}

/** Format a Tarabish game state as terminal text. */
export function formatTarabishState(state: TarabishResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Tarabish'));
  lines.push(
    `round ${state.roundNumber} | trick ${state.trickNumber + 1}/${TRICKS_PER_ROUND} | first to ${
      state.config.target
    } | ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(`score: yours=${state.scores[0] ?? 0} theirs=${state.scores[1] ?? 0}`);
  // **切り札の序列はこの系統の肝。** 盤面から読めないので常時出す。
  lines.push('trump: J(Jass)=20 > 9(Menel)=14 > A=11 > 10 > K=4 > Q=3');

  if (state.trumpTakerIdx >= 0) {
    const who = formatPlayerName(state.trumpTakerIdx, state.players[state.trumpTakerIdx]?.isHuman ?? false);
    lines.push(`trump: ${SUIT_SYMBOLS[state.trumpSuit] ?? '?'} (taken by ${who})`);
  } else if (state.upCard) {
    lines.push(`turned for trump: ${formatCard(state.upCard)}`);
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
    lines.push(
      `${marker}${formatPlayerName(p.id, p.isHuman)}[T${p.team}]: ${meldStr(p)} | ${p.trickCount} tricks | ${p.cardCount} cards`,
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
