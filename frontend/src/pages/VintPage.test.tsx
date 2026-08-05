import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { vintApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { CardDesign, VintHandResult, VintPlayer, VintResponse } from '../types/card';
import { VintPhase } from '../types/phases';
import { VintPage } from './VintPage';

vi.mock('../api/gameApi', () => ({
  vintApi: { exec: vi.fn() },
  actionLogApi: { vint: vi.fn() },
}));

const mockExec = vi.mocked(vintApi.exec);

const card = (design: CardDesign, value: number) => ({ design, value });

function seat(id: number, isHuman: boolean, overrides?: Partial<VintPlayer>): VintPlayer {
  return {
    id,
    isHuman,
    team: id % 2,
    cardCount: 3,
    cards: isHuman ? [card('SPADE', 1), card('HEART', 2), card('CLOVER', 3)] : [],
    tricksWon: 0,
    isDealer: id === 3,
    isDeclarer: false,
    isCurrentTurn: id === 0,
    ...overrides,
  };
}

function result(overrides?: Partial<VintHandResult>): VintHandResult {
  return {
    trickPoints: [210, 180],
    honourPoints: [600, 0],
    acePoints: [1200, 0],
    penalty: [0, 0],
    made: true,
    declarerTricks: 9,
    trickValue: 30,
    ...overrides,
  };
}

function makeState(overrides?: Partial<VintResponse>): VintResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false)],
    phase: VintPhase.PLAY,
    handNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 3,
    bids: [],
    highBid: { player: 0, level: 3, denom: 3, trickValue: 30 },
    declarerIdx: 0,
    trumpSuit: 3,
    trick: [],
    validPlays: [0, 1, 2],
    trickLeaderIdx: 0,
    trickNumber: 0,
    teamTricks: [0, 0],
    below: [0, 0],
    above: [0, 0],
    gamesWon: [0, 0],
    lastResult: null,
    trickValues: [4, 6, 8, 10, 12],
    gameTarget: 500,
    minLevel: 1,
    maxLevel: 7,
    gameEndFlag: false,
    winnerTeam: -1,
    message: '',
    ...overrides,
  };
}

/** The hand-card buttons, in order. */
function handButtons() {
  return screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'play');
}

describe('VintPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<VintPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **ダミーが無いのがブリッジとの違い。**常時表示する。
  it('states that there is no dummy', async () => {
    renderWithProviders(<VintPage />);
    await waitFor(() => expect(screen.getByTestId('vint-no-dummy')).toBeInTheDocument());
    expect(screen.getByTestId('vint-no-dummy')).toHaveTextContent('ダミーはありません');
  });

  // **♠ が最弱で NT が最強。**ブリッジと逆なので単価つきで常時表示する。
  it('shows the reversed bidding order with its trick values', async () => {
    renderWithProviders(<VintPage />);
    await waitFor(() => expect(screen.getByTestId('vint-ladder')).toBeInTheDocument());
    const ladder = screen.getByTestId('vint-ladder');
    const text = ladder.textContent ?? '';

    expect(text).toContain('♠ (4)');
    expect(text).toContain('♣ (6)');
    expect(text).toContain('NT (12)');
    // 並びは ♠ が先頭。
    expect(text.indexOf('♠')).toBeLessThan(text.indexOf('♣'));
    expect(text.indexOf('♣')).toBeLessThan(text.indexOf('NT'));
    expect(ladder).toHaveTextContent('♠が最弱でNTが最強');
  });

  it('plays exactly one card', async () => {
    renderWithProviders(<VintPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    fireEvent.click(handButtons()[1]);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 1 }));
  });

  // **追随は強制。**サーバーが出せる札を決め、それ以外は押せない。
  it('disables cards the server did not list as playable', async () => {
    mockExec.mockResolvedValue(makeState({ validPlays: [2] }));
    renderWithProviders(<VintPage />);
    await waitFor(() => expect(screen.getByTestId('vint-play-notice')).toBeInTheDocument());

    const hand = handButtons();
    expect(hand[0]).toBeDisabled();
    expect(hand[1]).toBeDisabled();
    expect(hand[2]).toBeEnabled();
  });

  it('bids a level and a denomination', async () => {
    mockExec.mockResolvedValue(makeState({ phase: VintPhase.BID, highBid: null, declarerIdx: -1, trumpSuit: 0 }));
    renderWithProviders(<VintPage />);
    await waitFor(() => expect(screen.getByLabelText(/レベル/)).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText(/レベル/), { target: { value: '3' } });
    fireEvent.change(screen.getByLabelText(/スート/), { target: { value: '4' } });

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '宣言する' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { level: 3, denom: 4 }));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'パス' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
  });

  // レベルはサーバーが送る範囲どおりに出す。
  it('offers only the levels the server allows', async () => {
    mockExec.mockResolvedValue(makeState({ phase: VintPhase.BID, highBid: null, declarerIdx: -1 }));
    renderWithProviders(<VintPage />);
    await waitFor(() => expect(screen.getByLabelText(/レベル/)).toBeInTheDocument());

    const select = screen.getByLabelText(/レベル/) as HTMLSelectElement;
    const values = Array.from(select.options).map((o) => o.value);
    expect(values).toEqual(['1', '2', '3', '4', '5', '6', '7']);
  });

  // **パートナーは向かい合わせ。**
  it('shows each seat with its team', async () => {
    renderWithProviders(<VintPage />);
    await waitFor(() => expect(screen.getAllByTestId('vint-player')).toHaveLength(4));
    const players = screen.getAllByTestId('vint-player');
    expect(players[0]).toHaveTextContent('チーム0');
    expect(players[1]).toHaveTextContent('チーム1');
    expect(players[2]).toHaveTextContent('チーム0');
  });

  it('shows both score lines and the game target', async () => {
    mockExec.mockResolvedValue(makeState({ below: [120, 340], above: [50, 0], gamesWon: [1, 0] }));
    renderWithProviders(<VintPage />);
    await waitFor(() => expect(screen.getByTestId('vint-scores')).toBeInTheDocument());
    const scores = screen.getByTestId('vint-scores');
    expect(scores).toHaveTextContent('120');
    expect(scores).toHaveTextContent('340');
    expect(scores).toHaveTextContent('500');
  });

  // **両チームが線下に得点する。**ブリッジとの決定的な違いを精算に書く。
  it('reports both sides below-the-line points at the settlement', async () => {
    mockExec.mockResolvedValue(makeState({ phase: VintPhase.HAND_END, lastResult: result() }));
    renderWithProviders(<VintPage />);
    await waitFor(() => expect(screen.getByTestId('vint-settlement')).toBeInTheDocument());
    const settlement = screen.getByTestId('vint-settlement');
    expect(settlement).toHaveTextContent('宣言達成');
    expect(settlement).toHaveTextContent('210');
    expect(settlement).toHaveTextContent('180');
    expect(settlement).toHaveTextContent('両チームとも');
    // オナーとエースは別勘定であることも書く。
    expect(settlement).toHaveTextContent('3枚以上から');
    expect(settlement).toHaveTextContent('多く持つ側が総取り');
  });

  // ペナルティは未達のときだけ出す。
  it('shows the penalty only when the contract failed', async () => {
    mockExec.mockResolvedValue(makeState({ phase: VintPhase.HAND_END, lastResult: result() }));
    renderWithProviders(<VintPage />);
    await waitFor(() => expect(screen.getByTestId('vint-settlement')).toBeInTheDocument());
    expect(screen.queryByTestId('vint-penalty')).not.toBeInTheDocument();

    mockExec.mockResolvedValue(
      makeState({
        phase: VintPhase.HAND_END,
        lastResult: result({ made: false, declarerTricks: 6, penalty: [0, 1500] }),
      }),
    );
    renderWithProviders(<VintPage />);
    await waitFor(() => expect(screen.getAllByTestId('vint-penalty').length).toBeGreaterThan(0));
    expect(screen.getAllByText(/宣言失敗/).length).toBeGreaterThan(0);
  });

  it('advances to the next hand', async () => {
    mockExec.mockResolvedValue(makeState({ phase: VintPhase.HAND_END, lastResult: result() }));
    renderWithProviders(<VintPage />);
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
        makeState({ phase: VintPhase.GAME_END, gameEndFlag: true, winnerTeam: team, lastResult: result() }),
      );
      renderWithProviders(<VintPage />);
      await waitFor(() => expect(screen.getAllByText(text).length).toBeGreaterThan(0));
    }
  });

  // **ヒント経路はページ側からも踏む。**ファクトリ単体テストだけだと
  // `hintFactories` の登録行と、ページのトグル／ツールチップが一度も
  // 実行されない（#4596 / #4600 のレビュー指摘）。
  it('turns the frontend hint on from the checkbox', async () => {
    localStorage.clear();
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<VintPage />);

    const toggle = await screen.findByRole('checkbox', { name: 'ヒント表示' });
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();

    fireEvent.click(toggle);
    expect(await screen.findByTestId('hint-tooltip')).toBeInTheDocument();
  });

  // **既存の宣言を上回らない組は出せない。**サーバは同格も弾くので、選ばせて
  // からエラーで返すのではなく、選べないようにする (#4940)。
  describe('bid must beat the standing bid', () => {
    const bidding = (over?: Partial<VintResponse>) =>
      makeState({ phase: VintPhase.BID, declarerIdx: -1, currentPlayerIdx: 0, ...over });

    it('leaves every level and denomination open before the first bid', async () => {
      mockExec.mockResolvedValue(bidding({ highBid: null }));
      renderWithProviders(<VintPage />);
      const levels = (await screen.findByLabelText(/レベル/)) as HTMLSelectElement;
      for (const o of Array.from(levels.options)) expect(o.disabled).toBe(false);
      const denoms = screen.getByLabelText(/スート|デノミ/) as HTMLSelectElement;
      for (const o of Array.from(denoms.options)) expect(o.disabled).toBe(false);
      expect(screen.getByTestId('vint-bid-button')).not.toBeDisabled();
    });

    it('closes the levels that no denomination can beat', async () => {
      // ♥4 が立っている → レベル 4 は NT だけ残り、1〜3 は全滅。
      mockExec.mockResolvedValue(bidding({ highBid: { player: 1, level: 4, denom: 3, trickValue: 40 } }));
      renderWithProviders(<VintPage />);
      const levels = (await screen.findByLabelText(/レベル/)) as HTMLSelectElement;
      const byValue = Object.fromEntries(Array.from(levels.options).map((o) => [o.value, o.disabled]));
      expect(byValue['3']).toBe(true);
      expect(byValue['4']).toBe(false);
      expect(byValue['5']).toBe(false);
    });

    // **選択が追い越されたら前に送る。**押せないだけで次に何を選べばよいか
    // 分からない状態を残さない（レビュー指摘）。
    it('moves the pickers to the next legal bid when outbid underneath', async () => {
      mockExec.mockResolvedValue(bidding({ highBid: { player: 1, level: 4, denom: 2, trickValue: 40 } }));
      renderWithProviders(<VintPage />);

      const levels = (await screen.findByLabelText(/レベル/)) as HTMLSelectElement;
      const denoms = screen.getByLabelText(/スート|デノミ/) as HTMLSelectElement;
      // 既定の 1/♠ のままではなく、♦4 の次 = ♥4 に送られている。
      await waitFor(() => expect(levels.value).toBe('4'));
      expect(denoms.value).toBe('3');
      expect(screen.getByTestId('vint-bid-button')).not.toBeDisabled();
      expect(screen.queryByTestId('vint-bid-too-low')).not.toBeInTheDocument();
    });

    // **同格も通らない。**「レベルが上でなければ不可」だと出せる宣言を潰しすぎ、
    // 「以上なら可」だと弾かれる宣言を残す。
    it('disables the denominations at or below the standing bid', async () => {
      mockExec.mockResolvedValue(bidding({ highBid: { player: 1, level: 3, denom: 2, trickValue: 30 } }));
      renderWithProviders(<VintPage />);
      const levels = (await screen.findByLabelText(/レベル/)) as HTMLSelectElement;
      fireEvent.change(levels, { target: { value: '3' } });

      const denoms = screen.getByLabelText(/スート|デノミ/) as HTMLSelectElement;
      const disabled = Array.from(denoms.options).map((o) => o.disabled);
      // 0(♠) 1(♣) 2(♦) は不可、3(♥) 4(NT) は可。
      expect(disabled).toEqual([true, true, true, false, false]);
    });
  });
});
