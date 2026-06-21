import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { doubleklondikeApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, DoubleKlondikeResponse } from '../types/card';
import { DoubleKlondikePage } from './DoubleKlondikePage';

vi.mock('../api/gameApi', () => ({
  doubleklondikeApi: { exec: vi.fn() },
  actionLogApi: { doubleklondike: vi.fn() },
}));

const mockExec = vi.mocked(doubleklondikeApi.exec);

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<DoubleKlondikeResponse> = {}): DoubleKlondikeResponse {
  const tableau: DoubleKlondikeResponse['tableau'] = Array.from({ length: 9 }, () => []);
  tableau[0] = [{ card: card('SPADE', 9), faceUp: true }];
  tableau[1] = [
    { card: null, faceUp: false },
    { card: card('HEART', 8), faceUp: true },
  ];
  return {
    tableau,
    stockCount: 59,
    waste: [card('DIAMOND', 1)],
    foundation: Array.from({ length: 8 }, () => []),
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('DoubleKlondikePage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<DoubleKlondikePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<DoubleKlondikePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders the 9 tableau columns and 8 foundations', async () => {
    renderWithProviders(<DoubleKlondikePage />);
    await waitFor(() => expect(screen.getByTestId('column-0')).toBeInTheDocument());
    expect(screen.getByTestId('column-8')).toBeInTheDocument();
    expect(screen.getByTestId('foundation-7')).toBeInTheDocument();
  });

  it('draws from the stock', async () => {
    renderWithProviders(<DoubleKlondikePage />);
    const stock = await screen.findByTestId('stock');
    mockExec.mockClear();
    fireEvent.click(stock);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('d'));
  });

  it('moves the waste card onto a tableau column', async () => {
    renderWithProviders(<DoubleKlondikePage />);
    fireEvent.click(await screen.findByTestId('waste'));
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('card-0-0'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('mwt', { col: 0 }));
  });

  it('moves the waste card onto a foundation', async () => {
    renderWithProviders(<DoubleKlondikePage />);
    fireEvent.click(await screen.findByTestId('waste'));
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('foundation-0'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('mwf'));
  });

  it('moves a tableau card to another column', async () => {
    renderWithProviders(<DoubleKlondikePage />);
    fireEvent.click(await screen.findByTestId('card-1-1'));
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('card-0-0'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('mtt', { fromCol: 1, cardIndex: 1, toCol: 0 }));
  });

  it('moves a tableau card to a foundation', async () => {
    renderWithProviders(<DoubleKlondikePage />);
    fireEvent.click(await screen.findByTestId('card-0-0'));
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('foundation-0'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('mtf', { col: 0 }));
  });

  it('moves onto an empty column', async () => {
    renderWithProviders(<DoubleKlondikePage />);
    fireEvent.click(await screen.findByTestId('card-0-0'));
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('column-5-drop'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('mtt', { fromCol: 0, cardIndex: 0, toCol: 5 }));
  });

  it('auto-completes, undoes, hints and gives up', async () => {
    mockExec.mockResolvedValue(makeState({ canUndo: true }));
    renderWithProviders(<DoubleKlondikePage />);
    await screen.findByTestId('hint-button');
    for (const [testid, cmd] of [
      ['auto-button', 'ac'],
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
    renderWithProviders(<DoubleKlondikePage />);
    await waitFor(() => expect(screen.getByTestId('column-0')).toBeInTheDocument());
    expect(screen.queryByTestId('hint-button')).not.toBeInTheDocument();
  });
});
