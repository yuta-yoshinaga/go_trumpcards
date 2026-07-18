import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { bridgeApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BridgeResponse } from '../types/card';
import { BridgePage } from './BridgePage';

vi.mock('../api/gameApi', () => ({
  bridgeApi: { exec: vi.fn() },
  actionLogApi: { bridge: vi.fn() },
}));

const mockExec = vi.mocked(bridgeApi.exec);

const bidPhaseState: BridgeResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 13,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 13 },
      ],
      team: 0,
      trickCount: 0,
    },
    { id: 1, isHuman: false, cardCount: 13, cards: [], team: 1, trickCount: 0 },
    { id: 2, isHuman: false, cardCount: 13, cards: [], team: 0, trickCount: 0 },
    { id: 3, isHuman: false, cardCount: 13, cards: [], team: 1, trickCount: 0 },
  ],
  phase: 0,
  roundNumber: 1,
  trickNumber: 0,
  currentPlayerIdx: 0,
  bidPlayerIdx: 0,
  dealerIdx: 0,
  trumpSuit: 0,
  contractLevel: 0,
  contractSuit: 0,
  doubled: 0,
  declarerIdx: -1,
  dummyIdx: -1,
  bidHistory: [],
  vulnerability: [false, false],
  currentTrick: [],
  teamScores: [0, 0],
  gamesWon: [0, 0],
  belowLine: [0, 0],
  gameEndFlag: false,
  winnerTeam: -1,
  leadPlayerIdx: -1,
  openingLeadDone: false,
  dummyHand: null,
  message: '',
  config: { cpuDifficulty: 1 },
};

const playPhaseState: BridgeResponse = {
  ...bidPhaseState,
  phase: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  contractLevel: 1,
  contractSuit: -1,
  declarerIdx: 0,
  dummyIdx: 2,
  openingLeadDone: true,
  dummyHand: [
    { design: 'DIAMOND', value: 10 },
    { design: 'CLOVER', value: 7 },
  ],
};

const cpuBidTurnState: BridgeResponse = {
  ...bidPhaseState,
  bidPlayerIdx: 1,
};

const cpuPlayTurnState: BridgeResponse = {
  ...playPhaseState,
  currentPlayerIdx: 1,
};

const trickEndState: BridgeResponse = {
  ...playPhaseState,
  phase: 2,
  currentTrick: [
    { playerIdx: 0, card: { design: 'SPADE', value: 1 } },
    { playerIdx: 1, card: { design: 'HEART', value: 5 } },
  ],
};

const roundEndState: BridgeResponse = {
  ...playPhaseState,
  phase: 3,
};

const gameEndState: BridgeResponse = {
  ...playPhaseState,
  phase: 4,
  gameEndFlag: true,
  winnerTeam: 0,
  message: 'Game end!',
};

const gameEndByFlagState: BridgeResponse = {
  ...playPhaseState,
  phase: 1,
  gameEndFlag: true,
  winnerTeam: 0,
  message: 'Game end!',
};

const bidHistoryState: BridgeResponse = {
  ...bidPhaseState,
  bidHistory: [
    { playerIdx: 0, bidType: 1, level: 1, suit: 5 },
    { playerIdx: 1, bidType: 0, level: 0, suit: 0 },
    { playerIdx: 2, bidType: 2, level: 0, suit: 0 },
  ],
};

const vulnerableState: BridgeResponse = {
  ...bidPhaseState,
  vulnerability: [true, false],
};

const doubledContractState: BridgeResponse = {
  ...playPhaseState,
  doubled: 1,
};

const redoubledContractState: BridgeResponse = {
  ...playPhaseState,
  doubled: 2,
};

// Human (seat 0, team 0) to bid; opponent (seat 1, team 1) made the last bid,
// so Double is legal.
const doublableBidTurnState: BridgeResponse = {
  ...bidPhaseState,
  contractLevel: 1,
  contractSuit: 5,
  doubled: 0,
  bidHistory: [{ playerIdx: 1, bidType: 1, level: 1, suit: 5 }],
};

// Human (seat 0, team 0) made the last bid and the opponent doubled it, so
// Redouble is legal for the human.
const redoublableBidTurnState: BridgeResponse = {
  ...bidPhaseState,
  contractLevel: 1,
  contractSuit: 5,
  doubled: 1,
  bidHistory: [
    { playerIdx: 0, bidType: 1, level: 1, suit: 5 },
    { playerIdx: 1, bidType: 2, level: 0, suit: 0 },
  ],
};

beforeEach(() => {
  mockExec.mockResolvedValue(bidPhaseState);
});

describe('BridgePage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<BridgePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<BridgePage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, undefined, {
        cpuDifficulty: 1,
      }),
    );
  });

  it('renders bid phase with bid controls when human turn', async () => {
    renderWithProviders(<BridgePage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '\u30d3\u30c3\u30c9' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '\u30d1\u30b9' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '\u30c0\u30d6\u30eb' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '\u30ea\u30c0\u30d6\u30eb' })).toBeInTheDocument();
    });
  });

  it('shows bid phase instruction when human bid turn', async () => {
    renderWithProviders(<BridgePage />);
    await waitFor(() => {
      expect(screen.getByText('\u30d3\u30c3\u30c9\u3057\u3066\u304f\u3060\u3055\u3044')).toBeInTheDocument();
    });
  });

  it('does not show bid controls when CPU bid turn', async () => {
    mockExec.mockResolvedValue(cpuBidTurnState);
    renderWithProviders(<BridgePage />);
    await waitFor(() => expect(screen.getByText('\u30c1\u30fc\u30e0\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '\u30d3\u30c3\u30c9' })).not.toBeInTheDocument();
  });

  it('calls bid command when bid button is clicked', async () => {
    renderWithProviders(<BridgePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30d3\u30c3\u30c9' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30d3\u30c3\u30c9' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', undefined, 1, 1, 5));
  });

  it('calls pass command when pass button is clicked', async () => {
    renderWithProviders(<BridgePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30d1\u30b9' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30d1\u30b9' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', undefined, 0, undefined, undefined));
  });

  it('calls double command when double button is clicked', async () => {
    mockExec.mockResolvedValue(doublableBidTurnState);
    renderWithProviders(<BridgePage />);
    await waitFor(() => expect(screen.getByTestId('br-double')).toBeEnabled());

    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30c0\u30d6\u30eb' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', undefined, 2, undefined, undefined));
  });

  it('disables Double when no opposing bid precedes it', async () => {
    // bidPhaseState has no bid yet -> Double is illegal.
    renderWithProviders(<BridgePage />);
    await waitFor(() => expect(screen.getByTestId('br-double')).toBeInTheDocument());
    expect(screen.getByTestId('br-double')).toBeDisabled();
    expect(screen.getByTestId('br-double')).toHaveAttribute(
      'title',
      '\u76f4\u524d\u306b\u76f8\u624b\u5074\u306e\u30d3\u30c3\u30c9\u304c\u306a\u3044\u305f\u3081\u30c0\u30d6\u30eb\u3067\u304d\u307e\u305b\u3093',
    );
  });

  it('enables Double after an opposing bid', async () => {
    mockExec.mockResolvedValue(doublableBidTurnState);
    renderWithProviders(<BridgePage />);
    await waitFor(() => expect(screen.getByTestId('br-double')).toBeEnabled());
  });

  it('disables Redouble when the contract is not doubled', async () => {
    renderWithProviders(<BridgePage />);
    await waitFor(() => expect(screen.getByTestId('br-redouble')).toBeInTheDocument());
    expect(screen.getByTestId('br-redouble')).toBeDisabled();
  });

  it('enables Redouble when our bid has been doubled by the opponent', async () => {
    mockExec.mockResolvedValue(redoublableBidTurnState);
    renderWithProviders(<BridgePage />);
    await waitFor(() => expect(screen.getByTestId('br-redouble')).toBeEnabled());
  });

  it('disables the bid button when the selection is at or below the current contract', async () => {
    // Contract already 1NT (level 1, suit 5); default selectors are level 1 / NT (suit 5),
    // which is not higher, so the bid button is disabled.
    mockExec.mockResolvedValue(doublableBidTurnState);
    renderWithProviders(<BridgePage />);
    await waitFor(() => expect(screen.getByTestId('br-bid-submit')).toBeDisabled());
  });

  it('renders play phase with human cards', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<BridgePage />);
    await waitFor(() => {
      expect(screen.getByAltText('\u2660 A')).toBeInTheDocument();
      expect(screen.getByAltText('\u2665 K')).toBeInTheDocument();
    });
  });

  it('renders the human hand via MobileHandGrid on a narrow mobile viewport', async () => {
    const original = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    try {
      mockExec.mockResolvedValue(playPhaseState);
      renderWithProviders(<BridgePage />);
      // MobileHandGrid lays the hand out in fanned rows instead of a scrolling strip.
      await waitFor(() => expect(screen.getAllByTestId('hand-row').length).toBeGreaterThan(0));
      expect(screen.getByAltText('\u2660 A')).toBeInTheDocument();
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: original });
    }
  });

  it('play button disabled when not 1 card selected', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<BridgePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u51fa\u3059' })).toBeDisabled());
  });

  it('play button enabled when 1 card selected', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<BridgePage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement);
    expect(screen.getByRole('button', { name: '\u51fa\u3059' })).not.toBeDisabled();
  });

  it('does not show play button when not human turn', async () => {
    mockExec.mockResolvedValue(cpuPlayTurnState);
    renderWithProviders(<BridgePage />);
    await waitFor(() => expect(screen.getByText('\u30c1\u30fc\u30e0\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '\u51fa\u3059' })).not.toBeInTheDocument();
  });

  it('shows next trick button on trick end', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<BridgePage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '\u6b21\u306e\u30c8\u30ea\u30c3\u30af' })).toBeInTheDocument(),
    );
  });

  it('shows next round button on round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<BridgePage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '\u6b21\u306e\u30e9\u30a6\u30f3\u30c9' })).toBeInTheDocument(),
    );
  });

  it('shows game end with action log button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<BridgePage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b')).toBeInTheDocument();
    });
  });

  it('shows game end via gameEndFlag with non-4 phase', async () => {
    mockExec.mockResolvedValue(gameEndByFlagState);
    renderWithProviders(<BridgePage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b')).toBeInTheDocument();
    });
  });

  it('shows error alert', async () => {
    renderWithProviders(<BridgePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('settings panel changes cpuDifficulty', async () => {
    renderWithProviders(<BridgePage />);
    await waitFor(() => expect(screen.getByText('\u30c1\u30fc\u30e0\u30b9\u30b3\u30a2')).toBeInTheDocument());

    fireEvent.click(screen.getByText('\u8a2d\u5b9a'));
    const selects = screen.getAllByRole('combobox');
    const difficultySelect = selects.find((s) => s.getAttribute('id') === 'cpuDifficulty') ?? selects[0];
    fireEvent.change(difficultySelect, { target: { value: '2' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, undefined, {
        cpuDifficulty: 2,
      }),
    );
  });

  it('card selection toggle via aria-pressed', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<BridgePage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('card buttons have aria-label with card name', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<BridgePage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-label', '\u2660 A');

    const cardBtn2 = screen.getByAltText('\u2665 K').closest('button') as HTMLButtonElement;
    expect(cardBtn2).toHaveAttribute('aria-label', '\u2665 K');
  });

  it('reset button calls apiExec', async () => {
    renderWithProviders(<BridgePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, undefined, {
        cpuDifficulty: 1,
      }),
    );
  });

  it('shows dummy hand when opening lead done', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<BridgePage />);
    await waitFor(() => {
      expect(screen.getByText('\u30c0\u30df\u30fc\u306e\u624b\u672d')).toBeInTheDocument();
    });
  });

  it('shows bid history when bids exist', async () => {
    mockExec.mockResolvedValue(bidHistoryState);
    renderWithProviders(<BridgePage />);
    await waitFor(() => {
      expect(screen.getByText('\u30d3\u30c3\u30c9\u5c65\u6b74')).toBeInTheDocument();
    });
  });

  it('shows vulnerability info', async () => {
    mockExec.mockResolvedValue(vulnerableState);
    renderWithProviders(<BridgePage />);
    await waitFor(() => {
      const allText = document.body.textContent ?? '';
      expect(allText).toContain('\u30d0\u30eb');
      expect(allText).toContain('\u30ce\u30f3\u30d0\u30eb');
    });
  });

  it('shows doubled contract display', async () => {
    mockExec.mockResolvedValue(doubledContractState);
    renderWithProviders(<BridgePage />);
    await waitFor(() => {
      expect(screen.getByText(/\u30c0\u30d6\u30eb/)).toBeInTheDocument();
    });
  });

  it('shows redoubled contract display', async () => {
    mockExec.mockResolvedValue(redoubledContractState);
    renderWithProviders(<BridgePage />);
    await waitFor(() => {
      expect(screen.getByText(/\u30ea\u30c0\u30d6\u30eb/)).toBeInTheDocument();
    });
  });

  it('shows declarer and dummy info when set', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<BridgePage />);
    await waitFor(() => {
      const allText = document.body.textContent ?? '';
      expect(allText).toContain('\u30c7\u30a3\u30af\u30ec\u30a2\u30e9\u30fc');
      expect(allText).toContain('\u30c0\u30df\u30fc');
    });
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<BridgePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<BridgePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('tutorial can be skipped', async () => {
    renderWithProviders(<BridgePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders CPU info as collapsible details on mobile', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    try {
      mockExec.mockResolvedValue(playPhaseState);
      const { container } = renderWithProviders(<BridgePage />);
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
      const { container } = renderWithProviders(<BridgePage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const scoreDetails = container.querySelector('details[data-tutorial="br-team-scores"]');
      expect(scoreDetails).toBeInTheDocument();
      const summary = scoreDetails?.querySelector('summary');
      expect(summary).toHaveTextContent('チームスコア');
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });
});
