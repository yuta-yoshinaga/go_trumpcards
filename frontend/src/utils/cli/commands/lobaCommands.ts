import type { lobaApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type LobaArgs = Parameters<typeof lobaApi.exec>;

const VALID_COMMANDS = [
  'ds',
  'drawstock',
  'dd',
  'drawdiscard',
  'm',
  'meld',
  'o',
  'layoff',
  'd',
  'discard',
  'n',
  'next',
  'log',
  'r',
  'reset',
  'h',
  'hint',
  'help',
  '?',
];

/** Parse a Loba CLI command into API exec arguments. */
export function parseLobaCommand(input: string): CliParseResult<LobaArgs> {
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
        const n = Number(p.trim());
        if (!Number.isInteger(n) || n < 0) return { error: `Invalid card index: ${p.trim()}` };
        idxs.push(n);
      }
      return { args: ['meld', undefined, undefined, idxs] };
    }
    case 'o':
    case 'layoff': {
      if (args.length < 2) return { error: `Usage: ${cmd} <card> <meld>` };
      const card = Number(args[0]);
      if (!Number.isInteger(card) || card < 0) return { error: `Invalid card index: ${args[0]}` };
      const meld = Number(args[1]);
      if (!Number.isInteger(meld) || meld < 0) return { error: `Invalid meld index: ${args[1]}` };
      // A hand index and a table index are different things; keeping them in
      // their own positions means one can never be read as the other.
      return { args: ['layoff', card, meld] };
    }
    case 'd':
    case 'discard': {
      if (args.length === 0) return { error: `Usage: ${cmd} <index>` };
      const n = Number(args[0]);
      if (!Number.isInteger(n) || n < 0) return { error: `Invalid card index: ${args[0]}` };
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

/** Help text for Loba CLI mode. */
export const LOBA_HELP: string[] = [
  'ds              - Draw from the stock',
  'dd              - Take the top discard',
  'm <i,j,k>       - Put down hand cards i, j and k as a meld',
  'o <i> <m>       - Add hand card i to meld m',
  'd <i>           - Discard hand card i',
  'n/next          - Deal the next round',
  'log             - Show action log',
  'r/reset         - New game',
  'h/hint      - Get a hint',
];
