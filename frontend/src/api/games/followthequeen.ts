// API client for followthequeen. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { FollowTheQueenResponse } from '../../types/card';
import { createHoldemLikeApi } from './holdem';

/** Configuration options for Follow the Queen game settings. */
export interface FollowTheQueenConfigInput {
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

/** API client for the Follow the Queen /followthequeen/exec endpoint. */
export const followTheQueenApi = createHoldemLikeApi<FollowTheQueenResponse, FollowTheQueenConfigInput>(
  'followthequeen',
);
