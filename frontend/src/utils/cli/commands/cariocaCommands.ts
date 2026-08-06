import type { cariocaApi } from '../../../api/gameApi';
import { parseIntArg, parseIntSlice, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type CariocaArgs = Parameters<typeof cariocaApi.exec>;

const VALID_COMMANDS = [
  'ds',
  'drawstock',
  'dd',
  'drawdiscard',
  'mc',
  'meldcontract',
  'me',
  'meldextra',
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

/**
 * Parse one "a,b,c" argument into card indices. The contract is met with one
 * such group per slot, mirroring `parseSlotIndices` in
 * `internal/adapter/controller/CariocaCuiController.go`.
 */
function parseSlotGroup(arg: string): number[] | null {
  const values: number[] = [];
  for (const part of arg.split(',')) {
    const n = Number(part);
    if (part === '' || !Number.isInteger(n)) return null;
    values.push(n);
  }
  return values.length > 0 ? values : null;
}

/** Parse a Carioca CLI command into API exec arguments. */
export function parseCariocaCommand(input: string): CliParseResult<CariocaArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'ds':
    case 'drawstock':
      return { args: ['drawstock'] };
    case 'dd':
    case 'drawdiscard':
      return { args: ['drawdiscard'] };
    case 'mc':
    case 'meldcontract': {
      // mc <a,b,c> <d,e,f> — one comma-separated group per contract slot.
      if (args.length === 0) return { error: 'Usage: mc <a,b,c> <d,e,f>' };
      const slots: number[][] = [];
      for (const arg of args) {
        const group = parseSlotGroup(arg);
        if (group === null) return { error: 'Usage: mc <a,b,c> <d,e,f>' };
        slots.push(group);
      }
      return { args: ['meldcontract', { indicesPerSlot: slots }] };
    }
    case 'me':
    case 'meldextra': {
      const parsed = parseIntSlice(args);
      if ('error' in parsed || parsed.values.length === 0) return { error: 'Usage: me <idx...>' };
      return { args: ['meldextra', { cardIndices: parsed.values }] };
    }
    case 'lo':
    case 'layoff': {
      // layoff <targetPlayerIdx> <meldIdx> <cardIndex>
      const parsed = parseIntSlice(args);
      if ('error' in parsed || parsed.values.length < 3) {
        return { error: 'Usage: lo <playerIdx> <meldIdx> <cardIdx>' };
      }
      return {
        args: ['layoff', { targetPlayerIdx: parsed.values[0], meldIdx: parsed.values[1], cardIndex: parsed.values[2] }],
      };
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

/** Help text for Carioca CLI mode. */
export const CARIOCA_HELP: string[] = [
  'ds/drawstock          - Draw from stock',
  'dd/drawdiscard        - Take the discard top',
  'mc <a,b,c> <d,e,f>    - Meet the round contract (one group per slot)',
  'me <idx...>           - Lay an extra meld (after the contract is met)',
  'lo <p> <m> <idx>      - Lay a card off onto player p, meld m',
  'd <idx>               - Discard a card',
  'nr/nextround          - Next round',
  'log                   - Show action log',
  'r/reset               - Reset game',
];
