import type { omahaApi } from '../../../api/gameApi';
import type { CliParseResult } from '../types';
import { HOLDEM_HELP, parseHoldemCommand } from './holdemCommands';

type OmahaArgs = Parameters<typeof omahaApi.exec>;

/** Parse an Omaha Hold'em CLI command into API exec arguments. */
export function parseOmahaCommand(input: string): CliParseResult<OmahaArgs> {
  return parseHoldemCommand(input) as CliParseResult<OmahaArgs>;
}

/** Help text for Omaha Hold'em CLI mode. */
export const OMAHA_HELP: string[] = HOLDEM_HELP;
