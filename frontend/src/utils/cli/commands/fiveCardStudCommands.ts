import type { fiveCardStudApi } from '../../../api/gameApi';
import type { CliParseResult } from '../types';

type FiveCardStudArgs = Parameters<typeof fiveCardStudApi.exec>;

const VALID_COMMANDS = [
  'reset',
  'fold',
  'check',
  'call',
  'bet',
  'raise',
  'allin',
  'rebuy',
  'skiprebuy',
  'addon',
  'skipaddon',
  'muck',
  'show',
] as const;

type FiveCardStudCommand = (typeof VALID_COMMANDS)[number];

/** Parse a Five Card Stud CLI command into API exec arguments. */
export function parseFiveCardStudCommand(input: string): CliParseResult<FiveCardStudArgs> {
  const parts = input.trim().split(/\s+/);
  const cmd = parts[0]?.toLowerCase() ?? '';
  const amount = parts[1] ? Number.parseInt(parts[1], 10) : undefined;
  if ((VALID_COMMANDS as readonly string[]).includes(cmd)) {
    return { args: [cmd as FiveCardStudCommand, amount] as FiveCardStudArgs };
  }
  return { error: `Unknown command: ${cmd}` };
}

/** Help text for Five Card Stud CLI mode. */
export const FIVECARDSTUD_HELP: string[] = [
  'f/fold      - Fold',
  'ck/check    - Check',
  'c/call      - Call',
  'b/bet <amt> - Bet',
  'ra/raise <amt> - Raise',
  'a/allin     - All-in',
  'rebuy       - Rebuy chips',
  'skiprebuy   - Skip rebuy',
  'addon       - Add-on chips',
  'skipaddon   - Skip add-on',
  'muck        - Muck hand',
  'show        - Show hand',
  'r/reset     - Reset game',
];
