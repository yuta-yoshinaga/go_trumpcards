import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { kalookiApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, KalookiResponse } from '../types/card';
import { KalookiPage } from './KalookiPage';

vi.mock('../api/gameApi', () => ({
  kalookiApi: { exec: vi.fn() },
  actionLogApi: { kalooki: vi.fn() },
}));

const mockExec = vi.mocked(kalookiApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const baseHand: Card[] = [
  card('SPADE', 5),
  card('HEART', 5),
  card('DIAMOND', 5),
  card('CLOVER', 13),
  card('SPADE', 13),
  card('HEART', 13),
  card('SPADE', 2),
  card('SPADE', 3),
  card('SPADE', 4),
  card('SPADE', 6),
  card('SPADE', 7),
  card('SPADE', 8),
  card('SPADE', 9),
];

const cpu = (id: number): KalookiResponse['players'][number] => ({
  id,
  isHuman: false,
  cardCount: 13,
  cards: [],
  melds: [],
  hasOpened: false,
  roundScore: 0,
  cumulativeScore: 0,
});

const drawState: KalookiResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 13,
      cards: baseHand,
      melds: [],
      hasOpened: false,
      roundScore: 0,
      cumulativeScore: 0,
    },
    cpu(1),
    cpu(2),
  ],
  phase: 0,
  openingThreshold: 51,
  currentPlayerIdx: 0,
  discardTop: card('HEART', 7),
  drawPileCount: 60,
  gameEndFlag: false,
  winnerIdx: -1,
  roundWinnerIdx: -1,
  config: { cpuDifficulty: 1, playerCount: 3, openingThreshold: 51 },
  message: '',
};

const meldState: KalookiResponse = { ...drawState, phase: 1 };

const openedState: KalookiResponse = {
  ...meldState,
  players: [
    { ...drawState.players[0], hasOpened: true },
    { ...cpu(1), hasOpened: true, melds: [{ cards: [card('SPADE', 5), card('HEART', 5), card('DIAMOND', 5)] }] },
    cpu(2),
  ],
};

const roundEndState: KalookiResponse = { ...drawState, phase: 2, roundWinnerIdx: 0 };

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
  mockExec.mockResolvedValue(drawState);
});

// Hand-card buttons are scoped to the hand section so they don't collide with
// the face-up meld buttons (which also contain card <img> elements).
const handButtons = () => {
  const hand = document.querySelector('[data-tutorial="kalooki-hand"]');
  if (!hand) return [];
  return Array.from(hand.querySelectorAll('button')).filter((b) => b.querySelector('img'));
};

describe('KalookiPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<KalookiPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows the opening-threshold and stock banner', async () => {
    renderWithProviders(<KalookiPage />);
    // The banner span renders "<label>: <value>"; match the value-bearing node to
    // avoid colliding with the settings-panel <label> of the same text.
    await waitFor(() => expect(screen.getByText(/(Opening points|オープニング点数)\s*:\s*51/)).toBeInTheDocument());
    expect(screen.getByText(/(Stock|山札)\s*:\s*60/)).toBeInTheDocument();
  });

  it('shows draw-phase buttons during the human draw turn', async () => {
    renderWithProviders(<KalookiPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /Draw from stock|山札から引く/ })).toBeInTheDocument(),
    );
    expect(screen.getByRole('button', { name: /Take discard|捨て札を取る/ })).toBeInTheDocument();
  });

  it('invokes drawstock when the stock button is clicked', async () => {
    renderWithProviders(<KalookiPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /Draw from stock|山札から引く/ })).toBeInTheDocument(),
    );
    fireEvent.click(screen.getByRole('button', { name: /Draw from stock|山札から引く/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('invokes drawdiscard when the discard button is clicked', async () => {
    renderWithProviders(<KalookiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /Take discard|捨て札を取る/ })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: /Take discard|捨て札を取る/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawdiscard'));
  });

  it('shows the opening-points hint while the human has not opened', async () => {
    mockExec.mockResolvedValue(meldState);
    renderWithProviders(<KalookiPage />);
    await waitFor(() => expect(screen.getByTestId('kalooki-opening-hint')).toBeInTheDocument());
  });

  it('shows meld-phase action buttons after drawing', async () => {
    mockExec.mockResolvedValue(meldState);
    renderWithProviders(<KalookiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /Add group|グループに追加/ })).toBeInTheDocument());
    expect(screen.getByTestId('kalooki-submit-meld')).toBeInTheDocument();
  });

  it('toggles a card and enables Add group only with 3+ cards', async () => {
    mockExec.mockResolvedValue(meldState);
    renderWithProviders(<KalookiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /Add group|グループに追加/ })).toBeInTheDocument());
    const addGroup = screen.getByRole('button', { name: /Add group|グループに追加/ });
    expect(addGroup).toBeDisabled();
    const cards = handButtons();
    fireEvent.click(cards[0]);
    fireEvent.click(cards[1]);
    expect(addGroup).toBeDisabled();
    fireEvent.click(cards[2]);
    expect(addGroup).not.toBeDisabled();
    fireEvent.click(cards[2]); // toggle off
    expect(addGroup).toBeDisabled();
  });

  it('labels hand cards by name and reflects selection with aria-pressed', async () => {
    mockExec.mockResolvedValue(meldState);
    renderWithProviders(<KalookiPage />);
    // baseHand[0] is ♠5 → cardAlt reads "♠ 5".
    const first = await screen.findByRole('button', { name: '♠ 5' });
    expect(first).toHaveAttribute('aria-pressed', 'false');
    fireEvent.click(first);
    expect(screen.getByRole('button', { name: '♠ 5' })).toHaveAttribute('aria-pressed', 'true');
  });

  it('stages a meld group and submits the meld with meldGroups', async () => {
    mockExec.mockResolvedValue(meldState);
    renderWithProviders(<KalookiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /Add group|グループに追加/ })).toBeInTheDocument());
    const cards = handButtons();
    fireEvent.click(cards[0]);
    fireEvent.click(cards[1]);
    fireEvent.click(cards[2]);
    fireEvent.click(screen.getByRole('button', { name: /Add group|グループに追加/ }));
    fireEvent.click(screen.getByTestId('kalooki-submit-meld'));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('meld', expect.objectContaining({ meldGroups: [[0, 1, 2]] })),
    );
  });

  // Stage a meld group from the first `count` hand cards (starting at `start`).
  const stageGroup = (start: number, count: number) => {
    const cards = handButtons();
    for (let i = start; i < start + count; i++) fireEvent.click(cards[i]);
    fireEvent.click(screen.getByRole('button', { name: /Add group|グループに追加/ }));
  };

  it('visualizes each staged meld group with its cards', async () => {
    mockExec.mockResolvedValue(meldState);
    renderWithProviders(<KalookiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /Add group|グループに追加/ })).toBeInTheDocument());
    // No staged-groups panel until a group is added.
    expect(screen.queryByTestId('kalooki-staged-groups')).not.toBeInTheDocument();
    stageGroup(0, 3);
    const group0 = screen.getByTestId('kalooki-staged-group-0');
    expect(group0).toBeInTheDocument();
    // The three staged card indices render as three mini card images.
    expect(group0.querySelectorAll('img')).toHaveLength(3);
  });

  it('removes an individual group and leaves the other groups staged', async () => {
    mockExec.mockResolvedValue(meldState);
    renderWithProviders(<KalookiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /Add group|グループに追加/ })).toBeInTheDocument());
    stageGroup(0, 3); // group 0 → indices 0,1,2
    stageGroup(3, 3); // group 1 → indices 3,4,5
    expect(screen.getByTestId('kalooki-staged-group-0')).toBeInTheDocument();
    expect(screen.getByTestId('kalooki-staged-group-1')).toBeInTheDocument();

    // Remove the FIRST group; the second must remain (not a clear-all).
    fireEvent.click(screen.getByTestId('kalooki-remove-group-0'));
    expect(screen.getByTestId('kalooki-staged-group-0')).toBeInTheDocument();
    expect(screen.queryByTestId('kalooki-staged-group-1')).not.toBeInTheDocument();

    // The surviving group's cards (indices 3,4,5) are submitted; the removed ones are gone.
    fireEvent.click(screen.getByTestId('kalooki-submit-meld'));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('meld', expect.objectContaining({ meldGroups: [[3, 4, 5]] })),
    );
  });

  it('re-enables hand-card selection after its group is removed', async () => {
    mockExec.mockResolvedValue(meldState);
    renderWithProviders(<KalookiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /Add group|グループに追加/ })).toBeInTheDocument());
    stageGroup(0, 3);
    // Card index 0 is locked (disabled) while it belongs to a staged group.
    expect(screen.getByTestId('kalooki-hand-0')).toBeDisabled();
    fireEvent.click(screen.getByTestId('kalooki-remove-group-0'));
    // Removing the group frees the card for re-selection.
    expect(screen.getByTestId('kalooki-hand-0')).not.toBeDisabled();
  });

  it('discards the selected card in the meld phase', async () => {
    mockExec.mockResolvedValue(meldState);
    renderWithProviders(<KalookiPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /Discard a card|カードを捨てる/ })).toBeInTheDocument(),
    );
    const discard = screen.getByRole('button', { name: /Discard a card|カードを捨てる/ });
    expect(discard).toBeDisabled();
    fireEvent.click(handButtons()[0]);
    expect(discard).not.toBeDisabled();
    fireEvent.click(discard);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', { cardIndex: 0 }));
  });

  it('shows the opened badge for an opened player', async () => {
    mockExec.mockResolvedValue(openedState);
    renderWithProviders(<KalookiPage />);
    await waitFor(() => expect(screen.getByTestId('kalooki-opened-0')).toBeInTheDocument());
    expect(screen.getByTestId('kalooki-opened-1')).toBeInTheDocument();
  });

  it('shows the layoff button once the human has opened', async () => {
    mockExec.mockResolvedValue(openedState);
    renderWithProviders(<KalookiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /Lay off|レイオフ/ })).toBeInTheDocument());
  });

  it('selecting an opponent meld highlights it and shows the layoff target', async () => {
    mockExec.mockResolvedValue(openedState);
    renderWithProviders(<KalookiPage />);
    const meldBtn = await screen.findByRole('button', { name: /メルド1|meld 1/ });
    fireEvent.click(meldBtn);
    expect(meldBtn).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByTestId('kalooki-layoff-target')).toBeInTheDocument();
  });

  it('invokes layoff with the target meld and selected card', async () => {
    mockExec.mockResolvedValue(openedState);
    renderWithProviders(<KalookiPage />);
    const meldBtn = await screen.findByRole('button', { name: /メルド1|meld 1/ });
    fireEvent.click(meldBtn);
    fireEvent.click(handButtons()[0]);
    fireEvent.click(screen.getByRole('button', { name: /Lay off|レイオフ/ }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('layoff', { targetPlayerIdx: 1, meldIdx: 0, cardIndex: 0 }),
    );
  });

  it('does not let you target an opponent meld before opening', async () => {
    const notOpened: KalookiResponse = {
      ...meldState,
      players: [
        { ...drawState.players[0], hasOpened: false },
        { ...cpu(1), hasOpened: true, melds: [{ cards: [card('SPADE', 5), card('HEART', 5), card('DIAMOND', 5)] }] },
        cpu(2),
      ],
    };
    mockExec.mockResolvedValue(notOpened);
    renderWithProviders(<KalookiPage />);
    const meldBtn = await screen.findByRole('button', { name: /メルド1|meld 1/ });
    expect(meldBtn).not.toHaveAttribute('aria-pressed');
    fireEvent.click(meldBtn);
    expect(screen.queryByTestId('kalooki-layoff-target')).not.toBeInTheDocument();
  });

  it('shows the next-round button at round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<KalookiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /Next round|次のラウンドへ/ })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: /Next round|次のラウンドへ/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('keeps CPU hands hidden during the draw phase', async () => {
    // drawState (from beforeEach) is phase 0 — CPU faces must not be revealed.
    renderWithProviders(<KalookiPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('kalooki-reveal-1')).not.toBeInTheDocument();
  });

  it('reveals CPU hands at round end', async () => {
    const revealState: KalookiResponse = {
      ...roundEndState,
      players: [
        roundEndState.players[0],
        { ...cpu(1), cardCount: 2, cards: [card('SPADE', 5), card('HEART', 9)] },
        cpu(2),
      ],
    };
    mockExec.mockResolvedValue(revealState);
    renderWithProviders(<KalookiPage />);
    const reveal = await screen.findByTestId('kalooki-reveal-1');
    expect(reveal.querySelectorAll('img')).toHaveLength(2);
    // A CPU with no cards left (winner) shows no reveal block.
    expect(screen.queryByTestId('kalooki-reveal-2')).not.toBeInTheDocument();
  });

  it('shows the winner banner at game end', async () => {
    mockExec.mockResolvedValue({ ...drawState, phase: 3, gameEndFlag: true, winnerIdx: 0 });
    renderWithProviders(<KalookiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /Next Game|次のゲーム/ })).toBeInTheDocument());
  });

  it('toggles CLI mode and renders the terminal', async () => {
    renderWithProviders(<KalookiPage />);
    const toggle = await screen.findByRole('button', { name: /CLIモードに切り替え|Switch to CLI/i });
    fireEvent.click(toggle);
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
  });

  it('resets with the chosen settings config after confirming', async () => {
    renderWithProviders(<KalookiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /^リセット$|^Reset$/ })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: /^リセット$|^Reset$/ }));
    fireEvent.click(await screen.findByRole('button', { name: /^確認$|^Confirm$/ }));
    await waitFor(() => {
      const resetWithConfig = mockExec.mock.calls.filter((c) => c[0] === 'reset' && c[1] !== undefined);
      expect(resetWithConfig.length).toBeGreaterThan(0);
    });
  });
});
