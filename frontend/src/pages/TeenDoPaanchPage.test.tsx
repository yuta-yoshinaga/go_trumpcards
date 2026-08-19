import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { teendopaanchApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, TeenDoPaanchResponse } from '../types/card';
import { TeenDoPaanchPage } from './TeenDoPaanchPage';

vi.mock('../api/gameApi', () => ({
  teendopaanchApi: { exec: vi.fn() },
  actionLogApi: { teendopaanch: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(teendopaanchApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 3,
  cards: id === 0 ? [card('HEART', 1), card('SPADE', 9), card('CLOVER', 13)] : [],
  target: [5, 3, 2][id] ?? 0,
  trickCount: 0,
  met: 0,
  ...over,
});

function makeState(overrides: Partial<TeenDoPaanchResponse> = {}): TeenDoPaanchResponse {
  return {
    players: [seat(0), seat(1), seat(2)],
    phase: 0,
    roundNumber: 1,
    trickNumber: 0,
    trumpSuit: 0,
    fivePlayerIdx: 0,
    lastExchange: 0,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    currentTrick: [],
    validPlays: [0, 1, 2],
    gameEndFlag: false,
    winnerIdx: -1,
    config: { rounds: 3 },
    message: '',
    ...overrides,
  } as unknown as TeenDoPaanchResponse;
}

/** A state where trump is settled and it is the human's turn to play. */
const playing = (over: Partial<TeenDoPaanchResponse> = {}) =>
  makeState({ phase: 1, trumpSuit: 3, ...over } as Partial<TeenDoPaanchResponse>);

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
});

describe('TeenDoPaanchPage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<TeenDoPaanchPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **ノルマは宣言ではなく割り当て。** これが読めないと何をすべきか分からない。
  it('states that the targets are assigned rather than bid', async () => {
    renderWithProviders(<TeenDoPaanchPage />);
    expect(await screen.findByTestId('td-targets')).toHaveTextContent(/割り当て/);
  });

  // **あと何トリック要るかが読めないと打ち方が決まらない。**
  it('shows each seat target against the tricks it has taken', async () => {
    mockExec.mockResolvedValue(
      playing({
        players: [seat(0, { trickCount: 4, met: 2 }), seat(1), seat(2)],
      } as Partial<TeenDoPaanchResponse>),
    );
    renderWithProviders(<TeenDoPaanchPage />);

    const mine = await screen.findByTestId('td-seat-0');
    expect(mine).toHaveTextContent('5');
    expect(mine).toHaveTextContent('4');
    expect(mine).toHaveTextContent('2');
  });

  // **3・2・5 が必ず 1 つずつ出る。**
  it('renders all three targets', async () => {
    mockExec.mockResolvedValue(playing());
    renderWithProviders(<TeenDoPaanchPage />);
    for (const [id, target] of [
      [0, '5'],
      [1, '3'],
      [2, '2'],
    ] as const) {
      expect(await screen.findByTestId(`td-seat-${id}`)).toHaveTextContent(target);
    }
  });

  it('marks the seat that declares trump', async () => {
    mockExec.mockResolvedValue(playing({ fivePlayerIdx: 1 } as Partial<TeenDoPaanchResponse>));
    renderWithProviders(<TeenDoPaanchPage />);
    expect(await screen.findByTestId('td-seat-1')).toHaveTextContent('切り札決定');
    expect(screen.getByTestId('td-seat-0')).not.toHaveTextContent('切り札決定');
  });

  it('offers all four trump suits to the 5-target seat', async () => {
    renderWithProviders(<TeenDoPaanchPage />);
    for (const suit of [1, 2, 3, 4]) {
      expect(await screen.findByTestId(`td-trump-${suit.toString()}-btn`)).toBeInTheDocument();
    }
  });

  it('hides the trump buttons when the human does not owe five', async () => {
    mockExec.mockResolvedValue(makeState({ fivePlayerIdx: 2 }));
    renderWithProviders(<TeenDoPaanchPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('td-trump-1-btn')).not.toBeInTheDocument();
  });

  // **切り札は4番目の引数で送る。** 位置がずれると別の値として届く。
  it.each([1, 2, 3, 4])('sends trump suit %s', async (suit) => {
    renderWithProviders(<TeenDoPaanchPage />);
    const btn = await screen.findByTestId(`td-trump-${suit.toString()}-btn`);
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('trump', undefined, undefined, suit));
  });

  it('plays the clicked card by its hand index', async () => {
    mockExec.mockResolvedValue(playing());
    renderWithProviders(<TeenDoPaanchPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    mockExec.mockClear();
    fireEvent.click(cards[2]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 2));
  });

  // 切り札は未宣言と確定の両側を踏む。
  it('shows the trump once declared', async () => {
    const { unmount } = renderWithProviders(<TeenDoPaanchPage />);
    expect(await screen.findByTestId('td-trump')).toHaveTextContent(/未宣言/);
    unmount();

    mockExec.mockResolvedValue(playing({ trumpSuit: 4 } as Partial<TeenDoPaanchResponse>));
    renderWithProviders(<TeenDoPaanchPage />);
    expect(await screen.findByTestId('td-trump')).toHaveTextContent('♦');
  });

  // **前ラウンドの札のやり取りは盤面に痕跡が残らない。**
  it('reports the exchange only when cards actually moved', async () => {
    mockExec.mockResolvedValue(playing());
    const { unmount } = renderWithProviders(<TeenDoPaanchPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('td-exchange')).not.toBeInTheDocument();
    unmount();

    mockExec.mockResolvedValue(playing({ lastExchange: 2 } as Partial<TeenDoPaanchResponse>));
    renderWithProviders(<TeenDoPaanchPage />);
    expect(await screen.findByTestId('td-exchange')).toHaveTextContent('2');
  });

  it('shows the round out of the configured total', async () => {
    mockExec.mockResolvedValue(playing({ roundNumber: 2, config: { rounds: 6 } } as Partial<TeenDoPaanchResponse>));
    renderWithProviders(<TeenDoPaanchPage />);
    const round = await screen.findByTestId('td-round');
    expect(round).toHaveTextContent('2');
    expect(round).toHaveTextContent('6');
  });

  it('advances to the next round when the button is pressed', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2, trumpSuit: 3 }));
    renderWithProviders(<TeenDoPaanchPage />);

    const btn = await screen.findByRole('button', { name: '次のラウンドへ' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<TeenDoPaanchPage />);
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
      mockExec.mockResolvedValue(makeState({ gameEndFlag: true, phase: 3, winnerIdx }));
      const { unmount } = renderWithProviders(<TeenDoPaanchPage />);
      expect(await screen.findByTestId('td-result')).toHaveTextContent(expected);
      unmount();
    }
  });

  it('disables the hand while it is a CPU turn', async () => {
    mockExec.mockResolvedValue(playing({ currentPlayerIdx: 1 } as Partial<TeenDoPaanchResponse>));
    renderWithProviders(<TeenDoPaanchPage />);
    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    expect(cards[0]).toBeDisabled();
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'trump-3', reason: 'hint.teendopaanchSelectTrump', confidence: 'moderate' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<TeenDoPaanchPage />);
    expect(await screen.findByText(/いちばん長いスート/)).toBeInTheDocument();
  });
});

// **合計だけでは、自分の手札から何が抜かれたのか分からない** (#5757)。
// 誰の最強札が誰に渡ったのかがこのゲームの名物。
describe('TeenDoPaanchPage exchange breakdown', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
  });

  it('names both sides of every exchange', async () => {
    mockExec.mockResolvedValue(
      makeState({
        lastExchange: 3,
        lastExchangePairs: [
          { giver: 1, taker: 0, count: 2 },
          { giver: 2, taker: 0, count: 1 },
        ],
      } as Partial<TeenDoPaanchResponse>),
    );
    renderWithProviders(<TeenDoPaanchPage />);

    const line = await screen.findByTestId('td-exchange');
    // **複数ペアぶん全部出る** (受け入れ条件2)。
    expect(line).toHaveTextContent('CPU 1→あなた 2枚');
    expect(line).toHaveTextContent('CPU 2→あなた 1枚');
    // 合計も残る。
    expect(line).toHaveTextContent('3');
  });

  it('shows nothing when no cards moved', async () => {
    mockExec.mockResolvedValue(makeState({ lastExchange: 0, lastExchangePairs: [] } as Partial<TeenDoPaanchResponse>));
    renderWithProviders(<TeenDoPaanchPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('td-exchange')).not.toBeInTheDocument();
  });
});
