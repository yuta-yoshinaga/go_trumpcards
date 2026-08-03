import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { popejoanApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { CardDesign, PopeJoanPlayer, PopeJoanResponse } from '../types/card';
import { PopeJoanPhase } from '../types/phases';
import { PopeJoanPage } from './PopeJoanPage';

vi.mock('../api/gameApi', () => ({
  popejoanApi: { exec: vi.fn() },
  actionLogApi: { popejoan: vi.fn() },
}));

const mockExec = vi.mocked(popejoanApi.exec);

const card = (design: CardDesign, value: number) => ({ design, value });

const COMPS = ['ace', 'king', 'queen', 'jack', 'game', 'pope', 'matrimony', 'intrigue'];
const DRESS = [1, 1, 1, 1, 1, 6, 2, 2];

function seat(id: number, isHuman: boolean, overrides?: Partial<PopeJoanPlayer>): PopeJoanPlayer {
  return {
    id,
    isHuman,
    cardCount: 3,
    cards: isHuman ? [card('SPADE', 3), card('HEART', 9), card('CLOVER', 13)] : [],
    chips: -15,
    holdsPope: false,
    hidden: !isHuman,
    ...overrides,
  };
}

function makeState(overrides?: Partial<PopeJoanResponse>): PopeJoanResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false)],
    phase: PopeJoanPhase.PLAY,
    currentPlayerIdx: 0,
    compartments: COMPS.map((name, i) => ({ name, chips: DRESS[i] })),
    trumpSuit: 1,
    turnUp: card('SPADE', 5),
    awards: [],
    playedPile: [],
    runSuit: -1,
    runRank: 0,
    dealNo: 0,
    targetDeals: 5,
    dealWinner: -1,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    ...overrides,
  };
}

/** Click the hand card at the given index. */
function pickHand(i: number) {
  const hand = screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'play');
  fireEvent.click(hand[i]);
}

describe('PopeJoanPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<PopeJoanPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows both rules permanently', async () => {
    renderWithProviders(<PopeJoanPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.getByText(/トランプの札でしか取れません/)).toBeInTheDocument();
    expect(screen.getByText(/♦8 が抜いてあるので/)).toBeInTheDocument();
  });

  // **8 区画すべてが出ていないと、持ち越しがどこに乗っているか読めない。**
  it("shows all eight compartments with the dealer's fixed dress", async () => {
    renderWithProviders(<PopeJoanPage />);
    await waitFor(() => expect(screen.getAllByTestId('popejoan-compartment')).toHaveLength(8));
    const pope = screen.getAllByTestId('popejoan-compartment').find((el) => el.textContent?.includes('ポープ'));
    expect(pope).toHaveTextContent('6');
  });

  // めくり札での即取りは通常の獲得と区別して読めなければならない。
  it('marks a turn-up award apart from an ordinary one', async () => {
    mockExec.mockResolvedValue(makeState({ awards: [{ compartment: 'pope', player: 0, chips: 6, byTurnUp: true }] }));
    renderWithProviders(<PopeJoanPage />);
    await waitFor(() => expect(screen.getByTestId('popejoan-awards')).toHaveTextContent('めくり札'));
  });

  it('plays exactly one card', async () => {
    renderWithProviders(<PopeJoanPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    pickHand(1);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 1));
  });

  // 止まっているかどうかで出せる札がまるで違うので、案内も変わる。
  it('tells a stopped run apart from one in progress', async () => {
    renderWithProviders(<PopeJoanPage />);
    await waitFor(() => expect(screen.getByText(/最も低い札を選んでください/)).toBeInTheDocument());
  });

  it('asks for the next higher card while a run is live', async () => {
    mockExec.mockResolvedValue(makeState({ runSuit: 1, runRank: 5, playedPile: [card('SPADE', 5)] }));
    renderWithProviders(<PopeJoanPage />);
    await waitFor(() => expect(screen.getByText(/同じスートの次に高い札/)).toBeInTheDocument());
  });

  // **Pope 保持者は支払いを免除される。**伏せ手でも見えていないと精算が読めない。
  it('marks a hidden opponent as holding the Pope', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [seat(0, true), seat(1, false, { holdsPope: true }), seat(2, false), seat(3, false)],
      }),
    );
    renderWithProviders(<PopeJoanPage />);
    await waitFor(() => expect(screen.getAllByText(/ポープ保持/).length).toBeGreaterThan(0));
  });

  it('advances to the next deal', async () => {
    mockExec.mockResolvedValue(makeState({ phase: PopeJoanPhase.DEAL_END, dealWinner: 2 }));
    renderWithProviders(<PopeJoanPage />);
    await waitFor(() => expect(screen.getByTestId('popejoan-deal-result')).toHaveTextContent('席2'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のディールへ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('reports each outcome', async () => {
    for (const [winner, code, text] of [
      [0, 'popejoan.win', '最も多く稼ぎました'],
      [2, 'popejoan.lose', '及びませんでした'],
    ] as const) {
      mockExec.mockResolvedValue(
        makeState({ phase: PopeJoanPhase.GAME_END, gameEndFlag: true, winnerIdx: winner, messageCode: code }),
      );
      renderWithProviders(<PopeJoanPage />);
      await waitFor(() => expect(screen.getAllByText(text).length).toBeGreaterThan(0));
    }
  });
});
