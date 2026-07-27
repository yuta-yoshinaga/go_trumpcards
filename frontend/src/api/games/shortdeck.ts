// API client for shortdeck. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ShortDeckResponse } from '../../types/card';
import { createHoldemLikeApi } from './holdem';

/** API client for the Short Deck Hold'em /shortdeck/exec endpoint. */
export const shortdeckApi = createHoldemLikeApi<ShortDeckResponse>('shortdeck');
