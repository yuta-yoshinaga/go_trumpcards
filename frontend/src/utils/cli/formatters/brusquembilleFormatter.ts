import type { BrusquembilleResponse } from '../../../types/card';
import { BrusquembillePhase } from '../../../types/phases';
import { suitSymbolAt } from '../../../utils/cardAlt';
import { formatCard, formatHeader, formatPlayerName, formatSeparator, isRequestedHint } from '../formatterBase';

const PHASE_NAMES = ['Play', 'TrickEnd', 'GameEnd'];

/** Format a Brusquembille game state as terminal text. */
export function formatBrusquembilleState(state: BrusquembilleResponse): string {
  const lines = [formatHeader('Brusquembille')];
  lines.push(
    `trick: ${state.trickNumber}  stock: ${state.stockRemaining}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(`trump: ${state.trumpCard ? formatCard(state.trumpCard) : suitSymbolAt(state.trumpSuit, 'none')}`);
  const humanTurn = state.players[state.currentPlayerIdx]?.isHuman === true;
  for (const player of state.players) {
    lines.push(`${formatPlayerName(player.id, player.isHuman)}: points=${player.points} tricks=${player.trickCount}`);
    if (player.isHuman && player.cards.length > 0) {
      lines.push(
        `  ${player.cards
          .map((card, index) => {
            const unplayable =
              humanTurn &&
              state.phase === BrusquembillePhase.PLAY &&
              state.followRequired &&
              (state.validIndices ?? []).length > 0 &&
              !state.validIndices.includes(index);
            return `[${index}]${formatCard(card)}${unplayable ? 'x' : ''}`;
          })
          .join('  ')}`,
      );
    }
  }
  if (state.currentTrick.length) {
    lines.push(
      `trick cards: ${state.currentTrick.map(({ playerIdx, card }) => `${formatPlayerName(playerIdx, state.players[playerIdx]?.isHuman ?? false)}=${formatCard(card)}`).join(', ')}`,
    );
  }
  if (state.hint && isRequestedHint(state)) {
    lines.push(
      state.hint.cardIndex !== undefined
        ? `HINT: play card [${state.hint.cardIndex}] (${state.hint.reason})`
        : `HINT: ${state.hint.reason}`,
    );
  }
  if (state.message) lines.push(state.message);
  if (state.phase === BrusquembillePhase.TRICK_END) lines.push('Press n for next trick.');
  else if (state.phase === BrusquembillePhase.PLAY && humanTurn) lines.push('Press p <idx> to play a card.');
  if (state.gameEndFlag)
    lines.push(state.winnerIdx >= 0 ? `Game Over! Player ${state.winnerIdx} wins!` : 'Game Over! Draw.');
  lines.push(formatSeparator());
  return lines.join('\n');
}
