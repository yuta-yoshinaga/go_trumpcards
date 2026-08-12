import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { crazyfourpokerApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CrazyFourPokerResponse } from '../types/card';
import { CRAZY_FOUR_POKER_RESULT } from '../types/games/crazyfourpoker';
import { CrazyFourPokerPhase } from '../types/phases';
import { CrazyFourPokerPage } from './CrazyFourPokerPage';

vi.mock('../api/gameApi', () => ({
  crazyfourpokerApi: { exec: vi.fn() },
  actionLogApi: { crazyfourpoker: vi.fn() },
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

const mockApi = vi.mocked(crazyfourpokerApi.exec);
const mockUseCliMode = vi.mocked(useCliMode);

// **文脈型が付かないので明示する。** 単体の配列だと design が string に推論されます。
const ace: Card = { design: 'SPADE', value: 1 };
const king: Card = { design: 'HEART', value: 13 };
const five: Card = { design: 'CLOVER', value: 5 };

const base: CrazyFourPokerResponse = {
  phase: CrazyFourPokerPhase.BET,
  playerHand: [],
  dealerHand: [],
  playerBest: [],
  dealerBest: [],
  playerHandRank: 0,
  dealerHandRank: 0,
  hasAcesOrBetter: false,
  maxMultiplier: 1,
  dealerQualifies: false,
  anteBet: 0,
  superBet: 0,
  queensUpBet: 0,
  playBet: 0,
  playMultiplier: 0,
  result: 0,
  payout: 0,
  chips: 1000,
  minTotalWager: 30,
  roundNumber: 1,
  remainingCards: 52,
  gameEndFlag: false,
  message: '',
};

const withState = (over: Partial<CrazyFourPokerResponse>): CrazyFourPokerResponse => ({ ...base, ...over });

const dealt = (over: Partial<CrazyFourPokerResponse> = {}) =>
  withState({
    phase: CrazyFourPokerPhase.DECIDE,
    anteBet: 50,
    superBet: 50,
    playerHand: [ace, ace, king, five, five],
    playerBest: [ace, ace, king, five],
    playerHandRank: 2,
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

describe('CrazyFourPokerPage', () => {
  it('マウント時に reset を呼ぶ', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<CrazyFourPokerPage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('賭けフェーズでは配るボタンを出す', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<CrazyFourPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '配る' })).toBeInTheDocument());
  });

  // **アンティを送ると同額の Super Bonus が付く**ので、送るのはアンティだけ。
  it('配るはアンティと Queens Up を送る', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<CrazyFourPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '配る' })).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '配る' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', { ante: 50, queensUp: 0 }));
  });

  // **エース未満のときは 1 倍のボタンしか出ない。**
  //
  // ページが手役から倍率を計算し直すとドメインとずれるので、maxMultiplier に従う。
  it('エース未満では 1 倍だけを出す', async () => {
    mockApi.mockResolvedValue(dealt({ hasAcesOrBetter: false, maxMultiplier: 1 }));
    renderWithProviders(<CrazyFourPokerPage />);
    await waitFor(() => expect(screen.getByTestId('c4p-play-1')).toBeInTheDocument());
    expect(screen.queryByTestId('c4p-play-2')).not.toBeInTheDocument();
    expect(screen.queryByTestId('c4p-play-3')).not.toBeInTheDocument();
    expect(screen.getByTestId('c4p-multiplier-notice')).toHaveTextContent('エース未満');
  });

  it('エースのペア以上では 3 倍まで出す', async () => {
    mockApi.mockResolvedValue(dealt({ hasAcesOrBetter: true, maxMultiplier: 3 }));
    renderWithProviders(<CrazyFourPokerPage />);
    await waitFor(() => expect(screen.getByTestId('c4p-play-3')).toBeInTheDocument());
    expect(screen.getByTestId('c4p-play-1')).toBeInTheDocument();
    expect(screen.getByTestId('c4p-play-2')).toBeInTheDocument();
    expect(screen.getByTestId('c4p-multiplier-notice')).toHaveTextContent('3倍');

    mockApi.mockClear();
    fireEvent.click(screen.getByTestId('c4p-play-3'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('play', { multiplier: 3 }));
  });

  it('降りるを送れる', async () => {
    mockApi.mockResolvedValue(dealt());
    renderWithProviders(<CrazyFourPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '降りる' })).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '降りる' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('fold'));
  });

  // **判断中はディーラーの手を出さない。** サーバも送っていない。
  it('判断中はディーラーの手を伏せる', async () => {
    mockApi.mockResolvedValue(dealt());
    renderWithProviders(<CrazyFourPokerPage />);
    await waitFor(() => expect(screen.getByTestId('c4p-player-hand')).toBeInTheDocument());
    expect(screen.queryByTestId('c4p-dealer-hand')).not.toBeInTheDocument();
    expect(screen.getByTestId('c4p-dealer-hidden')).toBeInTheDocument();
  });

  it('決着後はディーラーの手と収支を出す', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: CrazyFourPokerPhase.RESULT,
        anteBet: 50,
        superBet: 50,
        playBet: 50,
        playerHand: [ace, ace, king, five, five],
        playerBest: [ace, ace, king, five],
        playerHandRank: 2,
        dealerHand: [king, king, five, five, ace],
        dealerBest: [king, king, five, five],
        dealerHandRank: 3,
        dealerQualifies: true,
        result: CRAZY_FOUR_POKER_RESULT.lose,
        payout: 50,
      }),
    );
    renderWithProviders(<CrazyFourPokerPage />);
    await waitFor(() => expect(screen.getByTestId('c4p-dealer-hand')).toBeInTheDocument());
    expect(screen.queryByTestId('c4p-dealer-hidden')).not.toBeInTheDocument();
    expect(screen.getByTestId('c4p-result')).toHaveTextContent('ディーラーの勝ち');
    expect(screen.getByTestId('c4p-result')).toHaveTextContent('-100');
    expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument();
  });

  it('ディーラー不成立を表示する', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: CrazyFourPokerPhase.RESULT,
        anteBet: 50,
        superBet: 50,
        playBet: 50,
        playerHand: [ace, ace, king, five, five],
        playerBest: [ace, ace, king, five],
        playerHandRank: 2,
        dealerHand: [five, five, king, ace, ace],
        dealerBest: [five, five, king, ace],
        dealerHandRank: 1,
        dealerQualifies: false,
        result: CRAZY_FOUR_POKER_RESULT.dealerNotQualified,
        payout: 200,
      }),
    );
    renderWithProviders(<CrazyFourPokerPage />);
    await waitFor(() => expect(screen.getByTestId('c4p-dealer-rank')).toBeInTheDocument());
    expect(screen.getByTestId('c4p-dealer-rank')).toHaveTextContent('不成立');
  });

  it('チップを見出しに出す', async () => {
    mockApi.mockResolvedValue(withState({ chips: 777 }));
    renderWithProviders(<CrazyFourPokerPage />);
    await waitFor(() => expect(screen.getByTestId('c4p-chips')).toHaveTextContent('777'));
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
    renderWithProviders(<CrazyFourPokerPage />);
    await waitFor(() => expect(screen.queryByTestId('card-area')).not.toBeInTheDocument());
  });
});
