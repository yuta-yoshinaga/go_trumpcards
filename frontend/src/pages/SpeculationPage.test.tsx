import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { speculationApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, SpeculationResponse } from '../types/card';
import { SpeculationPhase } from '../types/phases';
import { SpeculationPage } from './SpeculationPage';

vi.mock('../api/gameApi', () => ({
  speculationApi: { exec: vi.fn() },
  actionLogApi: { speculation: vi.fn() },
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

const mockApi = vi.mocked(speculationApi.exec);
const mockUseCliMode = vi.mocked(useCliMode);

const card = (design: string, value: number): Card => ({ design, value }) as Card;

const seat = (name: string, chips: number, hiddenCount: number, best?: Card) => ({ name, chips, hiddenCount, best });

const base: SpeculationResponse = {
  phase: SpeculationPhase.FLIP,
  seats: [seat('You', 190, 3), seat('CPU1', 190, 3), seat('CPU2', 190, 2)],
  trumpSuit: 3,
  trumpCard: card('HEART', 9),
  pot: 30,
  turnSeat: 0,
  bestSeat: -1,
  offerFrom: -1,
  offerTo: -1,
  offerAmount: 0,
  roundNo: 0,
  winnerSeat: -1,
  gameEndFlag: false,
  config: { players: 3, initialChips: 200, stake: 10, rounds: 5 },
  message: '',
};

const withState = (over: Partial<SpeculationResponse>): SpeculationResponse => ({ ...base, ...over });

/** Auction where a CPU bids for the human's card (human sells). */
const sellingState = withState({
  phase: SpeculationPhase.AUCTION,
  offerFrom: 2,
  offerTo: 0,
  offerAmount: 25,
  bestSeat: 0,
  seats: [seat('You', 190, 2, card('HEART', 12)), seat('CPU1', 190, 3), seat('CPU2', 190, 2)],
});

/** Auction where a CPU asks a price of the human (human buys). */
const buyingState = withState({
  phase: SpeculationPhase.AUCTION,
  offerFrom: 0,
  offerTo: 2,
  offerAmount: 25,
  bestSeat: 2,
  seats: [seat('You', 190, 2), seat('CPU1', 190, 3), seat('CPU2', 190, 2, card('HEART', 12))],
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

describe('SpeculationPage', () => {
  it('マウント時に reset を呼ぶ', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<SpeculationPage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('全席のチップを出す', async () => {
    mockApi.mockResolvedValue(withState({ seats: [seat('You', 190, 3), seat('CPU1', 150, 3), seat('CPU2', 240, 2)] }));
    renderWithProviders(<SpeculationPage />);
    await waitFor(() => expect(screen.getByTestId('sp-seat-chips-0')).toBeInTheDocument());
    expect(screen.getByTestId('sp-seat-chips-1')).toHaveTextContent('150');
    expect(screen.getByTestId('sp-seat-chips-2')).toHaveTextContent('240');
  });

  // **伏せ札の中身はこのゲームの秘密そのもの。** 出るのは枚数だけで、
  // サーバもカードを送ってこない。
  it('伏せ札は枚数だけを出す', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<SpeculationPage />);
    await waitFor(() => expect(screen.getByTestId('sp-hidden-0')).toBeInTheDocument());
    expect(screen.getByTestId('sp-hidden-0')).toHaveTextContent('3');
    expect(screen.getByTestId('sp-hidden-2')).toHaveTextContent('2');
    // 誰の伏せ札も「最高札」として表には出ていない。
    expect(screen.queryByTestId('sp-best-0')).not.toBeInTheDocument();
    expect(screen.queryByTestId('sp-best-1')).not.toBeInTheDocument();
  });

  it('切り札スートを出す', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<SpeculationPage />);
    await waitFor(() => expect(screen.getByTestId('sp-trump')).toHaveTextContent('ハート'));
  });

  it('ラウンドとポットを出す', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<SpeculationPage />);
    // roundNo は消化済み数なので、表示は 1 / 5。
    await waitFor(() => expect(screen.getByTestId('sp-round-line')).toHaveTextContent('1 / 5'));
    expect(screen.getByTestId('sp-round-line')).toHaveTextContent('30');
  });

  // **「最高札の持ち主がいない」は -1。** 0 を「いない」と読むと、まだ切り札が
  // 1 枚も出ていない盤面で人間の席に最高札の印が付く。
  it('bestSeat が -1 なら座席0に最高札の印を付けない', async () => {
    mockApi.mockResolvedValue(withState({ bestSeat: -1 }));
    renderWithProviders(<SpeculationPage />);
    await waitFor(() => expect(screen.getByTestId('sp-seat-0')).toBeInTheDocument());
    expect(screen.getByTestId('sp-no-best')).toBeInTheDocument();
    expect(screen.getByTestId('sp-seat-0').className).not.toContain('ring-ds-success');
  });

  it('bestSeat の席に最高札を出す', async () => {
    mockApi.mockResolvedValue(buyingState);
    renderWithProviders(<SpeculationPage />);
    await waitFor(() => expect(screen.getByTestId('sp-best-2')).toBeInTheDocument());
    expect(screen.getByTestId('sp-seat-2').className).toContain('ring-ds-success');
    expect(screen.queryByTestId('sp-no-best')).not.toBeInTheDocument();
  });

  it('手番の席を示す', async () => {
    mockApi.mockResolvedValue(withState({ turnSeat: 1 }));
    renderWithProviders(<SpeculationPage />);
    await waitFor(() => expect(screen.getByTestId('sp-turn-guide')).toBeInTheDocument());
    expect(screen.getByTestId('sp-turn-guide')).toHaveTextContent('CPU1');
    expect(screen.getByTestId('sp-seat-1').className).toContain('ring-ds-accent');
  });

  it('めくるボタンが flip を送る', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<SpeculationPage />);
    await waitFor(() => expect(screen.getByTestId('sp-flip')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('sp-flip'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('flip'));
  });

  it('めくりフェーズでは競りの操作を出さない', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<SpeculationPage />);
    await waitFor(() => expect(screen.getByTestId('sp-flip')).toBeInTheDocument());
    expect(screen.queryByTestId('sp-offer')).not.toBeInTheDocument();
    expect(screen.queryByTestId('sp-accept')).not.toBeInTheDocument();
  });

  it('売り手のときは買い手の名前と提示額を出す', async () => {
    mockApi.mockResolvedValue(sellingState);
    renderWithProviders(<SpeculationPage />);
    await waitFor(() => expect(screen.getByTestId('sp-offer')).toBeInTheDocument());
    expect(screen.getByTestId('sp-offer')).toHaveTextContent('CPU2');
    expect(screen.getByTestId('sp-offer')).toHaveTextContent('25');
  });

  it('受けるボタンが accept を送る', async () => {
    mockApi.mockResolvedValue(sellingState);
    renderWithProviders(<SpeculationPage />);
    await waitFor(() => expect(screen.getByTestId('sp-accept')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('sp-accept'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('accept'));
  });

  it('断るボタンが decline を送る', async () => {
    mockApi.mockResolvedValue(sellingState);
    renderWithProviders(<SpeculationPage />);
    await waitFor(() => expect(screen.getByTestId('sp-decline')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('sp-decline'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('decline'));
  });

  // **上乗せできるのは人間が買い手のときだけ。** 売り手の側から値を吊り上げる
  // 手はドメインの Bid に無く、サーバは offerFrom != 0 の bid を弾く。
  it('売り手のときは上乗せの操作を出さない', async () => {
    mockApi.mockResolvedValue(sellingState);
    renderWithProviders(<SpeculationPage />);
    await waitFor(() => expect(screen.getByTestId('sp-offer')).toBeInTheDocument());
    expect(screen.queryByTestId('sp-raise')).not.toBeInTheDocument();
    expect(screen.queryByTestId('sp-bid')).not.toBeInTheDocument();
  });

  it('買い手のときは上乗せの操作を出す', async () => {
    mockApi.mockResolvedValue(buyingState);
    renderWithProviders(<SpeculationPage />);
    await waitFor(() => expect(screen.getByTestId('sp-raise')).toBeInTheDocument());
  });

  // **上乗せは提示額を超えていなければ成立しない。** ドメインは
  // `amount <= offerAmount` を errSpeculationBadAmount で弾く。
  it('上乗せは提示額より大きい額で bid を送る', async () => {
    mockApi.mockResolvedValue(buyingState);
    renderWithProviders(<SpeculationPage />);
    await waitFor(() => expect(screen.getByTestId('sp-bid')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('sp-bid'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bid', expect.objectContaining({ amount: 26 })));

    const call = mockApi.mock.calls.find(([cmd]) => cmd === 'bid');
    if (!call) throw new Error('bid was never sent');
    const amount = (call[1] as { amount: number }).amount;
    expect(amount).toBeGreaterThan(buyingState.offerAmount);
  });

  it('入力した上乗せ額をそのまま送る', async () => {
    mockApi.mockResolvedValue(buyingState);
    renderWithProviders(<SpeculationPage />);
    await waitFor(() => expect(screen.getByTestId('sp-bid')).toBeInTheDocument());
    fireEvent.change(screen.getByLabelText('上乗せ額'), { target: { value: '40' } });
    fireEvent.click(screen.getByTestId('sp-bid'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bid', { amount: 40 }));
  });

  it('決着で勝った席を出し、次のラウンドへ送れる', async () => {
    mockApi.mockResolvedValue(withState({ phase: SpeculationPhase.RESULT, winnerSeat: 2 }));
    renderWithProviders(<SpeculationPage />);
    await waitFor(() => expect(screen.getByTestId('sp-result')).toHaveTextContent('CPU2'));
    fireEvent.click(screen.getByTestId('sp-next'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('next'));
  });

  // **流局は「誰かが負けた」ではない。** 勝者 -1 を席 0 と読むと、切り札が
  // 1 枚も出なかったラウンドを人間の勝ちとして表示する。
  it('winnerSeat が -1 なら流局として出す', async () => {
    mockApi.mockResolvedValue(withState({ phase: SpeculationPhase.RESULT, winnerSeat: -1 }));
    renderWithProviders(<SpeculationPage />);
    await waitFor(() => expect(screen.getByTestId('sp-result')).toBeInTheDocument());
    expect(screen.getByTestId('sp-result')).toHaveTextContent('切り札が1枚も出ませんでした');
    expect(screen.getByTestId('sp-result')).not.toHaveTextContent('あなたがポットを取りました');
  });

  it('ゲーム終了で最終チップを出す', async () => {
    mockApi.mockResolvedValue(
      withState({ phase: SpeculationPhase.GAME_END, gameEndFlag: true, winnerSeat: 0, seats: [seat('You', 320, 0)] }),
    );
    renderWithProviders(<SpeculationPage />);
    await waitFor(() => expect(screen.getByTestId('sp-final-chips')).toHaveTextContent('320'));
    expect(screen.queryByTestId('sp-next')).not.toBeInTheDocument();
  });
});
