import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { sergeantmajorApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, SergeantMajorResponse } from '../types/card';
import { SergeantMajorPage } from './SergeantMajorPage';

vi.mock('../api/gameApi', () => ({
  sergeantmajorApi: { exec: vi.fn() },
  actionLogApi: { sergeantmajor: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(sergeantmajorApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 5,
  cards: id === 0 ? [card('HEART', 1), card('SPADE', 9), card('CLOVER', 13), card('DIAMOND', 4), card('SPADE', 2)] : [],
  target: [8, 5, 3][id] ?? 0,
  trickCount: 0,
  score: 0,
  ...over,
});

function makeState(overrides: Partial<SergeantMajorResponse> = {}): SergeantMajorResponse {
  return {
    players: [seat(0), seat(1), seat(2)],
    phase: 0,
    roundNumber: 1,
    trickNumber: 0,
    trumpSuit: 0,
    kittySize: 4,
    discardCount: 4,
    lastExchange: 0,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    dealerIdx: 0,
    currentTrick: [],
    validPlays: [0, 1, 2],
    gameEndFlag: false,
    winnerIdx: -1,
    config: { rounds: 3 },
    message: '',
    ...overrides,
  } as unknown as SergeantMajorResponse;
}

/** A state where trump is settled and it is the human's turn to play. */
const playing = (over: Partial<SergeantMajorResponse> = {}) =>
  makeState({ phase: 2, trumpSuit: 3, kittySize: 0, ...over } as Partial<SergeantMajorResponse>);

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
});

describe('SergeantMajorPage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<SergeantMajorPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **ノルマは席順で決まる。** これが読めないと何をすべきか分からない。
  it('states that the targets follow the seats', async () => {
    renderWithProviders(<SergeantMajorPage />);
    expect(await screen.findByTestId('sm-rule')).toHaveTextContent(/席順/);
  });

  // **8/5/3 が必ず 1 つずつ出る。**
  it('renders all three seat targets', async () => {
    mockExec.mockResolvedValue(playing());
    renderWithProviders(<SergeantMajorPage />);
    for (const [id, target] of [
      [0, '8'],
      [1, '5'],
      [2, '3'],
    ] as const) {
      expect(await screen.findByTestId(`sm-seat-${id}`)).toHaveTextContent(target);
    }
  });

  it('marks the dealer', async () => {
    mockExec.mockResolvedValue(playing({ dealerIdx: 1 } as Partial<SergeantMajorResponse>));
    renderWithProviders(<SergeantMajorPage />);
    expect(await screen.findByTestId('sm-seat-1')).toHaveTextContent(/親/);
    expect(screen.getByTestId('sm-seat-0')).not.toHaveTextContent(/親/);
  });

  it('offers all four trump suits to the dealer', async () => {
    renderWithProviders(<SergeantMajorPage />);
    for (const suit of [1, 2, 3, 4]) {
      expect(await screen.findByTestId(`sm-trump-${suit.toString()}-btn`)).toBeInTheDocument();
    }
  });

  it('hides the trump buttons when the human is not the dealer', async () => {
    mockExec.mockResolvedValue(makeState({ dealerIdx: 2 }));
    renderWithProviders(<SergeantMajorPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('sm-trump-1-btn')).not.toBeInTheDocument();
  });

  // **切り札は4番目の引数で送る。** 位置がずれると別の値として届く。
  it.each([1, 2, 3, 4])('sends trump suit %s', async (suit) => {
    renderWithProviders(<SergeantMajorPage />);
    const btn = await screen.findByTestId(`sm-trump-${suit.toString()}-btn`);
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('trump', undefined, undefined, suit));
  });

  // **ちょうど 4 枚選ぶまで確定できない。** サーバが必ず拒否する操作は出さない。
  it('requires exactly four cards before the discard confirms', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, dealerIdx: 0, kittySize: 0 } as Partial<SergeantMajorResponse>));
    renderWithProviders(<SergeantMajorPage />);

    const confirm = await screen.findByTestId('sm-discard-btn');
    expect(confirm).toBeDisabled();

    const cards = screen.getAllByRole('button', { name: /捨て札に選ぶ/ });
    fireEvent.click(cards[0]);
    fireEvent.click(cards[1]);
    fireEvent.click(cards[2]);
    expect(screen.getByTestId('sm-discard-btn')).toBeDisabled();

    fireEvent.click(cards[3]);
    expect(cards[3]).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByTestId('sm-discard-btn')).toBeEnabled();

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('sm-discard-btn'));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('discard', undefined, undefined, undefined, [0, 1, 2, 3]),
    );
  });

  // **選び直せる。** 押し直すと外れる。
  it('lets you unpick a card', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, dealerIdx: 0, kittySize: 0 } as Partial<SergeantMajorResponse>));
    renderWithProviders(<SergeantMajorPage />);

    const cards = await screen.findAllByRole('button', { name: /捨て札に選ぶ/ });
    fireEvent.click(cards[0]);
    expect(cards[0]).toHaveAttribute('aria-pressed', 'true');
    fireEvent.click(cards[0]);
    expect(cards[0]).toHaveAttribute('aria-pressed', 'false');
  });

  it('plays the clicked card by its hand index', async () => {
    mockExec.mockResolvedValue(playing());
    renderWithProviders(<SergeantMajorPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    mockExec.mockClear();
    fireEvent.click(cards[2]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 2));
  });

  // 切り札は未宣言と確定の両側を踏む。
  it('mentions the kitty, then the trump', async () => {
    const { unmount } = renderWithProviders(<SergeantMajorPage />);
    expect(await screen.findByTestId('sm-trump')).toHaveTextContent(/未宣言/);
    unmount();

    mockExec.mockResolvedValue(playing({ trumpSuit: 4 } as Partial<SergeantMajorResponse>));
    renderWithProviders(<SergeantMajorPage />);
    expect(await screen.findByTestId('sm-trump')).toHaveTextContent('♦');
  });

  // **前ラウンドの札のやり取りは盤面に痕跡が残らない。**
  it('reports the exchange only when cards actually moved', async () => {
    mockExec.mockResolvedValue(playing());
    const { unmount } = renderWithProviders(<SergeantMajorPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('sm-exchange')).not.toBeInTheDocument();
    unmount();

    mockExec.mockResolvedValue(playing({ lastExchange: 3 } as Partial<SergeantMajorResponse>));
    renderWithProviders(<SergeantMajorPage />);
    expect(await screen.findByTestId('sm-exchange')).toHaveTextContent('3');
  });

  it('advances to the next round when the button is pressed', async () => {
    mockExec.mockResolvedValue(playing({ phase: 3 } as Partial<SergeantMajorResponse>));
    renderWithProviders(<SergeantMajorPage />);

    const btn = await screen.findByRole('button', { name: '次のラウンドへ' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<SergeantMajorPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '投了' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('renders the result banner for each outcome', async () => {
    for (const [winnerIdx, expected] of [
      [0, /あなたの勝ち/],
      [1, /CPU1 の勝ち/],
      [-1, /同点/],
    ] as const) {
      mockExec.mockResolvedValue(playing({ gameEndFlag: true, phase: 4, winnerIdx }));
      const { unmount } = renderWithProviders(<SergeantMajorPage />);
      expect(await screen.findByTestId('sm-result')).toHaveTextContent(expected);
      unmount();
    }
  });

  it('disables the hand while it is a CPU turn', async () => {
    mockExec.mockResolvedValue(playing({ currentPlayerIdx: 1 } as Partial<SergeantMajorResponse>));
    renderWithProviders(<SergeantMajorPage />);
    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    expect(cards[0]).toBeDisabled();
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'trump-3', reason: 'hint.sergeantmajorSelectTrump', confidence: 'moderate' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<SergeantMajorPage />);
    expect(await screen.findByTestId('hint-tooltip')).toHaveTextContent(/いちばん強いスート/);
  });
});
