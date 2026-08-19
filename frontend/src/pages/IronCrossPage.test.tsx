import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ironcrossApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, IronCrossResponse } from '../types/card';
import { IronCrossPhase } from '../types/phases';
import { IronCrossPage } from './IronCrossPage';

vi.mock('../api/gameApi', () => ({
  ironcrossApi: { exec: vi.fn() },
  actionLogApi: { ironcross: vi.fn() },
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

const mockApi = vi.mocked(ironcrossApi.exec);
const mockUseCliMode = vi.mocked(useCliMode);

const card = (value: number): Card => ({ design: 'SPADE', value });
const hand = () => [card(1), card(2), card(3), card(4)];

const seat = (over: Partial<IronCrossResponse['seats'][number]> = {}) =>
  ({
    name: 'YOU',
    isHuman: true,
    chips: 1000,
    bet: 0,
    cards: hand(),
    folded: false,
    allIn: false,
    isTurn: true,
    line: 0,
    handRank: 0,
    bestHand: [],
    wonAmount: 0,
    ...over,
  }) as IronCrossResponse['seats'][number];

const base: IronCrossResponse = {
  phase: IronCrossPhase.BETTING,
  seats: [seat(), seat({ name: 'CPU1', isHuman: false, cards: [], isTurn: false })],
  cross: [null, null, null, null, null],
  revealedCount: 0,
  crossTotal: 5,
  verticalIndexes: [1, 0, 2],
  horizontalIndexes: [3, 0, 4],
  pot: 40,
  currentBet: 0,
  toCall: 0,
  raiseCount: 0,
  canRaise: true,
  turnSeat: 0,
  humanSeat: 0,
  isHumanTurn: true,
  isChoosing: false,
  handNumber: 1,
  remainingCards: 30,
  winnerSeat: 0,
  gameEndFlag: false,
  message: '',
};

const withState = (over: Partial<IronCrossResponse>): IronCrossResponse => ({ ...base, ...over });

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

describe('IronCrossPage', () => {
  it('マウント時に reset を呼ぶ', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<IronCrossPage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('自分の手札4枚を出す', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<IronCrossPage />);
    await waitFor(() => expect(screen.getByTestId('ic-hand')).toBeInTheDocument());
    expect(screen.getByTestId('ic-hand').children).toHaveLength(4);
  });

  // **十字の位置は落とさない。** 伏せている枠も出さないと、開いた札が
  // 中央なのか左なのか画面から読めなくなる。
  it('伏せている位置も5つの枠として出す', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<IronCrossPage />);
    await waitFor(() => expect(screen.getByTestId('ic-cross')).toBeInTheDocument());
    for (const i of [0, 1, 2, 3, 4]) {
      expect(screen.getByTestId(`ic-cross-${i}`)).toBeInTheDocument();
    }
  });

  // **開いた札は届いた位置に出る。** 詰めて描くと縦横の選択が成り立たない。
  it('開いた札を届いた位置に置く', async () => {
    // 中央 [0] と上 [1] だけが開いている盤面。
    mockApi.mockResolvedValue(withState({ cross: [card(13), card(12), null, null, null], revealedCount: 2 }));
    renderWithProviders(<IronCrossPage />);
    await waitFor(() => expect(screen.getByTestId('ic-cross-0')).toBeInTheDocument());
    // 開いている位置には札があり、伏せている位置には無い。
    expect(screen.getByTestId('ic-cross-0').querySelector('img,svg,div')).not.toBeNull();
    expect(screen.getByTestId('ic-cross-2')).toBeEmptyDOMElement();
    expect(screen.getByTestId('ic-cross-3')).toBeEmptyDOMElement();
    expect(screen.getByTestId('ic-cross-4')).toBeEmptyDOMElement();
  });

  it('公開枚数と総数を出す', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<IronCrossPage />);
    await waitFor(() => expect(screen.getByTestId('ic-revealed')).toHaveTextContent('0'));
    expect(screen.getByTestId('ic-revealed')).toHaveTextContent('5');
  });

  // **CPU の手札はサーバが送っていない。** 届いていなければ伏せ表示。
  it('届いていない手札は伏せ表示にする', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<IronCrossPage />);
    await waitFor(() => expect(screen.getByTestId('ic-seat-cards-1')).toBeInTheDocument());
    expect(screen.getByTestId('ic-seat-cards-1')).toHaveTextContent('伏せ');
  });

  // **届いていれば開く。** 一方向だけの検査は壊れているときに通る。
  it('ショーダウンで届いた手札を開く', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: IronCrossPhase.SHOWDOWN,
        seats: [seat({ isTurn: false }), seat({ name: 'CPU1', isHuman: false, isTurn: false })],
      }),
    );
    renderWithProviders(<IronCrossPage />);
    await waitFor(() => expect(screen.getByTestId('ic-seat-cards-1')).toBeInTheDocument());
    expect(screen.getByTestId('ic-seat-cards-1')).not.toHaveTextContent('伏せ');
    expect(screen.getByTestId('ic-seat-cards-1').children).toHaveLength(4);
  });

  // **選んだ列は届いたときだけ出す。** CPU の分はショーダウンまで 0 で届く。
  it('選んだ列は届いた席だけに出す', async () => {
    mockApi.mockResolvedValue(base);
    const { unmount } = renderWithProviders(<IronCrossPage />);
    await waitFor(() => expect(screen.getByTestId('ic-seat-0')).toBeInTheDocument());
    expect(screen.queryByTestId('ic-line-0')).not.toBeInTheDocument();
    expect(screen.queryByTestId('ic-line-1')).not.toBeInTheDocument();
    unmount();

    mockApi.mockResolvedValue(
      withState({
        phase: IronCrossPhase.SHOWDOWN,
        seats: [seat({ isTurn: false, line: 1 }), seat({ name: 'CPU1', isHuman: false, isTurn: false, line: 2 })],
      }),
    );
    renderWithProviders(<IronCrossPage />);
    await waitFor(() => expect(screen.getByTestId('ic-line-0')).toHaveTextContent('縦'));
    expect(screen.getByTestId('ic-line-1')).toHaveTextContent('横');
  });

  // **選ぶ場面はサーバが決める。** isChoosing に従い、ベットの手とは分ける。
  it('選ぶ場面でだけ縦横のボタンを出す', async () => {
    mockApi.mockResolvedValue(base);
    const { unmount } = renderWithProviders(<IronCrossPage />);
    await waitFor(() => expect(screen.getByTestId('ic-check')).toBeInTheDocument());
    expect(screen.queryByTestId('ic-vertical')).not.toBeInTheDocument();
    expect(screen.queryByTestId('ic-horizontal')).not.toBeInTheDocument();
    unmount();

    mockApi.mockResolvedValue(
      withState({
        phase: IronCrossPhase.CHOOSE_LINE,
        isChoosing: true,
        isHumanTurn: false,
        cross: [card(13), card(12), card(11), card(10), card(9)],
        revealedCount: 5,
      }),
    );
    renderWithProviders(<IronCrossPage />);
    await waitFor(() => expect(screen.getByTestId('ic-vertical')).toBeInTheDocument());
    expect(screen.getByTestId('ic-horizontal')).toBeInTheDocument();
    // ベットの手は出さない ── 一度きりの選択が打ち間違いで潰れる。
    expect(screen.queryByTestId('ic-check')).not.toBeInTheDocument();
    expect(screen.queryByTestId('ic-fold')).not.toBeInTheDocument();
  });

  // **後戻りできない一発勝負の選択** (#5781)。押す前にどの 3 枚が使われるかを
  // 十字の上で示す。位置はサーバの配列そのままで、ページで並べ直さない。
  it('列ボタンにホバー/フォーカスで使う3枚を示す', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: IronCrossPhase.CHOOSE_LINE,
        isChoosing: true,
        isHumanTurn: false,
        cross: [card(13), card(12), card(11), card(10), card(9)],
        revealedCount: 5,
        verticalIndexes: [1, 0, 2],
        horizontalIndexes: [3, 0, 4],
      }),
    );
    renderWithProviders(<IronCrossPage />);
    await waitFor(() => expect(screen.getByTestId('ic-vertical')).toBeInTheDocument());

    const previewed = () =>
      [0, 1, 2, 3, 4].filter((i) => screen.getByTestId(`ic-cross-${i}`).getAttribute('data-previewed') === 'true');
    expect(previewed()).toEqual([]);

    fireEvent.mouseEnter(screen.getByTestId('ic-vertical'));
    expect(previewed()).toEqual([0, 1, 2]);

    fireEvent.mouseLeave(screen.getByTestId('ic-vertical'));
    expect(previewed()).toEqual([]);

    fireEvent.focus(screen.getByTestId('ic-horizontal'));
    expect(previewed()).toEqual([0, 3, 4]);

    fireEvent.blur(screen.getByTestId('ic-horizontal'));
    expect(previewed()).toEqual([]);
  });

  // 4 つのハンドラすべてが同じ状態を触る。**片方の列だけ配線を忘れる**のが
  // ありがちな取りこぼしなので、縦=フォーカス / 横=ホバーの側も見る。
  it('縦はフォーカスでも、横はホバーでも光る', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: IronCrossPhase.CHOOSE_LINE,
        isChoosing: true,
        isHumanTurn: false,
        // **伏せたままの位置でも光る。** 開いた札だけ光ると、まだ見えていない
        // 枝がどちらの列に属すのか読めない。
        cross: [card(13), null, card(11), null, card(9)],
        revealedCount: 3,
      }),
    );
    renderWithProviders(<IronCrossPage />);
    await waitFor(() => expect(screen.getByTestId('ic-vertical')).toBeInTheDocument());

    fireEvent.focus(screen.getByTestId('ic-vertical'));
    expect(screen.getByTestId('ic-cross-1')).toHaveAttribute('data-previewed', 'true');
    fireEvent.blur(screen.getByTestId('ic-vertical'));
    expect(screen.getByTestId('ic-cross-1')).not.toHaveAttribute('data-previewed');

    fireEvent.mouseEnter(screen.getByTestId('ic-horizontal'));
    expect(screen.getByTestId('ic-cross-3')).toHaveAttribute('data-previewed', 'true');
    fireEvent.mouseLeave(screen.getByTestId('ic-horizontal'));
    expect(screen.getByTestId('ic-cross-3')).not.toHaveAttribute('data-previewed');
  });

  // 選択フェーズが終わったらプレビューは畳む（レビュー指摘）。クリックで
  // ボタンが消えるとき mouseleave も blur も飛ばない環境がある。
  it('列を選び終えたらハイライトを畳む', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: IronCrossPhase.CHOOSE_LINE,
        isChoosing: true,
        isHumanTurn: false,
        cross: [card(13), card(12), card(11), card(10), card(9)],
        revealedCount: 5,
      }),
    );
    renderWithProviders(<IronCrossPage />);
    await waitFor(() => expect(screen.getByTestId('ic-vertical')).toBeInTheDocument());

    fireEvent.mouseEnter(screen.getByTestId('ic-vertical'));
    expect(screen.getByTestId('ic-cross-1')).toHaveAttribute('data-previewed', 'true');

    // 次の応答で選択フェーズが終わる。
    mockApi.mockResolvedValue(
      withState({
        phase: IronCrossPhase.SHOWDOWN,
        isChoosing: false,
        isHumanTurn: false,
        cross: [card(13), card(12), card(11), card(10), card(9)],
        revealedCount: 5,
      }),
    );
    fireEvent.click(screen.getByTestId('ic-vertical'));

    await waitFor(() => expect(screen.queryByTestId('ic-vertical')).not.toBeInTheDocument());
    expect(screen.getByTestId('ic-cross-1')).not.toHaveAttribute('data-previewed');
  });

  // **配列はサーバのものを使う。** 位置をページで決め直さない。
  it('サーバが別の位置を送ってきたらそちらを光らせる', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: IronCrossPhase.CHOOSE_LINE,
        isChoosing: true,
        isHumanTurn: false,
        cross: [card(13), card(12), card(11), card(10), card(9)],
        revealedCount: 5,
        verticalIndexes: [3, 4],
        horizontalIndexes: [1, 2],
      }),
    );
    renderWithProviders(<IronCrossPage />);
    await waitFor(() => expect(screen.getByTestId('ic-vertical')).toBeInTheDocument());

    fireEvent.mouseEnter(screen.getByTestId('ic-vertical'));
    expect(screen.getByTestId('ic-cross-3')).toHaveAttribute('data-previewed', 'true');
    expect(screen.getByTestId('ic-cross-1')).not.toHaveAttribute('data-previewed');
  });

  // **選んだ列がそのまま送られる。** 強いほうに直したりしない。
  it('押した列をそのまま送る', async () => {
    mockApi.mockResolvedValue(withState({ phase: IronCrossPhase.CHOOSE_LINE, isChoosing: true, isHumanTurn: false }));
    renderWithProviders(<IronCrossPage />);
    await waitFor(() => expect(screen.getByTestId('ic-vertical')).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByTestId('ic-vertical'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('vertical'));

    mockApi.mockClear();
    fireEvent.click(screen.getByTestId('ic-horizontal'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('horizontal'));
  });

  // **チェックとコールは場況で入れ替わる。** サーバの toCall に従う。
  it('賭けが無ければチェック、あればコールを出す', async () => {
    mockApi.mockResolvedValue(base);
    const { unmount } = renderWithProviders(<IronCrossPage />);
    await waitFor(() => expect(screen.getByTestId('ic-check')).toBeInTheDocument());
    expect(screen.queryByTestId('ic-call')).not.toBeInTheDocument();
    expect(screen.getByTestId('ic-bet')).toBeInTheDocument();
    unmount();

    mockApi.mockResolvedValue(withState({ toCall: 20, currentBet: 20 }));
    renderWithProviders(<IronCrossPage />);
    await waitFor(() => expect(screen.getByTestId('ic-call')).toBeInTheDocument());
    expect(screen.queryByTestId('ic-check')).not.toBeInTheDocument();
    expect(screen.getByTestId('ic-raise')).toBeInTheDocument();
  });

  // **レイズの可否はサーバが決める。** 上限に達したら出さない。
  it('レイズ上限に達したらレイズを出さない', async () => {
    mockApi.mockResolvedValue(withState({ toCall: 20, currentBet: 20, canRaise: false }));
    renderWithProviders(<IronCrossPage />);
    await waitFor(() => expect(screen.getByTestId('ic-call')).toBeInTheDocument());
    expect(screen.queryByTestId('ic-raise')).not.toBeInTheDocument();
  });

  it('各操作をそのまま送る', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<IronCrossPage />);
    await waitFor(() => expect(screen.getByTestId('ic-check')).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByTestId('ic-check'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('check'));

    mockApi.mockClear();
    fireEvent.click(screen.getByTestId('ic-fold'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('fold'));

    mockApi.mockClear();
    fireEvent.click(screen.getByTestId('ic-bet'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', { amount: 20 }));
  });

  it('他人の手番では操作ボタンを出さない', async () => {
    mockApi.mockResolvedValue(withState({ isHumanTurn: false, turnSeat: 1 }));
    renderWithProviders(<IronCrossPage />);
    await waitFor(() => expect(screen.getByTestId('ic-hand')).toBeInTheDocument());
    expect(screen.queryByTestId('ic-check')).not.toBeInTheDocument();
    expect(screen.queryByTestId('ic-fold')).not.toBeInTheDocument();
  });

  it('ショーダウンで獲得額と次のハンドを出す', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: IronCrossPhase.SHOWDOWN,
        isHumanTurn: false,
        seats: [seat({ isTurn: false, wonAmount: 80 }), seat({ name: 'CPU1', isHuman: false, isTurn: false })],
      }),
    );
    renderWithProviders(<IronCrossPage />);
    await waitFor(() => expect(screen.getByTestId('ic-won-0')).toHaveTextContent('80'));

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のハンドへ' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('next'));
  });

  it('終局で勝者を出し、次のハンドを出さない', async () => {
    mockApi.mockResolvedValue(
      withState({ phase: IronCrossPhase.GAME_END, gameEndFlag: true, isHumanTurn: false, winnerSeat: 1 }),
    );
    renderWithProviders(<IronCrossPage />);
    await waitFor(() => expect(screen.getByTestId('ic-winner')).toHaveTextContent('CPU1'));
    expect(screen.queryByRole('button', { name: '次のハンドへ' })).not.toBeInTheDocument();
  });

  it('チップ・ハンド数・ポットを出す', async () => {
    mockApi.mockResolvedValue(
      withState({
        handNumber: 3,
        pot: 120,
        seats: [seat({ chips: 870 }), seat({ name: 'CPU1', isHuman: false, cards: [] })],
      }),
    );
    renderWithProviders(<IronCrossPage />);
    await waitFor(() => expect(screen.getByTestId('ic-chips')).toHaveTextContent('870'));
    expect(screen.getByTestId('ic-hand-line')).toHaveTextContent('3');
    expect(screen.getByTestId('ic-hand-line')).toHaveTextContent('120');
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
    renderWithProviders(<IronCrossPage />);
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
    expect(screen.queryByTestId('ic-check')).not.toBeInTheDocument();
  });
});
