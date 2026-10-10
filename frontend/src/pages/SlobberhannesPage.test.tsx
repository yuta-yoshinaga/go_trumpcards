import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, slobberhannesApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, SlobberhannesResponse } from '../types/card';
import { SlobberhannesPage } from './SlobberhannesPage';

vi.mock('../api/gameApi', () => ({
  slobberhannesApi: { exec: vi.fn() },
  actionLogApi: { slobberhannes: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(slobberhannesApi.exec);
const mockActionLog = vi.mocked(actionLogApi.slobberhannes);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 3,
  cards: id === 0 ? [card('SPADE', 1), card('HEART', 10), card('CLOVER', 12)] : [],
  score: 0,
  trickCount: 0,
  tookFirstTrick: false,
  tookLastTrick: false,
  tookQueen: false,
  ...over,
});

function makeState(overrides: Partial<SlobberhannesResponse> = {}): SlobberhannesResponse {
  return {
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    roundNumber: 1,
    trickNumber: 3,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    dealerIdx: 0,
    currentTrick: [],
    lastTrick: [],
    lastTrickWinner: -1,
    validPlays: [0, 1, 2],
    gameEndFlag: false,
    winnerIdx: -1,
    config: { rounds: 4 },
    message: '',
    ...overrides,
  } as unknown as SlobberhannesResponse;
}

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
  mockActionLog.mockResolvedValue({
    entries: [{ turnNumber: 0, playerIdx: -1, actionType: 'deal', detail: '配札を開始' }],
  });
});

describe('SlobberhannesPage', () => {
  it('opens the action log during play and shows its entries', async () => {
    renderWithProviders(<SlobberhannesPage />);
    const button = await screen.findByRole('button', { name: '棋譜を見る' });
    fireEvent.click(button);
    expect(await screen.findByText(/配札を開始/)).toBeInTheDocument();
  });
  it('shows the resolved trick and its winner until the next trick has cards', async () => {
    const trick = [0, 1, 2, 3].map((playerIdx) => ({ playerIdx, card: card('spade', 7 + playerIdx) }));
    mockExec.mockResolvedValue(makeState({ lastTrick: trick, lastTrickWinner: 2 }));
    const { unmount } = renderWithProviders(<SlobberhannesPage />);
    expect(await screen.findAllByTestId('trick-display-cards')).toHaveLength(1);
    expect(screen.getByTestId('trick-display-cards').children).toHaveLength(4);
    expect(screen.getAllByTestId('trick-winner-badge')).toHaveLength(1);

    mockExec.mockResolvedValue(
      makeState({
        currentTrick: [trick[0]],
        lastTrick: trick,
        lastTrickWinner: 2,
      }),
    );
    unmount();
    renderWithProviders(<SlobberhannesPage />);
    expect(await screen.findByTestId('trick-display-cards')).toBeInTheDocument();
    expect(screen.getByTestId('trick-display-cards').children).toHaveLength(1);
    expect(screen.queryByTestId('trick-winner-badge')).not.toBeInTheDocument();
  });

  it('does not show a previous trick after round end or game end', async () => {
    const trick = [0, 1, 2, 3].map((playerIdx) => ({ playerIdx, card: card('spade', 7 + playerIdx) }));
    mockExec.mockResolvedValue(makeState({ phase: 1, lastTrick: trick, lastTrickWinner: 2 }));
    const { unmount } = renderWithProviders(<SlobberhannesPage />);
    await waitFor(() => expect(screen.queryByTestId('trick-display-cards')).not.toBeInTheDocument());
    mockExec.mockResolvedValue(makeState({ phase: 2, gameEndFlag: true, lastTrick: trick, lastTrickWinner: 2 }));
    unmount();
    renderWithProviders(<SlobberhannesPage />);
    expect(screen.queryByTestId('trick-display-cards')).not.toBeInTheDocument();
    expect(screen.queryByTestId('trick-winner-badge')).not.toBeInTheDocument();
  });

  it('keeps the original card label off turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1 }));
    renderWithProviders(<SlobberhannesPage />);
    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    expect(cards[0]).toHaveAccessibleName(/♠ A を出す/);
    expect(cards[0]).not.toHaveAccessibleName(/出せる|出せない/);
  });

  it('announces playable and unplayable cards from validPlays', async () => {
    mockExec.mockResolvedValue(makeState({ validPlays: [1] }));
    renderWithProviders(<SlobberhannesPage />);
    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    expect(cards).toHaveLength(3);
    expect(cards[0]).toHaveAccessibleName(/出せない/);
    expect(cards[1]).toHaveAccessibleName(/出せる/);
    expect(cards[2]).toHaveAccessibleName(/出せない/);
  });
  it('resets on mount', async () => {
    renderWithProviders(<SlobberhannesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('plays the clicked card by its hand index', async () => {
    renderWithProviders(<SlobberhannesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    mockExec.mockClear();
    fireEvent.click(cards[2]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 2));
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<SlobberhannesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '投了' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  // **位置による罰は盤面に出ない情報。** 最初と最後のトリックでだけ警告し、
  // 中間では出さない。三方向とも踏む。
  it('warns on the first trick', async () => {
    mockExec.mockResolvedValue(makeState({ trickNumber: 0 }));
    renderWithProviders(<SlobberhannesPage />);
    expect(await screen.findByTestId('sh-position-warning')).toHaveTextContent('最初のトリック');
  });

  it('warns on the last trick', async () => {
    mockExec.mockResolvedValue(makeState({ trickNumber: 7 }));
    renderWithProviders(<SlobberhannesPage />);
    expect(await screen.findByTestId('sh-position-warning')).toHaveTextContent('最後のトリック');
  });

  it('stays silent on a middle trick', async () => {
    mockExec.mockResolvedValue(makeState({ trickNumber: 3 }));
    renderWithProviders(<SlobberhannesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('sh-position-warning')).not.toBeInTheDocument();
  });

  // 罰の内訳が席ごとに出る。無傷とそうでない場合の両方。
  it('shows each seat penalty marks, and "clean" when it has none', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [seat(0, { tookFirstTrick: true, tookQueen: true, score: -2 }), seat(1), seat(2), seat(3)],
      } as Partial<SlobberhannesResponse>),
    );
    renderWithProviders(<SlobberhannesPage />);

    expect(await screen.findByTestId('sh-seat-0')).toHaveTextContent('初');
    expect(screen.getByTestId('sh-seat-0')).toHaveTextContent('♣Q');
    expect(screen.getByTestId('sh-seat-0')).toHaveTextContent('-2点');
    // **「無傷」は失点なしではなく +1 の加点。**点数まで出さないと区別できない。
    expect(screen.getByTestId('sh-seat-1')).toHaveTextContent('無傷(+1点)');
    // 罰を受けた席にはボーナス表示が出ない。
    expect(screen.getByTestId('sh-seat-0')).not.toHaveTextContent('無傷');
  });

  it('keeps score and penalty information visible in compact seat rows', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<SlobberhannesPage />);

    expect(await screen.findByTestId('sh-seat-0')).toHaveTextContent('あなた');
    expect(screen.getByTestId('sh-seat-0')).toHaveTextContent('0点');
    expect(screen.getByTestId('sh-seat-0')).toHaveTextContent('無傷(+1点)');
    expect(screen.getByTestId('sh-seat-0')).toHaveClass('px-2', 'py-1', 'text-xs');
  });

  it('shows the current round trick counts with scores at round and game end only', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 1,
        players: [
          seat(0, { score: -2, trickCount: 3, tookFirstTrick: true }),
          seat(1, { score: 4, trickCount: 5 }),
          seat(2, { trickCount: 0 }),
          seat(3, { trickCount: 0 }),
        ],
      }),
    );
    const { unmount } = renderWithProviders(<SlobberhannesPage />);
    expect(await screen.findByTestId('sh-seat-0')).toHaveTextContent('3トリック');
    expect(screen.getByTestId('sh-seat-0')).toHaveTextContent('-2点');
    expect(screen.getByTestId('sh-seat-1')).toHaveTextContent('5トリック');
    unmount();

    mockExec.mockResolvedValue(makeState({ players: [seat(0, { trickCount: 1 }), seat(1), seat(2), seat(3)] }));
    const activeRender = renderWithProviders(<SlobberhannesPage />);
    expect(await screen.findByTestId('sh-seat-0')).toBeInTheDocument();
    expect(screen.getByTestId('sh-seat-0')).not.toHaveTextContent('トリック');
    activeRender.unmount();

    mockExec.mockResolvedValue(
      makeState({
        phase: 2,
        gameEndFlag: true,
        players: [seat(0, { score: -3, trickCount: 2 }), seat(1, { score: 7, trickCount: 6 }), seat(2), seat(3)],
      }),
    );
    renderWithProviders(<SlobberhannesPage />);
    expect(await screen.findByTestId('sh-seat-0')).toHaveTextContent('2トリック');
    expect(screen.getByTestId('sh-seat-0')).toHaveTextContent('-3点');
  });

  it('marks only the dealer seat', async () => {
    mockExec.mockResolvedValue(makeState({ dealerIdx: 2 }));
    renderWithProviders(<SlobberhannesPage />);

    expect(await screen.findByTestId('sh-seat-2')).toHaveTextContent('親');
    expect(screen.getByTestId('sh-seat-0')).not.toHaveTextContent('親');
    expect(screen.getByTestId('sh-seat-1')).not.toHaveTextContent('親');
    expect(screen.getByTestId('sh-seat-3')).not.toHaveTextContent('親');
  });

  it('moves the dealer mark when dealerIdx changes', async () => {
    mockExec.mockResolvedValue(makeState({ dealerIdx: 1 }));
    renderWithProviders(<SlobberhannesPage />);

    expect(await screen.findByTestId('sh-seat-1')).toHaveTextContent('親');
    expect(screen.getByTestId('sh-seat-0')).not.toHaveTextContent('親');
    expect(screen.getByTestId('sh-seat-2')).not.toHaveTextContent('親');
  });

  // 次のラウンドへは、ラウンド終了時にだけ現れる。
  it('offers the next-round button only at a round end', async () => {
    renderWithProviders(<SlobberhannesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByRole('button', { name: '次のラウンドへ' })).not.toBeInTheDocument();
  });

  it('advances the round when the next-round button is pressed', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1 }));
    renderWithProviders(<SlobberhannesPage />);

    const btn = await screen.findByRole('button', { name: '次のラウンドへ' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('renders the result banner for each outcome', async () => {
    for (const [winnerIdx, expected] of [
      [0, /あなたの勝ち/],
      [1, /CPU1 の勝ち/],
      [-1, /同点/],
    ] as const) {
      mockExec.mockResolvedValue(makeState({ gameEndFlag: true, phase: 2, winnerIdx }));
      const { unmount } = renderWithProviders(<SlobberhannesPage />);
      expect(await screen.findByText(expected)).toBeInTheDocument();
      unmount();
    }
  });

  it('disables the hand while it is a CPU turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1 }));
    renderWithProviders(<SlobberhannesPage />);
    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    expect(cards[0]).toBeDisabled();
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'card-1', reason: 'hint.slobberhannesAvoid', confidence: 'strong' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<SlobberhannesPage />);
    expect(await screen.findByText(/取らないように/)).toBeInTheDocument();
  });
});

// **♣Q は位置ではなく中身の罰点** (#5745)。最初/最後のトリックは
// trickNumber から警告できるが、♣Q は「今場に出ているか」がリスクの本体で、
// これまでは取ってから penaltyMarks で気づくしかなかった。
describe('SlobberhannesPage queen of clubs warning', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
  });

  it('warns while the queen is on the table', async () => {
    mockExec.mockResolvedValue(
      makeState({
        currentTrick: [{ playerIdx: 1, card: card('CLOVER', 12) }],
      } as unknown as Partial<SlobberhannesResponse>),
    );
    renderWithProviders(<SlobberhannesPage />);
    expect(await screen.findByTestId('sh-queen-warning')).toHaveTextContent('♣Q が場に出ています');
  });

  it('warns that the queen location is unknown before it appears', async () => {
    mockExec.mockResolvedValue(
      makeState({ players: [seat(0, { cards: [card('SPADE', 1)] }), seat(1), seat(2), seat(3)] }),
    );
    renderWithProviders(<SlobberhannesPage />);
    expect(await screen.findByTestId('sh-queen-unseen-warning')).toHaveTextContent('♣Q はまだ場に出ていません');
    expect(screen.queryByTestId('sh-queen-warning')).not.toBeInTheDocument();
  });

  it('does not call the queen unseen when the human holds it', async () => {
    mockExec.mockResolvedValue(
      makeState({ players: [seat(0, { cards: [card('CLOVER', 12)] }), seat(1), seat(2), seat(3)] }),
    );
    renderWithProviders(<SlobberhannesPage />);
    expect(await screen.findByText(/♣Q を持っています/)).toBeInTheDocument();
    expect(screen.queryByTestId('sh-queen-unseen-warning')).not.toBeInTheDocument();
  });

  it('does not show the unknown-location warning after the queen is taken', async () => {
    mockExec.mockResolvedValue(makeState({ players: [seat(0), seat(1, { tookQueen: true }), seat(2), seat(3)] }));
    renderWithProviders(<SlobberhannesPage />);
    await screen.findByTestId('sh-seat-0');
    expect(screen.queryByTestId('sh-queen-unseen-warning')).not.toBeInTheDocument();
  });

  it('stays quiet for any other card on the table', async () => {
    mockExec.mockResolvedValue(
      makeState({
        currentTrick: [{ playerIdx: 1, card: card('SPADE', 12) }],
      } as unknown as Partial<SlobberhannesResponse>),
    );
    renderWithProviders(<SlobberhannesPage />);
    await waitFor(() => expect(screen.getByTestId('sh-seat-0')).toBeInTheDocument());
    expect(screen.queryByTestId('sh-queen-warning')).not.toBeInTheDocument();
  });

  // 位置の警告と同時に出ても壊れない (受け入れ条件4)。
  it('shows the position warning alongside it on the first trick', async () => {
    mockExec.mockResolvedValue(
      makeState({
        trickNumber: 0,
        currentTrick: [{ playerIdx: 1, card: card('CLOVER', 12) }],
      } as unknown as Partial<SlobberhannesResponse>),
    );
    renderWithProviders(<SlobberhannesPage />);
    expect(await screen.findByTestId('sh-queen-warning')).toBeInTheDocument();
    expect(screen.getByTestId('sh-position-warning')).toHaveTextContent('最初のトリック');
  });

  it('marks the queen in the human hand', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<SlobberhannesPage />);
    // fixture の手札は ♠A / ♥10 / ♣Q。印が付くのは 1 枚だけ。
    expect(await screen.findByTestId('sh-queen-in-hand')).toBeInTheDocument();
    expect(screen.getAllByTestId('sh-queen-in-hand')).toHaveLength(1);
    const queenButton = screen.getByRole('button', { name: /♣ Q を出す/ });
    expect(queenButton).toHaveAccessibleName(/♣Q を持っています/);
    expect(queenButton.className).toContain('outline-ds-error');
  });
});
