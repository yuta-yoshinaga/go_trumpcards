import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { hasenpfefferApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, HasenpfefferResponse } from '../types/card';
import { HasenpfefferPage } from './HasenpfefferPage';

vi.mock('../api/gameApi', () => ({
  hasenpfefferApi: { exec: vi.fn() },
  actionLogApi: { hasenpfeffer: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(hasenpfefferApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  team: id % 2,
  cardCount: 3,
  cards: id === 0 ? [card('HEART', 11), card('SPADE', 9), card('CLOVER', 1)] : [],
  bid: -1,
  trickCount: 0,
  ...over,
});

function makeState(overrides: Partial<HasenpfefferResponse> = {}): HasenpfefferResponse {
  return {
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    handNumber: 1,
    trickNumber: 0,
    trumpSuit: 0,
    declarerIdx: -1,
    contract: 0,
    minBid: 3,
    mustBid: false,
    blindSize: 1,
    scores: [0, 0],
    teamTricks: [0, 0],
    lastHandEuchred: false,
    lastHandTricks: 0,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    dealerIdx: 3,
    currentTrick: [],
    validPlays: [0, 1, 2],
    gameEndFlag: false,
    winnerTeam: -1,
    config: { target: 10 },
    message: '',
    ...overrides,
  } as unknown as HasenpfefferResponse;
}

/** A state where trump is settled and it is the human's turn to play. */
const playing = (over: Partial<HasenpfefferResponse> = {}) =>
  makeState({
    phase: 2,
    trumpSuit: 3,
    declarerIdx: 1,
    contract: 4,
    blindSize: 0,
    ...over,
  } as Partial<HasenpfefferResponse>);

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
});

describe('HasenpfefferPage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<HasenpfefferPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **ジョーカーが最強という序列は知らないと打ち方が変わる。**
  it('states the joker ranking', async () => {
    renderWithProviders(<HasenpfefferPage />);
    expect(await screen.findByTestId('hpf-rule')).toHaveTextContent(/Best Bower/);
  });

  // **サーバが必ず拒否する額は出さない (#5304)。**
  it('offers only bids at or above the minimum', async () => {
    mockExec.mockResolvedValue(makeState({ minBid: 5 }));
    renderWithProviders(<HasenpfefferPage />);

    expect(await screen.findByTestId('hpf-bid-5-btn')).toBeInTheDocument();
    expect(screen.getByTestId('hpf-bid-6-btn')).toBeInTheDocument();
    for (const n of [3, 4]) {
      expect(screen.queryByTestId(`hpf-bid-${n.toString()}-btn`)).not.toBeInTheDocument();
    }
  });

  // **上限が立っていたら宣言できない。** 降りるボタンだけが残る。
  it('offers no bid at all once the maximum is standing', async () => {
    mockExec.mockResolvedValue(makeState({ minBid: 0 }));
    renderWithProviders(<HasenpfefferPage />);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    for (const n of [3, 4, 5, 6]) {
      expect(screen.queryByTestId(`hpf-bid-${n.toString()}-btn`)).not.toBeInTheDocument();
    }
    expect(screen.getByTestId('hpf-pass-btn')).toBeInTheDocument();
  });

  // **親が降りられない場面では降りるボタンを出さない。**
  it('hides the pass button when the dealer cannot pass', async () => {
    mockExec.mockResolvedValue(makeState({ mustBid: true }));
    renderWithProviders(<HasenpfefferPage />);

    expect(await screen.findByTestId('hpf-must-bid')).toBeInTheDocument();
    expect(screen.queryByTestId('hpf-pass-btn')).not.toBeInTheDocument();
    // 負のコントロール: 宣言ボタンは出ている
    expect(screen.getByTestId('hpf-bid-3-btn')).toBeInTheDocument();
  });

  // **宣言は5番目の引数で送る。** 位置がずれると別の値として届く。
  it.each([3, 4, 5, 6])('sends bid %s', async (n) => {
    renderWithProviders(<HasenpfefferPage />);
    const btn = await screen.findByTestId(`hpf-bid-${n.toString()}-btn`);
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', undefined, undefined, undefined, n));
  });

  it('sends a pass as bid 0', async () => {
    renderWithProviders(<HasenpfefferPage />);
    const btn = await screen.findByTestId('hpf-pass-btn');
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', undefined, undefined, undefined, 0));
  });

  // **捨て札は札を選んでからスートを選ぶ。** 選ぶ前は確定できない。
  it('requires a card before the trump buttons work', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: 1, declarerIdx: 0, contract: 4, blindSize: 0 } as Partial<HasenpfefferResponse>),
    );
    renderWithProviders(<HasenpfefferPage />);

    const suitBtn = await screen.findByTestId('hpf-discard-3-btn');
    expect(suitBtn).toBeDisabled();
    mockExec.mockClear();
    fireEvent.click(suitBtn);
    await waitFor(() => expect(mockExec).not.toHaveBeenCalled());

    const cards = screen.getAllByRole('button', { name: /捨て札に選ぶ/ });
    fireEvent.click(cards[1]);
    expect(cards[1]).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(screen.getByTestId('hpf-discard-3-btn'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', 1, undefined, 3));
  });

  it('plays the clicked card by its hand index', async () => {
    mockExec.mockResolvedValue(playing());
    renderWithProviders(<HasenpfefferPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    mockExec.mockClear();
    fireEvent.click(cards[2]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 2));
  });

  // **宣言の状態は 3 通り。**
  it('renders every bid state on the seats', async () => {
    mockExec.mockResolvedValue(
      playing({
        players: [seat(0, { bid: 4 }), seat(1, { bid: 0 }), seat(2), seat(3)],
      } as Partial<HasenpfefferResponse>),
    );
    renderWithProviders(<HasenpfefferPage />);

    expect(await screen.findByTestId('hpf-seat-0')).toHaveTextContent('4');
    expect(screen.getByTestId('hpf-seat-1')).toHaveTextContent(/降り/);
    expect(screen.getByTestId('hpf-seat-2')).toHaveTextContent(/未宣言/);
    expect(screen.getByTestId('hpf-seat-1')).toHaveTextContent('[落札]');
  });

  // 伏せ札・未宣言・確定の 3 状態を踏む。
  it('shows the blind, then trump', async () => {
    const { unmount } = renderWithProviders(<HasenpfefferPage />);
    expect(await screen.findByTestId('hpf-trump')).toHaveTextContent(/伏せ札/);
    unmount();

    mockExec.mockResolvedValue(playing({ trumpSuit: 4 } as Partial<HasenpfefferResponse>));
    renderWithProviders(<HasenpfefferPage />);
    expect(await screen.findByTestId('hpf-trump')).toHaveTextContent('♦');
  });

  // **落としたのか達成したのかは盤面から読めない。** 両側を踏む。
  it.each([
    [false, /達成/],
    [true, /落とし/],
  ])('explains how the hand ended (euchred=%s)', async (lastHandEuchred, expected) => {
    mockExec.mockResolvedValue(
      playing({ phase: 3, lastHandEuchred, lastHandTricks: 3 } as Partial<HasenpfefferResponse>),
    );
    const { unmount } = renderWithProviders(<HasenpfefferPage />);
    expect(await screen.findByTestId('hpf-hand-result')).toHaveTextContent(expected);
    unmount();
  });

  it('advances to the next hand when the button is pressed', async () => {
    mockExec.mockResolvedValue(playing({ phase: 3 } as Partial<HasenpfefferResponse>));
    renderWithProviders(<HasenpfefferPage />);

    const btn = await screen.findByRole('button', { name: '次のハンドへ' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<HasenpfefferPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '投了' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('renders the result banner for each outcome', async () => {
    for (const [winnerTeam, expected] of [
      [0, /あなたのチームの勝ち/],
      [1, /相手チームの勝ち/],
      [-1, /同点/],
    ] as const) {
      mockExec.mockResolvedValue(playing({ gameEndFlag: true, phase: 4, winnerTeam }));
      const { unmount } = renderWithProviders(<HasenpfefferPage />);
      expect(await screen.findByTestId('hpf-result')).toHaveTextContent(expected);
      unmount();
    }
  });

  it('disables the hand while it is a CPU turn', async () => {
    mockExec.mockResolvedValue(playing({ currentPlayerIdx: 1 } as Partial<HasenpfefferResponse>));
    renderWithProviders(<HasenpfefferPage />);
    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    expect(cards[0]).toBeDisabled();
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'bid-3', reason: 'hint.hasenpfefferMustBid', confidence: 'strong' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<HasenpfefferPage />);
    // **ルール帯にも同じ言い回しが出る。** ツールチップに絞って見る。
    expect(await screen.findByTestId('hint-tooltip')).toHaveTextContent(/最低額で受けましょう/);
  });
});
