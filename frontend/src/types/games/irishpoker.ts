// Type declarations for irishpoker. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { PineappleResponse } from './pineapple';

/** Irish Poker shares the same response shape as Pineapple. */
export type IrishPokerResponse = PineappleResponse;
