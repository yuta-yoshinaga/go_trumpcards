import type { BeziqueResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const PHASE_NAMES = ['Play', 'Meld', 'RoundEnd', 'GameEnd'];
const SUIT_SYMBOLS = ['none', '♠', '♣', '♥', '♦'];
const MELD_NAMES = ['marriage', 'Bezique', 'four aces', 'four kings', 'four queens', 'four jacks'];

/** Returns a short human-readable label for a meld (type + suit + points). */
function meldLabel(m: { type: number; suit: number; points: number }): string {
  const base = MELD_NAMES[m.type] ?? `meld ${m.type}`;
  const suit = m.suit >= 1 && m.suit <= 4 ? ` ${SUIT_SYMBOLS[m.suit]}` : '';
  return `${base}${suit} (${m.points})`;
}

/** Format a Bezique game state as terminal text. */
export function formatBeziqueState(state: BeziqueResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Bezique'));
  lines.push(
    `deal: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  const trumpText = state.trumpSuit >= 1 && state.trumpSuit <= 4 ? SUIT_SYMBOLS[state.trumpSuit] : 'undeclared';
  const trumpCardText = state.trumpCard ? ` (${formatCard(state.trumpCard)})` : '';
  lines.push(`trump: ${trumpText}${trumpCardText}`);
  lines.push(`stock: ${state.stockRemaining}${state.isEndgame ? '  [endgame — must follow]' : ''}`);
  lines.push(`match score: P0=${state.matchScore[0] ?? 0}  P1=${state.matchScore[1] ?? 0}`);
  lines.push(`deal points: P0=${state.dealPoints[0] ?? 0}  P1=${state.dealPoints[1] ?? 0}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(`${name}: cards=${p.cardCount} tricks=${p.trickCount} dealPts=${p.roundScore}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.currentTrick.length > 0) {
    const trickParts = state.currentTrick.map((tc) => {
      const name = formatPlayerName(tc.playerIdx, state.players[tc.playerIdx]?.isHuman ?? false);
      return `${name}=${formatCard(tc.card)}`;
    });
    lines.push(`trick: ${trickParts.join(', ')}`);
  }

  if (state.phase === 1 && state.availableMelds.length > 0) {
    lines.push('available melds:');
    state.availableMelds.forEach((m, i) => {
      lines.push(`  [${i}] ${meldLabel(m)}`);
    });
  }

  if (state.hint && isRequestedHint(state)) {
    if (state.hint.meldIndex != null) {
      lines.push(
        state.hint.meldIndex < 0
          ? `HINT: skip the meld (${state.hint.reason})`
          : `HINT: declare meld index [${state.hint.meldIndex}] (${state.hint.reason})`,
      );
    } else if (state.hint.cardIndex != null) {
      lines.push(`HINT: play card index [${state.hint.cardIndex}] (${state.hint.reason})`);
    }
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.winnerIdx >= 0) {
    lines.push(
      `Game Over! Winner: ${formatPlayerName(state.winnerIdx, state.players[state.winnerIdx]?.isHuman ?? false)}`,
    );
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
