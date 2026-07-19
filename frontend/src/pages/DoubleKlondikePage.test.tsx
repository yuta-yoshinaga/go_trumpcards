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

  it('renders face-down tableau cards as a card-back image, not "##"', async () => {
    renderWithProviders(<DoubleKlondikePage />);
    // makeState seeds a face-down card at tableau[1][0].
    await waitFor(() => expect(screen.getByTestId('column-1')).toBeInTheDocument());
    // Assert by the locale-independent card-back image src.
    const cardBacks = screen.queryAllByRole('img').filter((img) => img.getAttribute('src') === '/images/z01.png');
    expect(cardBacks.length).toBeGreaterThan(0);
    expect(screen.queryByText('##')).not.toBeInTheDocument();
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

  it('fans out the top 3 waste cards, keeping only the top selectable', async () => {
    mockExec.mockResolvedValue(
      makeState({ waste: [card('CLOVER', 2), card('SPADE', 3), card('HEART', 4), card('DIAMOND', 5)] }),
    );
    renderWithProviders(<DoubleKlondikePage />);
    // The top card stays interactive as the `waste` source button.
    await screen.findByTestId('waste');
    // The two cards beneath it are shown for visibility only (not selectable).
    expect(screen.getByTestId('waste-under-0')).toBeInTheDocument();
    expect(screen.getByTestId('waste-under-1')).toBeInTheDocument();
    // Only 3 of the 4 waste cards are rendered (the fan is capped at 3).
    expect(screen.queryByTestId('waste-under-2')).not.toBeInTheDocument();
    // The under-cards are plain divs, not buttons.
    expect(screen.getAllByTestId('waste', { exact: true }).length).toBe(1);
  });

  it('shows fewer waste cards when fewer than 3 are present', async () => {
    mockExec.mockResolvedValue(makeState({ waste: [card('CLOVER', 2), card('SPADE', 3)] }));
    renderWithProviders(<DoubleKlondikePage />);
    await screen.findByTestId('waste');
    expect(screen.getByTestId('waste-under-0')).toBeInTheDocument();
    expect(screen.queryByTestId('waste-under-1')).not.toBeInTheDocument();
  });

  it('shows an empty waste placeholder when the waste is empty', async () => {
    mockExec.mockResolvedValue(makeState({ waste: [] }));
    renderWithProviders(<DoubleKlondikePage />);
    await waitFor(() => expect(screen.getByTestId('waste-empty')).toBeInTheDocument());
    expect(screen.queryByTestId('waste')).not.toBeInTheDocument();
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
