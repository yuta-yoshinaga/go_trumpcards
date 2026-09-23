// API client for doubleexposure. Split out of gameApi.ts;
// gameApi.ts re-exports this file, so existing imports keep working.

import { createBlackJackLikeApi } from './blackjack';

/** API client for the Double Exposure Blackjack /doubleexposure/exec endpoint. */
export const doubleexposureApi = createBlackJackLikeApi('doubleexposure');
