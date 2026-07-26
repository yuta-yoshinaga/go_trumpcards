import type { contractrummyApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type ContractRummyArgs = Parameters<typeof contractrummyApi.exec>;

const CR_COMMANDS = [
  'ds',
  'drawstock',
  'dd',
  'drawdiscard',
  'x',
  'discard',
  'meld',
  'extra',
  'lo',
  'layoff',
  'nr',
  'nextround',
  'log',
  'l',
  'r',
  'reset',
  'help',
  '?',
];

/**
 * Parse a whitespace/comma-separated list of 0-based indices.
 * Returns null if the list is empty or any token is not a non-negative integer.
 */
function parseIdxList(s: string): number[] | null {
  const parts = s.split(/[\s,]+/).filter((p) => p.length > 0);
  if (parts.length === 0) return null;
  const out: number[] = [];
  for (const p of parts) {
    const n = Number(p);
    if (!Number.isInteger(n) || n < 0) return null;
    out.push(n);
  }
  return out;
}

/** Parse a Contract Rummy CLI command into API exec arguments (indices are 0-based). */
export function parseContractRummyCommand(input: string): CliParseResult<ContractRummyArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'ds':
    case 'drawstock':
      return { args: ['drawstock'] as ContractRummyArgs };
    case 'dd':
    case 'drawdiscard':
      return { args: ['drawdiscard'] as ContractRummyArgs };
    case 'x':
    case 'discard': {
      const list = parseIdxList(args.join(' '));
      if (list?.length !== 1) return { error: 'Usage: discard <cardIdx>' };
      return { args: ['discard', { cardIndex: list[0] }] as ContractRummyArgs };
    }
    case 'meld': {
      // Slots are separated by '/', cards within a slot by spaces/commas.
      const slotStrs = args
        .join(' ')
        .split('/')
        .map((s) => s.trim())
        .filter((s) => s.length > 0);
      if (slotStrs.length === 0) return { error: 'Usage: meld <i,j,k> / <i,j,k> ...' };
      const indicesPerSlot: number[][] = [];
      for (const s of slotStrs) {
        const list = parseIdxList(s);
        if (!list) return { error: 'Usage: meld <i,j,k> / <i,j,k> ...' };
        indicesPerSlot.push(list);
      }
      return { args: ['meldcontract', { indicesPerSlot }] as ContractRummyArgs };
    }
    case 'extra': {
      const list = parseIdxList(args.join(' '));
      if (!list || list.length < 3) return { error: 'Usage: extra <i,j,k> (min 3 cards)' };
      return { args: ['meldextra', { cardIndices: list }] as ContractRummyArgs };
    }
    case 'lo':
    case 'layoff': {
      const list = parseIdxList(args.join(' '));
      if (list?.length !== 3) return { error: 'Usage: layoff <cardIdx> <playerIdx> <meldIdx>' };
      return {
        args: ['layoff', { cardIndex: list[0], targetPlayerIdx: list[1], meldIdx: list[2] }] as ContractRummyArgs,
      };
    }
    case 'nr':
    case 'nextround':
      return { args: ['nextround'] as ContractRummyArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as ContractRummyArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as ContractRummyArgs };
    default: {
      const suggestion = suggestCommand(cmd, CR_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Contract Rummy CLI mode. */
export const CONTRACTRUMMY_HELP: string[] = [
  'ds/drawstock          - Draw from the stock pile',
  'dd/drawdiscard        - Draw the discard top',
  'meld <i,j> / <i,j,k>  - Lay down the round contract (slots split by /)',
  'extra <i,j,k>         - Meld an extra set/run (min 3 cards)',
  'lo/layoff <c> <p> <m> - Lay off card c onto player p meld m',
  'x/discard <cardIdx>   - Discard a card',
  'nr/nextround          - Advance to the next round',
  'log                   - Show the action log',
  'r/reset               - Reset game',
];
