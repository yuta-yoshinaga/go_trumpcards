import type { rookApi } from '../../../api/gameApi';
import type { RookResponse } from '../../../types/card';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by rookApi.exec. */
export type RookCliArgs = Parameters<typeof rookApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bid',
  'pa',
  'pass',
  'e',
  'exchange',
  'p',
  'play',
  'n',
  'next',
  'nr',
  'nextround',
  'r',
  'reset',
  'h',
  'hint',
  'help',
  '?',
];

/** Parses a single CLI command line for the Rook (ルーク) game. */
export function parseRookCommand(input: string): CliParseResult<RookCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      const bid = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(bid)) return { error: 'Usage: b <points> (70-120, step 5)' };
      return { args: ['bid', { bid }] };
    }
    case 'pa':
    case 'pass':
      return { args: ['pass'] };
    case 'e':
    case 'exchange': {
      const nums = args.map((a) => Number.parseInt(a, 10));
      if (nums.length !== 6 || nums.some(Number.isNaN))
        return { error: 'Usage: e <i> <j> <k> <l> <m> <color> (discard 5, color 1=red 2=gold 3=green 4=black)' };
      const trumpColor = nums[5];
      const discardIndices = nums.slice(0, 5);
      return { args: ['exchange', { discardIndices, trumpColor }] };
    }
    case 'p':
    case 'play': {
      const idx = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(idx)) return { error: 'Usage: p <idx>' };
      return { args: ['play', { cardIndex: idx }] };
    }
    case 'n':
    case 'next':
      return { args: ['next'] };
    case 'nr':
    case 'nextround':
      return { args: ['nextround'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Color labels for a Rook trump-color id (1=red, 2=gold, 3=green, 4=black). */
const COLOR_NAMES: Record<number, string> = { 1: 'Red', 2: 'Gold', 3: 'Green', 4: 'Black' };

/** Renders a Rook game state as CLI-friendly text. */
export function formatRookState(s: RookResponse): string {
  const lines: string[] = [];
  lines.push(`Round ${s.roundNumber} · Trick ${s.trickNumber}`);
  lines.push(`Scores: Team0 ${s.teamScores[0]} / Team1 ${s.teamScores[1]}`);
  if (s.contractBid > 0) {
    const trump = s.trumpColor >= 1 ? (COLOR_NAMES[s.trumpColor] ?? '?') : 'undeclared';
    lines.push(`Contract: ${s.contractBid} (trump ${trump})`);
  } else if (s.highestBid > 0) {
    lines.push(`Highest bid: ${s.highestBid}`);
  }
  for (const p of s.players) {
    const tag = p.isHuman ? 'You' : `CPU${p.id}`;
    const role = p.isDeclarer ? ' [declarer]' : p.passed ? ' [pass]' : '';
    lines.push(`${tag} (T${p.team}): ${p.cardCount} cards, ${p.trickCount} tricks, ${p.points} pts${role}`);
  }
  if (s.currentTrick.length > 0) {
    lines.push(`Trick: ${s.currentTrick.map((tc) => tc.card.label ?? String(tc.card.value)).join(' ')}`);
  }
  if (s.message) lines.push(s.message);
  return lines.join('\n');
}

/** Help text shown in the CLI terminal for Rook (ルーク). */
export const ROOK_HELP = [
  'b <points>              - Bid points (70-120, step 5)',
  'pa                      - Pass',
  'e <i j k l m> <color>   - Exchange nest (discard 5, color 1=red 2=gold 3=green 4=black)',
  'p <idx>                 - Play a card',
  'n / nr                  - Next trick / next round',
  'r/reset                 - Reset game',
  'h/hint                  - Get a hint',
] as const;
