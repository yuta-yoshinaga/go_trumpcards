import type { kalookiApi } from '../../../api/gameApi';
import { parseIntArg, parseIntSlice, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type KalookiArgs = Parameters<typeof kalookiApi.exec>;

const VALID_COMMANDS = [
  'ds',
  'drawstock',
  'dd',
  'drawdiscard',
  'meld',
  'm',
  'lo',
  'layoff',
  'dis',
  'discard',
  'd',
  'nr',
  'nextround',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Kalooki CLI command into API exec arguments. */
export function parseKalookiCommand(input: string): CliParseResult<KalookiArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'ds':
    case 'drawstock':
      return { args: ['drawstock'] };
    case 'dd':
    case 'drawdiscard':
      return { args: ['drawdiscard'] };
    case 'm':
    case 'meld': {
      // meld <idx...> — all indices form one meld group
      const parsed = parseIntSlice(args);
      if ('error' in parsed) return { error: 'Usage: meld <idx...>' };
      return { args: ['meld', { meldGroups: [parsed.values] }] };
    }
    case 'lo':
    case 'layoff': {
      // layoff <playerIdx> <meldIdx> <cardIdx>
      const usage = 'Usage: lo <playerIdx> <meldIdx> <cardIdx>';
      const targetPlayerIdx = parseIntArg(args, 0);
      if ('error' in targetPlayerIdx || targetPlayerIdx.value < 0) return { error: usage };
      const meldIdx = parseIntArg(args, 1);
      if ('error' in meldIdx || meldIdx.value < 0) return { error: usage };
      const cardIndex = parseIntArg(args, 2);
      if ('error' in cardIndex || cardIndex.value < 0) return { error: usage };
      return {
        args: [
          'layoff',
          { targetPlayerIdx: targetPlayerIdx.value, meldIdx: meldIdx.value, cardIndex: cardIndex.value },
        ],
      };
    }
    case 'd':
    case 'dis':
    case 'discard': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: d <idx>' };
      return { args: ['discard', { cardIndex: parsed.value }] };
    }
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

/** Help text for Kalooki CLI mode. */
export const KALOOKI_HELP: string[] = [
  'ds/drawstock          - Draw from stock',
  'dd/drawdiscard        - Draw from discard',
  'meld <idx...>         - Lay down a meld',
  'lo <pIdx> <mIdx> <cIdx> - Lay off a card onto a meld',
  'd <idx>               - Discard a card',
  'nr/nextround          - Next round',
  'log                   - Show action log',
  'r/reset               - Reset game',
];
