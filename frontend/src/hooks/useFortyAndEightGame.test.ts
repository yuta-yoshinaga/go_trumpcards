import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fortyAndEightApi } from '../api/gameApi';
import type { FortyAndEightResponse } from '../types/card';
import { useFortyAndEightGame } from './useFortyAndEightGame';

vi.mock('../api/gameApi', () => ({ fortyAndEightApi: { exec: vi.fn() }, actionLogApi: { fortyandeight: vi.fn() } }));
const mockExec = vi.mocked(fortyAndEightApi.exec);
const state = {} as FortyAndEightResponse;
const wrapper = ({ children }: { children: ReactNode }) =>
  createElement(QueryClientProvider, { client: new QueryClient() }, children);

describe('useFortyAndEightGame', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(state);
  });
  it('resets on mount and forwards draw, redeal, and undo escape', async () => {
    const { result } = renderHook(() => useFortyAndEightGame(), { wrapper });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(mockExec).toHaveBeenCalledTimes(1);
    mockExec.mockClear();
    act(() => result.current.handleDraw());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
    mockExec.mockClear();
    act(() => result.current.handleRedeal());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('redeal'));
    mockExec.mockClear();
    act(() => result.current.handleUndoEscape(3));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, 3));
  });
  it('clears selection on undo', async () => {
    const { result } = renderHook(() => useFortyAndEightGame(), { wrapper });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 0 }));
    act(() => result.current.handleUndo());
    expect(result.current.selectedSource).toBeNull();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });
});
