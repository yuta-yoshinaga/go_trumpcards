import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tusacApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, TuSacResponse } from '../types/card';
import { TuSacPhase } from '../types/phases';
import { TuSacPage } from './TuSacPage';

vi.mock('../api/gameApi', () => ({
  tusacApi: { exec: vi.fn() },
  actionLogApi: { tusac: vi.fn() },
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

const mockApi = vi.mocked(tusacApi.exec);
const mockUseCliMode = vi.mocked(useCliMode);

// 四色牌の札は手続き描画で届く (design は共有の文字列、glyph が駒)。
const card = (glyph: string, color: string, value: number): Card =>
  ({ design: 'SPADE', value, glyph, label: glyph, color, deck: 'tusac' }) as Card;

const hand = () => [card('卒', 'red', 7), card('車', 'green', 4), card('馬', 'gold', 5), card('砲', 'black', 6)];

const seat = (over: Partial<TuSacResponse['seats'][number]> = {}) =>
  ({
    name: 'YOU',
    isHuman: true,
    cards: hand(),
    handCount: 4,
    melds: [],
    meldPoints: 0,
    score: 0,
    roundScore: 0,
    isTurn: true,
    wentOut: false,
    ...over,
  }) as TuSacResponse['seats'][number];

const base: TuSacResponse = {
  phase: TuSacPhase.DRAW,
  seats: [seat(), seat({ name: 'CPU1', isHuman: false, cards: [], handCount: 20, isTurn: false })],
  discardTop: card('象', 'red', 3),
  discardCount: 1,
  stockCount: 31,
  turnSeat: 0,
  humanSeat: 0,
  isHumanTurn: true,
  roundNumber: 1,
  rounds: 5,
  wentOutSeat: -1,
  handSize: 20,
  deckSize: 112,
  meldPointsByKind: [0, 2, 3, 5],
  winnerSeat: 0,
  gameEndFlag: false,
  message: '',
};

const withState = (over: Partial<TuSacResponse>): TuSacResponse => ({ ...base, ...over });

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

describe('TuSacPage', () => {
  it('マウント時に reset を呼ぶ', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<TuSacPage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('自分の手札を出す', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<TuSacPage />);
    await waitFor(() => expect(screen.getByTestId('tusac-hand')).toBeInTheDocument());
    expect(screen.getByTestId('tusac-hand').children).toHaveLength(4);
  });

  // **相手の手札は届かない。** 枚数だけが分かる。
  it('相手は枚数だけを出す', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<TuSacPage />);
    await waitFor(() => expect(screen.getByTestId('tusac-count-1')).toBeInTheDocument());
    expect(screen.getByTestId('tusac-count-1')).toHaveTextContent('20');
    // 相手の札そのものは描かれない。
    expect(screen.queryByTestId('tusac-card-4')).not.toBeInTheDocument();
  });

  // **場に出た組み合わせは全員ぶん見える。** そこから読むのがこのゲーム。
  it('場の組み合わせを全員ぶん出す', async () => {
    mockApi.mockResolvedValue(
      withState({
        seats: [
          seat(),
          seat({
            name: 'CPU1',
            isHuman: false,
            cards: [],
            handCount: 15,
            isTurn: false,
            melds: [{ kind: 3, points: 5, cards: [card('卒', 'red', 7), card('卒', 'gold', 7)] }],
            meldPoints: 5,
          }),
        ],
      }),
    );
    renderWithProviders(<TuSacPage />);
    await waitFor(() => expect(screen.getByTestId('tusac-melds-1')).toBeInTheDocument());
    expect(screen.getByTestId('tusac-melds-1')).toHaveTextContent('卒5枚');
    expect(screen.getByTestId('tusac-melds-1')).toHaveTextContent('5');
    // 自分はまだ出していない。
    expect(screen.getByTestId('tusac-melds-0')).toHaveTextContent('まだ出していません');
  });

  // **引く場面と出す場面でボタンが入れ替わる。** サーバのフェーズに従う。
  it('引く場面と捨てる場面で操作を切り替える', async () => {
    mockApi.mockResolvedValue(base);
    const { unmount } = renderWithProviders(<TuSacPage />);
    await waitFor(() => expect(screen.getByTestId('tusac-draw')).toBeInTheDocument());
    expect(screen.getByTestId('tusac-take')).toBeInTheDocument();
    expect(screen.queryByTestId('tusac-meld')).not.toBeInTheDocument();
    expect(screen.queryByTestId('tusac-discard')).not.toBeInTheDocument();
    unmount();

    mockApi.mockResolvedValue(withState({ phase: TuSacPhase.DISCARD }));
    renderWithProviders(<TuSacPage />);
    await waitFor(() => expect(screen.getByTestId('tusac-meld')).toBeInTheDocument());
    expect(screen.getByTestId('tusac-discard')).toBeInTheDocument();
    expect(screen.queryByTestId('tusac-draw')).not.toBeInTheDocument();
  });

  // **山と捨て札は別のコマンド。** 取り違えると狙った札が流れる。
  it('山と捨て札を別のコマンドで送る', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<TuSacPage />);
    await waitFor(() => expect(screen.getByTestId('tusac-draw')).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByTestId('tusac-draw'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('draw'));
    expect(mockApi).not.toHaveBeenCalledWith('take');

    mockApi.mockClear();
    fireEvent.click(screen.getByTestId('tusac-take'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('take'));
    expect(mockApi).not.toHaveBeenCalledWith('draw');
  });

  it('捨て札が無ければ拾えない', async () => {
    mockApi.mockResolvedValue(withState({ discardTop: null, discardCount: 0 }));
    renderWithProviders(<TuSacPage />);
    await waitFor(() => expect(screen.getByTestId('tusac-take')).toBeDisabled());
    expect(screen.getByTestId('tusac-discard-top')).toHaveTextContent('捨て札なし');
  });

  // **選ぶのは手札の位置。** 押した番号がそのままワイヤに乗る。
  it('選んだ位置をそのまま送る', async () => {
    mockApi.mockResolvedValue(withState({ phase: TuSacPhase.DISCARD }));
    renderWithProviders(<TuSacPage />);
    await waitFor(() => expect(screen.getByTestId('tusac-card-0')).toBeInTheDocument());

    // 何も選んでいなければ押せない。
    expect(screen.getByTestId('tusac-meld')).toBeDisabled();
    expect(screen.getByTestId('tusac-discard')).toBeDisabled();

    fireEvent.click(screen.getByTestId('tusac-card-1'));
    fireEvent.click(screen.getByTestId('tusac-card-3'));
    expect(screen.getByTestId('tusac-selected')).toHaveTextContent('2');
    // 2 枚選ぶと捨てるは押せない (捨てるのは 1 枚)。
    expect(screen.getByTestId('tusac-discard')).toBeDisabled();

    mockApi.mockClear();
    fireEvent.click(screen.getByTestId('tusac-meld'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('meld', { indexes: [1, 3] }));
  });

  it('1 枚だけ選ぶと捨てられる', async () => {
    mockApi.mockResolvedValue(withState({ phase: TuSacPhase.DISCARD }));
    renderWithProviders(<TuSacPage />);
    await waitFor(() => expect(screen.getByTestId('tusac-card-2')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('tusac-card-2'));
    expect(screen.getByTestId('tusac-discard')).toBeEnabled();

    mockApi.mockClear();
    fireEvent.click(screen.getByTestId('tusac-discard'));
    // **0 始まりで送る。** 画面の 3 枚目は位置 2。
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('discard', { index: 2 }));
  });

  it('選択は押すたびに切り替わる', async () => {
    mockApi.mockResolvedValue(withState({ phase: TuSacPhase.DISCARD }));
    renderWithProviders(<TuSacPage />);
    await waitFor(() => expect(screen.getByTestId('tusac-card-0')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('tusac-card-0'));
    expect(screen.getByTestId('tusac-card-0')).toHaveAttribute('aria-pressed', 'true');
    fireEvent.click(screen.getByTestId('tusac-card-0'));
    expect(screen.getByTestId('tusac-card-0')).toHaveAttribute('aria-pressed', 'false');
    expect(screen.queryByTestId('tusac-selected')).not.toBeInTheDocument();
  });

  it('ラウンドと山の残りを出す', async () => {
    mockApi.mockResolvedValue(withState({ roundNumber: 3, stockCount: 7 }));
    renderWithProviders(<TuSacPage />);
    await waitFor(() => expect(screen.getByTestId('tusac-round')).toHaveTextContent('3'));
    expect(screen.getByTestId('tusac-round')).toHaveTextContent('5');
    expect(screen.getByTestId('tusac-stock')).toHaveTextContent('7');
  });

  it('決着で得点の内訳と次のラウンドを出す', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: TuSacPhase.ROUND_END,
        isHumanTurn: false,
        wentOutSeat: 0,
        seats: [
          seat({ isTurn: false, handCount: 0, cards: [], meldPoints: 12, roundScore: 12, wentOut: true }),
          seat({ name: 'CPU1', isHuman: false, cards: [], handCount: 9, isTurn: false, meldPoints: 4, roundScore: -5 }),
        ],
      }),
    );
    renderWithProviders(<TuSacPage />);
    await waitFor(() => expect(screen.getByTestId('tusac-roundscore-0')).toBeInTheDocument());
    expect(screen.getByTestId('tusac-roundscore-0')).toHaveTextContent('12');
    expect(screen.getByTestId('tusac-roundscore-1')).toHaveTextContent('-5');
    expect(screen.getByTestId('tusac-wentout-0')).toBeInTheDocument();

    mockApi.mockClear();
    fireEvent.click(screen.getByTestId('tusac-next'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('next'));
  });

  it('終局で勝者を出し、次のラウンドを出さない', async () => {
    mockApi.mockResolvedValue(
      withState({ phase: TuSacPhase.GAME_END, gameEndFlag: true, isHumanTurn: false, winnerSeat: 1 }),
    );
    renderWithProviders(<TuSacPage />);
    await waitFor(() => expect(screen.getByTestId('tusac-winner')).toHaveTextContent('CPU1'));
    expect(screen.queryByTestId('tusac-next')).not.toBeInTheDocument();
    expect(screen.queryByTestId('tusac-draw')).not.toBeInTheDocument();
  });

  it('他人の手番では操作ボタンを出さない', async () => {
    mockApi.mockResolvedValue(withState({ isHumanTurn: false, turnSeat: 1 }));
    renderWithProviders(<TuSacPage />);
    await waitFor(() => expect(screen.getByTestId('tusac-hand')).toBeInTheDocument());
    expect(screen.queryByTestId('tusac-draw')).not.toBeInTheDocument();
    expect(screen.queryByTestId('tusac-meld')).not.toBeInTheDocument();
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
    renderWithProviders(<TuSacPage />);
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
    expect(screen.queryByTestId('tusac-draw')).not.toBeInTheDocument();
  });
});
