// API client for videopoker. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import { createVideoPokerApi } from '../gameExec';

/** API client for the Video Poker /videopoker/exec endpoint. */
export const videopokerApi = createVideoPokerApi('videopoker');
