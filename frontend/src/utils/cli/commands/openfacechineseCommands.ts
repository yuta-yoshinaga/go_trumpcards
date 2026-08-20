import type { openfacechineseApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type OpenFaceChineseArgs = Parameters<typeof openfacechineseApi.exec>;

const VALID_COMMANDS = ['p', 'place', 'n', 'next', 'nextround', 'log', 'r', 'reset', 'h', 'hint', 'help', '?'];

/** Maps a row token (name / initial / index) to the backend row index, or null if invalid. */
function parseRow(token: string | undefined): number | null {
  switch ((token ?? '').toLowerCase()) {
    case 'front':
    case 'f':
    case '0':
      return 0;
    case 'middle':
    case 'mid':
    case 'm':
    case '1':
      return 1;
    case 'back':
    case 'b':
    case '2':
      return 2;
    default:
      return null;
  }
}

/** Parse an Open Face Chinese Poker CLI command into API exec arguments. */
export function parseOpenfacechineseCommand(input: string): CliParseResult<OpenFaceChineseArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'place': {
      const row = parseRow(args[0]);
      if (row === null) return { error: 'Usage: place <front|middle|back>' };
      return { args: ['place', { row }] };
    }
    case 'n':
    case 'next':
    case 'nextround':
      return { args: ['nextround'] };
    case 'log':
      return { args: ['log'] };
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

/** Help text for Open Face Chinese Poker CLI mode. */
export const OPENFACECHINESE_HELP: string[] = [
  'p/place <f|m|b> - Place the pending card in the front / middle / back row',
  'n/next          - Deal the next round',
  'log             - Show action log',
  'r/reset         - Reset game',
  'h/hint      - Get a hint',
];
