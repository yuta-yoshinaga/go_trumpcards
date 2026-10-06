import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fortyThievesApi } from '../api/gameApi';
import type { FortyThievesResponse } from '../types/card';
import { useFortyThievesGame } from './useFortyThievesGame';

vi.mock('../api/gameApi', () => ({ fortyThievesApi: { exec: vi.fn() }, actionLogApi: { fortythieves: vi.fn() } }));
const mockExec = vi.mocked(fortyThievesApi.exec);
const state = {} as FortyThievesResponse;
const wrapper = ({ children }: { children: ReactNode }) =>
  createElement(QueryClientProvider, { client: new QueryClient() }, children);

describe('useFortyThievesGame', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(state);
  });
  it('resets on mount and forwards draw and undo escape', async () => {
    const { result } = renderHook(() => useFortyThievesGame(), { wrapper });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(mockExec).toHaveBeenCalledTimes(1);
    mockExec.mockClear();
    act(() => result.current.handleDraw());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
    mockExec.mockClear();
    act(() => result.current.handleUndoEscape(4));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, 4));
  });
  it('clears selection on give up', async () => {
    const { result } = renderHook(() => useFortyThievesGame(), { wrapper });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 0 }));
    act(() => result.current.handleGiveUp());
    expect(result.current.selectedSource).toBeNull();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });
});
