import type { RussianBankResponse } from '../../../types/card';
import { RussianBankPhase } from '../../../types/phases';
import {
  russianBankFoundationTarget,
  russianBankTableauTargets,
  selectedRussianBankCard,
} from '../../russianBankRules';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

const ZONES = ['reserve', 'waste', 'tableau'];

function formatTop(pile: RussianBankResponse['tableau'][number]): string {
  const top = pile[pile.length - 1];
  return top ? formatCard(top) : '-';
}

/** Format a Russian Bank state for the CLI terminal. */
export function formatRussianBankState(state: RussianBankResponse): string {
  const ended = state.phase === RussianBankPhase.GAME_END || state.gameEndFlag;
  const lines = [
    formatHeader('Russian Bank'),
    `phase: ${ended ? 'GameEnd' : state.phase === RussianBankPhase.PLAYING ? 'Playing' : 'Idle'}`,
    `Turn: ${state.isHumanTurn ? 'You' : 'CPU'}`,
    `moves: ${state.moveCount}`,
  ];
  for (const player of state.players) {
    const who = player.isHuman ? 'You' : `CPU ${player.id}`;
    lines.push(
      `${who}: reserve=${player.reserveTop ? formatCard(player.reserveTop) : '-'} (${player.reserveCount}), waste=${player.wasteTop ? formatCard(player.wasteTop) : '-'} (${player.wasteCount}), hand=${player.handCount}, stop=${player.stopPoints}`,
    );
  }
  lines.push('Foundations:');
  state.foundations.forEach((pile, i) => {
    lines.push(`  [${i}] ${formatTop(pile)} (${pile.length})`);
  });
  lines.push('Tableau (top is last):');
  state.tableau.forEach((pile, i) => {
    lines.push(`  [${i}] ${formatTop(pile)}${pile.length > 1 ? ` (${pile.length} cards)` : ''}`);
  });
  const player = state.players.find((p) => p.isHuman);
  const canAct = !ended && state.phase === RussianBankPhase.PLAYING && state.isHumanTurn;
  if (canAct) {
    if (state.canCallStop) lines.push('Use s to call stop.');
    if (state.canUndo) lines.push('Use u to undo.');
    if (player) {
      const sources = state.players
        .flatMap((owner) => [
          { zone: 0, fromOpp: !owner.isHuman, col: 0 },
          { zone: 1, fromOpp: !owner.isHuman, col: 0 },
        ])
        .concat(state.tableau.map((_, col) => ({ zone: 2, fromOpp: false, col })));
      const legal = sources.flatMap((source) => {
        const card = selectedRussianBankCard(state, source);
        if (!card) return [];
        return [
          ...(russianBankFoundationTarget(state, card) >= 0
            ? [
                `${source.fromOpp ? 'opponent ' : ''}${ZONES[source.zone]}${source.zone === 2 ? ` ${source.col}` : ''} → foundation`,
              ]
            : []),
          ...[...russianBankTableauTargets(state, card, source)].map(
            (col) =>
              `${source.fromOpp ? 'opponent ' : ''}${ZONES[source.zone]}${source.zone === 2 ? ` ${source.col}` : ''} → tableau ${col}`,
          ),
        ];
      });
      const canMoveFrom = (name: string) => legal.some((move) => move.startsWith(`${name} →`));
      lines.push(
        legal.length
          ? 'Commands: pf <zone> <fromOpp:0|1> <col>, mt <zone> <fromOpp:0|1> <col> <toCol>, d, h, r'
          : 'Commands: d, h, r',
      );
      lines.push(
        `Your sources: reserve ${player.reserveTop ? (canMoveFrom('reserve') ? '*available' : 'available') : 'empty'}; waste ${player.wasteTop ? (canMoveFrom('waste') ? '*available' : 'available') : 'empty'}.`,
      );
      const opponent = state.players.find((p) => !p.isHuman);
      if (opponent) {
        lines.push(
          `Opponent sources: reserve ${opponent.reserveTop ? (canMoveFrom('opponent reserve') ? '*available' : 'available') : 'empty'}; waste ${opponent.wasteTop ? (canMoveFrom('opponent waste') ? '*available' : 'available') : 'empty'}.`,
        );
      }
      state.tableau.forEach((pile, i) => {
        const available = pile.length > 0;
        const marked = canMoveFrom(`tableau ${i}`);
        lines.push(`  Tableau column ${i}: ${available ? (marked ? '*available' : 'available') : 'empty'}`);
      });
      if (legal.length) lines.push(`Legal moves (*): ${legal.join('; ')}`);
    } else {
      lines.push('Commands: d, h, r');
    }
  } else {
    lines.push('Use r to reset.');
  }
  if (state.messageCode === 'russianbank.noHint') lines.push('No hint available.');
  else if (state.hint && isRequestedHint(state)) {
    const source =
      state.hint.zone === 2
        ? `tableau column ${state.hint.col}`
        : `${state.hint.fromOpponent ? 'opponent' : 'your'} ${ZONES[state.hint.zone] ?? 'pile'}`;
    const dest = state.hint.toFoundation ? 'its foundation' : `tableau column ${state.hint.toCol}`;
    lines.push(`HINT: move ${source} to ${dest}.`);
  }
  if (state.message && !ended && state.messageCode !== 'russianbank.noHint') lines.push(state.message);
  if (ended)
    lines.push(state.winnerIdx < 0 ? 'Game Over! Draw.' : state.winnerIdx === 0 ? 'Game Over! You win!' : 'CPU wins!');
  lines.push(formatSeparator());
  return lines.join('\n');
}
