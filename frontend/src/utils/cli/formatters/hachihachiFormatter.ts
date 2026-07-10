import type { HachiHachiResponse } from '../../../types/card';
import { formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES = ['Play', 'RoundEnd', 'GameEnd'];

/** Format a Hachi-Hachi (八八) game state as terminal text. */
export function formatHachiHachiState(state: HachiHachiResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Hachi-Hachi'));
  lines.push(
    `round: ${state.roundNumber}  deck: ${state.remainingDeck}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );

  const fieldStr = state.fieldCards.length > 0 ? formatIndexedCards(state.fieldCards) : '(empty)';
  lines.push(`field: ${fieldStr}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const yaku = p.yaku.length > 0 ? ` yaku=[${p.yaku.map((y) => `${y.key}:${y.points}`).join(' ')}]` : '';
    lines.push(
      `${name}: ${p.cardCount} cards  captured=${p.capturedCount}  raw=${p.rawScore}  score=${p.score}${yaku}`,
    );
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.phase === 1 && state.lastRoundResult) {
    lines.push('result:');
    for (const s of state.lastRoundResult.scores) {
      const yaku = s.yaku.length > 0 ? ` [${s.yaku.map((y) => `${y.key}:${y.points}`).join(' ')}]` : '';
      const mark = s.playerIdx === state.lastRoundResult.best ? ' *' : '';
      const sign = s.delta >= 0 ? `+${s.delta}` : `${s.delta}`;
      lines.push(`  P${s.playerIdx}: raw=${s.rawScore}+bonus=${s.bonus} → ${sign}${yaku}${mark}`);
    }
  }

  if (state.hint) {
    lines.push(`HINT: play ${state.hint.cardIndex} field ${state.hint.fieldIndex} (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
