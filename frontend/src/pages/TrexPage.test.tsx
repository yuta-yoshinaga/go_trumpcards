import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { trexApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, TrexPlayer, TrexResponse } from '../types/card';
import { TrexContract, TrexPhase } from '../types/phases';
import { TrexPage } from './TrexPage';

vi.mock('../api/gameApi', () => ({
  trexApi: { exec: vi.fn() },
  actionLogApi: { trex: vi.fn() },
}));

const mockExec = vi.mocked(trexApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function seat(id: number, isHuman: boolean, overrides?: Partial<TrexPlayer>): TrexPlayer {
  return {
    id,
    isHuman,
    cardCount: 13,
    cards: isHuman ? [card('SPADE', 11), card('HEART', 1), card('CLOVER', 5)] : [],
    score: 0,
    dealScore: 0,
    tricksWon: 0,
    hidden: !isHuman,
    ...overrides,
  };
}

function makeState(overrides?: Partial<TrexResponse>): TrexResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false)],
    phase: TrexPhase.CHOOSE,
    currentPlayerIdx: 0,
    kingIdx: 0,
    contract: TrexContract.NONE,
    availableContracts: [
      TrexContract.KING_OF_HEARTS,
      TrexContract.DIAMONDS,
      TrexContract.QUEENS,
      TrexContract.TRICKS,
      TrexContract.DOMINOES,
    ],
    isTrix: false,
    dealNo: 0,
    totalDeals: 20,
    trick: [],
    trickNo: 0,
    runs: [
      { suit: 1, started: false, low: 0, high: 0 },
      { suit: 2, started: false, low: 0, high: 0 },
      { suit: 3, started: false, low: 0, high: 0 },
      { suit: 4, started: false, low: 0, high: 0 },
    ],
    finishOrder: [],
    validIndices: [],
    canPass: false,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    ...overrides,
  };
}

describe('TrexPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<TrexPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows both rules permanently and the deal counter', async () => {
    renderWithProviders(<TrexPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.getByText(/同じ契約は1王国に1度だけ/)).toBeInTheDocument();
    expect(screen.getByText(/Jを起点/)).toBeInTheDocument();
    expect(screen.getByText(/ディール: 0\/20/)).toBeInTheDocument();
  });

  it('offers only the contracts this king has left', async () => {
    // 1王国に1度ずつなので、消化済みを出すと押しても弾かれる袋小路になる。
    mockExec.mockResolvedValue(makeState({ availableContracts: [TrexContract.QUEENS, TrexContract.DOMINOES] }));
    renderWithProviders(<TrexPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    expect(screen.getByRole('button', { name: /クイーン/ })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /ドミノ/ })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /ダイヤ/ })).not.toBeInTheDocument();
  });

  it('sends contract zero, the king of hearts, as a value', async () => {
    // 省略にすると ♥K 契約だけ選べなくなる。
    renderWithProviders(<TrexPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: /♥K/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('choose', TrexContract.KING_OF_HEARTS));
  });

  it('waits quietly while another seat is king', async () => {
    // 人間が王でないときに契約ボタンを出すと、押せない選択肢が並ぶ。
    mockExec.mockResolvedValue(makeState({ kingIdx: 2 }));
    renderWithProviders(<TrexPage />);
    await waitFor(() => expect(screen.getByText(/席2 が契約を選んでいます/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: /クイーン/ })).not.toBeInTheDocument();
  });

  it('only plays the hand cards the server marked valid', async () => {
    mockExec.mockResolvedValue(makeState({ phase: TrexPhase.PLAY, contract: TrexContract.QUEENS, validIndices: [2] }));
    renderWithProviders(<TrexPage />);
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

  it('shows the four runs during the dominoes and the trick otherwise', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: TrexPhase.PLAY,
        contract: TrexContract.DOMINOES,
        isTrix: true,
        runs: [
          { suit: 1, started: true, low: 11, high: 13 },
          { suit: 2, started: false, low: 0, high: 0 },
          { suit: 3, started: false, low: 0, high: 0 },
          { suit: 4, started: false, low: 0, high: 0 },
        ],
      }),
    );
    renderWithProviders(<TrexPage />);
    await waitFor(() => expect(screen.getAllByTestId('trex-run')).toHaveLength(4));
    expect(screen.getAllByTestId('trex-run')[0]).toHaveTextContent('11–13');
  });

  it('offers the pass only in the dominoes with no legal play', async () => {
    // トリック契約にパスは存在しない。
    mockExec.mockResolvedValue(makeState({ phase: TrexPhase.PLAY, contract: TrexContract.QUEENS }));
    renderWithProviders(<TrexPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: 'パス' })).not.toBeInTheDocument();

    mockExec.mockResolvedValue(
      makeState({ phase: TrexPhase.PLAY, contract: TrexContract.DOMINOES, isTrix: true, canPass: true }),
    );
    renderWithProviders(<TrexPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'パス' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
  });

  it('moves to the next deal on request', async () => {
    mockExec.mockResolvedValue(makeState({ phase: TrexPhase.DEAL_END, contract: TrexContract.QUEENS }));
    renderWithProviders(<TrexPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のディールへ' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のディールへ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('reports each outcome', async () => {
    for (const [winner, code, text] of [
      [0, 'trex.win', '勝ちました'],
      [2, 'trex.lose', '負けました'],
    ] as const) {
      mockExec.mockResolvedValue(
        makeState({ phase: TrexPhase.GAME_END, gameEndFlag: true, winnerIdx: winner, messageCode: code }),
      );
      renderWithProviders(<TrexPage />);
      await waitFor(() => expect(screen.getAllByText(text).length).toBeGreaterThan(0));
    }
  });
});
