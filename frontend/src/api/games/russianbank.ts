// API client for russianbank. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { RussianBankResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Commands accepted by the Russian Bank /russianbank/exec endpoint. */
export type RussianBankCommand = 'reset' | 'pf' | 'mt' | 'd' | 's' | 'u' | 'hint' | 'log';

/** Options accepted by a Russian Bank move command. */
export interface RussianBankMoveOpts {
  zone?: number;
  fromOpp?: boolean;
  col?: number;
  toCol?: number;
  config?: { cpuDifficulty?: number };
}

/** API client for the Russian Bank /russianbank/exec endpoint. */
export const russianbankApi = {
  exec: (command: RussianBankCommand, opts?: RussianBankMoveOpts) =>
    gameExec<RussianBankResponse>('russianbank', {
      command,
      zone: opts?.zone,
      fromOpp: opts?.fromOpp,
      col: opts?.col,
      toCol: opts?.toCol,
      config: opts?.config,
    }),
};
