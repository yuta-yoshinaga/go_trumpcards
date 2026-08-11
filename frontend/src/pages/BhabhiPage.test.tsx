import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { bhabhiApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BhabhiResponse, Card } from '../types/card';
import { BhabhiPage } from './BhabhiPage';

vi.mock('../api/gameApi', () => ({
  bhabhiApi: { exec: vi.fn() },
  actionLogApi: { bhabhi: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(bhabhiApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 3,
  cards: id === 0 ? [card('HEART', 10), card('SPADE', 9), card('CLOVER', 1)] : [],
  rank: -1,
  pickups: 0,
  ...over,
});

function makeState(overrides: Partial<BhabhiResponse> = {}): BhabhiResponse {
  return {
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    trickNumber: 0,
    leadSuit: 0,
    pile: [],
    lastPickupIdx: -1,
    lastPickupSize: 0,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    validPlays: [0, 1, 2],
    aliveCount: 4,
    gameEndFlag: false,
    bhabhiIdx: -1,
    stalemate: false,
    stalemateTricks: 300,
    config: { playerCnt: 4 },
    message: '',
    ...overrides,
  } as unknown as BhabhiResponse;
}

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
});

describe('BhabhiPage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<BhabhiPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **勝者ではなく敗者を決めるゲーム。** 目的が読めないと何をすべきか分からない。
  it('leads with the goal of not being the Bhabhi', async () => {
    renderWithProviders(<BhabhiPage />);
    expect(await screen.findByTestId('bh-goal')).toHaveTextContent(/Bhabhi/);
  });

  // **場札の枚数が罰の大きさそのもの。** 未リードと確定の両側を踏む。
  it('shows the pile size once somebody has led', async () => {
    const { unmount } = renderWithProviders(<BhabhiPage />);
    expect(await screen.findByTestId('bh-pile')).toHaveTextContent(/場は空/);
    unmount();

    mockExec.mockResolvedValue(
      makeState({
        leadSuit: 3,
        pile: [
          { playerIdx: 1, card: card('HEART', 5) },
          { playerIdx: 2, card: card('HEART', 9) },
        ],
      } as Partial<BhabhiResponse>),
    );
    renderWithProviders(<BhabhiPage />);
    const pile = await screen.findByTestId('bh-pile');
    expect(pile).toHaveTextContent('♥');
    expect(pile).toHaveTextContent('2');
  });

  it('shows how many players are still in', async () => {
    mockExec.mockResolvedValue(makeState({ aliveCount: 2 }));
    renderWithProviders(<BhabhiPage />);
    const alive = await screen.findByTestId('bh-alive');
    expect(alive).toHaveTextContent('2');
    expect(alive).toHaveTextContent('4');
  });

  // **順位は上がった順であって強さではない。** 残っている席と区別が付くこと。
  it('distinguishes seats that are out from seats still holding cards', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [seat(0), seat(1, { rank: 1, cardCount: 0 }), seat(2, { pickups: 2 }), seat(3)],
      } as Partial<BhabhiResponse>),
    );
    renderWithProviders(<BhabhiPage />);
    expect(await screen.findByTestId('bh-seat-1')).toHaveTextContent(/上がり/);
    expect(screen.getByTestId('bh-seat-0')).not.toHaveTextContent(/上がり/);
    expect(screen.getByTestId('bh-seat-0')).toHaveTextContent('3');
    expect(screen.getByTestId('bh-seat-2')).toHaveTextContent('2');
  });

  it('plays the clicked card by its hand index', async () => {
    renderWithProviders(<BhabhiPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    mockExec.mockClear();
    fireEvent.click(cards[2]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 2));
  });

  // **直前の引き取りは盤面に痕跡が残らない。**
  it('reports the last pickup while the game runs', async () => {
    mockExec.mockResolvedValue(makeState({ lastPickupIdx: 2, lastPickupSize: 5 }));
    renderWithProviders(<BhabhiPage />);
    expect(await screen.findByTestId('bh-last-pickup')).toHaveTextContent('5');
  });

  it('drops the pickup line once the game is over', async () => {
    mockExec.mockResolvedValue(
      makeState({ lastPickupIdx: 2, lastPickupSize: 5, gameEndFlag: true, phase: 1, bhabhiIdx: 1 }),
    );
    renderWithProviders(<BhabhiPage />);
    await screen.findByTestId('bh-result');
    expect(screen.queryByTestId('bh-last-pickup')).not.toBeInTheDocument();
  });

  // **次のハンドへ進むボタンは無い。** 配り切りの 1 ゲームで終わる。
  it('offers no next-hand button', async () => {
    mockExec.mockResolvedValue(makeState({ gameEndFlag: true, phase: 1, bhabhiIdx: 1 }));
    renderWithProviders(<BhabhiPage />);
    await screen.findByTestId('bh-result');
    expect(screen.queryByRole('button', { name: /次の/ })).not.toBeInTheDocument();
    // 負のコントロール: リセットは出ている
    expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument();
  });

  // 終わり方は 3 通りあり、それぞれ別の文言になる。
  it('names who the Bhabhi is', async () => {
    mockExec.mockResolvedValue(makeState({ gameEndFlag: true, phase: 1, bhabhiIdx: 0 }));
    const { unmount } = renderWithProviders(<BhabhiPage />);
    expect(await screen.findByTestId('bh-result')).toHaveTextContent(/あなたが Bhabhi/);
    unmount();

    mockExec.mockResolvedValue(makeState({ gameEndFlag: true, phase: 1, bhabhiIdx: 2 }));
    renderWithProviders(<BhabhiPage />);
    expect(await screen.findByTestId('bh-result')).toHaveTextContent(/CPU2 が Bhabhi/);
  });

  // **膠着で終わったことは盤面から読めない。**
  it('says so when the game was cut short as deadlocked', async () => {
    mockExec.mockResolvedValue(
      makeState({ gameEndFlag: true, phase: 1, bhabhiIdx: 1, stalemate: true, trickNumber: 300 }),
    );
    renderWithProviders(<BhabhiPage />);
    const result = await screen.findByTestId('bh-result');
    expect(result).toHaveTextContent(/膠着/);
    expect(result).toHaveTextContent('300');
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<BhabhiPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '投了' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  // **人数はゲームの形そのものを変える。** 3〜7 人を選べる。
  it('deals a fresh game when the table size changes', async () => {
    renderWithProviders(<BhabhiPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const select = await screen.findByTestId('bh-player-cnt');
    mockExec.mockClear();
    fireEvent.change(select, { target: { value: '6' } });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { playerCnt: 6 }));
  });

  it('renders every seat the server sent', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [seat(0), seat(1), seat(2), seat(3), seat(4), seat(5), seat(6)],
        config: { playerCnt: 7 },
        aliveCount: 7,
      } as Partial<BhabhiResponse>),
    );
    renderWithProviders(<BhabhiPage />);
    for (const id of [0, 1, 2, 3, 4, 5, 6]) {
      expect(await screen.findByTestId(`bh-seat-${id.toString()}`)).toBeInTheDocument();
    }
  });

  it('disables the hand while it is a CPU turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1 }));
    renderWithProviders(<BhabhiPage />);
    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    expect(cards[0]).toBeDisabled();
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'card-0', reason: 'hint.bhabhiDumpHigh', confidence: 'strong' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<BhabhiPage />);
    expect(await screen.findByText(/高い札を落としましょう/)).toBeInTheDocument();
  });
});
