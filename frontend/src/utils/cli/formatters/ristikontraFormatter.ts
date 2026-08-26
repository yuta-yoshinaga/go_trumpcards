import type { RistikontraResponse } from '../../../types/card';
import { formatCard, formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<string, string> = {
  play: 'Play',
  roundEnd: 'RoundEnd',
  gameEnd: 'GameEnd',
};

/** Format a Ristikontra game state as terminal text. */
export function formatRistikontraState(state: RistikontraResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Ristikontra'));
  lines.push(`phase: ${PHASE_NAMES[state.phase] ?? state.phase}  stock: ${state.remainingDeck}`);

  const top = state.pileTop ? formatCard(state.pileTop) : '-';
  lines.push(`pile: top=${top} (${state.pileCount} cards)`);
  lines.push('----------');

  state.players.forEach((p, i) => {
    const name = formatPlayerName(i, p.isHuman);
    const turn = i === state.currentTurn && !state.gameEndFlag ? ' <- turn' : '';
    lines.push(`${name}: hand ${p.cardCount} / captured ${p.capturedCount} / team ${(p.id % 2) + 1}${turn}`);
  });
  lines.push('----------');

  const human = state.players.find((p) => p.isHuman);
  if (human && human.cards.length > 0) {
    lines.push(`your hand: ${formatIndexedCards(human.cards)}`);
  }

  if (state.phase === 'play' && state.currentTurn === 0 && !state.gameEndFlag) {
    lines.push('(your turn — play a card with "p <hand#>")');
  }

  if (state.message) lines.push(state.message);

  if (state.gameEndFlag) {
    lines.push('Game Over!');
    state.players.forEach((p, i) => {
      const score = state.finalScores[i] ?? p.finalScore;
      lines.push(`  ${formatPlayerName(i, p.isHuman)}: ${score} pts`);
    });
    if (state.winners.length > 0) {
      const names = state.winners.map((idx) => formatPlayerName(idx, idx === 0)).join(', ');
      lines.push(`Winner: ${names}`);
    }
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
