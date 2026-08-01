import type { EcarteResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const PHASE_NAMES = ['Exchange', 'Play', 'RoundEnd', 'GameEnd'];
const NEG_STEP_NAMES = ['ElderDecide', 'DealerRespond', 'ElderDiscard', 'DealerDiscard'];
const SUIT_SYMBOLS = ['none', '♠', '♣', '♥', '♦'];

/** Format an Écarté game state as terminal text. */
export function formatEcarteState(state: EcarteResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Écarté'));
  lines.push(
    `deal: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  if (state.phase === 0) {
    lines.push(`negotiation: ${NEG_STEP_NAMES[state.negStep] ?? state.negStep}`);
  }
  const trumpText = state.trumpSuit >= 1 && state.trumpSuit <= 4 ? SUIT_SYMBOLS[state.trumpSuit] : 'undeclared';
  const trumpCardText = state.trumpCard ? ` (${formatCard(state.trumpCard)})` : '';
  lines.push(`trump: ${trumpText}${trumpCardText}`);
  lines.push(`stock: ${state.stockRemaining}${state.refusalByDealer ? '  [dealer refused]' : ''}`);
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

  if (state.hint && isRequestedHint(state)) {
    if (state.hint.cardIndex != null) {
      lines.push(`HINT: play card index [${state.hint.cardIndex}] (${state.hint.reason})`);
    } else if (state.hint.action) {
      lines.push(`HINT: ${state.hint.action} (${state.hint.reason})`);
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
