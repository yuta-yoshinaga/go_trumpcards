import type { cegoApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type CegoArgs = Parameters<typeof cegoApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bid',
  'pass',
  'ct',
  'contract',
  'cego',
  'handspiel',
  'd',
  'discard',
  'keep',
  'p',
  'play',
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

/** Parse a Cego (チェゴ) CLI command into API exec arguments. */
export function parseCegoCommand(input: string): CliParseResult<CegoArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      const arg = args[0]?.toLowerCase() ?? '';
      if (arg !== 'play' && arg !== 'p' && arg !== '') return { error: 'Usage: bid play' };
      return { args: ['bid', { bid: 'play' }] };
    }
    case 'pass':
      return { args: ['pass'] };
    case 'ct':
    case 'contract': {
      const arg = args[0]?.toLowerCase() ?? '';
      if (arg === 'cego' || arg === 'c') return { args: ['contract', { contract: 'cego' }] };
      if (arg === 'handspiel' || arg === 'solo' || arg === 'h')
        return { args: ['contract', { contract: 'handspiel' }] };
      return { error: 'Usage: contract <cego|handspiel>' };
    }
    case 'cego':
      return { args: ['contract', { contract: 'cego' }] };
    case 'handspiel':
    case 'solo':
      return { args: ['contract', { contract: 'handspiel' }] };
    case 'd':
    case 'discard':
    case 'keep': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: keep <idx> (the single card to keep in the Cego exchange)' };
      return { args: ['discard', { cardIndices: [parsed.value] }] };
    }
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <idx>' };
      return { args: ['play', { cardIndex: parsed.value }] };
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

/** Help text for Cego (チェゴ) CLI mode. */
export const CEGO_HELP: string[] = [
  'bid play                         - Declare (bid to play the deal)',
  'pass                             - Pass the auction',
  'contract <cego|handspiel>        - Choose your contract (declarer)',
  'cego / handspiel                 - Shortcut for the contract choice',
  'keep <idx>                       - Cego exchange: keep this 1 card, lay down the rest',
  'p <idx>                          - Play a card (Play phase, must follow suit / trump)',
  'n/next                           - Next trick',
  'nr/nextround                     - Next deal',
  'h/hint                           - Show hint',
  'r/reset                          - Reset game',
];
