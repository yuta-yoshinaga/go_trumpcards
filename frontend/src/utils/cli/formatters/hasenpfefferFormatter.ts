import type { HasenpfefferResponse } from '../../../types/card';
import { HasenpfefferPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [HasenpfefferPhase.BID]: 'BID',
  [HasenpfefferPhase.DISCARD]: 'DISCARD',
  [HasenpfefferPhase.PLAY]: 'PLAY',
  [HasenpfefferPhase.HAND_END]: 'HAND END',
  [HasenpfefferPhase.GAME_END]: 'GAME END',
};

/** trumpSuit is a 1-based suit code, as elsewhere in this repo. */
const SUIT_SYMBOLS: Record<number, string> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** Render one seat's bid: not yet, passed, or a number. */
function bidText(bid: number): string {
  if (bid < 0) return '-';
  if (bid === 0) return 'passed';
  return String(bid);
}

/** Format a Hasenpfeffer game state as terminal text. */
export function formatHasenpfefferState(state: HasenpfefferResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Hasenpfeffer'));
  lines.push(
    `hand ${state.handNumber} | trick ${state.trickNumber + 1}/6 | first to ${state.config.target} | ${
      PHASE_NAMES[state.phase] ?? state.phase
    }`,
  );
  // **ジョーカーが全カード中最強。** 序列を知らないと打ち方が変わる。
  lines.push('the joker is the highest trump of all (Best Bower), then the right and left bowers');
  lines.push(`score: yours=${state.scores[0] ?? 0} theirs=${state.scores[1] ?? 0}`);

  if (state.trumpSuit > 0) {
    lines.push(
      `trump: ${SUIT_SYMBOLS[state.trumpSuit] ?? '?'} (${formatPlayerName(
        state.declarerIdx,
        state.declarerIdx === 0,
      )} bid ${state.contract})`,
    );
  } else if (state.blindSize > 0) {
    lines.push(`blind: ${state.blindSize} card (the declarer takes it and discards one)`);
  } else {
    lines.push('trump: not named yet');
  }
  // **親は降りられないことがある。** 選択肢が無い場面を明示する。
  if (state.mustBid) {
    lines.push('everyone else passed — as dealer you cannot pass');
  } else if (state.phase === HasenpfefferPhase.BID && state.minBid === 0) {
    lines.push('the maximum bid is standing; you can only pass');
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
    const role = p.id === state.declarerIdx ? '[declarer]' : p.id === state.dealerIdx ? '[dealer]' : '';
    lines.push(
      `${marker}${formatPlayerName(p.id, p.isHuman)}[T${p.team}]${role}: bid ${bidText(p.bid)} | took ${p.trickCount} | ${p.cardCount} cards`,
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

  // **落としたのか達成したのかは盤面から読めない。**
  if (state.phase === HasenpfefferPhase.HAND_END) {
    lines.push('----------');
    lines.push(
      state.lastHandEuchred
        ? `hand over — contract ${state.contract}, took ${state.lastHandTricks}: euchred`
        : `hand over — contract ${state.contract}, took ${state.lastHandTricks}: made`,
    );
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
