import { describe, expect, it, vi } from 'bun:test';
import { act, renderHook } from '@testing-library/react';
import { useConfirmDialog } from './useConfirmDialog';

describe('useConfirmDialog', () => {
  it('starts closed', () => {
    const { result } = renderHook(() => useConfirmDialog());
    expect(result.current.isOpen).toBe(false);
  });

  it('opens on requestConfirm', () => {
    const { result } = renderHook(() => useConfirmDialog());
    act(() => result.current.requestConfirm(vi.fn()));
    expect(result.current.isOpen).toBe(true);
  });

  it('confirm closes and calls callback', async () => {
    const callback = vi.fn();
    const { result } = renderHook(() => useConfirmDialog());
    act(() => result.current.requestConfirm(callback));
    await act(async () => result.current.confirm());
    expect(result.current.isOpen).toBe(false);
    expect(callback).toHaveBeenCalledTimes(1);
  });

  it('cancel closes without calling callback', () => {
    const callback = vi.fn();
    const { result } = renderHook(() => useConfirmDialog());
    act(() => result.current.requestConfirm(callback));
    act(() => result.current.cancel());
    expect(result.current.isOpen).toBe(false);
    expect(callback).not.toHaveBeenCalled();
  });

  it('confirm does not call callback after cancel', async () => {
    const callback = vi.fn();
    const { result } = renderHook(() => useConfirmDialog());
    act(() => result.current.requestConfirm(callback));
    act(() => result.current.cancel());
    await act(async () => result.current.confirm());
    expect(callback).not.toHaveBeenCalled();
  });
});
