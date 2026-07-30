import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { mushiApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { MushiCard, MushiPlayer, MushiResponse } from '../types/card';
import { MushiPage } from './MushiPage';

vi.mock('../api/gameApi', () => ({
  mushiApi: { exec: vi.fn() },
  actionLogApi: { mushi: vi.fn() },
}));

const mockExec = vi.mocked(mushiApi.exec);

const card = (month: number, index: number, category = 0, isWild = false): MushiCard => ({
  // `design`/`value` are the wire's generic card fields; hanafuda identity
  // travels in `month`/`index`, which is what the UI reads.
  design: 'SPADE',
  value: index,
  month,
  index,
  category,
  points: [1, 5, 10, 20][category] ?? 1,
  isWild,
});

function human(overrides?: Partial<MushiPlayer>): MushiPlayer {
  return {
    id: 0,
    isHuman: true,
    cardCount: 2,
    cards: [card(1, 1, 3), card(11, 4, 0, true)],
    captured: [card(3, 2, 1)],
    capturedPoints: 5,
    score: 0,
    roundResult: 0,
    hidden: false,
    ...overrides,
  };
}

function cpu(overrides?: Partial<MushiPlayer>): MushiPlayer {
  // A hidden seat arrives with a count and NO hand cards -- but its CAPTURED
  // cards are still present, because those are public.
  return {
    id: 1,
    isHuman: false,
    cardCount: 2,
    cards: [],
    captured: [card(12, 1, 3)],
    capturedPoints: 20,
    score: 0,
    roundResult: 0,
    hidden: true,
    ...overrides,
  };
}

function makeState(overrides?: Partial<MushiResponse>): MushiResponse {
  return {
    players: [human(), cpu()],
    field: [card(5, 3), card(9, 2, 1)],
    phase: 0,
    currentPlayerIdx: 0,
    dealerIdx: 0,
    roundNumber: 1,
    targetRounds: 12,
    stockCount: 16,
    selectableIndices: [],
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    ...overrides,
  };
}

describe('MushiPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<MushiPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows both seats captures but never the opponent hand', async () => {
    renderWithProviders(<MushiPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    // Captures are public for both sides -- the game is unplayable without them.
    expect(screen.getByText('あなたの取り札 (5 pt)')).toBeInTheDocument();
    expect(screen.getByText('CPU の取り札 (20 pt)')).toBeInTheDocument();
    // The opponent's HAND is drawn from its count, not its (absent) cards.
    expect(screen.getByText('CPU の手札 2 枚')).toBeInTheDocument();
  });

  it('plays a hand card', async () => {
    renderWithProviders(<MushiPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    mockExec.mockClear();
    const handButtons = screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'play');
    fireEvent.click(handButtons[0]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0));
  });

  it('only allows the field cards the server marked selectable', async () => {
    // The wild's "not another willow" rule lives on the server; the page must
    // not accept a click on a card the server did not offer.
    mockExec.mockResolvedValue(makeState({ phase: 2, pendingCard: card(11, 4, 0, true), selectableIndices: [1] }));
    renderWithProviders(<MushiPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    const fieldButtons = screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'select');
    mockExec.mockClear();

    fireEvent.click(fieldButtons[0]);
    // Without the flush this assertion cannot fail: nothing has had a chance
    // to dispatch yet, so "not called" is true for any implementation.
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    fireEvent.click(fieldButtons[1]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('select', undefined, 1));
  });

  it('offers the next round only once the round has settled', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<MushiPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '次のラウンド' })).not.toBeInTheDocument();

    mockExec.mockResolvedValue(makeState({ phase: 3, players: [human({ roundResult: -12 }), cpu()] }));
    renderWithProviders(<MushiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('このラウンド: -12')).toBeInTheDocument();

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のラウンド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('reports the outcome once the game ends', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 4, gameEndFlag: true, winnerIdx: 0, messageCode: 'mushi.win' }));
    renderWithProviders(<MushiPage />);
    await waitFor(() => expect(screen.getByText('あなたの勝ちです')).toBeInTheDocument());
  });
});
