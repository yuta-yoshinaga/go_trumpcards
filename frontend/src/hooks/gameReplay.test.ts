import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { REPLAY_DELAY_MS, type ReplayConfig, runReplay, shouldSkipReplay } from './gameReplay';
import { REPLAY_SPEED_STORAGE_KEY } from './useReplaySpeed';

describe('runReplay', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  const flushDelays = async () => {
    await vi.runAllTimersAsync();
  };

  it('sets final state immediately when no human action and no replay states', async () => {
    const setDisplayState = vi.fn();
    const config: ReplayConfig<string> = {
      buildReplayStates: () => [],
    };

    const promise = runReplay('final', setDisplayState, config);
    await flushDelays();
    await promise;

    expect(setDisplayState).toHaveBeenCalledTimes(1);
    expect(setDisplayState).toHaveBeenCalledWith('final');
  });

  it('shows human action state with delay before replay', async () => {
    const setDisplayState = vi.fn();
    const config: ReplayConfig<string> = {
      buildReplayStates: () => [],
      buildHumanActionState: () => 'human-action',
    };

    const promise = runReplay('final', setDisplayState, config);
    await flushDelays();
    await promise;

    expect(setDisplayState).toHaveBeenCalledTimes(2);
    expect(setDisplayState).toHaveBeenNthCalledWith(1, 'human-action');
    expect(setDisplayState).toHaveBeenNthCalledWith(2, 'final');
  });

  it('skips human action when builder returns null', async () => {
    const setDisplayState = vi.fn();
    const config: ReplayConfig<string> = {
      buildReplayStates: () => [],
      buildHumanActionState: () => null,
    };

    const promise = runReplay('final', setDisplayState, config);
    await flushDelays();
    await promise;

    expect(setDisplayState).toHaveBeenCalledTimes(1);
    expect(setDisplayState).toHaveBeenCalledWith('final');
  });

  it('loops through replay states with REPLAY_DELAY_MS', async () => {
    const setDisplayState = vi.fn();
    const config: ReplayConfig<string> = {
      buildReplayStates: () => ['r1', 'r2'],
    };

    const promise = runReplay('final', setDisplayState, config);
    await flushDelays();
    await promise;

    expect(setDisplayState).toHaveBeenCalledTimes(3);
    expect(setDisplayState).toHaveBeenNthCalledWith(1, 'r1');
    expect(setDisplayState).toHaveBeenNthCalledWith(2, 'r2');
    expect(setDisplayState).toHaveBeenNthCalledWith(3, 'final');
  });

  it('shows human action then replay states then final', async () => {
    const setDisplayState = vi.fn();
    const config: ReplayConfig<string> = {
      buildReplayStates: () => ['r1'],
      buildHumanActionState: () => 'human',
    };

    const promise = runReplay('final', setDisplayState, config);
    await flushDelays();
    await promise;

    expect(setDisplayState).toHaveBeenCalledTimes(3);
    expect(setDisplayState).toHaveBeenNthCalledWith(1, 'human');
    expect(setDisplayState).toHaveBeenNthCalledWith(2, 'r1');
    expect(setDisplayState).toHaveBeenNthCalledWith(3, 'final');
  });

  it('scales delays by the persisted CPU replay speed (fast)', async () => {
    localStorage.setItem(REPLAY_SPEED_STORAGE_KEY, 'fast');
    try {
      const setDisplayState = vi.fn();
      const config: ReplayConfig<string> = {
        buildReplayStates: () => ['r1'],
        buildHumanActionState: () => 'human',
      };

      const promise = runReplay('final', setDisplayState, config);

      // 'fast' = 0.3 multiplier, so 800 -> 240ms. After 239ms only the human
      // state has been shown; after the full 240ms we move on.
      await vi.advanceTimersByTimeAsync(239);
      expect(setDisplayState).toHaveBeenCalledTimes(1);
      expect(setDisplayState).toHaveBeenCalledWith('human');

      await vi.advanceTimersByTimeAsync(1);
      expect(setDisplayState).toHaveBeenNthCalledWith(2, 'r1');

      await vi.runAllTimersAsync();
      await promise;
      expect(setDisplayState).toHaveBeenLastCalledWith('final');
    } finally {
      localStorage.removeItem(REPLAY_SPEED_STORAGE_KEY);
    }
  });

  it('collapses delays to zero on instant speed', async () => {
    localStorage.setItem(REPLAY_SPEED_STORAGE_KEY, 'instant');
    try {
      const setDisplayState = vi.fn();
      const config: ReplayConfig<string> = {
        buildReplayStates: () => ['r1', 'r2'],
        buildHumanActionState: () => 'human',
      };

      const promise = runReplay('final', setDisplayState, config);
      await vi.runAllTimersAsync();
      await promise;

      expect(setDisplayState).toHaveBeenNthCalledWith(1, 'human');
      expect(setDisplayState).toHaveBeenNthCalledWith(2, 'r1');
      expect(setDisplayState).toHaveBeenNthCalledWith(3, 'r2');
      expect(setDisplayState).toHaveBeenNthCalledWith(4, 'final');
    } finally {
      localStorage.removeItem(REPLAY_SPEED_STORAGE_KEY);
    }
  });

  it('keeps base delay constant when speed is normal', () => {
    expect(REPLAY_DELAY_MS).toBe(800);
  });

  it('uses custom getActionDelay per action', async () => {
    const setDisplayState = vi.fn();
    const getActionDelay = vi.fn((_state: string, i: number) => (i === 0 ? 300 : 1200));
    const config: ReplayConfig<string> = {
      buildReplayStates: () => ['r1', 'r2'],
      getActionDelay,
    };

    const promise = runReplay('final', setDisplayState, config);
    await flushDelays();
    await promise;

    expect(getActionDelay).toHaveBeenCalledTimes(2);
    expect(getActionDelay).toHaveBeenCalledWith('final', 0);
    expect(getActionDelay).toHaveBeenCalledWith('final', 1);
  });
  // A replay sleeps between steps, so it routinely outlives a page the player left.
  // It must stop rather than keep driving state into a gone component (#4447).
  it('abandons the remaining steps once the component has unmounted', async () => {
    const setDisplayState = vi.fn();
    let mounted = true;
    const config: ReplayConfig<string> = {
      buildReplayStates: () => ['a', 'b', 'c'],
      isMounted: () => mounted,
    };
    const promise = runReplay('final', setDisplayState, config);
    // First step renders, then the player navigates away mid-sleep.
    mounted = false;
    await flushDelays();
    await promise;
    expect(setDisplayState).toHaveBeenCalledWith('a');
    expect(setDisplayState).not.toHaveBeenCalledWith('c');
    expect(setDisplayState).not.toHaveBeenCalledWith('final');
  });

  // The other bail-out: the human-action state renders first, then the replay sleeps
  // before the CPU steps. Losing the page during THAT sleep must stop it too (#4447).
  it('abandons the replay if the page is lost during the human-action delay', async () => {
    const setDisplayState = vi.fn();
    let mounted = true;
    const config: ReplayConfig<string> = {
      buildHumanActionState: () => 'human',
      buildReplayStates: () => ['a', 'b'],
      isMounted: () => mounted,
    };
    const promise = runReplay('final', setDisplayState, config);
    mounted = false;
    await flushDelays();
    await promise;
    expect(setDisplayState).toHaveBeenCalledWith('human');
    expect(setDisplayState).not.toHaveBeenCalledWith('a');
    expect(setDisplayState).not.toHaveBeenCalledWith('final');
  });

  it('runs every step when the component stays mounted', async () => {
    const setDisplayState = vi.fn();
    const config: ReplayConfig<string> = {
      buildReplayStates: () => ['a', 'b'],
      isMounted: () => true,
    };
    const promise = runReplay('final', setDisplayState, config);
    await flushDelays();
    await promise;
    expect(setDisplayState).toHaveBeenCalledWith('a');
    expect(setDisplayState).toHaveBeenCalledWith('b');
    expect(setDisplayState).toHaveBeenLastCalledWith('final');
  });
});

describe('shouldSkipReplay', () => {
  it('returns false and updates ref on first call (ref is undefined)', () => {
    const ref = { current: undefined as unknown[] | undefined };
    const setDisplayState = vi.fn();
    const actions = [{ id: 1 }];

    const skipped = shouldSkipReplay(actions, ref, 'state', setDisplayState);

    expect(skipped).toBe(false);
    expect(ref.current).toEqual(actions);
    expect(setDisplayState).not.toHaveBeenCalled();
  });

  it('returns true and calls setDisplayState when actions unchanged', () => {
    const actions = [{ id: 1 }];
    const ref = { current: [{ id: 1 }] as unknown[] | undefined };
    const setDisplayState = vi.fn();

    const skipped = shouldSkipReplay(actions, ref, 'state', setDisplayState);

    expect(skipped).toBe(true);
    expect(setDisplayState).toHaveBeenCalledWith('state');
  });

  it('returns false and updates ref when actions change', () => {
    const ref = { current: [{ id: 1 }] as unknown[] | undefined };
    const setDisplayState = vi.fn();
    const newActions = [{ id: 2 }];

    const skipped = shouldSkipReplay(newActions, ref, 'state', setDisplayState);

    expect(skipped).toBe(false);
    expect(ref.current).toEqual(newActions);
    expect(setDisplayState).not.toHaveBeenCalled();
  });

  it('returns false on first call with empty actions (ref undefined vs [])', () => {
    const ref = { current: undefined as unknown[] | undefined };
    const setDisplayState = vi.fn();

    const skipped = shouldSkipReplay([], ref, 'state', setDisplayState);

    expect(skipped).toBe(false);
    expect(ref.current).toEqual([]);
  });

  it('returns true when both current and new are empty arrays', () => {
    const ref = { current: [] as unknown[] | undefined };
    const setDisplayState = vi.fn();

    const skipped = shouldSkipReplay([], ref, 'state', setDisplayState);

    expect(skipped).toBe(true);
    expect(setDisplayState).toHaveBeenCalledWith('state');
  });
});

describe('shouldSkipReplay', () => {
  it('returns false and updates ref on first call (ref is undefined)', () => {
    const ref = { current: undefined as unknown[] | undefined };
    const setDisplayState = vi.fn();
    const actions = [{ id: 1 }];

    const skipped = shouldSkipReplay(actions, ref, 'state', setDisplayState);

    expect(skipped).toBe(false);
    expect(ref.current).toEqual(actions);
    expect(setDisplayState).not.toHaveBeenCalled();
  });

  it('returns true and calls setDisplayState when actions unchanged', () => {
    const actions = [{ id: 1 }];
    const ref = { current: [{ id: 1 }] as unknown[] | undefined };
    const setDisplayState = vi.fn();

    const skipped = shouldSkipReplay(actions, ref, 'state', setDisplayState);

    expect(skipped).toBe(true);
    expect(setDisplayState).toHaveBeenCalledWith('state');
  });

  it('returns false and updates ref when actions change', () => {
    const ref = { current: [{ id: 1 }] as unknown[] | undefined };
    const setDisplayState = vi.fn();
    const newActions = [{ id: 2 }];

    const skipped = shouldSkipReplay(newActions, ref, 'state', setDisplayState);

    expect(skipped).toBe(false);
    expect(ref.current).toEqual(newActions);
    expect(setDisplayState).not.toHaveBeenCalled();
  });

  it('returns false on first call with empty actions (ref undefined vs [])', () => {
    const ref = { current: undefined as unknown[] | undefined };
    const setDisplayState = vi.fn();

    const skipped = shouldSkipReplay([], ref, 'state', setDisplayState);

    expect(skipped).toBe(false);
    expect(ref.current).toEqual([]);
  });

  it('returns true when both current and new are empty arrays', () => {
    const ref = { current: [] as unknown[] | undefined };
    const setDisplayState = vi.fn();

    const skipped = shouldSkipReplay([], ref, 'state', setDisplayState);

    expect(skipped).toBe(true);
    expect(setDisplayState).toHaveBeenCalledWith('state');
  });
});
