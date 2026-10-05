import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { rollingstoneApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, RollingStoneResponse } from '../types/card';
import { RollingStonePage } from './RollingStonePage';

vi.mock('../api/gameApi', () => ({
  rollingstoneApi: { exec: vi.fn() },
  actionLogApi: { rollingstone: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(rollingstoneApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const hand = [card('SPADE', 9), card('SPADE', 13), card('HEART', 10), card('CLOVER', 7)];

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 8,
  cards: id === 0 ? hand : [],
  pickups: 0,
  finishedAt: 0,
  ...over,
});

function makeState(overrides: Partial<RollingStoneResponse> = {}): RollingStoneResponse {
  return {
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    mustPickUp: false,
    validPlays: [0, 1],
    currentTrick: [],
    lastTrick: [],
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    trickNumber: 2,
    lastPickupIdx: -1,
    finishedCnt: 0,
    deckSize: 32,
    discarded: 8,
    gameEndFlag: false,
    winnerIdx: -1,
    winReason: '',
    config: { playerCnt: 4 },
    message: '',
    ...overrides,
  } as unknown as RollingStoneResponse;
}

/** The human cannot follow, so the only move is to take the trick. */
const forcedPickUp = (over: Partial<RollingStoneResponse> = {}) =>
  makeState({
    mustPickUp: true,
    validPlays: [],
    currentTrick: [{ playerIdx: 1, card: card('DIAMOND', 9) }],
    leadSuit: 4,
    ...over,
  } as Partial<RollingStoneResponse>);

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
});

describe('RollingStonePage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<RollingStonePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('announces cards added to the trick without duplicating the visible trick', async () => {
    mockExec.mockResolvedValueOnce(makeState());
    mockExec.mockResolvedValueOnce(
      makeState({
        currentTrick: [
          { playerIdx: 0, card: hand[0] },
          { playerIdx: 1, card: card('HEART', 11) },
        ],
        currentPlayerIdx: 2,
      }),
    );
    renderWithProviders(<RollingStonePage />);
    const playButton = await screen.findByRole('button', { name: '♠ 9 を出す（合法手）' });
    fireEvent.click(playButton);
    expect(await screen.findByTestId('rs-trick-announcement')).toHaveTextContent('あなた: ♠ 9、CPU1: ♥ J');
    expect(screen.getByTestId('rs-trick-announcement')).toHaveClass('sr-only');
  });

  it('announces the final card, resolution, and new trick when one response advances a trick', async () => {
    mockExec.mockResolvedValueOnce(
      makeState({
        currentTrick: [
          { playerIdx: 0, card: hand[0] },
          { playerIdx: 1, card: card('SPADE', 10) },
          { playerIdx: 2, card: card('SPADE', 11) },
        ],
        trickNumber: 2,
      }),
    );
    mockExec.mockResolvedValueOnce(
      makeState({
        currentTrick: [{ playerIdx: 2, card: card('HEART', 9) }],
        lastTrick: [
          { playerIdx: 0, card: hand[0] },
          { playerIdx: 1, card: card('SPADE', 10) },
          { playerIdx: 2, card: card('SPADE', 11) },
          { playerIdx: 3, card: card('SPADE', 12) },
        ],
        trickNumber: 3,
        currentPlayerIdx: 0,
      }),
    );
    renderWithProviders(<RollingStonePage />);
    fireEvent.click(await screen.findByRole('button', { name: '♠ 9 を出す（合法手）' }));
    await waitFor(() =>
      expect(screen.getByTestId('rs-trick-announcement')).toHaveTextContent(
        'CPU3: ♠ Q、トリックが解決しました。、CPU2: ♥ 9',
      ),
    );
  });

  // **勝利条件が逆さまなのが規則そのもの。**
  it('states that taking tricks is worth nothing', async () => {
    renderWithProviders(<RollingStonePage />);
    expect(await screen.findByTestId('rs-rule')).toHaveTextContent(/得点にはならず/);
  });

  // **デッキ枚数は人数で変わる。**
  it('shows the deck size and how much is still in play', async () => {
    renderWithProviders(<RollingStonePage />);
    const deck = await screen.findByTestId('rs-deck');
    expect(deck).toHaveTextContent('32');
    expect(deck).toHaveTextContent('24');
  });

  it('shows the finished count for an active and finished game', async () => {
    const { unmount } = renderWithProviders(<RollingStonePage />);
    expect(await screen.findByTestId('rs-finished-count')).toHaveTextContent('上がり 0 / 全 4 人');
    unmount();

    mockExec.mockResolvedValue(
      makeState({
        finishedCnt: 2,
        gameEndFlag: true,
        phase: 1,
        winnerIdx: 0,
        players: [seat(0, { cardCount: 0, finishedAt: 1 }), seat(1, { cardCount: 0, finishedAt: 2 }), seat(2), seat(3)],
      }),
    );
    renderWithProviders(<RollingStonePage />);
    expect(await screen.findByTestId('rs-finished-count')).toHaveTextContent('上がり 2 / 全 4 人');
    expect(screen.getByTestId('rs-result')).toHaveTextContent('あなたの勝ち！');
  });

  // **手札の枚数がそのまま順位。** 得点表示は無い。
  it('shows every hand size and pickup count', async () => {
    mockExec.mockResolvedValue(
      makeState({ players: [seat(0, { cardCount: 11, pickups: 2 }), seat(1), seat(2), seat(3)] }),
    );
    renderWithProviders(<RollingStonePage />);
    const s0 = await screen.findByTestId('rs-seat-0');
    expect(s0).toHaveTextContent('手札11枚');
    expect(s0).toHaveTextContent('引き取り2回');
    expect(screen.getByTestId('rs-seat-3')).toBeInTheDocument();
  });

  it('structures seat summaries as a list with a heading for each seat', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<RollingStonePage />);
    const list = await screen.findByRole('list');
    expect(list).toHaveClass('list-none');
    expect(list.querySelectorAll(':scope > li')).toHaveLength(4);
    expect(screen.getByRole('heading', { name: 'あなた', level: 2 })).toBeInTheDocument();
    expect(screen.getAllByRole('heading', { name: /CPU/, level: 2 })).toHaveLength(3);
  });

  // **引き取った席と上がった席は盤面に痕跡が残らない。**
  it('marks the last pickup and finishers', async () => {
    const { unmount } = renderWithProviders(<RollingStonePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByTestId('rs-seat-2')).not.toHaveTextContent(/引き取り\]/);
    unmount();

    mockExec.mockResolvedValue(makeState({ lastPickupIdx: 2 }));
    const second = renderWithProviders(<RollingStonePage />);
    expect(await screen.findByTestId('rs-seat-2')).toHaveTextContent(/直前に引き取り/);
    second.unmount();

    mockExec.mockResolvedValue(
      makeState({ players: [seat(0, { cardCount: 0, finishedAt: 1 }), seat(1), seat(2), seat(3)] }),
    );
    renderWithProviders(<RollingStonePage />);
    expect(await screen.findByTestId('rs-seat-0')).toHaveTextContent(/1位で上がり/);
  });

  it('plays the clicked card by its hand index', async () => {
    renderWithProviders(<RollingStonePage />);
    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    mockExec.mockClear();
    fireEvent.click(cards[1]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 1));
  });

  it('announces legal cards by name only during the human turn when pickup is not required', async () => {
    renderWithProviders(<RollingStonePage />);
    const cards = await screen.findAllByRole('button', { name: /出す/ });
    expect(cards[0]).toHaveAccessibleName('♠ 9 を出す（合法手）');
    expect(cards[1]).toHaveAccessibleName('♠ K を出す（合法手）');
    expect(cards[2]).toHaveAccessibleName('♥ 10 を出す');
    expect(cards[3]).toHaveAccessibleName('♣ 7 を出す');
  });

  // **出せる札が無い局面は、手札を押させずに引き取らせる。**
  it('offers only the pickup when you cannot follow', async () => {
    mockExec.mockImplementation(async (action) => (action === 'pickup' ? makeState() : forcedPickUp()));
    renderWithProviders(<RollingStonePage />);

    // **枚数だけでは、なぜ出せないのかが分からない** (#5764)。追従できなかった
    // スートまで書く。
    const banner = await screen.findByTestId('rs-must-pickup');
    expect(banner).toHaveTextContent('1');
    expect(banner).toHaveTextContent('♦');
    expect(banner).toHaveTextContent(/現在のトリック.*手札/);
    const pickup = screen.getByTestId('rs-pickup-btn');
    expect(pickup).toBeEnabled();
    const cards = screen.getAllByRole('button', { name: /を出す/ });
    expect(cards[0]).toBeDisabled();
    expect(cards[0]).toHaveAccessibleName('♠ 9 を出す');

    mockExec.mockClear();
    fireEvent.click(pickup);
    // **引き取りは別のコマンド。** cardIndex は送らない。
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pickup'));
    await waitFor(() => {
      expect(screen.queryByTestId('rs-must-pickup')).not.toBeInTheDocument();
      expect(screen.queryByTestId('rs-pickup-btn')).not.toBeInTheDocument();
    });
  });

  // 場が空のまま引き取りが立つことは規則上ありえないが、そのときでも
  // バナー自体は壊れない。
  it('falls back to a placeholder when no suit has been led', async () => {
    mockExec.mockResolvedValue(forcedPickUp({ leadSuit: 0, currentTrick: [] }));
    renderWithProviders(<RollingStonePage />);

    const banner = await screen.findByTestId('rs-must-pickup');
    expect(banner).toHaveTextContent('?');
    expect(banner).not.toHaveTextContent('♦');
  });

  // **負のコントロール: フォローできるなら引き取りは出さない。**
  it('hides the pickup button while you can still follow', async () => {
    renderWithProviders(<RollingStonePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('rs-pickup-btn')).not.toBeInTheDocument();
    expect(screen.queryByTestId('rs-must-pickup')).not.toBeInTheDocument();
  });

  it('disables the hand while it is a CPU turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1 }));
    renderWithProviders(<RollingStonePage />);
    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    expect(cards[0]).toBeDisabled();
    expect(cards[0]).toHaveAccessibleName('♠ 9 を出す');
    expect(screen.queryByTestId('rs-pickup-btn')).not.toBeInTheDocument();
  });

  // **上限で切った局は「上がった」わけではない。** 言い分ける。
  it('distinguishes running out from the stalemate', async () => {
    mockExec.mockResolvedValue(
      makeState({
        gameEndFlag: true,
        phase: 1,
        winnerIdx: 0,
        players: [seat(0, { cardCount: 0, finishedAt: 1 }), seat(1), seat(2), seat(3)],
      }),
    );
    const { unmount } = renderWithProviders(<RollingStonePage />);
    expect(await screen.findByTestId('rs-result')).toHaveTextContent(/先に手札を出し切りました/);
    unmount();

    mockExec.mockResolvedValue(
      makeState({
        gameEndFlag: true,
        phase: 1,
        winnerIdx: 2,
        winReason: 'stalemate',
        players: [seat(0), seat(1), seat(2, { cardCount: 3 }), seat(3)],
      }),
    );
    renderWithProviders(<RollingStonePage />);
    const banner = await screen.findByTestId('rs-result');
    expect(banner).toHaveTextContent(/決着が付かなかった/);
    expect(banner).toHaveTextContent('3');
  });

  // サーバが席の情報を欠いたレスポンスを返しても、膠着の説明が壊れず 0 枚として出る。
  it('reads zero cards when the winning seat is missing from the response', async () => {
    mockExec.mockResolvedValue(
      makeState({
        gameEndFlag: true,
        phase: 1,
        winnerIdx: 2,
        winReason: 'stalemate',
        players: [seat(0), seat(1)],
      }),
    );
    renderWithProviders(<RollingStonePage />);
    const banner = await screen.findByTestId('rs-result');
    expect(banner).toHaveTextContent(/決着が付かなかった/);
    expect(banner).toHaveTextContent('0');
  });

  // **投了も膠着も勝者に札が残る。** 枚数で分岐していたので、投了した局にも
  // 「決着が付かなかった」と出ていた。
  it('says the human resigned rather than calling it a stalemate', async () => {
    mockExec.mockResolvedValue(
      makeState({
        gameEndFlag: true,
        phase: 1,
        winnerIdx: 2,
        winReason: 'giveUp',
        players: [seat(0), seat(1), seat(2, { cardCount: 3 }), seat(3)],
      }),
    );
    renderWithProviders(<RollingStonePage />);
    const banner = await screen.findByTestId('rs-result');
    expect(banner).toHaveTextContent('あなたが投了しました');
    expect(banner).not.toHaveTextContent(/決着が付かなかった/);
  });

  it('reports a CPU running out first', async () => {
    mockExec.mockResolvedValue(
      makeState({
        gameEndFlag: true,
        phase: 1,
        winnerIdx: 1,
        players: [seat(0), seat(1, { cardCount: 0, finishedAt: 1 }), seat(2), seat(3)],
      }),
    );
    renderWithProviders(<RollingStonePage />);
    expect(await screen.findByTestId('rs-result')).toHaveTextContent(/CPU1 が先に出し切りました/);
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<RollingStonePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '投了' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'pickup', reason: 'hint.rollingstonePickUp', confidence: 'strong' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<RollingStonePage />);
    expect(await screen.findByTestId('hint-tooltip')).toHaveTextContent(/引き取るしかありません/);
  });
});
