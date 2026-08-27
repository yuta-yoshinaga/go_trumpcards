import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, napoleonApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { NapoleonResponse } from '../types/card';
import { NapoleonPage } from './NapoleonPage';

vi.mock('../api/gameApi', () => ({
  napoleonApi: { exec: vi.fn() },
  actionLogApi: { napoleon: vi.fn() },
}));

const mockExec = vi.mocked(napoleonApi.exec);

const playPhaseState: NapoleonResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 10,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
      ],
      bid: 13,
      isNapoleon: true,
      isAdjutant: false,
      adjutantRevealed: false,
      pictureCards: 2,
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 10,
      cards: [],
      bid: 0,
      isNapoleon: false,
      isAdjutant: true,
      adjutantRevealed: false,
      pictureCards: 1,
      roundScore: 3,
      cumulativeScore: 10,
      trickCount: 1,
    },
    {
      id: 2,
      isHuman: false,
      cardCount: 10,
      cards: [],
      bid: 0,
      isNapoleon: false,
      isAdjutant: false,
      adjutantRevealed: false,
      pictureCards: 0,
      roundScore: 5,
      cumulativeScore: 20,
      trickCount: 2,
    },
    {
      id: 3,
      isHuman: false,
      cardCount: 10,
      cards: [],
      bid: 0,
      isNapoleon: false,
      isAdjutant: false,
      adjutantRevealed: false,
      pictureCards: 0,
      roundScore: 0,
      cumulativeScore: 5,
      trickCount: 0,
    },
  ],
  phase: 3,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  bidPlayerIdx: 0,
  currentTrick: [],
  trumpSuit: 1,
  adjutantCard: { design: 'HEART', value: 1 },
  napoleonIdx: 0,
  adjutantIdx: 1,
  adjutantRevealed: false,
  highestBid: 13,
  highestBidder: 0,
  kitty: [],
  gameEndFlag: false,
  winnerTeam: -1,
  message: '',
  config: { cpuDifficulty: 1, minBid: 12, pointLimit: 100 },
};

const bidPhaseState: NapoleonResponse = {
  ...playPhaseState,
  phase: 0,
  bidPlayerIdx: 0,
  trumpSuit: 0,
  adjutantCard: null,
  highestBid: 0,
  players: playPhaseState.players.map((p) => ({ ...p, bid: -1, isNapoleon: false, isAdjutant: false })),
};

const bidPhaseCpuTurnState: NapoleonResponse = {
  ...bidPhaseState,
  bidPlayerIdx: 1,
};

const trumpDeclarationState: NapoleonResponse = {
  ...playPhaseState,
  phase: 1,
  napoleonIdx: 0,
};

const trumpDeclarationCpuState: NapoleonResponse = {
  ...playPhaseState,
  phase: 1,
  napoleonIdx: 1,
  players: playPhaseState.players.map((p, i) => ({
    ...p,
    isNapoleon: i === 1,
  })),
};

const kittyExchangeState: NapoleonResponse = {
  ...playPhaseState,
  phase: 2,
  napoleonIdx: 0,
  kitty: [
    { design: 'DIAMOND', value: 7 },
    { design: 'CLOVER', value: 3 },
  ],
};

const kittyExchangeCpuState: NapoleonResponse = {
  ...playPhaseState,
  phase: 2,
  napoleonIdx: 1,
  kitty: [],
  players: playPhaseState.players.map((p, i) => ({
    ...p,
    isNapoleon: i === 1,
  })),
};

const trickEndState: NapoleonResponse = {
  ...playPhaseState,
  phase: 4,
  currentTrick: [
    { playerIdx: 0, card: { design: 'DIAMOND', value: 3 } },
    { playerIdx: 1, card: { design: 'HEART', value: 5 } },
  ],
};

const roundEndState: NapoleonResponse = {
  ...playPhaseState,
  phase: 5,
};

const gameEndState: NapoleonResponse = {
  ...playPhaseState,
  phase: 6,
  gameEndFlag: true,
  winnerTeam: 0,
  message: 'Game end!',
};

const gameEndByFlagState: NapoleonResponse = {
  ...playPhaseState,
  phase: 3,
  gameEndFlag: true,
  winnerTeam: 0,
  message: 'Game end!',
};

const cpuTurnState: NapoleonResponse = {
  ...playPhaseState,
  currentPlayerIdx: 1,
};

const adjutantRevealedState: NapoleonResponse = {
  ...playPhaseState,
  adjutantRevealed: true,
};

const noHighestBidState: NapoleonResponse = {
  ...playPhaseState,
  highestBid: 0,
};

beforeEach(() => {
  mockExec.mockResolvedValue(playPhaseState);
});

describe('NapoleonPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<NapoleonPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, undefined, undefined, undefined, {
        cpuDifficulty: 1,
        minBid: 12,
        pointLimit: 100,
      }),
    );
  });

  it('renders play phase with human cards', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => {
      expect(screen.getByAltText('\u2660 A')).toBeInTheDocument();
      expect(screen.getByAltText('\u2665 J')).toBeInTheDocument();
    });
  });

  it('renders bid phase with bid button and input', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '\u30d3\u30c3\u30c9' })).toBeInTheDocument();
      expect(screen.getByLabelText('ビッド数入力')).toBeInTheDocument();
    });
  });

  it('shows bid phase instruction when human bid turn', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => {
      expect(
        screen.getByText(/\u30d3\u30c3\u30c9\u3092\u5ba3\u8a00\u3057\u3066\u304f\u3060\u3055\u3044/),
      ).toBeInTheDocument();
    });
  });

  it('does not show bid instruction when cpu bid turn', async () => {
    mockExec.mockResolvedValue(bidPhaseCpuTurnState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByText(/\u30d3\u30c3\u30c9\u3092\u5ba3\u8a00/)).not.toBeInTheDocument();
  });

  it('shows pass button during human bid turn', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '\u30d1\u30b9' })).toBeInTheDocument();
    });
  });

  it('calls bid command when bid button is clicked', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30d3\u30c3\u30c9' })).toBeInTheDocument());

    const input = screen.getByLabelText('ビッド数入力');
    fireEvent.change(input, { target: { value: '14' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30d3\u30c3\u30c9' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', 14));
  });

  it('calls pass (bid 0) when pass button is clicked', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30d1\u30b9' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30d1\u30b9' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', 0));
  });

  it('shows Napoleon-side face-card progress (adjutant excluded until revealed)', async () => {
    mockExec.mockResolvedValue(playPhaseState); // napoleon 2 pics, bid 13, adjutant hidden
    renderWithProviders(<NapoleonPage />);
    const badge = await screen.findByTestId('np-face-progress');
    expect(badge).toHaveTextContent('絵札 2/13 枚');
  });

  it('includes the adjutant haul in the progress once revealed', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, adjutantRevealed: true });
    renderWithProviders(<NapoleonPage />);
    const badge = await screen.findByTestId('np-face-progress');
    expect(badge).toHaveTextContent('絵札 3/13 枚');
  });

  it('shows trump declaration controls and the adjutant card picker when human is napoleon', async () => {
    mockExec.mockResolvedValue(trumpDeclarationState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => {
      expect(screen.getByLabelText('切り札スート')).toBeInTheDocument();
      expect(screen.getByTestId('np-adjutant-picker')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '宣言' })).toBeInTheDocument();
    });
    // 52 suit cards + 1 joker = 53 tappable adjutant options.
    expect(screen.getAllByTestId(/^np-adjutant-option-/)).toHaveLength(53);
  });

  it('does not show trump declaration controls when cpu is napoleon', async () => {
    mockExec.mockResolvedValue(trumpDeclarationCpuState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByLabelText('切り札スート')).not.toBeInTheDocument();
    expect(screen.queryByTestId('np-adjutant-picker')).not.toBeInTheDocument();
  });

  it('shows trump declaration instruction when human is napoleon', async () => {
    mockExec.mockResolvedValue(trumpDeclarationState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => {
      expect(screen.getByText('切り札スートと副官カードを選択してください')).toBeInTheDocument();
    });
  });

  it('disables the declare button until an adjutant card is picked', async () => {
    mockExec.mockResolvedValue(trumpDeclarationState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '宣言' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '宣言' })).toBeDisabled();

    fireEvent.click(screen.getByTestId('np-adjutant-option-1-1'));
    expect(screen.getByRole('button', { name: '宣言' })).toBeEnabled();
  });

  it('designates the tapped card as adjutant when declare is clicked', async () => {
    mockExec.mockResolvedValue(trumpDeclarationState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByTestId('np-adjutant-picker')).toBeInTheDocument());

    // Tap Heart 13 (K): adjutant suit 3, value 13.
    fireEvent.click(screen.getByTestId('np-adjutant-option-3-13'));
    expect(screen.getByTestId('np-adjutant-selected')).toBeInTheDocument();

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '宣言' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('trump', undefined, 1, 3, 13));
  });

  it('submits suit 0 / value 1 when the joker card is picked', async () => {
    mockExec.mockResolvedValue(trumpDeclarationState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByTestId('np-adjutant-picker')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('np-adjutant-option-0-1'));

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '宣言' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('trump', undefined, 1, 0, 1));
  });

  it('dims cards the human already holds but still allows designating them', async () => {
    mockExec.mockResolvedValue(trumpDeclarationState);
    renderWithProviders(<NapoleonPage />);
    // Human holds SPADE A (suit 1, value 1) in trumpDeclarationState.
    const heldOption = await screen.findByTestId('np-adjutant-option-1-1');
    expect(heldOption).toHaveStyle({ opacity: '0.4' });

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(heldOption);
    fireEvent.click(screen.getByRole('button', { name: '宣言' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('trump', undefined, 1, 1, 1));
  });

  it('shows kitty exchange controls when human is napoleon', async () => {
    mockExec.mockResolvedValue(kittyExchangeState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => {
      expect(screen.getByText('\u5834\u672d')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '\u4ea4\u63db' })).toBeInTheDocument();
    });
  });

  it('does not show kitty exchange controls when cpu is napoleon', async () => {
    mockExec.mockResolvedValue(kittyExchangeCpuState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '\u4ea4\u63db' })).not.toBeInTheDocument();
  });

  it('shows kitty exchange instruction when human is napoleon', async () => {
    mockExec.mockResolvedValue(kittyExchangeState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => {
      expect(
        screen.getByText('\u624b\u672d\u304b\u30891\u679a\u6368\u3066\u3066\u304f\u3060\u3055\u3044'),
      ).toBeInTheDocument();
    });
  });

  it('exchange button disabled when no card selected', async () => {
    mockExec.mockResolvedValue(kittyExchangeState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u4ea4\u63db' })).toBeDisabled());
  });

  it('calls exchange when card selected and exchange clicked', async () => {
    mockExec.mockResolvedValue(kittyExchangeState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement);
    // With one card selected the exchange button shows the dynamic "\u2660 A \u3092\u6368\u3066\u308b" label.
    const exchangeBtn = screen.getByRole('button', { name: /\u3092\u6368\u3066\u308b/ });
    expect(exchangeBtn).toHaveTextContent('\u2660 A');
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(exchangeBtn);

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('exchange', undefined, undefined, undefined, undefined, 0),
    );
  });

  it('play button disabled when not 1 card selected', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u51fa\u3059' })).toBeDisabled());
  });

  it('play button enabled when 1 card selected', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement);
    expect(screen.getByRole('button', { name: '\u51fa\u3059' })).not.toBeDisabled();
  });

  it('does not show play button when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByText('CPU 1')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '\u51fa\u3059' })).not.toBeInTheDocument();
  });

  it('shows next trick button on trick end', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '\u6b21\u306e\u30c8\u30ea\u30c3\u30af' })).toBeInTheDocument(),
    );
  });

  it('shows next round button on round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '\u6b21\u306e\u30e9\u30a6\u30f3\u30c9' })).toBeInTheDocument(),
    );
  });

  it('shows game end with action log button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b')).toBeInTheDocument();
    });
  });

  it('shows game end via gameEndFlag with non-6 phase', async () => {
    mockExec.mockResolvedValue(gameEndByFlagState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b')).toBeInTheDocument();
    });
  });

  it('shows error alert', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('settings panel changes cpuDifficulty', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());

    fireEvent.click(screen.getByText('\u8a2d\u5b9a'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[0], { target: { value: '2' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, undefined, undefined, undefined, {
        cpuDifficulty: 2,
        pointLimit: 100,
        minBid: 12,
      }),
    );
  });

  it('settings panel changes pointLimit', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());

    fireEvent.click(screen.getByText('\u8a2d\u5b9a'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[1], { target: { value: '200' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, undefined, undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 200,
        minBid: 12,
      }),
    );
  });

  it('settings panel changes minBid', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());

    fireEvent.click(screen.getByText('\u8a2d\u5b9a'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[2], { target: { value: '14' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, undefined, undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 100,
        minBid: 14,
      }),
    );
  });

  it('card selection toggle via aria-pressed', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('card buttons have aria-label with card name', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-label', '\u2660 A');

    const cardBtn2 = screen.getByAltText('\u2665 J').closest('button') as HTMLButtonElement;
    expect(cardBtn2).toHaveAttribute('aria-label', '\u2665 J');
  });

  it('reset button calls exec', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, undefined, undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 100,
        minBid: 12,
      }),
    );
  });

  it('score table shows all players', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => {
      expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument();
      expect(screen.getByText('\u3042\u306a\u305f')).toBeInTheDocument();
      expect(screen.getByText('CPU 1')).toBeInTheDocument();
      expect(screen.getByText('CPU 2')).toBeInTheDocument();
      expect(screen.getByText('CPU 3')).toBeInTheDocument();
    });
  });

  it('score table headers have scope="col" for accessibility', async () => {
    const { container } = renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByText('\u3042\u306a\u305f')).toBeInTheDocument());
    const ths = container.querySelectorAll('th');
    ths.forEach((th) => {
      expect(th).toHaveAttribute('scope', 'col');
    });
  });

  it('score table has horizontal scroll wrapper', async () => {
    const { container } = renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    const scoreSection = container.querySelector('[data-tutorial="np-score-table"]');
    const scrollWrapper = scoreSection?.querySelector('.overflow-x-auto');
    expect(scrollWrapper).toBeInTheDocument();
    const table = scrollWrapper?.querySelector('table');
    expect(table?.className).toContain('min-w-[420px]');
  });

  it('score table renders ScrollFadeHint on mobile', async () => {
    const innerWidthSpy = vi.spyOn(window, 'innerWidth', 'get').mockReturnValue(375);
    try {
      const { container } = renderWithProviders(<NapoleonPage />);
      await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
      const scoreSection = container.querySelector('[data-tutorial="np-score-table"]');
      const fadeHint = scoreSection?.querySelector('.bg-gradient-to-l');
      expect(fadeHint).toBeInTheDocument();
    } finally {
      innerWidthSpy.mockRestore();
    }
  });

  it('shows trump suit info', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByText(/\u30b9\u30da\u30fc\u30c9/)).toBeInTheDocument());
  });

  it('does not show trump suit when trumpSuit is 0', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByText('\u5207\u308a\u672d:')).not.toBeInTheDocument();
  });

  it('shows highest bid info', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByText('\u6700\u9ad8\u30d3\u30c3\u30c9: 13')).toBeInTheDocument());
  });

  it('does not show highest bid when 0', async () => {
    mockExec.mockResolvedValue(noHighestBidState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByText(/\u6700\u9ad8\u30d3\u30c3\u30c9/)).not.toBeInTheDocument();
  });

  it('shows current trick cards', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => {
      expect(screen.getByText('\u73fe\u5728\u306e\u30c8\u30ea\u30c3\u30af')).toBeInTheDocument();
      expect(screen.getByAltText('\u2666 3')).toBeInTheDocument();
      expect(screen.getByAltText('\u2665 5')).toBeInTheDocument();
    });
  });

  it('does not show current trick when empty', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByText('\u73fe\u5728\u306e\u30c8\u30ea\u30c3\u30af')).not.toBeInTheDocument();
  });

  it('shows CPU player areas', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1.*10\u679a/)).toBeInTheDocument();
      expect(screen.getByText(/CPU 2.*10\u679a/)).toBeInTheDocument();
      expect(screen.getByText(/CPU 3.*10\u679a/)).toBeInTheDocument();
    });
  });

  it('shows loading state', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    let resolve!: (value: NapoleonResponse) => void;
    const slow = new Promise<NapoleonResponse>((r) => {
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
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u51fa\u3059' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('play', undefined, undefined, undefined, undefined, undefined, 0),
    );
  });

  it('calls next when next trick button is clicked', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<NapoleonPage />);
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
    renderWithProviders(<NapoleonPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '\u6b21\u306e\u30e9\u30a6\u30f3\u30c9' })).toBeInTheDocument(),
    );

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u6b21\u306e\u30e9\u30a6\u30f3\u30c9' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('round and trick info displayed', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => {
      expect(screen.getByText('\u30e9\u30a6\u30f3\u30c9 1')).toBeInTheDocument();
      expect(screen.getByText('\u30c8\u30ea\u30c3\u30af 1')).toBeInTheDocument();
    });
  });

  it('does not show message when empty', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByText('Game end!')).not.toBeInTheDocument();
  });

  it('shows confirm dialog when reset button is clicked', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: '\u30ad\u30e3\u30f3\u30bb\u30eb' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('executes reset on confirm', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, undefined, undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 100,
        minBid: 12,
      }),
    );
  });

  it('handles action log visibility and API fetch', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b')).toBeInTheDocument());

    vi.mocked(actionLogApi.napoleon).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b'));

    await waitFor(() => expect(actionLogApi.napoleon).toHaveBeenCalledTimes(1));
    expect(screen.getByText('\u68cb\u8b5c')).toBeInTheDocument();

    fireEvent.click(screen.getByText('\u9589\u3058\u308b'));
    await waitFor(() => expect(screen.queryByText(/^\u68cb\u8b5c$/)).not.toBeInTheDocument());
    expect(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b')).toBeInTheDocument();
  });

  it('does not show action log button when not game end', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByText('\u68cb\u8b5c\u3092\u898b\u308b')).not.toBeInTheDocument();
  });

  it('does not show bid controls in play phase', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByLabelText('ビッド数入力')).not.toBeInTheDocument();
  });

  it('disables buttons while loading', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    let resolve!: (value: NapoleonResponse) => void;
    const slow = new Promise<NapoleonResponse>((r) => {
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
    const stateWithBadIdx: NapoleonResponse = {
      ...trickEndState,
      currentTrick: [{ playerIdx: 99, card: { design: 'SPADE', value: 1 } }],
    };
    mockExec.mockResolvedValue(stateWithBadIdx);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByText('CPU 99')).toBeInTheDocument());
  });

  it('sets aria-busy on container', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    const container = screen
      .getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })
      .closest('[aria-busy]') as HTMLElement;
    expect(container).toHaveAttribute('aria-busy', 'false');
  });

  it('no human cards renders empty hand area', async () => {
    const noHuman: NapoleonResponse = {
      ...playPhaseState,
      players: playPhaseState.players.map((p) => ({ ...p, isHuman: false })),
    };
    mockExec.mockResolvedValue(noHuman);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByAltText('\u2660 A')).not.toBeInTheDocument();
  });

  it('isHumanTurn false when currentPlayerIdx points to cpu', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '\u51fa\u3059' })).not.toBeInTheDocument();
  });

  it('phase indicator shows your turn during bid phase', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() =>
      expect(screen.getByTestId('phase-indicator')).toHaveTextContent('\u3042\u306a\u305f\u306e\u30bf\u30fc\u30f3'),
    );
  });

  it('phase indicator shows your turn when human play turn', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() =>
      expect(screen.getByTestId('phase-indicator')).toHaveTextContent('\u3042\u306a\u305f\u306e\u30bf\u30fc\u30f3'),
    );
  });

  it('phase indicator shows waiting when cpu turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('\u5f85\u6a5f\u4e2d'));
  });

  it('phase indicator shows your turn during trump declaration', async () => {
    mockExec.mockResolvedValue(trumpDeclarationState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() =>
      expect(screen.getByTestId('phase-indicator')).toHaveTextContent('\u3042\u306a\u305f\u306e\u30bf\u30fc\u30f3'),
    );
  });

  it('phase indicator shows your turn during kitty exchange', async () => {
    mockExec.mockResolvedValue(kittyExchangeState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() =>
      expect(screen.getByTestId('phase-indicator')).toHaveTextContent('\u3042\u306a\u305f\u306e\u30bf\u30fc\u30f3'),
    );
  });

  it('pressing number key toggles card in play phase', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('Enter key triggers play in play phase', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.keyDown(document, { key: '1' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);

    fireEvent.keyDown(document, { key: 'Enter' });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('play', undefined, undefined, undefined, undefined, undefined, 0),
    );
  });

  it('Escape key clears selection', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.keyDown(document, { key: 'Escape' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('keyboard nav disabled when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByText('CPU 1')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('reset hides the action log panel if open', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b')).toBeInTheDocument());

    vi.mocked(actionLogApi.napoleon).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b'));
    await waitFor(() => expect(screen.getByText('\u68cb\u8b5c')).toBeInTheDocument());

    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u6b21\u306e\u30b2\u30fc\u30e0' }));

    await waitFor(() => expect(screen.queryByText('\u68cb\u8b5c')).not.toBeInTheDocument());
  });

  it('shows bid value for player with bid >= 0', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => {
      // CPU 1 has bid=0 (pass)
      expect(screen.getByText(/CPU 1.*\u30d3\u30c3\u30c9 0/)).toBeInTheDocument();
    });
  });

  it('shows unbid text for player with bid < 0', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1.*\u672a\u30d3\u30c3\u30c9/)).toBeInTheDocument();
    });
  });

  it('score table shows dash for bid < 0', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    const rows = screen.getAllByRole('row');
    // Header + 4 players = 5 rows
    expect(rows.length).toBe(5);
  });

  it('shows napoleon role badge in CPU area', async () => {
    const stateWithNapoleonCpu: NapoleonResponse = {
      ...playPhaseState,
      players: playPhaseState.players.map((p, i) => ({
        ...p,
        isNapoleon: i === 1,
        isAdjutant: false,
      })),
      napoleonIdx: 1,
    };
    mockExec.mockResolvedValue(stateWithNapoleonCpu);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1.*\u30ca\u30dd\u30ec\u30aa\u30f3/)).toBeInTheDocument();
    });
  });

  it('shows adjutant role badge when revealed', async () => {
    mockExec.mockResolvedValue(adjutantRevealedState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1.*\u526f\u5b98/)).toBeInTheDocument();
    });
  });

  it('does not show adjutant role badge when not revealed', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    // CPU 1 is adjutant but not revealed
    expect(screen.queryByText(/CPU 1.*\u526f\u5b98/)).not.toBeInTheDocument();
  });

  it('shows kitty cards during exchange phase', async () => {
    mockExec.mockResolvedValue(kittyExchangeState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => {
      expect(screen.getByText('\u5834\u672d')).toBeInTheDocument();
      expect(screen.getByAltText('\u2666 7')).toBeInTheDocument();
      expect(screen.getByAltText('\u2663 3')).toBeInTheDocument();
    });
  });

  it('does not show kitty during play phase', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByText('\u5834\u672d')).not.toBeInTheDocument();
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('tutorial can be skipped', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('renders mobile viewport with 2-row hand grid', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    try {
      mockExec.mockResolvedValue(playPhaseState);
      renderWithProviders(<NapoleonPage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const hand = document.querySelector('[data-tutorial="np-player-hand"]');
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
      renderWithProviders(<NapoleonPage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const hand = document.querySelector('[data-tutorial="np-player-hand"]');
      expect(hand?.className).toContain('flex-wrap');
      expect(hand?.querySelectorAll('[data-testid="hand-row"]')).toHaveLength(0);
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  it('renders CPU info as collapsible details on mobile', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    try {
      mockExec.mockResolvedValue(playPhaseState);
      const { container } = renderWithProviders(<NapoleonPage />);
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
      const { container } = renderWithProviders(<NapoleonPage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const scoreDetails = container.querySelector('details[data-tutorial="np-score-table"]');
      expect(scoreDetails).toBeInTheDocument();
      const summary = scoreDetails?.querySelector('summary');
      expect(summary).toHaveTextContent('スコア');
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });
});

// #5504: 目標点数は開始前の設定でしか見えず、対局中は Settings を開き直さない限り
// あと何点で決着するのか分からなかった。
describe('NapoleonPage point limit', () => {
  it('shows the configured target during play', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, config: { ...playPhaseState.config, pointLimit: 75 } });
    renderWithProviders(<NapoleonPage />);
    const line = await screen.findByTestId('np-point-limit');
    expect(line.textContent).toContain('75');
  });

  // **設定値を出していること。** 定数を書いているだけなら、変えても表示が動かない。
  it('follows a settings change', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, config: { ...playPhaseState.config, pointLimit: 30 } });
    renderWithProviders(<NapoleonPage />);
    const line = await screen.findByTestId('np-point-limit');
    expect(line.textContent).toContain('30');
    expect(line.textContent).not.toContain('75');
  });

  // 変わることが読み上げの条件なので、hint がある間だけ現れる内側の div ではなく、
  // 常設のラッパーがライブ領域でなければならない。
  it('exposes the hint through a region that was already mounted', async () => {
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    const region = await screen.findByTestId('np-hint-live');
    expect(region).toHaveAttribute('role', 'status');
    expect(region).toHaveAttribute('aria-live', 'polite');
    // Empty before a hint exists -- that emptiness is what makes the later
    // text a change the reader announces.
    expect(region).toHaveTextContent('');
  });

  it('announces each kind of hint the page can render', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<NapoleonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    const region = screen.getByTestId('np-hint-live');

    // A play hint: the branch the region exists to announce.
    mockExec.mockResolvedValue({
      ...playPhaseState,
      hint: { cardIndex: 1, reason: 'followSuit' },
    } as unknown as NapoleonResponse);
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(region).toHaveTextContent('推奨カード'));

    // A bid hint takes a different branch of the same nested ternary.
    mockExec.mockResolvedValue({
      ...bidPhaseState,
      hint: { bid: 14, reason: 'strongHand' },
    } as unknown as NapoleonResponse);
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(region).toHaveTextContent('推奨ビッド'));
  });
});
