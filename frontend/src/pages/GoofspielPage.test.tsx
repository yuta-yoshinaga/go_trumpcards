import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { goofspielApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, GoofspielResponse } from '../types/card';
import { GoofspielPage } from './GoofspielPage';

vi.mock('../api/gameApi', () => ({
  goofspielApi: { exec: vi.fn() },
  actionLogApi: { goofspiel: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(goofspielApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 12,
  cards: id === 0 ? [card('SPADE', 1), card('SPADE', 2), card('SPADE', 13)] : [card('CLOVER', 1), card('CLOVER', 2)],
  score: 0,
  hasBid: false,
  ...over,
});

function makeState(overrides: Partial<GoofspielResponse> = {}): GoofspielResponse {
  return {
    players: [seat(0), seat(1)],
    phase: 0,
    validPlays: [0, 1, 2],
    currentPrize: card('DIAMOND', 9),
    carriedPrizes: [],
    prizeValue: 9,
    prizeRemaining: 11,
    lastWinnerIdx: -1,
    lastGained: 0,
    roundNumber: 2,
    gameEndFlag: false,
    winnerIdx: -1,
    config: { playerCnt: 2, tieRule: 0 },
    message: '',
    ...overrides,
  } as unknown as GoofspielResponse;
}

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
});

describe('GoofspielPage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<GoofspielPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **同時入札であることが規則そのもの。**
  it('states that everyone bids at the same time', async () => {
    renderWithProviders(<GoofspielPage />);
    expect(await screen.findByTestId('gs-rule')).toHaveTextContent(/同時に入札します/);
  });

  it('shows the prize and the points at stake', async () => {
    renderWithProviders(<GoofspielPage />);
    expect(await screen.findByTestId('gs-prize')).toHaveTextContent('9');
    expect(screen.queryByTestId('gs-carried')).not.toBeInTheDocument();
  });

  // **サーバが carriedPrizes を省いても落ちないこと。**
  //
  // 最初の実装はサーバが null を返し、ページが `.length` で落ちて skeleton が
  // 消えませんでした。手で `[]` を渡すページテストでは通ってしまい、E2E だけが
  // 検出した経路です。
  it('survives a response with no carriedPrizes field', async () => {
    const bare = makeState();
    delete (bare as { carriedPrizes?: unknown }).carriedPrizes;
    mockExec.mockResolvedValue(bare);
    renderWithProviders(<GoofspielPage />);
    expect(await screen.findByTestId('gs-prize')).toBeInTheDocument();
    expect(screen.queryByTestId('gs-carried')).not.toBeInTheDocument();
  });

  // **持ち越しは「今回の賞が増える」こと。** 見えないと計算が合いません。
  it('notes a carry-over', async () => {
    mockExec.mockResolvedValue(makeState({ carriedPrizes: [card('DIAMOND', 4)], prizeValue: 13 }));
    renderWithProviders(<GoofspielPage />);
    expect(await screen.findByTestId('gs-carried')).toHaveTextContent('1');
    expect(screen.getByTestId('gs-prize')).toHaveTextContent('13');
  });

  it('bids the clicked card by its hand index', async () => {
    renderWithProviders(<GoofspielPage />);
    const cards = await screen.findAllByRole('button', { name: /で入札する$/ });
    mockExec.mockClear();
    fireEvent.click(cards[2]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', 2));
  });

  // **伏せたことは見せますが、中身は公開まで見せません。**
  it('shows that a seat has bid without showing the card', async () => {
    mockExec.mockResolvedValue(makeState({ players: [seat(0, { hasBid: true }), seat(1)] }));
    renderWithProviders(<GoofspielPage />);
    const s0 = await screen.findByTestId('gs-seat-0');
    expect(s0).toHaveTextContent(/入札済み/);
    expect(s0).not.toHaveTextContent(/出した札/);
    // 伏せたあとは押せない。
    expect(screen.getAllByRole('button', { name: /で入札する$/ })[0]).toBeDisabled();
  });

  it('shows the revealed bids and who took the points', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 1,
        currentPrize: undefined,
        lastWinnerIdx: 1,
        lastGained: 9,
        players: [seat(0, { revealedBid: card('SPADE', 3) }), seat(1, { revealedBid: card('CLOVER', 11), score: 9 })],
      }),
    );
    renderWithProviders(<GoofspielPage />);
    expect(await screen.findByTestId('gs-seat-1')).toHaveTextContent(/出した札/);
    expect(screen.getByTestId('gs-round-end')).toHaveTextContent(/CPU1 が 9 点/);
  });

  // **同点は誰も取りません。** 勝者が居ない結果を言い分けます。
  it('reports a tie as nobody taking the prize', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, currentPrize: undefined, lastWinnerIdx: -1, lastGained: 0 }));
    renderWithProviders(<GoofspielPage />);
    expect(await screen.findByTestId('gs-round-end')).toHaveTextContent(/誰も取りません/);
  });

  it('turns the next prize on request', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, currentPrize: undefined, lastWinnerIdx: 0, lastGained: 9 }));
    renderWithProviders(<GoofspielPage />);
    await screen.findByTestId('gs-round-end');

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('gs-next-btn'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  // **残り札は CPU の分も公開。** 使った札は場に出るので隠せていません。
  it('shows every seat remaining cards and score', async () => {
    mockExec.mockResolvedValue(makeState({ players: [seat(0), seat(1, { score: 21 })] }));
    renderWithProviders(<GoofspielPage />);
    const s1 = await screen.findByTestId('gs-seat-1');
    expect(s1).toHaveTextContent('12');
    expect(s1).toHaveTextContent('21');
    // **勝負はランクの大小比較** (#5769)。CPU の残り札も、alt の羅列ではなく
    // 自分の手札と同じカードの絵で出る。
    const hand = within(s1).getByTestId('gs-hand-1');
    const imgs = within(hand).getAllByRole('img');
    expect(imgs).toHaveLength(2);
    expect(imgs[0]).toHaveAccessibleName(expect.stringMatching(/♣|CLOVER/));
    // 13 枚でも折り返せること。
    expect(hand.className).toContain('flex-wrap');
  });

  it('reports who took the most', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 2,
        gameEndFlag: true,
        winnerIdx: 0,
        players: [seat(0, { score: 50 }), seat(1, { score: 41 })],
      }),
    );
    const { unmount } = renderWithProviders(<GoofspielPage />);
    const banner = await screen.findByTestId('gs-result');
    expect(banner).toHaveTextContent(/あなたの勝ち/);
    expect(banner).toHaveTextContent('50');
    unmount();

    mockExec.mockResolvedValue(
      makeState({
        phase: 2,
        gameEndFlag: true,
        winnerIdx: 1,
        players: [seat(0, { score: 30 }), seat(1, { score: 61 })],
      }),
    );
    renderWithProviders(<GoofspielPage />);
    expect(await screen.findByTestId('gs-result')).toHaveTextContent(/CPU1 が 61 点/);
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<GoofspielPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '投了' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('resets with the chosen table size and tie rule', async () => {
    renderWithProviders(<GoofspielPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    // **サーバは 2..3 しか受けない。** 弾かれる値を並べると黙って既定に戻される。
    const options = [...screen.getByTestId('gs-players-select').querySelectorAll('option')].map((o) => o.value);
    expect(options).toEqual(['2', '3']);

    fireEvent.change(screen.getByTestId('gs-players-select'), { target: { value: '3' } });
    fireEvent.change(screen.getByTestId('gs-tie-select'), { target: { value: '1' } });
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { playerCnt: 3, tieRule: 1 }));
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'card-2', reason: 'hint.goofspielHighPrize', confidence: 'moderate' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<GoofspielPage />);
    expect(await screen.findByTestId('hint-tooltip')).toHaveTextContent(/高い賞札です/);
  });
});
