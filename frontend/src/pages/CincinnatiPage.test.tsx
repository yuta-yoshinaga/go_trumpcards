import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { cincinnatiApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CincinnatiResponse } from '../types/card';
import { CincinnatiPhase } from '../types/phases';
import { CincinnatiPage } from './CincinnatiPage';

vi.mock('../api/gameApi', () => ({
  cincinnatiApi: { exec: vi.fn() },
  actionLogApi: { cincinnati: vi.fn() },
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

const mockApi = vi.mocked(cincinnatiApi.exec);
const mockUseCliMode = vi.mocked(useCliMode);

const card = (value: number): Card => ({ design: 'SPADE', value });
const hand = () => [card(1), card(2), card(3), card(4), card(5)];

const seat = (over: Partial<CincinnatiResponse['seats'][number]> = {}) =>
  ({
    name: 'YOU',
    isHuman: true,
    chips: 1000,
    bet: 0,
    cards: hand(),
    folded: false,
    allIn: false,
    isTurn: true,
    handRank: 0,
    bestHand: [],
    wonAmount: 0,
    ...over,
  }) as CincinnatiResponse['seats'][number];

const base: CincinnatiResponse = {
  phase: CincinnatiPhase.BETTING,
  seats: [seat(), seat({ name: 'CPU1', isHuman: false, cards: [], isTurn: false })],
  community: [],
  revealedCount: 0,
  communityTotal: 5,
  pot: 40,
  currentBet: 0,
  toCall: 0,
  raiseCount: 0,
  canRaise: true,
  turnSeat: 0,
  humanSeat: 0,
  isHumanTurn: true,
  handNumber: 1,
  remainingCards: 30,
  winnerSeat: 0,
  gameEndFlag: false,
  message: '',
};

const withState = (over: Partial<CincinnatiResponse>): CincinnatiResponse => ({ ...base, ...over });

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

describe('CincinnatiPage', () => {
  it('マウント時に reset を呼ぶ', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<CincinnatiPage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('自分の手札5枚を出す', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<CincinnatiPage />);
    await waitFor(() => expect(screen.getByTestId('cin-hand')).toBeInTheDocument());
    expect(screen.getByTestId('cin-hand').children).toHaveLength(5);
  });

  // **あと何枚めくれるかを必ず出す。** 残りの回数だけベットラウンドがある。
  it('公開枚数と総数を出す', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<CincinnatiPage />);
    await waitFor(() => expect(screen.getByTestId('cin-revealed')).toHaveTextContent('0'));
    expect(screen.getByTestId('cin-revealed')).toHaveTextContent('5');
  });

  // **CPU の手札はサーバが送っていない。** 届いていなければ伏せ表示。
  it('届いていない手札は伏せ表示にする', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<CincinnatiPage />);
    await waitFor(() => expect(screen.getByTestId('cin-seat-cards-1')).toBeInTheDocument());
    expect(screen.getByTestId('cin-seat-cards-1')).toHaveTextContent('伏せ');
  });

  // **届いていれば開く。** 一方向だけの検査は壊れているときに通る。
  it('ショーダウンで届いた手札を開く', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: CincinnatiPhase.SHOWDOWN,
        seats: [seat({ isTurn: false }), seat({ name: 'CPU1', isHuman: false, isTurn: false })],
      }),
    );
    renderWithProviders(<CincinnatiPage />);
    await waitFor(() => expect(screen.getByTestId('cin-seat-cards-1')).toBeInTheDocument());
    expect(screen.getByTestId('cin-seat-cards-1')).not.toHaveTextContent('伏せ');
    expect(screen.getByTestId('cin-seat-cards-1').children).toHaveLength(5);
  });

  // **チェックとコールは場況で入れ替わる。** サーバの toCall に従う。
  it('賭けが無ければチェック、あればコールを出す', async () => {
    mockApi.mockResolvedValue(base);
    const { unmount } = renderWithProviders(<CincinnatiPage />);
    await waitFor(() => expect(screen.getByTestId('cin-check')).toBeInTheDocument());
    expect(screen.queryByTestId('cin-call')).not.toBeInTheDocument();
    expect(screen.getByTestId('cin-bet')).toBeInTheDocument();
    unmount();

    mockApi.mockResolvedValue(withState({ toCall: 20, currentBet: 20 }));
    renderWithProviders(<CincinnatiPage />);
    await waitFor(() => expect(screen.getByTestId('cin-call')).toBeInTheDocument());
    expect(screen.queryByTestId('cin-check')).not.toBeInTheDocument();
    expect(screen.getByTestId('cin-raise')).toBeInTheDocument();
  });

  // **レイズの可否はサーバが決める。** 上限に達したら出さない。
  it('レイズ上限に達したらレイズを出さない', async () => {
    mockApi.mockResolvedValue(withState({ toCall: 20, currentBet: 20, canRaise: false }));
    renderWithProviders(<CincinnatiPage />);
    await waitFor(() => expect(screen.getByTestId('cin-call')).toBeInTheDocument());
    expect(screen.queryByTestId('cin-raise')).not.toBeInTheDocument();
  });

  it('各操作をそのまま送る', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<CincinnatiPage />);
    await waitFor(() => expect(screen.getByTestId('cin-check')).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByTestId('cin-check'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('check'));

    mockApi.mockClear();
    fireEvent.click(screen.getByTestId('cin-fold'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('fold'));

    mockApi.mockClear();
    fireEvent.click(screen.getByTestId('cin-bet'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', { amount: 20 }));
  });

  it('他人の手番では操作ボタンを出さない', async () => {
    mockApi.mockResolvedValue(withState({ isHumanTurn: false, turnSeat: 1 }));
    renderWithProviders(<CincinnatiPage />);
    await waitFor(() => expect(screen.getByTestId('cin-hand')).toBeInTheDocument());
    expect(screen.queryByTestId('cin-check')).not.toBeInTheDocument();
    expect(screen.queryByTestId('cin-fold')).not.toBeInTheDocument();
  });

  // **なぜその配当になったのかが最後まで分からなかった** (#5780)。
  it('ショーダウンで役名とベスト5枚を出す', async () => {
    const best = [card(10), card(11), card(12), card(13), card(1)];
    mockApi.mockResolvedValue(
      withState({
        phase: CincinnatiPhase.SHOWDOWN,
        isHumanTurn: false,
        seats: [
          seat({ isTurn: false, wonAmount: 80, handRank: 9, bestHand: best }),
          seat({ name: 'CPU1', isHuman: false, isTurn: false, handRank: 2, bestHand: best }),
        ],
      }),
    );
    renderWithProviders(<CincinnatiPage />);

    const mine = await screen.findByTestId('cin-showdown-0');
    expect(mine).toHaveTextContent('ロイヤルフラッシュ');
    expect(mine.querySelectorAll('img')).toHaveLength(5);
    expect(screen.getByTestId('cin-showdown-1')).toHaveTextContent('ツーペア');
  });

  // **負のコントロール: ショーダウン前は出さない。**
  it('ショーダウン前は役を出さない', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<CincinnatiPage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalled());
    expect(screen.queryByTestId('cin-showdown-0')).not.toBeInTheDocument();
  });
  it('ショーダウンで獲得額と次のハンドを出す', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: CincinnatiPhase.SHOWDOWN,
        isHumanTurn: false,
        seats: [seat({ isTurn: false, wonAmount: 80 }), seat({ name: 'CPU1', isHuman: false, isTurn: false })],
      }),
    );
    renderWithProviders(<CincinnatiPage />);
    await waitFor(() => expect(screen.getByTestId('cin-won-0')).toHaveTextContent('80'));

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のハンドへ' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('next'));
  });

  it('終局で勝者を出し、次のハンドを出さない', async () => {
    mockApi.mockResolvedValue(
      withState({ phase: CincinnatiPhase.GAME_END, gameEndFlag: true, isHumanTurn: false, winnerSeat: 1 }),
    );
    renderWithProviders(<CincinnatiPage />);
    await waitFor(() => expect(screen.getByTestId('cin-winner')).toHaveTextContent('CPU1'));
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
    renderWithProviders(<CincinnatiPage />);
    await waitFor(() => expect(screen.getByTestId('cin-chips')).toHaveTextContent('870'));
    expect(screen.getByTestId('cin-hand-line')).toHaveTextContent('3');
    expect(screen.getByTestId('cin-hand-line')).toHaveTextContent('120');
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
    renderWithProviders(<CincinnatiPage />);
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
    expect(screen.queryByTestId('cin-check')).not.toBeInTheDocument();
  });
});
