import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { sjavsApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, SjavsPlayer, SjavsResponse } from '../types/card';
import { SjavsPhase } from '../types/phases';
import { SjavsPage } from './SjavsPage';

vi.mock('../api/gameApi', () => ({
  sjavsApi: { exec: vi.fn() },
  actionLogApi: { sjavs: vi.fn() },
}));

const mockExec = vi.mocked(sjavsApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function seat(id: number, isHuman: boolean, overrides?: Partial<SjavsPlayer>): SjavsPlayer {
  return {
    id,
    isHuman,
    team: id % 2,
    cardCount: 8,
    cards: isHuman ? [card('CLOVER', 12), card('HEART', 1), card('SPADE', 7)] : [],
    bid: 0,
    hidden: !isHuman,
    ...overrides,
  };
}

function makeState(overrides?: Partial<SjavsResponse>): SjavsResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false)],
    phase: SjavsPhase.BID,
    currentPlayerIdx: 0,
    dealerIdx: 3,
    trumpSuit: -1,
    trumpCount: 0,
    bidderIdx: -1,
    bidLength: 0,
    minBid: 5,
    myLongest: 6,
    trick: [],
    trickNo: 0,
    validIndices: [],
    teamPoints: [0, 0],
    remaining: [24, 24],
    crosses: [0, 0],
    carryOver: 0,
    gameEndFlag: false,
    winnerTeam: -1,
    doubleVictory: false,
    message: '',
    ...overrides,
  };
}

describe('SjavsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<SjavsPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows the permanent trumps permanently', async () => {
    // 切札スートの札しか切札でないと思い込むのが、このゲーム最大の誤解。
    renderWithProviders(<SjavsPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.getByText(/♣Q ＞ ♠Q ＞ ♣J ＞ ♠J ＞ ♥J ＞ ♦J/)).toBeInTheDocument();
  });

  it('shows the trump as undecided while bidding and the count once it is fixed', async () => {
    renderWithProviders(<SjavsPage />);
    await waitFor(() => expect(screen.getByText(/切札: 未定/)).toBeInTheDocument());

    mockExec.mockResolvedValue(makeState({ phase: SjavsPhase.PLAY, trumpSuit: 2, trumpCount: 13, validIndices: [0] }));
    renderWithProviders(<SjavsPage />);
    // 13 はサーバーが数える。クライアントが切札スートだけ数えると必ず足りない。
    await waitFor(() => expect(screen.getAllByText(/切札13枚/).length).toBeGreaterThan(0));
  });

  it('offers only the bid lengths the rules allow', async () => {
    // 5 枚未満は申告できず、自分の最長を超える申告もできない。
    renderWithProviders(<SjavsPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    expect(screen.getByRole('button', { name: '5枚を申告' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '6枚を申告' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '4枚を申告' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '7枚を申告' })).not.toBeInTheDocument();
  });

  it('sends a pass as a bid of zero', async () => {
    // 省略にすると必須項目エラーになる。0 は値として送らなければならない。
    renderWithProviders(<SjavsPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'パス' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', 0));
  });

  it('sends the chosen bid length', async () => {
    renderWithProviders(<SjavsPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '6枚を申告' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', 6));
  });

  it('only plays the hand cards the server marked valid', async () => {
    mockExec.mockResolvedValue(makeState({ phase: SjavsPhase.PLAY, trumpSuit: 2, trumpCount: 13, validIndices: [2] }));
    renderWithProviders(<SjavsPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    const handButtons = screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'play');
    mockExec.mockClear();

    fireEvent.click(handButtons[0]);
    // Without the flush this cannot fail: nothing has had a chance to dispatch.
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    fireEvent.click(handButtons[2]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 2));
  });

  it('cannot play during the bidding phase', async () => {
    renderWithProviders(<SjavsPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    const handButtons = screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'play');
    mockExec.mockClear();
    fireEvent.click(handButtons[0]);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('reports a 60-60 hand as scoring nothing', async () => {
    // 何も出ないと、入力を取りこぼしたように見える。
    mockExec.mockResolvedValue(
      makeState({
        phase: SjavsPhase.HAND_END,
        handResult: {
          declarerTeam: 0,
          declarerPoints: 60,
          scoringTeam: -1,
          amount: 0,
          vol: false,
          trumpWasClubs: false,
        },
      }),
    );
    renderWithProviders(<SjavsPage />);
    await waitFor(() => expect(screen.getByTestId('sjavs-hand-result')).toHaveTextContent(/60-60/));
    expect(screen.getByRole('button', { name: '次のハンドへ' })).toBeInTheDocument();
  });

  it('moves to the next hand on request', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: SjavsPhase.HAND_END,
        handResult: {
          declarerTeam: 0,
          declarerPoints: 95,
          scoringTeam: 0,
          amount: 4,
          vol: false,
          trumpWasClubs: false,
        },
      }),
    );
    renderWithProviders(<SjavsPage />);
    await waitFor(() => expect(screen.getByTestId('sjavs-hand-result')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のハンドへ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('reports each rubber outcome', async () => {
    for (const [winner, code, text] of [
      [0, 'sjavs.win', 'ラバーを取りました'],
      [1, 'sjavs.lose', 'ラバーを取られました'],
    ] as const) {
      mockExec.mockResolvedValue(
        makeState({ phase: SjavsPhase.GAME_END, gameEndFlag: true, winnerTeam: winner, messageCode: code }),
      );
      renderWithProviders(<SjavsPage />);
      await waitFor(() => expect(screen.getAllByText(text).length).toBeGreaterThan(0));
    }
  });
});
