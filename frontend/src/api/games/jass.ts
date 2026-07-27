// API client for jass. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { JassResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Jass (Schieber) game configuration input shape. */
export interface JassConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
  lastTrickBonus?: number;
  enableWeis?: boolean;
}

/** API client for the Jass /jass/exec endpoint. */
export const jassApi = {
  exec: (
    command: 'reset' | 'calltrump' | 'schieben' | 'play' | 'next' | 'nextround' | 'hint',
    suit?: number,
    cardIndex?: number,
    config?: JassConfigInput,
  ) =>
    gameExec<JassResponse>('jass', {
      command,
      suit,
      cardIndex,
      config,
    }),
};
