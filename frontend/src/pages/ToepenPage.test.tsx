import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { toepenApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, ToepenPlayer, ToepenResponse } from '../types/card';
import { ToepenPage } from './ToepenPage';

vi.mock('../api/gameApi', () => ({
  toepenApi: { exec: vi.fn() },
  actionLogApi: { toepen: vi.fn() },
}));

const mockExec = vi.mocked(toepenApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function human(overrides?: Partial<ToepenPlayer>): ToepenPlayer {
  return {
    id: 0,
    isHuman: true,
    cardCount: 3,
    cards: [card('SPADE', 10), card('HEART', 11), card('CLOVER', 7)],
    lives: 0,
    folded: false,
    eliminated: false,
    hidden: false,
    ...overrides,
  };
}

function cpu(id: number, overrides?: Partial<ToepenPlayer>): ToepenPlayer {
  // A hidden seat arrives with a count and NO cards.
  return {
    id,
    isHuman: false,
    cardCount: 3,
    cards: [],
    lives: 0,
    folded: false,
    eliminated: false,
    hidden: true,
    ...overrides,
  };
}

function makeState(overrides?: Partial<ToepenResponse>): ToepenResponse {
  return {
    players: [human(), cpu(1), cpu(2), cpu(3)],
    phase: 0,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    dealerIdx: 0,
    currentTrick: [],
    leadSuit: -1,
    trickNumber: 0,
    handNumber: 1,
    stake: 1,
    knockerIdx: -1,
    pendingRespondent: -1,
    lastTrickWinner: -1,
    maxLives: 10,
    validPlayIndices: [0, 1, 2],
    canRedeal: false,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    ...overrides,
  };
}

describe('ToepenPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<ToepenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows the inverted ranking permanently', async () => {
    // It is the one thing about this game that is easy to get backwards, so it
    // is on screen at all times rather than only in the tutorial.
    renderWithProviders(<ToepenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.getByText(/10 > 9 > 8 > 7 > A > K > Q > J/)).toBeInTheDocument();
  });

  it('never renders the opponent hands as cards', async () => {
    renderWithProviders(<ToepenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    const handButtons = screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'play');
    expect(handButtons).toHaveLength(3);
  });

  it('only plays the cards the server marked legal', async () => {
    // The follow-suit obligation lives on the server; the page must not accept
    // a click on a card it did not offer.
    mockExec.mockResolvedValue(makeState({ validPlayIndices: [1] }));
    renderWithProviders(<ToepenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    const handButtons = screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'play');
    mockExec.mockClear();

    fireEvent.click(handButtons[0]);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    fireEvent.click(handButtons[1]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 1));
  });

  it('raises the stake with toep', async () => {
    renderWithProviders(<ToepenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: /toep/i }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('toep'));
  });

  it('offers stay and fold only while a toep is on the human, and prices the fold below the stake', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, pendingRespondent: 0, knockerIdx: 2, stake: 3 }));
    renderWithProviders(<ToepenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    // Folding costs the stake BEFORE the raise: 2, not 3.
    expect(screen.getByRole('button', { name: '降りる（2点）' })).toBeInTheDocument();

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '追随する' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('answer', undefined, true));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '降りる（2点）' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('answer', undefined, false));
  });

  it('offers the next hand only once the hand has settled', async () => {
    renderWithProviders(<ToepenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '次のハンド' })).not.toBeInTheDocument();

    mockExec.mockResolvedValue(makeState({ phase: 2 }));
    renderWithProviders(<ToepenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のハンド' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のハンド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('reports the outcome and marks an eliminated seat', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 3,
        gameEndFlag: true,
        winnerIdx: 0,
        messageCode: 'toepen.win',
        players: [human(), cpu(1, { lives: 10, eliminated: true }), cpu(2), cpu(3)],
      }),
    );
    renderWithProviders(<ToepenPage />);
    await waitFor(() => expect(screen.getByText('あなたの勝ちです')).toBeInTheDocument());
    expect(screen.getByText('10/10')).toBeInTheDocument();
  });
});

describe('ToepenPage redeal', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('offers the redeal only when the server says it is available', async () => {
    // Whether a hand qualifies, and whether the window is still open, are both
    // the server's call -- the page must not count ranks itself.
    mockExec.mockResolvedValue(makeState({ canRedeal: false }));
    renderWithProviders(<ToepenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '配り直し（貧民）' })).not.toBeInTheDocument();

    mockExec.mockResolvedValue(makeState({ canRedeal: true }));
    renderWithProviders(<ToepenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '配り直し（貧民）' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '配り直し（貧民）' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('redeal'));
  });
});
