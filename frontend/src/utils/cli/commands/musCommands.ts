import type { musApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type MusArgs = Parameters<typeof musApi.exec>;

const VALID_COMMANDS = [
  'm',
  'mus',
  'c',
  'cut',
  'corte',
  'd',
  'discard',
  'paso',
  'e',
  'envido',
  'ordago',
  'quiero',
  'nq',
  'noquiero',
  'n',
  'next',
  'h',
  'hint',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Mus CLI command into API exec arguments. */
export function parseMusCommand(input: string): CliParseResult<MusArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'm':
    case 'mus':
      return { args: ['mus', { mus: true }] };
    case 'c':
    case 'cut':
    case 'corte':
      return { args: ['mus', { mus: false }] };
    case 'd':
    case 'discard': {
      const indices = args.map((a) => Number(a)).filter((nn) => Number.isInteger(nn) && nn >= 0);
      return { args: ['discard', { discardIndices: indices }] };
    }
    case 'paso':
      return { args: ['bet', { betAction: 0, betAmount: 0 }] };
    case 'e':
    case 'envido': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: e <amount>' };
      return { args: ['bet', { betAction: 1, betAmount: parsed.value }] };
    }
    case 'ordago':
      return { args: ['bet', { betAction: 2, betAmount: 0 }] };
    case 'quiero':
      return { args: ['bet', { betAction: 3, betAmount: 0 }] };
    case 'nq':
    case 'noquiero':
      return { args: ['bet', { betAction: 4, betAmount: 0 }] };
    case 'n':
    case 'next':
      return { args: ['next'] };
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

/** Help text for Mus CLI mode. */
export const MUS_HELP: string[] = [
  'm/mus            - Call Mus / exchange (Mus phase)',
  'c/cut/corte      - Cut and start betting (Mus phase)',
  'd <i>...         - Discard cards by index (Discard phase)',
  'paso             - Pass (no bet)',
  'e/envido <n>     - Bet n amarrakos',
  'ordago           - Órdago (all-in)',
  'quiero           - Accept the bet',
  'nq/noquiero      - Decline the bet',
  'n/next           - Next round',
  'h/hint           - Show hint',
  'r/reset          - Reset game',
];
