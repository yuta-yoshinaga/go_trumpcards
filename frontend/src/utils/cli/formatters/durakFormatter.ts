import type { DurakResponse } from '../../../types/card';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

/** Format a Durak game state as terminal text. */
export function formatDurakState(state: DurakResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Durak'));

  const trump = state.trumpCard ? formatCard(state.trumpCard) : state.trumpSuit || '?';
  lines.push(`trump: ${trump} | stock: ${state.stockCount}`);

  const nameOf = (idx: number): string => {
    const p = state.players[idx];
    return p ? formatPlayerName(p.id, p.isHuman) : `P${idx}`;
  };
  lines.push(`attacker: ${nameOf(state.attackerIdx)} | defender: ${nameOf(state.defenderIdx)}`);
  lines.push('----------');

  // Table pairs (attack -> defense), indexed for the `defend` command.
  if (state.tablePairs.length === 0) {
    lines.push('table: (empty)');
  } else {
    lines.push('table:');
    state.tablePairs.forEach((pair, i) => {
      const def = pair.defense ? formatCard(pair.defense) : '(undefended)';
      lines.push(`  [${i}] ${formatCard(pair.attack)} -> ${def}`);
    });
  }
  lines.push('----------');

  // Opponents' card counts.
  state.players
    .filter((p) => !p.isHuman)
    .forEach((p) => {
      const done = p.isFinished ? ' (out)' : '';
      lines.push(`${formatPlayerName(p.id, false)}: ${p.cardCount} cards${done}`);
    });

  // Human hand, indexed for the `attack`/`defend` commands.
  const human = state.players.find((p) => p.isHuman);
  if (human) {
    lines.push('----------');
    const hand = human.cards.map((c, i) => `[${i}]${formatCard(c)}`).join('  ');
    lines.push(`your hand: ${hand || '(empty)'}`);
  }

  if (state.gameEndFlag) {
    const loser = state.loserIdx >= 0 ? nameOf(state.loserIdx) : null;
    lines.push('----------');
    lines.push(loser ? `game over — loser (durak): ${loser}` : 'game over — draw');
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
