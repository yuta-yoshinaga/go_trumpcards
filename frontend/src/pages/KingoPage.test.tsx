import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { kingoApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, KingoResponse } from '../types/card';
import { KingoPhase } from '../types/phases';
import { KingoPage } from './KingoPage';

vi.mock('../api/gameApi', () => ({
  kingoApi: { exec: vi.fn() },
  actionLogApi: { kingo: vi.fn() },
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

const mockApi = vi.mocked(kingoApi.exec);
const mockUseCliMode = vi.mocked(useCliMode);

const card = (value: number): Card => ({ design: 'SPADE', value });

const seat = (over: Partial<KingoResponse['seats'][number]> = {}) =>
  ({
    name: 'YOU',
    isHuman: true,
    chips: 1000,
    bet: 0,
    cards: [],
    rank: 0,
    matchedValue: 0,
    isBanker: false,
    wonAmount: 0,
    ...over,
  }) as KingoResponse['seats'][number];

const base: KingoResponse = {
  phase: KingoPhase.BET,
  seats: [
    seat(),
    seat({ name: 'CPU1', isHuman: false, bet: 10 }),
    seat({ name: 'CPU2', isHuman: false, isBanker: true }),
    seat({ name: 'CPU3', isHuman: false, bet: 20 }),
  ],
  bankerSeat: 2,
  roundNumber: 1,
  rounds: 10,
  humanSeat: 0,
  isHumanBanker: false,
  isHumanTurn: true,
  handSize: 3,
  payoutArashi: 3,
  payoutPair: 1,
  remainingCards: 28,
  winnerSeat: 0,
  gameEndFlag: false,
  message: '',
};

const withState = (over: Partial<KingoResponse>): KingoResponse => ({ ...base, ...over });

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

describe('KingoPage', () => {
  it('マウント時に reset を呼ぶ', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<KingoPage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  // **配る前は誰の手札も無い。** 隠しているのではなく、まだ存在しない。
  it('張りの段階では手札を出さない', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<KingoPage />);
    await waitFor(() => expect(screen.getByTestId('kingo-cards-0')).toBeInTheDocument());
    expect(screen.getByTestId('kingo-cards-0')).toHaveTextContent('配る前です');
    expect(screen.queryByTestId('kingo-rank-0')).not.toBeInTheDocument();
  });

  // **決着では全員ぶんが出る。** 一方向だけの検査は壊れているときに通る。
  it('決着で全員の手札と役を出す', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: KingoPhase.RESULT,
        isHumanTurn: false,
        seats: [
          seat({ bet: 10, cards: [card(3), card(3), card(8)], rank: 1, matchedValue: 3, wonAmount: 10 }),
          seat({ name: 'CPU1', isHuman: false, bet: 10, cards: [card(1), card(5), card(9)], wonAmount: -10 }),
          seat({ name: 'CPU2', isHuman: false, isBanker: true, cards: [card(7), card(7), card(7)], rank: 2 }),
          seat({ name: 'CPU3', isHuman: false, bet: 20, cards: [card(2), card(6), card(10)], wonAmount: -20 }),
        ],
      }),
    );
    renderWithProviders(<KingoPage />);
    await waitFor(() => expect(screen.getByTestId('kingo-cards-0').children).toHaveLength(3));
    expect(screen.getByTestId('kingo-rank-0')).toHaveTextContent('2枚そろい');
    expect(screen.getByTestId('kingo-rank-2')).toHaveTextContent('嵐');
    expect(screen.getByTestId('kingo-won-0')).toHaveTextContent('10');
  });

  // **同じ「嵐」でも K 3 枚と A 3 枚では強さの実感が違う** (#5783)。
  it('そろえた数字まで出す', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: KingoPhase.RESULT,
        isHumanTurn: false,
        seats: [
          seat({ cards: [card(3), card(3), card(8)], rank: 1, matchedValue: 3 }),
          seat({ name: 'CPU1', isHuman: false, cards: [card(13), card(13), card(13)], rank: 2, matchedValue: 13 }),
          seat({ name: 'CPU2', isHuman: false, cards: [card(1), card(1), card(1)], rank: 2, matchedValue: 1 }),
          seat({ name: 'CPU3', isHuman: false, cards: [card(2), card(6), card(10)], rank: 0, matchedValue: 10 }),
        ],
      }),
    );
    renderWithProviders(<KingoPage />);

    await waitFor(() => expect(screen.getByTestId('kingo-rank-0')).toHaveTextContent('3'));
    // 絵札は札の表記で出す。K の嵐と A の嵐が見分けられる。
    expect(screen.getByTestId('kingo-rank-1')).toHaveTextContent('嵐（K）');
    expect(screen.getByTestId('kingo-rank-2')).toHaveTextContent('嵐（A）');
    // **役なしの席には数字を出さない**（受け入れ条件3）。
    expect(screen.getByTestId('kingo-rank-3')).toHaveTextContent('役なし');
    expect(screen.getByTestId('kingo-rank-3').textContent).not.toContain('10');
  });

  // **親と子で出す操作が変わる。** サーバの isHumanBanker に従う。
  it('子には張る、親には配るを出す', async () => {
    mockApi.mockResolvedValue(base);
    const { unmount } = renderWithProviders(<KingoPage />);
    await waitFor(() => expect(screen.getByTestId('kingo-bet')).toBeInTheDocument());
    expect(screen.queryByTestId('kingo-deal')).not.toBeInTheDocument();
    unmount();

    mockApi.mockResolvedValue(
      withState({
        isHumanBanker: true,
        bankerSeat: 0,
        seats: [
          seat({ isBanker: true }),
          seat({ name: 'CPU1', isHuman: false, bet: 10 }),
          seat({ name: 'CPU2', isHuman: false, bet: 10 }),
          seat({ name: 'CPU3', isHuman: false, bet: 20 }),
        ],
      }),
    );
    renderWithProviders(<KingoPage />);
    await waitFor(() => expect(screen.getByTestId('kingo-deal')).toBeInTheDocument());
    expect(screen.queryByTestId('kingo-bet')).not.toBeInTheDocument();
    expect(screen.getByTestId('kingo-banker')).toHaveTextContent('あなたが親です');
  });

  it('各操作をそのまま送る', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<KingoPage />);
    await waitFor(() => expect(screen.getByTestId('kingo-bet')).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByTestId('kingo-bet'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', { amount: 10 }));
    expect(mockApi).not.toHaveBeenCalledWith('deal');
  });

  it('親のときは配るを送る', async () => {
    mockApi.mockResolvedValue(withState({ isHumanBanker: true, bankerSeat: 0 }));
    renderWithProviders(<KingoPage />);
    await waitFor(() => expect(screen.getByTestId('kingo-deal')).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByTestId('kingo-deal'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('deal'));
    expect(mockApi).not.toHaveBeenCalledWith('bet', { amount: 10 });
  });

  // **配当はサーバが送った値を出す。** 画面が 3 倍を持たない。
  it('サーバが送った配当倍率を出す', async () => {
    mockApi.mockResolvedValue(base);
    const { unmount } = renderWithProviders(<KingoPage />);
    await waitFor(() => expect(screen.getByTestId('kingo-payouts')).toHaveTextContent('3'));
    unmount();

    mockApi.mockResolvedValue(withState({ payoutArashi: 7, payoutPair: 2 }));
    renderWithProviders(<KingoPage />);
    await waitFor(() => expect(screen.getByTestId('kingo-payouts')).toHaveTextContent('7'));
    expect(screen.getByTestId('kingo-payouts')).toHaveTextContent('2');
  });

  it('ラウンドと親を出す', async () => {
    mockApi.mockResolvedValue(withState({ roundNumber: 4, rounds: 10 }));
    renderWithProviders(<KingoPage />);
    await waitFor(() => expect(screen.getByTestId('kingo-round')).toHaveTextContent('4'));
    expect(screen.getByTestId('kingo-round')).toHaveTextContent('10');
    expect(screen.getByTestId('kingo-banker')).toHaveTextContent('CPU2');
  });

  it('決着で次のラウンドへ進める', async () => {
    mockApi.mockResolvedValue(withState({ phase: KingoPhase.RESULT, isHumanTurn: false }));
    renderWithProviders(<KingoPage />);
    await waitFor(() => expect(screen.getByTestId('kingo-next')).toBeInTheDocument());
    expect(screen.queryByTestId('kingo-bet')).not.toBeInTheDocument();

    mockApi.mockClear();
    fireEvent.click(screen.getByTestId('kingo-next'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('next'));
  });

  it('終局で勝者を出し、次のラウンドを出さない', async () => {
    mockApi.mockResolvedValue(
      withState({ phase: KingoPhase.GAME_END, gameEndFlag: true, isHumanTurn: false, winnerSeat: 1 }),
    );
    renderWithProviders(<KingoPage />);
    await waitFor(() => expect(screen.getByTestId('kingo-winner')).toHaveTextContent('CPU1'));
    expect(screen.queryByTestId('kingo-next')).not.toBeInTheDocument();
    expect(screen.queryByTestId('kingo-bet')).not.toBeInTheDocument();
  });

  it('チップを出す', async () => {
    mockApi.mockResolvedValue(withState({ seats: [seat({ chips: 870 }), ...base.seats.slice(1)] }));
    renderWithProviders(<KingoPage />);
    await waitFor(() => expect(screen.getByTestId('kingo-chips')).toHaveTextContent('870'));
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
    renderWithProviders(<KingoPage />);
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
    expect(screen.queryByTestId('kingo-bet')).not.toBeInTheDocument();
  });
});
