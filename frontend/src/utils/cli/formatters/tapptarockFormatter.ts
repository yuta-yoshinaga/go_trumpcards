import type { TappTarockResponse } from '../../../types/card';
import { formatCard, formatHeader, formatIndexedCards, formatSeparator, isRequestedHint } from '../formatterBase';

const PHASE_NAMES = ['Bid', 'Talon', 'Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];
const CONTRACT_NAMES: Record<string, string> = {
  pass: '-',
  trischaken: 'Trischaken',
  dreier: 'Dreier',
  solo: 'Solo',
};

/** Format a TappTarock (タップ・タロック) game state as terminal text. */
export function formatTappTarockState(state: TappTarockResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('TappTarock'));
  lines.push(
    `deal: ${state.roundNumber}/${state.totalRounds}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(`contract: ${CONTRACT_NAMES[state.contractName] ?? state.contractName}`);
  lines.push(`scores: ${state.players.map((p) => `P${p.id}=${p.score}`).join('  ')}`);
  lines.push(formatSeparator());

  for (const p of state.players) {
    const role = p.isDeclarer ? ' [declarer]' : '';
    lines.push(`P${p.id}${p.isHuman ? ' (you)' : ''}${role}: ${p.cardCount} cards  ${p.cardPoints} pts`);
  }

  if (state.currentTrick.length > 0) {
    lines.push(formatSeparator());
    lines.push(`trick: ${state.currentTrick.map((tc) => `P${tc.playerIdx}:${formatCard(tc.card)}`).join('  ')}`);
  }

  const human = state.players.find((p) => p.isHuman);
  if (human && human.cards.length > 0) {
    lines.push(formatSeparator());
    lines.push(formatIndexedCards(human.cards));
  }

  if (isRequestedHint(state) && state.hint) {
    lines.push(`hint: ${state.hint.reason}`);
  }
  if (state.message) lines.push(state.message);
  return lines.join('\n');
}
