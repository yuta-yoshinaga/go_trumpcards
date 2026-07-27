import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi } from '../api/gameApi';
import { useActionLog } from './useActionLog';

vi.mock('../api/gameApi', () => ({
  actionLogApi: {
    blackjack: vi.fn(),
  },
}));

describe('useActionLog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('initializes with null actionLog', () => {
    const { result } = renderHook(() => useActionLog('blackjack'));
    expect(result.current.actionLog).toBeNull();
  });

  it('fetches and sets actionLog on showActionLog', async () => {
    const mockEntries = [{ turnNumber: 1, playerIdx: 0, actionType: 'hit', detail: 'test', cards: [] }];
    vi.mocked(actionLogApi.blackjack).mockResolvedValueOnce({ entries: mockEntries });

    const { result } = renderHook(() => useActionLog('blackjack'));

    await act(async () => {
      await result.current.showActionLog();
    });

    expect(actionLogApi.blackjack).toHaveBeenCalledTimes(1);
    expect(result.current.actionLog).toEqual(mockEntries);
  });

  it('clears actionLog on hideActionLog', async () => {
    const mockEntries = [{ turnNumber: 1, playerIdx: 0, actionType: 'hit', detail: 'test', cards: [] }];
    vi.mocked(actionLogApi.blackjack).mockResolvedValueOnce({ entries: mockEntries });

    const { result } = renderHook(() => useActionLog('blackjack'));

    await act(async () => {
      await result.current.showActionLog();
    });
    expect(result.current.actionLog).toEqual(mockEntries);

    act(() => {
      result.current.hideActionLog();
    });
    expect(result.current.actionLog).toBeNull();
  });

  // The log fetch can land after the player closes the page; writing state then is a
  // no-op until the environment is torn down, where it throws (#4447).
  it('does not store entries that arrive after unmount', async () => {
    let resolveLog: ((value: { entries: [] }) => void) | undefined;
    vi.mocked(actionLogApi.blackjack).mockImplementation(
      () =>
        new Promise((resolve) => {
          resolveLog = resolve;
        }) as ReturnType<typeof actionLogApi.blackjack>,
    );
    const { result, unmount } = renderHook(() => useActionLog('blackjack'));
    act(() => {
      void result.current.showActionLog();
    });
    unmount();
    resolveLog?.({ entries: [] });
    await new Promise((resolve) => {
      setTimeout(resolve, 0);
    });
    // Nothing to assert on the unmounted hook's state directly; the contract is that
    // resolving late does not throw and leaves the last rendered value untouched.
    expect(result.current.actionLog).toBeNull();
  });
});
