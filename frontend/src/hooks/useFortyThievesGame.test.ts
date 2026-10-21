import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fortyThievesApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { FortyThievesResponse } from '../types/card';
import { useFortyThievesGame } from './useFortyThievesGame';

vi.mock('../api/gameApi', () => ({ fortyThievesApi: { exec: vi.fn() }, actionLogApi: { fortythieves: vi.fn() } }));
const mockExec = vi.mocked(fortyThievesApi.exec);
const state = { legalTargets: {} } as FortyThievesResponse;
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

  it('toggles off an already selected source and replaces a different source', async () => {
    const { result } = renderHook(() => useFortyThievesGame(), { wrapper });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const source = { zone: 'tableau', col: 1, cardIndex: 2 };
    act(() => result.current.handleSelectSource(source));
    expect(result.current.selectedSource).toEqual(source);
    act(() => result.current.handleSelectSource({ ...source, col: 3 }));
    expect(result.current.selectedSource).toEqual({ ...source, col: 3 });
    act(() => result.current.handleSelectSource({ ...source, col: 3 }));
    expect(result.current.selectedSource).toBeNull();
  });

  it('sends a move to the server supplied foundation target', async () => {
    mockExec.mockResolvedValue({
      legalTargets: { 'tableau-3': { tableau: [], foundation: [5] } },
    } as unknown as FortyThievesResponse);
    const { result } = renderHook(() => useFortyThievesGame(), { wrapper });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    await waitFor(() => expect(result.current.state?.legalTargets['tableau-3']).toBeDefined());
    mockExec.mockClear();
    act(() => result.current.handleFoundationShortcut({ zone: 'tableau', col: 3 }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'tableau', col: 3 }, { zone: 'foundation', col: 5 }),
    );
  });

  it('follows server targets even when they disagree with client card rules', async () => {
    mockExec.mockResolvedValue({
      legalTargets: { waste: { tableau: [], foundation: [6] } },
    } as unknown as FortyThievesResponse);
    const { result } = renderHook(() => useFortyThievesGame(), { wrapper });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    await waitFor(() => expect(result.current.state?.legalTargets.waste).toBeDefined());
    mockExec.mockClear();
    act(() => result.current.handleFoundationShortcut({ zone: 'waste' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste' }, { zone: 'foundation', col: 6 }),
    );
  });

  it('does nothing when the source key is missing', async () => {
    const { result } = renderHook(() => useFortyThievesGame(), { wrapper });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    act(() => result.current.handleFoundationShortcut({ zone: 'waste' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('does nothing when a tableau source has no column index', async () => {
    const { result } = renderHook(() => useFortyThievesGame(), { wrapper });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    act(() => result.current.handleFoundationShortcut({ zone: 'tableau' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});
