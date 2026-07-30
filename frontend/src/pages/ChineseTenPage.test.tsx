import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { chinesetenApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { CardDesign, ChineseTenCard, ChineseTenPlayer, ChineseTenResponse } from '../types/card';
import { ChineseTenPage } from './ChineseTenPage';

vi.mock('../api/gameApi', () => ({
  chinesetenApi: { exec: vi.fn() },
  actionLogApi: { chineseten: vi.fn() },
}));

const mockExec = vi.mocked(chinesetenApi.exec);

const card = (design: CardDesign, value: number, points = 0): ChineseTenCard => ({
  design,
  value,
  points,
  isRed: design === 'HEART' || design === 'DIAMOND',
});

function human(overrides?: Partial<ChineseTenPlayer>): ChineseTenPlayer {
  return {
    id: 0,
    isHuman: true,
    cardCount: 3,
    cards: [card('SPADE', 1), card('HEART', 9, 10), card('CLOVER', 13)],
    captured: [card('HEART', 5, 5)],
    score: 5,
    hidden: false,
    ...overrides,
  };
}

function cpu(overrides?: Partial<ChineseTenPlayer>): ChineseTenPlayer {
  // A hidden seat arrives with a count and NO hand cards -- but its CAPTURES
  // are still present, because those are public.
  return {
    id: 1,
    isHuman: false,
    cardCount: 3,
    cards: [],
    captured: [card('DIAMOND', 1, 20)],
    score: 20,
    hidden: true,
    ...overrides,
  };
}

function makeState(overrides?: Partial<ChineseTenResponse>): ChineseTenResponse {
  return {
    players: [human(), cpu()],
    layout: [card('SPADE', 9), card('DIAMOND', 9, 10), card('CLOVER', 4)],
    phase: 0,
    currentPlayerIdx: 0,
    stockCount: 24,
    selectableIndices: [],
    tieScore: 105,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    ...overrides,
  };
}

describe('ChineseTenPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<ChineseTenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows the two capture rules permanently', async () => {
    // A-9 and 10-K capture differently, and that is what a player gets wrong.
    renderWithProviders(<ChineseTenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.getByText(/A〜9は合計10で取る/)).toBeInTheDocument();
  });

  it('shows both seats captures but never the opponent hand', async () => {
    renderWithProviders(<ChineseTenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    expect(screen.getByText('あなたの取り札 (5)')).toBeInTheDocument();
    expect(screen.getByText('CPU の取り札 (20)')).toBeInTheDocument();
    expect(screen.getByText('CPU の手札 3 枚')).toBeInTheDocument();
  });

  it('plays a hand card', async () => {
    renderWithProviders(<ChineseTenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    mockExec.mockClear();
    const handButtons = screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'play');
    fireEvent.click(handButtons[0]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0));
  });

  it('only allows the layout cards the server marked selectable', async () => {
    // Both capture rules live on the server; the page must not accept a click
    // on a card it did not offer.
    mockExec.mockResolvedValue(makeState({ phase: 1, pendingCard: card('SPADE', 1), selectableIndices: [1] }));
    renderWithProviders(<ChineseTenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    const layoutButtons = screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'select');
    mockExec.mockClear();

    fireEvent.click(layoutButtons[0]);
    // Without the flush this cannot fail: nothing has had a chance to dispatch.
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    fireEvent.click(layoutButtons[1]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('select', undefined, 1));
  });

  it('reports each outcome', async () => {
    for (const [winner, text] of [
      [0, 'あなたの勝ちです'],
      [1, 'あなたの負けです'],
      [-1, '引き分けです'],
    ] as const) {
      const code = winner === 0 ? 'chineseten.win' : winner === 1 ? 'chineseten.lose' : 'chineseten.draw';
      mockExec.mockResolvedValue(makeState({ phase: 2, gameEndFlag: true, winnerIdx: winner, messageCode: code }));
      renderWithProviders(<ChineseTenPage />);
      await waitFor(() => expect(screen.getAllByText(text).length).toBeGreaterThan(0));
    }
  });
});
