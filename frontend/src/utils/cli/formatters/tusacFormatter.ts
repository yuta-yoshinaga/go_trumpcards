import type { Card, TuSacResponse } from '../../../types/card';
import { TuSacPhase } from '../../../types/phases';
import { formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [TuSacPhase.DRAW]: 'DRAW',
  [TuSacPhase.DISCARD]: 'DISCARD',
  [TuSacPhase.ROUND_END]: 'ROUND END',
  [TuSacPhase.GAME_END]: 'GAME END',
};

const MELD_NAMES: Record<number, string> = {
  1: 'three of one colour',
  2: 'chariot-horse-cannon',
  3: 'five soldiers',
};

const COLOR_LETTER: Record<string, string> = {
  gold: 'Y',
  red: 'R',
  green: 'G',
  black: 'W',
};

/**
 * Renders one card as colour + piece.
 *
 * **Not a rank and a suit.** The shared `formatCard` prints French suits and a
 * number, which say nothing about a four-colour deck — the server sends the
 * colour and the piece glyph on the card itself (ADR-0033), so this reads
 * those rather than re-deriving them.
 */
function tusacCard(card: Card | null): string {
  if (!card) return '-';
  const colour = COLOR_LETTER[card.color ?? ''] ?? '?';
  return `${colour}${card.glyph ?? '?'}`;
}

/** Format a Tu Sac game state as terminal text. */
export function formatTuSacState(state: TuSacResponse): string {
  const lines: string[] = [formatHeader('Tu Sac')];

  lines.push(`Phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(`Round: ${state.roundNumber} of ${state.rounds}`);
  lines.push(`Stock: ${state.stockCount} / Top discard: ${tusacCard(state.discardTop)}`);

  lines.push(formatSeparator());
  state.seats.forEach((seat) => {
    const mark = seat.isTurn ? '*' : ' ';
    // **相手の手札は届いていない。** 枚数と場の組み合わせだけが見える。
    const melds =
      seat.melds.length > 0
        ? seat.melds.map((m) => `${MELD_NAMES[m.kind] ?? '?'}(${m.cards.map(tusacCard).join(' ')})`).join(' ')
        : '-';
    lines.push(`${mark}${seat.name} ${seat.handCount} cards / score ${seat.score} : ${melds}`);
  });

  const human = state.seats[state.humanSeat];
  if (human && human.cards.length > 0) {
    lines.push(formatSeparator());
    // **番号は 1 始まり。** 同じ札が 4 枚あるので名前では指定できない。
    lines.push(`Your hand: ${human.cards.map((c, i) => `${i + 1}:${tusacCard(c)}`).join(' ')}`);
  }

  if (!state.gameEndFlag && state.isHumanTurn) {
    lines.push(state.phase === TuSacPhase.DRAW ? 'Draw (draw) or take the discard (take)' : 'Meld or discard');
  }
  if (state.phase === TuSacPhase.ROUND_END || state.gameEndFlag) {
    lines.push(formatSeparator());
    state.seats.forEach((seat) => {
      const out = seat.wentOut ? ' (went out)' : '';
      lines.push(`${seat.name}: melds ${seat.meldPoints} - held ${seat.handCount} = ${seat.roundScore}${out}`);
    });
  }
  if (state.gameEndFlag) {
    lines.push(`Winner: ${state.seats[state.winnerSeat]?.name ?? '?'}`);
  }

  return lines.join('\n');
}
