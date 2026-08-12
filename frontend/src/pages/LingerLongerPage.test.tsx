import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { lingerlongerApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, LingerLongerResponse } from '../types/card';
import { LingerLongerPage } from './LingerLongerPage';

vi.mock('../api/gameApi', () => ({
  lingerlongerApi: { exec: vi.fn() },
  actionLogApi: { lingerlonger: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(lingerlongerApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const hand = [card('SPADE', 9), card('SPADE', 13), card('HEART', 10), card('CLOVER', 7)];

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 4,
  cards: id === 0 ? hand : [],
  tricksWon: 0,
  eliminatedAt: 0,
  ...over,
});

function makeState(overrides: Partial<LingerLongerResponse> = {}): LingerLongerResponse {
  return {
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    validPlays: [0, 1],
    stockSize: 30,
    currentTrick: [],
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    trickNumber: 2,
    lastDrawIdx: -1,
    eliminatedCnt: 0,
    discarded: 6,
    gameEndFlag: false,
    winnerIdx: -1,
    config: { playerCnt: 4 },
    message: '',
    ...overrides,
  } as unknown as LingerLongerResponse;
}

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
});

describe('LingerLongerPage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<LingerLongerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **取っても得点にならず、補充できるだけ。** 直感と逆なので毎回出す。
  it('states that a trick only buys you a card', async () => {
    renderWithProviders(<LingerLongerPage />);
    expect(await screen.findByTestId('ll-rule')).toHaveTextContent(/1枚補充する権利/);
  });

  it('shows the trick number and the stock', async () => {
    renderWithProviders(<LingerLongerPage />);
    expect(await screen.findByTestId('ll-stock')).toHaveTextContent('30');
  });

  // **山札が尽きた瞬間から局は終わりに向かう。** 盤面からは読み取れない。
  it('announces the empty stock', async () => {
    const { unmount } = renderWithProviders(<LingerLongerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('ll-no-stock')).not.toBeInTheDocument();
    unmount();

    mockExec.mockResolvedValue(makeState({ stockSize: 0 }));
    renderWithProviders(<LingerLongerPage />);
    expect(await screen.findByTestId('ll-no-stock')).toHaveTextContent(/補充できません/);
  });

  // **負のコントロール: 終局後は残っていても言わない。**
  it('does not announce the empty stock after the game ends', async () => {
    mockExec.mockResolvedValue(makeState({ stockSize: 0, gameEndFlag: true, phase: 1, winnerIdx: 0 }));
    renderWithProviders(<LingerLongerPage />);
    await screen.findByTestId('ll-result');
    expect(screen.queryByTestId('ll-no-stock')).not.toBeInTheDocument();
  });

  // **手札の枚数が生死そのもの。** 得点表示は無い。
  it('shows every hand size and trick count', async () => {
    mockExec.mockResolvedValue(
      makeState({ players: [seat(0, { cardCount: 7, tricksWon: 3 }), seat(1), seat(2), seat(3)] }),
    );
    renderWithProviders(<LingerLongerPage />);
    const s0 = await screen.findByTestId('ll-seat-0');
    expect(s0).toHaveTextContent('手札7枚');
    expect(s0).toHaveTextContent('獲得3回');
    expect(screen.getByTestId('ll-seat-3')).toBeInTheDocument();
  });

  // **補充した席と脱落した席は盤面に痕跡が残らない。**
  it('marks the last draw and the eliminated seats', async () => {
    const { unmount } = renderWithProviders(<LingerLongerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByTestId('ll-seat-2')).not.toHaveTextContent(/補充\]/);
    unmount();

    mockExec.mockResolvedValue(makeState({ lastDrawIdx: 2 }));
    const second = renderWithProviders(<LingerLongerPage />);
    expect(await screen.findByTestId('ll-seat-2')).toHaveTextContent(/直前に補充/);
    second.unmount();

    mockExec.mockResolvedValue(
      makeState({ players: [seat(0), seat(1, { cardCount: 0, eliminatedAt: 1 }), seat(2), seat(3)] }),
    );
    renderWithProviders(<LingerLongerPage />);
    expect(await screen.findByTestId('ll-seat-1')).toHaveTextContent(/1番目に脱落/);
  });

  it('plays the clicked card by its hand index', async () => {
    renderWithProviders(<LingerLongerPage />);
    const cards = await screen.findAllByRole('button', { name: /を出す$/ });
    mockExec.mockClear();
    fireEvent.click(cards[1]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 1));
  });

  it('disables the hand while it is a CPU turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1 }));
    renderWithProviders(<LingerLongerPage />);
    const cards = await screen.findAllByRole('button', { name: /を出す$/ });
    expect(cards[0]).toBeDisabled();
  });

  // **脱落しても局は続く。** 手番が二度と来ないので、黙っていると固まって見える。
  it('tells the human when they are out but the game continues', async () => {
    mockExec.mockResolvedValue(
      makeState({
        currentPlayerIdx: 1,
        players: [seat(0, { cardCount: 0, cards: [], eliminatedAt: 2 }), seat(1), seat(2), seat(3)],
      }),
    );
    renderWithProviders(<LingerLongerPage />);
    expect(await screen.findByTestId('ll-eliminated')).toHaveTextContent(/見守ります/);
  });

  // **負のコントロール: まだ在席なら出さない。**
  it('does not claim you are out while you still hold cards', async () => {
    renderWithProviders(<LingerLongerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('ll-eliminated')).not.toBeInTheDocument();
  });

  it('reports who held cards longest', async () => {
    const { unmount } = renderWithProviders(<LingerLongerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    unmount();

    mockExec.mockResolvedValue(makeState({ gameEndFlag: true, phase: 1, winnerIdx: 0 }));
    const second = renderWithProviders(<LingerLongerPage />);
    expect(await screen.findByTestId('ll-result')).toHaveTextContent(/最後まで手札を持ち続けました/);
    second.unmount();

    mockExec.mockResolvedValue(makeState({ gameEndFlag: true, phase: 1, winnerIdx: 2 }));
    renderWithProviders(<LingerLongerPage />);
    expect(await screen.findByTestId('ll-result')).toHaveTextContent(/CPU2 が最後まで持ち続けました/);
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<LingerLongerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '投了' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  // **3 人でも遊べる。** 配る枚数が人数と同じなので、卓の大きさが手札を決める。
  it('resets with the chosen table size', async () => {
    renderWithProviders(<LingerLongerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    fireEvent.change(screen.getByTestId('ll-players-select'), { target: { value: '6' } });
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    // リセットは確認ダイアログ越し。押しただけでは配り直さない。
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { playerCnt: 6 }));
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'card-0', reason: 'hint.lingerlongerWinTrick', confidence: 'strong' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<LingerLongerPage />);
    expect(await screen.findByTestId('hint-tooltip')).toHaveTextContent(/1枚補充できます/);
  });
});
