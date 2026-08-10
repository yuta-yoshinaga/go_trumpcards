import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tysiacApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeTysiacState } from '../test/stateFactories';
import { TysiacPage } from './TysiacPage';

vi.mock('../api/gameApi', () => ({
  tysiacApi: { exec: vi.fn() },
  actionLogApi: { tysiac: vi.fn() },
}));

const mockExec = vi.mocked(tysiacApi.exec);

const playPhaseState = makeTysiacState();
const bidPhaseState = makeTysiacState({
  phase: 0,
  trumpSuit: 0,
  currentPlayerIdx: 0,
  isHumanTurn: true,
  currentBid: 100,
});
const talonPhaseState = makeTysiacState({ phase: 1, declarerIdx: 0 });
const trickEndState = makeTysiacState({
  phase: 3,
  currentTrick: [
    { playerIdx: 0, card: { design: 'HEART', value: 12 } },
    { playerIdx: 1, card: { design: 'CLOVER', value: 13 } },
  ],
});
const roundEndState = makeTysiacState({
  phase: 4,
  roundCardPoints: [55, 35, 30],
  roundMarriage: [100, 0, 0],
});
const gameEndState = makeTysiacState({
  phase: 5,
  gameEndFlag: true,
  winnerPlayer: 0,
  message: 'ゲーム終了！ あなたの勝ちです！',
});
const cpuTurnState = makeTysiacState({ currentPlayerIdx: 1, isHumanTurn: false });

beforeEach(() => {
  localStorage.clear();
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('TysiacPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<TysiacPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<TysiacPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetPoints: 1000 },
      }),
    );
  });

  it('renders the play phase with the human cards and the Declarer badge', async () => {
    renderWithProviders(<TysiacPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
    // The human (seat 0) is the default Declarer.
    expect(screen.getByText('デクレアラー')).toBeInTheDocument();
  });

  it('shows a marriage banner during play (trump-suit K-Q ♥ scores +100)', async () => {
    // Default hand: ♥K + ♥Q with trump ♥ → a +100 marriage.
    renderWithProviders(<TysiacPage />);
    const banner = await screen.findByTestId('tysiac-marriage');
    expect(banner).toHaveTextContent('マリッジ可能');
    expect(banner).toHaveTextContent('♥ K-Q (+100)');
  });

  it('renders the bid phase with raise and pass buttons', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<TysiacPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'レイズ（110）' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'パス' })).toBeInTheDocument();
  });

  it('shows the next bid amount on the raise button and the current bid near the controls', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<TysiacPage />);
    // Raise button label reflects currentBid (100) + step (10).
    await waitFor(() => expect(screen.getByTestId('tysiac-bid-raise')).toHaveTextContent('レイズ（110）'));
    // Current highest bid is shown right by the bid controls.
    expect(screen.getByTestId('tysiac-current-bid')).toHaveTextContent('入札: 100');
  });

  it('reflects a higher current bid in the raise-button next amount', async () => {
    mockExec.mockResolvedValue(makeTysiacState({ phase: 0, currentPlayerIdx: 0, isHumanTurn: true, currentBid: 130 }));
    renderWithProviders(<TysiacPage />);
    await waitFor(() => expect(screen.getByTestId('tysiac-bid-raise')).toHaveTextContent('レイズ（140）'));
    expect(screen.getByTestId('tysiac-current-bid')).toHaveTextContent('入札: 130');
  });

  it('hides the bid display once bidding is over (talon phase)', async () => {
    mockExec.mockResolvedValue(talonPhaseState);
    renderWithProviders(<TysiacPage />);
    await waitFor(() => expect(screen.getByTestId('tysiac-talon-prompt')).toBeInTheDocument());
    expect(screen.queryByTestId('tysiac-current-bid')).not.toBeInTheDocument();
    expect(screen.queryByTestId('tysiac-bid-raise')).not.toBeInTheDocument();
  });

  it('raising the bid dispatches bid with raise=true', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<TysiacPage />);
    const raiseBtn = await screen.findByRole('button', { name: 'レイズ（110）' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(raiseBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { raise: true }));
  });

  it('renders the talon phase with the give-card button and prompt', async () => {
    mockExec.mockResolvedValue(talonPhaseState);
    renderWithProviders(<TysiacPage />);
    await waitFor(() => expect(screen.getByTestId('tysiac-talon-prompt')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'カードを渡す' })).toBeInTheDocument();
  });

  it('selecting a card then discarding dispatches discard', async () => {
    mockExec.mockResolvedValue(talonPhaseState);
    renderWithProviders(<TysiacPage />);
    const card = await screen.findByAltText('♥ Q');
    fireEvent.click(card);
    const giveBtn = await screen.findByRole('button', { name: 'カードを渡す' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(talonPhaseState);
    fireEvent.click(giveBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', { cardIndex: 0 }));
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<TysiacPage />);
    const card = await screen.findByAltText('♥ Q');
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('renders trick end with the next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<TysiacPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with the next round button and the round result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<TysiacPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<TysiacPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ちです！')).toBeInTheDocument());
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<TysiacPage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('renders a progress bar toward the target for each player', async () => {
    const state = makeTysiacState({
      players: [
        { id: 0, isHuman: true, cardCount: 7, cards: [], trickCount: 0, score: 250, isDeclarer: false },
        { id: 1, isHuman: false, cardCount: 7, cards: [], trickCount: 0, score: 500, isDeclarer: false },
        { id: 2, isHuman: false, cardCount: 7, cards: [], trickCount: 0, score: 100, isDeclarer: false },
      ],
    });
    mockExec.mockResolvedValue(state);
    renderWithProviders(<TysiacPage />);
    const bar0 = await screen.findByTestId('tysiac-progress-0');
    expect(bar0).toHaveAttribute('role', 'progressbar');
    expect(bar0).toHaveAttribute('aria-valuemax', '1000');
    expect(bar0).toHaveAttribute('aria-valuenow', '250');
    expect(screen.getByTestId('tysiac-progress-1')).toHaveAttribute('aria-valuenow', '500');
    expect(screen.getByTestId('tysiac-progress-2')).toHaveAttribute('aria-valuenow', '100');
    // Below 80%: accent (non-warning) fill.
    expect(bar0.querySelector('.bg-ds-accent')).not.toBeNull();
    expect(bar0.querySelector('.bg-ds-warning')).toBeNull();
  });

  it('turns the bar to the warning color once a player passes 80% of the target', async () => {
    const state = makeTysiacState({
      players: [
        { id: 0, isHuman: true, cardCount: 7, cards: [], trickCount: 0, score: 850, isDeclarer: false },
        { id: 1, isHuman: false, cardCount: 7, cards: [], trickCount: 0, score: 400, isDeclarer: false },
        { id: 2, isHuman: false, cardCount: 7, cards: [], trickCount: 0, score: 800, isDeclarer: false },
      ],
    });
    mockExec.mockResolvedValue(state);
    renderWithProviders(<TysiacPage />);
    const bar0 = await screen.findByTestId('tysiac-progress-0');
    // 850/1000 = 85% > 80% → warning fill.
    expect(bar0.querySelector('.bg-ds-warning')).not.toBeNull();
    // Exactly 80% is not "over" 80% → still accent.
    expect(screen.getByTestId('tysiac-progress-2').querySelector('.bg-ds-warning')).toBeNull();
  });

  it('overlays the contract-forecast marker on the Declarer bar only', async () => {
    const state = makeTysiacState({
      players: [
        { id: 0, isHuman: true, cardCount: 7, cards: [], trickCount: 0, score: 300, isDeclarer: true },
        { id: 1, isHuman: false, cardCount: 7, cards: [], trickCount: 0, score: 200, isDeclarer: false },
        { id: 2, isHuman: false, cardCount: 7, cards: [], trickCount: 0, score: 100, isDeclarer: false },
      ],
      contract: 120,
    });
    mockExec.mockResolvedValue(state);
    renderWithProviders(<TysiacPage />);
    const marker = await screen.findByTestId('tysiac-forecast-0');
    // (300 + 120) / 1000 = 42%.
    expect(marker).toHaveStyle({ left: '42%' });
    // Non-declarers get no forecast marker.
    expect(screen.queryByTestId('tysiac-forecast-1')).not.toBeInTheDocument();
    expect(screen.queryByTestId('tysiac-forecast-2')).not.toBeInTheDocument();
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, hint: { cardIndices: [0], reason: 'x' } });
    renderWithProviders(<TysiacPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    // バナーは推奨札の位置を `([0])` の形で含む。トグルのラベル (「ヒント表示」)
    // と紛れないよう、そこで判定する。
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });

  // **押したときは出る。**押していない側だけを見ていると、`isRequestedHint` を
  // 定数 false にしても通ってしまう。真の分岐も踏んでおく。
  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue({
      ...playPhaseState,
      hint: { cardIndices: [0], reason: 'x' },
      messageCode: 'tysiac.hintRequested',
    });
    renderWithProviders(<TysiacPage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });

  it('says a marriage only scores when the K or Q is led', async () => {
    // Holding the pair is not enough — the CUI has always said so, the web page
    // only listed the pairs.
    renderWithProviders(<TysiacPage />);
    const banner = await screen.findByTestId('tysiac-marriage');
    expect(banner.textContent).toMatch(/リード/);
  });
});
