import { renderHook } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { useMountReset } from './useMountReset';

describe('useMountReset', () => {
  it('calls apiExec with "reset" once on mount', () => {
    const exec = vi.fn();
    renderHook(() => useMountReset(exec));
    expect(exec).toHaveBeenCalledTimes(1);
    expect(exec).toHaveBeenCalledWith('reset');
  });

  it('does not re-fire when apiExec identity changes between renders', () => {
    const first = vi.fn();
    const second = vi.fn();
    const { rerender } = renderHook(({ fn }: { fn: typeof first }) => useMountReset(fn), {
      initialProps: { fn: first },
    });
    rerender({ fn: second });
    expect(first).toHaveBeenCalledTimes(1);
    expect(second).not.toHaveBeenCalled();
  });
});
