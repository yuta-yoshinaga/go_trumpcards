import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { stHelenaApi } from '../api/gameApi';
import type { StHelenaResponse } from '../types/card';
import { useStHelenaGame } from './useStHelenaGame';

vi.mock('../api/gameApi', () => ({ stHelenaApi: { exec: vi.fn() }, actionLogApi: { sthelena: vi.fn() } }));
const mockExec = vi.mocked(stHelenaApi.exec);
const state = {} as StHelenaResponse;
const wrapper = ({ children }: { children: ReactNode }) =>
  createElement(QueryClientProvider, { client: new QueryClient() }, children);

describe('useStHelenaGame', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(state);
  });
  it('resets on mount and forwards redeal and undo escape', async () => {
    const { result } = renderHook(() => useStHelenaGame(), { wrapper });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(mockExec).toHaveBeenCalledTimes(1);
    mockExec.mockClear();
    act(() => result.current.handleRedeal());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('redeal'));
    mockExec.mockClear();
    act(() => result.current.handleUndoEscape(3));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, 3));
  });
  it('clears selection on give up', async () => {
    const { result } = renderHook(() => useStHelenaGame(), { wrapper });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 0 }));
    act(() => result.current.handleGiveUp());
    expect(result.current.selectedSource).toBeNull();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });
});
