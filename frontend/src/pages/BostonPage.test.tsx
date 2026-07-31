import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { bostonApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BostonBidOption, BostonPlayer, BostonResponse, CardDesign } from '../types/card';
import { BostonPhase } from '../types/phases';
import { BostonPage } from './BostonPage';

vi.mock('../api/gameApi', () => ({
  bostonApi: { exec: vi.fn() },
  actionLogApi: { boston: vi.fn() },
}));

const mockExec = vi.mocked(bostonApi.exec);

const card = (design: CardDesign, value: number) => ({ design, value });

function option(
  level: number,
  name: string,
  kind: number,
  tricks: number,
  overrides?: Partial<BostonBidOption>,
): BostonBidOption {
  return {
    level,
    name,
    kind,
    tricks,
    needsTrump: kind === 1,
    exposed: false,
    canCallPartner: kind === 1,
    payout: level,
    ...overrides,
  };
}

/** The ladder in rank order — the miseres sit BETWEEN the trick bids. */
const LADDER: BostonBidOption[] = [
  option(1, 'five', 1, 5),
  option(2, 'six', 1, 6),
  option(3, 'littleMisere', 2, 0),
  option(4, 'seven', 1, 7),
  option(5, 'piccolissimo', 3, 1),
  option(9, 'littleMisereTable', 2, 0, { exposed: true }),
];

function seat(id: number, isHuman: boolean, overrides?: Partial<BostonPlayer>): BostonPlayer {
  return {
    id,
    isHuman,
    cardCount: 3,
    cards: isHuman ? [card('SPADE', 1), card('HEART', 2), card('CLOVER', 3)] : [],
    tricksWon: 0,
    chips: 0,
    isDealer: id === 3,
    isDeclarer: false,
    isPartner: false,
    isDeclarerSide: false,
    isCurrentTurn: id === 0,
    ...overrides,
  };
}

function makeState(overrides?: Partial<BostonResponse>): BostonResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false)],
    phase: BostonPhase.PLAY,
    handNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 3,
    bids: [],
    highBid: { player: 0, level: 4, name: 'seven', suit: 3 },
    bidOptions: LADDER,
    declarerIdx: 0,
    partnerIdx: -1,
    trumpSuit: 3,
    exposed: false,
    trick: [],
    validPlays: [0, 1, 2],
    trickLeaderIdx: 0,
    trickNumber: 0,
    declarerTricks: 4,
    bidMade: false,
    handSize: 13,
    targetHands: 8,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    ...overrides,
  };
}

/** The hand-card buttons, in order. */
function handButtons() {
  return screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'play');
}

describe('BostonPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<BostonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **序列はサーバーが送る順のまま出す。**ミゼールがトリック宣言の間に挟まる。
  it('shows the ladder in rank order with the miseres interleaved', async () => {
    renderWithProviders(<BostonPage />);
    await waitFor(() => expect(screen.getByTestId('boston-ladder')).toBeInTheDocument());
    const ladder = screen.getByTestId('boston-ladder');
    const text = ladder.textContent ?? '';

    expect(text).toContain('6トリック');
    expect(text).toContain('リトル・ミゼール');
    expect(text).toContain('7トリック');
    // 並びは 6トリック < リトル・ミゼール < 7トリック。
    expect(text.indexOf('2. 6トリック')).toBeLessThan(text.indexOf('3. リトル・ミゼール'));
    expect(text.indexOf('3. リトル・ミゼール')).toBeLessThan(text.indexOf('4. 7トリック'));
    // 注意書きも出す。
    expect(ladder).toHaveTextContent('ミゼールはトリック宣言の間に挟まります');
  });

  // **ピッコリッシモはちょうど1トリック。**勝利条件が第3の型であることを書く。
  it('spells out what each kind of bid asks for', async () => {
    renderWithProviders(<BostonPage />);
    await waitFor(() => expect(screen.getByTestId('boston-ladder')).toBeInTheDocument());
    const ladder = screen.getByTestId('boston-ladder');
    expect(ladder).toHaveTextContent('7トリック以上');
    expect(ladder).toHaveTextContent('1トリックも取らない');
    expect(ladder).toHaveTextContent('ちょうど1トリック');
    // 公開宣言とパートナー可否も出す。
    expect(ladder).toHaveTextContent('公開');
    expect(ladder).toHaveTextContent('味方を呼べる');
  });

  it('plays exactly one card', async () => {
    renderWithProviders(<BostonPage />);
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
    renderWithProviders(<BostonPage />);
    await waitFor(() => expect(screen.getByTestId('boston-play-notice')).toBeInTheDocument());

    const hand = handButtons();
    expect(hand[0]).toBeDisabled();
    expect(hand[1]).toBeDisabled();
    expect(hand[2]).toBeEnabled();
  });

  // **立っている宣言より上しか選べない。**
  it('offers only the steps that beat the standing bid', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: BostonPhase.BID,
        highBid: { player: 1, level: 3, name: 'littleMisere', suit: 0 },
        declarerIdx: -1,
      }),
    );
    renderWithProviders(<BostonPage />);
    await waitFor(() => expect(screen.getByLabelText(/契約/)).toBeInTheDocument());

    const select = screen.getByLabelText(/契約/) as HTMLSelectElement;
    const values = Array.from(select.options).map((o) => o.value);
    expect(values).not.toContain('2');
    expect(values).toContain('4');
    expect(values).toContain('5');
  });

  // **切札が要るのはトリック宣言だけ。**ミゼールを選んだらスート欄を出さない。
  it('asks for a suit only on a trick bid', async () => {
    mockExec.mockResolvedValue(makeState({ phase: BostonPhase.BID, highBid: null, declarerIdx: -1, trumpSuit: 0 }));
    renderWithProviders(<BostonPage />);
    await waitFor(() => expect(screen.getByLabelText(/契約/)).toBeInTheDocument());

    // ミゼール（段 3）を選ぶとスート欄は出ない。
    fireEvent.change(screen.getByLabelText(/契約/), { target: { value: '3' } });
    expect(screen.queryByLabelText(/切札/)).not.toBeInTheDocument();

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '宣言する' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { level: 3, suit: undefined }));

    // トリック宣言（段 4）ならスート欄が出る。
    fireEvent.change(screen.getByLabelText(/契約/), { target: { value: '4' } });
    await waitFor(() => expect(screen.getByLabelText(/切札/)).toBeInTheDocument());
  });

  it('passes', async () => {
    mockExec.mockResolvedValue(makeState({ phase: BostonPhase.BID, highBid: null, declarerIdx: -1 }));
    renderWithProviders(<BostonPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'パス' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
  });

  // **「各自個人戦」ではない。**トリック宣言なら 2 対 2 にできる。
  it('offers both a partner and going alone', async () => {
    mockExec.mockResolvedValue(makeState({ phase: BostonPhase.CALL_PARTNER, declarerIdx: 0 }));
    renderWithProviders(<BostonPage />);
    await waitFor(() => expect(screen.getByTestId('boston-partner-notice')).toBeInTheDocument());
    expect(screen.getByTestId('boston-partner-notice')).toHaveTextContent('2対2');

    // 自分は指名できない。
    expect(screen.queryByRole('button', { name: /あなた を味方に/ })).not.toBeInTheDocument();

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: /CPU 2 を味方に/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('callpartner', { partner: 2 }));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '単独で戦う' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('callpartner', { partner: -1 }));
  });

  it('marks the declarer and the partner', async () => {
    mockExec.mockResolvedValue(
      makeState({
        partnerIdx: 2,
        players: [
          seat(0, true, { isDeclarer: true, isDeclarerSide: true }),
          seat(1, false),
          seat(2, false, { isPartner: true, isDeclarerSide: true }),
          seat(3, false),
        ],
      }),
    );
    renderWithProviders(<BostonPage />);
    await waitFor(() => expect(screen.getAllByTestId('boston-player')).toHaveLength(4));
    const players = screen.getAllByTestId('boston-player');
    expect(players[0]).toHaveTextContent('宣言');
    expect(players[2]).toHaveTextContent('味方');
  });

  // 達成と失敗は取ったトリック数つきで区別する。
  it('tells a failed contract apart from a made one', async () => {
    mockExec.mockResolvedValue(makeState({ phase: BostonPhase.HAND_END, bidMade: false, declarerTricks: 5 }));
    renderWithProviders(<BostonPage />);
    await waitFor(() => expect(screen.getByTestId('boston-settlement')).toBeInTheDocument());
    expect(screen.getByTestId('boston-settlement')).toHaveTextContent('各相手に払います');

    mockExec.mockResolvedValue(makeState({ phase: BostonPhase.HAND_END, bidMade: true, declarerTricks: 8 }));
    renderWithProviders(<BostonPage />);
    await waitFor(() => expect(screen.getAllByText(/各相手から受け取ります/).length).toBeGreaterThan(0));
  });

  it('advances to the next hand', async () => {
    mockExec.mockResolvedValue(makeState({ phase: BostonPhase.HAND_END }));
    renderWithProviders(<BostonPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次の局へ' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次の局へ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('reports each outcome', async () => {
    for (const [winner, text] of [
      [0, /あなたの勝利です！/],
      [2, /CPU 2 の勝利です。/],
    ] as const) {
      mockExec.mockResolvedValue(makeState({ phase: BostonPhase.GAME_END, gameEndFlag: true, winnerIdx: winner }));
      renderWithProviders(<BostonPage />);
      await waitFor(() => expect(screen.getAllByText(text).length).toBeGreaterThan(0));
    }
  });
});
