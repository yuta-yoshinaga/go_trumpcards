import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { describe, expect, it, vi } from 'vitest';
import { useTablanetGame } from './useTablanetGame';

vi.mock('../api/gameApi', () => ({ tablanetApi: { exec: vi.fn() } }));

function makeWrapper() {
  const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) => createElement(QueryClientProvider, { client }, children);
}

/** Hand selection changes must invalidate table selections made for the prior card. */
describe('useTablanetGame', () => {
  it('clears table selection when selecting or deselecting a hand card', () => {
    const { result } = renderHook(() => useTablanetGame(), { wrapper: makeWrapper() });

    act(() => {
      result.current.setHandIndex(0);
      result.current.toggleTable(2);
    });
    expect(result.current.tableIndices).toEqual([2]);

    act(() => result.current.setHandIndex(1));
    expect(result.current.handIndex).toBe(1);
    expect(result.current.tableIndices).toEqual([]);

    act(() => {
      result.current.toggleTable(2);
      result.current.setHandIndex(1);
    });
    expect(result.current.handIndex).toBeNull();
    expect(result.current.tableIndices).toEqual([]);
  });
});
