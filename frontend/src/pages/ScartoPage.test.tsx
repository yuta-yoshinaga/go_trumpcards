import { fireEvent, screen, waitFor, within } from '@testing-library/react';
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
  // サーバが返す「捨てられる札」。ピップ 3 枚は足りているので切り札は入らない。
  discardableIndices: [0, 1, 2],
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

// 捨てられるピップが 1 枚しかない親の手。ドメインは非ブー切り札 (index 1, 2) を
// 解禁するので、サーバはそれを discardableIndices に載せて返す。
const shortOnPipsState = makeScartoState({
  phase: 0,
  isHumanTurn: false,
  isHumanScarto: true,
  scartoCount: 0,
  playableIndices: [],
  discardableIndices: [0, 1, 2],
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 28,
      cards: [
        suit(4, 'HEART', '♥', '4'),
        { design: 'JOKER' as const, value: 8, glyph: 'A', label: 'T8', color: 'purple', deck: 'tarot' },
        { design: 'JOKER' as const, value: 12, glyph: 'A', label: 'T12', color: 'purple', deck: 'tarot' },
        { design: 'JOKER' as const, value: 21, glyph: 'A', label: 'T21', color: 'purple', deck: 'tarot' },
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
  lastTrickWinner: 1,
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
  window.innerWidth = 375;
  window.dispatchEvent(new Event('resize'));
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

  it('renders CPU difficulty as a select and match length as a number input', async () => {
    renderWithProviders(<ScartoPage />);
    fireEvent.click(await screen.findByText('設定'));

    expect(screen.getByRole('combobox', { name: 'CPU難易度' })).toBeInTheDocument();
    expect(screen.getByRole('spinbutton', { name: 'マッチのディール数' })).toBeInTheDocument();
  });

  it('renders the play phase with the human cards and a dealer badge', async () => {
    renderWithProviders(<ScartoPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'D ♥' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '21 ✦' })).toBeInTheDocument();
    });
    expect(screen.getAllByText('親（ディーラー）').length).toBeGreaterThan(0);
  });

  it('shows the card point reference during play without revealing CPU hands', async () => {
    renderWithProviders(<ScartoPage />);
    fireEvent.click(await screen.findByText('ルール'));
    const reference = await screen.findByTestId('scarto-point-reference');
    expect(reference).toHaveTextContent('スート札のキング: 4.5点');
    expect(reference).toHaveTextContent('スート札のクイーン: 3.5点');
    expect(reference).toHaveTextContent('スート札のナイト: 2.5点');
    expect(reference).toHaveTextContent('スート札のジャック: 1.5点');
    expect(reference).toHaveTextContent('ブー（切り札の1・21とエクスキューズ）: 各4.5点');
    expect(reference).toHaveTextContent('その他の札: 各0.5点');
    fireEvent.click(screen.getByText('CPU対戦相手 (2)'));
    expect(screen.getByText(/CPU 1: 25枚/)).toBeInTheDocument();
    expect(screen.getByText(/CPU 2: 25枚/)).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /CPU 1/ })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /CPU 2/ })).not.toBeInTheDocument();
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
    expect(within(screen.getByTestId('game-footer-actions')).getByRole('button', { name: /捨てる/ })).toBe(discardBtn);
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
    expect(within(screen.getByTestId('game-footer-actions')).getByRole('button', { name: '出す' })).toBe(playBtn);
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
    expect(screen.getByTestId('trick-winner-badge')).toHaveTextContent('CPU 1 が獲得');
  });

  it('does not render a winner badge while a trick is in progress', async () => {
    mockExec.mockResolvedValue(
      makeScartoState({
        phase: 1,
        lastTrickWinner: 1,
        currentTrick: [{ playerIdx: 1, card: { design: 'HEART', value: 12 } }],
      }),
    );
    renderWithProviders(<ScartoPage />);
    await waitFor(() => expect(screen.getByTestId('trick-display-cards')).toBeInTheDocument());
    expect(screen.queryByTestId('trick-winner-badge')).not.toBeInTheDocument();
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

  it('shows clearly provisional average differences during play with decimal averages', async () => {
    mockExec.mockResolvedValue(
      makeScartoState({
        phase: 1,
        players: [
          { id: 0, isHuman: true, cardCount: 20, cards: [], trickCount: 1, cardPoints: 1, score: 0, isDealer: false },
          { id: 1, isHuman: false, cardCount: 20, cards: [], trickCount: 0, cardPoints: 0, score: 0, isDealer: false },
          { id: 2, isHuman: false, cardCount: 20, cards: [], trickCount: 0, cardPoints: 0, score: 0, isDealer: true },
        ],
      }),
    );
    renderWithProviders(<ScartoPage />);
    const provisional = await screen.findByTestId('scarto-provisional');
    expect(provisional).toHaveTextContent('暫定（ディール進行中）');
    expect(screen.getByTestId('scarto-provisional-breakdown')).toHaveTextContent('全体平均: 0.3点');
    expect(screen.getByTestId('scarto-provisional-breakdown')).toHaveTextContent('平均差 +0.7');
    expect(screen.getByTestId('scarto-provisional-breakdown')).toHaveTextContent('平均より上');
    expect(screen.getByTestId('scarto-provisional-breakdown')).toHaveTextContent('平均より下');
    expect(screen.queryByTestId('scarto-result')).not.toBeInTheDocument();
  });

  it('labels players exactly at the average as tied during play', async () => {
    mockExec.mockResolvedValue(
      makeScartoState({
        phase: 1,
        players: [
          { id: 0, isHuman: true, cardCount: 20, cards: [], trickCount: 0, cardPoints: 4, score: 0, isDealer: false },
          { id: 1, isHuman: false, cardCount: 20, cards: [], trickCount: 0, cardPoints: 4, score: 0, isDealer: false },
          { id: 2, isHuman: false, cardCount: 20, cards: [], trickCount: 0, cardPoints: 4, score: 0, isDealer: true },
        ],
      }),
    );
    renderWithProviders(<ScartoPage />);
    const breakdown = await screen.findByTestId('scarto-provisional-breakdown');

    expect(breakdown).toHaveTextContent('平均と同点');
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

  it('accepts target deal counts from 1 through 100', async () => {
    renderWithProviders(<ScartoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'D ♥' })).toBeInTheDocument());
    const difficulty = screen.getByLabelText('CPU難易度') as HTMLSelectElement;
    fireEvent.change(difficulty, { target: { value: '2' } });
    expect(difficulty.value).toBe('2');
    const deals = screen.getByLabelText('マッチのディール数') as HTMLInputElement;
    expect(deals).toHaveAttribute('min', '1');
    expect(deals).toHaveAttribute('max', '100');
    expect(screen.getByText('1〜100の整数を指定してください。')).toBeInTheDocument();
    fireEvent.change(deals, { target: { value: '100' } });
    expect(deals.value).toBe('100');
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

  // **#6236 の本体。** 捨てられるピップが 3 枚に満たない手では、ドメインは
  // 非ブー切り札を捨てることを許す。以前は画面側が色だけを見て切り札を常に
  // 除外していたので、**この手を引いた親は枚数を揃えられず先へ進めなかった**。
  it('lets the dealer bury a trump when there are too few pips', async () => {
    mockExec.mockResolvedValue(shortOnPipsState);
    renderWithProviders(<ScartoPage />);

    const cards = await screen.findAllByRole('button', { name: /♥|♠|★|A|T/ });
    // 切り札 (index 1, 2) が選べること。
    fireEvent.click(cards[0]);
    fireEvent.click(cards[1]);
    fireEvent.click(cards[2]);
    const button = screen.getByRole('button', { name: /捨てる/ });
    expect(button).toBeEnabled();
    fireEvent.click(button);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('scarto', { cardIndices: [0, 1, 2] }));
  });

  // ブーとエクスキューズは、ピップが足りなくても捨てられない。
  it('still refuses the bouts and the Excuse when pips run short', async () => {
    mockExec.mockResolvedValue(shortOnPipsState);
    renderWithProviders(<ScartoPage />);
    const cards = await screen.findAllByRole('button', { name: /♥|♠|★|A|T/ });
    // index 3 はブー (T21), index 4 はエクスキューズ。どちらも選択に加わらない。
    fireEvent.click(cards[3]);
    fireEvent.click(cards[4]);
    const button = screen.getByRole('button', { name: /捨てる/ });
    expect(button).toBeDisabled();
  });
  // **エクスキューズだけが勝敗の外にいる。**出した札が誰の手にも渡らず自分の得点山に
  // 戻るという規則は実装されているのに説明が無く、cardPoints の内訳が読み解けなかった
  // (#6514)。説明は「ルール」details に置く。
  it('keeps the excuse rule available inside the rules disclosure', async () => {
    renderWithProviders(<ScartoPage />);
    await screen.findByTestId('scarto-provisional');
    const rules = screen.getByText('ルール').closest('details');
    expect(rules).not.toHaveAttribute('open');
    expect(screen.getByTestId('scarto-rules')).not.toBeVisible();
    fireEvent.click(screen.getByText('ルール'));
    const note = await screen.findByTestId('scarto-excuse-note');
    // **キーではなく解決後の文言を見る** ── i18next は未知のキーをそのまま返す。
    expect(note).toHaveTextContent('エクスキューズは特別');
    expect(note.textContent).not.toContain('excuseReturnsNote');
  });

  it('keeps the CPU seat information collapsed on mobile', async () => {
    renderWithProviders(<ScartoPage />);
    await screen.findByTestId('scarto-provisional');
    const cpuAccordion = screen.getByTestId('cpu-accordion');
    const humanSeat = screen.getByTestId('scarto-human-seat');
    expect(cpuAccordion).not.toHaveAttribute('open');
    expect(humanSeat).toBeVisible();
    expect(humanSeat).toHaveAttribute('data-tutorial', 'scarto-human-seat');
    expect(humanSeat).toHaveAttribute('aria-label', 'あなた: 25枚 | 0トリック | 0点');
    expect(cpuAccordion).not.toContainElement(humanSeat);
    expect(humanSeat).toHaveTextContent('あなた: 25枚 | 0トリック | 0点');
    expect(within(cpuAccordion).queryByText(/CPU 1: 25枚/)).not.toBeVisible();
    expect(within(cpuAccordion).queryByText(/あなた:/)).not.toBeInTheDocument();
    fireEvent.click(within(cpuAccordion).getByText('CPU対戦相手 (2)'));
    expect(within(cpuAccordion).getByText(/CPU 1: 25枚/)).toBeVisible();
  });

  it('opens the rules and CPU information on desktop', async () => {
    window.innerWidth = 1280;
    window.dispatchEvent(new Event('resize'));
    renderWithProviders(<ScartoPage />);
    await screen.findByTestId('scarto-point-reference');
    expect(screen.getByText('ルール').closest('details')).toHaveAttribute('open');
    expect(screen.getByTestId('cpu-accordion')).toHaveAttribute('open');
  });

  // **催促は常設のライブ領域の中にある (#6880)。** フェーズ切り替えで現れる
  // テキストなので、領域が無いとスクリーンリーダには何も届かない。領域を
  // 出現と同時に付けても読み上げられないため、常設にして中身だけ差し替える。
  it('announces the prompt from an always-mounted live region', async () => {
    mockExec.mockResolvedValue(scartoPhaseState);
    renderWithProviders(<ScartoPage />);

    const live = await screen.findByTestId('scarto-prompt-live');
    expect(live).toHaveAttribute('role', 'status');
    expect(live).toHaveAttribute('aria-live', 'polite');
    // 催促が**その領域の中**にあること。隣に置いただけの実装は属性の検査を通る。
    expect(live).toContainElement(await screen.findByTestId('scarto-discard-prompt'));
  });
});
