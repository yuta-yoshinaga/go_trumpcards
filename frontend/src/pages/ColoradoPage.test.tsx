import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { coloradoApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, ColoradoResponse } from '../types/card';
import { ColoradoPage, coloradoNextRank } from './ColoradoPage';

vi.mock('../api/gameApi', () => ({
  coloradoApi: { exec: vi.fn() },
  actionLogApi: { colorado: vi.fn() },
}));

const mockExec = vi.mocked(coloradoApi.exec);
const card = (design: CardDesign, value: number): Card => ({ design, value });

const TABLEAU_CNT = 20;

/** 20 piles of one card, with pile 3 emptied so the gap paths are reachable. */
function tableau(): Card[][] {
  const piles = Array.from({ length: TABLEAU_CNT }, (_, i) => [card('SPADE', ((i % 13) + 1) as number)]);
  piles[3] = [];
  return piles;
}

function makeState(overrides: Partial<ColoradoResponse> = {}): ColoradoResponse {
  return {
    tableau: tableau(),
    // F0 spades ascending holds the Ace; F4 spades descending holds the King.
    foundation: [[card('SPADE', 1)], [], [], [], [card('SPADE', 13)], [], [], []],
    foundationAscending: [true, true, true, true, false, false, false, false],
    stockCount: 71,
    waste: [card('HEART', 6)],
    phase: 0,
    moveCount: 0,
    canUndo: false,
    message: '',
    messageCode: 'colorado.playing',
    ...overrides,
  };
}

describe('coloradoNextRank', () => {
  it('opens an ascending foundation at the Ace and a descending one at the King', () => {
    expect(coloradoNextRank(undefined, 0, true)).toBe(1);
    expect(coloradoNextRank(undefined, 0, false)).toBe(13);
  });

  it('steps up or down depending on the direction', () => {
    expect(coloradoNextRank(1, 1, true)).toBe(2);
    expect(coloradoNextRank(13, 1, false)).toBe(12);
  });

  it('returns null once the pile is complete', () => {
    expect(coloradoNextRank(13, 13, true)).toBeNull();
    expect(coloradoNextRank(1, 13, false)).toBeNull();
  });
});

describe('ColoradoPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // Half the foundations run the other way, and the board is unreadable without
  // knowing which is which.
  it('shows the next rank for both directions', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-foundation-next-0')).toHaveTextContent('2'));
    expect(screen.getByTestId('co-foundation-next-4')).toHaveTextContent('Q');
    // The empty piles still show their opening rank, one per direction.
    expect(screen.getByTestId('co-foundation-next-1')).toHaveTextContent('A');
    expect(screen.getByTestId('co-foundation-next-5')).toHaveTextContent('K');
  });

  it('renders every tableau pile', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-tableau-0')).toBeInTheDocument());
    expect(screen.getByTestId(`co-tableau-${(TABLEAU_CNT - 1).toString()}`)).toBeInTheDocument();
  });

  it('draws from the stock', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-draw-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('co-draw-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('disables the draw button once the stock is empty', async () => {
    mockExec.mockResolvedValue(makeState({ stockCount: 0 }));
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-draw-button')).toBeDisabled());
  });

  it('moves the waste card to a foundation once selected', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-waste-button')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('co-waste-button'));
    await waitFor(() => expect(screen.getByTestId('co-waste-button')).toHaveAttribute('aria-pressed', 'true'));

    fireEvent.click(screen.getByTestId('co-foundation-0'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste' }, { zone: 'foundation' }));
  });

  // The whole game is this move: the waste goes onto ANY pile, suit and rank
  // irrelevant, burying whatever was there.
  it('buries the waste card on any tableau pile', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-waste-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('co-waste-button'));
    await waitFor(() => expect(screen.getByTestId('co-waste-button')).toHaveAttribute('aria-pressed', 'true'));

    fireEvent.click(screen.getByTestId('co-tableau-7'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste' }, { zone: 'tableau', idx: 7 }));
  });

  it('moves a tableau top to a foundation', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-tableau-0')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('co-tableau-0'));
    await waitFor(() => expect(screen.getByTestId('co-tableau-0')).toHaveAttribute('aria-pressed', 'true'));

    fireEvent.click(screen.getByTestId('co-foundation-0'));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'tableau', idx: 0 }, { zone: 'foundation' }),
    );
  });

  it('fills an empty pile straight from the stock', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-stock-fill-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('co-stock-fill-button'));
    await waitFor(() => expect(screen.getByTestId('co-stock-fill-button')).toHaveAttribute('aria-pressed', 'true'));

    fireEvent.click(screen.getByTestId('co-tableau-3'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'stock' }, { zone: 'tableau', idx: 3 }));
  });

  // The stock may only fill a gap, so clicking an occupied pile must do nothing.
  it('refuses to put the stock card onto an occupied pile', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-stock-fill-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('co-stock-fill-button'));
    await waitFor(() => expect(screen.getByTestId('co-stock-fill-button')).toHaveAttribute('aria-pressed', 'true'));

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('co-tableau-5'));
    // Without this await the assertion passes whether or not a move fired (#4439).
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());

    // Negative control: the empty pile DOES accept it.
    fireEvent.click(screen.getByTestId('co-tableau-3'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'stock' }, { zone: 'tableau', idx: 3 }));
  });

  it('hides the gap-fill button when no pile is empty', async () => {
    const piles = Array.from({ length: TABLEAU_CNT }, () => [card('SPADE', 7)]);
    mockExec.mockResolvedValue(makeState({ tableau: piles }));
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-draw-button')).toBeInTheDocument());
    expect(screen.queryByTestId('co-stock-fill-button')).not.toBeInTheDocument();
  });

  it('does not move when nothing is selected', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-foundation-0')).toBeDisabled());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('co-foundation-0'));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());

    // Negative control: the same click DOES move once a source is selected.
    fireEvent.click(screen.getByTestId('co-waste-button'));
    await waitFor(() => expect(screen.getByTestId('co-waste-button')).toHaveAttribute('aria-pressed', 'true'));
    fireEvent.click(screen.getByTestId('co-foundation-0'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste' }, { zone: 'foundation' }));
  });

  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue(
      makeState({
        hint: { fromZone: 'tableau', fromIdx: 4, toZone: 'foundation', toIdx: 0 },
        messageCode: 'colorado.hintAvailable',
      }),
    );
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByRole('status')).toHaveTextContent('ヒントがあります'));
  });

  // The other half of the gate: a passive hint must not surface the banner.
  it('hides the hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue(
      makeState({ hint: { fromZone: 'tableau', fromIdx: 4, toZone: 'foundation', toIdx: 0 } }),
    );
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-draw-button')).toBeInTheDocument());
    expect(screen.queryByRole('status')).not.toBeInTheDocument();
  });

  it('names the stock in the hint banner when the hint is to draw', async () => {
    mockExec.mockResolvedValue(
      makeState({
        hint: { fromZone: 'stock', fromIdx: -1, toZone: 'waste', toIdx: -1 },
        messageCode: 'colorado.hintAvailable',
      }),
    );
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByRole('status')).toHaveTextContent('山札'));
  });

  it('requests a hint', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('undoes only when there is history', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled());

    mockExec.mockResolvedValue(makeState({ canUndo: true }));
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getAllByRole('button', { name: '元に戻す' })[1]).toBeEnabled());
    fireEvent.click(screen.getAllByRole('button', { name: '元に戻す' })[1]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  // "Bury the waste somewhere" is always legal, so gating auto-complete on any
  // hint at all would leave the button lit for the whole game.
  it('enables auto-complete only for a move to a foundation', async () => {
    mockExec.mockResolvedValue(makeState({ hint: { fromZone: 'waste', fromIdx: -1, toZone: 'tableau', toIdx: 2 } }));
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('autocomplete-button')).toBeDisabled());

    mockExec.mockResolvedValue(
      makeState({ hint: { fromZone: 'tableau', fromIdx: 4, toZone: 'foundation', toIdx: 0 } }),
    );
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getAllByTestId('autocomplete-button')[1]).toBeEnabled());
    fireEvent.click(screen.getAllByTestId('autocomplete-button')[1]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('deselects via the cancel button', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-waste-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('co-waste-button'));
    await waitFor(() => expect(screen.getByRole('button', { name: 'キャンセル' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    await waitFor(() => expect(screen.getByTestId('co-waste-button')).toHaveAttribute('aria-pressed', 'false'));
  });

  it('clicking the selected pile again deselects it', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-tableau-0')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('co-tableau-0'));
    await waitFor(() => expect(screen.getByTestId('co-tableau-0')).toHaveAttribute('aria-pressed', 'true'));
    fireEvent.click(screen.getByTestId('co-tableau-0'));
    await waitFor(() => expect(screen.getByTestId('co-tableau-0')).toHaveAttribute('aria-pressed', 'false'));
  });

  it('hides the playing controls once the game clears', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1 }));
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.queryByTestId('co-draw-button')).not.toBeInTheDocument());
  });

  it('gives up through the confirm dialog', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    await waitFor(() => expect(screen.getByRole('button', { name: '確認' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('hides the playing controls after a give-up', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2 }));
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.queryByTestId('co-draw-button')).not.toBeInTheDocument());
  });

  it('shows an error with a retry', async () => {
    mockExec.mockRejectedValue(new Error('boom'));
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /再試行|retry/i })).toBeInTheDocument());
  });
});
