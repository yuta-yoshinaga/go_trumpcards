// API client for sevencardstudhilo. Split-file layout introduced by issue
// #4434; gameApi.ts re-exports this file, so existing imports keep working.

import type { SevenCardStudResponse } from '../../types/card';
import { createHoldemLikeApi } from './holdem';
import type { SevenCardStudConfigInput } from './sevencardstud';

/** API client for the Seven Card Stud Hi-Lo /sevencardstudhilo/exec endpoint.
 * Shares the SevenCardStudResponse shape -- the split only adds fields, which
 * ride along as omitempty. */
export const sevenCardStudHiLoApi = createHoldemLikeApi<SevenCardStudResponse, SevenCardStudConfigInput>(
  'sevencardstudhilo',
);
