import type { DehlaPakadResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

/** Format a Dehla Pakad game state as terminal text. */
export function formatDehlaPakadState(state: DehlaPakadResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Dehla Pakad'));
  lines.push(`hand: ${state.handNumber}  phase: ${state.phase}  target: ${state.config.targetKots} kot(s)`);
  if (state.trumpSuit > 0) {
    lines.push(`trump: ${state.trumpSuitName}  trick: ${state.trickNumber}/${state.trickCount}`);
  }
  const tens = state.teamTens ?? [];
  const kots = state.teamKots ?? [];
  const you = state.humanTeam;
  lines.push(
    `tens: yours=${tens[you] ?? 0} theirs=${tens[1 - you] ?? 0}  kots: ${kots[you] ?? 0}/${kots[1 - you] ?? 0}`,
  );
  if (state.streakCount > 1 && state.streakTeam >= 0) {
    lines.push(`streak: team ${state.streakTeam} has won ${state.streakCount} in a row`);
  }
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const role = p.isDealer ? 'Dealer' : 'Player';
    lines.push(`${name} (team ${p.team} / ${role}): cards=${p.cardCount} gathered=${p.gatheredCount}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.currentTrick.length > 0) {
    const parts = state.currentTrick.map((tc) => {
      const name = formatPlayerName(tc.playerIdx, state.players[tc.playerIdx]?.isHuman ?? false);
      return `${name}=${formatCard(tc.card)}`;
    });
    lines.push(`trick: ${parts.join(', ')}`);
  }

  // **これがこのゲームの心臓部。** 取っただけでは札は手に入らない。
  if (state.centrePileCount > 0) {
    lines.push(`centre: ${state.centrePileCount} card(s), ${state.centrePileTens} ten(s)`);
    if (state.prevTrickWinner >= 0) {
      const name = formatPlayerName(state.prevTrickWinner, state.prevTrickWinner === 0);
      lines.push(`  ${name} takes the pile by winning the next trick too`);
    }
  }

  if (state.isTrumpPhase) {
    const chooser = formatPlayerName(state.trumpChooserIdx, state.trumpChooserIdx === 0);
    lines.push(`${chooser} calls the trump (1=spade 2=club 3=heart 4=diamond)`);
  }

  if (state.lastHand) {
    const h = state.lastHand;
    lines.push(
      `last hand: team ${h.winnerTeam} wins, tens ${h.teamTens.join('-')}${h.kot ? ` (KOT: ${h.kotReason})` : ''}`,
    );
  }

  if (state.hint && isRequestedHint(state)) {
    const indices = state.hint.cardIndices ?? [];
    lines.push(`HINT: card indices [${indices.join(', ')}] (${state.hint.reason})`);
  }
  if (state.hintTrumpSuit > 0 && isRequestedHint(state)) {
    lines.push(`HINT: call trump ${state.hintTrumpSuit}`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    lines.push(`Game Over! Team ${state.winnerTeam} takes the match!`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
