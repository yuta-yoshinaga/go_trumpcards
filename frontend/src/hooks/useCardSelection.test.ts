import { describe, expect, it } from 'bun:test';
import { act, renderHook } from '@testing-library/react';
import { useCardSelection } from './useCardSelection';

describe('useCardSelection', () => {
  it('starts with empty selection by default', () => {
    const { result } = renderHook(() => useCardSelection());
    expect(result.current.selected).toEqual([]);
  });

  it('starts with custom initial selection', () => {
    const { result } = renderHook(() => useCardSelection([1, 3]));
    expect(result.current.selected).toEqual([1, 3]);
  });

  it('toggle adds an index', () => {
    const { result } = renderHook(() => useCardSelection());
    act(() => result.current.toggle(2));
    expect(result.current.selected).toEqual([2]);
  });

  it('toggle removes an already-selected index', () => {
    const { result } = renderHook(() => useCardSelection([0, 2]));
    act(() => result.current.toggle(2));
    expect(result.current.selected).toEqual([0]);
  });

  it('clear resets to empty', () => {
    const { result } = renderHook(() => useCardSelection());
    act(() => {
      result.current.toggle(0);
      result.current.toggle(3);
    });
    expect(result.current.selected).toEqual([0, 3]);

    act(() => result.current.clear());
    expect(result.current.selected).toEqual([]);
  });

  it('setSelected allows custom state updates', () => {
    const { result } = renderHook(() => useCardSelection());
    act(() => result.current.setSelected([5, 7]));
    expect(result.current.selected).toEqual([5, 7]);
  });
});
