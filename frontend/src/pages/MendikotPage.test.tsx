import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { mendikotApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, MendikotResponse } from '../types/card';
import { MendikotPage } from './MendikotPage';

vi.mock('../api/gameApi', () => ({
  mendikotApi: { exec: vi.fn() },
  actionLogApi: { mendikot: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(mendikotApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  team: id % 2,
  cardCount: 3,
  cards: id === 0 ? [card('HEART', 10), card('SPADE', 9), card('CLOVER', 1)] : [],
  tens: 0,
  trickCount: 0,
  ...over,
});

function makeState(overrides: Partial<MendikotResponse> = {}): MendikotResponse {
  return {
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    handNumber: 1,
    trickNumber: 0,
    trumpSuit: 0,
    trumpChooserIdx: -1,
    tensInDeck: 4,
    teamTens: [0, 0],
    teamTricks: [0, 0],
    scores: [0, 0],
    lastHandWinner: -1,
    lastHandKind: '',
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    currentTrick: [],
    validPlays: [0, 1, 2],
    gameEndFlag: false,
    winnerTeam: -1,
    config: { target: 3 },
    message: '',
    ...overrides,
  } as unknown as MendikotResponse;
}

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
});

describe('MendikotPage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<MendikotPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **勝敗を決めるのは 10 の枚数。** 盤面から読めないので必ず出す。
  it('leads with the race for the four tens', async () => {
    mockExec.mockResolvedValue(makeState({ teamTens: [2, 1] } as Partial<MendikotResponse>));
    renderWithProviders(<MendikotPage />);

    const tens = await screen.findByTestId('md-tens');
    expect(tens).toHaveTextContent('2');
    expect(tens).toHaveTextContent('1');
    expect(tens).toHaveTextContent('4');
  });

  // **トリック数は 2-2 のときしか効かない。** 出すが主役ではない。
  it('shows the trick count as the tie-break', async () => {
    mockExec.mockResolvedValue(makeState({ teamTricks: [7, 3] } as Partial<MendikotResponse>));
    renderWithProviders(<MendikotPage />);
    const tricks = await screen.findByTestId('md-tricks');
    expect(tricks).toHaveTextContent('7');
    expect(tricks).toHaveTextContent('3');
  });

  // **切り札を選ぶボタンは存在しない。** 出すとサーバが受けない操作を勧めることになる。
  // 札のボタンは aria-label にスート記号を持つので、記号での否定検索は当たらない
  // ——「切り札」ラベルそのものと、送ったコマンドの両方で見る。
  it('offers no trump buttons at all', async () => {
    renderWithProviders(<MendikotPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    // 負のコントロール: 札のボタンは確かに描かれている
    expect(await screen.findAllByRole('button', { name: /を出す/ })).not.toHaveLength(0);
    expect(screen.queryByRole('button', { name: /^切り札/ })).not.toBeInTheDocument();

    // **切り札はスートを名指すコマンドで届かない。** 送るのは札のインデックスだけ。
    const cards = screen.getAllByRole('button', { name: /を出す/ });
    mockExec.mockClear();
    fireEvent.click(cards[0]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0));
    for (const call of mockExec.mock.calls) expect(call).toHaveLength(2);
  });

  // 切り札は未定と確定の両側を踏む。
  it('explains how trump gets decided, then names it', async () => {
    const { unmount } = renderWithProviders(<MendikotPage />);
    expect(await screen.findByTestId('md-trump')).toHaveTextContent(/未定/);
    unmount();

    mockExec.mockResolvedValue(makeState({ trumpSuit: 4, trumpChooserIdx: 2 } as Partial<MendikotResponse>));
    renderWithProviders(<MendikotPage />);
    expect(await screen.findByTestId('md-trump')).toHaveTextContent('♦');
    expect(screen.getByTestId('md-seat-2')).toHaveTextContent('切り札決定');
    expect(screen.getByTestId('md-seat-0')).not.toHaveTextContent('切り札決定');
  });

  it('plays the clicked card by its hand index', async () => {
    renderWithProviders(<MendikotPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    mockExec.mockClear();
    fireEvent.click(cards[2]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 2));
  });

  // **席ごとの 10 の枚数は、どちらが勝っているかの根拠そのもの。**
  it('shows each seat team and its tens', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [seat(0, { tens: 2, trickCount: 3 }), seat(1), seat(2, { tens: 1 }), seat(3)],
      } as Partial<MendikotResponse>),
    );
    renderWithProviders(<MendikotPage />);

    expect(await screen.findByTestId('md-seat-0')).toHaveTextContent('T0');
    expect(screen.getByTestId('md-seat-1')).toHaveTextContent('T1');
    expect(screen.getByTestId('md-seat-2')).toHaveTextContent('T0');
    expect(screen.getByTestId('md-seat-0')).toHaveTextContent('2');
    expect(screen.getByTestId('md-seat-0')).toHaveTextContent('3');
  });

  // **決まり方で 1/2/3 点と変わる。** 4 通りすべて別の文言になる。
  it.each([
    ['tens', /10 の枚数で勝ち/],
    ['tricks', /トリックの多い/],
    ['mendikot', /Mendikot/],
    ['whitewash', /Whitewash/],
  ])('explains a hand decided by %s', async (kind, expected) => {
    mockExec.mockResolvedValue(makeState({ phase: 1, lastHandWinner: 0, lastHandKind: kind }));
    const { unmount } = renderWithProviders(<MendikotPage />);
    expect(await screen.findByTestId('md-hand-result')).toHaveTextContent(expected);
    unmount();
  });

  it('shows the running hand points', async () => {
    mockExec.mockResolvedValue(makeState({ scores: [2, 1] } as Partial<MendikotResponse>));
    renderWithProviders(<MendikotPage />);
    expect(await screen.findByTestId('md-score')).toHaveTextContent('2');
    expect(screen.getByTestId('md-score')).toHaveTextContent('1');
  });

  it('advances to the next hand when the button is pressed', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, lastHandWinner: 0, lastHandKind: 'tens' }));
    renderWithProviders(<MendikotPage />);

    const btn = await screen.findByRole('button', { name: '次のハンドへ' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<MendikotPage />);
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
      mockExec.mockResolvedValue(makeState({ gameEndFlag: true, phase: 2, winnerTeam }));
      const { unmount } = renderWithProviders(<MendikotPage />);
      expect(await screen.findByText(expected)).toBeInTheDocument();
      unmount();
    }
  });

  it('disables the hand while it is a CPU turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1 } as Partial<MendikotResponse>));
    renderWithProviders(<MendikotPage />);
    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    expect(cards[0]).toBeDisabled();
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'card-0', reason: 'hint.mendikotChaseTen', confidence: 'strong' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<MendikotPage />);
    expect(await screen.findByText(/10 が場に出ています/)).toBeInTheDocument();
  });
});

// **切り札は宣言ではなく事故で決まる** (#5755)。フォローできない手番は
// ハンド全体を左右する一度きりの選択なのに、警告が無かった。
describe('MendikotPage trump-setting warning', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
  });

  it('warns when the human cannot follow and the trump is still open', async () => {
    mockExec.mockResolvedValue(makeState({ trumpSuit: 0, willSetTrump: true }));
    renderWithProviders(<MendikotPage />);
    expect(await screen.findByTestId('md-sets-trump-warning')).toHaveTextContent('切り札になります');
  });

  it('stays quiet when the human can follow', async () => {
    mockExec.mockResolvedValue(makeState({ trumpSuit: 0, willSetTrump: false }));
    renderWithProviders(<MendikotPage />);
    await waitFor(() => expect(screen.getByTestId('md-trump')).toBeInTheDocument());
    expect(screen.queryByTestId('md-sets-trump-warning')).not.toBeInTheDocument();
  });

  it('stays quiet once the trump is decided', async () => {
    mockExec.mockResolvedValue(makeState({ trumpSuit: 3, willSetTrump: false }));
    renderWithProviders(<MendikotPage />);
    await waitFor(() => expect(screen.getByTestId('md-trump')).toBeInTheDocument());
    expect(screen.queryByTestId('md-sets-trump-warning')).not.toBeInTheDocument();
  });
});
