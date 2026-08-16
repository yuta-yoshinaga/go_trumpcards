import type { durakApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type DurakArgs = Parameters<typeof durakApi.exec>;

const DURAK_COMMANDS = [
  'a',
  'attack',
  'd',
  'defend',
  'p',
  'pass',
  't',
  'take',
  'sort',
  'r',
  'reset',
  'h',
  'hint',
  'help',
  '?',
];

/** Map a sort argument (suit/value/s/v/0/1) to the numeric sort mode, or null if invalid. */
function parseSortMode(arg: string | undefined): number | null {
  switch (arg) {
    case 'suit':
    case 's':
    case '0':
      return 0;
    case 'value':
    case 'v':
    case '1':
      return 1;
    default:
      return null;
  }
}

/** Parse a Durak CLI command into API exec arguments (indices are 0-based). */
export function parseDurakCommand(input: string): CliParseResult<DurakArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'a':
    case 'attack': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: a <cardIdx>' };
      return { args: ['attack', parsed.value] as DurakArgs };
    }
    case 'd':
    case 'defend': {
      const card = parseIntArg(args, 0);
      const pair = parseIntArg(args, 1);
      if ('error' in card || 'error' in pair) return { error: 'Usage: d <cardIdx> <pairIdx>' };
      return { args: ['defend', card.value, pair.value] as DurakArgs };
    }
    case 'p':
    case 'pass':
      return { args: ['pass'] as DurakArgs };
    case 't':
    case 'take':
      return { args: ['take'] as DurakArgs };
    case 'sort': {
      const mode = parseSortMode(args[0]);
      if (mode === null) return { error: 'Usage: sort <suit|value>' };
      return { args: ['sort', undefined, undefined, undefined, mode] as DurakArgs };
    }
    case 'r':
    case 'reset':
      return { args: ['reset'] as DurakArgs };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    default: {
      const suggestion = suggestCommand(cmd, DURAK_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Durak CLI mode. */
export const DURAK_HELP: string[] = [
  'a <cardIdx>          - Attack with a hand card',
  'd <cardIdx> <pairIdx>- Defend a table pair with a hand card',
  'p/pass               - Pass (stop attacking)',
  't/take               - Take the table cards (give up defending)',
  'sort <suit|value>    - Sort your hand',
  'r/reset              - Reset game',
  'h/hint      - Get a hint',
];
