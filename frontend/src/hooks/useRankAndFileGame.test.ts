import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { rankAndFileApi } from '../api/gameApi';
import type { RankAndFileResponse } from '../types/card';
import { useRankAndFileGame } from './useRankAndFileGame';

vi.mock('../api/gameApi', () => ({ rankAndFileApi: { exec: vi.fn() }, actionLogApi: { rankandfile: vi.fn() } }));
const mockExec = vi.mocked(rankAndFileApi.exec);
const state = {} as RankAndFileResponse;
const wrapper = ({ children }: { children: ReactNode }) =>
  createElement(QueryClientProvider, { client: new QueryClient() }, children);

describe('useRankAndFileGame', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(state);
  });
  it('resets on mount and forwards draw and undo escape', async () => {
    const { result } = renderHook(() => useRankAndFileGame(), { wrapper });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(mockExec).toHaveBeenCalledTimes(1);
    mockExec.mockClear();
    act(() => result.current.handleDraw());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
    mockExec.mockClear();
    act(() => result.current.handleUndoEscape(2));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, 2));
  });
  it('clears selection on undo', async () => {
    const { result } = renderHook(() => useRankAndFileGame(), { wrapper });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 0 }));
    act(() => result.current.handleUndo());
    expect(result.current.selectedSource).toBeNull();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });
});
