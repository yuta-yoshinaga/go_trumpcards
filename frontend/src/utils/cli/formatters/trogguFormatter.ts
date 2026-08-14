import type { TrogguResponse } from '../../../types/card';
import { formatCard, formatHeader, formatIndexedCards, formatSeparator, isRequestedHint } from '../formatterBase';

const PHASE_NAMES = ['Bid', 'Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];
const CONTRACT_NAMES: Record<string, string> = {
  pass: '-',
  trois: 'Trois (3 tricks)',
  solo: 'Solo (most points)',
  piccolo: 'Piccolo (exactly 1 trick)',
  misere: 'Misere (no tricks)',
};

/** Format a Troggu (トロッグ) game state as terminal text. */
export function formatTrogguState(state: TrogguResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Troggu'));
  lines.push(
    `deal: ${state.roundNumber}/${state.totalRounds}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(`contract: ${CONTRACT_NAMES[state.contractName] ?? state.contractName}`);
  lines.push(`scores: ${state.players.map((p) => `P${p.id}=${p.score}`).join('  ')}`);
  lines.push(formatSeparator());

  for (const p of state.players) {
    const role = p.isDeclarer ? ' [declarer]' : '';
    lines.push(
      `P${p.id}${p.isHuman ? ' (you)' : ''}${role}: ${p.cardCount} cards  ${p.trickCount} tricks  ${p.cardPoints} pts`,
    );
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

  if (state.breakdown) {
    // **契約ごとに単位が違う。** ソロは点、他はトリック。
    const bd = state.breakdown;
    const got = bd.targetIsTricks ? `${bd.declarerTricks} tricks` : `${bd.declarerPoints} pts`;
    lines.push(`result: ${bd.won ? 'made' : 'failed'} — ${got} (target ${bd.target})`);
  }

  if (isRequestedHint(state) && state.hint) {
    lines.push(`hint: ${state.hint.reason}`);
  }
  if (state.message) lines.push(state.message);
  return lines.join('\n');
}
