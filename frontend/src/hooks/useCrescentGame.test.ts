import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { crescentApi } from '../api/gameApi';
import type { CrescentResponse } from '../types/card';
import { useCrescentGame } from './useCrescentGame';

vi.mock('../api/gameApi', () => ({ crescentApi: { exec: vi.fn() }, actionLogApi: { crescent: vi.fn() } }));
const mockExec = vi.mocked(crescentApi.exec);
const state = {} as CrescentResponse;
const wrapper = ({ children }: { children: ReactNode }) =>
  createElement(
    QueryClientProvider,
    { client: new QueryClient({ defaultOptions: { mutations: { retry: false } } }) },
    children,
  );

describe('useCrescentGame', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(state);
  });
  it('resets on mount and forwards redeal and undo escape', async () => {
    const { result } = renderHook(() => useCrescentGame(), { wrapper });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(mockExec).toHaveBeenCalledTimes(1);
    mockExec.mockClear();
    act(() => result.current.handleRedeal());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('redeal'));
    mockExec.mockClear();
    act(() => result.current.handleUndoEscape(2));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, 2));
  });
  it('clears selection on give up', async () => {
    const { result } = renderHook(() => useCrescentGame(), { wrapper });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 0 }));
    act(() => result.current.handleGiveUp());
    expect(result.current.selectedSource).toBeNull();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });
});
