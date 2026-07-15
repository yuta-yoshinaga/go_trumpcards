import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { simplesimonApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, SimpleSimonResponse } from '../types/card';
import { SimpleSimonPage } from './SimpleSimonPage';

vi.mock('../api/gameApi', () => ({
  simplesimonApi: { exec: vi.fn() },
  actionLogApi: { simplesimon: vi.fn() },
}));

const mockExec = vi.mocked(simplesimonApi.exec);

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<SimpleSimonResponse> = {}): SimpleSimonResponse {
  const columns: Card[][] = Array.from({ length: 10 }, () => []);
  columns[0] = [card('SPADE', 9)];
  columns[1] = [card('SPADE', 8)];
  return {
    columns,
    completedSuits: 0,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    message: '',
    ...overrides,
  };
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('SimpleSimonPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<SimpleSimonPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<SimpleSimonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders the 10 columns', async () => {
    renderWithProviders(<SimpleSimonPage />);
    await waitFor(() => expect(screen.getByTestId('column-0')).toBeInTheDocument());
    expect(screen.getByTestId('column-9')).toBeInTheDocument();
  });

  it('selects a card then moves the run to another column', async () => {
    renderWithProviders(<SimpleSimonPage />);
    const srcCard = await screen.findByTestId('card-1-0');
    fireEvent.click(srcCard);
    mockExec.mockClear();
    // Click a card in column 0 → move col1 onto col0.
    fireEvent.click(screen.getByTestId('card-0-0'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('m', { fromCol: 1, cardIndex: 0, toCol: 0 }));
  });

  it('labels cards with name+position and reflects selection with aria-pressed', async () => {
    renderWithProviders(<SimpleSimonPage />);
    // column[1] is ♠8 at the top → "♠ 8（列2・上から1枚目）".
    const src = await screen.findByTestId('card-1-0');
    expect(src).toHaveAttribute('aria-label', '♠ 8（列2・上から1枚目）');
    expect(src).toHaveAttribute('aria-pressed', 'false');
    fireEvent.click(src);
    expect(screen.getByTestId('card-1-0')).toHaveAttribute('aria-pressed', 'true');
    // Empty column drop targets are named too.
    expect(screen.getByTestId('column-2-drop')).toHaveAttribute('aria-label', '列3（空）');
    // The selection guidance is a live region.
    expect(screen.getByTestId('ss-guidance')).toHaveAttribute('role', 'status');
  });

  it('moves onto an empty column', async () => {
    renderWithProviders(<SimpleSimonPage />);
    const srcCard = await screen.findByTestId('card-0-0');
    fireEvent.click(srcCard);
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('column-5-drop'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('m', { fromCol: 0, cardIndex: 0, toCol: 5 }));
  });

  it('undoes, hints and gives up', async () => {
    mockExec.mockResolvedValue(makeState({ canUndo: true }));
    renderWithProviders(<SimpleSimonPage />);
    await screen.findByTestId('hint-button');
    for (const [testid, cmd] of [
      ['undo-button', 'u'],
      ['hint-button', 'hint'],
      ['giveup-button', 'g'],
    ] as const) {
      mockExec.mockClear();
      fireEvent.click(screen.getByTestId(testid));
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith(cmd));
    }
  });

  it('hides controls at game over', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2 }));
    renderWithProviders(<SimpleSimonPage />);
    await waitFor(() => expect(screen.getByTestId('column-0')).toBeInTheDocument());
    expect(screen.queryByTestId('hint-button')).not.toBeInTheDocument();
  });
});
