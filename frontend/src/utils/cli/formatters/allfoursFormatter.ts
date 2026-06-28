import type { AllFoursResponse } from '../../../types/card';
import { AllFoursPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [AllFoursPhase.BEG]: 'BEG',
  [AllFoursPhase.GIFT]: 'GIFT',
  [AllFoursPhase.PLAY]: 'PLAY',
  [AllFoursPhase.TRICK_END]: 'TRICK END',
  [AllFoursPhase.ROUND_END]: 'DEAL END',
  [AllFoursPhase.GAME_END]: 'GAME END',
};

const SUIT_NAMES: Record<number, string> = {
  0: '(unset)',
  1: 'Spades',
  2: 'Clubs',
  3: 'Hearts',
  4: 'Diamonds',
};

/** Format an All Fours game state as terminal text. */
export function formatAllFoursState(state: AllFoursResponse): string {
  const lines: string[] = [];
  lines.push(formatHeader('All Fours'));
  lines.push(
    `round: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`,
  );
  const turnUp = state.turnUp ? formatCard(state.turnUp) : '--';
  lines.push(`dealer: ${state.dealerIdx}  trump: ${SUIT_NAMES[state.trumpSuit] ?? '?'}  turn-up: ${turnUp}`);
  for (const p of state.players) {
    const tag = p.isHuman ? 'YOU' : `CPU${p.id}`;
    lines.push(`[${tag}] tricks=${p.trickCount} round=${p.roundScore} total=${p.cumulativeScore} cards=${p.cardCount}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  hand: ${p.cards.map((c, i) => `[${i}]${formatCard(c)}`).join(' ')}`);
    }
  }
  if (state.currentTrick.length > 0) {
    lines.push(`trick: ${state.currentTrick.map((tc) => `P${tc.playerIdx}:${formatCard(tc.card)}`).join('  ')}`);
  }
  if (state.message) lines.push(state.message);
  lines.push(formatSeparator());
  return lines.join('\n');
}
