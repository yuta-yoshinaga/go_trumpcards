import type { PasurResponse } from '../../../types/card';
import { PasurPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [PasurPhase.PLAY]: 'PLAY',
  [PasurPhase.GAME_END]: 'GAME END',
};

/** Format a Pasur game state as terminal text. */
export function formatPasurState(state: PasurResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Pasur'));
  lines.push(
    `pack ${state.packsDealt} | deck ${state.deckRemaining} left | ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  // **11 の合計と絵札の扱いが規則そのもの。** 毎回書く。
  lines.push('play a card so it and table numerals add to 11; J/Q/K take the same rank only');

  // **場の札には番号を振る。** `p <i> <t...>` の t はこの番号。
  lines.push(
    state.table.length > 0
      ? `table: ${state.table.map((c, i) => `[${i}]${formatCard(c)}`).join('  ')}`
      : 'table: empty',
  );

  lines.push('----------');

  state.players.forEach((p) => {
    const marker = p.id === state.currentPlayerIdx && !state.gameEndFlag ? '>' : ' ';
    const role = p.id === state.lastCaptureIdx ? '[last capture]' : '';
    lines.push(
      `${marker}${formatPlayerName(p.id, p.isHuman)}${role}: ${p.capturedCount} taken, ${p.soors} soor | score ${p.score} | ${p.cardCount} cards`,
    );
  });

  const human = state.players.find((p) => p.isHuman);
  if (human) {
    lines.push('----------');
    const hand = human.cards
      .map((c, i) => {
        // **取れる札に印を付ける。** 取れるときは場に置けないので、見分けが要る。
        const canCapture = (state.captureOptions[i]?.length ?? 0) > 0;
        return `[${i}]${formatCard(c)}${canCapture ? '*' : ''}`;
      })
      .join('  ');
    lines.push(`your hand: ${hand || '(empty)'}`);
    lines.push('* = can capture (capturing is compulsory for those cards)');
  }

  if (state.gameEndFlag) {
    lines.push('----------');
    lines.push(
      state.winners.length === 1
        ? `game over — ${formatPlayerName(state.winners[0], state.winners[0] === 0)} wins on points`
        : `game over — ${state.winners.length} players tie`,
    );
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
