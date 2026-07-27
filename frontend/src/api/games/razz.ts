// API client for razz. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SevenCardStudResponse } from '../../types/card';
import { createHoldemLikeApi } from './holdem';
import type { SevenCardStudConfigInput } from './sevencardstud';

/** API client for the Razz /razz/exec endpoint. */
export const razzApi = createHoldemLikeApi<SevenCardStudResponse, SevenCardStudConfigInput>('razz');
