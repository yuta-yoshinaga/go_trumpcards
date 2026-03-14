import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { type ReplayConfig, runReplay, shouldSkipReplay } from './gameReplay';

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
