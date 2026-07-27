// API client for wizard. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { WizardResponse } from '../../types/card';
import { createBidPlayApi } from '../gameExec';

/** Configuration options for Wizard game settings. */
export interface WizardConfigInput {
  cpuDifficulty?: number;
}

/** API client for the Wizard /wizard/exec endpoint. */
export const wizardApi = createBidPlayApi<WizardResponse, WizardConfigInput>('wizard');
