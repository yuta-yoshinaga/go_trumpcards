import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { pigApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, PigResponse } from '../types/card';
import { PigPage } from './PigPage';

vi.mock('../api/gameApi', () => ({
  pigApi: { exec: vi.fn() },
  actionLogApi: { pig: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(pigApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const hand = [card('SPADE', 13), card('HEART', 13), card('CLOVER', 1), card('DIAMOND', 12)];

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 4,
  cards: id === 0 ? hand : [],
  letters: 0,
  letterWord: '',
  eliminated: false,
  hasSignalled: false,
  noticedOrder: 0,
  hasChosenPass: false,
  ...over,
});

function makeState(overrides: Partial<PigResponse> = {}): PigResponse {
  return {
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    validPlays: [0, 1, 2, 3],
    signallerIdx: -1,
    noticedCnt: 0,
    roundLoserIdx: -1,
    roundNumber: 2,
    passCount: 3,
    deckSize: 16,
    currentPlayerIdx: 0,
    gameEndFlag: false,
    winnerIdx: -1,
    config: { playerCnt: 4, cpuDifficulty: 1 },
    message: '',
    ...overrides,
  } as unknown as PigResponse;
}

/** A signal is out and the human has not answered it yet. */
const liveSignal = (over: Partial<PigResponse> = {}) =>
  makeState({ phase: 1, signallerIdx: 2, noticedCnt: 1, ...over } as Partial<PigResponse>);

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
});

describe('PigPage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<PigPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **取り合うものが何もないのが規則そのもの。**
  it('states that the signal is silent and being late is the loss', async () => {
    renderWithProviders(<PigPage />);
    expect(await screen.findByTestId('pig-rule')).toHaveTextContent(/黙って手を鼻に当てます/);
  });

  it('shows the round, deck size and pass count', async () => {
    renderWithProviders(<PigPage />);
    const head = await screen.findByTestId('pig-round');
    expect(head).toHaveTextContent('2');
    expect(head).toHaveTextContent('16');
  });

  // **文字がそのまま残機。** 得点表示はありません。
  it('shows every hand size and letters', async () => {
    mockExec.mockResolvedValue(
      makeState({ players: [seat(0, { letters: 2, letterWord: 'PI' }), seat(1), seat(2), seat(3)] }),
    );
    renderWithProviders(<PigPage />);
    const s0 = await screen.findByTestId('pig-seat-0');
    expect(s0).toHaveTextContent('手札4枚');
    expect(s0).toHaveTextContent('文字[PI]');
    expect(screen.getByTestId('pig-seat-3')).toHaveTextContent('文字[-]');
  });

  // **選び終えた席・気づいた席・脱落した席は盤面に痕跡が残らない。**
  it('marks who has chosen, who noticed and who is out', async () => {
    const { unmount } = renderWithProviders(<PigPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByTestId('pig-seat-1')).not.toHaveTextContent(/選択済み/);
    unmount();

    mockExec.mockResolvedValue(makeState({ players: [seat(0), seat(1, { hasChosenPass: true }), seat(2), seat(3)] }));
    const second = renderWithProviders(<PigPage />);
    expect(await screen.findByTestId('pig-seat-1')).toHaveTextContent(/選択済み/);
    second.unmount();

    mockExec.mockResolvedValue(
      liveSignal({ players: [seat(0), seat(1), seat(2, { hasSignalled: true, noticedOrder: 1 }), seat(3)] }),
    );
    const third = renderWithProviders(<PigPage />);
    expect(await screen.findByTestId('pig-seat-2')).toHaveTextContent(/1番目に気づいた/);
    third.unmount();

    mockExec.mockResolvedValue(
      makeState({ players: [seat(0), seat(1, { eliminated: true, letters: 3, letterWord: 'PIG' }), seat(2), seat(3)] }),
    );
    renderWithProviders(<PigPage />);
    expect(await screen.findByTestId('pig-seat-1')).toHaveTextContent(/脱落/);
  });

  it('passes the clicked card by its hand index', async () => {
    renderWithProviders(<PigPage />);
    const cards = await screen.findAllByRole('button', { name: /へ渡す$/ });
    mockExec.mockClear();
    fireEvent.click(cards[1]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass', 1));
  });

  // **同時に渡すので、選んだあとは待ちになる。**
  it('locks the hand once you have chosen', async () => {
    mockExec.mockResolvedValue(makeState({ players: [seat(0, { hasChosenPass: true }), seat(1), seat(2), seat(3)] }));
    renderWithProviders(<PigPage />);
    expect(await screen.findByTestId('pig-waiting')).toHaveTextContent(/全員が選ぶまで待ちます/);
    expect(screen.getAllByRole('button', { name: /へ渡す$/ })[0]).toBeDisabled();
  });

  // **合図が出ている場面は、押すべきボタンが1つだけ。**
  it('offers the signal button while a signal is live', async () => {
    mockExec.mockResolvedValue(liveSignal());
    renderWithProviders(<PigPage />);

    expect(await screen.findByTestId('pig-signal-alert')).toHaveTextContent(/鼻に当てました/);
    const btn = screen.getByTestId('pig-signal-btn');
    expect(btn).toBeEnabled();
    // 合図の場面では札を渡せない。
    expect(screen.getAllByRole('button', { name: /へ渡す$/ })[0]).toBeDisabled();

    mockExec.mockClear();
    fireEvent.click(btn);
    // **合図は別のコマンド。** cardIndex は送らない。
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('signal'));
  });

  // **負のコントロール: 合図が出ていなければボタンは出ない。**
  it('hides the signal button while cards are still being passed', async () => {
    renderWithProviders(<PigPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('pig-signal-btn')).not.toBeInTheDocument();
    expect(screen.queryByTestId('pig-signal-alert')).not.toBeInTheDocument();
  });

  it('reports that you already signalled and hides the button', async () => {
    mockExec.mockResolvedValue(
      liveSignal({
        noticedCnt: 2,
        players: [seat(0, { hasSignalled: true, noticedOrder: 2 }), seat(1), seat(2), seat(3)],
      }),
    );
    renderWithProviders(<PigPage />);
    expect(await screen.findByTestId('pig-signal-done')).toHaveTextContent('2');
    expect(screen.queryByTestId('pig-signal-btn')).not.toBeInTheDocument();
  });

  // **罰は1ラウンドに1回の出来事。** 配り直す前に読ませる。
  it('shows the round result and deals the next round on request', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 2,
        roundLoserIdx: 1,
        players: [seat(0), seat(1, { letters: 1, letterWord: 'P' }), seat(2), seat(3)],
      }),
    );
    renderWithProviders(<PigPage />);
    expect(await screen.findByTestId('pig-round-end')).toHaveTextContent(/CPU1 が気づくのに遅れ/);

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('pig-next-btn'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  // **脱落しても局は続く。**
  it('tells the human when they are out but the game continues', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [seat(0, { eliminated: true, letters: 3, letterWord: 'PIG', cards: [] }), seat(1), seat(2), seat(3)],
      }),
    );
    renderWithProviders(<PigPage />);
    expect(await screen.findByTestId('pig-eliminated')).toHaveTextContent(/見守ります/);
  });

  it('reports who was left standing', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 3, gameEndFlag: true, winnerIdx: 0 }));
    const { unmount } = renderWithProviders(<PigPage />);
    expect(await screen.findByTestId('pig-result')).toHaveTextContent(/最後まで残りました/);
    unmount();

    mockExec.mockResolvedValue(makeState({ phase: 3, gameEndFlag: true, winnerIdx: 2 }));
    renderWithProviders(<PigPage />);
    expect(await screen.findByTestId('pig-result')).toHaveTextContent(/CPU2 が最後まで残りました/);
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<PigPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '投了' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('resets with the chosen table size and difficulty', async () => {
    renderWithProviders(<PigPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    // **サーバは 3..6 しか受けない。** 弾かれる値を並べると黙って既定に戻される。
    const options = [...screen.getByTestId('pig-players-select').querySelectorAll('option')].map((o) => o.value);
    expect(options).toEqual(['3', '4', '5', '6']);

    fireEvent.change(screen.getByTestId('pig-players-select'), { target: { value: '6' } });
    fireEvent.change(screen.getByTestId('pig-difficulty-select'), { target: { value: '2' } });
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { playerCnt: 6, cpuDifficulty: 2 }));
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'signal', reason: 'hint.pigSignal', confidence: 'strong' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<PigPage />);
    expect(await screen.findByTestId('hint-tooltip')).toHaveTextContent(/いますぐ/);
  });
});
