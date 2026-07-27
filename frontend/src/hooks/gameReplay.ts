import { getReplaySpeedMultiplier } from './useReplaySpeed';

/** Default delay in milliseconds between replay animation steps. */
export const REPLAY_DELAY_MS = 800;

/** Return a promise that resolves after the given milliseconds. */
export const delay = (ms: number) => new Promise<void>((resolve) => setTimeout(resolve, ms));

/**
 * Apply the user's CPU replay speed preference (#1649) to a base delay.
 * Multiplier is read fresh from localStorage so settings changes take effect
 * on the very next replay step.
 */
function scaledDelay(baseMs: number): number {
  return Math.round(baseMs * getReplaySpeedMultiplier());
}

/** Configuration for the replay animation runner. */
export interface ReplayConfig<TState> {
  buildReplayStates: (finalState: TState) => TState[];
  buildHumanActionState?: (finalState: TState) => TState | null;
  getActionDelay?: (finalState: TState, actionIndex: number) => number;
  /**
   * Whether the calling component is still mounted. A replay sleeps between steps,
   * so it routinely outlives the page a player navigates away from; without this it
   * keeps calling `setDisplayState` on a gone component, and such a write can throw
   * once the environment is torn down. Supplying it also stops the remaining timers
   * rather than merely suppressing their writes. See issue #4447.
   */
  isMounted?: () => boolean;
}

/**
 * Check if replay should be skipped because cpuActions haven't changed.
 * Updates the ref and returns true if replay was skipped (setDisplayState called directly).
 */
export function shouldSkipReplay<TState>(
  newActions: unknown[],
  lastReplayedActionsRef: { current: unknown[] | undefined },
  res: TState,
  setDisplayState: (state: TState) => void,
): boolean {
  if (JSON.stringify(lastReplayedActionsRef.current) === JSON.stringify(newActions)) {
    setDisplayState(res);
    return true;
  }
  lastReplayedActionsRef.current = newActions;
  return false;
}

/** Run a step-by-step replay animation of CPU actions. */
export async function runReplay<TState>(
  finalState: TState,
  setDisplayState: (state: TState) => void,
  config: ReplayConfig<TState>,
): Promise<void> {
  const humanState = config.buildHumanActionState?.(finalState) ?? null;
  if (humanState) {
    setDisplayState(humanState);
    await delay(scaledDelay(REPLAY_DELAY_MS));
    if (config.isMounted?.() === false) return;
  }

  const replayStates = config.buildReplayStates(finalState);
  if (replayStates.length === 0) {
    setDisplayState(finalState);
    return;
  }

  for (let i = 0; i < replayStates.length; i++) {
    setDisplayState(replayStates[i]);
    const actionDelay = config.getActionDelay?.(finalState, i) ?? REPLAY_DELAY_MS;
    await delay(scaledDelay(actionDelay));
    if (config.isMounted?.() === false) return;
  }

  setDisplayState(finalState);
}
