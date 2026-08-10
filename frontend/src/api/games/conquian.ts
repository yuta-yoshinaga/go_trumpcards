// API client for conquian. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ConquianResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Conquian game settings. */
export interface ConquianConfigInput {
  cpuDifficulty?: number;
  targetWins?: number;
}

/** API client for the Conquian /conquian/exec endpoint. */
export const conquianApi = {
  exec: (
    command: 'reset' | 'drawstock' | 'drawdiscard' | 'meld' | 'discard' | 'nextround' | 'log',
    cardIndex?: number,
    config?: ConquianConfigInput,
    meldGroups?: number[][],
    /**
     * どのテーブルメルドへ足すかの指定 (meldGroups と同じ並び)。省略すると
     * バックエンドは従来どおり最初に延長できるメルドを選ぶ (#4837)。
     */
    extendTargets?: number[],
  ) =>
    gameExec<ConquianResponse>('conquian', {
      command,
      cardIndex,
      config,
      meldGroups,
      extendTargets,
    }),
};
