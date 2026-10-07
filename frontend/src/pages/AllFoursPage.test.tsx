import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { allfoursApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { AllFoursResponse } from '../types/card';
import { AllFoursPhase } from '../types/phases';
import { AllFoursPage } from './AllFoursPage';

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

const mockUseCliMode = vi.mocked(useCliMode);

vi.mock('../api/gameApi', () => ({
  allfoursApi: { exec: vi.fn() },
  actionLogApi: { allfours: vi.fn() },
}));

const mockExec = vi.mocked(allfoursApi.exec);

const baseState: AllFoursResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 6,
      cards: [{ design: 'HEART', value: 5 }],
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
    },
    { id: 1, isHuman: false, cardCount: 6, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0 },
  ],
  phase: AllFoursPhase.BEG,
  roundNumber: 1,
  trickNumber: 0,
  dealerIdx: 1,
  nonDealerIdx: 0,
  currentPlayerIdx: 0,
  trumpSuit: 3,
  turnUp: { design: 'HEART', value: 7 },
  runCount: 0,
  currentTrick: [],
  gameEndFlag: false,
  winnerIdx: -1,
  leadPlayerIdx: -1,
  validPlayIndices: [0],
  config: { cpuDifficulty: 1, pointLimit: 7 },
  message: '',
  messageCode: 'allfours.begPhase',
};

const playState: AllFoursResponse = {
  ...baseState,
  phase: AllFoursPhase.PLAY,
  trickNumber: 1,
  currentPlayerIdx: 0,
  validPlayIndices: [0],
};

const gameEndState: AllFoursResponse = {
  ...baseState,
  phase: AllFoursPhase.GAME_END,
  gameEndFlag: true,
  winnerIdx: 0,
  players: [
    { id: 0, isHuman: true, cardCount: 0, cards: [], roundScore: 0, cumulativeScore: 7, trickCount: 6 },
    { id: 1, isHuman: false, cardCount: 0, cards: [], roundScore: 0, cumulativeScore: 3, trickCount: 0 },
  ],
};

beforeEach(() => {
  mockExec.mockResolvedValue(baseState);
  mockUseCliMode.mockReturnValue({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  });
});

describe('AllFoursPage', () => {
  it('names the player score table and associates each column heading', async () => {
    renderWithProviders(<AllFoursPage />);

    const table = await screen.findByRole('table', { name: 'プレイヤー別スコア表' });
    for (const heading of ['プレイヤー', 'トリック', 'ディール', '累計']) {
      expect(within(table).getByRole('columnheader', { name: heading })).toHaveAttribute('scope', 'col');
    }
  });

  it('keeps an unplayable card focusable and describes the follow-suit rule', async () => {
    mockExec.mockResolvedValue({
      ...playState,
      currentTrick: [{ playerIdx: 1, card: { design: 'HEART', value: 7 } }],
      players: [
        {
          ...baseState.players[0],
          cards: [
            { design: 'HEART', value: 5 },
            { design: 'SPADE', value: 6 },
          ],
        },
        baseState.players[1],
      ],
      validPlayIndices: [0],
    });
    renderWithProviders(<AllFoursPage />);

    const allowed = (await screen.findByAltText('♥ 5')).closest('button') as HTMLButtonElement;
    const blocked = screen.getByAltText('♠ 6').closest('button') as HTMLButtonElement;
    expect(allowed).not.toHaveAttribute('aria-disabled');
    expect(blocked).toHaveAttribute('aria-disabled', 'true');
    expect(blocked).not.toBeDisabled();
    expect(document.getElementById(blocked.getAttribute('aria-describedby') ?? '')).toHaveTextContent(
      'リードスートを持っている場合は、そのスートか切り札を出してください。',
    );

    fireEvent.click(blocked);
    expect(screen.getByRole('button', { name: '出す' })).toBeDisabled();
  });

  it('announces an unplayed trick separately from a completed trick', async () => {
    mockExec.mockResolvedValue(playState);
    const { unmount } = renderWithProviders(<AllFoursPage />);
    expect(await screen.findByTestId('af-trick-status')).toHaveTextContent('まだカードが出ていません');
    unmount();

    mockExec.mockResolvedValue({ ...playState, phase: AllFoursPhase.TRICK_END });
    renderWithProviders(<AllFoursPage />);
    expect(await screen.findByTestId('af-trick-status')).toHaveTextContent('トリックが終了しました');
  });

  it('announces a completed trick while its four cards remain visible', async () => {
    mockExec.mockResolvedValue({
      ...playState,
      phase: AllFoursPhase.TRICK_END,
      currentTrick: [
        { playerIdx: 0, card: { design: 'HEART', value: 5 } },
        { playerIdx: 1, card: { design: 'CLOVER', value: 6 } },
        { playerIdx: 0, card: { design: 'DIAMOND', value: 7 } },
        { playerIdx: 1, card: { design: 'SPADE', value: 8 } },
      ],
    });
    renderWithProviders(<AllFoursPage />);

    expect(await screen.findByText('トリックが終了しました')).toBeInTheDocument();
    expect(screen.getAllByRole('figure')).toHaveLength(4);
  });

  it('associates each played card with its player in the accessible name', async () => {
    mockExec.mockResolvedValueOnce({
      ...playState,
      currentTrick: [{ playerIdx: 0, card: { design: 'HEART', value: 5 } }],
    });
    renderWithProviders(<AllFoursPage />);
    expect(await screen.findByRole('figure', { name: 'あなた: ♥ 5' })).toBeInTheDocument();
  });

  it('renders the GameSkeleton while state is null', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<AllFoursPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<AllFoursPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows exactly one reset control mid-game that opens a confirm dialog', async () => {
    renderWithProviders(<AllFoursPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
    // Only the consolidated GameResetButton remains — no duplicate reset button.
    expect(screen.getAllByRole('button', { name: 'リセット' })).toHaveLength(1);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('shows a single 次のゲーム reset control at game end (no duplicate reset button)', async () => {
    mockExec.mockResolvedValueOnce(gameEndState);
    renderWithProviders(<AllFoursPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'リセット' })).not.toBeInTheDocument();
  });

  it('shows the action log via ActionLogSection at game end', async () => {
    mockExec.mockResolvedValueOnce(gameEndState);
    renderWithProviders(<AllFoursPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('shows stand and beg buttons in beg phase', async () => {
    renderWithProviders(<AllFoursPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スタンド' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ベグ' })).toBeInTheDocument();
  });

  it('stand button calls exec with beg=false', async () => {
    renderWithProviders(<AllFoursPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スタンド' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スタンド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('beg', false));
  });

  it('beg button calls exec with beg=true', async () => {
    renderWithProviders(<AllFoursPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベグ' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'ベグ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('beg', true));
  });

  it('shows gift/run buttons when dealer is human in gift phase', async () => {
    mockExec.mockResolvedValueOnce({ ...baseState, phase: AllFoursPhase.GIFT, dealerIdx: 0, nonDealerIdx: 1 });
    renderWithProviders(<AllFoursPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /ギフト/ })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /ラン/ })).toBeInTheDocument();
  });

  it('play button calls exec with play and card index', async () => {
    mockExec.mockResolvedValueOnce(playState);
    renderWithProviders(<AllFoursPage />);
    // Select the first (valid) card, then play.
    await waitFor(() => expect(screen.getByRole('button', { name: '♥5' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: '♥5' }));
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, undefined, 0));
  });

  it('announces which hand card is selected and updates when selection changes', async () => {
    mockExec.mockResolvedValueOnce({
      ...playState,
      players: [
        {
          ...playState.players[0],
          cards: [
            { design: 'HEART', value: 5 },
            { design: 'SPADE', value: 3 },
          ],
        },
        playState.players[1],
      ],
      validPlayIndices: [0, 1],
    });
    renderWithProviders(<AllFoursPage />);

    const firstCard = await screen.findByRole('button', { name: '♥5' });
    const secondCard = screen.getByRole('button', { name: '♠3' });
    expect(firstCard).toHaveAttribute('aria-label', '♥5');
    expect(firstCard).toHaveAttribute('aria-pressed', 'false');
    expect(secondCard).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(firstCard);
    expect(firstCard).toHaveAttribute('aria-pressed', 'true');
    expect(secondCard).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(secondCard);
    expect(firstCard).toHaveAttribute('aria-pressed', 'false');
    expect(secondCard).toHaveAttribute('aria-pressed', 'true');
  });

  it('reads out the trump suit and turn-up by name', async () => {
    renderWithProviders(<AllFoursPage />); // trumpSuit 3 = ♥, turnUp ♥7
    expect(await screen.findByRole('img', { name: '切り札: ハート' })).toBeInTheDocument();
    expect(screen.getByRole('img', { name: 'めくり札: ハート7' })).toBeInTheDocument();
  });

  it('spells a face-card turn-up with its letter (♠J → スペードJ)', async () => {
    mockExec.mockResolvedValueOnce({ ...baseState, trumpSuit: 1, turnUp: { design: 'SPADE', value: 11 } });
    renderWithProviders(<AllFoursPage />);
    expect(await screen.findByRole('img', { name: 'めくり札: スペードJ' })).toBeInTheDocument();
  });

  it('exposes the hint toggle as a labelled checkbox in the settings panel', async () => {
    renderWithProviders(<AllFoursPage />);
    const toggle = await screen.findByRole('checkbox', { name: /ヒント/ });
    expect(toggle).toBeInTheDocument();
  });

  it('reads the trump as unset and omits the turn-up before it is decided', async () => {
    mockExec.mockResolvedValueOnce({ ...baseState, trumpSuit: 0, turnUp: null });
    renderWithProviders(<AllFoursPage />);
    expect(await screen.findByRole('img', { name: '切り札: 未確定' })).toBeInTheDocument();
    // No turn-up card is shown before it is flipped.
    expect(screen.queryByRole('img', { name: /めくり札/ })).not.toBeInTheDocument();
  });

  it('shows winner message at game end', async () => {
    mockExec.mockResolvedValueOnce(gameEndState);
    renderWithProviders(<AllFoursPage />);
    await waitFor(() => expect(screen.getByText('あなたの勝利！')).toBeInTheDocument());
  });

  it('renders the High/Low/Jack/Game breakdown at round end', async () => {
    const roundEndState: AllFoursResponse = {
      ...baseState,
      phase: AllFoursPhase.ROUND_END,
      roundBreakdown: {
        high: { winnerIdx: 0, card: { design: 'HEART', value: 1 } },
        low: { winnerIdx: 1, card: { design: 'HEART', value: 2 } },
        jack: { winnerIdx: 0 },
        game: { winnerIdx: 0, points: [5, 0] },
        provisional: false,
      },
    };
    mockExec.mockResolvedValueOnce(roundEndState);
    renderWithProviders(<AllFoursPage />);
    const panel = await screen.findByTestId('af-breakdown');
    expect(panel).toHaveTextContent('得点内訳');
    const table = within(panel).getByRole('table', { name: '得点内訳' });
    expect(within(table).getByRole('columnheader', { name: '得点項目' })).toBeInTheDocument();
    expect(within(table).getByRole('columnheader', { name: '獲得者' })).toBeInTheDocument();
    expect(within(table).getByRole('rowheader', { name: 'High' })).toBeInTheDocument();
    expect(panel).toHaveTextContent('High');
    expect(panel).toHaveTextContent('Jack');
    // Per-player Game pip totals are shown.
    expect(panel).toHaveTextContent('5 / 0');
  });

  it('shows "−" for the Jack row when no trump Jack was captured', async () => {
    const roundEndState: AllFoursResponse = {
      ...baseState,
      phase: AllFoursPhase.ROUND_END,
      roundBreakdown: {
        high: { winnerIdx: 0, card: { design: 'HEART', value: 5 } },
        low: { winnerIdx: 1, card: { design: 'HEART', value: 3 } },
        jack: { winnerIdx: -1 },
        game: { winnerIdx: -1, points: [0, 0] },
        provisional: false,
      },
    };
    mockExec.mockResolvedValueOnce(roundEndState);
    renderWithProviders(<AllFoursPage />);
    const panel = await screen.findByTestId('af-breakdown');
    // The Jack and Game rows both fall back to the em-dash placeholder.
    expect(panel.querySelectorAll('td')).not.toHaveLength(0);
    expect(panel).toHaveTextContent('−');
  });

  it('omits the breakdown panel outside round/game end', async () => {
    mockExec.mockResolvedValueOnce(playState);
    renderWithProviders(<AllFoursPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('af-breakdown')).not.toBeInTheDocument();
  });

  // **手札が急に増えて切り札も変わったのに、最初の Beg と同じ文面だった** (#6479)。
  // サーバは `allfours.begAfterRun` を返すようになったので、画面が**訳文に解決する**
  // ことを見る ── 生の識別子が出たら翻訳を通していない。
  it('spells out that the cards were run', async () => {
    mockExec.mockResolvedValue({
      ...baseState,
      messageCode: 'allfours.begAfterRun',
      messageParams: { count: '3' },
    });
    renderWithProviders(<AllFoursPage />);

    // 連続回数がそのまま出る。
    const line = await screen.findByText(/ランザカード 3 回/);
    expect(line).toBeInTheDocument();
    // 生の識別子も未置換のプレースホルダも残らない。
    expect(line.textContent).not.toContain('allfours.');
    expect(line.textContent).not.toContain('{{');
  });

  it('spells out the run on the first lead too', async () => {
    mockExec.mockResolvedValue({
      ...baseState,
      messageCode: 'allfours.playAfterRun',
      messageParams: { count: '1' },
    });
    renderWithProviders(<AllFoursPage />);

    const line = await screen.findByText(/切り札が変わりました/);
    expect(line).toBeInTheDocument();
    expect(line.textContent).not.toContain('{{');
  });
});

// **High/Low/Jack/Game はトリックが進むたびに途中経過が確定していくのに、
// ラウンド終了まで一切見えなかった (#4771)。**「今どちらが何を握っているか」は
// このゲームの戦略そのもの。
describe('AllFoursPage live breakdown', () => {
  const playState = (provisional: boolean): AllFoursResponse => ({
    ...baseState,
    phase: provisional ? AllFoursPhase.PLAY : AllFoursPhase.ROUND_END,
    roundBreakdown: {
      high: { winnerIdx: 0, card: { design: 'HEART', value: 12 } },
      low: { winnerIdx: 1, card: { design: 'HEART', value: 3 } },
      jack: { winnerIdx: -1 },
      game: { winnerIdx: 0, points: [5, 0] },
      provisional,
    },
  });

  it('shows the breakdown during the play phase', async () => {
    mockExec.mockResolvedValueOnce(playState(true));
    renderWithProviders(<AllFoursPage />);
    expect(await screen.findByTestId('af-breakdown')).toBeInTheDocument();
  });

  // **暫定値を確定値として見せない。**まだ出ていないトランプで High も Low も
  // 引っくり返る。
  it('labels a mid-round breakdown as provisional', async () => {
    mockExec.mockResolvedValueOnce(playState(true));
    renderWithProviders(<AllFoursPage />);
    expect(await screen.findByTestId('af-breakdown-provisional')).toBeInTheDocument();
  });

  it('does not label the settled breakdown as provisional', async () => {
    mockExec.mockResolvedValueOnce(playState(false));
    renderWithProviders(<AllFoursPage />);
    await screen.findByTestId('af-breakdown');
    expect(screen.queryByTestId('af-breakdown-provisional')).not.toBeInTheDocument();
  });
});
