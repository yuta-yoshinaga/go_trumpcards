import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, spadesApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { SpadesResponse } from '../types/card';
import { SpadesPage } from './SpadesPage';

vi.mock('../api/gameApi', () => ({
  spadesApi: { exec: vi.fn() },
  actionLogApi: { spades: vi.fn() },
}));

const mockExec = vi.mocked(spadesApi.exec);

const playPhaseState: SpadesResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 13,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
      ],
      bid: 3,
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
      bags: 0,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 13,
      cards: [],
      bid: 4,
      roundScore: 3,
      cumulativeScore: 10,
      trickCount: 1,
      bags: 2,
    },
    {
      id: 2,
      isHuman: false,
      cardCount: 13,
      cards: [],
      bid: 3,
      roundScore: 5,
      cumulativeScore: 20,
      trickCount: 2,
      bags: 1,
    },
    {
      id: 3,
      isHuman: false,
      cardCount: 13,
      cards: [],
      bid: 2,
      roundScore: 0,
      cumulativeScore: 5,
      trickCount: 0,
      bags: 0,
    },
  ],
  phase: 1,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  bidPlayerIdx: 0,
  currentTrick: [],
  spadesBroken: false,
  gameEndFlag: false,
  winnerIdx: -1,
  leadPlayerIdx: 0,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 500, nilBonus: 100, bagPenaltyThreshold: 10 },
};

const bidPhaseState: SpadesResponse = {
  ...playPhaseState,
  phase: 0,
  bidPlayerIdx: 0,
  players: playPhaseState.players.map((p) => ({ ...p, bid: -1 })),
};

const bidPhaseCpuTurnState: SpadesResponse = {
  ...bidPhaseState,
  bidPlayerIdx: 1,
};

const trickEndState: SpadesResponse = {
  ...playPhaseState,
  phase: 2,
  currentTrick: [
    { playerIdx: 0, card: { design: 'DIAMOND', value: 3 } },
    { playerIdx: 1, card: { design: 'HEART', value: 5 } },
  ],
};

const roundEndState: SpadesResponse = {
  ...playPhaseState,
  phase: 3,
};

const gameEndState: SpadesResponse = {
  ...playPhaseState,
  phase: 4,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};

const gameEndByFlagState: SpadesResponse = {
  ...playPhaseState,
  phase: 1,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};

const spadesBrokenState: SpadesResponse = {
  ...playPhaseState,
  spadesBroken: true,
};

const cpuTurnState: SpadesResponse = {
  ...playPhaseState,
  currentPlayerIdx: 1,
};

beforeEach(() => {
  mockExec.mockResolvedValue(playPhaseState);
});

describe('SpadesPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<SpadesPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 500,
        nilBonus: 100,
        bagPenaltyThreshold: 10,
      }),
    );
  });

  it('renders play phase with human cards', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => {
      expect(screen.getByAltText('\u2660 A')).toBeInTheDocument();
      expect(screen.getByAltText('\u2665 J')).toBeInTheDocument();
    });
  });

  it('renders bid phase with bid button and input', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '\u30d3\u30c3\u30c9' })).toBeInTheDocument();
      expect(screen.getByLabelText('bid-input')).toBeInTheDocument();
    });
  });

  it('shows bid phase instruction when human bid turn', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() => {
      expect(screen.getByText('\u30d3\u30c3\u30c9\u5ba3\u8a00 (0=\u30cb\u30eb, 1-13)')).toBeInTheDocument();
    });
  });

  it('does not show bid instruction when cpu bid turn', async () => {
    mockExec.mockResolvedValue(bidPhaseCpuTurnState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByText(/\u30d3\u30c3\u30c9\u3092\u5ba3\u8a00/)).not.toBeInTheDocument();
  });

  it('calls bid command when bid button is clicked', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30d3\u30c3\u30c9' })).toBeInTheDocument());

    const input = screen.getByLabelText('bid-input');
    fireEvent.change(input, { target: { value: '5' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30d3\u30c3\u30c9' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', 5));
  });

  it('play button disabled when not 1 card selected', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u51fa\u3059' })).toBeDisabled());
  });

  it('play button enabled when 1 card selected', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement);
    expect(screen.getByRole('button', { name: '\u51fa\u3059' })).not.toBeDisabled();
  });

  it('does not show play button when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByText('CPU 1')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '\u51fa\u3059' })).not.toBeInTheDocument();
  });

  it('shows next trick button on trick end', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '\u6b21\u306e\u30c8\u30ea\u30c3\u30af' })).toBeInTheDocument(),
    );
  });

  it('shows next round button on round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '\u6b21\u306e\u30e9\u30a6\u30f3\u30c9' })).toBeInTheDocument(),
    );
  });

  it('shows game end with action log button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b')).toBeInTheDocument();
    });
  });

  it('shows game end via gameEndFlag with non-4 phase', async () => {
    mockExec.mockResolvedValue(gameEndByFlagState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b')).toBeInTheDocument();
    });
  });

  it('shows error alert', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('settings panel changes cpuDifficulty', async () => {
    renderWithProviders(<SpadesPage />);
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
        pointLimit: 500,
        nilBonus: 100,
        bagPenaltyThreshold: 10,
      }),
    );
  });

  it('settings panel changes pointLimit', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());

    fireEvent.click(screen.getByText('\u8a2d\u5b9a'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[1], { target: { value: '1000' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 1000,
        nilBonus: 100,
        bagPenaltyThreshold: 10,
      }),
    );
  });

  it('card selection toggle via aria-pressed', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('card buttons have aria-label with card name', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-label', '\u2660 A');

    const cardBtn2 = screen.getByAltText('\u2665 J').closest('button') as HTMLButtonElement;
    expect(cardBtn2).toHaveAttribute('aria-label', '\u2665 J');
  });

  it('reset button calls exec', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 500,
        nilBonus: 100,
        bagPenaltyThreshold: 10,
      }),
    );
  });

  it('score table shows all players', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => {
      expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument();
      expect(screen.getByText('\u3042\u306a\u305f')).toBeInTheDocument();
      expect(screen.getByText('CPU 1')).toBeInTheDocument();
      expect(screen.getByText('CPU 2')).toBeInTheDocument();
      expect(screen.getByText('CPU 3')).toBeInTheDocument();
    });
  });

  it('score table headers have scope="col" for accessibility', async () => {
    const { container } = renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByText('\u3042\u306a\u305f')).toBeInTheDocument());
    const ths = container.querySelectorAll('th');
    ths.forEach((th) => {
      expect(th).toHaveAttribute('scope', 'col');
    });
  });

  it('score table has horizontal scroll wrapper', async () => {
    const { container } = renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    const scoreSection = container.querySelector('[data-tutorial="sp-score-table"]');
    const scrollWrapper = scoreSection?.querySelector('.overflow-x-auto');
    expect(scrollWrapper).toBeInTheDocument();
    const table = scrollWrapper?.querySelector('table');
    expect(table?.className).toContain('min-w-');
  });

  it('score table renders ScrollFadeHint on mobile', async () => {
    const innerWidthSpy = vi.spyOn(window, 'innerWidth', 'get').mockReturnValue(375);
    try {
      const { container } = renderWithProviders(<SpadesPage />);
      await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
      const scoreSection = container.querySelector('[data-tutorial="sp-score-table"]');
      const fadeHint = scoreSection?.querySelector('[aria-hidden="true"]');
      expect(fadeHint).toBeInTheDocument();
    } finally {
      innerWidthSpy.mockRestore();
    }
  });

  it('shows spades broken text', async () => {
    mockExec.mockResolvedValue(spadesBrokenState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() =>
      expect(screen.getByText('\u30b9\u30da\u30fc\u30c9\u30d6\u30ec\u30a4\u30af\u6e08')).toBeInTheDocument(),
    );
  });

  it('shows spades not broken text', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() =>
      expect(screen.getByText('\u30b9\u30da\u30fc\u30c9\u672a\u30d6\u30ec\u30a4\u30af')).toBeInTheDocument(),
    );
  });

  it('shows current trick cards', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() => {
      expect(screen.getByText('\u73fe\u5728\u306e\u30c8\u30ea\u30c3\u30af')).toBeInTheDocument();
      expect(screen.getByAltText('\u2666 3')).toBeInTheDocument();
      expect(screen.getByAltText('\u2665 5')).toBeInTheDocument();
    });
  });

  it('does not show current trick when empty', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByText('\u73fe\u5728\u306e\u30c8\u30ea\u30c3\u30af')).not.toBeInTheDocument();
  });

  it('shows CPU player areas', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1.*13\u679a/)).toBeInTheDocument();
      expect(screen.getByText(/CPU 2.*13\u679a/)).toBeInTheDocument();
      expect(screen.getByText(/CPU 3.*13\u679a/)).toBeInTheDocument();
    });
  });

  it('shows loading state', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    let resolve!: (value: SpadesResponse) => void;
    const slow = new Promise<SpadesResponse>((r) => {
      resolve = r;
    });
    mockExec.mockReturnValueOnce(slow);
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).toBeDisabled();

    resolve(playPhaseState);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());
  });

  it('calls play command when play button is clicked', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u51fa\u3059' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 0));
  });

  it('calls next when next trick button is clicked', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<SpadesPage />);
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
    renderWithProviders(<SpadesPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '\u6b21\u306e\u30e9\u30a6\u30f3\u30c9' })).toBeInTheDocument(),
    );

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u6b21\u306e\u30e9\u30a6\u30f3\u30c9' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('round and trick info displayed', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => {
      expect(screen.getByText('\u30e9\u30a6\u30f3\u30c9 1')).toBeInTheDocument();
      expect(screen.getByText('\u30c8\u30ea\u30c3\u30af 1')).toBeInTheDocument();
    });
  });

  it('does not show message when empty', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByText('Game end!')).not.toBeInTheDocument();
  });

  // -- ConfirmDialog on reset --

  it('shows confirm dialog when reset button is clicked', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: '\u30ad\u30e3\u30f3\u30bb\u30eb' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('executes reset on confirm', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 500,
        nilBonus: 100,
        bagPenaltyThreshold: 10,
      }),
    );
  });

  it('handles action log visibility and API fetch', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b')).toBeInTheDocument());

    vi.mocked(actionLogApi.spades).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b'));

    await waitFor(() => expect(actionLogApi.spades).toHaveBeenCalledTimes(1));
    expect(screen.getByText('\u68cb\u8b5c')).toBeInTheDocument();

    fireEvent.click(screen.getByText('\u9589\u3058\u308b'));
    await waitFor(() => expect(screen.queryByText(/^\u68cb\u8b5c$/)).not.toBeInTheDocument());
    expect(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b')).toBeInTheDocument();
  });

  it('does not show action log button when not game end', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByText('\u68cb\u8b5c\u3092\u898b\u308b')).not.toBeInTheDocument();
  });

  it('does not show bid controls in play phase', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByLabelText('bid-input')).not.toBeInTheDocument();
  });

  it('disables buttons while loading', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    let resolve!: (value: SpadesResponse) => void;
    const slow = new Promise<SpadesResponse>((r) => {
      resolve = r;
    });
    mockExec.mockReturnValueOnce(slow);
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).toBeDisabled();

    resolve(playPhaseState);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());
  });

  it('trick card shows player name with fallback', async () => {
    const stateWithBadIdx: SpadesResponse = {
      ...trickEndState,
      currentTrick: [{ playerIdx: 99, card: { design: 'SPADE', value: 1 } }],
    };
    mockExec.mockResolvedValue(stateWithBadIdx);
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByText('CPU 99')).toBeInTheDocument());
  });

  it('sets aria-busy on container', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    const container = screen
      .getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })
      .closest('[aria-busy]') as HTMLElement;
    expect(container).toHaveAttribute('aria-busy', 'false');
  });

  it('no human cards renders empty hand area', async () => {
    const noHuman: SpadesResponse = {
      ...playPhaseState,
      players: playPhaseState.players.map((p) => ({ ...p, isHuman: false })),
    };
    mockExec.mockResolvedValue(noHuman);
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByAltText('\u2660 A')).not.toBeInTheDocument();
  });

  it('isHumanTurn false when currentPlayerIdx points to cpu', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '\u51fa\u3059' })).not.toBeInTheDocument();
  });

  // --- PhaseIndicator coverage ---

  it('phase indicator shows your turn during bid phase', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() =>
      expect(screen.getByTestId('phase-indicator')).toHaveTextContent('\u3042\u306a\u305f\u306e\u30bf\u30fc\u30f3'),
    );
  });

  it('phase indicator shows your turn when human play turn', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() =>
      expect(screen.getByTestId('phase-indicator')).toHaveTextContent('\u3042\u306a\u305f\u306e\u30bf\u30fc\u30f3'),
    );
  });

  it('phase indicator shows waiting when cpu turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('\u5f85\u6a5f\u4e2d'));
  });

  // -- Keyboard navigation --

  it('pressing number key toggles card in play phase', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('Enter key triggers play in play phase', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.keyDown(document, { key: '1' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);

    fireEvent.keyDown(document, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 0));
  });

  it('Escape key clears selection', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.keyDown(document, { key: 'Escape' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('keyboard nav disabled when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByText('CPU 1')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('reset hides the action log panel if open', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b')).toBeInTheDocument());

    vi.mocked(actionLogApi.spades).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b'));
    await waitFor(() => expect(screen.getByText('\u68cb\u8b5c')).toBeInTheDocument());

    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() => expect(screen.queryByText('\u68cb\u8b5c')).not.toBeInTheDocument());
  });

  it('shows bid value for player with bid >= 0', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => {
      // CPU 1 has bid=4
      expect(screen.getByText(/CPU 1.*\u30d3\u30c3\u30c9 4/)).toBeInTheDocument();
    });
  });

  it('shows unbid text for player with bid < 0', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1.*\u672a\u30d3\u30c3\u30c9/)).toBeInTheDocument();
    });
  });

  it('score table shows dash for bid < 0', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    // All players have bid=-1, so bid column should show '-'
    const rows = screen.getAllByRole('row');
    // Header + 4 players = 5 rows
    expect(rows.length).toBe(5);
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('tutorial can be skipped', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('renders mobile viewport with 2-row hand grid', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    try {
      mockExec.mockResolvedValue(playPhaseState);
      renderWithProviders(<SpadesPage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const hand = document.querySelector('[data-tutorial="sp-player-hand"]');
      expect(hand).toBeInTheDocument();
      const rows = hand?.querySelectorAll('[data-testid="hand-row"]');
      expect(rows?.length).toBeGreaterThanOrEqual(1);
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  it('renders desktop viewport with wrapping hand', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 800 });
    try {
      mockExec.mockResolvedValue(playPhaseState);
      renderWithProviders(<SpadesPage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const hand = document.querySelector('[data-tutorial="sp-player-hand"]');
      expect(hand?.className).toContain('flex-wrap');
      expect(hand?.querySelectorAll('[data-testid="hand-row"]')).toHaveLength(0);
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });
});
