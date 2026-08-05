import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { scartoApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeScartoState } from '../test/stateFactories';
import { ScartoPage } from './ScartoPage';

vi.mock('../api/gameApi', () => ({
  scartoApi: { exec: vi.fn() },
  actionLogApi: { scarto: vi.fn() },
}));

const mockExec = vi.mocked(scartoApi.exec);

const suit = (value: number, design: 'HEART' | 'SPADE' | 'CLOVER' | 'DIAMOND', glyph: string, label: string) => ({
  design,
  value,
  glyph,
  label,
  color: design === 'HEART' || design === 'DIAMOND' ? 'red' : 'black',
  deck: 'tarot',
});

const playPhaseState = makeScartoState();

// Human is the dealer performing the scarto. The hand mixes buryable pips (0–2)
// with a King (index 3) and the Excuse (index 4), which must NOT be buryable.
const scartoPhaseState = makeScartoState({
  phase: 0,
  isHumanTurn: false,
  isHumanScarto: true,
  scartoCount: 0,
  playableIndices: [],
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 28,
      cards: [
        suit(2, 'HEART', '♥', '2'),
        suit(3, 'HEART', '♥', '3'),
        suit(4, 'HEART', '♥', '4'),
        suit(14, 'SPADE', '♠', 'R'),
        { design: 'JOKER' as const, value: 0, glyph: '★', label: 'Excuse', color: 'gold', deck: 'tarot' },
      ],
      trickCount: 0,
      cardPoints: 0,
      score: 0,
      isDealer: true,
    },
    { id: 1, isHuman: false, cardCount: 25, cards: [], trickCount: 0, cardPoints: 0, score: 0, isDealer: false },
    { id: 2, isHuman: false, cardCount: 25, cards: [], trickCount: 0, cardPoints: 0, score: 0, isDealer: false },
  ],
});

// Scarto phase but a CPU is the dealer: the human waits.
const scartoWaitingState = makeScartoState({
  phase: 0,
  isHumanTurn: false,
  isHumanScarto: false,
  scartoCount: 0,
  playableIndices: [],
});

const trickEndState = makeScartoState({
  phase: 2,
  isHumanTurn: false,
  currentTrick: [
    { playerIdx: 0, card: suit(12, 'HEART', '♥', 'C') },
    { playerIdx: 1, card: suit(13, 'CLOVER', '♣', 'D') },
  ],
});

const roundEndState = makeScartoState({
  phase: 3,
  isHumanTurn: false,
  outcome: 1,
  dealScores: [6, -2, -4],
});

// Round end with captured card-points that make the average-difference settlement
// meaningful: totals 70+60+52 = 182, mean ≈ 60.7, so dealScores = 3·points − 182.
const settlementState = makeScartoState({
  phase: 3,
  isHumanTurn: false,
  outcome: 1,
  dealScores: [28, -2, -34],
  players: [
    { id: 0, isHuman: true, cardCount: 0, cards: [], trickCount: 5, cardPoints: 70, score: 28, isDealer: false },
    { id: 1, isHuman: false, cardCount: 0, cards: [], trickCount: 4, cardPoints: 60, score: -2, isDealer: false },
    { id: 2, isHuman: false, cardCount: 0, cards: [], trickCount: 3, cardPoints: 52, score: -34, isDealer: true },
  ],
});

const gameEndState = makeScartoState({
  phase: 4,
  isHumanTurn: false,
  gameEndFlag: true,
  winnerPlayer: 0,
  message: 'ゲーム終了！ あなたの勝ち！',
});

const drawState = makeScartoState({
  phase: 4,
  isHumanTurn: false,
  gameEndFlag: true,
  winnerPlayer: -1,
  message: 'ゲーム終了！ 引き分け！',
});

const cpuTurnState = makeScartoState({ currentPlayerIdx: 1, isHumanTurn: false });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('ScartoPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<ScartoPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<ScartoPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetDeals: 5 },
      }),
    );
  });

  it('renders the play phase with the human cards and a dealer badge', async () => {
    renderWithProviders(<ScartoPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'D ♥' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '21 ✦' })).toBeInTheDocument();
    });
    expect(screen.getAllByText('親（ディーラー）').length).toBeGreaterThan(0);
  });

  it('renders the scarto phase discard prompt and buries exactly 3 pip cards', async () => {
    mockExec.mockResolvedValue(scartoPhaseState);
    renderWithProviders(<ScartoPage />);
    await screen.findByTestId('scarto-discard-prompt');
    // Select the three buryable heart pips (indices 0–2).
    for (const label of ['2 ♥', '3 ♥', '4 ♥']) {
      fireEvent.click(screen.getByRole('button', { name: label }));
    }
    const discardBtn = screen.getByRole('button', { name: /捨てる/ });
    expect(discardBtn).toBeEnabled();
    mockExec.mockClear();
    mockExec.mockResolvedValue(scartoPhaseState);
    fireEvent.click(discardBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('scarto', { cardIndices: [0, 1, 2] }));
  });

  it('marks counting cards (King / Excuse) as non-buryable in the scarto phase', async () => {
    mockExec.mockResolvedValue(scartoPhaseState);
    renderWithProviders(<ScartoPage />);
    await screen.findByTestId('scarto-discard-prompt');
    expect(screen.getByRole('button', { name: 'R ♠' })).toHaveAttribute('aria-disabled', 'true');
    expect(screen.getByRole('button', { name: 'Excuse ★' })).toHaveAttribute('aria-disabled', 'true');
  });

  it('surfaces distinct un-buriable reasons for the King and the Excuse in the scarto phase', async () => {
    mockExec.mockResolvedValue(scartoPhaseState);
    renderWithProviders(<ScartoPage />);
    await screen.findByTestId('scarto-discard-prompt');
    // The King (counting card) exposes the court-specific reason on its tooltip.
    const kingBtn = screen.getByRole('button', { name: 'R ♠' });
    expect(kingBtn).toHaveAttribute('title', '得点札（K・コート札）は捨てられません。');
    // The Excuse exposes a different, excuse-specific reason.
    const excuseBtn = screen.getByRole('button', { name: 'Excuse ★' });
    expect(excuseBtn).toHaveAttribute('title', 'エクスキューズ（マット）は捨てられません。');
    // The two reasons differ.
    expect(kingBtn.getAttribute('title')).not.toBe(excuseBtn.getAttribute('title'));
  });

  it('shows no un-buriable tooltip on a freely buriable low pip in the scarto phase', async () => {
    mockExec.mockResolvedValue(scartoPhaseState);
    renderWithProviders(<ScartoPage />);
    await screen.findByTestId('scarto-discard-prompt');
    expect(screen.getByRole('button', { name: '2 ♥' })).not.toHaveAttribute('title');
  });

  it('keeps the bury button disabled until exactly 3 cards are chosen', async () => {
    mockExec.mockResolvedValue(scartoPhaseState);
    renderWithProviders(<ScartoPage />);
    await screen.findByTestId('scarto-discard-prompt');
    fireEvent.click(screen.getByRole('button', { name: '2 ♥' }));
    expect(screen.getByRole('button', { name: /捨てる/ })).toBeDisabled();
  });

  it('shows a waiting state when a CPU is the dealer during the scarto', async () => {
    mockExec.mockResolvedValue(scartoWaitingState);
    renderWithProviders(<ScartoPage />);
    await waitFor(() => expect(screen.getByTestId('scarto-waiting')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: /捨てる/ })).not.toBeInTheDocument();
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<ScartoPage />);
    const card = await screen.findByRole('button', { name: 'D ♥' });
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<ScartoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'D ♥' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('renders trick end with the next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<ScartoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with the next deal button and the deal settlement', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<ScartoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のディール' })).toBeInTheDocument());
    expect(screen.getByTestId('scarto-result')).toBeInTheDocument();
  });

  it('shows the average-difference settlement breakdown at round end', async () => {
    mockExec.mockResolvedValue(settlementState);
    renderWithProviders(<ScartoPage />);
    const breakdown = await screen.findByTestId('scarto-breakdown');
    // Table average of captured card-points (182 / 3 ≈ 60.7).
    expect(breakdown).toHaveTextContent('全体平均: 60.7点');
    // Each seat's earned card-points.
    expect(breakdown).toHaveTextContent('獲得 70点');
    expect(breakdown).toHaveTextContent('獲得 60点');
    expect(breakdown).toHaveTextContent('獲得 52点');
    // The displayed delta matches dealScores (existing per-player settlement line).
    expect(screen.getByTestId('scarto-result')).toHaveTextContent('+28');
  });

  // 上段の dealScores と内訳の平均差が N 倍で結び付くことを固定する (#4930)。
  it('spells out that the change is the average difference times the player count', async () => {
    mockExec.mockResolvedValue(settlementState);
    renderWithProviders(<ScartoPage />);
    const breakdown = await screen.findByTestId('scarto-breakdown');

    expect(screen.getByTestId('scarto-formula')).toHaveTextContent('平均差 × プレイヤー数（3人）');
    // 70 - 60.666… = +9.3、×3 で +28。上の行の dealScores と一致する。
    expect(breakdown).toHaveTextContent('平均差 +9.3');
    expect(breakdown).toHaveTextContent('変動 +28.0');
    // 負の側も出る。52 - 60.666… = -8.7、×3 で -26。
    expect(breakdown).toHaveTextContent('平均差 -8.7');
  });

  it('dispatches nextround from round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<ScartoPage />);
    const btn = await screen.findByRole('button', { name: '次のディール' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(roundEndState);
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<ScartoPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ち！')).toBeInTheDocument());
  });

  it('renders a draw without celebrating a false win', async () => {
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<ScartoPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ 引き分け！')).toBeInTheDocument());
  });

  it('the next-game button at game end resets immediately', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<ScartoPage />);
    const nextGame = await screen.findByRole('button', { name: '次のゲーム' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(gameEndState);
    fireEvent.click(nextGame);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', expect.anything()));
  });

  it('changing the CPU difficulty and target-deals selects updates the config', async () => {
    renderWithProviders(<ScartoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'D ♥' })).toBeInTheDocument());
    const difficulty = screen.getByLabelText('CPU難易度') as HTMLSelectElement;
    fireEvent.change(difficulty, { target: { value: '2' } });
    expect(difficulty.value).toBe('2');
    const deals = screen.getByLabelText('マッチのディール数') as HTMLSelectElement;
    fireEvent.change(deals, { target: { value: '3' } });
    expect(deals.value).toBe('3');
  });

  it('renders the backend hint banner with its card indices', async () => {
    mockExec.mockResolvedValue(
      makeScartoState({ hint: { cardIndices: [0, 2], reason: 'lead_low' }, messageCode: 'scarto.hintRequested' }),
    );
    renderWithProviders(<ScartoPage />);
    await waitFor(() => expect(screen.getByText(/\[0\], \[2\]/)).toBeInTheDocument());
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる
  // (#4605)。このテストは、それを正しいと固定していた旧テストの対。
  it('hides the hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue(makeScartoState({ hint: { cardIndices: [0, 2], reason: 'lead_low' } }));
    renderWithProviders(<ScartoPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByText(/\[0\], \[2\]/)).not.toBeInTheDocument();
  });
});
