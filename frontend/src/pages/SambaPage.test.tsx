import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { sambaApi } from '../api/gameApi';
import i18n from '../i18n';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeSambaState } from '../test/stateFactories';
import { SambaPage } from './SambaPage';

vi.mock('../api/gameApi', () => ({
  sambaApi: { exec: vi.fn() },
  actionLogApi: { samba: vi.fn() },
}));
const mockExec = vi.mocked(sambaApi.exec);

const drawPhaseState = makeSambaState();
const meldPhaseState = makeSambaState({ phase: 1, messageCode: 'samba.meldPhase' });
const discardPhaseState = makeSambaState({ phase: 2, messageCode: 'samba.discardPhase' });
const roundEndState = makeSambaState({ phase: 3, messageCode: 'samba.roundEnd' });
const gameEndState = makeSambaState({ phase: 4, gameEndFlag: true, winnerIdx: 0 });
// Not the human's turn — CPU (seat 1) is active during the draw phase.
const cpuTurnState = makeSambaState({ currentPlayerIdx: 1 });

describe('SambaPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(drawPhaseState);
  });

  afterEach(() => {
    localStorage.clear();
  });

  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<SambaPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<SambaPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1, pointLimit: 10000 }),
    );
  });

  it('shows draw phase buttons and team scores', async () => {
    renderWithProviders(<SambaPage />);
    const actions = await screen.findByTestId('game-footer-actions');
    expect(within(actions).getByRole('button', { name: '山札から引く' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument();
    expect(screen.getByTestId('sa-team-scores')).toBeInTheDocument();
    expect(screen.getByTestId('sa-meld-points')).toHaveAttribute('aria-live', 'polite');
  });

  it('announces only changed team scores after the initial state', async () => {
    renderWithProviders(<SambaPage />);
    const announce = await screen.findByTestId('sa-team-score-announce');
    expect(announce).toHaveAttribute('role', 'status');
    expect(announce).toHaveAttribute('aria-live', 'polite');
    expect(announce).toBeEmptyDOMElement();

    mockExec.mockResolvedValue(makeSambaState({ teamScores: [25, 0] }));
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(announce).toHaveTextContent('チーム0の得点は25点です'));

    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(screen.getByTestId('sa-team-scores')).toHaveTextContent('チーム0: 25'));
    expect(announce).toHaveTextContent('チーム0の得点は25点です');
  });

  it('announces when the discard pile becomes frozen', async () => {
    renderWithProviders(<SambaPage />);
    const announce = await screen.findByTestId('sa-frozen-announce');
    expect(announce).toHaveAttribute('role', 'status');
    expect(announce).toHaveAttribute('aria-live', 'polite');
    expect(announce).toHaveTextContent(''); // no transition yet
    // A draw resolves to a frozen state → isFrozen false→true triggers the announcement.
    mockExec.mockResolvedValue(makeSambaState({ isFrozen: true }));
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(screen.getByTestId('sa-frozen-announce')).toHaveTextContent('捨札が凍結されました'));
  });

  it('calls drawstock command when button clicked', async () => {
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(meldPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('hides action controls when it is not the human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '山札から引く' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '捨て札を取る' })).not.toBeInTheDocument();
  });

  it('shows meld phase buttons', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'メルドする' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'スキップ' })).toBeInTheDocument();
  });

  it('keeps unavailable meld actions focusable and ignores activation', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<SambaPage />);
    const addGroup = await screen.findByRole('button', { name: '選択カードをグループに追加' });
    const meld = screen.getByRole('button', { name: 'メルドする' });
    expect(addGroup).toHaveAttribute('aria-disabled', 'true');
    expect(meld).toHaveAttribute('aria-disabled', 'true');
    mockExec.mockClear();
    fireEvent.click(addGroup);
    fireEvent.click(meld);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
    expect(screen.queryByRole('region', { name: 'メルドグループ' })).not.toBeInTheDocument();
  });

  it('stages multiple selected groups and sends their card indices together', async () => {
    const sixCardMeldState = makeSambaState({
      phase: 1,
      messageCode: 'samba.meldPhase',
      players: [
        {
          ...meldPhaseState.players[0],
          cards: [
            { design: 'SPADE', value: 7 },
            { design: 'CLOVER', value: 7 },
            { design: 'HEART', value: 7 },
            { design: 'SPADE', value: 10 },
            { design: 'CLOVER', value: 10 },
            { design: 'HEART', value: 10 },
          ],
        },
        ...meldPhaseState.players.slice(1),
      ],
    });
    mockExec.mockResolvedValue(sixCardMeldState);
    renderWithProviders(<SambaPage />);
    const hand = await screen.findByRole('button', { name: 'メルドする' });
    const getCards = () => document.querySelectorAll('[data-tutorial="sa-player-hand"] button');
    for (const index of [0, 1, 2]) fireEvent.click(getCards()[index]);
    fireEvent.click(screen.getByRole('button', { name: '選択カードをグループに追加' }));
    fireEvent.click(getCards()[0]);
    expect(getCards()[0]).toHaveAttribute('aria-pressed', 'true');
    for (const index of [3, 4, 5]) fireEvent.click(getCards()[index]);
    fireEvent.click(screen.getByRole('button', { name: '選択カードをグループに追加' }));

    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(hand);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('meld', undefined, undefined, undefined, [
        [0, 1, 2],
        [3, 4, 5],
      ]),
    );
  });

  it('removes a staged meld group', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<SambaPage />);
    const getCards = () => document.querySelectorAll('[data-tutorial="sa-player-hand"] button');
    await screen.findByRole('button', { name: 'メルドする' });
    for (const index of [0, 1, 2]) fireEvent.click(getCards()[index]);
    fireEvent.click(screen.getByRole('button', { name: '選択カードをグループに追加' }));
    expect(screen.getByRole('region', { name: 'メルドグループ' })).toHaveTextContent('グループ1');
    fireEvent.click(screen.getByRole('button', { name: 'グループを取り消す' }));
    expect(screen.queryByRole('region', { name: 'メルドグループ' })).not.toBeInTheDocument();
  });

  it('does not remove a staged group while an API action is loading', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<SambaPage />);
    const getCards = () => document.querySelectorAll('[data-tutorial="sa-player-hand"] button');
    await screen.findByRole('button', { name: 'メルドする' });
    for (const index of [0, 1, 2]) fireEvent.click(getCards()[index]);
    fireEvent.click(screen.getByRole('button', { name: '選択カードをグループに追加' }));
    mockExec.mockReturnValue(new Promise(() => undefined));
    fireEvent.click(screen.getByRole('button', { name: 'メルドする' }));
    const remove = await screen.findByRole('button', { name: 'グループを取り消す' });
    await waitFor(() => expect(remove).toHaveAttribute('aria-disabled', 'true'));
    fireEvent.click(remove);
    expect(screen.getByRole('region', { name: 'メルドグループ' })).toHaveTextContent('グループ1');
  });

  it('shows the initial-meld minimum and selected total in the meld phase', async () => {
    mockExec.mockResolvedValue({ ...meldPhaseState, minMeld: 90 }); // the server's minimum wins; hasInitMeld false
    renderWithProviders(<SambaPage />);
    const info = await screen.findByTestId('sa-meld-points');
    expect(info).toHaveTextContent('初回メルド必要点: 90');
    expect(info).toHaveTextContent('選択合計: 0');
    expect(info).toHaveTextContent('あと90点不足しています');
    const hand = document.querySelector('[data-tutorial="sa-player-hand"] button');
    expect(hand).toHaveAttribute('aria-describedby', 'sa-meld-points');
  });

  it('does not describe meld points after the initial meld is complete', async () => {
    const stateAfterInitialMeld = makeSambaState({
      phase: 1,
      messageCode: 'samba.meldPhase',
      players: [{ ...meldPhaseState.players[0], hasInitMeld: true }, ...meldPhaseState.players.slice(1)],
    });
    mockExec.mockResolvedValue(stateAfterInitialMeld);
    renderWithProviders(<SambaPage />);

    const meldButton = await screen.findByRole('button', { name: 'メルドする' });
    const hand = document.querySelector('[data-tutorial="sa-player-hand"] button');
    expect(hand).not.toHaveAttribute('aria-describedby');
    expect(meldButton).not.toHaveAttribute('aria-describedby');
  });

  it('announces that the initial-meld requirement is met when selected points reach the minimum', async () => {
    const stateWithEnoughPoints = makeSambaState({
      phase: 1,
      messageCode: 'samba.meldPhase',
      players: [
        {
          ...meldPhaseState.players[0],
          cards: [
            { design: 'SPADE', value: 10 },
            { design: 'CLOVER', value: 10 },
            { design: 'HEART', value: 10 },
            { design: 'DIAMOND', value: 10 },
            { design: 'SPADE', value: 10 },
          ],
        },
        ...meldPhaseState.players.slice(1),
      ],
    });
    mockExec.mockResolvedValue(stateWithEnoughPoints);
    renderWithProviders(<SambaPage />);

    const handCards = await screen.findAllByRole('button', { name: /10/ });
    handCards.forEach((card) => {
      fireEvent.click(card);
    });

    expect(screen.getByTestId('sa-meld-points')).toHaveTextContent('初回メルド必要点: 50');
    expect(screen.getByTestId('sa-meld-points')).toHaveTextContent('選択合計: 50');
    expect(screen.getByTestId('sa-meld-points')).toHaveTextContent('必要点を満たしています');
    expect(screen.getByTestId('sa-meld-points')).not.toHaveTextContent('不足しています');
  });

  it('does not describe an empty meld-points region during a CPU meld turn', async () => {
    mockExec.mockResolvedValue(makeSambaState({ phase: 1, currentPlayerIdx: 1, messageCode: 'samba.meldPhase' }));
    renderWithProviders(<SambaPage />);

    await screen.findByTestId('sa-meld-points');
    const hand = document.querySelector('[data-tutorial="sa-player-hand"] button');
    expect(hand).not.toHaveAttribute('aria-describedby');
    expect(screen.getByTestId('sa-meld-points')).toHaveTextContent('');
  });

  it('calls skipmeld command when skip button clicked', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スキップ' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('skipmeld'));
  });

  it('shows discard phase buttons', async () => {
    mockExec.mockResolvedValue(
      makeSambaState({
        phase: 2,
        messageCode: 'samba.discardPhase',
        players: makeSambaState().players.map((p) => (p.isHuman ? { ...p, cards: [], cardCount: 0 } : p)),
      }),
    );
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨てる' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '上がる' })).toBeInTheDocument();
    expect(screen.getByTestId('sa-go-out-progress')).toHaveTextContent('完成メルド 0/2');
    expect(screen.getByRole('button', { name: '上がる' })).toHaveAttribute(
      'title',
      '上がるにはチームで完成メルドが2個必要です',
    );
  });

  it.each([
    [0, '完成メルド 0/2', '上がるにはチームで完成メルドが2個必要です'],
    [1, '完成メルド 1/2', '上がるにはチームで完成メルドが2個必要です'],
    [2, '完成メルド 2/2', undefined],
  ] as const)('shows go-out progress for %d completed melds', async (completed, progress, title) => {
    mockExec.mockResolvedValue(
      makeSambaState({
        phase: 2,
        completedMelds: [completed, 0],
        players: makeSambaState().players.map((p) => (p.isHuman ? { ...p, cards: [], cardCount: 0 } : p)),
      }),
    );
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(screen.getByTestId('sa-go-out-progress')).toHaveTextContent(progress));
    const goOut = screen.getByRole('button', { name: '上がる' });
    if (title) {
      expect(goOut).toHaveAttribute('title', title);
    } else {
      expect(goOut).toBeEnabled();
      fireEvent.click(goOut);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('goout'));
    }
  });

  it.each([
    {
      cards: [
        { design: 'SPADE' as const, value: 4 },
        { design: 'SPADE' as const, value: 5 },
      ],
      expected: '上がるには手札を1枚以下にしてください',
    },
    { cards: [{ design: 'HEART' as const, value: 3 }], expected: '赤3は捨てられないため上がれません' },
  ])('explains why go-out is unavailable for the hand', async ({ cards, expected }) => {
    const base = makeSambaState();
    mockExec.mockResolvedValue(
      makeSambaState({
        phase: 2,
        completedMelds: [2, 0],
        players: base.players.map((p) => (p.isHuman ? { ...p, cards, cardCount: cards.length } : p)),
      }),
    );
    renderWithProviders(<SambaPage />);
    const goOut = await screen.findByRole('button', { name: '上がる' });
    expect(goOut).toHaveAttribute('aria-disabled', 'true');
    expect(screen.getByTestId('sa-go-out-guidance')).toHaveTextContent(expected);
    mockExec.mockClear();
    fireEvent.click(goOut);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('allows going out with no cards or one non-red-three when melds are complete', async () => {
    const base = makeSambaState();
    for (const cards of [[], [{ design: 'SPADE' as const, value: 4 }]]) {
      mockExec.mockResolvedValue(
        makeSambaState({
          phase: 2,
          completedMelds: [2, 0],
          players: base.players.map((p) => (p.isHuman ? { ...p, cards, cardCount: cards.length } : p)),
        }),
      );
      const { unmount } = renderWithProviders(<SambaPage />);
      const goOut = await screen.findByRole('button', { name: '上がる' });
      expect(goOut).toHaveAttribute('aria-disabled', 'false');
      fireEvent.click(goOut);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('goout'));
      unmount();
      mockExec.mockClear();
    }
  });

  it('uses the server-provided go-out requirement for progress and tooltip', async () => {
    mockExec.mockResolvedValue(
      makeSambaState({
        phase: 2,
        completedMelds: [2, 0],
        players: makeSambaState().players.map((p) => (p.isHuman ? { ...p, cards: [], cardCount: 0 } : p)),
        config: { ...makeSambaState().config, goOutRequiredMelds: 3 },
      }),
    );
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(screen.getByTestId('sa-go-out-progress')).toHaveTextContent('完成メルド 2/3'));
    expect(screen.getByRole('button', { name: '上がる' })).toHaveAttribute(
      'title',
      '上がるにはチームで完成メルドが3個必要です',
    );
  });

  it('shows next round button at round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
  });

  it('renders hint toggle checkbox', async () => {
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());
    expect(screen.getByRole('checkbox', { name: 'ヒント表示' })).toBeInTheDocument();
  });

  it('shows HintTooltip when hint is enabled in draw phase', async () => {
    localStorage.setItem('hint_enabled_samba', 'true');
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('shows 次のゲーム at game-end and fires reset directly (no confirm)', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1, pointLimit: 10000 }),
    );
    expect(screen.queryByRole('button', { name: '確認' })).not.toBeInTheDocument();
  });

  it('shows the disabled reason for draw-from-discard when no cards are selected', async () => {
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument());
    expect(screen.getByTestId('sa-draw-discard-reason')).toHaveTextContent(
      '手札からトップカードと同ランクの2枚を選択してください',
    );
  });

  it('renders the frozen badge and reason when the discard pile is frozen', async () => {
    mockExec.mockResolvedValue(makeSambaState({ isFrozen: true }));
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(screen.getByTestId('sa-frozen-badge')).toBeInTheDocument());
    expect(screen.getByTestId('sa-discard-pile')).toHaveClass(
      'bg-ds-surface',
      'border-ds-border-subtle',
      'ring-2',
      'ring-ds-info',
    );
    expect(screen.getByTestId('sa-draw-discard-reason')).toHaveTextContent(
      'フリーズ中はワイルドカードでの代用ができません',
    );
  });

  it('enables the draw button and clears the reason when exactly 2 cards are selected', async () => {
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument());
    const handCards = screen.getAllByRole('button', { pressed: false }).filter((b) => b.hasAttribute('aria-pressed'));
    fireEvent.click(handCards[0]);
    fireEvent.click(handCards[1]);
    expect(screen.queryByTestId('sa-draw-discard-reason')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: '捨て札を取る' })).not.toBeDisabled();
  });

  it('shows canasta/samba progress and a completion pulse per meld', async () => {
    const base = makeSambaState();
    const human = {
      ...base.players[0],
      melds: [
        // Incomplete set: 5 cards → 2 more for a canasta.
        {
          cards: Array.from({ length: 5 }, () => ({ design: 'SPADE' as const, value: 4 })),
          kind: 0,
          isNatural: true,
          isCanasta: false,
          isSamba: false,
          rank: 4,
        },
        // Completed 7-card sequence → samba, with pulse emphasis.
        {
          cards: Array.from({ length: 7 }, (_, i) => ({ design: 'HEART' as const, value: i + 3 })),
          kind: 1,
          isNatural: true,
          isCanasta: false,
          isSamba: true,
          rank: 3,
        },
      ],
    };
    mockExec.mockResolvedValue(makeSambaState({ players: [human, ...base.players.slice(1)] }));
    renderWithProviders(<SambaPage />);

    const setProgress = await screen.findByTestId('sa-meld-progress-0-0');
    expect(setProgress).toHaveTextContent('あと2枚でカナスタ');
    expect(setProgress.className).not.toContain('animate-pulse');

    const sambaProgress = screen.getByTestId('sa-meld-progress-0-1');
    expect(sambaProgress).toHaveTextContent('サンバ成立！');
    expect(sambaProgress.className).toContain('animate-pulse');
  });

  it('announces human-team meld progress updates in Japanese', async () => {
    const base = makeSambaState();
    const initial = makeSambaState({
      players: [
        {
          ...base.players[0],
          melds: [
            {
              cards: Array.from({ length: 4 }, () => ({ design: 'SPADE' as const, value: 4 })),
              kind: 0,
              isNatural: true,
              isCanasta: false,
              isSamba: false,
              rank: 4,
            },
          ],
        },
        ...base.players.slice(1),
      ],
    });
    const updated = makeSambaState({
      players: [
        {
          ...initial.players[0],
          melds: [
            {
              ...initial.players[0].melds[0],
              cards: [...initial.players[0].melds[0].cards, { design: 'HEART', value: 4 }],
            },
          ],
        },
        ...initial.players.slice(1),
      ],
    });
    const completed = makeSambaState({
      players: [
        {
          ...updated.players[0],
          melds: [
            {
              ...updated.players[0].melds[0],
              cards: Array.from({ length: 7 }, () => ({ design: 'SPADE' as const, value: 4 })),
            },
          ],
        },
        ...updated.players.slice(1),
      ],
    });
    mockExec.mockResolvedValueOnce(initial).mockResolvedValueOnce(updated).mockResolvedValueOnce(completed);
    renderWithProviders(<SambaPage />);

    const announce = await screen.findByTestId('sa-meld-progress-announce');
    expect(announce).toHaveAttribute('role', 'status');
    expect(announce).toHaveAttribute('aria-live', 'polite');
    expect(announce).toHaveTextContent('');
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(announce).toHaveTextContent('あなたのチームのメルド: あと2枚でカナスタ'));
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(announce).toHaveTextContent('あなたのチームのメルド: カナスタ成立！'));
  });

  it('announces human-team meld progress updates in English', async () => {
    const base = makeSambaState();
    const initial = makeSambaState({
      players: [
        {
          ...base.players[0],
          melds: [
            {
              cards: Array.from({ length: 4 }, () => ({ design: 'SPADE' as const, value: 4 })),
              kind: 0,
              isNatural: true,
              isCanasta: false,
              isSamba: false,
              rank: 4,
            },
          ],
        },
        ...base.players.slice(1),
      ],
    });
    const updated = makeSambaState({
      players: [
        {
          ...initial.players[0],
          melds: [
            {
              ...initial.players[0].melds[0],
              cards: [...initial.players[0].melds[0].cards, { design: 'HEART', value: 4 }],
            },
          ],
        },
        ...initial.players.slice(1),
      ],
    });
    mockExec.mockResolvedValueOnce(initial).mockResolvedValueOnce(updated);
    await i18n.changeLanguage('en');
    try {
      renderWithProviders(<SambaPage />);
      const announce = await screen.findByTestId('sa-meld-progress-announce');
      fireEvent.click(screen.getByRole('button', { name: 'Draw from stock' }));
      await waitFor(() => expect(announce).toHaveTextContent("Your team's meld: 2 more for Canasta"));
    } finally {
      await i18n.changeLanguage('ja');
    }
  });

  it('does not announce opponent meld updates', async () => {
    const base = makeSambaState();
    const initial = makeSambaState({ players: [base.players[0], { ...base.players[1], melds: [] }] });
    const opponentUpdate = makeSambaState({
      players: [
        base.players[0],
        {
          ...base.players[1],
          melds: [
            {
              cards: Array.from({ length: 5 }, () => ({ design: 'SPADE' as const, value: 4 })),
              kind: 0,
              isNatural: true,
              isCanasta: false,
              isSamba: false,
              rank: 4,
            },
          ],
        },
      ],
    });
    mockExec.mockResolvedValueOnce(initial).mockResolvedValueOnce(opponentUpdate);
    renderWithProviders(<SambaPage />);

    const announce = await screen.findByTestId('sa-meld-progress-announce');
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
    expect(announce).toHaveTextContent('');
  });

  it('localizes the CPU hand-count label at round end', async () => {
    // The label is split across text nodes, so match on the row's textContent.
    const cpuRows = (re: RegExp) =>
      screen.queryAllByText((_, el) => el?.tagName === 'DIV' && re.test(el.textContent ?? ''));
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(cpuRows(/^CPU \d+: \d+枚$/).length).toBeGreaterThan(0));
    expect(cpuRows(/^CPU \d+: \d+ cards$/)).toHaveLength(0);
  });

  // 捨て札の一番上のカードが存在する場合のみ、捨て札置き場を表示する。
  it('renders sa-discard-pile when discardTop exists', async () => {
    mockExec.mockResolvedValue(drawPhaseState);
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(screen.getByTestId('sa-discard-pile')).toBeInTheDocument());
  });

  it('does not render sa-discard-pile when discardTop is null', async () => {
    mockExec.mockResolvedValue(makeSambaState({ discardTop: null }));
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(screen.getByTestId('sa-team-scores')).toBeInTheDocument());
    expect(screen.queryByTestId('sa-discard-pile')).not.toBeInTheDocument();
  });
});
