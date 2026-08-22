import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, threecardrummyApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, ThreeCardRummyResponse } from '../types/card';
import { ThreeCardRummyPage } from './ThreeCardRummyPage';

vi.mock('../api/gameApi', () => ({
  threecardrummyApi: { exec: vi.fn() },
  actionLogApi: { threecardrummy: vi.fn() },
}));

const mockExec = vi.mocked(threecardrummyApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

/** What the server sends for a card the player must not see yet. */
const masked = { design: '' as CardDesign, value: 0 };

/** ♠3 ♥4 ♦5 — mixed suits, no set, no run → 3+4+5 = 12 points. */
const plainHand = [card('SPADE', 3), card('HEART', 4), card('DIAMOND', 5)];
/** ♠7 ♥7 ♦7 — three of a rank is a meld, which scores 0: the best hand there is. */
const meldHand = [card('SPADE', 7), card('HEART', 7), card('DIAMOND', 7)];
/** ♣5 ♦2 ♥7 — the dealer's revealed hand, 5+2+7 = 14 points. */
const dealerHand = [card('CLOVER', 5), card('DIAMOND', 2), card('HEART', 7)];

const betPhaseState: ThreeCardRummyResponse = {
  playerHand: [],
  dealerHand: [],
  phase: 1,
  chips: 1000,
  anteBet: 0,
  lowBonusBet: 0,
  playBet: 0,
  result: 0,
  antePayout: 0,
  playPayout: 0,
  anteBonusPayout: 0,
  lowBonusPayout: 0,
  totalPayout: 0,
  dealerQualified: false,
  playerScore: 0,
  dealerScore: 0,
  message: '',
};

const actionPhaseState: ThreeCardRummyResponse = {
  ...betPhaseState,
  phase: 2,
  playerHand: plainHand,
  // The dealer's three cards arrive masked; they are only filled in at the end.
  dealerHand: [masked, masked, masked],
  anteBet: 100,
  chips: 900,
  playerScore: 12,
  dealerScore: 0,
};

/** Same action phase, but the player was dealt a meld — a total of 0. */
const actionPhaseMeldState: ThreeCardRummyResponse = {
  ...actionPhaseState,
  playerHand: meldHand,
  playerScore: 0,
};

const endPhasePlayerWins: ThreeCardRummyResponse = {
  ...betPhaseState,
  phase: 3,
  playerHand: plainHand,
  dealerHand,
  chips: 1400,
  anteBet: 100,
  playBet: 100,
  result: 1,
  antePayout: 200,
  playPayout: 200,
  totalPayout: 400,
  dealerQualified: true,
  playerScore: 12,
  dealerScore: 14,
  // The Go presenter sends an English `message` plus a `messageCode`. The GUI
  // must render the translated code, never the English fallback.
  message: 'Player wins!',
  messageCode: 'threecardrummy.result.playerWins',
};

/** End phase where the player held a meld and cashed both bonuses. */
const endPhaseMeld: ThreeCardRummyResponse = {
  ...endPhasePlayerWins,
  playerHand: meldHand,
  playerScore: 0,
  lowBonusBet: 50,
  // ドメインの式どおり: アンテ 100 × 9 = 900、ローボーナスは賭け金も戻るので
  // 50 + 50 × 100 = 5050。合計 = アンテ配当 200 + プレイ配当 200 + 900 + 5050。
  anteBonusPayout: 900,
  lowBonusPayout: 5050,
  totalPayout: 6350,
};

const endPhaseFold: ThreeCardRummyResponse = {
  ...endPhasePlayerWins,
  chips: 900,
  playBet: 0,
  result: -1,
  antePayout: 0,
  playPayout: 0,
  totalPayout: 0,
  // **降りてもディーラーの手は開く。** `Fold()` は dealerHand を消さず、Web
  // プレゼンタは End フェーズなら必ず開示する。`[]` はサーバが送らない状態で、
  // それを置くと降りたあとの表示が丸ごと未検証になる。
  dealerHand,
  dealerQualified: true,
  dealerScore: 14,
  message: 'Player folded.',
  messageCode: 'threecardrummy.result.fold',
};

beforeEach(() => {
  vi.clearAllMocks();
  // useGameHint persists the toggle; without this the hint test leaks into the rest.
  localStorage.removeItem('hint_enabled_threecardrummy');
});

describe('ThreeCardRummyPage', () => {
  // ── Bet phase ─────────────────────────────────────────────────────────────

  it('renders the bet controls on mount', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<ThreeCardRummyPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.getByLabelText('アンテ')).toBeInTheDocument();
    expect(screen.getByLabelText('ローボーナス')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument();
  });

  it('explains in the bet phase that a low total is good and where the dealer qualifies', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<ThreeCardRummyPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    // These two lines are the only place a first-time player learns the game is
    // inverted; losing either of them leaves the scoring unexplained.
    expect(
      screen.getByText('3枚の合計が低いほど強く、0点が最強。絵札=10、A=1。同ランク3枚か同スート連番3枚は0点。'),
    ).toBeInTheDocument();
    expect(screen.getByText('ディーラーは合計20点以下でクオリファイします。')).toBeInTheDocument();
  });

  it('bets the ante and low bonus typed into the inputs', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<ThreeCardRummyPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText('アンテ'), { target: { value: '200' } });
    fireEvent.change(screen.getByLabelText('ローボーナス'), { target: { value: '50' } });
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 200, 50));
  });

  // ── The player's score ────────────────────────────────────────────────────

  it("shows the player's total from the action phase, before the play/fold decision", async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    const { container } = renderWithProviders(<ThreeCardRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ' })).toBeInTheDocument());
    // ♠3 ♥4 ♦5 = 12. The total is the whole basis of the play/fold call, so it
    // has to be on screen while the decision is still open.
    expect(container.querySelector('[data-tutorial="tcr-results"]')).toHaveTextContent('点数: 12点');
  });

  it('renders a meld (total 0) as the perfect-hand text rather than a bare 0', async () => {
    mockExec.mockResolvedValue(actionPhaseMeldState);
    const { container } = renderWithProviders(<ThreeCardRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ' })).toBeInTheDocument());
    // ♠7 ♥7 ♦7 is a meld — 0 points, the strongest hand in the game. An earlier
    // version rendered the score only when it was non-zero and hid exactly this.
    const results = container.querySelector('[data-tutorial="tcr-results"]');
    expect(results).toHaveTextContent('点数: 0点（役 = 最強）');
    // jest-dom は空白を正規化するので「点数: 0点 」(末尾スペース) を否定しても
    // 常に通る。素の 0 が出ていないことは、役の文言そのものを否定して見る。
    expect(results?.textContent).not.toMatch(/点数: 0(?!点（)/);
  });

  it('keeps showing the perfect-hand text for a meld at the end phase', async () => {
    mockExec.mockResolvedValue(endPhaseMeld);
    const { container } = renderWithProviders(<ThreeCardRummyPage />);
    await waitFor(() => expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument());
    expect(container.querySelector('[data-tutorial="tcr-results"]')).toHaveTextContent('点数: 0点（役 = 最強）');
  });

  // ── The dealer's masked hand ──────────────────────────────────────────────

  it("renders the dealer's cards face down while the phase is action", async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<ThreeCardRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ' })).toBeInTheDocument());

    // Three backs, each announced as a hidden card rather than as a suit.
    expect(screen.getAllByRole('img', { name: '？' })).toHaveLength(3);
    expect(screen.getAllByTestId('animated-card-back')).toHaveLength(3);
    // The only face-up cards are the player's own three.
    expect(screen.getByAltText('♠ 3')).toBeInTheDocument();
    expect(screen.queryByAltText('♣ 5')).not.toBeInTheDocument();
  });

  it("shows the hidden placeholder for the dealer's score before the end", async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<ThreeCardRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ' })).toBeInTheDocument());
    // The player's own total is a number in the same render, so a bare "no
    // number anywhere" would not satisfy this.
    expect(screen.getByText('点数: 12点')).toBeInTheDocument();
    expect(screen.getByText('点数: ？')).toBeInTheDocument();
    expect(screen.queryByText('点数: 14点')).not.toBeInTheDocument();
    expect(screen.queryByText(/クオリファイ/)).not.toBeInTheDocument();
  });

  it("reveals the dealer's cards and score at the end phase", async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<ThreeCardRummyPage />);
    await waitFor(() => expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument());

    expect(screen.queryAllByTestId('animated-card-back')).toHaveLength(0);
    expect(screen.queryAllByRole('img', { name: '？' })).toHaveLength(0);
    expect(screen.getByAltText('♣ 5')).toBeInTheDocument();
    expect(screen.getByAltText('♦ 2')).toBeInTheDocument();
    // ♣5 ♦2 ♥7 = 14, and 14 <= 20 so the dealer qualified.
    expect(screen.getByText('点数: 14点')).toBeInTheDocument();
    expect(screen.getByText('クオリファイ')).toBeInTheDocument();
  });

  // ── Action phase commands ─────────────────────────────────────────────────

  it('sends play when the play button is pressed', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<ThreeCardRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(endPhasePlayerWins);
    fireEvent.click(screen.getByRole('button', { name: 'プレイ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play'));
  });

  it('sends fold when the fold button is pressed', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<ThreeCardRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(endPhaseFold);
    fireEvent.click(screen.getByRole('button', { name: 'フォールド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('fold'));
  });

  it('shows the ante and the amount the play bet will cost during the action phase', async () => {
    mockExec.mockResolvedValue({ ...actionPhaseState, lowBonusBet: 50 });
    renderWithProviders(<ThreeCardRummyPage />);
    await waitFor(() => expect(screen.getByTestId('action-bet-slip')).toBeInTheDocument());
    const slip = screen.getByTestId('action-bet-slip');
    expect(slip).toHaveTextContent('アンテ: 100');
    expect(slip).toHaveTextContent('ローボーナス: 50');
    expect(slip).toHaveTextContent('プレイに必要: 100');
  });

  // ── End phase ─────────────────────────────────────────────────────────────

  it('shows the result message and the payout breakdown at the end phase', async () => {
    mockExec.mockResolvedValueOnce(actionPhaseState).mockResolvedValueOnce(endPhasePlayerWins);
    renderWithProviders(<ThreeCardRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'プレイ' }));
    await waitFor(() => expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument());

    // NOTE: `messageCode.threecardrummy.*` is missing from the frontend locales,
    // so today the box falls back to the server's raw English `message`. Either
    // side of that fix satisfies this matcher; an unrendered box does not.
    // **英語のフォールバックではなく翻訳が出ること。** messageCode の
    // common.json エントリが無いと素の "Player wins!" が日本語 UI に出る。
    expect(screen.getByText('勝利！')).toBeInTheDocument();
    expect(screen.queryByText('Player wins!')).not.toBeInTheDocument();

    const breakdown = screen.getByTestId('payout-breakdown');
    expect(breakdown).toHaveTextContent('アンテ: 200');
    expect(breakdown).toHaveTextContent('プレイ: 200');
    expect(breakdown).toHaveTextContent('合計: 400');
  });

  it('lists the ante bonus and low bonus rows when they paid', async () => {
    mockExec.mockResolvedValue(endPhaseMeld);
    renderWithProviders(<ThreeCardRummyPage />);
    await waitFor(() => expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument());
    const breakdown = screen.getByTestId('payout-breakdown');
    expect(breakdown).toHaveTextContent('アンテボーナス: 900');
    expect(breakdown).toHaveTextContent('ローボーナス: 5050');
    expect(breakdown).toHaveTextContent('合計: 6350');
  });

  it('omits the low bonus row when no side bet was placed', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<ThreeCardRummyPage />);
    await waitFor(() => expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument());
    expect(screen.getByTestId('payout-breakdown')).not.toHaveTextContent('ローボーナス');
  });

  it('reveals the dealer after a fold and pays nothing', async () => {
    // 降りてもラウンドは終わっているので、ディーラーの手は開く。以前は
    // `dealerHand: []` という**サーバが送らない状態**を渡していたため、
    // 降りたあとの表示が丸ごと未検証だった。
    mockExec.mockResolvedValue(endPhaseFold);
    renderWithProviders(<ThreeCardRummyPage />);
    await waitFor(() => expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument());
    expect(screen.getByText('フォールド', { selector: ':not(button)' })).toBeInTheDocument();
    expect(screen.queryByText('Player folded.')).not.toBeInTheDocument();
    // 伏せ札ではなく実物が出る。
    expect(screen.getByAltText('♣ 5')).toBeInTheDocument();
    expect(screen.getByText('点数: 14点')).toBeInTheDocument();
    expect(screen.getByTestId('payout-breakdown')).toHaveTextContent('合計: 0');
  });

  // ── Rebet ─────────────────────────────────────────────────────────────────

  it('offers a rebet after a round and re-sends the previous ante and low bonus', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(endPhasePlayerWins);
    renderWithProviders(<ThreeCardRummyPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText('アンテ'), { target: { value: '200' } });
    fireEvent.change(screen.getByLabelText('ローボーナス'), { target: { value: '50' } });
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));

    await waitFor(() => expect(screen.getByTestId('tcr-rebet-button')).toBeInTheDocument());
    // 200 ante + 50 low bonus is the outlay the button advertises.
    expect(screen.getByTestId('tcr-rebet-button')).toHaveTextContent('250');

    mockExec.mockClear();
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(actionPhaseState);
    fireEvent.click(screen.getByTestId('tcr-rebet-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 200, 50));
  });

  it('offers no rebet at the end phase when no bet was placed this session', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<ThreeCardRummyPage />);
    await waitFor(() => expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument());
    expect(screen.queryByTestId('tcr-rebet-button')).not.toBeInTheDocument();
  });

  // ── Hint ──────────────────────────────────────────────────────────────────

  it('surfaces the hint tooltip during the action phase once the hint toggle is on', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<ThreeCardRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ' })).toBeInTheDocument());
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('checkbox', { name: 'ヒント表示' }));

    // 12 points is under the dealer's 20-point qualifying ceiling → play.
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
    expect(within(screen.getByTestId('hint-tooltip')).getByText(/プレイ推奨/)).toBeInTheDocument();
  });

  // ── Misc ──────────────────────────────────────────────────────────────────

  it('renders the skeleton until the first state arrives', () => {
    mockExec.mockReturnValue(new Promise(() => {}));
    renderWithProviders(<ThreeCardRummyPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('shows the action log on request', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    vi.mocked(actionLogApi.threecardrummy).mockResolvedValue({ entries: [] as never[] });
    renderWithProviders(<ThreeCardRummyPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    fireEvent.click(screen.getByText('棋譜を見る'));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());
  });
});
