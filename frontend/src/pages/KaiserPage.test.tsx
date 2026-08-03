import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { kaiserApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { CardDesign, KaiserPlayer, KaiserResponse } from '../types/card';
import { KaiserPhase } from '../types/phases';
import { KaiserPage } from './KaiserPage';

vi.mock('../api/gameApi', () => ({
  kaiserApi: { exec: vi.fn() },
  actionLogApi: { kaiser: vi.fn() },
}));

const mockExec = vi.mocked(kaiserApi.exec);

const card = (design: CardDesign, value: number) => ({ design, value });

/** The human hand: the two scoring cards plus an ordinary one. */
const HUMAN_HAND = [card('HEART', 5), card('SPADE', 3), card('CLOVER', 7)];

function seat(id: number, isHuman: boolean, overrides?: Partial<KaiserPlayer>): KaiserPlayer {
  return {
    id,
    isHuman,
    team: id % 2,
    cardCount: 3,
    cards: isHuman ? HUMAN_HAND : [],
    isDealer: id === 3,
    isDeclarer: false,
    isCurrentTurn: id === 0,
    ...overrides,
  };
}

function makeState(overrides?: Partial<KaiserResponse>): KaiserResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false)],
    phase: KaiserPhase.PLAY,
    handNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 3,
    bids: [],
    highBid: { player: 0, value: 8, contract: 0 },
    declarerIdx: 0,
    trumpSuit: 3,
    contract: 0,
    kittySize: 0,
    trick: [],
    trickLeaderIdx: 0,
    trickNumber: 0,
    validPlays: [0, 1, 2],
    teamHandPoints: [0, 0],
    teamScores: [0, 0],
    heartFiveBy: -1,
    spadeThreeBy: -1,
    bidMade: false,
    targetScore: 52,
    minBid: 7,
    maxBid: 12,
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

describe('KaiserPage', () => {
  // **サーバーが弾く選択肢は出さない。**設定でノートランプを切っていると
  // Bid が error を返すので、選べてしまうと押した瞬間に必ず失敗する。
  describe('the no-trump setting', () => {
    const biddingState = (allowNoTrump: boolean) =>
      makeState({
        phase: KaiserPhase.BID,
        highBid: null,
        declarerIdx: -1,
        trumpSuit: 0,
        config: { cpuDifficulty: 0, allowNoTrump },
      });

    it('offers all three contracts while it is on', async () => {
      mockExec.mockResolvedValue(biddingState(true));
      renderWithProviders(<KaiserPage />);
      await waitFor(() => expect(screen.getByLabelText(/契約/)).toBeInTheDocument());
      const select = screen.getByLabelText(/契約/) as HTMLSelectElement;
      expect(select.options).toHaveLength(3);
    });

    it('offers only the trump contract while it is off', async () => {
      mockExec.mockResolvedValue(biddingState(false));
      renderWithProviders(<KaiserPage />);
      await waitFor(() => expect(screen.getByLabelText(/契約/)).toBeInTheDocument());
      const select = screen.getByLabelText(/契約/) as HTMLSelectElement;
      expect(select.options).toHaveLength(1);
      expect(select.options[0]?.textContent).toBe('切札あり');
    });
  });

  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<KaiserPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **♥5 と ♠3 でトリック8点と同じ重みが動く。**常時表示する。
  it('states both scoring cards permanently', async () => {
    renderWithProviders(<KaiserPage />);
    await waitFor(() => expect(screen.getByTestId('kaiser-specials')).toBeInTheDocument());
    const specials = screen.getByTestId('kaiser-specials');
    expect(specials).toHaveTextContent('♥5 = +5点');
    expect(specials).toHaveTextContent('♠3 = −3点');
    // デッキが 34 枚であることも書く。issue の 32 枚が最大の誤り。
    expect(specials).toHaveTextContent('34枚');
  });

  it('plays exactly one card', async () => {
    renderWithProviders(<KaiserPage />);
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
    renderWithProviders(<KaiserPage />);
    await waitFor(() => expect(screen.getByTestId('kaiser-play-notice')).toBeInTheDocument());

    const hand = handButtons();
    expect(hand[0]).toBeDisabled();
    expect(hand[1]).toBeDisabled();
    expect(hand[2]).toBeEnabled();
  });

  // **宣言するのは点数でトリック数ではない。**最低は 7。
  it('bids points from the minimum the server sends', async () => {
    mockExec.mockResolvedValue(makeState({ phase: KaiserPhase.BID, highBid: null, declarerIdx: -1, trumpSuit: 0 }));
    renderWithProviders(<KaiserPage />);
    await waitFor(() => expect(screen.getByTestId('kaiser-bid-notice')).toBeInTheDocument());
    expect(screen.getByTestId('kaiser-bid-notice')).toHaveTextContent('トリック数ではありません');

    // 6 以下のボタンは出ない。
    expect(screen.queryByRole('button', { name: '6 を宣言' })).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: '7 を宣言' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '12 を宣言' })).toBeInTheDocument();

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '8 を宣言' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 8, contract: 0 }));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'パス' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
  });

  // ロー・ノートランプはランクが逆転する別契約なので選べる必要がある。
  it('carries the chosen contract into the bid', async () => {
    mockExec.mockResolvedValue(makeState({ phase: KaiserPhase.BID, highBid: null, declarerIdx: -1, trumpSuit: 0 }));
    renderWithProviders(<KaiserPage />);
    await waitFor(() => expect(screen.getByLabelText(/契約/)).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText(/契約/), { target: { value: '2' } });
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '9 を宣言' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 9, contract: 2 }));
  });

  // **切札を決めてからでないと捨てられない。**捨てる判断が切札に依る。
  it('asks for trump before the discard', async () => {
    mockExec.mockResolvedValue(makeState({ phase: KaiserPhase.DISCARD, trumpSuit: 0, kittySize: 0 }));
    renderWithProviders(<KaiserPage />);
    await waitFor(() => expect(screen.getByTestId('kaiser-trump-notice')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '捨てる' })).not.toBeInTheDocument();

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: /♥ を切札に/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('trump', { suit: 3 }));
  });

  // **♥5 と ♠3 は捨てられない。**押せないようにする。
  it('refuses to select either scoring card for the discard', async () => {
    mockExec.mockResolvedValue(makeState({ phase: KaiserPhase.DISCARD }));
    renderWithProviders(<KaiserPage />);
    await waitFor(() => expect(screen.getByTestId('kaiser-discard-notice')).toBeInTheDocument());
    expect(screen.getByTestId('kaiser-discard-notice')).toHaveTextContent('♥5 と ♠3 は捨てられません');

    const hand = handButtons();
    expect(hand[0]).toBeDisabled(); // ♥5
    expect(hand[1]).toBeDisabled(); // ♠3
    expect(hand[2]).toBeEnabled();
  });

  // 捨てるのはちょうど 2 枚。
  it('discards exactly two cards', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: KaiserPhase.DISCARD,
        players: [
          seat(0, true, {
            isDeclarer: true,
            cards: [card('CLOVER', 7), card('CLOVER', 8), card('CLOVER', 9)],
          }),
          seat(1, false),
          seat(2, false),
          seat(3, false),
        ],
      }),
    );
    renderWithProviders(<KaiserPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨てる' })).toBeInTheDocument());

    const discard = screen.getByRole('button', { name: '捨てる' });
    expect(discard).toBeDisabled();

    fireEvent.click(handButtons()[0]);
    expect(screen.getByRole('button', { name: '捨てる' })).toBeDisabled();

    fireEvent.click(handButtons()[2]);
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '捨てる' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', { indices: [0, 2] }));
  });

  // **パートナーは向かい合わせ。**誰が味方か読めないと戦えない。
  it('shows each seat with its team', async () => {
    renderWithProviders(<KaiserPage />);
    await waitFor(() => expect(screen.getAllByTestId('kaiser-player')).toHaveLength(4));
    const players = screen.getAllByTestId('kaiser-player');
    expect(players[0]).toHaveTextContent('チーム0');
    expect(players[1]).toHaveTextContent('チーム1');
    expect(players[2]).toHaveTextContent('チーム0');
  });

  it('reports where the scoring cards went', async () => {
    mockExec.mockResolvedValue(makeState({ heartFiveBy: 0, spadeThreeBy: 1 }));
    renderWithProviders(<KaiserPage />);
    await waitFor(() => expect(screen.getByTestId('kaiser-heart-five-taken')).toBeInTheDocument());
    expect(screen.getByTestId('kaiser-spade-three-taken')).toBeInTheDocument();
  });

  // ベートは達成と意味がまるで違うので字面で分ける。
  it('tells a set hand apart from a made one', async () => {
    mockExec.mockResolvedValue(makeState({ phase: KaiserPhase.HAND_END, bidMade: false }));
    renderWithProviders(<KaiserPage />);
    await waitFor(() => expect(screen.getByTestId('kaiser-settlement')).toBeInTheDocument());
    expect(screen.getByTestId('kaiser-settlement')).toHaveTextContent('宣言額がそのままマイナス');

    mockExec.mockResolvedValue(makeState({ phase: KaiserPhase.HAND_END, bidMade: true }));
    renderWithProviders(<KaiserPage />);
    await waitFor(() => expect(screen.getAllByText(/両チームとも取った点を加算/).length).toBeGreaterThan(0));
  });

  it('advances to the next hand', async () => {
    mockExec.mockResolvedValue(makeState({ phase: KaiserPhase.HAND_END }));
    renderWithProviders(<KaiserPage />);
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
      mockExec.mockResolvedValue(makeState({ phase: KaiserPhase.GAME_END, gameEndFlag: true, winnerTeam: team }));
      renderWithProviders(<KaiserPage />);
      await waitFor(() => expect(screen.getAllByText(text).length).toBeGreaterThan(0));
    }
  });

  // **ヒント経路はページ側からも踏む。**ファクトリ単体テストだけだと
  // `hintFactories` の登録行と、ページのトグル／ツールチップが一度も
  // 実行されない（#4596 / #4600 のレビュー指摘）。
  it('turns the frontend hint on from the checkbox', async () => {
    localStorage.clear();
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<KaiserPage />);

    const toggle = await screen.findByRole('checkbox', { name: 'ヒント表示' });
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();

    fireEvent.click(toggle);
    expect(await screen.findByTestId('hint-tooltip')).toBeInTheDocument();
  });
});
