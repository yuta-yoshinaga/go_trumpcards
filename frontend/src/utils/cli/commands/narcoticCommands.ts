import type { narcoticApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type NarcoticArgs = Parameters<typeof narcoticApi.exec>;

const VALID_COMMANDS = [
  'd',
  'draw',
  'rm',
  'remove',
  'mv',
  'move',
  'rd',
  'redeal',
  'g',
  'giveup',
  'h',
  'hint',
  'log',
  'u',
  'undo',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Narcotic CLI command into API exec arguments. */
export function parseNarcoticCommand(input: string): CliParseResult<NarcoticArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'd':
    case 'draw':
      return { args: ['draw'] };
    // **列を取らない。**揃った4枚をまとめて捨てるので、選ぶ余地が無い。
    // クローン元 (Aces Up) は `rm <col>` で、列ごとに捨てる。
    case 'rm':
    case 'remove':
      return { args: ['remove'] };
    case 'rd':
    case 'redeal':
      return { args: ['redeal'] };
    case 'mv':
    case 'move': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: mv <col>' };
      return { args: ['move', parsed.value] };
    }
    case 'g':
    case 'giveup':
      return { args: ['giveup'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    case 'log':
      return { args: ['log'] };
    case 'u':
    case 'undo':
      return { args: ['undo'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Narcotic CLI mode. */
export const NARCOTIC_HELP: string[] = [
  'd/draw      - Deal one card to each of the four piles',
  'rm          - Discard all four exposed cards when their ranks match',
  'mv <col>    - Stack onto the leftmost pile showing the same rank',
  'rd/redeal   - Gather from the right and re-deal, unshuffled (no limit)',
  'u/undo      - Undo last move',
  'h/hint      - Get a hint',
  'g/giveup    - Give up',
  'log         - Show action log',
  'r/reset     - Reset game',
];
