import type { macauApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import { STANDARD_SUIT_MAP as SUIT_MAP } from '../suitMaps';
import type { CliParseResult } from '../types';

type MacauArgs = Parameters<typeof macauApi.exec>;

const VALID_COMMANDS = [
  'p',
  'play',
  'd',
  'draw',
  'suit',
  'dc',
  'declare',
  'sk',
  'skipdeclare',
  'nr',
  'nextround',
  'r',
  'reset',
  'h',
  'hint',
  'help',
  '?',
];

/** Parse a Macau CLI command into API exec arguments. */
export function parseMacauCommand(input: string): CliParseResult<MacauArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <idx>' };
      return { args: ['play', parsed.value] };
    }
    case 'd':
    case 'draw':
      return { args: ['draw'] };
    case 'suit': {
      if (args.length === 0) return { error: 'Usage: suit <suit> (spade/clover/heart/diamond)' };
      const suit = SUIT_MAP[args[0].toLowerCase()];
      if (suit === undefined) return { error: 'Invalid suit. Use: spade/clover/heart/diamond' };
      return { args: ['suit', undefined, suit] };
    }
    case 'dc':
    case 'declare':
      return { args: ['declare'] };
    case 'sk':
    case 'skipdeclare':
      return { args: ['skipdeclare'] };
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

/** Help text for Macau CLI mode. */
export const MACAU_HELP: string[] = [
  'p <idx>     - Play a card (2=draw two, 7=skip, 8=wild, J=reverse)',
  'd/draw      - Draw / take the penalty stack',
  'suit <suit> - Choose suit (after 8)',
  'dc/declare  - Declare "Macau!" (one card left)',
  'sk          - Skip declaration (take penalty)',
  'nr/nextround- Next round',
  'r/reset     - Reset game',
  'h/hint      - Get a hint',
];
