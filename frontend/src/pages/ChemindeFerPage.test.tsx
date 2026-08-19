import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { chemindeferApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, ChemindeFerResponse } from '../types/card';
import { ChemindeFerPhase } from '../types/phases';
import { ChemindeFerPage } from './ChemindeFerPage';

vi.mock('../api/gameApi', () => ({
  chemindeferApi: { exec: vi.fn() },
  actionLogApi: { chemindefer: vi.fn() },
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

const mockApi = vi.mocked(chemindeferApi.exec);
const mockUseCliMode = vi.mocked(useCliMode);

// **文脈型が付かないので明示する。** 単体の配列だと design が string に推論されます。
const king: Card = { design: 'SPADE', value: 13 };
const four: Card = { design: 'HEART', value: 4 };
const five: Card = { design: 'CLOVER', value: 5 };

const seat = (id: number, over: Partial<ChemindeFerResponse['players'][number]> = {}) => ({
  id,
  name: `Player${id + 1}`,
  isHuman: id === 0,
  chips: 1000,
  bet: 0,
  isBanker: id === 0,
  isRepresentative: false,
  lastNet: 0,
  ...over,
});

const base: ChemindeFerResponse = {
  players: [seat(0), seat(1), seat(2), seat(3), seat(4), seat(5)],
  phase: ChemindeFerPhase.STAKE,
  bankerIdx: 0,
  betTurn: -1,
  stake: 0,
  remainingStake: 0,
  totalBet: 0,
  stakeMin: 10,
  stakeMax: 1000,
  betMin: 0,
  betMax: 0,
  representativeIdx: -1,
  punterMayChoose: false,
  bankerHand: [],
  punterHand: [],
  bankerTotal: 0,
  punterTotal: 0,
  punterDrew: false,
  result: 0,
  roundNumber: 1,
  remainingCards: 312,
  isHumanTurn: true,
  gameEndFlag: false,
  message: '',
};

const withState = (over: Partial<ChemindeFerResponse>): ChemindeFerResponse => ({ ...base, ...over });

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

describe('ChemindeFerPage', () => {
  it('マウント時に reset を呼ぶ', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<ChemindeFerPage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('親のときは張りの操作を出す', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<ChemindeFerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '張る' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '賭ける' })).not.toBeInTheDocument();
  });

  it('賭けフェーズでは賭けと降りるを出す', async () => {
    mockApi.mockResolvedValue(
      withState({ phase: ChemindeFerPhase.BET, stake: 300, remainingStake: 300, betTurn: 0, betMax: 300 }),
    );
    renderWithProviders(<ChemindeFerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '賭ける' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '降りる' })).toBeInTheDocument();
  });

  // **降りるは賭け額 0 を送ること。** 送らないのとは違うので、値まで固定する。
  it('降りるは amount 0 を送る', async () => {
    mockApi.mockResolvedValue(
      withState({ phase: ChemindeFerPhase.BET, stake: 300, remainingStake: 300, betTurn: 0, betMax: 300 }),
    );
    renderWithProviders(<ChemindeFerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '降りる' })).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '降りる' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', { amount: 0 }));
  });

  // **選べない合計ではボタンを出さない。**
  //
  // ページが合計から判定し直すとドメインとずれるので、`punterMayChoose` を使う。
  // ここが壊れると、規則で決まっている手を player に選ばせてしまう。
  it('子の合計が選べないときは引く/立つを出さず、理由を表示する', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: ChemindeFerPhase.PUNTER_DRAW,
        punterMayChoose: false,
        punterHand: [king, four],
        punterTotal: 4,
        bankerHand: [king, king],
        representativeIdx: 0,
      }),
    );
    renderWithProviders(<ChemindeFerPage />);
    await waitFor(() => expect(screen.getByTestId('cdf-forced-notice')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '引く' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '立つ' })).not.toBeInTheDocument();
  });

  it('子の合計が 5 のときだけ引く/立つを出す', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: ChemindeFerPhase.PUNTER_DRAW,
        punterMayChoose: true,
        punterHand: [king, five],
        punterTotal: 5,
        bankerHand: [king, king],
        representativeIdx: 0,
      }),
    );
    renderWithProviders(<ChemindeFerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '引く' })).toBeInTheDocument());
    expect(screen.queryByTestId('cdf-forced-notice')).not.toBeInTheDocument();

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '引く' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('pd'));
  });

  // **親はどの合計でも選べる。** 子と同じ扱いにすると自由が消える。
  it('親の判断ではどの合計でも引く/立つを出す', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: ChemindeFerPhase.BANKER_DRAW,
        punterMayChoose: false,
        bankerHand: [king, { design: 'SPADE', value: 7 }],
        bankerTotal: 7,
        punterHand: [king, four],
        punterTotal: 4,
        representativeIdx: 1,
      }),
    );
    renderWithProviders(<ChemindeFerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '引く' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '立つ' })).toBeInTheDocument();
    expect(screen.queryByTestId('cdf-forced-notice')).not.toBeInTheDocument();

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '立つ' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bs'));
  });

  it('ラウンド終了では次へと、親のときだけバンクを渡すを出す', async () => {
    mockApi.mockResolvedValue(
      withState({ phase: ChemindeFerPhase.ROUND_END, result: 1, isHumanTurn: false, bankerIdx: 0 }),
    );
    renderWithProviders(<ChemindeFerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'バンクを渡す' })).toBeInTheDocument();
    expect(screen.getByTestId('cdf-result')).toHaveTextContent('親の勝ち');
  });

  // **卓の結果と自分の損益は別の情報** (#5774)。
  it('自分の賭けが勝ったか負けたかを金額付きで出す', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: ChemindeFerPhase.ROUND_END,
        result: 2,
        isHumanTurn: false,
        players: [seat(0, { isHuman: true, lastNet: 200 }), seat(1, { isBanker: true, lastNet: -200 })],
      }),
    );
    renderWithProviders(<ChemindeFerPage />);

    const net = await screen.findByTestId('cdf-net');
    expect(net).toHaveTextContent('+200');
    expect(net.className).toContain('text-ds-success');
  });

  it('負けた回は赤で、賭けていない回は増減なしと出す', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: ChemindeFerPhase.ROUND_END,
        result: 1,
        isHumanTurn: false,
        players: [seat(0, { isHuman: true, lastNet: -50 }), seat(1, { isBanker: true, lastNet: 50 })],
      }),
    );
    const { unmount } = renderWithProviders(<ChemindeFerPage />);
    const lost = await screen.findByTestId('cdf-net');
    expect(lost).toHaveTextContent('-50');
    expect(lost.className).toContain('text-ds-error');
    unmount();

    // **賭けていない回に行ごと消すと、勝ったのか賭けていないのかが読めない。**
    mockApi.mockResolvedValue(withState({ phase: ChemindeFerPhase.ROUND_END, result: 3, isHumanTurn: false }));
    renderWithProviders(<ChemindeFerPage />);
    expect(await screen.findByTestId('cdf-net')).toHaveTextContent('増減なし');
  });

  it('親でなければバンクを渡すは出ない', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: ChemindeFerPhase.ROUND_END,
        result: 2,
        isHumanTurn: false,
        bankerIdx: 2,
        players: [seat(0, { isBanker: false }), seat(1, { isBanker: false }), seat(2, { isBanker: true })],
      }),
    );
    renderWithProviders(<ChemindeFerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'バンクを渡す' })).not.toBeInTheDocument();
  });

  it('席ごとにチップと親の印を出す', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: ChemindeFerPhase.BET,
        stake: 200,
        betTurn: 1,
        isHumanTurn: false,
        players: [
          seat(0, { isBanker: true }),
          seat(1, { isBanker: false, bet: 50, chips: 950 }),
          seat(2, { isBanker: false }),
        ],
      }),
    );
    renderWithProviders(<ChemindeFerPage />);
    await waitFor(() => expect(screen.getByTestId('cdf-seat-0')).toBeInTheDocument());
    expect(screen.getByTestId('cdf-seat-0')).toHaveTextContent('★');
    expect(screen.getByTestId('cdf-seat-1')).toHaveTextContent('950');
    expect(screen.getByTestId('cdf-seat-1')).toHaveTextContent('50');
  });

  it('自分の手番でないときは待機を表示する', async () => {
    mockApi.mockResolvedValue(withState({ phase: ChemindeFerPhase.BET, stake: 200, betTurn: 3, isHumanTurn: false }));
    renderWithProviders(<ChemindeFerPage />);
    await waitFor(() => expect(screen.getByTestId('cdf-wait-notice')).toBeInTheDocument());
  });

  it('両者の手札と合計を表示する', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: ChemindeFerPhase.ROUND_END,
        isHumanTurn: false,
        punterHand: [king, five],
        punterTotal: 5,
        bankerHand: [king, four],
        bankerTotal: 4,
        result: 2,
      }),
    );
    renderWithProviders(<ChemindeFerPage />);
    await waitFor(() => expect(screen.getByTestId('cdf-punter-hand-total')).toHaveTextContent('5'));
    expect(screen.getByTestId('cdf-banker-hand-total')).toHaveTextContent('4');
  });

  it('CLI モードでは端末を出す', async () => {
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
    renderWithProviders(<ChemindeFerPage />);
    await waitFor(() => expect(screen.queryByTestId('card-area')).not.toBeInTheDocument());
  });
});
