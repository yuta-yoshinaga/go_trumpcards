import type { Card } from '../../../types/card';
import {
  formatCard,
  formatCardList,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
} from '../formatterBase';

/** Generic player-like structure. */
export interface GenericPlayer {
  id: number;
  isHuman: boolean;
  cards?: Card[];
  cardCount?: number;
  chips?: number;
  score?: number;
  cumulativeScore?: number;
  roundScore?: number;
  trickCount?: number;
  bid?: number;
  folded?: boolean;
  allIn?: boolean;
  isFinished?: boolean;
  handName?: string;
}

/** Generic trick card. */
export interface GenericTrickCard {
  playerIdx: number;
  card: Card;
}

/** Options for the generic formatter. */
export interface GenericFormatOptions {
  title: string;
  players?: GenericPlayer[];
  phase?: number | string;
  phaseNames?: Record<number, string>;
  currentTurn?: number;
  pot?: number;
  communityCards?: Card[];
  tableCards?: Card[];
  currentTrick?: GenericTrickCard[];
  roundNumber?: number;
  trickNumber?: number;
  message?: string;
  gameEndFlag?: boolean;
  customLines?: string[];
  /** Score display mode: 'chips' | 'score' | 'trick' */
  scoreMode?: 'chips' | 'score' | 'trick';
}

/** Format a generic game state as terminal text. */
export function formatGenericState(opts: GenericFormatOptions): string {
  const lines: string[] = [];

  lines.push(formatHeader(opts.title));

  // Phase
  if (opts.phase !== undefined) {
    const phaseName =
      typeof opts.phase === 'string' ? opts.phase : (opts.phaseNames?.[opts.phase] ?? String(opts.phase));
    lines.push(`phase: ${phaseName}`);
  }
  if (opts.pot !== undefined) lines.push(`pot: ${opts.pot}`);
  if (opts.roundNumber !== undefined)
    lines.push(`round: ${opts.roundNumber}${opts.trickNumber !== undefined ? `  trick: ${opts.trickNumber}` : ''}`);
  lines.push('');

  // Players
  if (opts.players) {
    for (const p of opts.players) {
      const name = formatPlayerName(p.id, p.isHuman);
      const parts: string[] = [name];
      if (opts.scoreMode === 'chips' && p.chips !== undefined) parts.push(`chips=${p.chips}`);
      if (opts.scoreMode === 'score' && p.cumulativeScore !== undefined) parts.push(`total=${p.cumulativeScore}`);
      if (opts.scoreMode === 'trick' && p.trickCount !== undefined) parts.push(`tricks=${p.trickCount}`);
      if (p.roundScore !== undefined) parts.push(`round=${p.roundScore}`);
      if (p.bid !== undefined && p.bid >= 0) parts.push(`bid=${p.bid}`);
      if (p.cardCount !== undefined) parts.push(`${p.cardCount} cards`);
      if (p.folded) parts.push('[Folded]');
      if (p.allIn) parts.push('[All-in]');
      if (p.isFinished) parts.push('[Finished]');
      if (p.handName) parts.push(`[${p.handName}]`);
      lines.push(parts.join(' '));
      if (p.isHuman && p.cards && p.cards.length > 0) {
        lines.push(`  ${formatIndexedCards(p.cards)}`);
      }
      if (opts.gameEndFlag && !p.isHuman && p.cards && p.cards.length > 0) {
        lines.push(`  ${p.cards.map(formatCard).join('  ')}`);
      }
    }
    lines.push('----------');
  }

  // Community cards
  if (opts.communityCards && opts.communityCards.length > 0) {
    lines.push(`board: ${formatCardList(opts.communityCards)}`);
  }

  // Table cards
  if (opts.tableCards && opts.tableCards.length > 0) {
    lines.push(`table: ${formatCardList(opts.tableCards)}`);
  }

  // Current trick
  if (opts.currentTrick && opts.currentTrick.length > 0) {
    const trickParts = opts.currentTrick.map((tc) => {
      const name = formatPlayerName(tc.playerIdx, opts.players?.[tc.playerIdx]?.isHuman ?? false);
      return `${name}=${formatCard(tc.card)}`;
    });
    lines.push(`trick: ${trickParts.join(', ')}`);
  }

  // Custom lines
  if (opts.customLines) {
    for (const line of opts.customLines) {
      lines.push(line);
    }
  }

  // Turn info
  if (opts.currentTurn !== undefined && !opts.gameEndFlag && opts.players) {
    const current = formatPlayerName(opts.currentTurn, opts.players[opts.currentTurn]?.isHuman ?? false);
    lines.push(`turn: ${current}`);
  }

  if (opts.message) lines.push(opts.message);
  if (opts.gameEndFlag) lines.push('Game Over');

  lines.push(formatSeparator());
  return lines.join('\n');
}
