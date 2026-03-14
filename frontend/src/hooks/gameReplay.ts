export const REPLAY_DELAY_MS = 800;

export const delay = (ms: number) => new Promise<void>((resolve) => setTimeout(resolve, ms));

export interface ReplayConfig<TState> {
  buildReplayStates: (finalState: TState) => TState[];
  buildHumanActionState?: (finalState: TState) => TState | null;
  getActionDelay?: (finalState: TState, actionIndex: number) => number;
}

export async function runReplay<TState>(
  finalState: TState,
  setDisplayState: (state: TState) => void,
  config: ReplayConfig<TState>,
): Promise<void> {
  const humanState = config.buildHumanActionState?.(finalState) ?? null;
  if (humanState) {
    setDisplayState(humanState);
    await delay(REPLAY_DELAY_MS);
  }

  const replayStates = config.buildReplayStates(finalState);
  if (replayStates.length === 0) {
    setDisplayState(finalState);
    return;
  }

  for (let i = 0; i < replayStates.length; i++) {
    setDisplayState(replayStates[i]);
    const actionDelay = config.getActionDelay?.(finalState, i) ?? REPLAY_DELAY_MS;
    await delay(actionDelay);
  }

  setDisplayState(finalState);
}
