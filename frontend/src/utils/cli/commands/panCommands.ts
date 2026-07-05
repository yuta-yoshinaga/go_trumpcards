import type { panApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type PanArgs = Parameters<typeof panApi.exec>;

const VALID_COMMANDS = [
  'ds',
  'drawstock',
  'dd',
  'drawdiscard',
  'm',
  'meld',
  'lo',
  'layoff',
  'd',
  'discard',
  'nr',
  'nextround',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Panguingue (Pan) CLI command into API exec arguments. */
export function parsePanCommand(input: string): CliParseResult<PanArgs> {
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
      const indices = args.map((a) => Number.parseInt(a, 10));
      if (indices.length < 3 || indices.some((n) => Number.isNaN(n))) {
        return { error: 'Usage: m <idx> <idx> <idx> [...]' };
      }
      return { args: ['meld', { cardIndices: indices }] };
    }
    case 'lo':
    case 'layoff': {
      const owner = parseIntArg(args, 0);
      const meldIdx = parseIntArg(args, 1);
      const cardIndex = parseIntArg(args, 2);
      if ('error' in owner || 'error' in meldIdx || 'error' in cardIndex) {
        return { error: 'Usage: lo <owner> <meld> <idx>' };
      }
      return { args: ['layoff', { meldOwner: owner.value, meldIdx: meldIdx.value, cardIndex: cardIndex.value }] };
    }
    case 'd':
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

/** Help text for Panguingue (Pan) CLI mode. */
export const PAN_HELP: string[] = [
  'ds/drawstock   - Draw from stock',
  'dd/drawdiscard - Draw the discard top',
  'm <idx...>     - Lay down a meld (3+ cards)',
  'lo <o> <m> <i> - Lay off card i onto owner o meld m',
  'd <idx>        - Discard a card',
  'nr/nextround   - Next round',
  'log            - Show action log',
  'r/reset        - Reset game',
];
