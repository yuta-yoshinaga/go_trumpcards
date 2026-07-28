// API client for sevencardstud. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SevenCardStudResponse } from '../../types/card';
import { createHoldemLikeApi } from './holdem';

/** Configuration options for Seven Card Stud game settings. */
export interface SevenCardStudConfigInput {
  ante?: number;
  bringIn?: number;
  smallBet?: number;
  bigBet?: number;
  tournamentMode?: boolean;
  anteLevelHands?: number;
  anteMultiplier?: number;
  bettingLimit?: number;
  tableSize?: number;
  rebuyEnabled?: boolean;
  rebuyMaxCount?: number;
  rebuyChips?: number;
  rebuyPeriodHands?: number;
  addonEnabled?: boolean;
  addonChips?: number;
  addonAfterHand?: number;
  cpuMetaAI?: boolean;
}

/** API client for the Seven Card Stud /sevencardstud/exec endpoint. */
export const sevenCardStudApi = createHoldemLikeApi<SevenCardStudResponse, SevenCardStudConfigInput>('sevencardstud');
