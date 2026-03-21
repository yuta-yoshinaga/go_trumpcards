import { afterEach, beforeEach, describe, expect, it, vi } from 'bun:test';
import { act, renderHook } from '@testing-library/react';
import { actionLogApi } from '../api/gameApi';
import { asMocked } from '../test/viCompat';
import { useActionLog } from './useActionLog';

describe('useActionLog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(actionLogApi, 'blackjack').mockImplementation(vi.fn());
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('initializes with null actionLog', () => {
    const { result } = renderHook(() => useActionLog('blackjack'));
    expect(result.current.actionLog).toBeNull();
  });

  it('fetches and sets actionLog on showActionLog', async () => {
    const mockEntries = [{ turnNumber: 1, playerIdx: 0, actionType: 'hit', detail: 'test', cards: [] }];
    asMocked(actionLogApi.blackjack).mockResolvedValueOnce({ entries: mockEntries });

    const { result } = renderHook(() => useActionLog('blackjack'));

    await act(async () => {
      await result.current.showActionLog();
    });

    expect(actionLogApi.blackjack).toHaveBeenCalledTimes(1);
    expect(result.current.actionLog).toEqual(mockEntries);
  });

  it('clears actionLog on hideActionLog', async () => {
    const mockEntries = [{ turnNumber: 1, playerIdx: 0, actionType: 'hit', detail: 'test', cards: [] }];
    asMocked(actionLogApi.blackjack).mockResolvedValueOnce({ entries: mockEntries });

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
});
