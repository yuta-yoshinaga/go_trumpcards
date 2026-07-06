import type { KoiKoiResponse } from '../../../types/card';
import { formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES = ['Play', 'KoiKoiDecision', 'RoundEnd', 'GameEnd'];

/** Format a Koi-Koi (こいこい) game state as terminal text. */
export function formatKoiKoiState(state: KoiKoiResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Koi-Koi'));
  lines.push(
    `round: ${state.roundNumber}  deck: ${state.remainingDeck}  koikoi: ${state.koikoiCount}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );

  const fieldStr = state.fieldCards.length > 0 ? formatIndexedCards(state.fieldCards) : '(empty)';
  lines.push(`field: ${fieldStr}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const yaku = p.yaku.length > 0 ? ` yaku=[${p.yaku.map((y) => `${y.key}:${y.points}`).join(' ')}]` : '';
    lines.push(`${name}: ${p.cardCount} cards  captured=${p.capturedCount}  score=${p.score}${yaku}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.phase === 1 && state.pendingYaku.length > 0) {
    const yaku = state.pendingYaku.map((y) => `${y.key}:${y.points}`).join(' ');
    lines.push(`decision: [${yaku}] = ${state.pendingPoints}  (koikoi / stop)`);
  }

  if (state.phase === 2 && state.lastRoundResult) {
    const d = state.lastRoundResult;
    const yaku = d.yaku.map((y) => `${y.key}:${y.points}`).join(' ');
    const winner = d.winner >= 0 ? formatPlayerName(d.winner, state.players[d.winner]?.isHuman ?? false) : 'draw';
    lines.push(`result: winner=${winner} [${yaku}] ${d.basePoints}×${d.multiplier}=${d.total}`);
  }

  if (state.hint) {
    const action = state.hint.koikoi === 1 ? 'koikoi' : 'stop';
    lines.push(`HINT: play ${state.hint.cardIndex} field ${state.hint.fieldIndex} (${state.hint.reason}) / ${action}`);
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
