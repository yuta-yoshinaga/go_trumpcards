import type { UnsunKarutaResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const PHASE_NAMES = ['Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];

/** Format an Unsun Karuta game state as terminal text. */
export function formatUnsunKarutaState(state: UnsunKarutaResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Unsun Karuta'));
  lines.push(
    `deal: ${state.roundNumber}  trick: ${state.trickNumber}/${state.trickCount}  ` +
      `phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  // **切り札はスート名で出す。** 数札の強弱がスートで逆になるので、
  // どのスートが切り札かが分からないと強さを読めない。
  lines.push(`trump: ${state.trumpSuitName}${state.trumpCard ? ` (${formatCard(state.trumpCard)})` : ''}`);
  const tricks = state.teamTricks ?? [];
  const scores = state.teamScores ?? [];
  lines.push(`ko: team0=${tricks[0] ?? 0} team1=${tricks[1] ?? 0}  match: ${scores[0] ?? 0}/${scores[1] ?? 0}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const role = p.isDealer ? 'Dealer' : 'Player';
    lines.push(`${name} (team ${p.team} / ${role}): cards=${p.cardCount} tricks=${p.trickCount}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.currentTrick.length > 0) {
    const trickParts = state.currentTrick.map((tc) => {
      const name = formatPlayerName(tc.playerIdx, state.players[tc.playerIdx]?.isHuman ?? false);
      return `${name}=${formatCard(tc.card)}`;
    });
    lines.push(`trick: ${trickParts.join(', ')}`);
  }

  // **宣言の有無は盤面の規則そのもの。** 出さないと、なぜ札が限られるのかが
  // 端末から読めない。
  if (state.mustFollow) lines.push('declared: must follow the led suit');
  if (state.canDeclare) lines.push('you lead: meri <idx> declares');

  if (state.hint && isRequestedHint(state)) {
    const indices = state.hint.cardIndices ?? [];
    lines.push(`HINT: card indices [${indices.join(', ')}] (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    lines.push(state.winnerTeam >= 0 ? `Game Over! Team ${state.winnerTeam} wins!` : 'Game Over! Draw!');
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
