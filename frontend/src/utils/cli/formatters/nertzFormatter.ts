import type { NertzResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  0: 'INIT',
  1: 'PLAYING',
  2: 'ROUND_END',
  3: 'MATCH_END',
};

/** Format a Nertz game state as terminal text. */
export function formatNertzState(state: NertzResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Nertz'));
  lines.push(
    `round: ${state.roundNumber}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}  target: ${state.targetScore}`,
  );

  // Foundations
  const fnd = state.foundations.map((f) => (f.top ? formatCard(f.top) : '[  ]'));
  lines.push(`foundation: ${fnd.join(' | ')}`);
  lines.push('----------');

  // Per-player view (humans show full tableau, CPU shows summary)
  for (const p of state.players) {
    const tag = p.isHuman ? '[YOU]' : '[CPU]';
    lines.push(`${tag} ${p.name} score=${p.score} stock=${p.stockSize} waste=${p.wasteSize} nertz=${p.nertzSize}`);
    if (p.isHuman) {
      const wasteStr = p.wasteTop ? formatCard(p.wasteTop) : '[  ]';
      const nertzStr = p.nertzTop ? formatCard(p.nertzTop) : '[  ]';
      lines.push(`  waste: ${wasteStr}  nertz: ${nertzStr}`);
      for (let col = 0; col < p.tableau.length; col++) {
        const column = p.tableau[col];
        if (column.length === 0) {
          lines.push(`  t${col}: [empty]`);
          continue;
        }
        const cardStrs = column.map((c, i) => (c.card ? `[${i}]${formatCard(c.card)}` : '[?]'));
        lines.push(`  t${col}: ${cardStrs.join(' ')}`);
      }
    }
  }
  lines.push('----------');

  lines.push(`moves: ${state.moveCount}  undo:${state.canUndo ? 'yes' : 'no'}`);

  if (state.hint && isRequestedHint(state)) {
    const fromCol = state.hint.fromCol >= 0 ? state.hint.fromCol : '';
    const toCol = state.hint.toCol >= 0 ? state.hint.toCol : '';
    lines.push(`HINT: ${state.hint.fromZone}${fromCol}[${state.hint.cardIndex}] → ${state.hint.toZone}${toCol}`);
  }
  if (state.message) lines.push(state.message);
  if (state.matchWinner >= 0) {
    lines.push(`Match Winner: ${state.players[state.matchWinner]?.name ?? 'unknown'}`);
  } else if (state.winnerIdx >= 0) {
    lines.push(`Round Winner: ${state.players[state.winnerIdx]?.name ?? 'unknown'}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
