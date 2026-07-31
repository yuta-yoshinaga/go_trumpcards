import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { bidEuchreApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BidEuchreHandResult, BidEuchrePlayer, BidEuchreResponse, CardDesign } from '../types/card';
import { BidEuchrePhase } from '../types/phases';
import { BidEuchrePage } from './BidEuchrePage';

vi.mock('../api/gameApi', () => ({
  bidEuchreApi: { exec: vi.fn() },
  actionLogApi: { bideuchre: vi.fn() },
}));

const mockExec = vi.mocked(bidEuchreApi.exec);

const card = (design: CardDesign, value: number) => ({ design, value });

function seat(id: number, isHuman: boolean, overrides?: Partial<BidEuchrePlayer>): BidEuchrePlayer {
  return {
    id,
    isHuman,
    team: id % 2,
    cardCount: 3,
    cards: isHuman ? [card('SPADE', 1), card('HEART', 11), card('CLOVER', 11)] : [],
    tricksWon: 0,
    isDealer: id === 3,
    isDeclarer: false,
    isCurrentTurn: id === 0,
    ...overrides,
  };
}

function result(overrides?: Partial<BidEuchreHandResult>): BidEuchreHandResult {
  return {
    points: [4, 2],
    tricks: [4, 2],
    made: true,
    bid: 3,
    ...overrides,
  };
}

function makeState(overrides?: Partial<BidEuchreResponse>): BidEuchreResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false)],
    phase: BidEuchrePhase.PLAY,
    handNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 3,
    bids: [],
    highBid: { player: 0, value: 4 },
    declarerIdx: 0,
    trump: 0,
    trumpSuit: 1,
    trumpChosen: true,
    trick: [],
    validPlays: [0, 1, 2],
    trickLeaderIdx: 0,
    trickNumber: 0,
    teamTricks: [0, 0],
    scores: [0, 0],
    lastResult: null,
    gameTarget: 32,
    minBid: 3,
    maxBid: 6,
    handSize: 6,
    gameEndFlag: false,
    winnerTeam: -1,
    config: { cpuDifficulty: 0, allowNoTrump: true },
    message: '',
    ...overrides,
  };
}

/** The hand-card buttons, in order. */
function handButtons() {
  return screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'play');
}

describe('BidEuchrePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<BidEuchrePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **24 / 4 = 6 でちょうど配り切るのでキティが無い。**常時表示する。
  it('states that there is no kitty', async () => {
    renderWithProviders(<BidEuchrePage />);
    await waitFor(() => expect(screen.getByTestId('bideuchre-no-kitty')).toBeInTheDocument());
    expect(screen.getByTestId('bideuchre-no-kitty')).toHaveTextContent('キティはありません');
  });

  it('plays exactly one card', async () => {
    renderWithProviders(<BidEuchrePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    fireEvent.click(handButtons()[1]);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 1 }));
  });

  // **追随は強制で、左ボワーは切札扱い。**サーバーが出せる札を決める。
  it('disables cards the server did not list as playable', async () => {
    mockExec.mockResolvedValue(makeState({ validPlays: [2] }));
    renderWithProviders(<BidEuchrePage />);
    await waitFor(() => expect(screen.getByTestId('bideuchre-play-notice')).toBeInTheDocument());

    const hand = handButtons();
    expect(hand[0]).toBeDisabled();
    expect(hand[1]).toBeDisabled();
    expect(hand[2]).toBeEnabled();
    expect(screen.getByTestId('bideuchre-play-notice')).toHaveTextContent('左ボワー');
  });

  it('bids a trick count', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: BidEuchrePhase.BID, highBid: null, declarerIdx: -1, trumpChosen: false }),
    );
    renderWithProviders(<BidEuchrePage />);
    await waitFor(() => expect(screen.getByLabelText(/宣言/)).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText(/宣言/), { target: { value: '5' } });

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '宣言する' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { value: 5 }));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'パス' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
  });

  // **最低は 3。**サーバーが送る範囲どおりに出す。
  it('offers only three tricks and up', async () => {
    mockExec.mockResolvedValue(makeState({ phase: BidEuchrePhase.BID, highBid: null, declarerIdx: -1 }));
    renderWithProviders(<BidEuchrePage />);
    await waitFor(() => expect(screen.getByLabelText(/宣言/)).toBeInTheDocument());

    const select = screen.getByLabelText(/宣言/) as HTMLSelectElement;
    expect(Array.from(select.options).map((o) => o.value)).toEqual(['3', '4', '5', '6']);
  });

  // **上回る宣言だけが通るので、通らない値は出さない。**
  it('offers only bids that can actually stand', async () => {
    // 人間は親ではない (dealerIdx=3)。立っている 4 を上回る 5-6 だけ。
    mockExec.mockResolvedValue(
      makeState({ phase: BidEuchrePhase.BID, highBid: { player: 1, value: 4 }, declarerIdx: -1, dealerIdx: 3 }),
    );
    renderWithProviders(<BidEuchrePage />);
    await waitFor(() => expect(screen.getByLabelText(/宣言/)).toBeInTheDocument());
    expect(Array.from((screen.getByLabelText(/宣言/) as HTMLSelectElement).options).map((o) => o.value)).toEqual([
      '5',
      '6',
    ]);
  });

  // **親だけは同額でも奪えるので、同額も出す。**
  it('lets the dealer pick the equal bid', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: BidEuchrePhase.BID, highBid: { player: 1, value: 4 }, declarerIdx: -1, dealerIdx: 0 }),
    );
    renderWithProviders(<BidEuchrePage />);
    await waitFor(() => expect(screen.getByLabelText(/宣言/)).toBeInTheDocument());
    expect(Array.from((screen.getByLabelText(/宣言/) as HTMLSelectElement).options).map((o) => o.value)).toEqual([
      '4',
      '5',
      '6',
    ]);

    // 既定の 3 は選べないので、出せる最小値へ寄せて送る。
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '宣言する' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { value: 4 }));
  });

  // **ノートランプは設定で切れる。**サーバーが弾く選択肢は出さない。
  it('hides the no-trump forms when the config switches them off', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: BidEuchrePhase.CHOOSE_TRUMP,
        declarerIdx: 0,
        trumpChosen: false,
        config: { cpuDifficulty: 0, allowNoTrump: false },
      }),
    );
    renderWithProviders(<BidEuchrePage />);
    await waitFor(() => expect(screen.getByLabelText(/切札/)).toBeInTheDocument());
    expect(Array.from((screen.getByLabelText(/切札/) as HTMLSelectElement).options).map((o) => o.textContent)).toEqual([
      '♠',
      '♣',
      '♦',
      '♥',
    ]);
  });

  // **親だけは同額でも奪える。**入札画面で読めること。
  it('says the dealer alone may equal the standing bid', async () => {
    mockExec.mockResolvedValue(makeState({ phase: BidEuchrePhase.BID, highBid: null, declarerIdx: -1 }));
    renderWithProviders(<BidEuchrePage />);
    await waitFor(() => expect(screen.getByTestId('bideuchre-dealer-note')).toBeInTheDocument());
    expect(screen.getByTestId('bideuchre-dealer-note')).toHaveTextContent('親だけは同額');
  });

  // **ノートランプが 2 種類あり、ローは序列が逆転する。**
  it('offers both no-trump forms and warns that low reverses the ranking', async () => {
    mockExec.mockResolvedValue(makeState({ phase: BidEuchrePhase.CHOOSE_TRUMP, declarerIdx: 0, trumpChosen: false }));
    renderWithProviders(<BidEuchrePage />);
    await waitFor(() => expect(screen.getByLabelText(/切札/)).toBeInTheDocument());

    const select = screen.getByLabelText(/切札/) as HTMLSelectElement;
    const labels = Array.from(select.options).map((o) => o.textContent);
    expect(labels).toEqual(['♠', '♣', '♦', '♥', 'NTハイ', 'NTロー']);
    expect(screen.getByTestId('bideuchre-ntlow-note')).toHaveTextContent('9が最強');

    fireEvent.change(select, { target: { value: '5' } });
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '指定する' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('trump', { trump: 5 }));
  });

  // **パートナーは向かい合わせ。**
  it('shows each seat with its team', async () => {
    renderWithProviders(<BidEuchrePage />);
    await waitFor(() => expect(screen.getAllByTestId('bideuchre-player')).toHaveLength(4));
    const players = screen.getAllByTestId('bideuchre-player');
    expect(players[0]).toHaveTextContent('チーム0');
    expect(players[1]).toHaveTextContent('チーム1');
    expect(players[2]).toHaveTextContent('チーム0');
  });

  it('shows both scores and the 32-point target', async () => {
    mockExec.mockResolvedValue(makeState({ scores: [18, 25] }));
    renderWithProviders(<BidEuchrePage />);
    await waitFor(() => expect(screen.getByTestId('bideuchre-scores')).toBeInTheDocument());
    const scores = screen.getByTestId('bideuchre-scores');
    expect(scores).toHaveTextContent('18');
    expect(scores).toHaveTextContent('25');
    expect(scores).toHaveTextContent('32');
  });

  // **未達で失うのは宣言額。**取ったトリック数ではない。
  it('explains that a set costs the bid, not the tricks taken', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: BidEuchrePhase.HAND_END,
        lastResult: result({ made: false, bid: 5, points: [-5, 4], tricks: [2, 4] }),
      }),
    );
    renderWithProviders(<BidEuchrePage />);
    await waitFor(() => expect(screen.getByTestId('bideuchre-settlement')).toBeInTheDocument());
    const settlement = screen.getByTestId('bideuchre-settlement');
    expect(settlement).toHaveTextContent('宣言失敗');
    expect(settlement).toHaveTextContent('-5');
    expect(screen.getByTestId('bideuchre-set-note')).toHaveTextContent('未達で失うのは宣言額');
    expect(screen.getByTestId('bideuchre-set-note')).toHaveTextContent('守備側は未達でも取ったトリックを得点');
  });

  // 達成時はその注記を出さない。
  it('omits the set note when the contract was made', async () => {
    mockExec.mockResolvedValue(makeState({ phase: BidEuchrePhase.HAND_END, lastResult: result() }));
    renderWithProviders(<BidEuchrePage />);
    await waitFor(() => expect(screen.getByTestId('bideuchre-settlement')).toBeInTheDocument());
    expect(screen.getByTestId('bideuchre-settlement')).toHaveTextContent('宣言達成');
    expect(screen.queryByTestId('bideuchre-set-note')).not.toBeInTheDocument();
  });

  it('advances to the next hand', async () => {
    mockExec.mockResolvedValue(makeState({ phase: BidEuchrePhase.HAND_END, lastResult: result() }));
    renderWithProviders(<BidEuchrePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次の局へ' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次の局へ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  // **勝敗はチームで決まる。**席ではない。
  it('reports the outcome by team', async () => {
    for (const [team, text] of [
      [0, /あなたのチームの勝利です！/],
      [1, /相手チームの勝利です。/],
    ] as const) {
      mockExec.mockResolvedValue(
        makeState({ phase: BidEuchrePhase.GAME_END, gameEndFlag: true, winnerTeam: team, lastResult: result() }),
      );
      renderWithProviders(<BidEuchrePage />);
      await waitFor(() => expect(screen.getAllByText(text).length).toBeGreaterThan(0));
    }
  });
});
