import type { TablanetResponse } from '../../../types/card';
import { formatHeader, formatIndexedCards, formatPlayerName, formatSeparator, isRequestedHint } from '../formatterBase';

const PHASE_NAMES = ['Play', 'GameEnd'];

/** Format a Tablanet (Tablić) game state as terminal text. */
export function formatTablanetState(state: TablanetResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Tablanet'));
  lines.push(
    `deal: ${state.roundNumber}  deck: ${state.remainingDeck}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );

  const tableStr = state.tableCards.length > 0 ? formatIndexedCards(state.tableCards) : '(empty)';
  lines.push(`table: ${tableStr}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(`${name}: ${p.cardCount} cards  captured=${p.capturedCount}  tabla=${p.tablaCount}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.phase === 1 && state.lastDealDetail) {
    const d = state.lastDealDetail;
    const scores = state.players.map((p) => `${formatPlayerName(p.id, p.isHuman)}=${d.gained[p.id] ?? 0}`).join(' ');
    lines.push(`final: ${scores}`);
    if (state.winners.length > 0) {
      lines.push(
        `winner: ${state.winners.map((i) => formatPlayerName(i, state.players[i]?.isHuman ?? false)).join(', ')}`,
      );
    }
  }

  if (state.hint && isRequestedHint(state)) {
    const cards = state.hint.cardIndices ?? [];
    const targets = state.hint.tableIndices ?? [];
    lines.push(`HINT: play [${cards.join(', ')}] capture [${targets.join(', ')}] (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
