import type { maoApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import { STANDARD_SUIT_MAP as SUIT_MAP } from '../suitMaps';
import type { CliParseResult } from '../types';

type MaoArgs = Parameters<typeof maoApi.exec>;

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
  'dw',
  'declareword',
  'nr',
  'nextround',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Mao CLI command into API exec arguments. */
export function parseMaoCommand(input: string): CliParseResult<MaoArgs> {
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
    case 'dw':
    case 'declareword': {
      const word = args.join(' ').trim();
      if (word.length === 0) return { error: 'Usage: dw <word>' };
      return { args: ['declareword', undefined, undefined, undefined, word] };
    }
    case 'nr':
    case 'nextround':
      return { args: ['nextround'] };
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

/** Help text for Mao CLI mode. */
export const MAO_HELP: string[] = [
  'p <idx>     - Play a card (2=draw two, A=skip, 8=wild)',
  'd/draw      - Draw / take the penalty stack',
  'suit <suit> - Choose suit (after 8)',
  'dc/declare  - Declare "Mao!" (one card left)',
  'sk          - Skip declaration (take penalty)',
  'dw <word>   - Say a word (the hidden rule may require it — you must guess when!)',
  'nr/nextround- Next round',
  'r/reset     - Reset game',
];
