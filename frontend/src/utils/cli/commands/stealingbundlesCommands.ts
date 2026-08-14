import type { stealingbundlesApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type StealingBundlesArgs = Parameters<typeof stealingbundlesApi.exec>;

const VALID_COMMANDS = ['t', 'take', 's', 'steal', 'd', 'trail', 'h', 'hint', 'g', 'giveup', 'r', 'reset', 'log', 'l'];

/** Parse a Stealing Bundles CLI command into API exec arguments (indices are 0-based). */
export function parseStealingBundlesCommand(input: string): CliParseResult<StealingBundlesArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 't':
    case 'take': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: t <cardIdx>' };
      return { args: ['take', idx.value] as StealingBundlesArgs };
    }
    case 'd':
    case 'trail': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: d <cardIdx>' };
      return { args: ['trail', idx.value] as StealingBundlesArgs };
    }
    case 's':
    case 'steal': {
      // **札と相手の両方が要ります。** 相手を書かないと誰の束か決まりません。
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: s <cardIdx> <victimIdx>' };
      const victim = parseIntArg(args, 1);
      if ('error' in victim) return { error: 'Usage: s <cardIdx> <victimIdx>' };
      return { args: ['steal', idx.value, victim.value] as StealingBundlesArgs };
    }
    case 'h':
    case 'hint':
      return { args: ['hint'] as StealingBundlesArgs };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] as StealingBundlesArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as StealingBundlesArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as StealingBundlesArgs };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Stealing Bundles CLI mode. */
export const STEALINGBUNDLES_HELP: string[] = [
  't <cardIdx> - Capture matching cards from the table',
  "s <cardIdx> <victimIdx> - Steal that seat's whole bundle (its top card must match)",
  'd <cardIdx> - Place a card on the table (only when nothing can be captured)',
  'h/hint      - Show a hint',
  'g/giveup    - Give up',
  'log         - Show the action log',
  'r/reset     - Reset game',
];
