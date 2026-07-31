import type { desmocheApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type DesmocheArgs = Parameters<typeof desmocheApi.exec>;

const VALID_COMMANDS = [
  'ds',
  'drawstock',
  'dd',
  'drawdiscard',
  'm',
  'meld',
  'o',
  'layoff',
  'x',
  'desmoche',
  'd',
  'discard',
  'n',
  'next',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a non-negative integer argument, or null when it is not one. */
function parseIdx(raw: string | undefined): number | null {
  const n = Number(raw);
  if (raw === undefined || raw.trim() === '' || !Number.isInteger(n) || n < 0) return null;
  return n;
}

/** Parse a Desmoche CLI command into API exec arguments. */
export function parseDesmocheCommand(input: string): CliParseResult<DesmocheArgs> {
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
      if (args.length === 0) return { error: `Usage: ${cmd} <i,j,k>` };
      // Several cards in one command, so they are comma-separated. Join the
      // arguments back together first: splitCommand cuts on whitespace, so
      // "m 0, 2, 5" would otherwise only see "0,".
      const parts = args.join('').split(',');
      const idxs: number[] = [];
      for (const p of parts) {
        const n = parseIdx(p.trim());
        if (n === null) return { error: `Invalid card index: ${p.trim()}` };
        idxs.push(n);
      }
      return { args: ['meld', undefined, undefined, idxs] };
    }
    case 'o':
    case 'layoff': {
      if (args.length < 2) return { error: `Usage: ${cmd} <card> <meld>` };
      const card = parseIdx(args[0]);
      if (card === null) return { error: `Invalid card index: ${args[0]}` };
      const meld = parseIdx(args[1]);
      if (meld === null) return { error: `Invalid meld index: ${args[1]}` };
      // A hand index and a table index are different things; keeping them in
      // their own positions means one can never be read as the other.
      return { args: ['layoff', card, meld] };
    }
    case 'x':
    case 'desmoche': {
      if (args.length < 3) return { error: `Usage: ${cmd} <from> <card> <to>` };
      const from = parseIdx(args[0]);
      if (from === null) return { error: `Invalid meld index: ${args[0]}` };
      // This index counts within the source meld, not the hand.
      const card = parseIdx(args[1]);
      if (card === null) return { error: `Invalid card index: ${args[1]}` };
      const to = parseIdx(args[2]);
      if (to === null) return { error: `Invalid meld index: ${args[2]}` };
      return { args: ['desmoche', card, undefined, undefined, { fromMeldIndex: from, toMeldIndex: to }] };
    }
    case 'd':
    case 'discard': {
      if (args.length === 0) return { error: `Usage: ${cmd} <index>` };
      const n = parseIdx(args[0]);
      if (n === null) return { error: `Invalid card index: ${args[0]}` };
      return { args: ['discard', n] };
    }
    case 'n':
    case 'next':
      return { args: ['next'] };
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

/** Help text for Desmoche CLI mode. */
export const DESMOCHE_HELP: string[] = [
  'ds              - Draw from the stock',
  'dd              - Take the top discard',
  'm <i,j,k>       - Put down hand cards i, j and k as a meld',
  'o <i> <m>       - Add hand card i to meld m',
  'x <f> <i> <t>   - Move card i of meld f into meld t',
  'd <i>           - Discard hand card i',
  'n/next          - Deal the next round',
  'log             - Show action log',
  'r/reset         - New game',
];
