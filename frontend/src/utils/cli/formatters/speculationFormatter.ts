import type { SpeculationResponse } from '../../../types/card';
import { SPECULATION_HUMAN_SEAT, SPECULATION_NO_SEAT } from '../../../types/games/speculation';
import { SpeculationPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [SpeculationPhase.FLIP]: 'TURN UP',
  [SpeculationPhase.AUCTION]: 'BIDDING',
  [SpeculationPhase.RESULT]: 'ROUND RESULT',
  [SpeculationPhase.GAME_END]: 'FINISHED',
};

/**
 * Suit symbols by the Go `CardDesign*` value (Card.go: joker 0, ♠1 ♣2 ♥3 ♦4).
 *
 * **An unknown suit prints "-", not "♠".** `trumpSuit` is -1 until the trump
 * card is turned, and a lookup that falls back to the first suit would announce
 * spades as trumps for a round that has none.
 */
const SUIT_SYMBOLS: Record<number, string> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** Format a Speculation game state as terminal text. */
export function formatSpeculationState(state: SpeculationResponse): string {
  const lines: string[] = [formatHeader('Speculation')];

  lines.push(`Phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(`Round: ${state.roundNo + 1}${state.config ? ` / ${state.config.rounds}` : ''}`);
  lines.push(
    `Trump: ${SUIT_SYMBOLS[state.trumpSuit] ?? '-'}${state.trumpCard ? ` (${formatCard(state.trumpCard)})` : ''}`,
  );
  lines.push(`Pot: ${state.pot}`);

  lines.push(formatSeparator());
  state.seats.forEach((seat, i) => {
    // **「最高札を持つ席」は -1 で「いない」。** 0 と混同すると、まだ誰も
    // 切り札を出していない盤面で人間の行に印が付く。
    const best = i === state.bestSeat ? '*' : ' ';
    const turn = i === state.turnSeat ? '>' : ' ';
    const held = seat.best ? `  [best ${formatCard(seat.best)}]` : '';
    // 伏せ札は枚数だけ。中身は送られてこないし、送られてきてはいけない。
    lines.push(`${best}${turn} ${seat.name}: ${seat.chips} chips / ${seat.hiddenCount} face down${held}`);
  });

  if (
    state.phase === SpeculationPhase.AUCTION &&
    state.offerFrom !== SPECULATION_NO_SEAT &&
    state.offerTo !== SPECULATION_NO_SEAT
  ) {
    const buyer = state.seats[state.offerFrom]?.name ?? '?';
    const owner = state.seats[state.offerTo]?.name ?? '?';
    lines.push(formatSeparator());
    lines.push(
      state.offerTo === SPECULATION_HUMAN_SEAT
        ? `${buyer} offers ${state.offerAmount} for your card. \`a\` to sell, \`d\` to decline.`
        : `${owner} will part with the best trump for ${state.offerAmount}. \`a\` to buy, \`bid <amount>\` to raise, \`d\` to pass.`,
    );
  }

  if (state.phase === SpeculationPhase.RESULT || state.gameEndFlag) {
    lines.push(formatSeparator());
    if (state.winnerSeat === SPECULATION_NO_SEAT) {
      lines.push('No trump appeared. The stakes are returned.');
    } else if (state.winnerSeat === SPECULATION_HUMAN_SEAT) {
      lines.push('You take the pot!');
    } else {
      lines.push(`${state.seats[state.winnerSeat]?.name ?? '?'} takes the pot.`);
    }
  }
  if (state.gameEndFlag && state.seats.length > 0) {
    lines.push(`Final chips: ${state.seats[SPECULATION_HUMAN_SEAT].chips}`);
  }

  return lines.join('\n');
}
