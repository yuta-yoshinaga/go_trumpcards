import type { sevenBridgeApi } from '../../../api/gameApi';
import { parseIntArg, parseIntSlice, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type SevenBridgeArgs = Parameters<typeof sevenBridgeApi.exec>;

const VALID_COMMANDS = [
  'd',
  'draw',
  'p',
  'pon',
  'c',
  'chi',
  'm',
  'meld',
  'lay',
  'layoff',
  'x',
  'discard',
  'n',
  'next',
  'nextround',
  'r',
  'reset',
  'log',
  'l',
  'h',
  'hint',
];

/** Parse a CLI input string into Seven Bridge API arguments. */
export function parseSevenBridgeCommand(input: string): CliParseResult<SevenBridgeArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'd':
    case 'draw':
      return { args: ['drawstock'] };
    case 'p':
    case 'pon': {
      const idx = parseIntSlice(args);
      if ('error' in idx || idx.values.length === 0) {
        return { error: 'Usage: p <i...> (hand indices forming the pon)' };
      }
      return { args: ['pon', undefined, undefined, idx.values] };
    }
    case 'c':
    case 'chi': {
      const idx = parseIntSlice(args);
      if ('error' in idx || idx.values.length === 0) {
        return { error: 'Usage: c <i...> (hand indices forming the chi)' };
      }
      return { args: ['chi', undefined, undefined, idx.values] };
    }
    case 'm':
    case 'meld': {
      const idx = parseIntSlice(args);
      if ('error' in idx || idx.values.length === 0) {
        return { error: 'Usage: m <i...> (hand indices forming the meld)' };
      }
      return { args: ['meld', undefined, undefined, idx.values] };
    }
    case 'x':
    case 'discard': {
      const card = parseIntArg(args, 0);
      if ('error' in card) return { error: 'Usage: x <i> (hand index to discard)' };
      return { args: ['discard', card.value] };
    }
    case 'lay':
    case 'layoff': {
      const card = parseIntArg(args, 0);
      const target = parseIntArg(args, 1);
      const meld = parseIntArg(args, 2);
      if ('error' in card || 'error' in target || 'error' in meld) {
        return { error: 'Usage: lay <handIdx> <targetPlayer> <meldIdx>' };
      }
      return { args: ['layoff', card.value, undefined, undefined, target.value, meld.value] };
    }
    case 'n':
    case 'next':
    case 'nextround':
      return { args: ['nextround'] };
    case 'log':
    case 'l':
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

/** Help text for Seven Bridge CLI mode. */
export const SEVENBRIDGE_HELP: string[] = [
  'd                       - Draw a card from the stock',
  'p <i...>                - Pon (claim discard with two matching hand cards)',
  'c <i...>                - Chi (claim discard to complete a run)',
  'm <i...>                - Meld a set/run from your hand',
  'lay <i> <pl> <meld>     - Lay off hand card i onto player pl meld',
  'x <i>                   - Discard hand card i',
  'n                       - Next round',
  'r                       - Reset / new game',
  'l / log                 - Action log',
  'h/hint       - Get a hint',
];
