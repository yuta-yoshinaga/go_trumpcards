import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { laughandliedownApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, LaughAndLieDownPlayer, LaughAndLieDownResponse } from '../types/card';
import { LaughAndLieDownPage } from './LaughAndLieDownPage';

vi.mock('../api/gameApi', () => ({
  laughandliedownApi: { exec: vi.fn() },
  actionLogApi: { laughandliedown: vi.fn() },
}));

const mockExec = vi.mocked(laughandliedownApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function human(overrides?: Partial<LaughAndLieDownPlayer>): LaughAndLieDownPlayer {
  return {
    id: 0,
    isHuman: true,
    cardCount: 3,
    cards: [card('SPADE', 7), card('HEART', 9), card('CLOVER', 13)],
    wonCount: 4,
    laidDown: false,
    score: 0,
    hidden: false,
    ...overrides,
  };
}

function cpu(id: number, overrides?: Partial<LaughAndLieDownPlayer>): LaughAndLieDownPlayer {
  return {
    id,
    isHuman: false,
    cardCount: 3,
    cards: [],
    wonCount: 2,
    laidDown: false,
    score: 0,
    hidden: true,
    ...overrides,
  };
}

function makeState(overrides?: Partial<LaughAndLieDownResponse>): LaughAndLieDownResponse {
  return {
    players: [human(), cpu(1), cpu(2), cpu(3), cpu(4)],
    layout: [card('CLOVER', 7), card('DIAMOND', 7), card('SPADE', 3)],
    phase: 0,
    currentPlayerIdx: 0,
    validIndices: [0],
    threeTakeIndices: [],
    dealerIdx: 0,
    lastInIdx: -1,
    pot: 11,
    gameEndFlag: false,
    message: '',
    ...overrides,
  };
}

describe('LaughAndLieDownPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<LaughAndLieDownPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows the pot, the dealer and both rules permanently', async () => {
    renderWithProviders(<LaughAndLieDownPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.getByText(/ポット: 11/)).toBeInTheDocument();
    expect(screen.getByText(/1枚か3枚/)).toBeInTheDocument();
    expect(screen.getByText(/手札を全部場に置いて降りる/)).toBeInTheDocument();
  });

  it('shows every seat won count and marks who has laid down', async () => {
    // 取り札の枚数は 8 との差がそのまま収支なので、常に見えている必要がある。
    mockExec.mockResolvedValue(makeState({ players: [human(), cpu(1, { laidDown: true, wonCount: 9 })] }));
    renderWithProviders(<LaughAndLieDownPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.getByText(/取り札9枚 · 降りた/)).toBeInTheDocument();
  });

  it('only plays the hand cards the server marked valid', async () => {
    renderWithProviders(<LaughAndLieDownPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    const handButtons = screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'play');
    mockExec.mockClear();

    fireEvent.click(handButtons[1]);
    // Without the flush this cannot fail: nothing has had a chance to dispatch.
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    fireEvent.click(handButtons[0]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0, 1));
  });

  it('offers the three-card take only where the server said three are on the table', async () => {
    mockExec.mockResolvedValue(makeState({ validIndices: [0, 1], threeTakeIndices: [0] }));
    renderWithProviders(<LaughAndLieDownPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    // 2 枚の合法手のうち、3 枚取りが提示されるのは 1 枚だけ。
    expect(screen.getAllByRole('button', { name: '3枚取る' })).toHaveLength(1);
  });

  it('sends a take count of three once the option is armed', async () => {
    mockExec.mockResolvedValue(makeState({ validIndices: [0], threeTakeIndices: [0] }));
    renderWithProviders(<LaughAndLieDownPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    fireEvent.click(screen.getByRole('button', { name: '3枚取る' }));
    mockExec.mockClear();

    const handButtons = screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'play');
    fireEvent.click(handButtons[0]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0, 3));
  });

  it('reports each outcome by net result, not by finishing order', async () => {
    for (const [score, text] of [
      [3, '勝ち越しました'],
      [0, '収支ゼロでした'],
      [-2, '負け越しました'],
    ] as const) {
      const code = score > 0 ? 'laughandliedown.win' : score === 0 ? 'laughandliedown.even' : 'laughandliedown.lose';
      mockExec.mockResolvedValue(
        makeState({
          phase: 1,
          gameEndFlag: true,
          players: [human({ cards: [], cardCount: 0, score }), cpu(1)],
          messageCode: code,
        }),
      );
      renderWithProviders(<LaughAndLieDownPage />);
      await waitFor(() => expect(screen.getAllByText(text).length).toBeGreaterThan(0));
    }
  });
});
