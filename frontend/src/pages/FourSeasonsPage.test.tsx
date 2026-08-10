import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fourseasonsApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, FourSeasonsResponse } from '../types/card';
import { FourSeasonsPage, fourseasonsNextRank, fourseasonsTableauNextRank } from './FourSeasonsPage';

vi.mock('../api/gameApi', () => ({
  fourseasonsApi: { exec: vi.fn() },
  actionLogApi: { fourseasons: vi.fn() },
}));

const mockExec = vi.mocked(fourseasonsApi.exec);
const card = (design: CardDesign, value: number): Card => ({ design, value });

/** Base rank 7: the foundations start at 7, not at Ace. */
function makeState(overrides: Partial<FourSeasonsResponse> = {}): FourSeasonsResponse {
  return {
    tableau: [[card('HEART', 12)], [card('CLOVER', 3)], [card('DIAMOND', 9)], [], [card('SPADE', 13)]],
    foundation: [[card('SPADE', 7)], [], [], []],
    stockCount: 44,
    waste: [card('HEART', 6)],
    baseRank: 7,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    message: '',
    messageCode: 'fourseasons.playing',
    ...overrides,
  };
}

describe('fourseasonsNextRank', () => {
  it('opens an empty foundation at the base rank, not at Ace', () => {
    expect(fourseasonsNextRank(undefined, 0, 7)).toBe(7);
    expect(fourseasonsNextRank(undefined, 0, 1)).toBe(1);
  });

  it('ascends and wraps King to Ace', () => {
    expect(fourseasonsNextRank(7, 1, 7)).toBe(8);
    expect(fourseasonsNextRank(13, 7, 7)).toBe(1);
  });

  it('returns null once the pile is complete', () => {
    expect(fourseasonsNextRank(6, 13, 7)).toBeNull();
  });
});

describe('fourseasonsTableauNextRank', () => {
  it('descends and wraps Ace to King', () => {
    expect(fourseasonsTableauNextRank(9)).toBe(8);
    expect(fourseasonsTableauNextRank(1)).toBe(13);
  });

  it('returns null for an empty pile, which accepts anything', () => {
    expect(fourseasonsTableauNextRank(undefined)).toBeNull();
  });
});

describe('FourSeasonsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<FourSeasonsPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // The base rank drives every placement rule, so it has to be on screen.
  it('shows the deal base rank', async () => {
    renderWithProviders(<FourSeasonsPage />);
    await waitFor(() => expect(screen.getByTestId('fs-base-rank')).toHaveTextContent('7'));
  });

  it('shows the next required rank per foundation, from the base rank', async () => {
    renderWithProviders(<FourSeasonsPage />);
    // F0 holds the 7, so it wants an 8; the empty corners still want the base 7.
    await waitFor(() => expect(screen.getByTestId('fs-foundation-next-0')).toHaveTextContent('8'));
    expect(screen.getByTestId('fs-foundation-next-1')).toHaveTextContent('7');
  });

  it('draws from the stock', async () => {
    renderWithProviders(<FourSeasonsPage />);
    await waitFor(() => expect(screen.getByTestId('fs-draw-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('fs-draw-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('disables the draw button once the stock is empty', async () => {
    mockExec.mockResolvedValue(makeState({ stockCount: 0 }));
    renderWithProviders(<FourSeasonsPage />);
    await waitFor(() => expect(screen.getByTestId('fs-draw-button')).toBeDisabled());
  });

  it('moves the waste card to a corner once selected', async () => {
    renderWithProviders(<FourSeasonsPage />);
    await waitFor(() => expect(screen.getByTestId('fs-waste-button')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('fs-waste-button'));
    await waitFor(() => expect(screen.getByTestId('fs-waste-button')).toHaveAttribute('aria-pressed', 'true'));

    fireEvent.click(screen.getByTestId('fs-foundation-0'));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste' }, { zone: 'foundation', idx: 0 }),
    );
  });

  it('moves a cross pile top to another cross pile', async () => {
    renderWithProviders(<FourSeasonsPage />);
    await waitFor(() => expect(screen.getByTestId('fs-tableau-0')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('fs-tableau-0'));
    await waitFor(() => expect(screen.getByTestId('fs-tableau-0')).toHaveAttribute('aria-pressed', 'true'));

    fireEvent.click(screen.getByTestId('fs-tableau-1'));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'tableau', idx: 0 }, { zone: 'tableau', idx: 1 }),
    );
  });

  // An empty cross space takes any card, so it must be clickable as a target
  // even though it holds nothing to select.
  it('accepts a card onto an empty cross space', async () => {
    renderWithProviders(<FourSeasonsPage />);
    await waitFor(() => expect(screen.getByTestId('fs-waste-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('fs-waste-button'));
    await waitFor(() => expect(screen.getByTestId('fs-tableau-3')).toBeEnabled());

    fireEvent.click(screen.getByTestId('fs-tableau-3'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste' }, { zone: 'tableau', idx: 3 }));
  });

  it('does not move when nothing is selected', async () => {
    renderWithProviders(<FourSeasonsPage />);
    await waitFor(() => expect(screen.getByTestId('fs-foundation-0')).toBeDisabled());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('fs-foundation-0'));
    // Without this await the assertion passes whether or not a move fired (#4439).
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());

    // Negative control: the same click DOES move once a source is selected.
    fireEvent.click(screen.getByTestId('fs-waste-button'));
    await waitFor(() => expect(screen.getByTestId('fs-waste-button')).toHaveAttribute('aria-pressed', 'true'));
    fireEvent.click(screen.getByTestId('fs-foundation-0'));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste' }, { zone: 'foundation', idx: 0 }),
    );
  });

  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue(
      makeState({
        hint: { fromZone: 'tableau', fromIdx: 4, toZone: 'foundation', toIdx: 0 },
        messageCode: 'fourseasons.hintAvailable',
      }),
    );
    renderWithProviders(<FourSeasonsPage />);
    await waitFor(() => expect(screen.getByRole('status')).toHaveTextContent('ヒントがあります'));
  });

  // The other half of the gate: a passive hint must not surface the banner.
  it('hides the hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue(
      makeState({ hint: { fromZone: 'tableau', fromIdx: 4, toZone: 'foundation', toIdx: 0 } }),
    );
    renderWithProviders(<FourSeasonsPage />);
    await waitFor(() => expect(screen.getByTestId('fs-draw-button')).toBeInTheDocument());
    expect(screen.queryByRole('status')).not.toBeInTheDocument();
  });

  it('hides the playing controls once the game clears', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1 }));
    renderWithProviders(<FourSeasonsPage />);
    await waitFor(() => expect(screen.queryByTestId('fs-draw-button')).not.toBeInTheDocument());
  });
  it('requests a hint', async () => {
    renderWithProviders(<FourSeasonsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('undoes only when there is history', async () => {
    renderWithProviders(<FourSeasonsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled());

    mockExec.mockResolvedValue(makeState({ canUndo: true }));
    renderWithProviders(<FourSeasonsPage />);
    await waitFor(() => expect(screen.getAllByRole('button', { name: '元に戻す' })[1]).toBeEnabled());
    fireEvent.click(screen.getAllByRole('button', { name: '元に戻す' })[1]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  // Auto-complete is only offered when the server says a card can actually go,
  // so an idle board does not dangle a button that would just error.
  it('enables auto-complete only when a move to a corner exists', async () => {
    renderWithProviders(<FourSeasonsPage />);
    await waitFor(() => expect(screen.getByTestId('autocomplete-button')).toBeDisabled());

    mockExec.mockResolvedValue(
      makeState({ hint: { fromZone: 'tableau', fromIdx: 4, toZone: 'foundation', toIdx: 0 } }),
    );
    renderWithProviders(<FourSeasonsPage />);
    await waitFor(() => expect(screen.getAllByTestId('autocomplete-button')[1]).toBeEnabled());
    fireEvent.click(screen.getAllByTestId('autocomplete-button')[1]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('deselects via the cancel button', async () => {
    renderWithProviders(<FourSeasonsPage />);
    await waitFor(() => expect(screen.getByTestId('fs-waste-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('fs-waste-button'));
    await waitFor(() => expect(screen.getByRole('button', { name: 'キャンセル' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    await waitFor(() => expect(screen.getByTestId('fs-waste-button')).toHaveAttribute('aria-pressed', 'false'));
  });

  it('clicking the selected pile again deselects it', async () => {
    renderWithProviders(<FourSeasonsPage />);
    await waitFor(() => expect(screen.getByTestId('fs-tableau-0')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('fs-tableau-0'));
    await waitFor(() => expect(screen.getByTestId('fs-tableau-0')).toHaveAttribute('aria-pressed', 'true'));
    fireEvent.click(screen.getByTestId('fs-tableau-0'));
    await waitFor(() => expect(screen.getByTestId('fs-tableau-0')).toHaveAttribute('aria-pressed', 'false'));
  });

  it('moves a cross pile top to a corner', async () => {
    renderWithProviders(<FourSeasonsPage />);
    await waitFor(() => expect(screen.getByTestId('fs-tableau-4')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('fs-tableau-4'));
    await waitFor(() => expect(screen.getByTestId('fs-tableau-4')).toHaveAttribute('aria-pressed', 'true'));

    fireEvent.click(screen.getByTestId('fs-foundation-0'));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'tableau', idx: 4 }, { zone: 'foundation', idx: 0 }),
    );
  });

  it('gives up through the confirm dialog', async () => {
    renderWithProviders(<FourSeasonsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    await waitFor(() => expect(screen.getByRole('button', { name: '確認' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('hides the playing controls after a give-up', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2 }));
    renderWithProviders(<FourSeasonsPage />);
    await waitFor(() => expect(screen.queryByTestId('fs-draw-button')).not.toBeInTheDocument());
  });

  it('shows an error with a retry', async () => {
    mockExec.mockRejectedValue(new Error('boom'));
    renderWithProviders(<FourSeasonsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /再試行|retry/i })).toBeInTheDocument());
  });
});
