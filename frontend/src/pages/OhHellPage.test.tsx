import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, ohHellApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
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

  it('renders bid phase with bid button and input', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<OhHellPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '\u30d3\u30c3\u30c9' })).toBeInTheDocument();
      expect(screen.getByLabelText('bid-input')).toBeInTheDocument();
    });
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

  it('calls bid command when bid button is clicked', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<OhHellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30d3\u30c3\u30c9' })).toBeInTheDocument());

    const input = screen.getByLabelText('bid-input');
    fireEvent.change(input, { target: { value: '3' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30d3\u30c3\u30c9' }));

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
});
