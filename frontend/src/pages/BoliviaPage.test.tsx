import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { boliviaApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeBoliviaState } from '../test/stateFactories';
import { BoliviaPage } from './BoliviaPage';

const mobileState = vi.hoisted(() => ({ isMobile: true }));

vi.mock('../hooks/useCardDimensions', async (importOriginal) => ({
  ...(await importOriginal<typeof import('../hooks/useCardDimensions')>()),
  useIsMobile: () => mobileState.isMobile,
}));

vi.mock('../api/gameApi', () => ({
  boliviaApi: { exec: vi.fn() },
  actionLogApi: { bolivia: vi.fn() },
}));
const mockExec = vi.mocked(boliviaApi.exec);

const drawPhaseState = makeBoliviaState();
const meldPhaseState = makeBoliviaState({ phase: 1, messageCode: 'bolivia.meldPhase' });
const discardPhaseState = makeBoliviaState({ phase: 2, messageCode: 'bolivia.discardPhase' });
const roundEndState = makeBoliviaState({ phase: 3, messageCode: 'bolivia.roundEnd' });
const gameEndState = makeBoliviaState({ phase: 4, gameEndFlag: true, winnerIdx: 0 });
// Not the human's turn — CPU (seat 1) is active during the draw phase.
const cpuTurnState = makeBoliviaState({ currentPlayerIdx: 1 });

describe('BoliviaPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(drawPhaseState);
  });

  afterEach(() => {
    localStorage.clear();
    mobileState.isMobile = true;
  });

  it('collapses the rule and score details on mobile while keeping team scores visible', async () => {
    renderWithProviders(<BoliviaPage />);

    const rule = await screen.findByTestId('bo-goout-rule');
    const scoreTable = screen.getByTestId('bo-score-table');
    expect(rule).not.toHaveAttribute('open');
    expect(within(rule).getByText('上がるには完成メルド 2 個以上＋そのうち最低 1 本がエスカレラ')).not.toBeVisible();
    expect(scoreTable).not.toHaveAttribute('open');
    expect(within(scoreTable).getByRole('table')).not.toBeVisible();
    expect(screen.getByTestId('sa-team-scores')).toBeVisible();
  });

  it('opens the score details on desktop', async () => {
    mobileState.isMobile = false;
    renderWithProviders(<BoliviaPage />);

    const scoreTable = await screen.findByTestId('bo-score-table');
    expect(scoreTable).toHaveAttribute('open');
    expect(within(scoreTable).getByRole('table')).toBeVisible();
  });

  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<BoliviaPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<BoliviaPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1, pointLimit: 15000 }),
    );
  });

  it('labels each team score as yours or the opponent while preserving team numbers', async () => {
    const state = makeBoliviaState({
      players: makeBoliviaState().players.map((player) => (player.isHuman ? { ...player, team: 1 } : player)),
      teamScores: [123, 456],
    });
    mockExec.mockResolvedValue(state);
    renderWithProviders(<BoliviaPage />);

    expect(await screen.findByTestId('sa-team-scores')).toHaveTextContent(
      'チーム0（相手チーム）: 123 / チーム1（自チーム）: 456',
    );
  });

  it('shows draw phase buttons and team scores', async () => {
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());
    expect(
      within(screen.getByTestId('game-footer-actions')).getByRole('button', { name: '山札から引く' }),
    ).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument();
    expect(screen.getByTestId('sa-team-scores')).toBeInTheDocument();
  });

  it('announces when the discard pile becomes frozen', async () => {
    renderWithProviders(<BoliviaPage />);
    const announce = await screen.findByTestId('sa-frozen-announce');
    expect(announce).toHaveAttribute('role', 'status');
    expect(announce).toHaveAttribute('aria-live', 'polite');
    expect(announce).toHaveTextContent(''); // no transition yet
    // A draw resolves to a frozen state → isFrozen false→true triggers the announcement.
    mockExec.mockResolvedValue(makeBoliviaState({ isFrozen: true }));
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(screen.getByTestId('sa-frozen-announce')).toHaveTextContent('捨札が凍結されました'));
  });

  it('calls drawstock command when button clicked', async () => {
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(meldPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('hides action controls when it is not the human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '山札から引く' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '捨て札を取る' })).not.toBeInTheDocument();
  });

  it('shows meld phase buttons', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'メルドする' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'スキップ' })).toBeInTheDocument();
  });

  it('sends staged hand selections as separate meld groups in one action', async () => {
    const state = makeBoliviaState({
      phase: 1,
      players: makeBoliviaState().players.map((player) =>
        player.isHuman
          ? {
              ...player,
              cardCount: 6,
              cards: Array.from({ length: 6 }, (_, value) => ({ design: 'SPADE' as const, value })),
            }
          : player,
      ),
    });
    mockExec.mockResolvedValue(state);
    renderWithProviders(<BoliviaPage />);
    await screen.findByRole('button', { name: 'メルドする' });
    const handCards = screen
      .getAllByRole('button', { pressed: false })
      .filter((button) => button.hasAttribute('aria-pressed'));
    for (const card of handCards.slice(0, 3)) fireEvent.click(card);
    fireEvent.click(screen.getByRole('button', { name: 'グループを追加' }));
    expect(screen.getByTestId('sa-meld-groups')).toHaveTextContent('グループ1（3枚）');
    for (const card of handCards.slice(3, 6)) fireEvent.click(card);
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'メルドする' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('meld', undefined, undefined, undefined, [
        [0, 1, 2],
        [3, 4, 5],
      ]),
    );
  });

  it('allows deselecting a staged meld group', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<BoliviaPage />);
    await screen.findByRole('button', { name: 'メルドする' });
    const handCards = screen
      .getAllByRole('button', { pressed: false })
      .filter((button) => button.hasAttribute('aria-pressed'));
    for (const card of handCards.slice(0, 3)) fireEvent.click(card);
    fireEvent.click(screen.getByRole('button', { name: 'グループを追加' }));
    fireEvent.click(screen.getByRole('button', { name: 'グループ1（3枚）を選択解除' }));
    expect(screen.queryByTestId('sa-meld-groups')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'メルドする' })).toBeDisabled();
  });

  it('does not select a card already in a staged group and makes it selectable after removal', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<BoliviaPage />);
    await screen.findByRole('button', { name: 'メルドする' });
    const handCards = screen
      .getAllByRole('button', { pressed: false })
      .filter((button) => button.hasAttribute('aria-pressed'));
    for (const card of handCards.slice(0, 3)) fireEvent.click(card);
    fireEvent.click(screen.getByRole('button', { name: 'グループを追加' }));

    const groupedCard = handCards[0];
    expect(groupedCard).toHaveAttribute('aria-disabled', 'true');
    expect(groupedCard).toHaveAccessibleName(/グループ1に追加済み/);
    fireEvent.click(groupedCard);
    expect(screen.getByRole('button', { name: 'グループを追加' })).toBeDisabled();

    fireEvent.click(screen.getByRole('button', { name: 'グループ1（3枚）を選択解除' }));
    expect(groupedCard).not.toHaveAttribute('aria-disabled');
    fireEvent.click(groupedCard);
    expect(screen.getByRole('button', { name: 'グループを追加' })).toBeDisabled();
    fireEvent.click(handCards[1]);
    fireEvent.click(handCards[2]);
    expect(screen.getByRole('button', { name: 'グループを追加' })).toBeEnabled();
  });

  it('keeps staged groups and shows the server error when a meld is rejected', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<BoliviaPage />);
    await screen.findByRole('button', { name: 'メルドする' });
    const handCards = screen
      .getAllByRole('button', { pressed: false })
      .filter((button) => button.hasAttribute('aria-pressed'));
    for (const card of handCards.slice(0, 3)) fireEvent.click(card);
    fireEvent.click(screen.getByRole('button', { name: 'グループを追加' }));
    mockExec.mockRejectedValueOnce(new Error('invalid meld group'));
    fireEvent.click(screen.getByRole('button', { name: 'メルドする' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('通信エラーが発生しました。もう一度お試しください。');
    expect(screen.getByTestId('sa-meld-groups')).toHaveTextContent('グループ1（3枚）');
  });

  it('shows the initial-meld minimum and selected total in the meld phase', async () => {
    mockExec.mockResolvedValue({ ...meldPhaseState, minMeld: 90 }); // team score 0 → min 50; hasInitMeld false
    renderWithProviders(<BoliviaPage />);
    const info = await screen.findByTestId('sa-meld-points');
    expect(info).toHaveTextContent('初回メルド最低点: 90');
    expect(info).toHaveTextContent('選択合計: 0');
  });

  it('calls skipmeld command when skip button clicked', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スキップ' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('skipmeld'));
  });

  it('shows discard phase buttons', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨てる' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '上がる' })).toBeInTheDocument();
  });

  it('disables go out when the server says it is unavailable', async () => {
    mockExec.mockResolvedValue(makeBoliviaState({ phase: 2, canGoOut: false }));
    renderWithProviders(<BoliviaPage />);
    const button = await screen.findByRole('button', { name: '上がる' });
    expect(button).toBeDisabled();
  });

  it('enables go out when the server says it is available', async () => {
    mockExec.mockResolvedValue(makeBoliviaState({ phase: 2, canGoOut: true }));
    renderWithProviders(<BoliviaPage />);
    const button = await screen.findByRole('button', { name: '上がる' });
    expect(button).toBeEnabled();
  });

  it('shows next round button at round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
  });

  it('renders hint toggle checkbox', async () => {
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());
    expect(screen.getByRole('checkbox', { name: 'ヒント表示' })).toBeInTheDocument();
  });

  it('shows HintTooltip when hint is enabled in draw phase', async () => {
    localStorage.setItem('hint_enabled_bolivia', 'true');
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('shows 次のゲーム at game-end and fires reset directly (no confirm)', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1, pointLimit: 15000 }),
    );
    expect(screen.queryByRole('button', { name: '確認' })).not.toBeInTheDocument();
  });

  it('shows the disabled reason for draw-from-discard when no cards are selected', async () => {
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument());
    expect(screen.getByTestId('sa-draw-discard-reason')).toHaveTextContent(
      '手札からトップカードと同ランクの2枚を選択してください',
    );
  });

  it('renders the frozen badge and reason when the discard pile is frozen', async () => {
    mockExec.mockResolvedValue(makeBoliviaState({ isFrozen: true }));
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(screen.getByTestId('sa-frozen-badge')).toBeInTheDocument());
    expect(screen.getByTestId('sa-draw-discard-reason')).toHaveTextContent(
      'フリーズ中はワイルドカードでの代用ができません',
    );
  });

  it('explains when a natural pair matches the discard top', async () => {
    mockExec.mockResolvedValue(makeBoliviaState({ discardTop: { design: 'SPADE', value: 7 } }));
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument());
    const handCards = screen.getAllByRole('button', { pressed: false }).filter((b) => b.hasAttribute('aria-pressed'));
    fireEvent.click(handCards[0]);
    fireEvent.click(handCards[1]);
    const button = screen.getByRole('button', { name: '捨て札を取る' });
    expect(screen.getByTestId('sa-draw-discard-reason')).toHaveTextContent('ナチュラルカード2枚がトップと同ランクです');
    expect(button).toHaveAttribute('aria-disabled', 'false');
    mockExec.mockClear();
    fireEvent.click(button);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawdiscard', undefined, undefined, [0, 1]));
  });

  it('blocks a selected pair with the wrong rank and shows why', async () => {
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument());
    const handCards = screen.getAllByRole('button', { pressed: false }).filter((b) => b.hasAttribute('aria-pressed'));
    fireEvent.click(handCards[0]);
    fireEvent.click(handCards[1]);
    const button = screen.getByRole('button', { name: '捨て札を取る' });
    expect(screen.getByTestId('sa-draw-discard-reason')).toHaveTextContent('同じランクである必要があります');
    expect(button).toHaveAttribute('aria-disabled', 'true');
    mockExec.mockClear();
    fireEvent.click(button);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('explains why a pair cannot take an empty or unusable discard pile', async () => {
    const handCards = () =>
      screen.getAllByRole('button', { pressed: false }).filter((button) => button.hasAttribute('aria-pressed'));

    mockExec.mockResolvedValue(makeBoliviaState({ discardTop: null }));
    const emptyPile = renderWithProviders(<BoliviaPage />);
    await screen.findByRole('button', { name: '捨て札を取る' });
    fireEvent.click(handCards()[0]);
    fireEvent.click(handCards()[1]);
    expect(screen.getByTestId('sa-draw-discard-reason')).toHaveTextContent('捨て札のトップがないため取得できません');

    emptyPile.unmount();
    mockExec.mockResolvedValue(makeBoliviaState({ discardTop: { design: 'SPADE', value: 2 } }));
    renderWithProviders(<BoliviaPage />);
    await screen.findByRole('button', { name: '捨て札を取る' });
    fireEvent.click(handCards()[0]);
    fireEvent.click(handCards()[1]);
    expect(screen.getByTestId('sa-draw-discard-reason')).toHaveTextContent(
      'トップがワイルドカードまたは黒3のため取得できません',
    );
  });

  it.each([
    [{ design: 'JOKER' as const, value: 0 }, 'joker top'],
    [{ design: 'CLOVER' as const, value: 3 }, 'black three top'],
  ])('explains why a pair cannot take a discard pile with a %s', async (discardTop, _caseName) => {
    mockExec.mockResolvedValue(makeBoliviaState({ discardTop }));
    renderWithProviders(<BoliviaPage />);
    await screen.findByRole('button', { name: '捨て札を取る' });
    const handCards = screen
      .getAllByRole('button', { pressed: false })
      .filter((button) => button.hasAttribute('aria-pressed'));
    fireEvent.click(handCards[0]);
    fireEvent.click(handCards[1]);
    expect(screen.getByTestId('sa-draw-discard-reason')).toHaveTextContent(
      'トップがワイルドカードまたは黒3のため取得できません',
    );
  });

  it('explains that a wild pair cannot take an unfrozen discard pile', async () => {
    const base = makeBoliviaState({ discardTop: { design: 'SPADE', value: 7 } });
    const human = {
      ...base.players[0],
      cards: [
        { design: 'JOKER' as const, value: 0 },
        { design: 'SPADE' as const, value: 7 },
      ],
    };
    mockExec.mockResolvedValue(makeBoliviaState({ ...base, players: [human, ...base.players.slice(1)] }));
    renderWithProviders(<BoliviaPage />);
    await screen.findByRole('button', { name: '捨て札を取る' });
    const handCards = screen
      .getAllByRole('button', { pressed: false })
      .filter((button) => button.hasAttribute('aria-pressed'));
    fireEvent.click(handCards[0]);
    fireEvent.click(handCards[1]);
    expect(screen.getByTestId('sa-draw-discard-reason')).toHaveTextContent(
      'ワイルドカードは使えません。ナチュラルカードを2枚選択してください',
    );
  });

  it('blocks wild cards in a frozen pair and explains the frozen restriction', async () => {
    const base = makeBoliviaState({ isFrozen: true, discardTop: { design: 'SPADE', value: 7 } });
    const human = {
      ...base.players[0],
      cards: [
        { design: 'JOKER' as const, value: 0 },
        { design: 'SPADE' as const, value: 7 },
      ],
    };
    mockExec.mockResolvedValue(makeBoliviaState({ ...base, players: [human, ...base.players.slice(1)] }));
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument());
    const handCards = screen.getAllByRole('button', { pressed: false }).filter((b) => b.hasAttribute('aria-pressed'));
    fireEvent.click(handCards[0]);
    fireEvent.click(handCards[1]);
    expect(screen.getByTestId('sa-draw-discard-reason')).toHaveTextContent('フリーズ中は2枚ともナチュラルカード');
    expect(screen.getByRole('button', { name: '捨て札を取る' })).toHaveAttribute('aria-disabled', 'true');
  });

  it('shows canasta/bolivia progress and a completion pulse per meld', async () => {
    const base = makeBoliviaState();
    const human = {
      ...base.players[0],
      melds: [
        // Incomplete set: 5 cards → 2 more for a canasta.
        {
          cards: Array.from({ length: 5 }, () => ({ design: 'SPADE' as const, value: 4 })),
          kind: 0,
          isNatural: true,
          isCanasta: false,
          isEscalera: false,
          isBolivia: false,
          rank: 4,
        },
        // Completed 7-card sequence → bolivia, with pulse emphasis.
        {
          cards: Array.from({ length: 7 }, (_, i) => ({ design: 'HEART' as const, value: i + 3 })),
          kind: 1,
          isNatural: true,
          isCanasta: false,
          isEscalera: false,
          isBolivia: true,
          rank: 3,
        },
      ],
    };
    mockExec.mockResolvedValue(makeBoliviaState({ players: [human, ...base.players.slice(1)] }));
    renderWithProviders(<BoliviaPage />);

    const setProgress = await screen.findByTestId('sa-meld-progress-0-0');
    expect(setProgress).toHaveTextContent('あと2枚でカナスタ');
    expect(setProgress.className).not.toContain('animate-pulse');

    const escaleraProgress = screen.getByTestId('sa-meld-progress-0-1');
    // **7 枚のシーケンスはエスカレラ。** ボリビアはワイルド 7 枚のほう。
    expect(escaleraProgress).toHaveTextContent('エスカレラ成立！');
    expect(escaleraProgress.className).toContain('animate-pulse');
  });

  it('localizes the CPU hand-count label at round end', async () => {
    // The label is split across text nodes, so match on the row's textContent.
    const cpuRows = (re: RegExp) =>
      screen.queryAllByText((_, el) => el?.tagName === 'DIV' && re.test(el.textContent ?? ''));
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(cpuRows(/^CPU \d+: \d+枚$/).length).toBeGreaterThan(0));
    expect(cpuRows(/^CPU \d+: \d+ cards$/)).toHaveLength(0);
  });

  // **メルドは 3 種類あり、進捗の言い回しも 3 つ。**
  //
  // クローン元 (Samba) はセットとシーケンスの 2 分岐しか持たないので、
  // そのままだと**ワイルドだけのメルドが「カナスタ」と表示される** ── この
  // ゲームの名前になっている役が、別の役の名前で出てしまう。
  it('names each of the three meld kinds, including the wild-only one', async () => {
    const base = makeBoliviaState();
    const seven = (design: 'SPADE' | 'HEART', from: number) =>
      Array.from({ length: 7 }, (_, i) => ({ design, value: from + i }));
    mockExec.mockResolvedValue(
      makeBoliviaState({
        players: base.players.map((p, i) =>
          i === 0
            ? {
                ...p,
                melds: [
                  {
                    cards: seven('SPADE', 4),
                    kind: 1,
                    isNatural: true,
                    isCanasta: false,
                    isEscalera: true,
                    isBolivia: false,
                    rank: 4,
                  },
                  {
                    cards: [
                      { design: 'SPADE', value: 2 },
                      { design: 'HEART', value: 2 },
                      { design: 'JOKER', value: 0 },
                      { design: 'CLOVER', value: 2 },
                      { design: 'DIAMOND', value: 2 },
                      { design: 'JOKER', value: 0 },
                      { design: 'JOKER', value: 0 },
                    ],
                    kind: 2,
                    isNatural: false,
                    isCanasta: false,
                    isEscalera: false,
                    isBolivia: true,
                    rank: 0,
                  },
                ],
              }
            : p,
        ),
      }),
    );
    renderWithProviders(<BoliviaPage />);
    expect(await screen.findByTestId('sa-meld-progress-0-0')).toHaveTextContent('エスカレラ成立！');
    const wild = screen.getByTestId('sa-meld-progress-0-1');
    expect(wild).toHaveTextContent('ボリビア成立！');
    // 負のコントロール: ワイルドのメルドをカナスタと呼ばない。
    expect(wild).not.toHaveTextContent('カナスタ');
  });

  // **ラベルの側も 3 分岐であること (レビュー指摘)。** 進捗表示だけを直して
  // ラベルを 2 分岐のまま残していたので、完成したボリビアが枚数表示に
  // 落ちて、画面のどこにも「ボリビア」と出ていなかった。
  it('labels a completed wild meld a bolivia in the meld label, not just the progress', async () => {
    const base = makeBoliviaState();
    mockExec.mockResolvedValue(
      makeBoliviaState({
        players: base.players.map((p, i) =>
          i === 0
            ? {
                ...p,
                melds: [
                  {
                    cards: [
                      { design: 'SPADE', value: 2 },
                      { design: 'HEART', value: 2 },
                      { design: 'CLOVER', value: 2 },
                      { design: 'DIAMOND', value: 2 },
                      { design: 'JOKER', value: 0 },
                      { design: 'JOKER', value: 0 },
                      { design: 'JOKER', value: 0 },
                    ],
                    kind: 2,
                    isNatural: false,
                    isCanasta: false,
                    isEscalera: false,
                    isBolivia: true,
                    rank: 0,
                  },
                ],
              }
            : p,
        ),
      }),
    );
    renderWithProviders(<BoliviaPage />);
    const area = await screen.findByTestId('sa-meld-progress-0-0');
    const row = area.parentElement as HTMLElement;
    expect(row).toHaveTextContent('ボリビア');
    // 負のコントロール: 枚数だけの表示に落ちていないこと。
    expect(row).not.toHaveTextContent('(7)');
  });

  // **席の印もエスカレラとボリビアを別に出す。**
  it('marks the escalera and the bolivia with distinct seat badges', async () => {
    const base = makeBoliviaState();
    mockExec.mockResolvedValue(
      makeBoliviaState({
        // 席の枠はメルドか赤3があるときだけ描かれるので、エスカレラを 1 本置く。
        players: base.players.map((p, i) =>
          i === 0
            ? {
                ...p,
                hasEscalera: true,
                hasBolivia: false,
                melds: [
                  {
                    cards: Array.from({ length: 7 }, (_, k) => ({ design: 'HEART' as const, value: 4 + k })),
                    kind: 1,
                    isNatural: true,
                    isCanasta: false,
                    isEscalera: true,
                    isBolivia: false,
                    rank: 4,
                  },
                ],
              }
            : p,
        ),
      }),
    );
    renderWithProviders(<BoliviaPage />);
    expect(await screen.findByTestId('bo-tag-escalera-0')).toBeInTheDocument();
    expect(screen.queryByTestId('bo-tag-bolivia-0')).not.toBeInTheDocument();
  });

  // **赤3は初回メルド未達なら罰点化する (#6634)**
  it('warns about red 3 penalty when team has not completed initial meld', async () => {
    const base = makeBoliviaState();
    mockExec.mockResolvedValue(
      makeBoliviaState({
        players: base.players.map((p, i) =>
          i === 0
            ? {
                ...p,
                hasInitMeld: false,
                red3s: [{ design: 'HEART' as const, value: 3 }],
                red3Count: 1,
              }
            : p,
        ),
      }),
    );
    renderWithProviders(<BoliviaPage />);
    const warning = await screen.findByTestId('bo-red3-warning-0');
    expect(warning).toHaveTextContent('初回メルド未達なら罰点');
    expect(warning).toHaveClass('text-ds-warning');
  });

  // 負のコントロール 1: 自身が初回メルド済みなら警告は出ない (受け入れ条件2)
  it('does not warn about red 3 penalty once the player has melded', async () => {
    const base = makeBoliviaState();
    mockExec.mockResolvedValue(
      makeBoliviaState({
        players: base.players.map((p, i) =>
          i === 0
            ? {
                ...p,
                hasInitMeld: true,
                red3s: [{ design: 'HEART' as const, value: 3 }],
                red3Count: 1,
              }
            : p,
        ),
      }),
    );
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(screen.getByText('赤3')).toBeInTheDocument());
    expect(screen.queryByTestId('bo-red3-warning-0')).not.toBeInTheDocument();
    expect(screen.queryByText(/初回メルド未達なら罰点/)).not.toBeInTheDocument();
  });

  // 負のコントロール 2: パートナー (席2, チーム0) が初回メルド済みなら警告は出ない (受け入れ条件2)
  it('does not warn about red 3 penalty when teammate has melded', async () => {
    const base = makeBoliviaState();
    mockExec.mockResolvedValue(
      makeBoliviaState({
        players: base.players.map((p, i) => {
          if (i === 0) {
            return {
              ...p,
              hasInitMeld: false,
              red3s: [{ design: 'HEART' as const, value: 3 }],
              red3Count: 1,
            };
          }
          if (i === 2) {
            return {
              ...p,
              hasInitMeld: true,
            };
          }
          return p;
        }),
      }),
    );
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(screen.getByText('赤3')).toBeInTheDocument());
    expect(screen.queryByTestId('bo-red3-warning-0')).not.toBeInTheDocument();
  });

  // 負のコントロール 3: 赤3がなければ警告は出ない
  it('stays silent when player has no red 3s', async () => {
    const base = makeBoliviaState();
    mockExec.mockResolvedValue(base);
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('bo-red3-warning-0')).not.toBeInTheDocument();
  });

  // 場に捨て札(discardTop)があるときだけ捨て札の山が表示される
  it('shows the discard pile only when discardTop is present', async () => {
    const base = makeBoliviaState();

    // 捨て札なし
    mockExec.mockResolvedValueOnce(makeBoliviaState({ ...base, discardTop: null }));
    const { unmount } = renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('sa-discard-pile')).not.toBeInTheDocument();
    unmount();

    // 捨て札あり
    mockExec.mockResolvedValueOnce(makeBoliviaState({ ...base, discardTop: { design: 'SPADE', value: 4 } }));
    renderWithProviders(<BoliviaPage />);
    expect(await screen.findByTestId('sa-discard-pile')).toBeInTheDocument();
  });

  it('shows server score breakdown values only at round or game end', async () => {
    const state = makeBoliviaState({
      players: makeBoliviaState().players.map((player) =>
        player.team === 0
          ? {
              ...player,
              scoreBreakdown: {
                cardPoints: 37,
                naturalCanastaBonus: 500,
                mixedCanastaBonus: 0,
                escaleraBonus: 1500,
                boliviaBonus: 2500,
                red3Bonus: 300,
                red3Penalty: 100,
                goOutBonus: 100,
                handPenalty: 65,
              },
            }
          : player,
      ),
    });
    mockExec.mockResolvedValue(state);
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('bo-score-breakdown-0')).not.toBeInTheDocument();

    mockExec.mockResolvedValue(makeBoliviaState({ ...state, phase: 3 }));
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    const breakdown = await screen.findByTestId('bo-score-breakdown-0');
    expect(breakdown).toHaveTextContent('チーム0');
    expect(breakdown).toHaveTextContent('あなたのチーム');
    expect(breakdown).toHaveTextContent('37');
    expect(breakdown).toHaveTextContent('500');
    expect(breakdown).toHaveTextContent('1500');
    expect(breakdown).toHaveTextContent('2500');
    expect(breakdown).toHaveTextContent('300');
    expect(breakdown).toHaveTextContent('−100');
    expect(breakdown).toHaveTextContent('−65');
    expect(breakdown).toHaveTextContent('100');
    expect(breakdown).toHaveTextContent('ミックスカナスタ0');
    const opponentBreakdown = await screen.findByTestId('bo-score-breakdown-1');
    expect(opponentBreakdown).toHaveTextContent('チーム1');
    expect(opponentBreakdown).not.toHaveTextContent('あなたのチーム');
    expect(opponentBreakdown).not.toHaveTextContent('−0');
  });

  it('marks each player name as a row header for that player’s scores', async () => {
    const state = makeBoliviaState({
      players: makeBoliviaState().players.map((player, index) => ({
        ...player,
        roundScore: 100 + index,
        cumulativeScore: 1000 + index,
      })),
    });
    mockExec.mockResolvedValue(state);
    renderWithProviders(<BoliviaPage />);

    const playerHeader = await screen.findByRole('rowheader', { name: 'あなた' });
    expect(playerHeader).toHaveAttribute('scope', 'row');
    const playerRow = playerHeader.closest('tr');
    expect(playerRow).toHaveTextContent('100');
    expect(playerRow).toHaveTextContent('1000');
    expect(playerRow?.querySelectorAll('td')).toHaveLength(3);
  });
});
