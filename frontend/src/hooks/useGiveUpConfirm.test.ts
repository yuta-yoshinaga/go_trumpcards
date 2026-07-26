import { act, renderHook } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { useGiveUpConfirm } from './useGiveUpConfirm';

describe('useGiveUpConfirm', () => {
  it('routes the give-up action through requestGiveUpConfirm', () => {
    const giveUp = vi.fn();
    // requestGiveUpConfirm stores the action; a real dialog runs it on confirm.
    let stored: (() => void) | null = null;
    const requestGiveUpConfirm = vi.fn((action: () => void) => {
      stored = action;
    });

    const { result } = renderHook(() => useGiveUpConfirm(giveUp, requestGiveUpConfirm));

    // Invoking the handler opens the dialog but does NOT fire give-up yet.
    act(() => result.current());
    expect(requestGiveUpConfirm).toHaveBeenCalledTimes(1);
    expect(giveUp).not.toHaveBeenCalled();

    // Confirming runs the stored give-up action.
    act(() => stored?.());
    expect(giveUp).toHaveBeenCalledTimes(1);
  });

  it('returns a stable callback across renders while deps are unchanged', () => {
    const giveUp = vi.fn();
    const requestGiveUpConfirm = vi.fn();
    const { result, rerender } = renderHook(() => useGiveUpConfirm(giveUp, requestGiveUpConfirm));
    const first = result.current;
    rerender();
    expect(result.current).toBe(first);
  });
});
