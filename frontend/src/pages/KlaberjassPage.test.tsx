import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { klaberjassApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { CardDesign, KlaberjassPlayer, KlaberjassResponse } from '../types/card';
import { KlaberjassPhase } from '../types/phases';
import { KlaberjassPage } from './KlaberjassPage';

vi.mock('../api/gameApi', () => ({
  klaberjassApi: { exec: vi.fn() },
  actionLogApi: { klaberjass: vi.fn() },
}));

const mockExec = vi.mocked(klaberjassApi.exec);

const card = (design: CardDesign, value: number) => ({ design, value });

function seat(id: number, isHuman: boolean, overrides?: Partial<KlaberjassPlayer>): KlaberjassPlayer {
  return {
    id,
    isHuman,
    cardCount: 3,
    cards: isHuman ? [card('SPADE', 11), card('HEART', 1), card('CLOVER', 7)] : [],
    sequences: [],
    handPoints: 0,
    score: 0,
    isMaker: false,
    isDealer: id === 0,
    isCurrentTurn: id === 0,
    ...overrides,
  };
}

function makeState(overrides?: Partial<KlaberjassResponse>): KlaberjassResponse {
  return {
    players: [seat(0, true), seat(1, false, { isMaker: true })],
    phase: KlaberjassPhase.PLAY,
    dealNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 0,
    trumpSuit: 1,
    turnUpCard: null,
    makerIdx: 1,
    trick: [],
    trickLeaderIdx: 0,
    trickNumber: 0,
    validPlays: [0, 1, 2],
    sequenceWinner: -1,
    belaHolder: -1,
    belaScored: false,
    dixUsed: false,
    bete: false,
    schmeissBy: -1,
    targetScore: 501,
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

describe('KlaberjassPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<KlaberjassPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **切札の J と 9 が A の上に来る**のが最大の落とし穴なので、常時表示する。
  it('shows the trump order permanently', async () => {
    renderWithProviders(<KlaberjassPage />);
    await waitFor(() => expect(screen.getByTestId('klaberjass-ladder')).toBeInTheDocument());
    const ladder = screen.getByTestId('klaberjass-ladder');
    expect(ladder).toHaveTextContent('J (20)');
    expect(ladder).toHaveTextContent('9 (14)');
    expect(ladder).toHaveTextContent('A (11)');
    // シーケンスは点数順ではないという注意書きも出す。
    expect(ladder).toHaveTextContent(/7-8-9-10-J-Q-K-A/);
  });

  it('plays exactly one card', async () => {
    renderWithProviders(<KlaberjassPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    pickHand(1);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 1 }));
  });

  // **追随・切札・上乗せが全部強制。**サーバーが出せる札を決め、それ以外は押せない。
  it('disables cards the server did not list as playable', async () => {
    mockExec.mockResolvedValue(makeState({ validPlays: [2] }));
    renderWithProviders(<KlaberjassPage />);
    await waitFor(() => expect(screen.getByTestId('klaberjass-play-notice')).toBeInTheDocument());

    const hand = screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'play');
    expect(hand[0]).toBeDisabled();
    expect(hand[1]).toBeDisabled();
    expect(hand[2]).toBeEnabled();
  });

  it('takes the turn-up suit or passes in the first round', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: KlaberjassPhase.BID_TURN_UP, trumpSuit: 0, turnUpCard: card('HEART', 13), makerIdx: -1 }),
    );
    renderWithProviders(<KlaberjassPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '取る' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '取る' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('accept'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'パス' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
  });

  // **断られた表向きのスートは第2ラウンドで選べない。**ボタン自体を出さない。
  it('withholds the refused turn-up suit in the free round', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: KlaberjassPhase.BID_FREE, trumpSuit: 0, turnUpCard: card('HEART', 13), makerIdx: -1 }),
    );
    renderWithProviders(<KlaberjassPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /♠ を切札に/ })).toBeInTheDocument());

    expect(screen.queryByRole('button', { name: /♥ を切札に/ })).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: /♣ を切札に/ })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /♦ を切札に/ })).toBeInTheDocument();

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: /♠ を切札に/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('call', { suit: 1 }));
  });

  // **拒否は「流さない」ではなく「相手を宣言側にする」。**案内でそう伝える。
  it('warns that refusing a schmeiss makes the thrower the maker', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: KlaberjassPhase.SCHMEISS, trumpSuit: 0, turnUpCard: card('HEART', 13), schmeissBy: 1 }),
    );
    renderWithProviders(<KlaberjassPage />);
    await waitFor(() => expect(screen.getByTestId('klaberjass-schmeiss-notice')).toBeInTheDocument());
    expect(screen.getByTestId('klaberjass-schmeiss-notice')).toHaveTextContent('相手が】宣言側');

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '拒否する' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('answerschmeiss', { accept: false }));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '同意する' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('answerschmeiss', { accept: true }));
  });

  it('offers to throw the deal in while bidding', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: KlaberjassPhase.BID_TURN_UP, trumpSuit: 0, turnUpCard: card('HEART', 13), makerIdx: -1 }),
    );
    renderWithProviders(<KlaberjassPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '投げる' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '投げる' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('schmeiss'));
  });

  // ベートかどうかは精算の意味がまるで違うので、字面で分ける。
  it('tells a bete apart from an ordinary settlement', async () => {
    mockExec.mockResolvedValue(makeState({ phase: KlaberjassPhase.HAND_END, bete: true }));
    renderWithProviders(<KlaberjassPage />);
    await waitFor(() => expect(screen.getByTestId('klaberjass-settlement')).toBeInTheDocument());
    expect(screen.getByTestId('klaberjass-settlement')).toHaveTextContent('得点は全部相手のもの');

    mockExec.mockResolvedValue(makeState({ phase: KlaberjassPhase.HAND_END, bete: false }));
    renderWithProviders(<KlaberjassPage />);
    await waitFor(() => expect(screen.getAllByText(/両者とも自分の点/).length).toBeGreaterThan(0));
  });

  // 役の勝負は「良い方が全部」か「引き分けで誰も取らない」かのどちらか。
  it('reports the sequence contest either way', async () => {
    mockExec.mockResolvedValue(makeState({ phase: KlaberjassPhase.HAND_END, sequenceWinner: 0 }));
    renderWithProviders(<KlaberjassPage />);
    await waitFor(() => expect(screen.getByTestId('klaberjass-settlement')).toHaveTextContent('総取り'));

    mockExec.mockResolvedValue(makeState({ phase: KlaberjassPhase.HAND_END, sequenceWinner: -1 }));
    renderWithProviders(<KlaberjassPage />);
    await waitFor(() => expect(screen.getAllByText(/どちらも得点なし/).length).toBeGreaterThan(0));
  });

  it('reports the bela and the dix when they happened', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: KlaberjassPhase.HAND_END, belaHolder: 0, belaScored: true, dixUsed: true }),
    );
    renderWithProviders(<KlaberjassPage />);
    await waitFor(() => expect(screen.getByTestId('klaberjass-settlement')).toHaveTextContent('ベラ成立'));
    expect(screen.getByTestId('klaberjass-settlement')).toHaveTextContent('切札の7');
  });

  it('advances to the next deal', async () => {
    mockExec.mockResolvedValue(makeState({ phase: KlaberjassPhase.HAND_END }));
    renderWithProviders(<KlaberjassPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のディールへ' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のディールへ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('reports each outcome', async () => {
    for (const [winner, text] of [
      [0, /あなたの勝利です！/],
      [1, /CPU 1 の勝利です。/],
    ] as const) {
      mockExec.mockResolvedValue(makeState({ phase: KlaberjassPhase.GAME_END, gameEndFlag: true, winnerIdx: winner }));
      renderWithProviders(<KlaberjassPage />);
      await waitFor(() => expect(screen.getAllByText(text).length).toBeGreaterThan(0));
    }
  });
});
