import type { EstimationPlayer, EstimationResponse } from '../../../types/card';
import { EstimationCall, EstimationPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [EstimationPhase.TRUMP]: 'TRUMP',
  [EstimationPhase.BID]: 'BID',
  [EstimationPhase.PLAY]: 'PLAY',
  [EstimationPhase.ROUND_END]: 'ROUND END',
  [EstimationPhase.GAME_END]: 'GAME END',
};

/** Tricks per round (thirteen cards each). */
const TRICKS_PER_ROUND = 13;

/** trumpSuit is a 1-based suit code, as elsewhere in this repo. */
const SUIT_SYMBOLS: Record<number, string> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** A seat's call, spelled out with its kind rather than as a bare number. */
function bidStr(p: EstimationPlayer): string {
  if (p.bid < 0) return 'no call';
  if (p.callType === EstimationCall.DASH) return 'Dash (0)';
  if (p.callType === EstimationCall.RISK) return `Risk (${p.bid})`;
  return `call ${p.bid}`;
}

/** Format an Estimation game state as terminal text. */
export function formatEstimationState(state: EstimationResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Estimation'));
  lines.push(
    `round ${state.roundNumber}/${state.config.rounds} | trick ${state.trickNumber + 1}/${TRICKS_PER_ROUND} | ${
      PHASE_NAMES[state.phase] ?? state.phase
    }`,
  );
  // **得点表は盤面から読めない。** Dash と Risk の振れ幅を常時出す。
  lines.push('score: exact +(10+call) / missed -(10+call); Dash (0) is ±23; Risk doubles');
  lines.push(state.trumpSuit > 0 ? `trump: ${SUIT_SYMBOLS[state.trumpSuit] ?? '?'}` : 'trump: undecided');

  // **押せない宣言があるなら先に言う。** 出してから拒否されるのでは遅い。
  if (state.restrictedBid >= 0) {
    lines.push(`you call last: ${state.restrictedBid} is barred (it would make the total 13)`);
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
      `${marker}${formatPlayerName(p.id, p.isHuman)}: ${bidStr(p)} | ${p.trickCount} tricks | ${p.totalScore} total | ${p.cardCount} cards`,
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
        ? `game over — ${formatPlayerName(state.winnerIdx, state.winnerIdx === 0)} wins`
        : 'game over — tie',
    );
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
