import type { cribbageApi } from '../../../api/gameApi';
import { parseIntArg, parseIntSlice, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type CribbageArgs = Parameters<typeof cribbageApi.exec>;

const VALID_COMMANDS = [
  'dis',
  'discard',
  'c',
  'cut',
  'peg',
  'go',
  'sn',
  'shownext',
  'nr',
  'nextround',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Cribbage CLI command into API exec arguments. */
export function parseCribbageCommand(input: string): CliParseResult<CribbageArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'dis':
    case 'discard': {
      const parsed = parseIntSlice(args);
      if ('error' in parsed) return { error: 'Usage: dis <idx...> (2 cards to crib)' };
      return { args: ['discard', undefined, parsed.values] };
    }
    case 'c':
    case 'cut':
      return { args: ['cut'] };
    case 'peg': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: peg <idx>' };
      return { args: ['peg', parsed.value] };
    }
    case 'go':
      return { args: ['go'] };
    case 'sn':
    case 'shownext':
      return { args: ['shownext'] };
    case 'nr':
    case 'nextround':
      return { args: ['nextround'] };
    case 'log':
      return { args: ['log'] };
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

/** Help text for Cribbage CLI mode. */
export const CRIBBAGE_HELP: string[] = [
  'dis <idx...>   - Discard to crib (2 cards)',
  'c/cut          - Cut the deck to reveal the starter',
  'peg <idx>      - Play card in pegging',
  'go             - Say "Go"',
  'sn/shownext    - Show next scoring',
  'nr/nextround   - Next round',
  'log            - Show action log',
  'r/reset        - Reset game',
];
