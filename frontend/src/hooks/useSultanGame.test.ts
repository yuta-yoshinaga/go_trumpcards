import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { sultanApi } from '../api/gameApi';
import type { SultanResponse } from '../types/card';
import { useSultanGame } from './useSultanGame';

vi.mock('../api/gameApi', () => ({ sultanApi: { exec: vi.fn() }, actionLogApi: { sultan: vi.fn() } }));
const mockExec = vi.mocked(sultanApi.exec);
const state = {} as SultanResponse;
const wrapper = ({ children }: { children: ReactNode }) =>
  createElement(QueryClientProvider, { client: new QueryClient() }, children);

describe('useSultanGame', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(state);
  });
  it('resets on mount and forwards draw, redeal, play, and undo escape', async () => {
    const { result } = renderHook(() => useSultanGame(), { wrapper });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(mockExec).toHaveBeenCalledTimes(1);
    mockExec.mockClear();
    act(() => result.current.handleDraw());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
    mockExec.mockClear();
    act(() => result.current.handleRedeal());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('redeal'));
    mockExec.mockClear();
    act(() => result.current.handlePlay({ zone: 'waste' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste' }));
    mockExec.mockClear();
    act(() => result.current.handleUndoEscape(2));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, 2));
  });
});
