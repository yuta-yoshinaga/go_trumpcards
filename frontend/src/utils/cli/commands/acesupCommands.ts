import type { acesupApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type AcesUpArgs = Parameters<typeof acesupApi.exec>;

const VALID_COMMANDS = [
  'd',
  'draw',
  'rm',
  'remove',
  'mv',
  'move',
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

/** Parse an Aces Up CLI command into API exec arguments. */
export function parseAcesUpCommand(input: string): CliParseResult<AcesUpArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'd':
    case 'draw':
      return { args: ['draw'] };
    case 'rm':
    case 'remove': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: rm <col>' };
      return { args: ['remove', parsed.value] };
    }
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

/** Help text for Aces Up CLI mode. */
export const ACESUP_HELP: string[] = [
  'd/draw      - Deal a card to each column',
  'rm <col>    - Remove top card from column',
  'mv <col>    - Move top card to an empty column',
  'u/undo      - Undo last move',
  'h/hint      - Get a hint',
  'g/giveup    - Give up',
  'log         - Show action log',
  'r/reset     - Reset game',
];
