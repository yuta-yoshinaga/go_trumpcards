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
const mobileState = vi.hoisted(() => ({ isMobile: true, largeDesktop: false }));

vi.mock('../hooks/useCardDimensions', async (importOriginal) => ({
  ...(await importOriginal<typeof import('../hooks/useCardDimensions')>()),
  useIsMobile: () => mobileState.isMobile,
  useBreakpoint: () => (mobileState.largeDesktop ? 'largeDesktop' : 'desktop'),
}));

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
    mobileState.isMobile = true;
    mobileState.largeDesktop = false;
    mockExec.mockResolvedValue(drawPhaseState);
  });

  it('uses a single-row CardFan hand on large desktop', async () => {
    mobileState.isMobile = false;
    mobileState.largeDesktop = true;
    renderWithProviders(<SambaPage />);
    expect(await screen.findByTestId('sa-player-hand')).toBeInTheDocument();
    expect(screen.getByTestId('sa-player-hand').querySelectorAll('button')).toHaveLength(
      drawPhaseState.players[0].cards.length,
    );
    fireEvent.click(within(screen.getByTestId('sa-player-hand')).getAllByRole('button')[0]);
    expect(within(screen.getByTestId('sa-player-hand')).getAllByRole('button')[0]).toHaveAttribute(
      'aria-pressed',
      'true',
    );
  });

  it('omits the large-desktop meld footer when the human team has no melds or red threes', async () => {
    mobileState.isMobile = false;
    mobileState.largeDesktop = true;
    renderWithProviders(<SambaPage />);
    await screen.findByTestId('sa-player-hand');
    expect(screen.queryByTestId('sa-footer-melds')).not.toBeInTheDocument();
  });

  it('places teammate melds beside the large-desktop hand in a bounded scroll area', async () => {
    const base = makeSambaState();
    const teammate = base.players[2];
    mockExec.mockResolvedValue(
      makeSambaState({
        players: [
          base.players[0],
          base.players[1],
          { ...teammate, red3s: [{ design: 'HEART', value: 3 }], red3Count: 1 },
          base.players[3],
        ],
      }),
    );

    mobileState.isMobile = false;
    mobileState.largeDesktop = true;
    renderWithProviders(<SambaPage />);
    const meldArea = await screen.findByTestId('sa-footer-melds');
    const hand = await screen.findByTestId('sa-player-hand');
    const row = screen.getByTestId('sa-footer-hand-row');
    expect(meldArea.style.maxHeight).toBeTruthy();
    expect(meldArea).toHaveClass('shrink-0', 'max-w-[30%]', 'overflow-y-auto');
    expect(meldArea.parentElement).toBe(row);
    expect(hand.parentElement).toBe(row);
    expect(meldArea).toHaveTextContent('赤3');
    expect(meldArea.closest('[data-testid="game-footer-content"]')).toBeInTheDocument();
  });

  afterEach(() => {
    localStorage.clear();
  });

  it('uses stacked MobileHandGrid rows for a large mobile hand', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { configurable: true, value: 375 });
    window.dispatchEvent(new Event('resize'));
    try {
      const base = makeSambaState();
      mockExec.mockResolvedValue(
        makeSambaState({
          players: [
            {
              ...base.players[0],
              cards: Array.from({ length: 16 }, (_, i) => ({ design: 'SPADE' as const, value: (i % 13) + 1 })),
            },
            ...base.players.slice(1),
          ],
        }),
      );
      renderWithProviders(<SambaPage />);

      const rows = await screen.findAllByTestId('hand-row');
      expect(rows.length).toBeGreaterThanOrEqual(2);
      expect(Number.parseFloat((rows[1] as HTMLElement).style.marginTop)).toBeLessThan(0);
    } finally {
      Object.defineProperty(window, 'innerWidth', { configurable: true, value: originalWidth });
      window.dispatchEvent(new Event('resize'));
    }
  });

  it('selects tappable mobile cards and keeps staged meld cards unavailable', async () => {
    mobileState.isMobile = true;
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<SambaPage />);
    const cards = await screen.findAllByTestId('hand-row');
    const buttons = cards.flatMap((row) => Array.from(row.querySelectorAll('button')));
    fireEvent.click(buttons[0]);
    expect(buttons[0]).toHaveAttribute('aria-pressed', 'true');
    fireEvent.click(buttons[1]);
    fireEvent.click(buttons[2]);
    fireEvent.click(screen.getByRole('button', { name: '選択カードをグループに追加' }));

    await waitFor(() => expect(buttons[0]).toHaveAttribute('aria-disabled', 'true'));
    fireEvent.click(buttons[0]);
    expect(buttons[0]).toHaveAttribute('aria-pressed', 'true');
    fireEvent.click(buttons[3]);
    expect(buttons[3]).toHaveAttribute('aria-pressed', 'true');
  });

  it('collapses opponents melds on mobile while keeping the human team visible', async () => {
    const base = makeSambaState();
    const opponentMeld = {
      cards: Array.from({ length: 3 }, () => ({ design: 'SPADE' as const, value: 4 })),
      kind: 0,
      isNatural: true,
      isCanasta: false,
      isSamba: false,
      rank: 4,
    };
    mockExec.mockResolvedValue(
      makeSambaState({
        players: [
          { ...base.players[0], melds: [opponentMeld] },
          { ...base.players[1], melds: [opponentMeld] },
          ...base.players.slice(2),
        ],
      }),
    );
    renderWithProviders(<SambaPage />);

    const details = await screen.findByTestId('sa-opponent-meld-details');
    expect(details).not.toHaveAttribute('open');
    expect(within(details).getByText('相手チームのメルド (1)')).toBeInTheDocument();
    expect(within(details).getByText('あと4枚でカナスタ')).not.toBeVisible();
    expect(screen.getByTestId('sa-meld-progress-0-0')).toBeVisible();
  });

  it('expands opponents melds on desktop', async () => {
    mobileState.isMobile = false;
    const base = makeSambaState();
    const opponentMeld = {
      cards: Array.from({ length: 3 }, () => ({ design: 'SPADE' as const, value: 4 })),
      kind: 0,
      isNatural: true,
      isCanasta: false,
      isSamba: false,
      rank: 4,
    };
    mockExec.mockResolvedValue(
      makeSambaState({
        players: [base.players[0], { ...base.players[1], melds: [opponentMeld] }, ...base.players.slice(2)],
      }),
    );
    renderWithProviders(<SambaPage />);

    const details = await screen.findByTestId('sa-opponent-meld-details');
    expect(details).toHaveAttribute('open');
    expect(within(details).getByText('あと4枚でカナスタ')).toBeVisible();
  });

  it('does not render opponent meld details when opponents have no melds or red threes', async () => {
    renderWithProviders(<SambaPage />);
    await screen.findByTestId('game-footer-actions');
    expect(screen.queryByTestId('sa-opponent-meld-details')).not.toBeInTheDocument();
  });

  it('keeps only the hand and its selection feedback in the footer content', async () => {
    const { container } = renderWithProviders(<SambaPage />);
    const footer = await screen.findByTestId('game-footer-content');
    const playArea = container.querySelector('.flex-1.overflow-y-auto');

    expect(playArea).not.toBeNull();
    expect(footer.querySelector('[data-tutorial="sa-player-hand"]')).not.toBeNull();
    expect(within(footer).getByTestId('sa-meld-points')).toBeInTheDocument();
    expect(within(footer).queryByTestId('samba-kbd-shortcuts')).not.toBeInTheDocument();
    expect(within(footer).queryByRole('button', { name: 'リセット' })).not.toBeInTheDocument();
    expect(within(playArea as HTMLElement).getByTestId('samba-kbd-shortcuts')).toBeInTheDocument();
    expect(within(playArea as HTMLElement).getByRole('button', { name: 'リセット' })).toBeInTheDocument();
  });

  it('shows an opponent meld and counts only melds (not red threes) in the summary', async () => {
    const base = makeSambaState();
    const meld = {
      cards: [
        { design: 'SPADE' as const, value: 4 },
        { design: 'HEART' as const, value: 4 },
      ],
      kind: 0,
      isNatural: true,
      isCanasta: false,
      isSamba: false,
      rank: 4,
    };
    const red3 = { design: 'HEART' as const, value: 3 };
    mockExec.mockResolvedValue(
      makeSambaState({
        players: [base.players[0], { ...base.players[1], melds: [meld], red3s: [red3] }, ...base.players.slice(2)],
      }),
    );
    renderWithProviders(<SambaPage />);

    const details = await screen.findByTestId('sa-opponent-meld-details');
    expect(within(details).getByText('相手チームのメルド (1)')).toBeInTheDocument();
    expect(within(details).getByAltText('♠ 4')).toBeInTheDocument();
    expect(within(details).getByAltText('♥ 3')).toBeInTheDocument();
  });

  it('shows an opponent red three when that seat has no melds', async () => {
    const base = makeSambaState();
    mockExec.mockResolvedValue(
      makeSambaState({
        players: [
          base.players[0],
          { ...base.players[1], red3s: [{ design: 'DIAMOND', value: 3 }] },
          ...base.players.slice(2),
        ],
      }),
    );
    renderWithProviders(<SambaPage />);

    const details = await screen.findByTestId('sa-opponent-meld-details');
    expect(within(details).getByAltText('♦ 3')).toBeInTheDocument();
  });

  it('omits an empty opponent seat when the other opponent has a meld', async () => {
    const base = makeSambaState();
    const meld = {
      cards: [{ design: 'SPADE' as const, value: 4 }],
      kind: 0,
      isNatural: true,
      isCanasta: false,
      isSamba: false,
      rank: 4,
    };
    mockExec.mockResolvedValue(
      makeSambaState({
        players: [base.players[0], { ...base.players[1], melds: [meld] }, base.players[2], base.players[3]],
      }),
    );
    renderWithProviders(<SambaPage />);

    const details = await screen.findByTestId('sa-opponent-meld-details');
    expect(within(details).getByText(/CPU 1/)).toBeInTheDocument();
    expect(within(details).queryByText(/CPU 3/)).not.toBeInTheDocument();
    expect(within(details).getAllByTestId('animated-card')).toHaveLength(1);
  });

  it('keeps the human team meld outside opponent details', async () => {
    const base = makeSambaState();
    const meld = {
      cards: [{ design: 'SPADE' as const, value: 4 }],
      kind: 0,
      isNatural: true,
      isCanasta: false,
      isSamba: false,
      rank: 4,
    };
    mockExec.mockResolvedValue(
      makeSambaState({
        players: [
          { ...base.players[0], melds: [meld] },
          { ...base.players[1], melds: [meld] },
          ...base.players.slice(2),
        ],
      }),
    );
    renderWithProviders(<SambaPage />);

    const details = await screen.findByTestId('sa-opponent-meld-details');
    const allMeldProgress = screen.getAllByText('あと6枚でカナスタ');
    expect(details).not.toContainElement(allMeldProgress[0]);
    expect(details).toContainElement(allMeldProgress[1]);
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

  it('collapses the score table on mobile and opens it on desktop', async () => {
    const { unmount } = renderWithProviders(<SambaPage />);
    await screen.findByTestId('sa-score-details');
    const mobileDetails = screen.getByTestId('sa-score-details');
    expect(mobileDetails).not.toHaveAttribute('open');
    expect(within(mobileDetails).getByRole('table')).not.toBeVisible();
    unmount();

    mobileState.isMobile = false;
    renderWithProviders(<SambaPage />);
    const desktopDetails = await screen.findByTestId('sa-score-details');
    expect(desktopDetails).toHaveAttribute('open');
    expect(within(desktopDetails).getByRole('table')).toBeVisible();
  });

  it('renders the score summary using the active locale', async () => {
    await i18n.changeLanguage('ja');
    const { unmount } = renderWithProviders(<SambaPage />);
    expect(await within(await screen.findByTestId('sa-score-details')).findByText('スコア')).toBeInTheDocument();
    expect(screen.queryByText('label.score')).not.toBeInTheDocument();
    unmount();

    await i18n.changeLanguage('en');
    renderWithProviders(<SambaPage />);
    expect(await within(await screen.findByTestId('sa-score-details')).findByText('Scores')).toBeInTheDocument();
    expect(screen.queryByText('label.score')).not.toBeInTheDocument();
    await i18n.changeLanguage('ja');
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
  });

  it('describes meld points from desktop hand cards during the initial meld', async () => {
    mobileState.isMobile = false;
    mockExec.mockResolvedValue({ ...meldPhaseState, minMeld: 90 });
    renderWithProviders(<SambaPage />);

    await screen.findByTestId('sa-meld-points');
    const hand = document.querySelector('[data-tutorial="sa-player-hand"] button');
    expect(hand).toHaveAttribute('aria-describedby', 'sa-meld-points');
  });

  it('does not describe meld points after the initial meld is complete', async () => {
    mobileState.isMobile = false;
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

  it('does not add meld-point announcements to mobile card labels before or after the initial meld', async () => {
    const pointsLabel = '初回メルド必要点';
    mockExec.mockResolvedValue(meldPhaseState);
    const { unmount } = renderWithProviders(<SambaPage />);
    const beforeCards = await screen.findAllByTestId('hand-row');
    const beforeCard = beforeCards[0]?.querySelector('button');
    expect(beforeCard).toBeInTheDocument();
    expect(beforeCard?.getAttribute('aria-label')).not.toContain(pointsLabel);
    unmount();

    const stateAfterInitialMeld = makeSambaState({
      phase: 1,
      messageCode: 'samba.meldPhase',
      players: [{ ...meldPhaseState.players[0], hasInitMeld: true }, ...meldPhaseState.players.slice(1)],
    });
    mockExec.mockResolvedValue(stateAfterInitialMeld);
    renderWithProviders(<SambaPage />);
    const afterCards = await screen.findAllByTestId('hand-row');
    const afterCard = afterCards[0]?.querySelector('button');
    expect(afterCard).toBeInTheDocument();
    expect(afterCard?.getAttribute('aria-label')).not.toContain(pointsLabel);
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

  it('keeps go-out progress in the play area during the discard phase', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    const { container } = renderWithProviders(<SambaPage />);
    const progress = await screen.findByTestId('sa-go-out-progress');
    const footer = await screen.findByTestId('game-footer-content');
    const playArea = container.querySelector('.flex-1.overflow-y-auto');

    expect(playArea).not.toBeNull();
    expect(within(footer).queryByTestId('sa-go-out-progress')).not.toBeInTheDocument();
    expect(within(playArea as HTMLElement).getByTestId('sa-go-out-progress')).toBe(progress);
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
