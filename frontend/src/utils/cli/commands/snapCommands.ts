import type { snapApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type SnapArgs = Parameters<typeof snapApi.exec>;

const VALID_COMMANDS = ['s', 'step', 'n', 'snap', 't', 'tick', 'h', 'hint', 'g', 'giveup', 'r', 'reset', 'log', 'l'];

/** Parse a Snap CLI command into API exec arguments. */
export function parseSnapCommand(input: string): CliParseResult<SnapArgs> {
  const { cmd } = splitCommand(input);

  switch (cmd) {
    case 's':
    case 'step':
      return { args: ['step'] as SnapArgs };
    case 'n':
    case 'snap':
      // **席は取らない。** 席を選べると CPU に誤宣言させられる。
      return { args: ['snap'] as SnapArgs };
    case 't':
    case 'tick':
      return { args: ['tick'] as SnapArgs };
    case 'h':
    case 'hint':
      return { args: ['hint'] as SnapArgs };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] as SnapArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as SnapArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as SnapArgs };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Snap CLI mode. */
export const SNAP_HELP: string[] = [
  's/step      - Turn the top card of your stock over',
  'n/snap      - Call snap (no seat argument: it is always you)',
  't/tick      - Let the CPUs react',
  'h/hint      - Show a hint',
  'g/giveup    - Give up',
  'log         - Show the action log',
  'r/reset     - Reset game',
];
