// API client for jokerpoker. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import { createVideoPokerApi } from '../gameExec';

/** API client for the Joker Poker /jokerpoker/exec endpoint. */
export const jokerpokerApi = createVideoPokerApi('jokerpoker');
