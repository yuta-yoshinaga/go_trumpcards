import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { slobberhannesApi } from '../api/gameApi';
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
});

describe('SlobberhannesPage', () => {
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
    expect(screen.getByTestId('sh-seat-1')).toHaveTextContent('無傷');
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
