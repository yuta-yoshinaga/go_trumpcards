import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { gaigelApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, GaigelResponse } from '../types/card';
import { GaigelPhase } from '../types/phases';
import { GaigelPage } from './GaigelPage';

vi.mock('../api/gameApi', () => ({
  gaigelApi: { exec: vi.fn() },
  actionLogApi: { gaigel: vi.fn() },
}));

const mockPlaySound = vi.fn();
const mockSoundValue = { playSound: mockPlaySound, muted: false, toggleMute: vi.fn() };
vi.mock('../providers/SoundProvider', () => ({
  SoundProvider: ({ children }: { children: React.ReactNode }) => children,
  useSound: () => mockSoundValue,
  useOptionalSound: () => mockSoundValue,
}));

const mockExec = vi.mocked(gaigelApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<GaigelResponse> = {}): GaigelResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 5,
        cards: [card('SPADE', 13), card('SPADE', 12), card('HEART', 10)],
        team: 0,
        trickCount: 0,
      },
      { id: 1, isHuman: false, cardCount: 5, cards: [], team: 1, trickCount: 0 },
      { id: 2, isHuman: false, cardCount: 5, cards: [], team: 0, trickCount: 0 },
      { id: 3, isHuman: false, cardCount: 5, cards: [], team: 1, trickCount: 0 },
    ],
    phase: GaigelPhase.PLAY,
    roundNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    dealerIdx: 3,
    trumpSuit: 1,
    stockRemaining: 28,
    isEndgame: false,
    currentTrick: [],
    teamScores: [0, 0],
    roundPoints: [0, 0],
    roundMarriage: [0, 0],
    marriageIndices: [],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    trickWinnerIdx: -1,
    message: '',
    config: { cpuDifficulty: 1, targetScore: 101 },
    ...overrides,
  };
}

const initialState = makeState();
const gameEndState = makeState({
  phase: GaigelPhase.GAME_END,
  gameEndFlag: true,
  winnerTeam: 0,
  teamScores: [101, 60],
});

beforeEach(() => {
  mockExec.mockReset();
  mockPlaySound.mockClear();
  mockExec.mockResolvedValue(initialState);
});

describe('GaigelPage', () => {
  it('calls reset on mount with default config', async () => {
    renderWithProviders(<GaigelPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        targetScore: 101,
      }),
    );
  });

  // The phase key map must hold bare keys; usePhaseNames adds the `phase.`
  // prefix itself, so a prefixed key resolved to the literal
  // "phase.phase.play" on screen. See issue #4374.
  it('renders the translated phase name, not the raw i18n key', async () => {
    renderWithProviders(<GaigelPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('プレイ'));
    expect(screen.getByTestId('phase-indicator')).not.toHaveTextContent('phase.');
  });

  it('renders the team score table and stock readout during play', async () => {
    renderWithProviders(<GaigelPage />);
    await waitFor(() => expect(screen.getByText('チームスコア')).toBeInTheDocument());
    expect(screen.getByText('山札: 28')).toBeInTheDocument();
  });

  it('shows the active target score and non-negative points remaining for each team', async () => {
    mockExec.mockResolvedValue(makeState({ teamScores: [201, 80], config: { cpuDifficulty: 2, targetScore: 201 } }));
    renderWithProviders(<GaigelPage />);

    await screen.findByRole('table');
    expect(screen.getByText('目標: 201点')).toBeInTheDocument();
    expect(screen.getByText('チーム0の残り: 0点')).toBeInTheDocument();
    expect(screen.getByText('チーム1の残り: 121点')).toBeInTheDocument();

    const i18n = (await import('../i18n')).default;
    const previousLanguage = i18n.language;
    await i18n.changeLanguage('en');
    try {
      expect(await screen.findByText('Target: 201 pts')).toBeInTheDocument();
      expect(screen.getByText('Team 0 remaining: 0 pts')).toBeInTheDocument();
      expect(screen.getByText('Team 1 remaining: 121 pts')).toBeInTheDocument();
    } finally {
      await i18n.changeLanguage(previousLanguage);
    }
  });

  it('highlights only the trick winner at trick end and clears the highlight on the next trick', async () => {
    const completedTrick = makeState({
      phase: GaigelPhase.TRICK_END,
      leadPlayerIdx: 0,
      trickWinnerIdx: 2,
      currentTrick: [
        { playerIdx: 0, card: card('SPADE', 13) },
        { playerIdx: 1, card: card('SPADE', 12) },
        { playerIdx: 2, card: card('HEART', 10) },
        { playerIdx: 3, card: card('SPADE', 11) },
      ],
    });
    mockExec.mockResolvedValueOnce(completedTrick).mockResolvedValueOnce(makeState());

    renderWithProviders(<GaigelPage />);
    expect(await screen.findAllByTestId('trick-winner-badge')).toHaveLength(1);
    expect(screen.getByTestId('trick-winner-badge')).toHaveTextContent('勝者');
    const trickCards = screen.getByTestId('trick-display-cards');
    expect(trickCards.querySelectorAll('[data-trick-winner="true"]')).toHaveLength(1);
    const highlightedCard = trickCards.querySelector('[data-trick-winner="true"]');
    expect(highlightedCard).toHaveAttribute('data-player-idx', '2');
    expect(trickCards.querySelector('[data-player-idx="0"][data-trick-winner="true"]')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: '次のトリック' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
    await waitFor(() => expect(screen.queryByTestId('trick-winner-badge')).not.toBeInTheDocument());
  });

  it('does not show a winner during play', async () => {
    mockExec.mockResolvedValue(
      makeState({
        currentTrick: [{ playerIdx: 0, card: card('SPADE', 13) }],
        leadPlayerIdx: 0,
      }),
    );

    renderWithProviders(<GaigelPage />);
    expect(await screen.findByTestId('trick-display-cards')).toBeInTheDocument();
    expect(screen.queryByTestId('trick-winner-badge')).not.toBeInTheDocument();
  });

  it('marks the dealer beside the matching human or CPU player', async () => {
    const cpuDealerView = renderWithProviders(<GaigelPage />);
    const cpuBadge = await screen.findByTestId('gaigel-dealer-badge');
    expect(cpuBadge).toHaveTextContent('ディーラー');
    expect(cpuBadge.parentElement).toHaveAttribute('data-testid', 'gaigel-player-3');
    expect(screen.getAllByTestId('gaigel-dealer-badge')).toHaveLength(1);

    cpuDealerView.unmount();
    mockExec.mockResolvedValue(makeState({ dealerIdx: 0 }));
    renderWithProviders(<GaigelPage />);
    const humanBadge = await screen.findByTestId('gaigel-dealer-badge');
    expect(humanBadge.parentElement).toHaveAttribute('data-testid', 'gaigel-player-0');
    expect(screen.getAllByTestId('gaigel-dealer-badge')).toHaveLength(1);

    const i18n = (await import('../i18n')).default;
    const previousLanguage = i18n.language;
    await i18n.changeLanguage('en');
    try {
      expect(await screen.findByTestId('gaigel-dealer-badge')).toHaveTextContent('[Dealer]');
    } finally {
      await i18n.changeLanguage(previousLanguage);
    }
  });

  it('labels score rows and relates each team score to its column header', async () => {
    mockExec.mockResolvedValue(makeState({ teamScores: [21, 34], roundPoints: [11, 17], roundMarriage: [20, 0] }));
    renderWithProviders(<GaigelPage />);

    const table = await screen.findByRole('table');
    const rowHeaders = within(table).getAllByRole('rowheader');
    expect(rowHeaders.map((header) => header.textContent)).toEqual(['累計得点', 'ラウンド点', 'マリッジ点']);
    expect(rowHeaders.map((header) => header.getAttribute('scope'))).toEqual(['row', 'row', 'row']);
    expect(within(table).getByRole('columnheader', { name: 'チーム0' })).toBeInTheDocument();
    expect(within(table).getByRole('columnheader', { name: 'チーム1' })).toBeInTheDocument();
    expect(within(table).getByRole('row', { name: /累計得点/ })).toHaveTextContent('21');
    expect(within(table).getByRole('row', { name: /ラウンド点/ })).toHaveTextContent('11');
    expect(within(table).getByRole('row', { name: /マリッジ点/ })).toHaveTextContent('20');
  });

  it('keeps cumulative and round row headers when the marriage row is hidden', async () => {
    renderWithProviders(<GaigelPage />);

    const table = await screen.findByRole('table');
    expect(
      within(table)
        .getAllByRole('rowheader')
        .map((header) => header.textContent),
    ).toEqual(['累計得点', 'ラウンド点']);
    expect(within(table).getByRole('row', { name: /累計得点/ })).toBeInTheDocument();
    expect(within(table).getByRole('row', { name: /ラウンド点/ })).toBeInTheDocument();
  });

  it('renders the face-up turn-up card when the stock still holds it', async () => {
    mockExec.mockResolvedValue(makeState({ trumpCard: card('HEART', 10) }));
    renderWithProviders(<GaigelPage />);
    await waitFor(() => expect(screen.getByTestId('gaigel-trump-card')).toBeInTheDocument());
    expect(screen.getByText('めくり札')).toBeInTheDocument();
  });

  it('omits the turn-up card once the stock is exhausted', async () => {
    mockExec.mockResolvedValue(makeState({ trumpCard: undefined, stockRemaining: 0 }));
    renderWithProviders(<GaigelPage />);
    await waitFor(() => expect(screen.getByText('チームスコア')).toBeInTheDocument());
    expect(screen.queryByTestId('gaigel-trump-card')).not.toBeInTheDocument();
  });

  it('dispatches play with the selected card index during play', async () => {
    renderWithProviders(<GaigelPage />);
    const cardBtn = await screen.findByRole('button', { name: '♠ K' });
    fireEvent.click(cardBtn);
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 0));
  });

  it('shows the marriage button when the selected card can declare and dispatches marriage', async () => {
    mockExec.mockResolvedValue(makeState({ marriageIndices: [0] }));
    renderWithProviders(<GaigelPage />);
    // 結婚できる札は名前にバッジの説明が付くので、完全一致では引けない (#6612)。
    const cardBtn = await screen.findByRole('button', { name: /^♠ K/ });
    fireEvent.click(cardBtn);
    const marriageBtn = await screen.findByRole('button', { name: 'マリッジ' });
    mockExec.mockClear();
    fireEvent.click(marriageBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('marriage', undefined, 0));
  });

  it('hides the marriage button when the selected card cannot declare', async () => {
    renderWithProviders(<GaigelPage />);
    const cardBtn = await screen.findByRole('button', { name: '♠ K' });
    fireEvent.click(cardBtn);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'マリッジ' })).not.toBeInTheDocument();
  });

  it('badges both King and Queen when a marriage is available on the human turn', async () => {
    mockExec.mockResolvedValue(makeState({ marriageIndices: [0, 1] }));
    renderWithProviders(<GaigelPage />);
    await waitFor(() => expect(screen.getByTestId('card-role-badge-0')).toBeInTheDocument());
    expect(screen.getByTestId('card-role-badge-1')).toBeInTheDocument();
    // The non-marriage card (index 2) gets no badge.
    expect(screen.queryByTestId('card-role-badge-2')).not.toBeInTheDocument();
    // 切り札のマリッジは 40 点。通常の 💍 と区別が付くよう 👑 を出す。
    expect(screen.getByTestId('card-role-badge-0')).toHaveTextContent('👑');
    expect(screen.getByRole('button', { name: /♠ K.*40/ })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /♠ Q.*40/ })).toBeInTheDocument();
  });

  it('shows 20 points for a non-trump marriage and no badge for a non-candidate', async () => {
    mockExec.mockResolvedValue(
      makeState({
        trumpSuit: 1,
        marriageIndices: [1],
        players: [
          { ...makeState().players[0], cards: [card('SPADE', 13), card('HEART', 12), card('HEART', 13)] },
          ...makeState().players.slice(1),
        ],
      }),
    );
    renderWithProviders(<GaigelPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /♥ Q.*20/ })).toBeInTheDocument());
    // 通常スートは 20 点で 💍 のまま。切り札の 👑 と取り違えないこと。
    expect(screen.getByTestId('card-role-badge-1')).toHaveTextContent('💍');
    expect(screen.getByTestId('card-role-badge-1')).not.toHaveTextContent('👑');
    expect(screen.queryByTestId('card-role-badge-0')).not.toBeInTheDocument();
    expect(screen.queryByTestId('card-role-badge-2')).not.toBeInTheDocument();
  });

  it('shows no marriage badge when no marriage is available', async () => {
    renderWithProviders(<GaigelPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeInTheDocument());
    expect(screen.queryByTestId('card-role-badge-0')).not.toBeInTheDocument();
    expect(screen.queryByTestId('card-role-badge-1')).not.toBeInTheDocument();
  });

  it('shows no marriage badge when it is not the human turn', async () => {
    mockExec.mockResolvedValue(makeState({ marriageIndices: [0, 1], currentPlayerIdx: 1 }));
    renderWithProviders(<GaigelPage />);
    await waitFor(() => expect(screen.getByText('チームスコア')).toBeInTheDocument());
    expect(screen.queryByTestId('card-role-badge-0')).not.toBeInTheDocument();
    expect(screen.queryByTestId('card-role-badge-1')).not.toBeInTheDocument();
  });

  it('shows reset button mid-game and opens confirm dialog', async () => {
    renderWithProviders(<GaigelPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('shows 次のゲーム at game end with no confirm', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<GaigelPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, expect.any(Object)));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('translates the server hint reason instead of leaking the raw code', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<GaigelPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    mockExec.mockResolvedValue(makeState({ hint: { cardIndex: 2, reason: 'lead_trump', isMarriage: false } }));
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByText(/切り札でリード/)).toBeInTheDocument());
    expect(screen.queryByText(/lead_trump/)).not.toBeInTheDocument();
  });

  it("shows '-' when the hint carries no card index", async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<GaigelPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    mockExec.mockResolvedValue(makeState({ hint: { reason: 'marriage', isMarriage: true } }));
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByText(/\[-\]/)).toBeInTheDocument());
  });

  it('falls back to the raw code for an unknown hint reason', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<GaigelPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    mockExec.mockResolvedValue(makeState({ hint: { cardIndex: 0, reason: 'no_such_reason', isMarriage: false } }));
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByText(/no_such_reason/)).toBeInTheDocument());
  });

  // **山が尽きた瞬間にマストフォローへ切り替わる** ── `validateEndgameFollow` が
  // 発動する規則の転換なのに、Gaigel だけどこにも出していなかった (#6482)。
  it('announces the switch to must-follow when the stock runs out', async () => {
    mockExec.mockResolvedValue(makeState({ isEndgame: true, stockRemaining: 0 }));
    renderWithProviders(<GaigelPage />);

    const notice = await screen.findByTestId('ga-endgame-notice');
    expect(notice).toHaveAttribute('role', 'status');
    expect(notice).toHaveTextContent('第2フェーズ');
    expect(notice.textContent).not.toContain('{{');
  });

  // **残り枚数からは判断しない。**判定はドメインが持つので、0 枚でも
  // `isEndgame` が false のうちは出さない。
  it('does not announce must-follow while the domain says otherwise', async () => {
    mockExec.mockResolvedValue(makeState({ isEndgame: false, stockRemaining: 0 }));
    renderWithProviders(<GaigelPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('ga-endgame-notice')).not.toBeInTheDocument();
  });
});
