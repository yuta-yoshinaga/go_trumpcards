import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { baseballpokerApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BaseballPokerResponse, Card } from '../types/card';
import { BaseballPhase } from '../types/phases';
import { BaseballPokerPage } from './BaseballPokerPage';

vi.mock('../api/gameApi', () => ({
  baseballpokerApi: { exec: vi.fn() },
  actionLogApi: { baseballpoker: vi.fn() },
}));

vi.mock('../hooks/useCliMode', () => ({
  useCliMode: vi.fn(() => ({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  })),
}));

const mockApi = vi.mocked(baseballpokerApi.exec);
const mockUseCliMode = vi.mocked(useCliMode);

const card = (value: number): Card => ({ design: 'SPADE', value });

const seat = (over: Partial<BaseballPokerResponse['seats'][number]> = {}) =>
  ({
    name: 'YOU',
    isHuman: true,
    chips: 1000,
    bet: 0,
    cards: [card(1), card(2), card(5)],
    faceUp: [false, false, true],
    bonusCards: 0,
    folded: false,
    allIn: false,
    isTurn: true,
    isBuying: false,
    handRank: 0,
    usedWild: false,
    bestHand: [],
    wonAmount: 0,
    ...over,
  }) as BaseballPokerResponse['seats'][number];

// CPU 席: 伏せ 2 枚は null で届き、表札だけが実物で届く。
const cpuSeat = (over: Partial<BaseballPokerResponse['seats'][number]> = {}) =>
  seat({
    name: 'CPU1',
    isHuman: false,
    isTurn: false,
    cards: [null, null, card(7)],
    faceUp: [false, false, true],
    ...over,
  });

const base: BaseballPokerResponse = {
  phase: BaseballPhase.BETTING,
  seats: [seat(), cpuSeat()],
  street: 1,
  streetTotal: 4,
  wildValues: [3, 9],
  bonusValue: 4,
  buyInValue: 3,
  pot: 40,
  currentBet: 0,
  toCall: 0,
  raiseCount: 0,
  canRaise: true,
  turnSeat: 0,
  humanSeat: 0,
  isHumanTurn: true,
  buyerSeat: -1,
  buyCost: 0,
  isBuying: false,
  handNumber: 1,
  remainingCards: 30,
  winnerSeat: 0,
  gameEndFlag: false,
  message: '',
};

const withState = (over: Partial<BaseballPokerResponse>): BaseballPokerResponse => ({ ...base, ...over });

beforeEach(() => {
  vi.clearAllMocks();
  mockUseCliMode.mockReturnValue({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  } as unknown as ReturnType<typeof useCliMode>);
});

describe('BaseballPokerPage', () => {
  it('マウント時に reset を呼ぶ', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<BaseballPokerPage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('自分の手札を出す', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<BaseballPokerPage />);
    await waitFor(() => expect(screen.getByTestId('bb-hand')).toBeInTheDocument());
    expect(screen.getByTestId('bb-hand').children).toHaveLength(3);
  });

  // **表札は相手のぶんも出す。それがスタッドの読み合いの材料。**
  it('CPU の表札を出し、伏せ札だけを裏向きにする', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<BaseballPokerPage />);
    await waitFor(() => expect(screen.getByTestId('bb-seat-cards-1')).toBeInTheDocument());
    // 3 枚ぶんの枠があり、そのうち 2 枚が裏向き。
    expect(screen.getByTestId('bb-seat-cards-1').children).toHaveLength(3);
    expect(screen.getByTestId('bb-seat-cards-1-hidden-0')).toBeInTheDocument();
    expect(screen.getByTestId('bb-seat-cards-1-hidden-1')).toBeInTheDocument();
    expect(screen.queryByTestId('bb-seat-cards-1-hidden-2')).not.toBeInTheDocument();
  });

  // **ショーダウンでは届いた札を全部出す。** 一方向だけの検査は壊れても通る。
  it('ショーダウンで届いた伏せ札を開く', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: BaseballPhase.SHOWDOWN,
        isHumanTurn: false,
        seats: [seat({ isTurn: false }), cpuSeat({ cards: [card(11), card(12), card(7)] })],
      }),
    );
    renderWithProviders(<BaseballPokerPage />);
    await waitFor(() => expect(screen.getByTestId('bb-seat-cards-1')).toBeInTheDocument());
    expect(screen.queryByTestId('bb-seat-cards-1-hidden-0')).not.toBeInTheDocument();
    expect(screen.getByTestId('bb-seat-cards-1').children).toHaveLength(3);
  });

  // **説明文も同じ値から作る** (#5782)。印だけ動いて文言が固定だと、規則が
  // 変わったとき片方だけ嘘になる。
  it('ワイルドの説明文もサーバの値から作る', async () => {
    mockApi.mockResolvedValue(withState({ wildValues: [2, 7], bonusValue: 5, buyInValue: 6 }));
    renderWithProviders(<BaseballPokerPage />);

    const notice = await screen.findByTestId('bb-wild-notice');
    expect(notice).toHaveTextContent('2 / 7');
    expect(notice).toHaveTextContent('5');
    expect(notice).toHaveTextContent('6');
    expect(notice.textContent).not.toContain('3 と 9');
  });

  // **ワイルドの印はサーバが送った値から出す。** ページが 3 と 9 を持たない。
  it('サーバが送ったワイルドの値に印を付ける', async () => {
    mockApi.mockResolvedValue(
      withState({ seats: [seat({ cards: [card(3), card(2), card(9)], faceUp: [false, false, true] }), cpuSeat()] }),
    );
    const { unmount } = renderWithProviders(<BaseballPokerPage />);
    await waitFor(() => expect(screen.getByTestId('bb-hand-wild-0')).toBeInTheDocument());
    expect(screen.queryByTestId('bb-hand-wild-1')).not.toBeInTheDocument();
    expect(screen.getByTestId('bb-hand-wild-2')).toBeInTheDocument();
    unmount();

    // ワイルドの値が変われば印の位置も変わる ── 画面が値を持っていない証拠。
    mockApi.mockResolvedValue(
      withState({
        wildValues: [2],
        seats: [seat({ cards: [card(3), card(2), card(9)], faceUp: [false, false, true] }), cpuSeat()],
      }),
    );
    renderWithProviders(<BaseballPokerPage />);
    await waitFor(() => expect(screen.getByTestId('bb-hand-wild-1')).toBeInTheDocument());
    expect(screen.queryByTestId('bb-hand-wild-0')).not.toBeInTheDocument();
    expect(screen.queryByTestId('bb-hand-wild-2')).not.toBeInTheDocument();
  });

  it('ストリートと総数を出す', async () => {
    mockApi.mockResolvedValue(withState({ street: 3 }));
    renderWithProviders(<BaseballPokerPage />);
    await waitFor(() => expect(screen.getByTestId('bb-street')).toHaveTextContent('3'));
    expect(screen.getByTestId('bb-street')).toHaveTextContent('4');
  });

  it('追加札の枚数を出す', async () => {
    mockApi.mockResolvedValue(withState({ seats: [seat({ bonusCards: 2 }), cpuSeat()] }));
    renderWithProviders(<BaseballPokerPage />);
    await waitFor(() => expect(screen.getByTestId('bb-bonus-0')).toHaveTextContent('2'));
  });

  // **買い増しの返事はベットの手と分ける。** 同じ列だと打ち間違いで払ってしまう。
  it('買い増しの場面でだけ支払いのボタンを出す', async () => {
    mockApi.mockResolvedValue(base);
    const { unmount } = renderWithProviders(<BaseballPokerPage />);
    await waitFor(() => expect(screen.getByTestId('bb-check')).toBeInTheDocument());
    expect(screen.queryByTestId('bb-pay')).not.toBeInTheDocument();
    unmount();

    mockApi.mockResolvedValue(
      withState({
        phase: BaseballPhase.BUY_IN,
        isBuying: true,
        isHumanTurn: false,
        buyerSeat: 0,
        buyCost: 80,
        seats: [seat({ isTurn: false, isBuying: true }), cpuSeat()],
      }),
    );
    renderWithProviders(<BaseballPokerPage />);
    await waitFor(() => expect(screen.getByTestId('bb-pay')).toBeInTheDocument());
    expect(screen.getByTestId('bb-buyfold')).toBeInTheDocument();
    // 払う額を出す。
    expect(screen.getByTestId('bb-buy-guide')).toHaveTextContent('80');
    // ベットの手は出さない。
    expect(screen.queryByTestId('bb-check')).not.toBeInTheDocument();
    expect(screen.queryByTestId('bb-fold')).not.toBeInTheDocument();
  });

  // **降りるつもりが支払いに化けない。** 両方の返事を別々に送る。
  it('買い増しの返事をそのまま送る', async () => {
    mockApi.mockResolvedValue(
      withState({ phase: BaseballPhase.BUY_IN, isBuying: true, isHumanTurn: false, buyerSeat: 0, buyCost: 80 }),
    );
    renderWithProviders(<BaseballPokerPage />);
    await waitFor(() => expect(screen.getByTestId('bb-pay')).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByTestId('bb-buyfold'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('buyfold'));
    expect(mockApi).not.toHaveBeenCalledWith('pay');

    mockApi.mockClear();
    fireEvent.click(screen.getByTestId('bb-pay'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('pay'));
  });

  // **チェックとコールは場況で入れ替わる。** サーバの toCall に従う。
  it('賭けが無ければチェック、あればコールを出す', async () => {
    mockApi.mockResolvedValue(base);
    const { unmount } = renderWithProviders(<BaseballPokerPage />);
    await waitFor(() => expect(screen.getByTestId('bb-check')).toBeInTheDocument());
    expect(screen.queryByTestId('bb-call')).not.toBeInTheDocument();
    expect(screen.getByTestId('bb-bet')).toBeInTheDocument();
    unmount();

    mockApi.mockResolvedValue(withState({ toCall: 20, currentBet: 20 }));
    renderWithProviders(<BaseballPokerPage />);
    await waitFor(() => expect(screen.getByTestId('bb-call')).toBeInTheDocument());
    expect(screen.queryByTestId('bb-check')).not.toBeInTheDocument();
    expect(screen.getByTestId('bb-raise')).toBeInTheDocument();
  });

  it('レイズ上限に達したらレイズを出さない', async () => {
    mockApi.mockResolvedValue(withState({ toCall: 20, currentBet: 20, canRaise: false }));
    renderWithProviders(<BaseballPokerPage />);
    await waitFor(() => expect(screen.getByTestId('bb-call')).toBeInTheDocument());
    expect(screen.queryByTestId('bb-raise')).not.toBeInTheDocument();
  });

  it('各操作をそのまま送る', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<BaseballPokerPage />);
    await waitFor(() => expect(screen.getByTestId('bb-check')).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByTestId('bb-check'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('check'));

    mockApi.mockClear();
    fireEvent.click(screen.getByTestId('bb-fold'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('fold'));

    mockApi.mockClear();
    fireEvent.click(screen.getByTestId('bb-bet'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', { amount: 20 }));
  });

  it('他人の手番では操作ボタンを出さない', async () => {
    mockApi.mockResolvedValue(withState({ isHumanTurn: false, turnSeat: 1 }));
    renderWithProviders(<BaseballPokerPage />);
    await waitFor(() => expect(screen.getByTestId('bb-hand')).toBeInTheDocument());
    expect(screen.queryByTestId('bb-check')).not.toBeInTheDocument();
    expect(screen.queryByTestId('bb-pay')).not.toBeInTheDocument();
  });

  it('ショーダウンで獲得額とワイルド使用を出し、次のハンドへ進める', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: BaseballPhase.SHOWDOWN,
        isHumanTurn: false,
        seats: [seat({ isTurn: false, wonAmount: 80, usedWild: true }), cpuSeat()],
      }),
    );
    renderWithProviders(<BaseballPokerPage />);
    await waitFor(() => expect(screen.getByTestId('bb-won-0')).toHaveTextContent('80'));
    expect(screen.getByTestId('bb-usedwild-0')).toBeInTheDocument();

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のハンドへ' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('next'));
  });

  it('終局で勝者を出し、次のハンドを出さない', async () => {
    mockApi.mockResolvedValue(
      withState({ phase: BaseballPhase.GAME_END, gameEndFlag: true, isHumanTurn: false, winnerSeat: 1 }),
    );
    renderWithProviders(<BaseballPokerPage />);
    await waitFor(() => expect(screen.getByTestId('bb-winner')).toHaveTextContent('CPU1'));
    expect(screen.queryByRole('button', { name: '次のハンドへ' })).not.toBeInTheDocument();
  });

  it('チップ・ハンド数・ポットを出す', async () => {
    mockApi.mockResolvedValue(withState({ handNumber: 3, pot: 120, seats: [seat({ chips: 870 }), cpuSeat()] }));
    renderWithProviders(<BaseballPokerPage />);
    await waitFor(() => expect(screen.getByTestId('bb-chips')).toHaveTextContent('870'));
    expect(screen.getByTestId('bb-hand-line')).toHaveTextContent('3');
    expect(screen.getByTestId('bb-hand-line')).toHaveTextContent('120');
  });

  it('CLIモードでは端末を出す', async () => {
    mockUseCliMode.mockReturnValue({
      cliEnabled: true,
      toggleCli: vi.fn(),
      logEntries: [],
      addInput: vi.fn(),
      addOutput: vi.fn(),
      addError: vi.fn(),
      clearLog: vi.fn(),
    } as unknown as ReturnType<typeof useCliMode>);
    mockApi.mockResolvedValue(base);
    renderWithProviders(<BaseballPokerPage />);
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
    expect(screen.queryByTestId('bb-check')).not.toBeInTheDocument();
  });
});
