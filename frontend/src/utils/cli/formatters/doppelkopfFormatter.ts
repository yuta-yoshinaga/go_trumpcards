import type { DoppelkopfResponse } from '../../../types/card';
import { formatSignedDelta } from '../../formatSignedDelta';
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
  if (state.roundScoreHistory.length > 0) {
    lines.push('ROUND SCORES:');
    state.roundScoreHistory.forEach((scores, roundIndex) => {
      lines.push(
        `Round ${roundIndex + 1}: ${scores.map((score, playerIdx) => `P${playerIdx + 1} ${formatSignedDelta(score)}`).join(' | ')}`,
      );
    });
  }
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

  // 進行中の獲得点。GUI のパネルと CUI の 1 行が出しているのに、CLI モードだけが
  // 何も出していなかった (#6435)。ラウンド結果は下の行が引き継ぐので、ここは進行中だけ。
  if (state.phase === 0 || state.phase === 1) {
    lines.push(`card points: Re=${state.liveRePoints} Kontra=${state.liveKontraPoints}`);
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
