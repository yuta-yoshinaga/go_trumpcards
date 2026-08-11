import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { estimationApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, EstimationResponse } from '../types/card';
import { EstimationPage } from './EstimationPage';

vi.mock('../api/gameApi', () => ({
  estimationApi: { exec: vi.fn() },
  actionLogApi: { estimation: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(estimationApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 3,
  cards: id === 0 ? [card('HEART', 11), card('SPADE', 9), card('CLOVER', 1)] : [],
  bid: -1,
  callType: 0,
  trickCount: 0,
  roundScore: 0,
  totalScore: 0,
  ...over,
});

function makeState(overrides: Partial<EstimationResponse> = {}): EstimationResponse {
  return {
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    roundNumber: 1,
    trickNumber: 0,
    trumpSuit: 0,
    restrictedBid: -1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    leadPlayerIdx: 0,
    // 既定では親を 0 にして、人間が切り札を選べる状況にする。
    dealerIdx: 0,
    currentTrick: [],
    validPlays: [0, 1, 2],
    gameEndFlag: false,
    winnerIdx: -1,
    config: { rounds: 5 },
    message: '',
    ...overrides,
  } as unknown as EstimationResponse;
}

/** A state where trump and calls are settled and it is the human's turn. */
const playing = (over: Partial<EstimationResponse> = {}) =>
  makeState({ phase: 2, trumpSuit: 3, ...over } as Partial<EstimationResponse>);

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
});

describe('EstimationPage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<EstimationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **的中だけが得点になることは盤面から読めない。** 常に出ていること。
  it('always states the scoring', async () => {
    renderWithProviders(<EstimationPage />);
    const box = await screen.findByTestId('est-score');
    expect(box).toHaveTextContent(/Dash Call/);
    expect(box).toHaveTextContent(/Risk/);
  });

  it('offers all four trump suits to the dealer', async () => {
    renderWithProviders(<EstimationPage />);
    for (const suit of [1, 2, 3, 4]) {
      expect(await screen.findByTestId(`est-trump-${suit.toString()}-btn`)).toBeInTheDocument();
    }
  });

  it('hides the trump buttons when the human is not the dealer', async () => {
    mockExec.mockResolvedValue(makeState({ dealerIdx: 2 }));
    renderWithProviders(<EstimationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('est-trump-1-btn')).not.toBeInTheDocument();
  });

  // **切り札は4番目の引数で送る。** 位置がずれると別の値として届く。
  it.each([1, 2, 3, 4])('sends trump suit %s', async (suit) => {
    renderWithProviders(<EstimationPage />);
    const btn = await screen.findByTestId(`est-trump-${suit.toString()}-btn`);
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('trump', undefined, undefined, suit));
  });

  it('offers every call from 0 to 13 while bidding', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, trumpSuit: 3 }));
    renderWithProviders(<EstimationPage />);
    expect(await screen.findByTestId('est-bid-0-btn')).toBeInTheDocument();
    expect(screen.getByTestId('est-bid-13-btn')).toBeInTheDocument();
  });

  // **宣言は5番目の引数。** 0（Dash Call）も同じ経路で送る。
  it.each([0, 5, 13])('sends call %s', async (bid) => {
    mockExec.mockResolvedValue(makeState({ phase: 1, trumpSuit: 3 }));
    renderWithProviders(<EstimationPage />);
    const btn = await screen.findByTestId(`est-bid-${bid.toString()}-btn`);
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', undefined, undefined, undefined, bid));
  });

  // **禁止値は押せない。** サーバが必ず拒否するので出させない。負のコントロール付き。
  it('disables only the barred call', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, trumpSuit: 3, restrictedBid: 4 }));
    renderWithProviders(<EstimationPage />);

    expect(await screen.findByTestId('est-bid-4-btn')).toBeDisabled();
    expect(screen.getByTestId('est-bid-3-btn')).toBeEnabled();
    expect(screen.getByTestId('est-bid-5-btn')).toBeEnabled();
  });

  it('enables every call when nothing is barred', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, trumpSuit: 3, restrictedBid: -1 }));
    renderWithProviders(<EstimationPage />);
    expect(await screen.findByTestId('est-bid-4-btn')).toBeEnabled();
  });

  it('hides the call buttons once play starts', async () => {
    mockExec.mockResolvedValue(playing());
    renderWithProviders(<EstimationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('est-bid-0-btn')).not.toBeInTheDocument();
  });

  it('plays the clicked card by its hand index', async () => {
    mockExec.mockResolvedValue(playing());
    renderWithProviders(<EstimationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    mockExec.mockClear();
    fireEvent.click(cards[2]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 2));
  });

  // **宣言の種類は盤面から読めない。** 席ごとに出す。3 種すべて踏む。
  it('labels each seat with its call kind', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          seat(0, { bid: 0, callType: 1 }),
          seat(1, { bid: 6, callType: 2, totalScore: 44 }),
          seat(2, { bid: 3, callType: 0 }),
          seat(3),
        ],
      } as Partial<EstimationResponse>),
    );
    renderWithProviders(<EstimationPage />);

    expect(await screen.findByTestId('est-seat-0')).toHaveTextContent('Dash(0)');
    expect(screen.getByTestId('est-seat-1')).toHaveTextContent('Risk(6)');
    expect(screen.getByTestId('est-seat-1')).toHaveTextContent('44');
    expect(screen.getByTestId('est-seat-2')).toHaveTextContent('宣言3');
    expect(screen.getByTestId('est-seat-3')).toHaveTextContent('未宣言');
  });

  // 切り札は未定と確定の両側を踏む。
  it('shows the trump once chosen', async () => {
    mockExec.mockResolvedValue(makeState());
    const { unmount } = renderWithProviders(<EstimationPage />);
    expect(await screen.findByTestId('est-trump')).toHaveTextContent(/未定/);
    unmount();

    mockExec.mockResolvedValue(playing({ trumpSuit: 4 } as Partial<EstimationResponse>));
    renderWithProviders(<EstimationPage />);
    expect(await screen.findByTestId('est-trump')).toHaveTextContent('♦');
  });

  it('advances the round when the next-round button is pressed', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 3 }));
    renderWithProviders(<EstimationPage />);

    const btn = await screen.findByRole('button', { name: '次のラウンドへ' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<EstimationPage />);
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
      mockExec.mockResolvedValue(makeState({ gameEndFlag: true, phase: 4, winnerIdx }));
      const { unmount } = renderWithProviders(<EstimationPage />);
      expect(await screen.findByText(expected)).toBeInTheDocument();
      unmount();
    }
  });

  it('disables the hand while it is a CPU turn', async () => {
    mockExec.mockResolvedValue(playing({ currentPlayerIdx: 1 } as Partial<EstimationResponse>));
    renderWithProviders(<EstimationPage />);
    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    expect(cards[0]).toBeDisabled();
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'bid-0', reason: 'hint.estimationDashCall', confidence: 'moderate' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<EstimationPage />);
    // 得点表にも "Dash Call" が出るので、ヒント固有の文言で絞る。
    expect(await screen.findByText(/強い札がありません/)).toBeInTheDocument();
  });
});
