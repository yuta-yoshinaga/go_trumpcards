import type { RussianBankResponse } from '../../../types/card';
import { RussianBankPhase } from '../../../types/phases';
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
    lines.push('Commands: pf <zone> <fromOpp:0|1> <col>, mt <zone> <fromOpp:0|1> <col> <toCol>, d, h, r');
    if (state.canCallStop) lines.push('Use s to call stop.');
    if (state.canUndo) lines.push('Use u to undo.');
    if (player) {
      lines.push(
        `Your sources: reserve ${player.reserveTop ? 'available' : 'empty'}; waste ${player.wasteTop ? 'available' : 'empty'}.`,
      );
      state.tableau.forEach((pile, i) => {
        lines.push(`  Tableau column ${i}: ${pile.length ? 'available' : 'empty'}`);
      });
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
    lines.push(
      state.winnerIdx < 0
        ? 'Game Over! Draw.'
        : `Game Over! ${state.winnerIdx === 0 ? 'You' : `CPU ${state.winnerIdx}`} win!`,
    );
  lines.push(formatSeparator());
  return lines.join('\n');
}
