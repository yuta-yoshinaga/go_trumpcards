// API client for nainjaune. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { NainJauneResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Commands the /nainjaune/exec endpoint accepts. */
export type NainJauneCommand = 'reset' | 'play' | 'next' | 'hint' | 'log';

/** API client for the Le Nain Jaune /nainjaune/exec endpoint. */
export const nainjauneApi = {
  exec: (command: NainJauneCommand, cardIndex?: number) =>
    gameExec<NainJauneResponse>('nainjaune', { command, cardIndex }),
};
