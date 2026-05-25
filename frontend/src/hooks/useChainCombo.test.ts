import { renderHook } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { useChainCombo } from './useChainCombo';

describe('useChainCombo', () => {
  it('returns 0 when state is not yet loaded', () => {
    const { result } = renderHook(() => useChainCombo(undefined, undefined));
    expect(result.current).toBe(0);
  });

  it('increments while moveCount rises and stockCount holds', () => {
    const { result, rerender } = renderHook(({ m, s }) => useChainCombo(m, s), {
      initialProps: { m: 0, s: 24 },
    });
    expect(result.current).toBe(0);
    rerender({ m: 1, s: 24 });
    expect(result.current).toBe(1);
    rerender({ m: 2, s: 24 });
    expect(result.current).toBe(2);
  });

  it('resets when stockCount changes (player drew from stock)', () => {
    const { result, rerender } = renderHook(({ m, s }) => useChainCombo(m, s), {
      initialProps: { m: 0, s: 24 },
    });
    rerender({ m: 1, s: 24 });
    rerender({ m: 2, s: 24 });
    expect(result.current).toBe(2);
    rerender({ m: 3, s: 23 }); // drew a card
    expect(result.current).toBe(0);
  });

  it('resets when moveCount drops to 0 (game reset)', () => {
    const { result, rerender } = renderHook(({ m, s }) => useChainCombo(m, s), {
      initialProps: { m: 0, s: 24 },
    });
    rerender({ m: 1, s: 24 });
    rerender({ m: 2, s: 24 });
    expect(result.current).toBe(2);
    rerender({ m: 0, s: 24 }); // reset
    expect(result.current).toBe(0);
  });

  it('does not increment if moveCount stays the same', () => {
    const { result, rerender } = renderHook(({ m, s }) => useChainCombo(m, s), {
      initialProps: { m: 0, s: 24 },
    });
    rerender({ m: 1, s: 24 });
    expect(result.current).toBe(1);
    rerender({ m: 1, s: 24 });
    expect(result.current).toBe(1);
  });

  it('resets when moveCount decreases (undo)', () => {
    const { result, rerender } = renderHook(({ m, s }) => useChainCombo(m, s), {
      initialProps: { m: 0, s: 24 },
    });
    rerender({ m: 1, s: 24 });
    rerender({ m: 2, s: 24 });
    expect(result.current).toBe(2);
    rerender({ m: 1, s: 24 }); // undo
    expect(result.current).toBe(0);
    rerender({ m: 2, s: 24 }); // redo would otherwise inflate; combo starts fresh
    expect(result.current).toBe(1);
  });
});
