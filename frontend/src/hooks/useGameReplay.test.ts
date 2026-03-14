import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { REPLAY_DELAY_MS, type ReplayConfig, runReplay } from './useGameReplay';

describe('REPLAY_DELAY_MS', () => {
  it('is 800', () => {
    expect(REPLAY_DELAY_MS).toBe(800);
  });
});

describe('runReplay', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  const flushDelays = async () => {
    // Flush all pending promises and timers
    for (let i = 0; i < 20; i++) {
      await vi.advanceTimersByTimeAsync(REPLAY_DELAY_MS);
    }
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
