import type { koikoiApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type KoiKoiArgs = Parameters<typeof koikoiApi.exec>;

const VALID_COMMANDS = [
  'p',
  'play',
  'kk',
  'koikoi',
  'sb',
  'stop',
  'shobu',
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

/** Parse a Koi-Koi (こいこい) CLI command into API exec arguments. */
export function parseKoiKoiCommand(input: string): CliParseResult<KoiKoiArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <handIdx> [fieldIdx]' };
      if (args.length > 1) {
        const field = Number.parseInt(args[1], 10);
        if (Number.isNaN(field)) return { error: `Invalid field index: ${args[1]}` };
        return { args: ['play', { cardIndex: parsed.value, fieldIndex: field }] };
      }
      return { args: ['play', { cardIndex: parsed.value }] };
    }
    case 'kk':
    case 'koikoi':
      return { args: ['koikoi'] };
    case 'sb':
    case 'stop':
    case 'shobu':
      return { args: ['stop'] };
    case 'n':
    case 'next':
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

/** Help text for Koi-Koi (こいこい) CLI mode. */
export const KOIKOI_HELP: string[] = [
  'p <h> [f]    - Play hand card h, capturing field card f (omit f unless a 2-way match)',
  'kk/koikoi    - Call koi-koi (continue for more yaku, doubling the stakes)',
  'sb/stop      - Call shobu (stop and score the completed yaku)',
  'nr/nextround - Deal the next round',
  'h/hint       - Show hint',
  'r/reset      - Reset game',
];
