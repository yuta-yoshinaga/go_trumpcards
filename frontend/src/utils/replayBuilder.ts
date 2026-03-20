/** Generic two-pass replay state builder.
 *
 *  Many games share the same pattern: reverse all CPU actions from the final
 *  state to reconstruct the "before" context, then replay forward to produce
 *  one intermediate display state per action. This utility encapsulates that
 *  pattern so each game only provides the game-specific callbacks.
 */

/** Configuration for building replay animation states from CPU actions. */
export interface ReplayBuilderConfig<TResponse, TAction, TCtx> {
  actions: TAction[];
  finalState: TResponse;
  initContext: (finalState: TResponse) => TCtx;
  reverseAction: (ctx: TCtx, action: TAction) => void;
  applyAction: (ctx: TCtx, action: TAction) => void;
  buildState: (
    finalState: TResponse,
    ctx: TCtx,
    action: TAction,
    processedActions: TAction[],
    isLast: boolean,
  ) => TResponse;
}

/** Build intermediate display states by reversing then replaying CPU actions. */
export function buildReplayStates<TResponse, TAction, TCtx>(
  config: ReplayBuilderConfig<TResponse, TAction, TCtx>,
): TResponse[] {
  const { actions, finalState, initContext, reverseAction, applyAction, buildState } = config;
  if (actions.length === 0) return [];

  const ctx = initContext(finalState);
  for (let i = actions.length - 1; i >= 0; i--) {
    reverseAction(ctx, actions[i]);
  }

  const states: TResponse[] = [];
  for (let i = 0; i < actions.length; i++) {
    applyAction(ctx, actions[i]);
    const isLast = i === actions.length - 1;
    states.push(buildState(finalState, ctx, actions[i], actions.slice(0, i + 1), isLast));
  }
  return states;
}

/** Configuration for building the pre-CPU-action state after a human action. */
export interface HumanActionStateConfig<TResponse, TAction, TCtx> {
  actions: TAction[];
  finalState: TResponse;
  initContext: (finalState: TResponse) => TCtx;
  reverseAction: (ctx: TCtx, action: TAction) => void;
  buildState: (finalState: TResponse, ctx: TCtx) => TResponse;
}

/** Build the display state showing the human action before CPU replay begins. */
export function buildHumanActionState<TResponse, TAction, TCtx>(
  config: HumanActionStateConfig<TResponse, TAction, TCtx>,
): TResponse | null {
  const { actions, finalState, initContext, reverseAction, buildState } = config;
  if (actions.length === 0) return null;

  const ctx = initContext(finalState);
  for (let i = actions.length - 1; i >= 0; i--) {
    reverseAction(ctx, actions[i]);
  }

  return buildState(finalState, ctx);
}
