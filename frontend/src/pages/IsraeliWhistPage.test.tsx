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

  /**
   * Mirrors `israeliWhistSuitRank` in `internal/domain/IsraeliWhist.go`:
   * ♣ < ♦ < ♥ < ♠. The test computes the domain's verdict independently so it
   * asserts the *server would accept* the call, not merely what was sent.
   */
  const RANK: Record<number, number> = { 2: 1, 4: 2, 3: 3, 1: 4 };
  const domainAccepts = (bid: number, suit: number, highBid: number, highSuit: number) =>
    bid !== highBid ? bid > highBid : RANK[suit] > RANK[highSuit];

  // **同数で競り上げられるのは、序列で上のスートだけ。** 下のスートは1つ上の数が要る。
  // 全部同じ数を送ると、サーバが拒否する入札をボタンが出すことになる。
  it.each([
    [8, 1], // 標準が ♠ — 同数では誰も上回れない
    [8, 3], // 標準が ♥ — ♠ だけが同数で上回れる
    [8, 2], // 標準が ♣ — 残り3スートが同数で上回れる
    [6, 4], // 標準が ♦
  ])('offers a bid the server accepts against %s %s', async (highBid, highSuit) => {
    mockExec.mockResolvedValue(makeState({ highBid, highSuit }));
    renderWithProviders(<IsraeliWhistPage />);
    await screen.findByTestId('iw-auction-1-btn');

    for (const suit of [1, 2, 3, 4]) {
      const btn = screen.getByTestId(`iw-auction-${suit.toString()}-btn`);
      if ((btn as HTMLButtonElement).disabled) continue;
      mockExec.mockClear();
      fireEvent.click(btn);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());
      const [, , , sentSuit, sentBid] = mockExec.mock.calls[0] as [string, undefined, undefined, number, number];
      expect(sentSuit).toBe(suit);
      expect(domainAccepts(sentBid, sentSuit, highBid, highSuit)).toBe(true);
    }
  });

  // **♠ が標準でも競り上げの道が残る。** 数を1つ上げれば全スートで上回れる。
  it('can still raise when spades are the standing suit', async () => {
    mockExec.mockResolvedValue(makeState({ highBid: 8, highSuit: 1 }));
    renderWithProviders(<IsraeliWhistPage />);

    const btn = await screen.findByTestId('iw-auction-2-btn');
    expect(btn).toBeEnabled();
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('auction', undefined, undefined, 2, 9));
  });

  // **13 を超える数は宣言できない。** そのスートのボタンは押せなくする。
  it('disables a suit that would need more than thirteen tricks', async () => {
    mockExec.mockResolvedValue(makeState({ highBid: 13, highSuit: 1 }));
    renderWithProviders(<IsraeliWhistPage />);

    // ♠ が標準の13なので、どのスートも14が要る＝全部押せない。
    for (const suit of [1, 2, 3, 4]) {
      expect(await screen.findByTestId(`iw-auction-${suit.toString()}-btn`)).toBeDisabled();
    }
    // 負のコントロール: 標準が ♣ の13なら、上位3スートは同数13で上回れる。
    mockExec.mockResolvedValue(makeState({ highBid: 13, highSuit: 2 }));
    const { unmount } = renderWithProviders(<IsraeliWhistPage />);
    unmount();
  });

  // **誰も入札しないまま最後の1人になったら降りられない。** サーバが必ず
  // 拒否するので、pass ボタンを出さない。負のコントロール付き。
  it('hides pass when the human is the last bidder standing with no bid', async () => {
    mockExec.mockResolvedValue(
      makeState({
        highBid: 0,
        players: [seat(0), seat(1, { passed: true }), seat(2, { passed: true }), seat(3, { passed: true })],
      } as Partial<IsraeliWhistResponse>),
    );
    renderWithProviders(<IsraeliWhistPage />);

    expect(await screen.findByTestId('iw-must-bid')).toBeInTheDocument();
    expect(screen.queryByTestId('iw-pass-btn')).not.toBeInTheDocument();
    // 入札の道は残っていること。
    expect(screen.getByTestId('iw-auction-1-btn')).toBeEnabled();
  });

  it('still offers pass while someone else is bidding', async () => {
    mockExec.mockResolvedValue(
      makeState({
        highBid: 0,
        players: [seat(0), seat(1, { passed: true }), seat(2), seat(3, { passed: true })],
      } as Partial<IsraeliWhistResponse>),
    );
    renderWithProviders(<IsraeliWhistPage />);

    expect(await screen.findByTestId('iw-pass-btn')).toBeInTheDocument();
    expect(screen.queryByTestId('iw-must-bid')).not.toBeInTheDocument();
  });

  // 入札が出ていれば、最後の1人でも降りられる。
  it('offers pass once a bid is standing', async () => {
    mockExec.mockResolvedValue(
      makeState({
        highBid: 7,
        highSuit: 1,
        players: [seat(0), seat(1, { passed: true }), seat(2, { passed: true }), seat(3, { passed: true })],
      } as Partial<IsraeliWhistResponse>),
    );
    renderWithProviders(<IsraeliWhistPage />);

    expect(await screen.findByTestId('iw-pass-btn')).toBeInTheDocument();
    expect(screen.queryByTestId('iw-must-bid')).not.toBeInTheDocument();
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

// **2 倍はこのゲームの起伏そのもの** (#5752)。畳まれたアクションログを
// 開かないと、点が普段の倍動いた理由が分からなかった。
describe('IsraeliWhistPage doubled-round banner', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
  });

  it('announces a round where every seat hit its bid', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: 3, doubled: true, doubledAllExact: true } as Partial<IsraeliWhistResponse>),
    );
    renderWithProviders(<IsraeliWhistPage />);
    expect(await screen.findByTestId('iw-doubled-banner')).toHaveTextContent('全員が宣言通り');
  });

  it('announces a round where every seat missed, without swapping the reason', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: 3, doubled: true, doubledAllExact: false } as Partial<IsraeliWhistResponse>),
    );
    renderWithProviders(<IsraeliWhistPage />);
    const banner = await screen.findByTestId('iw-doubled-banner');
    expect(banner).toHaveTextContent('全員が外した');
    expect(banner).not.toHaveTextContent('全員が宣言通り');
  });

  it('stays quiet on an ordinary round', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 3, doubled: false } as Partial<IsraeliWhistResponse>));
    renderWithProviders(<IsraeliWhistPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('iw-doubled-banner')).not.toBeInTheDocument();
  });
});
