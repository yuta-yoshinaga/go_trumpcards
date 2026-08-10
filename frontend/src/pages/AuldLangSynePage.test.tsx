import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { auldlangsyneApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { AuldLangSyneResponse, Card, CardDesign } from '../types/card';
import {
  AuldLangSynePage,
  auldlangsyneDealsLeft,
  auldlangsyneNextRank,
  auldlangsyneUpcomingRanks,
} from './AuldLangSynePage';

vi.mock('../api/gameApi', () => ({
  auldlangsyneApi: { exec: vi.fn() },
  actionLogApi: { auldlangsyne: vi.fn() },
}));

const mockExec = vi.mocked(auldlangsyneApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

/** The opening board: four Aces on the foundations, one card on each waste. */
const playingState: AuldLangSyneResponse = {
  foundations: [[card('SPADE', 1)], [card('HEART', 1)], [card('DIAMOND', 1)], [card('CLOVER', 1)]],
  wastes: [[card('SPADE', 2)], [card('HEART', 9)], [card('DIAMOND', 7)], [card('CLOVER', 13)]],
  stockCount: 44,
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
  messageCode: 'auldlangsyne.playing',
};

const stockEmptyState: AuldLangSyneResponse = {
  ...playingState,
  stockCount: 0,
  canUndo: true,
};

const stalemateState: AuldLangSyneResponse = {
  ...stockEmptyState,
  isStalemate: true,
  undoToEscape: 2,
  messageCode: 'auldlangsyne.stalemate',
};

const gameClearState: AuldLangSyneResponse = {
  ...playingState,
  phase: 1,
  stockCount: 0,
  moveCount: 100,
  messageCode: 'auldlangsyne.gameClear',
  messageParams: { moveCount: '100' },
};

/**
 * The response to an explicit `hint` command. `messageCode` ending in
 * `.hintAvailable` is what `isRequestedHint` keys off -- the board also carries
 * a passive `hint` on every response since #4483, so the code is the only thing
 * separating "the player asked" from "the tooltip needs data" (#4605).
 */
const hintRequestedState: AuldLangSyneResponse = {
  ...playingState,
  hint: { wasteIdx: 2, foundationIdx: 1 },
  messageCode: 'auldlangsyne.hintAvailable',
};

/** The same hint present passively, i.e. nobody asked for it. */
const passiveHintState: AuldLangSyneResponse = {
  ...playingState,
  hint: { wasteIdx: 2, foundationIdx: 1 },
  messageCode: 'auldlangsyne.playing',
};

const gameOverState: AuldLangSyneResponse = {
  ...playingState,
  phase: 2,
  messageCode: 'auldlangsyne.gameOver',
};

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
  mockExec.mockResolvedValue(playingState);
});

describe('auldlangsyneNextRank', () => {
  it('advances one rank at a time regardless of suit', () => {
    expect(auldlangsyneNextRank(1, 1)).toBe(2);
    expect(auldlangsyneNextRank(12, 12)).toBe(13);
  });

  it('opens an empty pile with an Ace', () => {
    expect(auldlangsyneNextRank(undefined, 0)).toBe(1);
  });

  it('returns null once the pile is complete', () => {
    expect(auldlangsyneNextRank(13, 13)).toBeNull();
  });
});

describe('auldlangsyneUpcomingRanks', () => {
  it('lists the ranks still needed, stopping at the King', () => {
    expect(auldlangsyneUpcomingRanks(10, 10)).toEqual([11, 12, 13]);
  });

  it('is empty for a complete pile', () => {
    expect(auldlangsyneUpcomingRanks(13, 13)).toEqual([]);
  });

  it('honours the look-ahead cap', () => {
    expect(auldlangsyneUpcomingRanks(1, 1, 2)).toEqual([2, 3]);
  });
});

describe('auldlangsyneDealsLeft', () => {
  it('divides the stock across the four wastes', () => {
    expect(auldlangsyneDealsLeft(44)).toBe(11);
    expect(auldlangsyneDealsLeft(4)).toBe(1);
    expect(auldlangsyneDealsLeft(0)).toBe(0);
  });
});

describe('AuldLangSynePage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<AuldLangSynePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders the heading', async () => {
    renderWithProviders(<AuldLangSynePage />);
    await waitFor(() => expect(screen.getByText('オールド・ラング・サイン')).toBeInTheDocument());
  });

  it('shows how many deals are left rather than the next card', async () => {
    renderWithProviders(<AuldLangSynePage />);
    await waitFor(() => expect(screen.getByTestId('als-deals-left')).toHaveTextContent('11'));
  });

  it('deals when the deal button is pressed', async () => {
    renderWithProviders(<AuldLangSynePage />);
    await waitFor(() => expect(screen.getByTestId('als-deal-button')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('als-deal-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('deal'));
  });

  it('disables the deal button once the stock is empty', async () => {
    mockExec.mockResolvedValue(stockEmptyState);
    renderWithProviders(<AuldLangSynePage />);
    await waitFor(() => expect(screen.getByTestId('als-deal-button')).toBeDisabled());
  });

  it('moves a waste top to a foundation once a waste is selected', async () => {
    renderWithProviders(<AuldLangSynePage />);
    await waitFor(() => expect(screen.getByTestId('als-waste-button-0')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('als-waste-button-0'));
    await waitFor(() => expect(screen.getByTestId('als-waste-button-0')).toHaveAttribute('aria-pressed', 'true'));

    fireEvent.click(screen.getByRole('button', { name: /ファンデーション 0/ }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste', idx: 0 }, { zone: 'foundation', idx: 0 }),
    );
  });

  // A foundation click with nothing selected must not fire a move: there is no
  // stock source here, so the selection is the only thing that can supply a card.
  it('does not move when no waste is selected', async () => {
    renderWithProviders(<AuldLangSynePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /ファンデーション 0/ })).toBeDisabled());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: /ファンデーション 0/ }));
    // Without this await the assertion passes whether or not a move fired — the
    // dispatch has not been flushed yet (#4439).
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());

    // Negative control: the same click DOES move once a waste is selected, so the
    // assertion above is testing the guard rather than a broken click path.
    fireEvent.click(screen.getByTestId('als-waste-button-0'));
    await waitFor(() => expect(screen.getByTestId('als-waste-button-0')).toHaveAttribute('aria-pressed', 'true'));
    fireEvent.click(screen.getByRole('button', { name: /ファンデーション 0/ }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste', idx: 0 }, { zone: 'foundation', idx: 0 }),
    );
  });

  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue(hintRequestedState);
    renderWithProviders(<AuldLangSynePage />);
    await waitFor(() => expect(screen.getByRole('status')).toHaveTextContent('ヒントがあります'));
  });

  // The other half of the gate: a passive hint on an ordinary response must not
  // surface the banner, or every turn shows one to a player who never asked.
  it('hides the hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue(passiveHintState);
    renderWithProviders(<AuldLangSynePage />);
    await waitFor(() => expect(screen.getByTestId('als-deal-button')).toBeInTheDocument());
    expect(screen.queryByRole('status')).not.toBeInTheDocument();
  });

  it('toggles a waste selection off when clicked twice', async () => {
    renderWithProviders(<AuldLangSynePage />);
    await waitFor(() => expect(screen.getByTestId('als-waste-button-1')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('als-waste-button-1'));
    await waitFor(() => expect(screen.getByTestId('als-waste-button-1')).toHaveAttribute('aria-pressed', 'true'));

    fireEvent.click(screen.getByTestId('als-waste-button-1'));
    await waitFor(() => expect(screen.getByTestId('als-waste-button-1')).toHaveAttribute('aria-pressed', 'false'));
  });

  it('enables auto-complete only once the stock is empty', async () => {
    renderWithProviders(<AuldLangSynePage />);
    await waitFor(() => expect(screen.getByTestId('autocomplete-button')).toBeDisabled());
  });

  it('runs auto-complete when the stock is empty', async () => {
    mockExec.mockResolvedValue(stockEmptyState);
    renderWithProviders(<AuldLangSynePage />);
    await waitFor(() => expect(screen.getByTestId('autocomplete-button')).toBeEnabled());

    fireEvent.click(screen.getByTestId('autocomplete-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('requests a hint', async () => {
    renderWithProviders(<AuldLangSynePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('undoes only when there is history', async () => {
    renderWithProviders(<AuldLangSynePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled());

    mockExec.mockResolvedValue(stockEmptyState);
    renderWithProviders(<AuldLangSynePage />);
    await waitFor(() => expect(screen.getAllByRole('button', { name: '元に戻す' })[1]).toBeEnabled());
  });

  it('offers the stalemate escape when stuck', async () => {
    mockExec.mockResolvedValue(stalemateState);
    renderWithProviders(<AuldLangSynePage />);
    await waitFor(() => expect(screen.getByText(/オールド・ラング・サイン/)).toBeInTheDocument());
    expect(screen.getByTestId('als-deal-button')).toBeDisabled();
  });

  it('hides the playing controls once the game clears', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<AuldLangSynePage />);
    await waitFor(() => expect(screen.queryByTestId('als-deal-button')).not.toBeInTheDocument());
  });

  it('hides the playing controls after a give-up', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<AuldLangSynePage />);
    await waitFor(() => expect(screen.queryByTestId('als-deal-button')).not.toBeInTheDocument());
  });
});
