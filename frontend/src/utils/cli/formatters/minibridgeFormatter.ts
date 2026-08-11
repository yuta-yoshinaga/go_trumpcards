import type { MinibridgeResponse } from '../../../types/card';
import { MinibridgePhase } from '../../../types/phases';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [MinibridgePhase.CONTRACT]: 'CONTRACT',
  [MinibridgePhase.PLAY]: 'PLAY',
  [MinibridgePhase.ROUND_END]: 'ROUND END',
  [MinibridgePhase.GAME_END]: 'GAME END',
};

/** Contract denominations. **`0` is no-trump**, which is a choice, not a blank. */
const SUIT_SYMBOLS: Record<number, string> = { 0: 'NT', 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** Format a Minibridge game state as terminal text. */
export function formatMinibridgeState(state: MinibridgeResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Minibridge'));
  lines.push(
    `deal ${state.roundNumber}/${state.config.rounds} | trick ${state.trickNumber + 1}/13 | ${
      PHASE_NAMES[state.phase] ?? state.phase
    }`,
  );
  // **競りが無いこと自体が規則。** 毎回書く。
  lines.push('no auction: everyone announces HCP (40 in total) and the stronger pair declares');

  lines.push(
    state.contractLevel > 0
      ? `contract: ${state.contractLevel}${SUIT_SYMBOLS[state.contractSuit] ?? '?'} by ${formatPlayerName(
          state.declarerIdx,
          state.declarerIdx === 0,
        )} — needs ${state.requiredTricks} tricks`
      : 'contract: not yet chosen',
  );
  lines.push(`totals: your pair ${state.teamScores[0] ?? 0} | their pair ${state.teamScores[1] ?? 0}`);

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
    const role = p.id === state.declarerIdx ? '[declarer]' : p.id === state.dummyIdx ? '[dummy]' : '';
    lines.push(
      `${marker}${formatPlayerName(p.id, p.isHuman)}${role}: team ${p.team}, HCP ${p.hcp}, took ${p.trickCount} | ${p.cardCount} cards`,
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

  // **ダミーは契約が決まってから公開される。** デクレアラーはここからも出す。
  if (state.dummyHand.length > 0) {
    lines.push(`dummy's hand: ${state.dummyHand.map((c, i) => `[${i}]${formatCard(c)}`).join('  ')}`);
  }

  if (state.phase === MinibridgePhase.ROUND_END) {
    lines.push('----------');
    lines.push(
      state.lastMade
        ? `contract made: ${state.lastTricks} of ${state.requiredTricks} tricks`
        : `contract down: ${state.lastTricks} of ${state.requiredTricks} tricks`,
    );
  }

  if (state.gameEndFlag) {
    lines.push('----------');
    lines.push(
      state.winnerTeam >= 0
        ? `game over — ${state.winnerTeam === 0 ? 'your pair' : 'their pair'} wins on points`
        : 'game over — tie',
    );
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
