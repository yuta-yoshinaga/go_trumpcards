import type { sheepsheadApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type SheepsheadArgs = Parameters<typeof sheepsheadApi.exec>;

const VALID_COMMANDS = [
  'pick',
  'pass',
  'bury',
  'call',
  'p',
  'play',
  'n',
  'next',
  'nr',
  'nextround',
  'h',
  'hint',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Sheepshead CLI command into API exec arguments. */
export function parseSheepsheadCommand(input: string): CliParseResult<SheepsheadArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'pick':
      return { args: ['pick', { pick: true }] };
    case 'pass':
      return { args: ['pick', { pick: false }] };
    case 'bury': {
      const indices = args.map((a) => Number(a)).filter((n) => Number.isInteger(n) && n >= 0);
      if (indices.length !== 2) return { error: 'Usage: bury <idx> <idx>' };
      return { args: ['bury', { buryIndices: indices }] };
    }
    case 'call': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: call <suit> (1=spade 2=club 3=heart)' };
      return { args: ['call', { callSuit: parsed.value }] };
    }
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <idx>' };
      return { args: ['play', { cardIndex: parsed.value }] };
    }
    case 'n':
    case 'next':
      return { args: ['next'] };
    case 'nr':
    case 'nextround':
      return { args: ['nextround'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
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

/** Help text for Sheepshead CLI mode. */
export const SHEEPSHEAD_HELP: string[] = [
  'pick             - Take the blind (Pick phase)',
  'pass             - Pass on the blind (Pick phase)',
  'bury <i> <j>     - Bury 2 cards (picker, Bury phase)',
  'call <suit>      - Call partner suit 1=spade 2=club 3=heart (Call phase)',
  'p <idx>          - Play a card (Play phase)',
  'n/next           - Next trick',
  'nr/nextround     - Next round',
  'h/hint           - Show hint',
  'r/reset          - Reset game',
];
