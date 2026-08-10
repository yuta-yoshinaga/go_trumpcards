import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { pitchApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { CardDesign, PitchResponse } from '../types/card';
import { PitchPhase } from '../types/phases';
import { PitchPage } from './PitchPage';

vi.mock('../api/gameApi', () => ({
  pitchApi: { exec: vi.fn() },
  actionLogApi: { pitch: vi.fn() },
}));

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

const mockApi = vi.mocked(pitchApi.exec);
void useCliMode; // kept for vi.mock side-effect; unused locally

const makeCard = (design: CardDesign, value: number) => ({ design, value });

const baseConfig = { cpuDifficulty: 1, pointLimit: 7 };

const bidState: PitchResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 6,
      cards: [makeCard('SPADE', 5), makeCard('HEART', 9)],
      bid: -1,
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
    },
    { id: 1, isHuman: false, cardCount: 6, cards: [], bid: -1, roundScore: 0, cumulativeScore: 0, trickCount: 0 },
    { id: 2, isHuman: false, cardCount: 6, cards: [], bid: -1, roundScore: 0, cumulativeScore: 0, trickCount: 0 },
    { id: 3, isHuman: false, cardCount: 6, cards: [], bid: -1, roundScore: 0, cumulativeScore: 0, trickCount: 0 },
  ],
  phase: PitchPhase.BID,
  roundNumber: 1,
  trickNumber: 0,
  dealerIdx: 3,
  currentPlayerIdx: -1,
  bidPlayerIdx: 0,
  currentBid: 0,
  bidWinnerIdx: -1,
  trumpSuit: 0,
  currentTrick: [],
  lastTrick: [],
  lastTrickWinner: -1,
  gameEndFlag: false,
  winnerIdx: -1,
  leadPlayerIdx: -1,
  validPlayIndices: [],
  message: '',
  config: baseConfig,
};

const playState: PitchResponse = {
  ...bidState,
  phase: PitchPhase.PLAY,
  trickNumber: 1,
  currentPlayerIdx: 0,
  bidPlayerIdx: 0,
  currentBid: 3,
  bidWinnerIdx: 0,
  trumpSuit: 1,
  validPlayIndices: [0, 1],
  leadPlayerIdx: 0,
  players: bidState.players.map((p, i) => (i === 0 ? { ...p, bid: 3 } : { ...p, bid: 0 })),
};

// A play-phase trick with trump = SPADE (1); lead card is a HEART, so HEART is
// the lead suit, and the SPADE card is a trump cut.
const trickState: PitchResponse = {
  ...playState,
  trumpSuit: 1,
  currentTrick: [
    { playerIdx: 3, card: makeCard('HEART', 9) },
    { playerIdx: 0, card: makeCard('HEART', 4) },
    { playerIdx: 1, card: makeCard('SPADE', 12) },
  ],
};

const gameEndState: PitchResponse = {
  ...bidState,
  phase: PitchPhase.GAME_END,
  gameEndFlag: true,
  winnerIdx: 0,
};

beforeEach(() => {
  vi.clearAllMocks();
  mockApi.mockResolvedValue(bidState);
});

describe('PitchPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<PitchPage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('renders bid phase with pass + bid buttons', async () => {
    mockApi.mockResolvedValue(bidState);
    renderWithProviders(<PitchPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /パス/ })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /ビッド 2/ })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /ビッド 3/ })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /ビッド 4/ })).toBeInTheDocument();
  });

  it('renders play phase with player hand', async () => {
    mockApi.mockResolvedValue(playState);
    renderWithProviders(<PitchPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /出す/ })).toBeInTheDocument());
  });

  it('bid phase: pressing "p" passes and "2" bids two', async () => {
    mockApi.mockResolvedValue(bidState);
    renderWithProviders(<PitchPage />);
    await screen.findByRole('button', { name: /パス/ });
    mockApi.mockClear();
    fireEvent.keyDown(document.body, { key: 'p' });
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bid', 0));
    mockApi.mockClear();
    fireEvent.keyDown(document.body, { key: '2' });
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bid', 2));
  });

  it('play phase: a number key selects a valid card and Enter plays it', async () => {
    mockApi.mockResolvedValue(playState);
    renderWithProviders(<PitchPage />);
    await screen.findByRole('button', { name: /出す/ });
    mockApi.mockClear();
    // "1" → hand index 0 (a valid play index); Enter confirms.
    fireEvent.keyDown(document.body, { key: '1' });
    fireEvent.keyDown(document.body, { key: 'Enter' });
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('play', undefined, 0));
  });

  it('play phase: an invalid card cannot be selected or played', async () => {
    // Only index 0 is a legal play; pressing "2" (index 1) must not select.
    mockApi.mockResolvedValue({ ...playState, validPlayIndices: [0] });
    renderWithProviders(<PitchPage />);
    await screen.findByRole('button', { name: /出す/ });
    mockApi.mockClear();
    fireEvent.keyDown(document.body, { key: '2' });
    fireEvent.keyDown(document.body, { key: 'Enter' });
    // No selection means handlePlay early-returns; no play command is sent.
    await waitFor(() => expect(mockApi).not.toHaveBeenCalled());
  });

  it('advertises the bid keyboard shortcut on the pass button', async () => {
    mockApi.mockResolvedValue(bidState);
    renderWithProviders(<PitchPage />);
    const pass = await screen.findByRole('button', { name: /パス/ });
    expect(pass).toHaveAttribute('aria-keyshortcuts', 'p');
    expect(pass.querySelector('kbd')?.textContent).toBe('P');
  });

  it('shows score table with players', async () => {
    mockApi.mockResolvedValue(bidState);
    renderWithProviders(<PitchPage />);
    await waitFor(() => expect(screen.getAllByText(/あなた/).length).toBeGreaterThan(0));
    expect(screen.getAllByText(/CPU 1/).length).toBeGreaterThan(0);
  });

  it('shows winner banner on game end', async () => {
    mockApi.mockResolvedValue(gameEndState);
    renderWithProviders(<PitchPage />);
    await waitFor(() => expect(screen.getByText(/あなたの勝利/)).toBeInTheDocument());
  });

  it('renders the Game-pip badge with the human hand total', async () => {
    // SPADE 5 (0 pips) + HEART 9 (0 pips) = 0
    mockApi.mockResolvedValue(bidState);
    renderWithProviders(<PitchPage />);
    const badge = await screen.findByTestId('pitch-game-pips-badge');
    expect(badge.textContent).toMatch(/Game値: 0/);
  });

  it('Game-pip badge tooltip has no trailing newline when pips = 0', async () => {
    // SPADE 5 (0) + HEART 9 (0) → no breakdown line should be appended
    mockApi.mockResolvedValue(bidState);
    renderWithProviders(<PitchPage />);
    const badge = await screen.findByTestId('pitch-game-pips-badge');
    const title = badge.getAttribute('title') ?? '';
    expect(title).not.toMatch(/\n$/);
    expect(title).not.toContain('\n');
  });

  it('Game-pip badge sums A, K, Q, J, 10 correctly', async () => {
    const pipHand: PitchResponse = {
      ...bidState,
      players: [
        {
          ...bidState.players[0],
          // A(4) + 10(10) + J(1) + K(3) + Q(2) + 7(0) = 20
          cards: [
            makeCard('SPADE', 1),
            makeCard('SPADE', 10),
            makeCard('SPADE', 11),
            makeCard('SPADE', 13),
            makeCard('SPADE', 12),
            makeCard('HEART', 7),
          ],
          cardCount: 6,
        },
        ...bidState.players.slice(1),
      ],
    };
    mockApi.mockResolvedValue(pipHand);
    renderWithProviders(<PitchPage />);
    const badge = await screen.findByTestId('pitch-game-pips-badge');
    expect(badge.textContent).toMatch(/Game値: 20/);
    // Tapping opens a breakdown popover (reachable on touch, unlike the title tooltip).
    expect(screen.queryByTestId('pitch-game-pips-popover')).not.toBeInTheDocument();
    fireEvent.click(badge);
    const popover = screen.getByTestId('pitch-game-pips-popover');
    // Breakdown lists only contributing cards plus the total; the zero-pip 7 is excluded.
    expect(popover.textContent).toContain('+4'); // A
    expect(popover.textContent).toContain('+10'); // 10
    expect(popover).toHaveTextContent('合計');
    expect(popover.textContent).toMatch(/20/);
    // Tapping again closes it.
    fireEvent.click(badge);
    expect(screen.queryByTestId('pitch-game-pips-popover')).not.toBeInTheDocument();
  });

  it('Game-pip badge meets the 44px tap-target minimum', async () => {
    mockApi.mockResolvedValue(bidState);
    renderWithProviders(<PitchPage />);
    const badge = await screen.findByTestId('pitch-game-pips-badge');
    expect(badge.className).toContain('min-h-[44px]');
    expect(badge.className).toContain('min-w-[44px]');
  });

  it('emphasizes the lead card and trump cards in the current trick', async () => {
    mockApi.mockResolvedValue(trickState);
    renderWithProviders(<PitchPage />);

    // Lead badge is on the first card of the trick, whose wrapper is ringed as
    // the lead suit (HEART here, distinct from the trump ring color).
    const leadBadge = await screen.findByTestId('pt-trick-lead-badge');
    expect(leadBadge).toHaveTextContent('リード');
    expect(leadBadge.parentElement?.className).toContain('ring-ds-info');

    // Trump card carries a suit-symbol badge (non-color cue) and an orange ring.
    const trumpBadge = screen.getByTestId('pt-trick-trump-badge');
    expect(trumpBadge).toHaveTextContent('♠');
    expect(trumpBadge.parentElement?.className).toContain('ring-ds-warning');

    // The header trump indicator is emphasized once trump is set.
    expect(screen.getByTestId('pt-trump-indicator').className).toContain('ring-ds-warning');
    // Legend explains both rings.
    expect(screen.getByTestId('pt-trick-legend')).toBeInTheDocument();
  });

  it('closes the pips popover on Escape and on an outside click', async () => {
    const pipHand: PitchResponse = {
      ...bidState,
      players: [
        { ...bidState.players[0], cards: [makeCard('SPADE', 1), makeCard('SPADE', 10)], cardCount: 2 },
        ...bidState.players.slice(1),
      ],
    };
    mockApi.mockResolvedValue(pipHand);
    renderWithProviders(<PitchPage />);
    const badge = await screen.findByTestId('pitch-game-pips-badge');

    fireEvent.click(badge);
    expect(screen.getByTestId('pitch-game-pips-popover')).toBeInTheDocument();
    fireEvent.keyDown(document, { key: 'Escape' });
    expect(screen.queryByTestId('pitch-game-pips-popover')).not.toBeInTheDocument();

    fireEvent.click(badge);
    expect(screen.getByTestId('pitch-game-pips-popover')).toBeInTheDocument();
    fireEvent.mouseDown(document.body);
    expect(screen.queryByTestId('pitch-game-pips-popover')).not.toBeInTheDocument();
  });

  it('previous-trick panel shows the just-completed trick and its winner', async () => {
    const lastTrickState: PitchResponse = {
      ...playState,
      trickNumber: 2,
      lastTrick: [
        { playerIdx: 0, card: makeCard('HEART', 9) },
        { playerIdx: 1, card: makeCard('HEART', 4) },
        { playerIdx: 2, card: makeCard('SPADE', 12) },
        { playerIdx: 3, card: makeCard('HEART', 2) },
      ],
      lastTrickWinner: 2,
    };
    mockApi.mockResolvedValue(lastTrickState);
    renderWithProviders(<PitchPage />);
    await waitFor(() => expect(screen.getByTestId('pt-previous-trick')).toBeInTheDocument());
    // The winner label is rendered (CPU 2 won the trick) and the empty placeholder is not.
    expect(screen.getByTestId('pt-previous-trick')).toHaveTextContent(/獲得/);
    expect(screen.queryByTestId('pt-previous-trick-empty')).not.toBeInTheDocument();
  });

  it('previous-trick panel is empty on the round first trick', async () => {
    mockApi.mockResolvedValue(playState); // trickNumber 1, lastTrick []
    renderWithProviders(<PitchPage />);
    await waitFor(() => expect(screen.getByTestId('pt-previous-trick-empty')).toBeInTheDocument());
  });

  it('renders the i18n skeleton instead of a hardcoded Loading label before state loads', () => {
    mockApi.mockReturnValue(new Promise(() => {}));
    renderWithProviders(<PitchPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
  });
});
