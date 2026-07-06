import type { wizardApi } from '../../../api/gameApi';
import { parseIntArg } from '../commandParserBase';
import type { CliParseResult } from '../types';
import { parseTrickCommand, TRICK_HELP } from './sharedTrickCommands';

type WizardArgs = Parameters<typeof wizardApi.exec>;

const EXTRA_COMMANDS = ['bid'];

/** Parse a Wizard CLI command into API exec arguments. */
export function parseWizardCommand(input: string): CliParseResult<WizardArgs> {
  const result = parseTrickCommand(input, EXTRA_COMMANDS, (cmd, args) => {
    if (cmd === 'bid') {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: bid <n>' };
      return { command: 'bid', bid: parsed.value };
    }
    return null;
  });

  if ('error' in result) return { error: result.error };
  if (result.command === 'bid') {
    return { args: ['bid', result.bid] };
  }
  return { args: [result.command as WizardArgs[0], undefined, result.cardIndex] };
}

/** Help text for Wizard CLI mode. */
export const WIZARD_HELP: string[] = ['bid <n>     - Place bid', ...TRICK_HELP];
