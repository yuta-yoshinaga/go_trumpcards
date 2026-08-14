import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { banluckApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BanLuckResponse, Card } from '../types/card';
import { BAN_LUCK_RANK } from '../types/games/banluck';
import { BanLuckPhase } from '../types/phases';
import { BanLuckPage } from './BanLuckPage';

vi.mock('../api/gameApi', () => ({
  banluckApi: { exec: vi.fn() },
  actionLogApi: { banluck: vi.fn() },
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

const mockApi = vi.mocked(banluckApi.exec);
const mockUseCliMode = vi.mocked(useCliMode);

const card = (value: number): Card => ({ design: 'SPADE', value });

const seat = (over: Partial<BanLuckResponse['seats'][number]> = {}) =>
  ({
    name: 'YOU',
    isHuman: true,
    chips: 1000,
    bet: 0,
    cards: [],
    score: 0,
    rank: BAN_LUCK_RANK.point,
    outcome: 1,
    roundBet: 0,
    delta: 0,
    busted: false,
    stood: false,
    isBanker: false,
    isTurn: false,
    ...over,
  }) as BanLuckResponse['seats'][number];

const base: BanLuckResponse = {
  phase: BanLuckPhase.BET,
  seats: [seat({ name: 'YOU' }), seat({ name: 'CPU1', isHuman: false, isBanker: true })],
  bankerSeat: 1,
  turnSeat: 0,
  humanSeat: 0,
  isHumanTurn: false,
  mustHit: false,
  roundNumber: 1,
  remainingCards: 52,
  winnerSeat: 0,
  gameEndFlag: false,
  message: '',
};

const withState = (over: Partial<BanLuckResponse>): BanLuckResponse => ({ ...base, ...over });

/** Play phase with the human seat holding a live hand and the turn. */
const playing = (over: Partial<BanLuckResponse> = {}): BanLuckResponse =>
  withState({
    phase: BanLuckPhase.PLAY,
    isHumanTurn: true,
    seats: [
      seat({ name: 'YOU', cards: [card(10), card(5)], score: 15, bet: 50, isTurn: true }),
      seat({ name: 'CPU1', isHuman: false, isBanker: true, cards: [card(9), card(8)], score: 17 }),
    ],
    ...over,
  });

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

describe('BanLuckPage', () => {
  it('マウント時に reset を呼ぶ', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<BanLuckPage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('賭けフェーズでは配るボタンを出す', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<BanLuckPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '配る' })).toBeInTheDocument());
  });

  it('子のときは入力した賭け金を送る', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<BanLuckPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '配る' })).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '配る' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', { bet: 50 }));
  });

  // **親のラウンドは 0 を送る。** 送らないのとは違う (サーバが 400 を返す)。
  it('親のときは 0 を送り、賭け金の入力を出さない', async () => {
    mockApi.mockResolvedValue(
      withState({
        bankerSeat: 0,
        seats: [seat({ name: 'YOU', isBanker: true }), seat({ name: 'CPU1', isHuman: false })],
      }),
    );
    renderWithProviders(<BanLuckPage />);
    await waitFor(() => expect(screen.getByTestId('bl-banker-notice')).toBeInTheDocument());
    expect(screen.queryByLabelText('賭け')).not.toBeInTheDocument();

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '配る' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', { bet: 0 }));
  });

  it('親を名指しする', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<BanLuckPage />);
    await waitFor(() => expect(screen.getByTestId('bl-banker')).toHaveTextContent('CPU1'));
    expect(screen.getByTestId('bl-banker-mark-1')).toBeInTheDocument();
    expect(screen.queryByTestId('bl-banker-mark-0')).not.toBeInTheDocument();
  });

  it('自分の手番では引くと止めるを出す', async () => {
    mockApi.mockResolvedValue(playing());
    renderWithProviders(<BanLuckPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '引く' })).toBeInTheDocument());
    expect(screen.getByTestId('bl-stand')).toBeInTheDocument();

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '引く' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('hit'));
  });

  // **止まれるかどうかはサーバの値に従う。** ページが点数から計算し直すと、
  // 規則が 2 か所に分かれてズレる (#5304)。
  it('義務があるときは止めるボタンを出さない', async () => {
    mockApi.mockResolvedValue(playing({ mustHit: true }));
    renderWithProviders(<BanLuckPage />);
    await waitFor(() => expect(screen.getByTestId('bl-must-hit')).toBeInTheDocument());
    expect(screen.queryByTestId('bl-stand')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: '引く' })).toBeInTheDocument();
  });

  // **点数が同じでもフラグが false なら止められる。** 一方向だけの検査は
  // 壊れているときに通ってしまう。
  it('同じ点数でもフラグが false なら止めるボタンが出る', async () => {
    mockApi.mockResolvedValue(
      playing({
        mustHit: false,
        seats: [
          seat({ name: 'YOU', cards: [card(10), card(4)], score: 14, isTurn: true }),
          seat({ name: 'CPU1', isHuman: false, isBanker: true }),
        ],
      }),
    );
    renderWithProviders(<BanLuckPage />);
    await waitFor(() => expect(screen.getByTestId('bl-stand')).toBeInTheDocument());
    expect(screen.queryByTestId('bl-must-hit')).not.toBeInTheDocument();
  });

  it('他人の手番では操作ボタンを出さない', async () => {
    mockApi.mockResolvedValue(playing({ isHumanTurn: false }));
    renderWithProviders(<BanLuckPage />);
    await waitFor(() => expect(screen.getByTestId('bl-seat-0')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '引く' })).not.toBeInTheDocument();
    expect(screen.queryByTestId('bl-stand')).not.toBeInTheDocument();
  });

  it('決着で役と収支を出す', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: BanLuckPhase.ROUND_END,
        seats: [
          seat({ name: 'YOU', cards: [card(1), card(1)], score: 12, rank: BAN_LUCK_RANK.banBan, delta: 150 }),
          seat({ name: 'CPU1', isHuman: false, isBanker: true, cards: [card(10), card(7)], score: 17, delta: -150 }),
        ],
      }),
    );
    renderWithProviders(<BanLuckPage />);
    await waitFor(() => expect(screen.getByTestId('bl-result-0')).toHaveTextContent('バンバン'));
    expect(screen.getByTestId('bl-result-0')).toHaveTextContent('150');
    expect(screen.getByRole('button', { name: '次のラウンドへ' })).toBeInTheDocument();
  });

  it('決着後は次のラウンドを送る', async () => {
    mockApi.mockResolvedValue(withState({ phase: BanLuckPhase.ROUND_END }));
    renderWithProviders(<BanLuckPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンドへ' })).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のラウンドへ' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('next'));
  });

  it('終局で勝者を出し、次のラウンドを出さない', async () => {
    mockApi.mockResolvedValue(withState({ phase: BanLuckPhase.GAME_END, gameEndFlag: true, winnerSeat: 1 }));
    renderWithProviders(<BanLuckPage />);
    await waitFor(() => expect(screen.getByTestId('bl-winner')).toHaveTextContent('CPU1'));
    expect(screen.queryByRole('button', { name: '次のラウンドへ' })).not.toBeInTheDocument();
  });

  it('チップとラウンドを出す', async () => {
    mockApi.mockResolvedValue(
      withState({
        roundNumber: 4,
        seats: [seat({ name: 'YOU', chips: 850 }), seat({ name: 'CPU1', isHuman: false, isBanker: true })],
      }),
    );
    renderWithProviders(<BanLuckPage />);
    await waitFor(() => expect(screen.getByTestId('bl-chips')).toHaveTextContent('850'));
    expect(screen.getByTestId('bl-round-line')).toHaveTextContent('4');
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
    renderWithProviders(<BanLuckPage />);
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '配る' })).not.toBeInTheDocument();
  });
});
