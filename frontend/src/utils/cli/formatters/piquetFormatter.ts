import type { PiquetResponse } from '../../../types/card';
import { PiquetDeclarationKind, PiquetExchangeTurn, PiquetPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [PiquetPhase.EXCHANGE]: 'Exchange',
  [PiquetPhase.DECLARATION]: 'Declaration',
  [PiquetPhase.PLAY]: 'Play',
  [PiquetPhase.SCORE]: 'Round end',
  [PiquetPhase.GAME_END]: 'Partie end',
};

const KIND_NAMES: Record<number, string> = {
  [PiquetDeclarationKind.POINT]: 'Point',
  [PiquetDeclarationKind.SEQUENCE]: 'Sequence',
  [PiquetDeclarationKind.SET]: 'Set',
};

/** Format a Piquet game state as terminal text. */
export function formatPiquetState(state: PiquetResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Piquet'));
  lines.push(`deal: ${state.dealNumber}/${state.dealsPerPartie}  phase: ${PHASE_NAMES[state.phase] ?? '?'}`);

  for (let i = 0; i < state.players.length; i++) {
    const p = state.players[i];
    const role = i === state.elderIdx ? 'Elder' : 'Younger';
    const name = `${role} ${formatPlayerName(p.id, p.isHuman)}`;
    const parts = [
      `tricks=${p.trickCount}`,
      `decl=${p.declScore}`,
      `trick=${p.trickScore}`,
      `bonus=${p.bonusScore}`,
      `round=${p.roundScore}`,
      `match=${p.matchScore}`,
    ];
    lines.push(`${name}: ${parts.join(' ')}`);
    if (state.carteBlanche[i]) lines.push(`  ★ carte blanche +10`);
    if (p.isHuman && p.cards.length > 0) lines.push(`  ${formatIndexedCards(p.cards)}`);
  }
  lines.push(formatSeparator());

  if (state.phase === PiquetPhase.EXCHANGE) {
    if (state.exchangeTurn === PiquetExchangeTurn.ELDER) lines.push('Waiting: Elder to exchange (1..5)');
    if (state.exchangeTurn === PiquetExchangeTurn.YOUNGER) lines.push('Waiting: Younger to exchange (0..3)');
  }

  if (state.declResults.length > 0) {
    lines.push('Declarations:');
    for (const r of state.declResults) {
      const kind = KIND_NAMES[r.kind] ?? '?';
      if (r.score === 0) {
        lines.push(`  ${kind}: tied`);
      } else {
        const who = r.scoredBy === state.elderIdx ? 'Elder' : 'Younger';
        lines.push(`  ${kind}: ${who} +${r.score}`);
      }
    }
  }

  if (state.phase === PiquetPhase.PLAY && state.currentTrick.length > 0) {
    const parts = state.currentTrick.map((tc) => `P${tc.playerIdx}=${formatCard(tc.card)}`);
    lines.push(`Trick: ${parts.join(' ')}`);
  }

  if (state.phase === PiquetPhase.GAME_END) {
    if (state.winnerIdx === -1) {
      lines.push('Partie ended in a draw');
    } else {
      lines.push(`★ Partie winner: player ${state.winnerIdx}`);
    }
  }

  if (state.message) lines.push(`message: ${state.message}`);
  return lines.join('\n');
}
