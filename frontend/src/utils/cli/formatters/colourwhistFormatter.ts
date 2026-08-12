import type { ColourWhistResponse } from '../../../types/card';
import { COLOUR_WHIST_NO_TRUMP } from '../../../types/games/colourwhist';
import { ColourWhistContract, ColourWhistPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [ColourWhistPhase.BID]: 'BID',
  [ColourWhistPhase.CALL]: 'CALL',
  [ColourWhistPhase.PLAY]: 'PLAY',
  [ColourWhistPhase.ROUND_END]: 'ROUND END',
  [ColourWhistPhase.GAME_END]: 'GAME END',
};

const CONTRACT_NAMES: Record<number, string> = {
  [ColourWhistContract.NONE]: 'undecided',
  [ColourWhistContract.SAMEN]: 'Samen',
  [ColourWhistContract.ALLEEN]: 'Alleen',
  [ColourWhistContract.MISERIE]: 'Miserie',
  [ColourWhistContract.TROEL]: 'Troel',
};

const SUIT_NAMES: Record<number, string> = { 1: 'Spades', 2: 'Clubs', 3: 'Hearts', 4: 'Diamonds' };

/** Renders the trump, naming no-trump rather than printing -1. */
function trumpName(suit: number): string {
  return suit === COLOUR_WHIST_NO_TRUMP ? 'none' : (SUIT_NAMES[suit] ?? '?');
}

/** Format a Colour Whist game state as terminal text. */
export function formatColourWhistState(state: ColourWhistResponse): string {
  const lines: string[] = [];
  lines.push(formatHeader('Colour Whist'));
  lines.push(`phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(`round: ${state.roundNumber}${state.config ? ` / ${state.config.rounds}` : ''}`);
  lines.push(`contract: ${CONTRACT_NAMES[state.contract] ?? '?'} (trump: ${trumpName(state.trumpSuit)})`);
  // **競りが飛ばされた理由を出す。** 出さないと不具合に見えます。
  if (state.troelForced) {
    lines.push(`troel: seat ${state.declarerIdx} held three aces, so there was no auction`);
  }
  if (state.declarerIdx >= 0) {
    lines.push(`declarer: seat ${state.declarerIdx} (${state.declarerTricks} tricks)`);
  }
  // **得点は負にもなります。** ゼロサムなので当然そうなります。
  lines.push(`scores: ${state.players.map((p) => `#${p.id}:${p.score}`).join(' ')}`);

  if (state.currentTrick.length > 0) {
    lines.push(formatSeparator());
    for (const tc of state.currentTrick) {
      lines.push(`  seat ${tc.playerIdx}: ${formatCard(tc.card)}`);
    }
  }

  const human = state.players.find((p) => p.isHuman);
  if (human && human.cards.length > 0) {
    lines.push(formatSeparator());
    const legal = new Set(state.validPlays);
    lines.push(`hand: ${human.cards.map((c, i) => `[${i}${legal.has(i) ? '*' : ' '}]${formatCard(c)}`).join(' ')}`);
  }

  if (state.gameEndFlag) lines.push(`winner: seat ${state.winnerIdx}`);
  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
