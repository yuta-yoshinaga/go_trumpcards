import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, maoApi } from '../api/gameApi';
import i18n from '../i18n';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { MaoResponse } from '../types/card';
import { MaoPage } from './MaoPage';

vi.mock('../api/gameApi', () => ({
  maoApi: { exec: vi.fn() },
  actionLogApi: { mao: vi.fn() },
}));

const mockPlaySound = vi.fn();
const mockSoundValue = { playSound: mockPlaySound, muted: false, toggleMute: vi.fn() };
vi.mock('../providers/SoundProvider', () => ({
  SoundProvider: ({ children }: { children: React.ReactNode }) => children,
  useSound: () => mockSoundValue,
  useOptionalSound: () => mockSoundValue,
}));

const mockExec = vi.mocked(maoApi.exec);

const playPhaseState: MaoResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 5,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
      ],
      roundScore: 0,
      cumulativeScore: 0,
      hasDeclared: false,
    },
    { id: 1, isHuman: false, cardCount: 5, cards: [], roundScore: 3, cumulativeScore: 10, hasDeclared: false },
    { id: 2, isHuman: false, cardCount: 5, cards: [], roundScore: 5, cumulativeScore: 20, hasDeclared: false },
    { id: 3, isHuman: false, cardCount: 5, cards: [], roundScore: 0, cumulativeScore: 5, hasDeclared: false },
  ],
  phase: 0,
  roundNumber: 1,
  currentPlayerIdx: 0,
  discardTop: { design: 'HEART', value: 7 },
  drawPileCount: 30,
  chosenSuit: 0,
  penaltyDrawCount: 0,
  direction: 1,
  gameEndFlag: false,
  winnerIdx: -1,
  awaitingWord: false,
  correctCount: 0,
  hintUnlocked: false,
  ruleHint: '',
  rulePenalty: false,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 200 },
};

const chooseSuitState: MaoResponse = { ...playPhaseState, phase: 1 };
const mustDeclareState: MaoResponse = { ...playPhaseState, phase: 2 };
const roundEndState: MaoResponse = { ...playPhaseState, phase: 3 };
const gameEndState: MaoResponse = {
  ...playPhaseState,
  phase: 4,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};
const penaltyState: MaoResponse = { ...playPhaseState, penaltyDrawCount: 4 };
const rulePenaltyState: MaoResponse = { ...playPhaseState, rulePenalty: true };

beforeEach(() => {
  mockExec.mockReset();
  mockPlaySound.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

/**
 * 唱えた言葉の読み上げ用リージョン。
 *
 * **LiveAnnouncement は中身より先に mount される**ので (領域と本文が同じ
 * コミットで現れると読み上げられない)、testid ではなく role と aria-live で引く。
 */
const sayWordLive = () => document.querySelector('[role="status"][aria-live="polite"][aria-atomic="true"]');

describe('MaoPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<MaoPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<MaoPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { cpuDifficulty: 1, pointLimit: 200 }),
    );
  });

  it('renders human cards', async () => {
    renderWithProviders(<MaoPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
      expect(screen.getByAltText('♥ J')).toBeInTheDocument();
    });
  });

  it('labels the human hand as a named region and preserves card selection state', async () => {
    renderWithProviders(<MaoPage />);
    const hand = await screen.findByRole('region', { name: 'あなたの手札' });
    const card = screen.getByRole('button', { name: '♠ A' });

    expect(hand).toContainElement(card);
    expect(card).toHaveAttribute('aria-pressed', 'false');
    fireEvent.click(card);
    expect(card).toHaveAttribute('aria-pressed', 'true');

    await i18n.changeLanguage('en');
    expect(screen.getByRole('region', { name: 'Your hand' })).toContainElement(card);
    await i18n.changeLanguage('ja');
  });

  it('calls play when play button clicked', async () => {
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0));
  });

  it('calls draw when draw button clicked', async () => {
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '引く' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '引く' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('announces each card added by a human draw once', async () => {
    const drawnCard = { design: 'DIAMOND', value: 12 } as const;
    mockExec.mockResolvedValueOnce(playPhaseState).mockResolvedValueOnce({
      ...playPhaseState,
      players: playPhaseState.players.map((player) =>
        player.isHuman ? { ...player, cards: [...(player.cards ?? []), drawnCard] } : player,
      ),
    });
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '引く' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: '引く' }));
    const live = await screen.findByTestId('mao-draw-announcement');
    await waitFor(() => expect(live).toHaveTextContent('♦ Q'));
    expect(live).toHaveAttribute('role', 'status');
    expect(live).toHaveAttribute('aria-live', 'polite');
  });

  it('clears a draw announcement after a successful play response', async () => {
    const drawnCard = { design: 'DIAMOND', value: 12 } as const;
    mockExec
      .mockResolvedValueOnce(playPhaseState)
      .mockResolvedValueOnce({
        ...playPhaseState,
        players: playPhaseState.players.map((player) =>
          player.isHuman ? { ...player, cards: [...(player.cards ?? []), drawnCard] } : player,
        ),
      })
      .mockResolvedValueOnce(playPhaseState);
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '引く' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: '引く' }));
    const live = await screen.findByTestId('mao-draw-announcement');
    await waitFor(() => expect(live).toHaveTextContent('♦ Q'));

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenLastCalledWith('play', 0));
    await waitFor(() => expect(live).toBeEmptyDOMElement());
  });

  it('does not announce a draw response when the human hand did not grow', async () => {
    mockExec.mockResolvedValue({
      ...playPhaseState,
      message: 'Draw response received',
      players: playPhaseState.players.map((player) => ({ ...player })),
    });
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '引く' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: '引く' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
    expect(await screen.findByText('Draw response received')).toBeInTheDocument();
    expect(await screen.findByTestId('mao-draw-announcement')).toBeEmptyDOMElement();
  });

  it('does not announce cards from a later play response after a rejected draw', async () => {
    const unrelatedCard = { design: 'DIAMOND', value: 12 } as const;
    mockExec
      .mockResolvedValueOnce(playPhaseState)
      .mockRejectedValueOnce(new Error('draw rejected'))
      .mockResolvedValueOnce({
        ...playPhaseState,
        players: playPhaseState.players.map((player) =>
          player.isHuman ? { ...player, cards: [...(player.cards ?? []), unrelatedCard] } : player,
        ),
      });
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '引く' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: '引く' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
    await waitFor(() => expect(screen.getByRole('button', { name: '引く' })).toBeEnabled());
    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenLastCalledWith('play', 0));
    expect(await screen.findByTestId('mao-draw-announcement')).toBeEmptyDOMElement();
  });

  it('does not announce cards from play or round transition responses', async () => {
    const unrelatedCard = { design: 'DIAMOND', value: 12 } as const;
    mockExec.mockResolvedValueOnce(roundEndState).mockResolvedValueOnce({
      ...playPhaseState,
      players: playPhaseState.players.map((player) =>
        player.isHuman ? { ...player, cards: [...(player.cards ?? []), unrelatedCard] } : player,
      ),
    });
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /次のラウンド|Next Round/ })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: /次のラウンド|Next Round/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
    expect(await screen.findByTestId('mao-draw-announcement')).toBeEmptyDOMElement();
  });

  it('buzzes and flashes the rule panel when a rule penalty lands', async () => {
    mockExec.mockResolvedValue(rulePenaltyState);
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByTestId('rule-penalty')).toBeInTheDocument());
    expect(mockPlaySound).toHaveBeenCalledWith('errorBuzz');
    const panel = screen.getByTestId('mao-rule-panel');
    expect(panel.className).toContain('motion-safe:animate-pulse');
    expect(panel.className).toContain('ring-ds-error');
  });

  it('announces a rule penalty through a live region', async () => {
    mockExec.mockResolvedValue(rulePenaltyState);
    renderWithProviders(<MaoPage />);

    const live = await waitFor(() => {
      const element = [...document.querySelectorAll('[role="status"][aria-live="polite"][aria-atomic="true"]')].find(
        (candidate) => candidate.textContent?.includes('ペナルティ'),
      );
      expect(element).not.toBeUndefined();
      return element;
    });
    expect(live).toHaveAttribute('aria-live', 'polite');
    expect(live).toHaveAttribute('aria-atomic', 'true');
  });

  it('does not buzz when there is no rule penalty', async () => {
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByTestId('mao-rule-panel')).toBeInTheDocument());
    expect(mockPlaySound).not.toHaveBeenCalledWith('errorBuzz');
  });

  it('shows penalty banner and take-penalty draw label when penalty active', async () => {
    mockExec.mockResolvedValue(penaltyState);
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByText(/ドローペナルティ/)).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /引き受ける/ })).toBeInTheDocument();
    expect(await screen.findByTestId('penalty-badge')).toHaveTextContent('4');
  });

  it('renders choose suit phase and calls suit', async () => {
    mockExec.mockResolvedValue(chooseSuitState);
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スペード' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ハート' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('suit', undefined, 3));
  });

  it('shows a color-coded suit symbol alongside each suit-selection button', async () => {
    mockExec.mockResolvedValue(chooseSuitState);
    renderWithProviders(<MaoPage />);
    // CrazyEightsSuit: SPADE=1, CLOVER=2, HEART=3, DIAMOND=4
    await waitFor(() => expect(screen.getByTestId('suit-symbol-1')).toBeInTheDocument());

    // Every suit button shows both its glyph and its text label.
    expect(screen.getByTestId('suit-symbol-1')).toHaveTextContent('♠');
    expect(screen.getByTestId('suit-symbol-2')).toHaveTextContent('♣');
    expect(screen.getByTestId('suit-symbol-3')).toHaveTextContent('♥');
    expect(screen.getByTestId('suit-symbol-4')).toHaveTextContent('♦');
    expect(screen.getByRole('button', { name: 'ダイヤ' })).toBeInTheDocument();

    // Red suits (hearts, diamonds) carry the red token; black suits use the ivory primary token.
    expect(screen.getByTestId('suit-symbol-3').className).toContain('text-ds-error-text');
    expect(screen.getByTestId('suit-symbol-4').className).toContain('text-ds-error-text');
    expect(screen.getByTestId('suit-symbol-1').className).toContain('text-ds-text-primary');
    expect(screen.getByTestId('suit-symbol-2').className).toContain('text-ds-text-primary');
  });

  it('calls declare when declare button clicked', async () => {
    mockExec.mockResolvedValue(mustDeclareState);
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'マオ！' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'マオ！' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declare'));
  });

  it('calls skipdeclare when skip button clicked', async () => {
    mockExec.mockResolvedValue(mustDeclareState);
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '宣言しない' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '宣言しない' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('skipdeclare'));
  });

  it('shows next round button and calls nextround', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のラウンド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('renders the hidden-rule panel with compliance progress', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, correctCount: 2 });
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByTestId('mao-rule-panel')).toBeInTheDocument());
    expect(screen.getByText(/ルール遵守: 2\/3/)).toBeInTheDocument();
  });

  it('says a word via declareword and clears the input', async () => {
    renderWithProviders(<MaoPage />);
    const input = await screen.findByLabelText('唱える言葉を入力…');
    fireEvent.change(input, { target: { value: 'mao' } });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '発言する' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declareword', undefined, undefined, undefined, 'mao'));
    await waitFor(() => expect((input as HTMLInputElement).value).toBe(''));
  });

  it('disables say-word button when the input is empty', async () => {
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '発言する' })).toBeDisabled());
  });

  it('hides the say-word history panel until a word is spoken', async () => {
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByTestId('mao-rule-panel')).toBeInTheDocument());
    expect(screen.queryByTestId('mao-sayword-history')).not.toBeInTheDocument();
  });

  it('logs a say-word attempt with an accepted outcome and the board card', async () => {
    renderWithProviders(<MaoPage />);
    const input = await screen.findByLabelText('唱える言葉を入力…');
    fireEvent.change(input, { target: { value: 'mao' } });
    // Response reports no penalty → the attempt was accepted.
    mockExec.mockResolvedValueOnce({ ...playPhaseState, rulePenalty: false, correctCount: 1 });
    fireEvent.click(screen.getByRole('button', { name: '発言する' }));
    const row = await screen.findByTestId('mao-sayword-history-row');
    expect(row).toHaveTextContent('mao');
    // Board captured from the discard top at submit time (HEART 7).
    expect(row).toHaveTextContent('♥ 7');
    expect(screen.getByTestId('sayword-outcome-correct')).toHaveTextContent('正解');
    expect(screen.queryByTestId('sayword-outcome-penalty')).not.toBeInTheDocument();
  });

  // #5668: 違反は role="status" の rule-penalty とブザー音で強く伝わるのに、成功は
  // 折りたたみ履歴に静かに積まれるだけだった。**隠しルールを試行錯誤で学ぶゲーム**
  // なので、耳に届く情報が違反だけなのは学習の材料として非対称。
  it('announces an accepted say-word through a live region', async () => {
    renderWithProviders(<MaoPage />);
    const input = await screen.findByLabelText('唱える言葉を入力…');
    fireEvent.change(input, { target: { value: 'mao' } });
    mockExec.mockResolvedValueOnce({ ...playPhaseState, rulePenalty: false, correctCount: 1 });
    fireEvent.click(screen.getByRole('button', { name: '発言する' }));

    // **領域は中身より先に mount される** ので、testid ではなく role で引く。
    await waitFor(() => expect(sayWordLive()).toHaveTextContent('mao'));
    expect(sayWordLive()).toHaveAttribute('aria-live', 'polite');
    expect(sayWordLive()).toHaveAttribute('aria-atomic', 'true');
    expect(sayWordLive()).toHaveTextContent('正解');
  });

  it('announces a penalty through the same live region', async () => {
    renderWithProviders(<MaoPage />);
    const input = await screen.findByLabelText('唱える言葉を入力…');
    fireEvent.change(input, { target: { value: 'oops' } });
    mockExec.mockResolvedValueOnce({ ...playPhaseState, rulePenalty: true });
    fireEvent.click(screen.getByRole('button', { name: '発言する' }));

    await waitFor(() => expect(sayWordLive()).toHaveTextContent('ペナルティ'));
  });

  // **2回目以降は最新を読む。**1件だけのテストは、先頭を読む実装でも末尾を読む
  // 実装でも通ってしまう。
  it('announces the newest attempt, not the first', async () => {
    renderWithProviders(<MaoPage />);
    const input = await screen.findByLabelText('唱える言葉を入力…');

    fireEvent.change(input, { target: { value: 'first' } });
    mockExec.mockResolvedValueOnce({ ...playPhaseState, rulePenalty: false, correctCount: 1 });
    fireEvent.click(screen.getByRole('button', { name: '発言する' }));
    await waitFor(() => expect(sayWordLive()).toHaveTextContent('first'));

    fireEvent.change(input, { target: { value: 'second' } });
    mockExec.mockResolvedValueOnce({ ...playPhaseState, rulePenalty: true, correctCount: 1 });
    fireEvent.click(screen.getByRole('button', { name: '発言する' }));

    await waitFor(() => expect(sayWordLive()).toHaveTextContent('second'));
    expect(sayWordLive()).toHaveTextContent('ペナルティ');
  });

  it('says nothing before any word is declared', async () => {
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByTestId('mao-rule-panel')).toBeInTheDocument());

    expect(screen.queryByTestId('sayword-live')).not.toBeInTheDocument();
  });

  it('color-codes a penalty say-word attempt in the history panel', async () => {
    renderWithProviders(<MaoPage />);
    const input = await screen.findByLabelText('唱える言葉を入力…');
    fireEvent.change(input, { target: { value: 'oops' } });
    mockExec.mockResolvedValueOnce({ ...playPhaseState, rulePenalty: true });
    fireEvent.click(screen.getByRole('button', { name: '発言する' }));
    const outcome = await screen.findByTestId('sayword-outcome-penalty');
    expect(outcome).toHaveTextContent('ペナルティ');
    expect(outcome.className).toContain('text-ds-error-text');
  });

  it('clears the say-word history on reset', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<MaoPage />);
    const input = await screen.findByLabelText('唱える言葉を入力…');
    fireEvent.change(input, { target: { value: 'mao' } });
    mockExec.mockResolvedValueOnce({ ...gameEndState, rulePenalty: false });
    fireEvent.click(screen.getByRole('button', { name: '発言する' }));
    await screen.findByTestId('mao-sayword-history-row');
    // At game end the reset button fires immediately (no confirm dialog) and wipes the log.
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() => expect(screen.queryByTestId('mao-sayword-history')).not.toBeInTheDocument());
  });

  it('shows a penalty notice when rulePenalty is true', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, rulePenalty: true });
    renderWithProviders(<MaoPage />);
    expect(await screen.findByTestId('rule-penalty')).toBeInTheDocument();
  });

  it('shows awaiting-word signal when awaitingWord is true', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, awaitingWord: true });
    renderWithProviders(<MaoPage />);
    expect(await screen.findByTestId('awaiting-word')).toBeInTheDocument();
  });

  it('shows the rule hint only when hintUnlocked', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, hintUnlocked: true, ruleHint: 'be polite' });
    renderWithProviders(<MaoPage />);
    expect(await screen.findByTestId('rule-hint')).toHaveTextContent('be polite');
  });

  // **コードが来たら訳す。**サーバの i18n 言語はプロセス全体で 1 つなので、
  // ruleHint をそのまま出すと英語ロケールでも日本語のままになる (#4917)。
  it('translates the rule hint code instead of the server string', async () => {
    mockExec.mockResolvedValue({
      ...playPhaseState,
      hintUnlocked: true,
      ruleHint: 'A word is required when a certain suit is played.',
      ruleHintCode: 'hintSuit',
    });
    renderWithProviders(<MaoPage />);
    const hint = await screen.findByTestId('rule-hint');
    expect(hint).toHaveTextContent('あるスートを出したときに言葉が必要です。');
    expect(hint).not.toHaveTextContent('hintSuit');
  });

  it('hides the rule hint when not unlocked', async () => {
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByTestId('mao-rule-panel')).toBeInTheDocument());
    expect(screen.queryByTestId('rule-hint')).not.toBeInTheDocument();
  });

  it('shows discard top card and round info', async () => {
    renderWithProviders(<MaoPage />);
    await waitFor(() => {
      expect(screen.getByText('捨て札')).toBeInTheDocument();
      expect(screen.getByAltText('♥ 7')).toBeInTheDocument();
      expect(screen.getByText('ラウンド 1')).toBeInTheDocument();
    });
  });

  it('shows game end with action log button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<MaoPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('opens an action log during play and disables card shortcuts while open', async () => {
    vi.mocked(actionLogApi.mao).mockResolvedValue({
      entries: [{ turnNumber: 1, playerIdx: 0, actionType: 'play', detail: 'Alice played ♥ 7' }],
    });
    renderWithProviders(<MaoPage />);
    const button = await screen.findByRole('button', { name: '棋譜を見る' });
    fireEvent.click(button);
    await waitFor(() => expect(screen.getByText(/Alice played ♥ 7/)).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: '1' });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('reset button calls exec with confirm', async () => {
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { cpuDifficulty: 1, pointLimit: 200 }),
    );
  });

  it('renders accessible h1 heading', async () => {
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  // **ヒント経路はページ側からも踏む。**ファクトリ単体テストだけだと
  // `hintFactories` の登録行と、ページのトグル／ツールチップが一度も
  // 実行されない（codecov が #4596 で同じ 3 ファイルを未到達と報告した）。
  it('turns the frontend hint on from the settings panel', async () => {
    localStorage.removeItem('hint_enabled_mao');
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<MaoPage />);

    const toggle = await screen.findByRole('checkbox', { name: 'ヒント表示' });
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();

    fireEvent.click(toggle);
    expect(await screen.findByTestId('hint-tooltip')).toBeInTheDocument();
  });

  // 隠しルールに触れないことを、ページ越しにも見る。
  it('shows no tooltip during the declaration phase', async () => {
    localStorage.setItem('hint_enabled_mao', 'true');
    mockExec.mockResolvedValue(mustDeclareState);
    renderWithProviders(<MaoPage />);
    await screen.findByRole('checkbox', { name: 'ヒント表示' });
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();
    localStorage.removeItem('hint_enabled_mao');
  });

  // スート指定がある場合のみ、捨て札背景に指定スートの透かしを表示する。
  it('renders chosen-suit-watermark when chosenSuit is greater than zero', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, chosenSuit: 1 });
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByTestId('chosen-suit-watermark')).toBeInTheDocument());
    expect(screen.getByTestId('chosen-suit-status')).toHaveTextContent('指定スート: スペード ♠');
  });

  it('falls back to the symbol alone for a suit value it has no name for', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, chosenSuit: 9 });
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByTestId('chosen-suit-status')).toHaveTextContent('指定スート: ?'));
  });

  it('shows the chosen suit name in English when English is selected', async () => {
    await i18n.changeLanguage('en');
    try {
      mockExec.mockResolvedValue({ ...playPhaseState, chosenSuit: 1 });
      renderWithProviders(<MaoPage />);
      await waitFor(() => expect(screen.getByTestId('chosen-suit-status')).toHaveTextContent('Chosen suit: Spade ♠'));
    } finally {
      await i18n.changeLanguage('ja');
    }
  });

  it('does not render chosen-suit-watermark when chosenSuit is zero', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    expect(screen.queryByTestId('chosen-suit-watermark')).not.toBeInTheDocument();
  });
});
