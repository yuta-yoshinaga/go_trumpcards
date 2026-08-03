import { type AluetteResponse, aluetteLuetteName, aluetteTeamOf } from '../../../types/card';
import { formatCard, formatHeader, formatPlayerName, formatSeparator, isRequestedHint } from '../formatterBase';

const PHASE_NAMES = ['Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];

/** Format an Aluette game state as terminal text. */
export function formatAluetteState(state: AluetteResponse): string {
  const lines: string[] = [];
  const luettes = state.luettes ?? [];

  lines.push(formatHeader('Aluette'));
  lines.push(
    `mene: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(
    `scores: team0=${state.teamScores[0]}  team1=${state.teamScores[1]} (first to ${state.config.targetPoints})`,
  );
  // **序列表を毎フレーム出す。**強さは値ではなく札ごとに決まるので、六枚を
  // 覚えるまでは手札を目で並べ替えることすらできない。
  lines.push(`luettes: ${luettes.map((l) => `${l.name}(${l.design[0]}${l.value})`).join(' > ')}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const dealer = p.isDealer ? ' [dealer]' : '';
    lines.push(`${name} (team ${p.team}): cards=${p.cardCount} tricks=${p.trickCount}${dealer}`);
    if (p.isHuman && p.cards.length > 0) {
      // 手札のリュエットには呼び名を添える。"D3" だけでは最強札と読めない。
      const hand = p.cards
        .map((c, i) => {
          const luette = aluetteLuetteName(luettes, c);
          return `[${i}]${formatCard(c)}${luette ? `<${luette}>` : ''}`;
        })
        .join(' ');
      lines.push(`  ${hand}`);
    }
  }
  lines.push('----------');

  if (state.currentTrick.length > 0) {
    const parts = state.currentTrick.map((tc) => {
      const name = formatPlayerName(tc.playerIdx, state.players[tc.playerIdx]?.isHuman ?? false);
      const luette = aluetteLuetteName(luettes, tc.card);
      return `${name}=${formatCard(tc.card)}${luette ? `<${luette}>` : ''}`;
    });
    lines.push(`trick: ${parts.join(', ')}`);
  }

  if (state.phase === 2 || state.phase === 3) {
    const tricks = state.roundTricks.map((v, i) => `P${i}(T${aluetteTeamOf(i)})=${v}`).join(' ');
    lines.push(`mene result: tricks ${tricks}`);
  }

  if (state.hint && isRequestedHint(state)) {
    const indices = state.hint.cardIndices ?? [];
    lines.push(`HINT: card indices [${indices.join(', ')}] (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    // A draw leaves winnerTeam at -1; "Team -1 wins" would be a lie.
    lines.push(state.winnerTeam >= 0 ? `Game Over! Team ${state.winnerTeam} wins!` : 'Game Over! Draw.');
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
