import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, ohHellApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { OhHellResponse } from '../types/card';
import { OhHellPage } from './OhHellPage';

vi.mock('../api/gameApi', () => ({
  ohHellApi: { exec: vi.fn() },
  actionLogApi: { ohhell: vi.fn() },
}));

const mockExec = vi.mocked(ohHellApi.exec);

const playPhaseState: OhHellResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 5,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
      ],
      bid: 2,
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 5,
      cards: [],
      bid: 1,
      roundScore: 3,
      cumulativeScore: 10,
      trickCount: 1,
    },
    {
      id: 2,
      isHuman: false,
      cardCount: 5,
      cards: [],
      bid: 2,
      roundScore: 5,
      cumulativeScore: 20,
      trickCount: 2,
    },
    {
      id: 3,
      isHuman: false,
      cardCount: 5,
      cards: [],
      bid: 0,
      roundScore: 0,
      cumulativeScore: 5,
      trickCount: 0,
    },
  ],
  phase: 1,
  roundNumber: 3,
  totalRounds: 19,
  handSize: 5,
  trickNumber: 1,
  currentPlayerIdx: 0,
  bidPlayerIdx: 0,
  dealerIdx: 3,
  currentTrick: [],
  trumpCard: { design: 'HEART', value: 7 },
  trumpSuit: 3,
  restrictedBid: -1,
  gameEndFlag: false,
  winnerIdx: -1,
  leadPlayerIdx: 0,
  message: '',
  config: { cpuDifficulty: 1, maxHandSize: 10, scoringVariant: 0, roundDirection: 1 },
};

const bidPhaseState: OhHellResponse = {
  ...playPhaseState,
  phase: 0,
  bidPlayerIdx: 0,
  restrictedBid: -1,
  players: playPhaseState.players.map((p) => ({ ...p, bid: -1 })),
};

const bidPhaseDealerState: OhHellResponse = {
  ...bidPhaseState,
  dealerIdx: 0,
  bidPlayerIdx: 0,
  restrictedBid: 3,
};

const bidPhaseCpuTurnState: OhHellResponse = {
  ...bidPhaseState,
  bidPlayerIdx: 1,
};

const trickEndState: OhHellResponse = {
  ...playPhaseState,
  phase: 2,
  currentTrick: [
    { playerIdx: 0, card: { design: 'DIAMOND', value: 3 } },
    { playerIdx: 1, card: { design: 'HEART', value: 5 } },
  ],
};

const roundEndState: OhHellResponse = {
  ...playPhaseState,
  phase: 3,
};

const gameEndState: OhHellResponse = {
  ...playPhaseState,
  phase: 4,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};

const gameEndByFlagState: OhHellResponse = {
  ...playPhaseState,
  phase: 1,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};

const cpuTurnState: OhHellResponse = {
  ...playPhaseState,
  currentPlayerIdx: 1,
};

beforeEach(() => {
  mockExec.mockResolvedValue(playPhaseState);
});

describe('OhHellPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<OhHellPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<OhHellPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        maxHandSize: 10,
        scoringVariant: 0,
        roundDirection: 1,
      }),
    );
  });

  it('renders play phase with human cards', async () => {
    renderWithProviders(<OhHellPage />);
    await waitFor(() => {
      expect(screen.getByAltText('\u2660 A')).toBeInTheDocument();
      expect(screen.getByAltText('\u2665 J')).toBeInTheDocument();
    });
  });

  it('shows the bid-progress chip in the header during play', async () => {
    renderWithProviders(<OhHellPage />);
    // Human bid 2, won 0, 5 cards left \u2192 still achievable \u2192 neutral info colors.
    const chip = await screen.findByTestId('bid-progress-chip');
    expect(chip).toHaveTextContent('\u5ba3\u8a00: 2 / \u7372\u5f97: 0');
    expect(chip.className).toContain('border-ds-border-subtle');
  });

  it('colors the progress chip green when tricks match the bid', async () => {
    mockExec.mockResolvedValue({
      ...playPhaseState,
      players: [{ ...playPhaseState.players[0], trickCount: 2 }, ...playPhaseState.players.slice(1)],
    });
    renderWithProviders(<OhHellPage />);
    const chip = await screen.findByTestId('bid-progress-chip');
    expect(chip).toHaveTextContent('\u5ba3\u8a00: 2 / \u7372\u5f97: 2');
    expect(chip.className).toContain('border-ds-success');
  });

  it('colors the progress chip yellow when tricks exceed the bid', async () => {
    mockExec.mockResolvedValue({
      ...playPhaseState,
      players: [{ ...playPhaseState.players[0], trickCount: 3 }, ...playPhaseState.players.slice(1)],
    });
    renderWithProviders(<OhHellPage />);
    const chip = await screen.findByTestId('bid-progress-chip');
    expect(chip.className).toContain('border-ds-warning');
  });

  it('colors the progress chip red when the bid is no longer reachable', async () => {
    mockExec.mockResolvedValue({
      ...playPhaseState,
      players: [{ ...playPhaseState.players[0], cardCount: 1 }, ...playPhaseState.players.slice(1)],
    });
    renderWithProviders(<OhHellPage />);
    // Needs 2 more tricks with only 1 card left.
    const chip = await screen.findByTestId('bid-progress-chip');
    expect(chip.className).toContain('border-ds-error');
  });

  it('does not turn red while the human card in the unresolved trick can still win', async () => {
    // 1 card in hand + 1 card already played into the open trick → 2 winnable tricks, bid 2 reachable.
    mockExec.mockResolvedValue({
      ...playPhaseState,
      currentTrick: [{ playerIdx: 0, card: { design: 'SPADE', value: 9 } }],
      players: [{ ...playPhaseState.players[0], cardCount: 1 }, ...playPhaseState.players.slice(1)],
    });
    renderWithProviders(<OhHellPage />);
    const chip = await screen.findByTestId('bid-progress-chip');
    expect(chip.className).toContain('border-ds-border-subtle');
    expect(chip.className).not.toContain('border-ds-error');
  });

  it('hides the progress chip during the bid phase and at game end', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    let view = renderWithProviders(<OhHellPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
    expect(screen.queryByTestId('bid-progress-chip')).not.toBeInTheDocument();
    view.unmount();

    mockExec.mockResolvedValue(gameEndState);
    view = renderWithProviders(<OhHellPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
    expect(screen.queryByTestId('bid-progress-chip')).not.toBeInTheDocument();
    view.unmount();

    mockExec.mockResolvedValue(gameEndByFlagState);
    renderWithProviders(<OhHellPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
    expect(screen.queryByTestId('bid-progress-chip')).not.toBeInTheDocument();
  });

  it('renders bid phase as a button group of bid choices', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<OhHellPage />);
    await waitFor(() => {
      // handSize = 5 \u2192 buttons 0..5
      for (let i = 0; i <= 5; i++) {
        expect(screen.getByRole('button', { name: `\u30d3\u30c3\u30c9 ${i}` })).toBeInTheDocument();
      }
      // No legacy number input
      expect(screen.queryByLabelText('bid-input')).not.toBeInTheDocument();
    });
  });

  it('shows the bid-total chip during the bid phase (under when nothing bid yet)', async () => {
    mockExec.mockResolvedValue(bidPhaseState); // handSize 5, no bids placed
    renderWithProviders(<OhHellPage />);
    const chip = await screen.findByTestId('bid-total-chip');
    expect(chip).toHaveTextContent('総ビッド 0 / 手札 5');
    expect(chip).toHaveTextContent('(アンダー)');
  });

  it('colors the bid-total chip for over and exact tables', async () => {
    // Three players bid totaling 7 vs handSize 5 -> over.
    mockExec.mockResolvedValue({
      ...bidPhaseState,
      players: bidPhaseState.players.map((p, i) => ({ ...p, bid: i < 3 ? [3, 2, 2, -1][i] : -1 })),
    });
    const { unmount } = renderWithProviders(<OhHellPage />);
    expect(await screen.findByTestId('bid-total-chip')).toHaveTextContent('(オーバー)');
    unmount();

    // Bids totaling exactly 5.
    mockExec.mockResolvedValue({
      ...bidPhaseState,
      players: bidPhaseState.players.map((p, i) => ({ ...p, bid: [2, 2, 1, -1][i] })),
    });
    renderWithProviders(<OhHellPage />);
    expect(await screen.findByTestId('bid-total-chip')).toHaveTextContent('(ぴったり)');
  });

  it('makes the restricted bid focusable via aria-disabled and ignores its click', async () => {
    mockExec.mockResolvedValue(bidPhaseDealerState);
    renderWithProviders(<OhHellPage />);
    const restricted = await screen.findByTestId('ohhell-restricted-bid');
    // aria-disabled (stays focusable), NOT HTML-disabled; reason in the label.
    expect(restricted).toHaveAttribute('aria-disabled', 'true');
    expect(restricted).not.toBeDisabled();
    expect(restricted).toHaveAttribute(
      'aria-label',
      '3 \u3092\u30d3\u30c3\u30c9\uff08\u30c7\u30a3\u30fc\u30e9\u30fc\u5236\u7d04\u306b\u3088\u308a\u9078\u629e\u3067\u304d\u307e\u305b\u3093\uff09',
    );
    expect(restricted).toHaveAttribute(
      'title',
      '\u30c7\u30a3\u30fc\u30e9\u30fc\u5236\u7d04\u306e\u305f\u3081\u9078\u629e\u3067\u304d\u307e\u305b\u3093',
    );

    // Clicking the restricted bid dispatches nothing.
    mockExec.mockClear();
    fireEvent.click(restricted);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    // Other choices remain enabled and normal.
    expect(screen.getByRole('button', { name: '\u30d3\u30c3\u30c9 0' })).not.toBeDisabled();
    expect(screen.getByRole('button', { name: '\u30d3\u30c3\u30c9 5' })).not.toBeDisabled();
  });

  it('shows bid phase instruction when human bid turn', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<OhHellPage />);
    await waitFor(() => {
      expect(screen.getByText('\u30d3\u30c3\u30c9\u5ba3\u8a00 (0-5)')).toBeInTheDocument();
    });
  });

  it('shows restricted bid info for dealer', async () => {
    mockExec.mockResolvedValue(bidPhaseDealerState);
    renderWithProviders(<OhHellPage />);
    await waitFor(() => {
      expect(
        screen.getByText(
          '\u203b \u30c7\u30a3\u30fc\u30e9\u30fc\u306f\u30d3\u30c3\u30c9 3 \u3092\u9078\u629e\u3067\u304d\u307e\u305b\u3093',
        ),
      ).toBeInTheDocument();
    });
  });

  it('does not show bid instruction when cpu bid turn', async () => {
    mockExec.mockResolvedValue(bidPhaseCpuTurnState);
    renderWithProviders(<OhHellPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByText(/\u30d3\u30c3\u30c9\u5ba3\u8a00/)).not.toBeInTheDocument();
  });

  it('calls bid command with the clicked button value', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<OhHellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30d3\u30c3\u30c9 3' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30d3\u30c3\u30c9 3' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', 3));
  });

  it('play button disabled when not 1 card selected', async () => {
    renderWithProviders(<OhHellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u51fa\u3059' })).toBeDisabled());
  });

  it('play button enabled when 1 card selected', async () => {
    renderWithProviders(<OhHellPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement);
    expect(screen.getByRole('button', { name: '\u51fa\u3059' })).not.toBeDisabled();
  });

  it('does not show play button when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<OhHellPage />);
    await waitFor(() => expect(screen.getByText('CPU 1')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '\u51fa\u3059' })).not.toBeInTheDocument();
  });

  it('shows next trick button on trick end', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<OhHellPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '\u6b21\u306e\u30c8\u30ea\u30c3\u30af' })).toBeInTheDocument(),
    );
  });

  it('highlights the trick winner card at trick end', async () => {
    // trickEndState.leadPlayerIdx is 0, so the trick winner badge appears.
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<OhHellPage />);
    await waitFor(() => expect(screen.getByTestId('trick-winner-badge')).toBeInTheDocument());
  });

  it('does not highlight a trick winner during active play', async () => {
    mockExec.mockResolvedValue({
      ...trickEndState,
      phase: 1, // PLAY
    });
    renderWithProviders(<OhHellPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('trick-winner-badge')).not.toBeInTheDocument();
  });

  it('shows next round button on round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<OhHellPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '\u6b21\u306e\u30e9\u30a6\u30f3\u30c9' })).toBeInTheDocument(),
    );
  });

  it('shows game end with action log button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<OhHellPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b')).toBeInTheDocument();
    });
  });

  it('shows game end via gameEndFlag with non-4 phase', async () => {
    mockExec.mockResolvedValue(gameEndByFlagState);
    renderWithProviders(<OhHellPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b')).toBeInTheDocument();
    });
  });

  it('shows error alert', async () => {
    renderWithProviders(<OhHellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('settings panel changes cpuDifficulty', async () => {
    renderWithProviders(<OhHellPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());

    fireEvent.click(screen.getByText('\u8a2d\u5b9a'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[0], { target: { value: '2' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 2,
        maxHandSize: 10,
        scoringVariant: 0,
        roundDirection: 1,
      }),
    );
  });

  it('settings panel changes maxHandSize', async () => {
    renderWithProviders(<OhHellPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());

    fireEvent.click(screen.getByText('\u8a2d\u5b9a'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[1], { target: { value: '13' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        maxHandSize: 13,
        scoringVariant: 0,
        roundDirection: 1,
      }),
    );
  });

  it('card selection toggle via aria-pressed', async () => {
    renderWithProviders(<OhHellPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('reset button calls exec', async () => {
    renderWithProviders(<OhHellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        maxHandSize: 10,
        scoringVariant: 0,
        roundDirection: 1,
      }),
    );
  });

  it('score table shows all players', async () => {
    renderWithProviders(<OhHellPage />);
    await waitFor(() => {
      expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument();
      expect(screen.getByText('\u3042\u306a\u305f')).toBeInTheDocument();
      expect(screen.getByText('CPU 1')).toBeInTheDocument();
      expect(screen.getByText('CPU 2')).toBeInTheDocument();
      expect(screen.getByText('CPU 3')).toBeInTheDocument();
    });
  });

  it('score table headers have scope="col" for accessibility', async () => {
    const { container } = renderWithProviders(<OhHellPage />);
    await waitFor(() => expect(screen.getByText('\u3042\u306a\u305f')).toBeInTheDocument());
    const ths = container.querySelectorAll('th');
    ths.forEach((th) => {
      expect(th).toHaveAttribute('scope', 'col');
    });
  });

  it('shows current trick cards', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<OhHellPage />);
    await waitFor(() => {
      expect(screen.getByText('\u73fe\u5728\u306e\u30c8\u30ea\u30c3\u30af')).toBeInTheDocument();
      expect(screen.getByAltText('\u2666 3')).toBeInTheDocument();
      expect(screen.getByAltText('\u2665 5')).toBeInTheDocument();
    });
  });

  it('does not show current trick when empty', async () => {
    renderWithProviders(<OhHellPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByText('\u73fe\u5728\u306e\u30c8\u30ea\u30c3\u30af')).not.toBeInTheDocument();
  });

  it('shows CPU player areas', async () => {
    renderWithProviders(<OhHellPage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1.*5\u679a/)).toBeInTheDocument();
      expect(screen.getByText(/CPU 2.*5\u679a/)).toBeInTheDocument();
      expect(screen.getByText(/CPU 3.*5\u679a/)).toBeInTheDocument();
    });
  });

  it('calls play command when play button is clicked', async () => {
    renderWithProviders(<OhHellPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u51fa\u3059' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 0));
  });

  it('calls next when next trick button is clicked', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<OhHellPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '\u6b21\u306e\u30c8\u30ea\u30c3\u30af' })).toBeInTheDocument(),
    );

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u6b21\u306e\u30c8\u30ea\u30c3\u30af' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('calls nextround when next round button is clicked', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<OhHellPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '\u6b21\u306e\u30e9\u30a6\u30f3\u30c9' })).toBeInTheDocument(),
    );

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u6b21\u306e\u30e9\u30a6\u30f3\u30c9' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('round and trick info displayed', async () => {
    renderWithProviders(<OhHellPage />);
    await waitFor(() => {
      expect(screen.getByText('\u30e9\u30a6\u30f3\u30c9 3/19')).toBeInTheDocument();
      expect(screen.getByText('\u30c8\u30ea\u30c3\u30af 1')).toBeInTheDocument();
    });
  });

  it('trump and dealer info displayed', async () => {
    renderWithProviders(<OhHellPage />);
    await waitFor(() => {
      expect(screen.getByText('\u5207\u308a\u672d: \u30cf\u30fc\u30c8')).toBeInTheDocument();
      expect(screen.getByText('\u30c7\u30a3\u30fc\u30e9\u30fc: CPU 3')).toBeInTheDocument();
    });
  });

  it('shows confirm dialog when reset button is clicked', async () => {
    renderWithProviders(<OhHellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    renderWithProviders(<OhHellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: '\u30ad\u30e3\u30f3\u30bb\u30eb' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('handles action log visibility and API fetch', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<OhHellPage />);
    await waitFor(() => expect(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b')).toBeInTheDocument());

    vi.mocked(actionLogApi.ohhell).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b'));

    await waitFor(() => expect(actionLogApi.ohhell).toHaveBeenCalledTimes(1));
    expect(screen.getByText('\u68cb\u8b5c')).toBeInTheDocument();

    fireEvent.click(screen.getByText('\u9589\u3058\u308b'));
    await waitFor(() => expect(screen.queryByText(/^\u68cb\u8b5c$/)).not.toBeInTheDocument());
    expect(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b')).toBeInTheDocument();
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<OhHellPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<OhHellPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '\u30c1\u30e5\u30fc\u30c8\u30ea\u30a2\u30eb' })).toBeInTheDocument(),
    );
  });

  it('score table shows dash for bid < 0', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<OhHellPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    const rows = screen.getAllByRole('row');
    // Header + 4 players = 5 rows
    expect(rows.length).toBe(5);
  });

  it('shows bid value for player with bid >= 0', async () => {
    renderWithProviders(<OhHellPage />);
    await waitFor(() => {
      // CPU 1 has bid=1
      expect(screen.getByText(/CPU 1.*\u30d3\u30c3\u30c9 1/)).toBeInTheDocument();
    });
  });

  it('shows unbid text for player with bid < 0', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<OhHellPage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1.*\u672a\u30d3\u30c3\u30c9/)).toBeInTheDocument();
    });
  });

  it('renders CPU info as collapsible details on mobile', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    try {
      mockExec.mockResolvedValue(playPhaseState);
      const { container } = renderWithProviders(<OhHellPage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const allDetails = container.querySelectorAll('details');
      const cpuDetails = Array.from(allDetails).find((d) =>
        d.querySelector('summary')?.textContent?.includes('CPU対戦相手'),
      );
      expect(cpuDetails).toBeInTheDocument();
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  it('renders score table as collapsible details on mobile', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    try {
      mockExec.mockResolvedValue(playPhaseState);
      const { container } = renderWithProviders(<OhHellPage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const scoreDetails = container.querySelector('details[data-tutorial="oh-score-table"]');
      expect(scoreDetails).toBeInTheDocument();
      const summary = scoreDetails?.querySelector('summary');
      expect(summary).toHaveTextContent('スコア');
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  // **ヒントは数字をテキストで言うだけで、どのボタンを押せばよいか視覚的に
  // 示していなかった (#4738)。**同じコードベースの TwoTenJack はヒント札を
  // 光らせている。
  it('highlights the bid button the hint recommends', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<OhHellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    // hint はフックが 'hint' コマンドの応答から取る。要求して初めて出る。
    mockExec.mockResolvedValue({ ...bidPhaseState, hint: { bid: 2, reason: 'bid' } });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(document.querySelectorAll('[data-hint-suggested="true"]')).toHaveLength(1));
    expect(screen.getByRole('button', { name: /2/ })).toHaveAttribute('data-hint-suggested', 'true');
  });

  // **制限ビッドとは別状態。**ディーラーが選べないビッドを推奨として光らせると、
  // 押せないボタンを勧めることになる。
  it('does not highlight a restricted bid even if the hint names it', async () => {
    mockExec.mockResolvedValue({ ...bidPhaseState, restrictedBid: 2 });
    renderWithProviders(<OhHellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockResolvedValue({ ...bidPhaseState, restrictedBid: 2, hint: { bid: 2, reason: 'bid' } });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
    expect(document.querySelectorAll('[data-hint-suggested="true"]')).toHaveLength(0);
  });

  it('highlights no bid button before a hint arrives', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<OhHellPage />);

    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(document.querySelectorAll('[data-hint-suggested="true"]')).toHaveLength(0);
  });

  // **カード側のハイライトも踏む。**ビッドボタンだけ確かめて満足すると、
  // 同じ PR で足したもう一方が未検証のまま残る (codecov が partial として検出)。
  it('highlights the recommended card during the play phase', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<OhHellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockResolvedValue({ ...playPhaseState, hint: { cardIndex: 1, reason: 'follow_suit' } });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(document.querySelectorAll('[data-hint-suggested="true"]')).toHaveLength(1));
  });

  it('highlights no card before a play hint arrives', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<OhHellPage />);

    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(document.querySelectorAll('[data-hint-suggested="true"]')).toHaveLength(0);
  });
});
