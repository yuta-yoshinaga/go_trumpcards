import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { sixBidSoloApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { CardDesign, SixBidSoloHandResult, SixBidSoloPlayer, SixBidSoloResponse } from '../types/card';
import { SixBidSoloPhase } from '../types/phases';
import { SixBidSoloPage } from './SixBidSoloPage';

vi.mock('../api/gameApi', () => ({
  sixBidSoloApi: { exec: vi.fn() },
  actionLogApi: { sixbidsolo: vi.fn() },
}));

const mockExec = vi.mocked(sixBidSoloApi.exec);

const card = (design: CardDesign, value: number) => ({ design, value });

function seat(id: number, isHuman: boolean, overrides?: Partial<SixBidSoloPlayer>): SixBidSoloPlayer {
  return {
    id,
    isHuman,
    cardCount: 3,
    cards: isHuman ? [card('SPADE', 1), card('HEART', 10), card('CLOVER', 13)] : [],
    points: 0,
    tricksWon: 0,
    score: 0,
    isDealer: id === 2,
    isDeclarer: false,
    isCurrentTurn: id === 0,
    ...overrides,
  };
}

function result(overrides?: Partial<SixBidSoloHandResult>): SixBidSoloHandResult {
  return {
    kind: 1,
    declarer: 0,
    declarerPoints: 65,
    widowPoints: 25,
    target: 61,
    made: true,
    value: 10,
    deltas: [20, -10, -10],
    ...overrides,
  };
}

function makeState(overrides?: Partial<SixBidSoloResponse>): SixBidSoloResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false)],
    phase: SixBidSoloPhase.PLAY,
    handNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 2,
    bids: [],
    highBid: { player: 0, kind: 1 },
    declarerIdx: 0,
    trumpSuit: 1,
    declared: true,
    calledCard: null,
    spreadOpen: false,
    widow: [],
    widowSize: 3,
    trick: [],
    validPlays: [0, 1, 2],
    trickLeaderIdx: 0,
    trickNumber: 0,
    lastResult: null,
    bidTargets: [0, 61, 61, 0, 80, 0, 120],
    totalPoints: 120,
    baseTarget: 60,
    handSize: 11,
    targetHands: 6,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    config: { cpuDifficulty: 0, targetHands: 6 },
    ...overrides,
  };
}

/** The hand-card buttons, in order. */
function handButtons() {
  return screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'play');
}

describe('SixBidSoloPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<SixBidSoloPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **11枚ずつ＋ウィドウ3枚。**issue の「12枚ずつ・スキャットなし」は誤り。
  it('shows the widow face down and explains that it is credited to the declarer', async () => {
    renderWithProviders(<SixBidSoloPage />);
    await waitFor(() => expect(screen.getByTestId('sixbidsolo-widow')).toBeInTheDocument());
    const widow = screen.getByTestId('sixbidsolo-widow');
    expect(widow).toHaveTextContent('伏せ 3枚');
    expect(widow).toHaveTextContent('11枚ずつ配って3枚がウィドウ');
    expect(widow).toHaveTextContent('宣言者の得点へ加算');
  });

  it('plays exactly one card', async () => {
    renderWithProviders(<SixBidSoloPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    fireEvent.click(handButtons()[1]);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 1 }));
  });

  // **追随は強制。**サーバーが出せる札を決める。
  it('disables cards the server did not list as playable', async () => {
    mockExec.mockResolvedValue(makeState({ validPlays: [2] }));
    renderWithProviders(<SixBidSoloPage />);
    await waitFor(() => expect(screen.getByTestId('sixbidsolo-play-notice')).toBeInTheDocument());

    const hand = handButtons();
    expect(hand[0]).toBeDisabled();
    expect(hand[1]).toBeDisabled();
    expect(hand[2]).toBeEnabled();
  });

  // **通常ビッドは61点以上、ミゼールは0点。**入札画面で両方読めること。
  it('states the real targets while bidding', async () => {
    mockExec.mockResolvedValue(makeState({ phase: SixBidSoloPhase.BID, highBid: null, declarerIdx: -1 }));
    renderWithProviders(<SixBidSoloPage />);
    await waitFor(() => expect(screen.getByTestId('sixbidsolo-base-note')).toBeInTheDocument());

    expect(screen.getByTestId('sixbidsolo-base-note')).toHaveTextContent('60点ちょうどでは足りません');
    expect(screen.getByTestId('sixbidsolo-base-note')).toHaveTextContent('61点以上');
    expect(screen.getByTestId('sixbidsolo-misere-note')).toHaveTextContent('「0トリック」ではなく「0点」');
  });

  // **上回る宣言だけが通る。**通らない値は出さない。
  it('offers the six bids, then only the ones that can beat the standing bid', async () => {
    mockExec.mockResolvedValue(makeState({ phase: SixBidSoloPhase.BID, highBid: null, declarerIdx: -1 }));
    renderWithProviders(<SixBidSoloPage />);
    await waitFor(() => expect(screen.getByLabelText(/宣言/)).toBeInTheDocument());
    expect(Array.from((screen.getByLabelText(/宣言/) as HTMLSelectElement).options).map((o) => o.value)).toEqual([
      '1',
      '2',
      '3',
      '4',
      '5',
      '6',
    ]);

    mockExec.mockResolvedValue(
      makeState({ phase: SixBidSoloPhase.BID, highBid: { player: 1, kind: 4 }, declarerIdx: -1 }),
    );
    renderWithProviders(<SixBidSoloPage />);
    await waitFor(() =>
      expect(
        Array.from((screen.getAllByLabelText(/宣言/)[1] as HTMLSelectElement).options).map((o) => o.value),
      ).toEqual(['5', '6']),
    );
  });

  it('bids and passes', async () => {
    mockExec.mockResolvedValue(makeState({ phase: SixBidSoloPhase.BID, highBid: null, declarerIdx: -1 }));
    renderWithProviders(<SixBidSoloPage />);
    await waitFor(() => expect(screen.getByLabelText(/宣言/)).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText(/宣言/), { target: { value: '3' } });
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '宣言する' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 3 }));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'パス' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
  });

  it('declares a trump without a called card', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: SixBidSoloPhase.DECLARE, declarerIdx: 0, declared: false, highBid: { player: 0, kind: 1 } }),
    );
    renderWithProviders(<SixBidSoloPage />);
    await waitFor(() => expect(screen.getByLabelText(/^切札/)).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText(/^切札/), { target: { value: '3' } });
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '指定する' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declare', { suit: 3 }));

    // コール・ソロでないので指名欄は出ない。
    expect(screen.queryByLabelText(/指名スート/)).not.toBeInTheDocument();
  });

  // **コール・ソロだけが札を指名する。**持っていた人は交換に応じる。
  it('asks for a called card only at a call solo', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: SixBidSoloPhase.DECLARE, declarerIdx: 0, declared: false, highBid: { player: 0, kind: 6 } }),
    );
    renderWithProviders(<SixBidSoloPage />);
    await waitFor(() => expect(screen.getByLabelText(/指名スート/)).toBeInTheDocument());

    expect(screen.getByTestId('sixbidsolo-call-notice')).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText(/^切札/), { target: { value: '2' } });
    fireEvent.change(screen.getByLabelText(/指名スート/), { target: { value: '3' } });
    fireEvent.change(screen.getByLabelText(/指名ランク/), { target: { value: '13' } });

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '指定する' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declare', { suit: 2, calledSuit: 3, calledValue: 13 }));
  });

  it('shows each seat with its points, tricks and running total', async () => {
    mockExec.mockResolvedValue(
      makeState({ players: [seat(0, true, { points: 40, tricksWon: 3, score: 120 }), seat(1, false), seat(2, false)] }),
    );
    renderWithProviders(<SixBidSoloPage />);
    await waitFor(() => expect(screen.getAllByTestId('sixbidsolo-player')).toHaveLength(3));
    const players = screen.getAllByTestId('sixbidsolo-player');
    expect(players[0]).toHaveTextContent('40');
    expect(players[0]).toHaveTextContent('120');
  });

  // **ウィドウ加算はミゼール系では 0。**精算でその区別が読めること。
  it('reports the widow credit at the settlement', async () => {
    mockExec.mockResolvedValue(makeState({ phase: SixBidSoloPhase.HAND_END, lastResult: result() }));
    renderWithProviders(<SixBidSoloPage />);
    await waitFor(() => expect(screen.getByTestId('sixbidsolo-settlement')).toBeInTheDocument());
    const settlement = screen.getByTestId('sixbidsolo-settlement');
    expect(settlement).toHaveTextContent('達成');
    expect(settlement).toHaveTextContent('65');
    expect(settlement).toHaveTextContent('61');
    expect(screen.getByTestId('sixbidsolo-widow-credit')).toHaveTextContent('25');

    mockExec.mockResolvedValue(
      makeState({
        phase: SixBidSoloPhase.HAND_END,
        lastResult: result({ kind: 3, made: false, declarerPoints: 8, widowPoints: 0, target: 0 }),
      }),
    );
    renderWithProviders(<SixBidSoloPage />);
    await waitFor(() => expect(screen.getAllByText(/失敗/).length).toBeGreaterThan(0));
    // ミゼール系はウィドウが加算されない。
    expect(screen.getAllByTestId('sixbidsolo-widow-credit')[1]).toHaveTextContent('0');
  });

  // **スプレッド・ミゼールでは宣言者の手札が公開される。**
  it('says when the declarer hand is exposed', async () => {
    mockExec.mockResolvedValue(makeState({ spreadOpen: true }));
    renderWithProviders(<SixBidSoloPage />);
    await waitFor(() => expect(screen.getByTestId('sixbidsolo-spread')).toBeInTheDocument());
    expect(screen.getByTestId('sixbidsolo-spread')).toHaveTextContent('公開されます');
  });

  it('advances to the next hand', async () => {
    mockExec.mockResolvedValue(makeState({ phase: SixBidSoloPhase.HAND_END, lastResult: result() }));
    renderWithProviders(<SixBidSoloPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次の局へ' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次の局へ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('reports the outcome by seat', async () => {
    for (const [seatIdx, text] of [
      [0, /あなたの勝利です！/],
      [1, /CPUの勝利です。/],
    ] as const) {
      mockExec.mockResolvedValue(
        makeState({ phase: SixBidSoloPhase.GAME_END, gameEndFlag: true, winnerIdx: seatIdx, lastResult: result() }),
      );
      renderWithProviders(<SixBidSoloPage />);
      await waitFor(() => expect(screen.getAllByText(text).length).toBeGreaterThan(0));
    }
  });

  // **ヒント経路はページ側からも踏む。**ファクトリ単体テストだけだと
  // `hintFactories` の登録行と、ページのトグル／ツールチップが一度も
  // 実行されない（#4596 / #4600 のレビュー指摘）。
  it('turns the frontend hint on from the checkbox', async () => {
    localStorage.clear();
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<SixBidSoloPage />);

    const toggle = await screen.findByRole('checkbox', { name: 'ヒント表示' });
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();

    fireEvent.click(toggle);
    expect(await screen.findByTestId('hint-tooltip')).toBeInTheDocument();
  });
});
