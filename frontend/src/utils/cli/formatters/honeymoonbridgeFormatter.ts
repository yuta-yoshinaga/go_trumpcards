import type { HoneymoonBridgeResponse } from '../../../types/card';
import { HoneymoonBridgePhase } from '../../../types/phases';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [HoneymoonBridgePhase.DRAW]: 'DRAW',
  [HoneymoonBridgePhase.BID]: 'BID',
  [HoneymoonBridgePhase.PLAY]: 'PLAY',
  [HoneymoonBridgePhase.ROUND_END]: 'ROUND END',
  [HoneymoonBridgePhase.GAME_END]: 'GAME END',
};

/** Contract suits. **`0` is no-trump**, which is a bid, not a missing value. */
const SUIT_SYMBOLS: Record<number, string> = { 0: 'NT', 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** Format a Honeymoon Bridge game state as terminal text. */
export function formatHoneymoonBridgeState(state: HoneymoonBridgeResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Honeymoon Bridge'));
  lines.push(
    `deal ${state.roundNumber} | trick ${state.trickNumber + 1}/13 | target ${state.config.target} | ${
      PHASE_NAMES[state.phase] ?? state.phase
    }`,
  );

  // **前半のトリックは得点にならない。** 山札の残りだけが意味を持つ。
  if (state.phase === HoneymoonBridgePhase.DRAW) {
    lines.push(`stock: ${state.stockSize} left — tricks in this half do not score`);
  }

  lines.push(
    state.contractLevel > 0
      ? `contract: ${state.contractLevel}${SUIT_SYMBOLS[state.trumpSuit] ?? '?'} by ${formatPlayerName(
          state.declarerIdx,
          state.declarerIdx === 0,
        )} — needs ${state.requiredTricks} tricks`
      : 'contract: not yet decided',
  );

  // **通る最小の宣言を出す。** これが無いと拒否される値を打ち込むことになる。
  if (state.phase === HoneymoonBridgePhase.BID) {
    lines.push(
      state.minBidLevel > 0
        ? `lowest bid that outbids: ${state.minBidLevel}${SUIT_SYMBOLS[state.minBidSuit] ?? '?'}`
        : 'the contract is at the ceiling (7NT) — pass is the only move',
    );
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
    const role = p.id === state.declarerIdx ? '[declarer]' : '';
    const bid = p.bidLevel > 0 ? `bid ${p.bidLevel}${SUIT_SYMBOLS[p.bidSuit] ?? '?'}` : 'no bid';
    lines.push(
      `${marker}${formatPlayerName(p.id, p.isHuman)}${role}: ${bid}, took ${p.trickCount} | total ${p.score} | ${p.cardCount} cards`,
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

  if (state.phase === HoneymoonBridgePhase.ROUND_END && state.contractLevel > 0) {
    lines.push('----------');
    // **CLI モードも「何点動いたか」を出す** (#5760 レビュー)。ページのバナーと
    // Go の CUI が出しているのに、ここだけトリックの過不足で止まっていた。
    const points = state.lastPoints.toString();
    lines.push(
      state.lastMade
        ? `contract made: ${state.lastTricks} of ${state.requiredTricks} tricks (+${points} to the declarer)`
        : `contract down: ${state.lastTricks} of ${state.requiredTricks} tricks (+${points} to the opponent)`,
    );
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
