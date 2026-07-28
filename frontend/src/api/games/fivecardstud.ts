// API client for fivecardstud. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { FiveCardStudResponse } from '../../types/card';
import { createHoldemLikeApi } from './holdem';

/** Configuration options for Five Card Stud game settings. */
export interface FiveCardStudConfigInput {
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

/** API client for the Five Card Stud /fivecardstud/exec endpoint. */
export const fiveCardStudApi = createHoldemLikeApi<FiveCardStudResponse, FiveCardStudConfigInput>('fivecardstud');
