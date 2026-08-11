import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { reversisApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, ReversisResponse } from '../types/card';
import { ReversisPage } from './ReversisPage';

vi.mock('../api/gameApi', () => ({
  reversisApi: { exec: vi.fn() },
  actionLogApi: { reversis: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(reversisApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 3,
  cards: id === 0 ? [card('HEART', 11), card('SPADE', 9), card('DIAMOND', 1)] : [],
  chips: 45,
  roundPenalty: 0,
  trickCount: 0,
  tookQuinola: false,
  tookDiamondAce: false,
  ...over,
});

function makeState(overrides: Partial<ReversisResponse> = {}): ReversisResponse {
  return {
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    roundNumber: 1,
    trickNumber: 2,
    pool: 20,
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
  } as unknown as ReversisResponse;
}

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
});

describe('ReversisPage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<ReversisPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('plays the clicked card by its hand index', async () => {
    renderWithProviders(<ReversisPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    mockExec.mockClear();
    fireEvent.click(cards[2]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 2));
  });

  // **プールと失点配分は盤面から読めない。** 常に出ていなければならない。
  it('always shows the pool and the penalty scale', async () => {
    renderWithProviders(<ReversisPage />);
    expect(await screen.findByTestId('rv-pool')).toHaveTextContent('20');
    expect(screen.getByTestId('rv-penalty-rule')).toHaveTextContent(/A=4 K=3 Q=2 J=1/);
    expect(screen.getByTestId('rv-penalty-rule')).toHaveTextContent(/♥J（キノラ）と♦A/);
  });

  it('tracks a growing pool', async () => {
    mockExec.mockResolvedValue(makeState({ pool: 45 }));
    renderWithProviders(<ReversisPage />);
    expect(await screen.findByTestId('rv-pool')).toHaveTextContent('45');
  });

  // 印付きの札を取ったかどうかが席ごとに出る。両側を踏む。
  it('shows which marked cards a seat has taken, and "clean" for one that has none', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          seat(0, { tookQuinola: true, roundPenalty: 7, chips: 40 }),
          seat(1, { tookDiamondAce: true }),
          seat(2),
          seat(3),
        ],
      } as Partial<ReversisResponse>),
    );
    renderWithProviders(<ReversisPage />);

    expect(await screen.findByTestId('rv-seat-0')).toHaveTextContent('♥J');
    expect(screen.getByTestId('rv-seat-0')).toHaveTextContent('40チップ / 失点7');
    expect(screen.getByTestId('rv-seat-1')).toHaveTextContent('♦A');
    expect(screen.getByTestId('rv-seat-2')).toHaveTextContent('無傷');
  });

  it('offers the next-round button only at a round end', async () => {
    renderWithProviders(<ReversisPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByRole('button', { name: '次のラウンドへ' })).not.toBeInTheDocument();
  });

  it('advances the round when the next-round button is pressed', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1 }));
    renderWithProviders(<ReversisPage />);

    const btn = await screen.findByRole('button', { name: '次のラウンドへ' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<ReversisPage />);
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
      mockExec.mockResolvedValue(makeState({ gameEndFlag: true, phase: 2, winnerIdx }));
      const { unmount } = renderWithProviders(<ReversisPage />);
      expect(await screen.findByText(expected)).toBeInTheDocument();
      unmount();
    }
  });

  it('disables the hand while it is a CPU turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1 }));
    renderWithProviders(<ReversisPage />);
    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    expect(cards[0]).toBeDisabled();
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'card-1', reason: 'hint.reversisAvoidMarked', confidence: 'strong' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<ReversisPage />);
    expect(await screen.findByText(/キノラか♦Aが乗っています/)).toBeInTheDocument();
  });
});
