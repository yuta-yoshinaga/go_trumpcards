import type { DoppelkopfResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const PHASE_NAMES = ['Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];

/** Format a Doppelkopf game state as terminal text. */
export function formatDoppelkopfState(state: DoppelkopfResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Doppelkopf'));
  lines.push(
    `round: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(`you are: ${state.youAreRe ? 'Re' : 'Kontra'}${state.soloRe ? ' (solo Re)' : ''}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const team = state.teamsRevealed ? (p.isRe ? ' [Re]' : ' [Kontra]') : '';
    lines.push(`${name}${team}: cards=${p.cardCount} tricks=${p.trickCount} chips=${p.chips}`);
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

  if (state.reAnnounced || state.kontraAnnounced) {
    const calls: string[] = [];
    if (state.reAnnounced) calls.push('Re');
    if (state.kontraAnnounced) calls.push('Kontra');
    lines.push(`announced: ${calls.join(', ')}`);
  }

  if (state.phase === 2 || state.phase === 3) {
    lines.push(
      `round result: Re points=${state.roundRePoints} ` +
        `reWon=${state.roundReWon ? 'yes' : 'no'} gamePoints=${state.roundGamePoints}`,
    );
  }

  if (state.hint && isRequestedHint(state)) {
    const indices = state.hint.cardIndices ?? [];
    lines.push(`HINT: card indices [${indices.join(', ')}] (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.winnerIdx >= 0) {
    lines.push(
      `Game Over! Winner: ${formatPlayerName(state.winnerIdx, state.players[state.winnerIdx]?.isHuman ?? false)}`,
    );
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
