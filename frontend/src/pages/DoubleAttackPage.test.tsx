import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { doubleattackApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, DoubleAttackResponse } from '../types/card';
import { DOUBLE_ATTACK_RESULT } from '../types/games/doubleattack';
import { DoubleAttackPhase } from '../types/phases';
import { DoubleAttackPage } from './DoubleAttackPage';

vi.mock('../api/gameApi', () => ({
  doubleattackApi: { exec: vi.fn() },
  actionLogApi: { doubleattack: vi.fn() },
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

const mockApi = vi.mocked(doubleattackApi.exec);
const mockUseCliMode = vi.mocked(useCliMode);

const card = (value: number): Card => ({ design: 'SPADE', value });

const hand = (over: Partial<DoubleAttackResponse['hands'][number]> = {}) =>
  ({
    cards: [card(13), card(7)],
    score: 17,
    bet: 100,
    isSoft: false,
    stood: false,
    doubled: false,
    busted: false,
    blackjack: false,
    result: 0,
    ...over,
  }) as DoubleAttackResponse['hands'][number];

const base: DoubleAttackResponse = {
  phase: DoubleAttackPhase.BET,
  hands: [],
  activeHand: 0,
  dealerCards: [],
  dealerScore: 0,
  dealerHoleDealt: false,
  maxAttackBet: 0,
  canDouble: false,
  canSplit: false,
  anteBet: 0,
  attackBet: 0,
  bustItBet: 0,
  payout: 0,
  bustItPayout: 0,
  chips: 1000,
  roundNumber: 1,
  remainingCards: 384,
  gameEndFlag: false,
  message: '',
};

const withState = (over: Partial<DoubleAttackResponse>): DoubleAttackResponse => ({ ...base, ...over });

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

describe('DoubleAttackPage', () => {
  it('マウント時に reset を呼ぶ', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('賭けフェーズでは配るボタンを出す', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '配る' })).toBeInTheDocument());
  });

  it('配るはアンティと Bust It を送る', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '配る' })).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '配る' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', { ante: 50, bustIt: 0 }));
  });

  // **追加ベットの前は 1 枚だけで、点数も出ない。**
  //
  // これは伏せているのではなく、サーバがまだ 2 枚目を持っていないため。
  it('追加ベットの前はアップカードだけを出し、点数を出さない', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: DoubleAttackPhase.ATTACK,
        anteBet: 50,
        dealerCards: [card(6)],
        dealerHoleDealt: false,
        maxAttackBet: 50,
        hands: [hand()],
      }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByTestId('da-dealer-hidden')).toBeInTheDocument());
    expect(screen.queryByTestId('da-dealer-score')).not.toBeInTheDocument();
    expect(screen.getByTestId('da-attack-notice')).toBeInTheDocument();
  });

  // **見送りは amount 0 を送ること。** 送らないのとは違う。
  it('見送りは amount 0 を送る', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: DoubleAttackPhase.ATTACK,
        anteBet: 50,
        dealerCards: [card(6)],
        maxAttackBet: 50,
        hands: [hand()],
      }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '見送る' })).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '見送る' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('attack', { amount: 0 }));
  });

  it('2枚目が配られたら点数を出す', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: DoubleAttackPhase.PLAY,
        anteBet: 50,
        dealerCards: [card(6), card(9)],
        dealerScore: 15,
        dealerHoleDealt: true,
        hands: [hand()],
      }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByTestId('da-dealer-score')).toHaveTextContent('15'));
    expect(screen.queryByTestId('da-dealer-hidden')).not.toBeInTheDocument();
  });

  it('プレイフェーズではヒットとスタンドを出す', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: DoubleAttackPhase.PLAY,
        anteBet: 50,
        dealerCards: [card(6), card(9)],
        dealerHoleDealt: true,
        hands: [hand()],
      }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'スタンド' })).toBeInTheDocument();

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'ヒット' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('hit'));
  });

  // **押せるかどうかはサーバが決める。** 手札から計算し直さない。
  it('canDouble / canSplit が false のときボタンを出さない', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: DoubleAttackPhase.PLAY,
        anteBet: 50,
        dealerCards: [card(6), card(9)],
        dealerHoleDealt: true,
        hands: [hand()],
        canDouble: false,
        canSplit: false,
      }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());
    expect(screen.queryByTestId('da-double')).not.toBeInTheDocument();
    expect(screen.queryByTestId('da-split')).not.toBeInTheDocument();
  });

  it('canDouble / canSplit が true のときボタンを出す', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: DoubleAttackPhase.PLAY,
        anteBet: 50,
        dealerCards: [card(6), card(9)],
        dealerHoleDealt: true,
        hands: [hand({ cards: [card(8), card(8)], score: 16 })],
        canDouble: true,
        canSplit: true,
      }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByTestId('da-double')).toBeInTheDocument());
    expect(screen.getByTestId('da-split')).toBeInTheDocument();

    mockApi.mockClear();
    fireEvent.click(screen.getByTestId('da-split'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('split'));
  });

  it('スプリット後は手札が複数出る', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: DoubleAttackPhase.PLAY,
        anteBet: 50,
        dealerCards: [card(6), card(9)],
        dealerHoleDealt: true,
        hands: [hand(), hand({ score: 19 })],
        activeHand: 1,
      }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByTestId('da-hand-0')).toBeInTheDocument());
    expect(screen.getByTestId('da-hand-1')).toBeInTheDocument();
  });

  it('決着では収支と次へを出す', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: DoubleAttackPhase.RESULT,
        anteBet: 50,
        attackBet: 50,
        bustItBet: 20,
        dealerCards: [card(6), card(9), card(10)],
        dealerScore: 25,
        dealerHoleDealt: true,
        hands: [hand({ result: DOUBLE_ATTACK_RESULT.win, bet: 100 })],
        payout: 200,
        bustItPayout: 60,
      }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByTestId('da-result')).toBeInTheDocument());
    // 賭け 100 + Bust It 20 = 120 に対し払い戻し 200 -> 収支 +80
    expect(screen.getByTestId('da-result')).toHaveTextContent('80');
    // **賭けたのに結果が見えない状態をなくす** (#5776)。
    expect(screen.getByTestId('da-bustit-result')).toHaveTextContent('Bust It 的中');
    expect(screen.getByTestId('da-bustit-result')).toHaveTextContent('60');
    expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument();
  });

  // **払い戻し 0 でも「外れ」と言う。** 何も出ないと、賭けたこと自体が
  // 無かったように見える (#5776)。
  it('Bust It が外れた回もそう言う', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: DoubleAttackPhase.RESULT,
        anteBet: 50,
        attackBet: 50,
        bustItBet: 20,
        dealerHoleDealt: true,
        hands: [hand({ result: DOUBLE_ATTACK_RESULT.lose, bet: 100 })],
        payout: 0,
        bustItPayout: 0,
      }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByTestId('da-bustit-result')).toBeInTheDocument());
    expect(screen.getByTestId('da-bustit-result')).toHaveTextContent('外れ');
  });

  // **負のコントロール: 賭けていない回に行は出さない。**
  it('Bust It に賭けていない回は行を出さない', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: DoubleAttackPhase.RESULT,
        anteBet: 50,
        attackBet: 50,
        bustItBet: 0,
        dealerHoleDealt: true,
        hands: [hand({ result: DOUBLE_ATTACK_RESULT.win, bet: 100 })],
        payout: 200,
        bustItPayout: 0,
      }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByTestId('da-result')).toBeInTheDocument());
    expect(screen.queryByTestId('da-bustit-result')).not.toBeInTheDocument();
  });

  it('チップを見出しに出す', async () => {
    mockApi.mockResolvedValue(withState({ chips: 777 }));
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByTestId('da-chips')).toHaveTextContent('777'));
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
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.queryByTestId('card-area')).not.toBeInTheDocument());
  });
});
