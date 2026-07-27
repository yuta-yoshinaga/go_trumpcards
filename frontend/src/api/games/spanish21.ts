// API client for spanish21. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import { createBlackJackLikeApi } from './blackjack';

/** API client for the Spanish 21 /spanish21/exec endpoint (shares BlackJack response shape). */
export const spanish21Api = createBlackJackLikeApi('spanish21');
