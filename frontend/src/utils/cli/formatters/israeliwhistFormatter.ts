import type { IsraeliWhistPlayer, IsraeliWhistResponse } from '../../../types/card';
import { IsraeliWhistPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [IsraeliWhistPhase.AUCTION]: 'AUCTION',
  [IsraeliWhistPhase.BID]: 'BID',
  [IsraeliWhistPhase.PLAY]: 'PLAY',
  [IsraeliWhistPhase.ROUND_END]: 'ROUND END',
  [IsraeliWhistPhase.GAME_END]: 'GAME END',
};

/** Tricks per round (thirteen cards each). */
const TRICKS_PER_ROUND = 13;

/** trumpSuit is a 1-based suit code, as elsewhere in this repo. */
const SUIT_SYMBOLS: Record<number, string> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** A seat's standing in the auction, which the calling round does not replace. */
function roleStr(p: IsraeliWhistPlayer, declarer: boolean): string {
  if (declarer) return `won ${p.auctionBid}`;
  return p.passed ? 'passed' : 'bidding';
}

/** Format an Israeli Whist game state as terminal text. */
export function formatIsraeliWhistState(state: IsraeliWhistResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Israeli Whist'));
  lines.push(
    `round ${state.roundNumber}/${state.config.rounds} | trick ${state.trickNumber + 1}/${TRICKS_PER_ROUND} | ${
      PHASE_NAMES[state.phase] ?? state.phase
    }`,
  );
  // **得点表は盤面から読めない。** 2 乗で跳ねることと全員一致の倍率を常時出す。
  lines.push(
    'score: exact +(call^2 + 10), an exact 0 is +25, missing costs -(10 x diff); all-exact or all-miss doubles',
  );

  lines.push(
    state.trumpSuit > 0
      ? `trump: ${SUIT_SYMBOLS[state.trumpSuit] ?? '?'} (won with ${state.highBid})`
      : `auction: high bid ${state.highBid} ${SUIT_SYMBOLS[state.highSuit] ?? '-'}`,
  );

  // **押せない宣言があるなら先に言う。** 出してから拒否されるのでは遅い。
  if (state.minimumBid > 0) lines.push(`you won the auction: call at least ${state.minimumBid}`);
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
    const call = p.bid < 0 ? 'no call' : `call ${p.bid}`;
    lines.push(
      `${marker}${formatPlayerName(p.id, p.isHuman)}[${roleStr(p, p.id === state.declarerIdx)}]: ${call} | ${p.trickCount} tricks | ${p.totalScore} total | ${p.cardCount} cards`,
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
