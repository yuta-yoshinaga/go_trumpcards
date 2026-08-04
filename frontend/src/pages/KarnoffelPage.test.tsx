import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { karnoffelApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { CardDesign, KarnoffelHandResult, KarnoffelPlayer, KarnoffelResponse } from '../types/card';
import { KarnoffelPhase } from '../types/phases';
import { KarnoffelPage } from './KarnoffelPage';

vi.mock('../api/gameApi', () => ({
  karnoffelApi: { exec: vi.fn() },
  actionLogApi: { karnoffel: vi.fn() },
}));

const mockExec = vi.mocked(karnoffelApi.exec);

const card = (design: CardDesign, value: number) => ({ design, value });

function seat(id: number, isHuman: boolean, overrides?: Partial<KarnoffelPlayer>): KarnoffelPlayer {
  return {
    id,
    isHuman,
    team: id % 2,
    cardCount: 3,
    cards: isHuman ? [card('SPADE', 13), card('HEART', 11), card('CLOVER', 6)] : [],
    upCard: card('HEART', 3 + id),
    tricksWon: 0,
    isDealer: id === 3,
    isCurrentTurn: id === 0,
    ...overrides,
  };
}

function result(overrides?: Partial<KarnoffelHandResult>): KarnoffelHandResult {
  return { winnerTeam: 0, tricks: [3, 1], chosenSuit: 3, ...overrides };
}

function makeState(overrides?: Partial<KarnoffelResponse>): KarnoffelResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false)],
    phase: KarnoffelPhase.PLAY,
    handNumber: 1,
    currentPlayerIdx: 0,
    dealerIdx: 3,
    chosenSuit: 3,
    trick: [],
    validPlays: [0, 1, 2],
    trickLeaderIdx: 0,
    trickNumber: 0,
    teamTricks: [0, 0],
    handsWon: [0, 0],
    lastResult: null,
    tricksToWin: 3,
    handSize: 5,
    targetHands: 3,
    gameEndFlag: false,
    winnerTeam: -1,
    message: '',
    config: { cpuDifficulty: 0, targetHands: 3 },
    ...overrides,
  };
}

/** The hand-card buttons, in order. */
function handButtons() {
  return screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'play');
}

describe('KarnoffelPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<KarnoffelPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **切札は表向きの4枚のうち最も低い札が決める。**「最後の1枚をめくる」ではない。
  it('explains how the chosen suit was picked', async () => {
    renderWithProviders(<KarnoffelPage />);
    await waitFor(() => expect(screen.getByTestId('karnoffel-chosen-note')).toBeInTheDocument());
    expect(screen.getByTestId('karnoffel-chosen-note')).toHaveTextContent('表向きに配られた4枚のうち最も低い札');
    expect(screen.getByTestId('karnoffel-chosen')).toHaveTextContent('♥');
  });

  // **序列が普通と違うので常時表示する。**悪魔の特殊性は文章でしか伝わらない。
  it('shows the irregular ranking with the devil rule', async () => {
    renderWithProviders(<KarnoffelPage />);
    await waitFor(() => expect(screen.getByTestId('karnoffel-ladder')).toBeInTheDocument());
    const ladder = screen.getByTestId('karnoffel-ladder');
    expect(ladder).toHaveTextContent('J（カルニッフェル）');
    expect(ladder).toHaveTextContent('7（悪魔・リード時のみ）');
    expect(ladder).toHaveTextContent('6（法王）');
    // **部分切札の負け方も出す。**
    expect(ladder).toHaveTextContent('3はKに、4はK・Qに、5は絵札すべてに負けます');
    expect(ladder).toHaveTextContent('追随して出した7はあらゆる札に負けます');

    const text = ladder.textContent ?? '';
    expect(text.indexOf('J（カルニッフェル）')).toBeLessThan(text.indexOf('7（悪魔'));
  });

  // **表向きの札は全員ぶん見える。**切札の根拠がそこにある。
  it('shows every seat face-up card', async () => {
    renderWithProviders(<KarnoffelPage />);
    await waitFor(() => expect(screen.getAllByTestId('karnoffel-player')).toHaveLength(4));
    for (const player of screen.getAllByTestId('karnoffel-player')) {
      expect(player).toHaveTextContent('表');
    }
  });

  it('plays exactly one card', async () => {
    renderWithProviders(<KarnoffelPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    fireEvent.click(handButtons()[1]);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 1 }));
  });

  // **追随の義務は無い。**制限は「第1トリックのリードに悪魔は使えない」だけ。
  it('says there is no follow-suit rule but the devil cannot open', async () => {
    renderWithProviders(<KarnoffelPage />);
    await waitFor(() => expect(screen.getByTestId('karnoffel-play-notice')).toBeInTheDocument());
    const notice = screen.getByTestId('karnoffel-play-notice');
    expect(notice).toHaveTextContent('追随の義務はありません');
    expect(notice).toHaveTextContent('第1トリックのリードに悪魔は使えません');
  });

  // その一つの制限はサーバーが決める。
  it('disables cards the server did not list as playable', async () => {
    mockExec.mockResolvedValue(makeState({ validPlays: [1, 2] }));
    renderWithProviders(<KarnoffelPage />);
    await waitFor(() => expect(screen.getByTestId('karnoffel-play-notice')).toBeInTheDocument());

    const hand = handButtons();
    expect(hand[0]).toBeDisabled();
    expect(hand[1]).toBeEnabled();
    expect(hand[2]).toBeEnabled();
  });

  // **パートナーは向かい合わせ。**
  it('shows each seat with its team', async () => {
    renderWithProviders(<KarnoffelPage />);
    await waitFor(() => expect(screen.getAllByTestId('karnoffel-player')).toHaveLength(4));
    const players = screen.getAllByTestId('karnoffel-player');
    expect(players[0]).toHaveTextContent('チーム0');
    expect(players[1]).toHaveTextContent('チーム1');
    expect(players[2]).toHaveTextContent('チーム0');
  });

  // **3トリック先取で1局、規定局数先取で勝ち。**
  it('shows the hands won and what it takes to win', async () => {
    mockExec.mockResolvedValue(makeState({ handsWon: [2, 1], teamTricks: [1, 2] }));
    renderWithProviders(<KarnoffelPage />);
    await waitFor(() => expect(screen.getByTestId('karnoffel-scores')).toBeInTheDocument());
    const scores = screen.getByTestId('karnoffel-scores');
    expect(scores).toHaveTextContent('2局');
    expect(scores).toHaveTextContent('3局先取');
    expect(scores).toHaveTextContent('3トリック先取');
  });

  // **3トリックに届かなければ勝者なし。**その区別を出す。
  it('reports the hand result, drawn hands included', async () => {
    mockExec.mockResolvedValue(makeState({ phase: KarnoffelPhase.HAND_END, lastResult: result() }));
    renderWithProviders(<KarnoffelPage />);
    await waitFor(() => expect(screen.getByTestId('karnoffel-result')).toBeInTheDocument());
    expect(screen.getByTestId('karnoffel-result')).toHaveTextContent('チーム0が局を取りました');

    mockExec.mockResolvedValue(
      makeState({
        phase: KarnoffelPhase.HAND_END,
        lastResult: result({ winnerTeam: -1, tricks: [2, 2] }),
      }),
    );
    renderWithProviders(<KarnoffelPage />);
    await waitFor(() => expect(screen.getAllByText(/どちらも3トリックに届きませんでした/).length).toBeGreaterThan(0));
  });

  it('advances to the next hand', async () => {
    mockExec.mockResolvedValue(makeState({ phase: KarnoffelPhase.HAND_END, lastResult: result() }));
    renderWithProviders(<KarnoffelPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次の局へ' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次の局へ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  // **勝敗はチームで決まる。**
  it('reports the outcome by team', async () => {
    for (const [team, text] of [
      [0, /あなたのチームの勝利です！/],
      [1, /相手チームの勝利です。/],
    ] as const) {
      mockExec.mockResolvedValue(
        makeState({ phase: KarnoffelPhase.GAME_END, gameEndFlag: true, winnerTeam: team, lastResult: result() }),
      );
      renderWithProviders(<KarnoffelPage />);
      await waitFor(() => expect(screen.getAllByText(text).length).toBeGreaterThan(0));
    }
  });

  // **ヒント経路はページ側からも踏む。**ファクトリ単体テストだけだと
  // `hintFactories` の登録行と、ページのトグル／ツールチップが一度も
  // 実行されない（#4596 / #4600 のレビュー指摘）。
  it('turns the frontend hint on from the checkbox', async () => {
    localStorage.clear();
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<KarnoffelPage />);

    const toggle = await screen.findByRole('checkbox', { name: 'ヒント表示' });
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();

    fireEvent.click(toggle);
    expect(await screen.findByTestId('hint-tooltip')).toBeInTheDocument();
  });

  it('badges only the titled cards of the suit chosen this deal', async () => {
    localStorage.clear();
    mockExec.mockReset();
    // Hand is ♠K, ♥J, ♣6 with hearts chosen: only the ♥J is the Karnöffel.
    // The ♣6 would be the Pope in clubs, which is exactly the trap.
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<KarnoffelPage />);
    await waitFor(() => expect(screen.getByTestId('karnoffel-rank-karnoffel')).toBeInTheDocument());
    expect(screen.queryByTestId('karnoffel-rank-pope')).not.toBeInTheDocument();
    expect(document.querySelectorAll('[data-testid^="karnoffel-rank-"]')).toHaveLength(1);
  });
});
