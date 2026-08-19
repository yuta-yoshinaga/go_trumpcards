import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { honeymoonbridgeApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, HoneymoonBridgeResponse } from '../types/card';
import { HoneymoonBridgePage } from './HoneymoonBridgePage';

vi.mock('../api/gameApi', () => ({
  honeymoonbridgeApi: { exec: vi.fn() },
  actionLogApi: { honeymoonbridge: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(honeymoonbridgeApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 5,
  cards: id === 0 ? [card('HEART', 1), card('SPADE', 9), card('CLOVER', 13), card('DIAMOND', 4), card('SPADE', 2)] : [],
  bidLevel: 0,
  bidSuit: 0,
  trickCount: 0,
  score: 0,
  ...over,
});

function makeState(overrides: Partial<HoneymoonBridgeResponse> = {}): HoneymoonBridgeResponse {
  return {
    players: [seat(0), seat(1)],
    phase: 0,
    roundNumber: 1,
    trickNumber: 0,
    stockSize: 26,
    trumpSuit: 0,
    declarerIdx: -1,
    contractLevel: 0,
    requiredTricks: 0,
    minBidLevel: 0,
    minBidSuit: 0,
    lastMade: false,
    lastTricks: 0,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    dealerIdx: 0,
    currentTrick: [],
    validPlays: [0, 1, 2],
    gameEndFlag: false,
    winnerIdx: -1,
    config: { target: 100 },
    message: '',
    ...overrides,
  } as unknown as HoneymoonBridgeResponse;
}

/** The human's turn in the auction, with nobody having bid yet. */
const bidding = (over: Partial<HoneymoonBridgeResponse> = {}) =>
  makeState({ phase: 1, stockSize: 0, minBidLevel: 1, minBidSuit: 1, ...over } as Partial<HoneymoonBridgeResponse>);

/** The human's turn playing out a settled contract. */
const playing = (over: Partial<HoneymoonBridgeResponse> = {}) =>
  makeState({
    phase: 2,
    stockSize: 0,
    trumpSuit: 3,
    declarerIdx: 0,
    contractLevel: 2,
    requiredTricks: 8,
    ...over,
  } as Partial<HoneymoonBridgeResponse>);

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
});

describe('HoneymoonBridgePage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<HoneymoonBridgePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **前半のトリックは得点にならない。** これが読めないと打ち方を間違える。
  it('says the first half does not score', async () => {
    renderWithProviders(<HoneymoonBridgePage />);
    expect(await screen.findByTestId('hb-rule')).toHaveTextContent(/勝っても得点にならず/);
    expect(await screen.findByTestId('hb-stock')).toHaveTextContent('26');
  });

  // **山札は競りに入る前に尽きる。** 残りを出し続けると誤解を招く。
  it('hides the stock line once the auction starts', async () => {
    mockExec.mockResolvedValue(bidding());
    renderWithProviders(<HoneymoonBridgePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('hb-stock')).not.toBeInTheDocument();
  });

  it('shows the contract once it is bought, and the needed trick count', async () => {
    const { unmount } = renderWithProviders(<HoneymoonBridgePage />);
    expect(await screen.findByTestId('hb-contract')).toHaveTextContent(/未決定/);
    unmount();

    mockExec.mockResolvedValue(playing());
    renderWithProviders(<HoneymoonBridgePage />);
    const contract = await screen.findByTestId('hb-contract');
    expect(contract).toHaveTextContent('♥');
    // **「2♥」だけでは何トリック要るか分からない。**
    expect(contract).toHaveTextContent('8');
  });

  // **サーバが必ず拒否する値を出させない。** 通る最小の宣言を明示する。
  it('names the lowest bid that outbids the table', async () => {
    mockExec.mockResolvedValue(bidding({ minBidLevel: 3, minBidSuit: 3 } as Partial<HoneymoonBridgeResponse>));
    renderWithProviders(<HoneymoonBridgePage />);
    const line = await screen.findByTestId('hb-minbid');
    expect(line).toHaveTextContent('3');
    expect(line).toHaveTextContent('♥');
  });

  // **ページは競りの序列を作り直さない。** どのボタンもドメインが受理する値を送る。
  // 押せる宣言はすべて `outbids` を満たす（[[feedback_page_rederives_a_domain_rule]]）。
  it.each([
    [1, 1],
    [2, 3],
    [4, 0],
    [7, 4],
  ])('only offers bids the domain accepts at minimum %s/%s', async (minBidLevel, minBidSuit) => {
    mockExec.mockResolvedValue(bidding({ minBidLevel, minBidSuit } as Partial<HoneymoonBridgeResponse>));
    renderWithProviders(<HoneymoonBridgePage />);
    await screen.findByTestId('hb-level-select');

    // ドメインの規則をテスト側で独立に計算する（送信の記録を見るだけでは合法性を見ていない）。
    const rank = (s: number) => (s === 0 ? 5 : s);
    const domainAccepts = (level: number, suit: number) =>
      level !== minBidLevel ? level > minBidLevel : rank(suit) >= rank(minBidSuit);

    for (const suit of [1, 2, 3, 4, 0]) {
      const btn = screen.getByTestId(`hb-bid-${suit.toString()}-btn`);
      // 選択レベルは既定で minBidLevel。
      expect(btn.hasAttribute('disabled')).toBe(!domainAccepts(minBidLevel, suit));
    }
  });

  // **NT はいちばん強い。** 同レベルでも ♦ の上に置ける。
  it('offers no-trump at the minimum level even when the minimum suit is diamonds', async () => {
    mockExec.mockResolvedValue(bidding({ minBidLevel: 2, minBidSuit: 4 } as Partial<HoneymoonBridgeResponse>));
    renderWithProviders(<HoneymoonBridgePage />);
    expect(await screen.findByTestId('hb-bid-0-btn')).toBeEnabled();
    expect(screen.getByTestId('hb-bid-4-btn')).toBeEnabled();
    expect(screen.getByTestId('hb-bid-3-btn')).toBeDisabled();
  });

  // **レベルを上げれば下のスートも通る。**
  it('re-enables the weaker suits once you raise the level', async () => {
    mockExec.mockResolvedValue(bidding({ minBidLevel: 2, minBidSuit: 4 } as Partial<HoneymoonBridgeResponse>));
    renderWithProviders(<HoneymoonBridgePage />);
    expect(await screen.findByTestId('hb-bid-1-btn')).toBeDisabled();

    fireEvent.change(screen.getByTestId('hb-level-select'), { target: { value: '3' } });
    expect(screen.getByTestId('hb-bid-1-btn')).toBeEnabled();
  });

  // **上限に張り付いたら pass だけ。** レベルも選ばせない。
  it('offers only pass when nothing can outbid 7NT', async () => {
    mockExec.mockResolvedValue(bidding({ minBidLevel: 0, minBidSuit: 0 } as Partial<HoneymoonBridgeResponse>));
    renderWithProviders(<HoneymoonBridgePage />);
    expect(await screen.findByTestId('hb-minbid')).toHaveTextContent(/7NT/);
    expect(screen.queryByTestId('hb-level-select')).not.toBeInTheDocument();
    expect(screen.queryByTestId('hb-bid-0-btn')).not.toBeInTheDocument();
    expect(screen.getByTestId('hb-pass-btn')).toBeEnabled();
  });

  // **レベルとスートは 4・5 番目の引数で送る。** ずれると別の契約として届く。
  it('sends the selected level and suit', async () => {
    mockExec.mockResolvedValue(bidding({ minBidLevel: 2, minBidSuit: 1 } as Partial<HoneymoonBridgeResponse>));
    renderWithProviders(<HoneymoonBridgePage />);

    fireEvent.change(await screen.findByTestId('hb-level-select'), { target: { value: '4' } });
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('hb-bid-3-btn'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', undefined, undefined, 4, 3));
  });

  // **ノートランプは suit 0 で送る。** 「省略」と混ざらないこと。
  it('sends no-trump as suit zero', async () => {
    mockExec.mockResolvedValue(bidding());
    renderWithProviders(<HoneymoonBridgePage />);
    const btn = await screen.findByTestId('hb-bid-0-btn');
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', undefined, undefined, 1, 0));
  });

  it('passes when the pass button is pressed', async () => {
    mockExec.mockResolvedValue(bidding());
    renderWithProviders(<HoneymoonBridgePage />);
    const btn = await screen.findByTestId('hb-pass-btn');
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
  });

  it('hides the auction controls when it is the opponent bidding', async () => {
    mockExec.mockResolvedValue(bidding({ currentPlayerIdx: 1 } as Partial<HoneymoonBridgeResponse>));
    renderWithProviders(<HoneymoonBridgePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('hb-pass-btn')).not.toBeInTheDocument();
    expect(screen.queryByTestId('hb-level-select')).not.toBeInTheDocument();
  });

  it('shows each seat, and marks the declarer', async () => {
    mockExec.mockResolvedValue(playing({ declarerIdx: 1 } as Partial<HoneymoonBridgeResponse>));
    renderWithProviders(<HoneymoonBridgePage />);
    expect(await screen.findByTestId('hb-seat-1')).toHaveTextContent(/落札者/);
    expect(screen.getByTestId('hb-seat-0')).not.toHaveTextContent(/落札者/);
  });

  it("shows each seat's bid, or that they have none", async () => {
    mockExec.mockResolvedValue(
      playing({
        players: [seat(0, { bidLevel: 2, bidSuit: 3 }), seat(1)],
      } as Partial<HoneymoonBridgeResponse>),
    );
    renderWithProviders(<HoneymoonBridgePage />);
    expect(await screen.findByTestId('hb-seat-0')).toHaveTextContent('♥');
    expect(screen.getByTestId('hb-seat-1')).toHaveTextContent(/宣言なし/);
  });

  it('plays the clicked card by its hand index', async () => {
    mockExec.mockResolvedValue(playing());
    renderWithProviders(<HoneymoonBridgePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    mockExec.mockClear();
    fireEvent.click(cards[2]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 2));
  });

  // **引き合いでも札は出せる。** 得点にならないだけ。
  it('lets you play during the draw phase', async () => {
    renderWithProviders(<HoneymoonBridgePage />);
    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    expect(cards[0]).toBeEnabled();
  });

  it('disables the hand while it is a CPU turn', async () => {
    mockExec.mockResolvedValue(playing({ currentPlayerIdx: 1 } as Partial<HoneymoonBridgeResponse>));
    renderWithProviders(<HoneymoonBridgePage />);
    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    expect(cards[0]).toBeDisabled();
  });

  // ディールの結果は 3 通りすべて出す。
  it('reports the deal result', async () => {
    for (const [over, expected] of [
      [{ lastMade: true, lastTricks: 9 }, /成立/],
      [{ lastMade: false, lastTricks: 6 }, /失敗/],
      [{ contractLevel: 0, requiredTricks: 0, declarerIdx: -1 }, /流れました/],
    ] as const) {
      mockExec.mockResolvedValue(playing({ phase: 3, ...over } as Partial<HoneymoonBridgeResponse>));
      const { unmount } = renderWithProviders(<HoneymoonBridgePage />);
      expect(await screen.findByTestId('hb-round-result')).toHaveTextContent(expected);
      unmount();
    }
  });

  it('advances to the next deal when the button is pressed', async () => {
    mockExec.mockResolvedValue(playing({ phase: 3 } as Partial<HoneymoonBridgeResponse>));
    renderWithProviders(<HoneymoonBridgePage />);

    const btn = await screen.findByRole('button', { name: '次のディールへ' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<HoneymoonBridgePage />);
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
      const { unmount } = renderWithProviders(<HoneymoonBridgePage />);
      expect(await screen.findByTestId('hb-result')).toHaveTextContent(expected);
      unmount();
    }
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'bid-2-3', reason: 'hint.honeymoonbridgeBid', confidence: 'moderate' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<HoneymoonBridgePage />);
    expect(await screen.findByTestId('hint-tooltip')).toHaveTextContent(/長いスート/);
  });
});

// **得点式は細かい** (契約レベル×10 + オーバートリック×5 / 失敗は不足×10)。
// トリックの過不足だけでは、そのディールが何点だったのか読めない (#5760)。
describe('HoneymoonBridgePage round points', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
  });

  it('states what a made contract was worth, overtricks included', async () => {
    mockExec.mockResolvedValue(
      playing({
        phase: 3,
        contractLevel: 3,
        requiredTricks: 9,
        lastTricks: 11,
        lastMade: true,
        lastPoints: 40,
      } as Partial<HoneymoonBridgeResponse>),
    );
    renderWithProviders(<HoneymoonBridgePage />);
    const banner = await screen.findByTestId('hb-round-result');
    expect(banner).toHaveTextContent('+40点');
    expect(banner).toHaveTextContent('11');
  });

  it('states what going down handed the opponent', async () => {
    mockExec.mockResolvedValue(
      playing({
        phase: 3,
        contractLevel: 4,
        requiredTricks: 10,
        lastTricks: 7,
        lastMade: false,
        lastPoints: 30,
      } as Partial<HoneymoonBridgeResponse>),
    );
    renderWithProviders(<HoneymoonBridgePage />);
    const banner = await screen.findByTestId('hb-round-result');
    expect(banner).toHaveTextContent('+30点');
    expect(banner).toHaveTextContent('相手に');
  });
});
