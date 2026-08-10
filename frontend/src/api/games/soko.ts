// API client for soko. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { FiveCardStudResponse } from '../../types/card';
import type { FiveCardStudConfigInput } from './fivecardstud';
import { createHoldemLikeApi } from './holdem';

/**
 * API client for the Soko /soko/exec endpoint.
 *
 * Soko shares Five Card Stud's request and response shapes exactly — the deal,
 * the commands and the phases are the same game, and only the showdown ranking
 * differs, which the server resolves into `handName`/`handRank` before the page
 * sees it. It reuses `FiveCardStudConfigInput` for the same reason.
 *
 * The one caveat: `handRank` is on Soko's own scale (a four-card straight and a
 * four-card flush sit between one pair and two pair), so it must not be compared
 * against another game's `handRank`.
 */
export const sokoApi = createHoldemLikeApi<FiveCardStudResponse, FiveCardStudConfigInput>('soko');
