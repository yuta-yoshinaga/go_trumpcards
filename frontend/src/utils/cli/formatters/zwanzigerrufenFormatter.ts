import type { ZwanzigerrufenResponse } from '../../../types/card';
import { formatCard, formatHeader, formatIndexedCards, formatSeparator, isRequestedHint } from '../formatterBase';

const PHASE_NAMES = ['Bid', 'Talon', 'Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];
const CONTRACT_NAMES: Record<string, string> = {
  pass: '-',
  trischaken: 'Trischaken',
  rufer: 'Call the XX',
  solo: 'Solo',
};

/** Format a Zwanzigerrufen (ツヴァンツィガールーフェン) game state as terminal text. */
export function formatZwanzigerrufenState(state: ZwanzigerrufenResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Zwanzigerrufen'));
  lines.push(
    `deal: ${state.roundNumber}/${state.totalRounds}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(`contract: ${CONTRACT_NAMES[state.contractName] ?? state.contractName}`);
  if (state.calledTrump > 0) {
    // **持ち主は判明するまで書かない。** 呼び札そのものは公開情報。
    const holder = state.partnerRevealed && state.partnerIdx >= 0 ? ` (P${state.partnerIdx})` : ' (holder unknown)';
    lines.push(`called: trump ${state.calledTrump}${holder}`);
  }
  lines.push(`scores: ${state.players.map((p) => `P${p.id}=${p.score}`).join('  ')}`);
  lines.push(formatSeparator());

  for (const p of state.players) {
    const role = p.isDeclarer ? ' [declarer]' : p.isPartner ? ' [partner]' : '';
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
