import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { freebetApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, FreeBetResponse } from '../types/card';
import { FREE_BET_RESULT } from '../types/games/freebet';
import { FreeBetPhase } from '../types/phases';
import { FreeBetPage } from './FreeBetPage';

vi.mock('../api/gameApi', () => ({
  freebetApi: { exec: vi.fn() },
  actionLogApi: { freebet: vi.fn() },
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

const mockApi = vi.mocked(freebetApi.exec);
const mockUseCliMode = vi.mocked(useCliMode);

const card = (value: number): Card => ({ design: 'SPADE', value });

const hand = (over: Partial<FreeBetResponse['hands'][number]> = {}) =>
  ({
    cards: [card(13), card(7)],
    score: 17,
    bet: 50,
    freeBet: 0,
    isSoft: false,
    stood: false,
    doubled: false,
    busted: false,
    blackjack: false,
    result: 0,
    ...over,
  }) as FreeBetResponse['hands'][number];

const base: FreeBetResponse = {
  phase: FreeBetPhase.BET,
  hands: [],
  activeHand: 0,
  dealerCards: [],
  dealerScore: 0,
  dealerPushed22: false,
  canFreeDouble: false,
  canFreeSplit: false,
  anteBet: 0,
  payout: 0,
  chips: 1000,
  roundNumber: 1,
  remainingCards: 312,
  gameEndFlag: false,
  message: '',
};

const withState = (over: Partial<FreeBetResponse>): FreeBetResponse => ({ ...base, ...over });

const playing = (over: Partial<FreeBetResponse> = {}): FreeBetResponse =>
  withState({
    phase: FreeBetPhase.PLAY,
    anteBet: 50,
    dealerCards: [card(6), card(9)],
    dealerScore: 15,
    hands: [hand()],
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

describe('FreeBetPage', () => {
  it('マウント時に reset を呼ぶ', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<FreeBetPage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('賭けフェーズでは配るボタンを出す', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<FreeBetPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '配る' })).toBeInTheDocument());
  });

  it('配るはアンティを送る', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<FreeBetPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '配る' })).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '配る' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', { ante: 50 }));
  });

  it('プレイフェーズではヒットとスタンドを出す', async () => {
    mockApi.mockResolvedValue(playing());
    renderWithProviders(<FreeBetPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'スタンド' })).toBeInTheDocument();

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'ヒット' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('hit'));
  });

  // **可否はサーバの値に従う。** ページが手札から計算し直すと、規則が 2 か所に
  // 分かれてズレる (#5304)。フラグが false のときにボタンが出ないことまで見る。
  it('無料ダブル / 無料スプリットはサーバのフラグでのみ出る', async () => {
    mockApi.mockResolvedValue(playing({ hands: [hand({ score: 10, cards: [card(5), card(5)] })] }));
    const { unmount } = renderWithProviders(<FreeBetPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());
    // 5-5 はハード10 かつ対 — ドメインなら両方使える形だが、フラグが false なら出ない。
    expect(screen.queryByTestId('fb-freedouble')).not.toBeInTheDocument();
    expect(screen.queryByTestId('fb-freesplit')).not.toBeInTheDocument();
    unmount();

    mockApi.mockResolvedValue(playing({ canFreeDouble: true, canFreeSplit: true }));
    renderWithProviders(<FreeBetPage />);
    await waitFor(() => expect(screen.getByTestId('fb-freedouble')).toBeInTheDocument());
    expect(screen.getByTestId('fb-freesplit')).toBeInTheDocument();
  });

  it('無料ダブルと無料スプリットは引数なしで送る', async () => {
    mockApi.mockResolvedValue(playing({ canFreeDouble: true, canFreeSplit: true }));
    renderWithProviders(<FreeBetPage />);
    await waitFor(() => expect(screen.getByTestId('fb-freedouble')).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByTestId('fb-freedouble'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('freedouble'));

    mockApi.mockClear();
    fireEvent.click(screen.getByTestId('fb-freesplit'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('freesplit'));
  });

  // **自分の金とハウスの金を合算しない。** 合算すると「いくら失うのか」が消える。
  it('ハウスの出資を賭け金と別に出す', async () => {
    mockApi.mockResolvedValue(playing({ hands: [hand({ bet: 50, freeBet: 50, doubled: true })] }));
    renderWithProviders(<FreeBetPage />);
    await waitFor(() => expect(screen.getByTestId('fb-free-0')).toBeInTheDocument());
    expect(screen.getByTestId('fb-hand-0')).toHaveTextContent('賭け 50');
    expect(screen.getByTestId('fb-free-0')).toHaveTextContent('ハウス 50');
    expect(screen.getByTestId('fb-hand-0')).not.toHaveTextContent('賭け 100');
  });

  it('ハウスの出資が無い手札には表示しない', async () => {
    mockApi.mockResolvedValue(playing());
    renderWithProviders(<FreeBetPage />);
    await waitFor(() => expect(screen.getByTestId('fb-hand-0')).toBeInTheDocument());
    expect(screen.queryByTestId('fb-free-0')).not.toBeInTheDocument();
  });

  // **収支はプレイヤーが出した金だけで測る。** ハウスの出資を賭け金に数えると、
  // 無料ダブルして勝ったラウンドの黒字が実際より小さく出る。
  it('収支の計算にハウスの出資を含めない', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: FreeBetPhase.RESULT,
        anteBet: 50,
        dealerCards: [card(6), card(9), card(5)],
        dealerScore: 20,
        // 賭け 50 + ハウス 50 で勝つと払い戻しは 150。自腹は 50 なので収支は +100。
        hands: [hand({ bet: 50, freeBet: 50, doubled: true, result: FREE_BET_RESULT.win })],
        payout: 150,
      }),
    );
    renderWithProviders(<FreeBetPage />);
    await waitFor(() => expect(screen.getByTestId('fb-result')).toHaveTextContent('収支: 100'));
  });

  // **22 は名指しする。** 無料ダブル / 無料スプリットの対価がこれ。
  it('ディーラーの22を名指しする', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: FreeBetPhase.RESULT,
        anteBet: 50,
        dealerCards: [card(6), card(9), card(7)],
        dealerScore: 22,
        dealerPushed22: true,
        hands: [hand({ result: FREE_BET_RESULT.dealer22Push })],
        payout: 50,
      }),
    );
    renderWithProviders(<FreeBetPage />);
    await waitFor(() => expect(screen.getByTestId('fb-dealer22')).toBeInTheDocument());
    expect(screen.getByTestId('fb-hand-0')).toHaveTextContent('引き分け（ディーラー22）');
  });

  it('22でないときは名指ししない', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: FreeBetPhase.RESULT,
        anteBet: 50,
        dealerCards: [card(6), card(9), card(8)],
        dealerScore: 23,
        hands: [hand({ result: FREE_BET_RESULT.win })],
        payout: 100,
      }),
    );
    renderWithProviders(<FreeBetPage />);
    await waitFor(() => expect(screen.getByTestId('fb-result')).toBeInTheDocument());
    expect(screen.queryByTestId('fb-dealer22')).not.toBeInTheDocument();
  });

  it('決着後は次のラウンドを押せる', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: FreeBetPhase.RESULT,
        anteBet: 50,
        hands: [hand({ result: FREE_BET_RESULT.lose })],
      }),
    );
    renderWithProviders(<FreeBetPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のラウンド' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('next'));
  });

  it('破産したら次のラウンドを出さない', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: FreeBetPhase.RESULT,
        anteBet: 50,
        chips: 0,
        gameEndFlag: true,
        hands: [hand({ result: FREE_BET_RESULT.lose })],
      }),
    );
    renderWithProviders(<FreeBetPage />);
    await waitFor(() => expect(screen.getByTestId('fb-result')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '次のラウンド' })).not.toBeInTheDocument();
  });

  it('チップとラウンドを出す', async () => {
    mockApi.mockResolvedValue(withState({ chips: 850, roundNumber: 3, anteBet: 50 }));
    renderWithProviders(<FreeBetPage />);
    await waitFor(() => expect(screen.getByTestId('fb-chips')).toHaveTextContent('850'));
    expect(screen.getByTestId('fb-bet-line')).toHaveTextContent('3');
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
    renderWithProviders(<FreeBetPage />);
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '配る' })).not.toBeInTheDocument();
  });
});
