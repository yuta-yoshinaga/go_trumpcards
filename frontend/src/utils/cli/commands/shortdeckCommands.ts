import type { shortdeckApi } from '../../../api/gameApi';
import type { CliParseResult } from '../types';
import { HOLDEM_HELP, parseHoldemCommand } from './holdemCommands';

type ShortDeckArgs = Parameters<typeof shortdeckApi.exec>;

/** Parse a Short Deck Hold'em CLI command into API exec arguments. */
export function parseShortdeckCommand(input: string): CliParseResult<ShortDeckArgs> {
  return parseHoldemCommand(input) as CliParseResult<ShortDeckArgs>;
}

/** Help text for Short Deck Hold'em CLI mode. */
export const SHORTDECK_HELP: string[] = HOLDEM_HELP;
