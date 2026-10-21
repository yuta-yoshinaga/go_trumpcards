import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { penguinApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { PenguinResponse } from '../types/card';
import { usePenguinGame } from './usePenguinGame';

vi.mock('../api/gameApi', () => ({ penguinApi: { exec: vi.fn() }, actionLogApi: { penguin: vi.fn() } }));

const mockExec = vi.mocked(penguinApi.exec);
const response = (legalTargets: PenguinResponse['legalTargets']) => ({ legalTargets }) as PenguinResponse;
const wrapper = ({ children }: { children: ReactNode }) =>
  createElement(QueryClientProvider, { client: new QueryClient() }, children);

describe('usePenguinGame foundation shortcut', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(response({ 'tableau-2-4': { tableau: [], foundation: [6] } }));
  });

  it('sends the server foundation destination', async () => {
    const { result } = renderHook(() => usePenguinGame(), { wrapper });
    await waitFor(() => expect(result.current.state?.legalTargets).toBeDefined());
    act(() => result.current.handleFoundationShortcut({ zone: 'tableau', col: 2, cardIndex: 4 }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'move',
        { zone: 'tableau', col: 2, cardIndex: 4 },
        { zone: 'foundation', col: 6 },
      ),
    );
  });

  it('sends the server foundation destination for a freecell source', async () => {
    mockExec.mockResolvedValue(response({ 'freecell-1': { tableau: [], foundation: [2] } }));
    const { result } = renderHook(() => usePenguinGame(), { wrapper });
    await waitFor(() => expect(result.current.state?.legalTargets).toBeDefined());
    act(() => result.current.handleFoundationShortcut({ zone: 'freecell', cell: 1 }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'freecell', cell: 1 }, { zone: 'foundation', col: 2 }),
    );
  });

  it('sends the server foundation destination for a tableau source', async () => {
    mockExec.mockResolvedValue(response({ 'tableau-3-4': { tableau: [], foundation: [0] } }));
    const { result } = renderHook(() => usePenguinGame(), { wrapper });
    await waitFor(() => expect(result.current.state?.legalTargets).toBeDefined());
    act(() => result.current.handleFoundationShortcut({ zone: 'tableau', col: 3, cardIndex: 4 }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'move',
        { zone: 'tableau', col: 3, cardIndex: 4 },
        { zone: 'foundation', col: 0 },
      ),
    );
  });

  it('does nothing when the source key is missing', async () => {
    mockExec.mockResolvedValue(response({}));
    const { result } = renderHook(() => usePenguinGame(), { wrapper });
    await waitFor(() => expect(result.current.state?.legalTargets).toBeDefined());
    mockExec.mockClear();
    act(() => result.current.handleFoundationShortcut({ zone: 'tableau', col: 2, cardIndex: 4 }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('does nothing when a tableau source has no card index', async () => {
    const { result } = renderHook(() => usePenguinGame(), { wrapper });
    await waitFor(() => expect(result.current.state?.legalTargets).toBeDefined());
    mockExec.mockClear();
    act(() => result.current.handleFoundationShortcut({ zone: 'tableau', col: 2 }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('does nothing when a freecell source has no cell', async () => {
    const { result } = renderHook(() => usePenguinGame(), { wrapper });
    await waitFor(() => expect(result.current.state?.legalTargets).toBeDefined());
    mockExec.mockClear();
    act(() => result.current.handleFoundationShortcut({ zone: 'freecell' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('does nothing while the initial game response is pending', async () => {
    let resolveReset!: (value: PenguinResponse) => void;
    mockExec.mockReturnValue(
      new Promise((resolve) => {
        resolveReset = resolve;
      }),
    );
    const { result } = renderHook(() => usePenguinGame(), { wrapper });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(result.current.state).toBeNull();
    mockExec.mockClear();
    act(() => result.current.handleFoundationShortcut({ zone: 'tableau', col: 3, cardIndex: 4 }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
    await act(async () => resolveReset(response({})));
  });
});
