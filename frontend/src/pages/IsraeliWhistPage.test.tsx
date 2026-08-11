import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { israeliwhistApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, IsraeliWhistResponse } from '../types/card';
import { IsraeliWhistPage } from './IsraeliWhistPage';

vi.mock('../api/gameApi', () => ({
  israeliwhistApi: { exec: vi.fn() },
  actionLogApi: { israeliwhist: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(israeliwhistApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 3,
  cards: id === 0 ? [card('HEART', 11), card('SPADE', 9), card('CLOVER', 1)] : [],
  auctionBid: -1,
  auctionSuit: 0,
  passed: false,
  bid: -1,
  trickCount: 0,
  roundScore: 0,
  totalScore: 0,
  ...over,
});

function makeState(overrides: Partial<IsraeliWhistResponse> = {}): IsraeliWhistResponse {
  return {
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    roundNumber: 1,
    trickNumber: 0,
    trumpSuit: 0,
    declarerIdx: -1,
    highBid: 0,
    highSuit: 0,
    minimumBid: 0,
    restrictedBid: -1,
    currentPlayerIdx: 0,
    auctionPlayerIdx: 0,
    bidPlayerIdx: 0,
    leadPlayerIdx: 0,
    dealerIdx: 3,
    currentTrick: [],
    validPlays: [0, 1, 2],
    gameEndFlag: false,
    winnerIdx: -1,
    config: { rounds: 4 },
    message: '',
    ...overrides,
  } as unknown as IsraeliWhistResponse;
}

/** A state where both bidding rounds are settled and it is the human's turn. */
const playing = (over: Partial<IsraeliWhistResponse> = {}) =>
  makeState({ phase: 2, trumpSuit: 3, declarerIdx: 1, highBid: 7, ...over } as Partial<IsraeliWhistResponse>);

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
});

describe('IsraeliWhistPage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<IsraeliWhistPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **2乗で跳ねることと全員一致の倍率は盤面から読めない。** 常に出ていること。
  it('always states the scoring', async () => {
    renderWithProviders(<IsraeliWhistPage />);
    const box = await screen.findByTestId('iw-score');
    expect(box).toHaveTextContent(/2/);
    expect(box).toHaveTextContent(/全員/);
  });

  it('offers all four suits plus pass during the auction', async () => {
    renderWithProviders(<IsraeliWhistPage />);
    for (const suit of [1, 2, 3, 4]) {
      expect(await screen.findByTestId(`iw-auction-${suit.toString()}-btn`)).toBeInTheDocument();
    }
    expect(screen.getByTestId('iw-pass-btn')).toBeInTheDocument();
  });

  // **入札は数とスートの両方を送る。** 位置がずれると別の入札になる。
  it.each([1, 2, 3, 4])('sends an auction bid in suit %s', async (suit) => {
    renderWithProviders(<IsraeliWhistPage />);
    const btn = await screen.findByTestId(`iw-auction-${suit.toString()}-btn`);
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('auction', undefined, undefined, suit, 5));
  });

  // **競り上げは現在の最高入札を上回る数で送る。**
  it('offers a bid that beats the standing one', async () => {
    mockExec.mockResolvedValue(makeState({ highBid: 8, highSuit: 1 }));
    renderWithProviders(<IsraeliWhistPage />);
    const btn = await screen.findByTestId('iw-auction-4-btn');
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('auction', undefined, undefined, 4, 8));
  });

  it('sends pass as its own command', async () => {
    renderWithProviders(<IsraeliWhistPage />);
    const btn = await screen.findByTestId('iw-pass-btn');
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
  });

  // **降りた席には入札ボタンを出さない。**
  it('hides the auction buttons once the human has passed', async () => {
    mockExec.mockResolvedValue(
      makeState({ players: [seat(0, { passed: true }), seat(1), seat(2), seat(3)] } as Partial<IsraeliWhistResponse>),
    );
    renderWithProviders(<IsraeliWhistPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('iw-auction-1-btn')).not.toBeInTheDocument();
    expect(screen.queryByTestId('iw-pass-btn')).not.toBeInTheDocument();
  });

  it('offers every call from 0 to 13 in the second round', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, trumpSuit: 3, declarerIdx: 1, highBid: 7 }));
    renderWithProviders(<IsraeliWhistPage />);
    expect(await screen.findByTestId('iw-bid-0-btn')).toBeInTheDocument();
    expect(screen.getByTestId('iw-bid-13-btn')).toBeInTheDocument();
  });

  it.each([0, 5, 13])('sends call %s', async (bid) => {
    mockExec.mockResolvedValue(makeState({ phase: 1, trumpSuit: 3, declarerIdx: 1, highBid: 7 }));
    renderWithProviders(<IsraeliWhistPage />);
    const btn = await screen.findByTestId(`iw-bid-${bid.toString()}-btn`);
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', undefined, undefined, undefined, bid));
  });

  // **サーバが必ず拒否する宣言は押せない。** ノルマ未満と禁止値の両方。
  it('disables calls below the quota', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, trumpSuit: 3, declarerIdx: 0, highBid: 9, minimumBid: 9 }));
    renderWithProviders(<IsraeliWhistPage />);

    expect(await screen.findByTestId('iw-bid-8-btn')).toBeDisabled();
    expect(screen.getByTestId('iw-bid-9-btn')).toBeEnabled();
    expect(screen.getByTestId('iw-bid-0-btn')).toBeDisabled();
  });

  it('disables the barred call', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, trumpSuit: 3, declarerIdx: 1, highBid: 5, restrictedBid: 4 }));
    renderWithProviders(<IsraeliWhistPage />);

    expect(await screen.findByTestId('iw-bid-4-btn')).toBeDisabled();
    expect(screen.getByTestId('iw-bid-3-btn')).toBeEnabled();
    expect(screen.getByTestId('iw-bid-5-btn')).toBeEnabled();
  });

  it('enables every call when nothing is barred', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, trumpSuit: 3, declarerIdx: 1, highBid: 5 }));
    renderWithProviders(<IsraeliWhistPage />);
    expect(await screen.findByTestId('iw-bid-4-btn')).toBeEnabled();
    expect(screen.getByTestId('iw-bid-0-btn')).toBeEnabled();
  });

  it('plays the clicked card by its hand index', async () => {
    mockExec.mockResolvedValue(playing());
    renderWithProviders(<IsraeliWhistPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    mockExec.mockClear();
    fireEvent.click(cards[2]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 2));
  });

  // **2段階ぶんの状態を同時に出す。** 3 種の立場すべてを踏む。
  it('shows each seat with both its auction standing and its call', async () => {
    mockExec.mockResolvedValue(
      makeState({
        declarerIdx: 0,
        players: [
          seat(0, { auctionBid: 7, bid: 8, totalScore: 59 }),
          seat(1, { passed: true, bid: 2 }),
          seat(2),
          seat(3),
        ],
      } as Partial<IsraeliWhistResponse>),
    );
    renderWithProviders(<IsraeliWhistPage />);

    expect(await screen.findByTestId('iw-seat-0')).toHaveTextContent('落札 7');
    expect(screen.getByTestId('iw-seat-0')).toHaveTextContent('宣言8');
    expect(screen.getByTestId('iw-seat-0')).toHaveTextContent('59');
    expect(screen.getByTestId('iw-seat-1')).toHaveTextContent('降り');
    expect(screen.getByTestId('iw-seat-2')).toHaveTextContent('入札中');
    expect(screen.getByTestId('iw-seat-2')).toHaveTextContent('未宣言');
  });

  // オークション中と決着後で表示が入れ替わる。両側を踏む。
  it('shows the standing bid, then the trump', async () => {
    mockExec.mockResolvedValue(makeState({ highBid: 6, highSuit: 1 }));
    const { unmount } = renderWithProviders(<IsraeliWhistPage />);
    expect(await screen.findByTestId('iw-trump')).toHaveTextContent('♠');
    unmount();

    mockExec.mockResolvedValue(playing({ trumpSuit: 4 } as Partial<IsraeliWhistResponse>));
    renderWithProviders(<IsraeliWhistPage />);
    expect(await screen.findByTestId('iw-trump')).toHaveTextContent('♦');
  });

  it('advances the round when the next-round button is pressed', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 3 }));
    renderWithProviders(<IsraeliWhistPage />);

    const btn = await screen.findByRole('button', { name: '次のラウンドへ' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<IsraeliWhistPage />);
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
      const { unmount } = renderWithProviders(<IsraeliWhistPage />);
      expect(await screen.findByText(expected)).toBeInTheDocument();
      unmount();
    }
  });

  it('disables the hand while it is a CPU turn', async () => {
    mockExec.mockResolvedValue(playing({ currentPlayerIdx: 1 } as Partial<IsraeliWhistResponse>));
    renderWithProviders(<IsraeliWhistPage />);
    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    expect(cards[0]).toBeDisabled();
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'auction-7-3', reason: 'hint.israeliwhistAuctionBid', confidence: 'moderate' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<IsraeliWhistPage />);
    expect(await screen.findByText(/競り落とす価値があります/)).toBeInTheDocument();
  });
});
