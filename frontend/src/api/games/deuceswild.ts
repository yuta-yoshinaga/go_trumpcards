// API client for deuceswild. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import { createVideoPokerApi } from '../gameExec';

/** API client for the Deuces Wild /deuceswild/exec endpoint. */
export const deuceswildApi = createVideoPokerApi('deuceswild');
