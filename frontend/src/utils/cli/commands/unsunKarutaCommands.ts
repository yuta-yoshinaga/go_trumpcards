import type { unsunKarutaApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type UnsunKarutaArgs = Parameters<typeof unsunKarutaApi.exec>;

const VALID_COMMANDS = [
  'p',
  'play',
  'meri',
  'monchi',
  'm',
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

/**
 * Parse an Unsun Karuta CLI command into API exec arguments.
 *
 * **The declaration rides with the card.** `meri 0` plays index 0 *and*
 * declares; there is no way to declare without playing, because that state does
 * not exist on the board.
 */
export function parseUnsunKarutaCommand(input: string): CliParseResult<UnsunKarutaArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <idx>' };
      return { args: ['play', { cardIndex: parsed.value, declare: false }] };
    }
    case 'm':
    case 'meri':
    case 'monchi': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: meri <idx>' };
      return { args: ['play', { cardIndex: parsed.value, declare: true }] };
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

/** Help text for Unsun Karuta CLI mode. */
export const UNSUN_KARUTA_HELP: string[] = [
  'p <idx>                          - Play a card',
  'meri <idx>                       - Play it and declare (leader only; everyone must then follow)',
  'n/next                           - Next trick',
  'nr/nextround                     - Next deal',
  'h/hint                           - Show hint',
  'r/reset                          - Reset game',
];
