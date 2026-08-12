import type { StealingBundlesResponse } from '../../../types/card';
import { StealingBundlesPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [StealingBundlesPhase.PLAY]: 'PLAY',
  [StealingBundlesPhase.GAME_END]: 'GAME END',
};

/** Format a Stealing Bundles game state as terminal text. */
export function formatStealingBundlesState(state: StealingBundlesResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Stealing Bundles'));
  lines.push(`turn ${state.turnNumber + 1} | deck ${state.deckRemaining} | ${PHASE_NAMES[state.phase] ?? state.phase}`);
  // **束の一番上が弱点、というのが規則そのもの。**
  lines.push('matching a rank captures it — and a rival bundle goes whole if its top card matches');

  lines.push('----------');
  // **空の場も情報。** 行が消えると見落としと区別が付きません。
  const table = state.tableCards.map((c) => formatCard(c)).join('  ');
  lines.push(`table: ${table || '(empty)'}`);
  lines.push('----------');

  state.players.forEach((p) => {
    const marker = p.id === state.currentPlayerIdx && !state.gameEndFlag ? '>' : ' ';
    const role = p.id === state.lastCaptureIdx ? '[captured last]' : '';
    // **一番上は全員に見えます。** そこが狙われる場所だからです。
    const top = p.bundleTop ? formatCard(p.bundleTop) : 'none';
    lines.push(
      `${marker}${formatPlayerName(p.id, p.isHuman)}${role}: ${p.cardCount} cards, bundle ${p.bundleSize}, top ${top}`,
    );
  });

  const human = state.players.find((p) => p.isHuman);
  if (human) {
    lines.push('----------');
    const hand = human.cards
      .map((c, i) => {
        const takes = state.tableMatches[String(i)]?.length ?? 0;
        const steals = state.stealTargets[String(i)] ?? [];
        const marks = `${takes > 0 ? '*' : ''}${steals.length > 0 ? `!${steals.join('')}` : ''}`;
        return `[${i}]${formatCard(c)}${marks}`;
      })
      .join('  ');
    lines.push(`your hand: ${hand || '(empty)'}`);
    lines.push('  * = captures from the table, !N = steals seat N');
  }

  if (!state.gameEndFlag) {
    // **取れるときは置けません。** 黙っていると trail が弾かれる理由が読めません。
    lines.push(
      state.canCapture
        ? 'a capture is available — you cannot place a card on the table'
        : 'nothing can be captured — place a card on the table',
    );
  } else {
    lines.push('----------');
    const winner = state.players[state.winnerIdx];
    lines.push(
      `game over — ${formatPlayerName(state.winnerIdx, state.winnerIdx === 0)} collected the most (${winner?.bundleSize ?? 0})`,
    );
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
