import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { pasurApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, PasurResponse } from '../types/card';
import { PasurPage } from './PasurPage';

vi.mock('../api/gameApi', () => ({
  pasurApi: { exec: vi.fn() },
  actionLogApi: { pasur: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(pasurApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const hand = [card('DIAMOND', 4), card('SPADE', 13), card('HEART', 9), card('CLOVER', 6)];

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 4,
  cards: id === 0 ? hand : [],
  capturedCount: 0,
  soors: 0,
  score: 0,
  ...over,
});

function makeState(overrides: Partial<PasurResponse> = {}): PasurResponse {
  return {
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    table: [card('SPADE', 7), card('HEART', 4), card('CLOVER', 3)],
    // ♦4 は ♠7 単独、または ♥4+♣3。♠K は取れない。
    captureOptions: [[[0], [1, 2]], [], [], []],
    deckRemaining: 32,
    packsDealt: 1,
    lastCaptureIdx: -1,
    currentPlayerIdx: 0,
    gameEndFlag: false,
    winners: [],
    config: { playerCnt: 4 },
    message: '',
    ...overrides,
  } as unknown as PasurResponse;
}

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
});

describe('PasurPage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<PasurPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **11 の合計と絵札の扱いが規則そのもの。**
  it('states the capture rule', async () => {
    renderWithProviders(<PasurPage />);
    expect(await screen.findByTestId('ps-rule')).toHaveTextContent(/合計が11/);
  });

  it('shows the table, and says when it is empty', async () => {
    const { unmount } = renderWithProviders(<PasurPage />);
    expect(await screen.findByTestId('ps-table')).toHaveTextContent(/場/);
    unmount();

    mockExec.mockResolvedValue(makeState({ table: [] }));
    renderWithProviders(<PasurPage />);
    expect(await screen.findByTestId('ps-table')).toHaveTextContent(/なし/);
  });

  it('shows captures and soors for every seat', async () => {
    mockExec.mockResolvedValue(
      makeState({ players: [seat(0, { capturedCount: 6, soors: 2, score: 9 }), seat(1), seat(2), seat(3)] }),
    );
    renderWithProviders(<PasurPage />);
    const s0 = await screen.findByTestId('ps-seat-0');
    expect(s0).toHaveTextContent('捕獲6枚');
    expect(s0).toHaveTextContent('スール2');
    expect(s0).toHaveTextContent('得点9');
    expect(screen.getByTestId('ps-seat-3')).toBeInTheDocument();
  });

  // **場に残った札の行き先が読めること。**
  it('marks the last capturer only once someone has captured', async () => {
    const { unmount } = renderWithProviders(<PasurPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByTestId('ps-seat-2')).not.toHaveTextContent(/最後に捕獲/);
    unmount();

    mockExec.mockResolvedValue(makeState({ lastCaptureIdx: 2 }));
    renderWithProviders(<PasurPage />);
    expect(await screen.findByTestId('ps-seat-2')).toHaveTextContent(/最後に捕獲/);
  });

  // **候補はサーバが送ったものだけ。** ここで 11 の部分集合を作り直さない。
  it('offers exactly the capture options the server sent', async () => {
    renderWithProviders(<PasurPage />);
    const cards = await screen.findAllByRole('button', { name: /を選ぶ$/ });
    fireEvent.click(cards[0]);

    expect(await screen.findByTestId('ps-options')).toBeInTheDocument();
    expect(screen.getByTestId('ps-take-0-btn')).toBeInTheDocument();
    expect(screen.getByTestId('ps-take-1-2-btn')).toBeInTheDocument();
    // 送られていない組み合わせは出さない。
    expect(screen.queryByTestId('ps-take-1-btn')).not.toBeInTheDocument();
    expect(screen.queryByTestId('ps-take-0-1-2-btn')).not.toBeInTheDocument();
  });

  it('sends the chosen card and table indices', async () => {
    renderWithProviders(<PasurPage />);
    const cards = await screen.findAllByRole('button', { name: /を選ぶ$/ });
    fireEvent.click(cards[0]);

    mockExec.mockClear();
    fireEvent.click(await screen.findByTestId('ps-take-1-2-btn'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0, undefined, [1, 2]));
  });

  // **取れる組み合わせがあるときは場に置けない。** サーバが必ず拒否する。
  it('hides the lay-down button while a capture is available', async () => {
    renderWithProviders(<PasurPage />);
    const cards = await screen.findAllByRole('button', { name: /を選ぶ$/ });

    fireEvent.click(cards[0]);
    expect(await screen.findByTestId('ps-must-capture')).toBeInTheDocument();
    expect(screen.queryByTestId('ps-trail-btn')).not.toBeInTheDocument();

    // ♠K は取れないので置ける。
    fireEvent.click(cards[0]);
    fireEvent.click(cards[1]);
    expect(await screen.findByTestId('ps-trail-btn')).toBeEnabled();
    expect(screen.queryByTestId('ps-must-capture')).not.toBeInTheDocument();
  });

  it('sends an empty selection when laying a card down', async () => {
    renderWithProviders(<PasurPage />);
    const cards = await screen.findAllByRole('button', { name: /を選ぶ$/ });
    fireEvent.click(cards[1]);

    mockExec.mockClear();
    fireEvent.click(await screen.findByTestId('ps-trail-btn'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 1, undefined, []));
  });

  it('lets you deselect and choose again', async () => {
    renderWithProviders(<PasurPage />);
    const cards = await screen.findAllByRole('button', { name: /を選ぶ$/ });

    fireEvent.click(cards[0]);
    expect(cards[0]).toHaveAttribute('aria-pressed', 'true');
    fireEvent.click(cards[0]);
    expect(cards[0]).toHaveAttribute('aria-pressed', 'false');
    expect(screen.queryByTestId('ps-options')).not.toBeInTheDocument();

    fireEvent.click(cards[0]);
    fireEvent.click(screen.getByRole('button', { name: '選び直す' }));
    expect(screen.queryByTestId('ps-options')).not.toBeInTheDocument();
  });

  it('disables the hand while it is a CPU turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1 }));
    renderWithProviders(<PasurPage />);
    const cards = await screen.findAllByRole('button', { name: /を選ぶ$/ });
    expect(cards[0]).toBeDisabled();
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<PasurPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '投了' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('renders the result banner for each outcome', async () => {
    for (const [winners, expected] of [
      [[0], /あなたの勝ち/],
      [[2], /CPU2 の勝ち/],
      [[1, 3], /2 人が同点/],
    ] as const) {
      mockExec.mockResolvedValue(makeState({ gameEndFlag: true, phase: 1, winners: [...winners] }));
      const { unmount } = renderWithProviders(<PasurPage />);
      expect(await screen.findByTestId('ps-result')).toHaveTextContent(expected);
      unmount();
    }
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'card-0-take-0', reason: 'hint.pasurSoor', confidence: 'strong' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<PasurPage />);
    expect(await screen.findByTestId('hint-tooltip')).toHaveTextContent(/場を空に/);
  });
});

// **スールは「取った結果、場が空になる」こと** (#5762)。倍化を狙うなら、どの
// 選択肢がそれに当たるかがボタンから読めないと選べない。
describe('PasurPage soor options', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
  });

  it('badges the option that clears the table and not the partial one', async () => {
    mockExec.mockResolvedValue(
      makeState({
        table: [card('SPADE', 3), card('HEART', 4), card('CLOVER', 7)],
        // 手札 0 で「場を全部」か「1 枚だけ」かを選べる状況。
        captureOptions: [[[0, 1, 2], [0]], [], [], []],
      } as Partial<PasurResponse>),
    );
    renderWithProviders(<PasurPage />);

    // 手札の 1 枚目を選ぶと、その札の取得候補がボタンになる。
    const handButtons = await screen.findAllByRole('button', { name: /を選ぶ/ });
    fireEvent.click(handButtons[0]);

    expect(await screen.findByTestId('ps-soor-0-1-2')).toHaveTextContent('スール');
    expect(screen.queryByTestId('ps-soor-0')).not.toBeInTheDocument();
    // 読み上げには倍化まで入る。
    expect(screen.getByTestId('ps-soor-0-1-2')).toHaveTextContent('2 倍');
  });

  it('does not badge a full-length option when the table is empty', async () => {
    mockExec.mockResolvedValue(makeState({ table: [], captureOptions: [[], [], [], []] } as Partial<PasurResponse>));
    renderWithProviders(<PasurPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('ps-options')).not.toBeInTheDocument();
  });
});
