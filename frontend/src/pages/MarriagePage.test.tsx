import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi } from '../api/gameApi';
import { marriageApi } from '../api/games/marriage';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { MarriagePlayer, MarriageResponse } from '../types/games/marriage';
import { MarriagePage } from './MarriagePage';

vi.mock('../api/games/marriage', () => ({
  marriageApi: { exec: vi.fn() },
}));

vi.mock('../api/gameApi', async (importOriginal) => ({
  ...(await importOriginal<typeof import('../api/gameApi')>()),
  actionLogApi: { marriage: vi.fn() },
}));

const mockExec = vi.mocked(marriageApi.exec);

function player(overrides: Partial<MarriagePlayer> = {}): MarriagePlayer {
  return {
    id: 0,
    isHuman: true,
    cardCount: 21,
    cards: [],
    roundScore: 0,
    cumulativeScore: 0,
    deadwood: 0,
    hasPureSequence: false,
    maal: 0,
    ...overrides,
  };
}

const drawPhaseState: MarriageResponse = {
  players: [
    player({
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
      ],
    }),
    player({ id: 1, isHuman: false, roundScore: 3, cumulativeScore: 10 }),
  ],
  phase: 0,
  roundNumber: 1,
  targetRounds: 3,
  currentPlayerIdx: 0,
  dealerIdx: 0,
  discardTop: { design: 'HEART', value: 7 },
  drawPileCount: 40,
  wildJoker: { design: 'CLOVER', value: 5 },
  wildRank: 5,
  gameEndFlag: false,
  winnerIdx: -1,
  declarerIdx: -1,
  declarationValid: false,
  humanDeadwood: 0,
  humanHasPureSequence: false,
  message: '',
  config: { playerCount: 5, cpuDifficulty: 1, targetRounds: 3 },
};

const discardPhaseState: MarriageResponse = { ...drawPhaseState, phase: 1 };

// 21-card arrangement that IS a valid declaration (three pure runs + four sets),
// verified against the Go domain. The 22nd card (♣2) is the finish/discard card.
const declareValidHand: MarriagePlayer['cards'] = [
  { design: 'SPADE', value: 3 },
  { design: 'SPADE', value: 4 },
  { design: 'SPADE', value: 5 },
  { design: 'HEART', value: 7 },
  { design: 'HEART', value: 8 },
  { design: 'HEART', value: 9 },
  { design: 'DIAMOND', value: 11 },
  { design: 'DIAMOND', value: 12 },
  { design: 'DIAMOND', value: 13 },
  { design: 'DIAMOND', value: 2 },
  { design: 'HEART', value: 2 },
  { design: 'SPADE', value: 2 },
  { design: 'DIAMOND', value: 6 },
  { design: 'CLOVER', value: 6 },
  { design: 'SPADE', value: 6 },
  { design: 'DIAMOND', value: 10 },
  { design: 'CLOVER', value: 10 },
  { design: 'SPADE', value: 10 },
  { design: 'CLOVER', value: 13 },
  { design: 'HEART', value: 13 },
  { design: 'SPADE', value: 13 },
  { design: 'CLOVER', value: 2 }, // finish card (index 21)
];

const declareValidState: MarriageResponse = {
  ...discardPhaseState,
  wildJoker: null,
  wildRank: 0,
  players: [player({ cards: declareValidHand }), player({ id: 1, isHuman: false })],
};

// 21-card arrangement that is INVALID: only sets, no sequence at all (no pure
// sequence). The 22nd card (♥8) is the finish card.
const declareNoPureHand: MarriagePlayer['cards'] = [
  { design: 'DIAMOND', value: 2 },
  { design: 'CLOVER', value: 2 },
  { design: 'SPADE', value: 2 },
  { design: 'DIAMOND', value: 4 },
  { design: 'CLOVER', value: 4 },
  { design: 'SPADE', value: 4 },
  { design: 'DIAMOND', value: 6 },
  { design: 'CLOVER', value: 6 },
  { design: 'SPADE', value: 6 },
  { design: 'DIAMOND', value: 8 },
  { design: 'CLOVER', value: 8 },
  { design: 'SPADE', value: 8 },
  { design: 'DIAMOND', value: 10 },
  { design: 'CLOVER', value: 10 },
  { design: 'SPADE', value: 10 },
  { design: 'DIAMOND', value: 12 },
  { design: 'CLOVER', value: 12 },
  { design: 'SPADE', value: 12 },
  { design: 'DIAMOND', value: 13 },
  { design: 'CLOVER', value: 13 },
  { design: 'SPADE', value: 13 },
  { design: 'HEART', value: 8 }, // finish card (index 21)
];

const declareNoPureState: MarriageResponse = {
  ...discardPhaseState,
  wildJoker: null,
  wildRank: 0,
  players: [player({ cards: declareNoPureHand }), player({ id: 1, isHuman: false })],
};

// 21-card arrangement that is INVALID: one pure run but 18 cards cannot all be
// melded (115 deadwood points). The 22nd card (♣9) is the finish card.
const declareUncoveredHand: MarriagePlayer['cards'] = [
  { design: 'SPADE', value: 3 },
  { design: 'SPADE', value: 4 },
  { design: 'SPADE', value: 5 },
  { design: 'HEART', value: 1 },
  { design: 'HEART', value: 1 },
  { design: 'HEART', value: 2 },
  { design: 'HEART', value: 4 },
  { design: 'HEART', value: 5 },
  { design: 'HEART', value: 7 },
  { design: 'HEART', value: 7 },
  { design: 'HEART', value: 10 },
  { design: 'HEART', value: 1 },
  { design: 'DIAMOND', value: 1 },
  { design: 'DIAMOND', value: 1 },
  { design: 'DIAMOND', value: 2 },
  { design: 'DIAMOND', value: 4 },
  { design: 'DIAMOND', value: 5 },
  { design: 'DIAMOND', value: 7 },
  { design: 'DIAMOND', value: 7 },
  { design: 'DIAMOND', value: 10 },
  { design: 'DIAMOND', value: 1 },
  { design: 'CLOVER', value: 9 }, // finish card (index 21)
];

const declareUncoveredState: MarriageResponse = {
  ...discardPhaseState,
  wildJoker: null,
  wildRank: 0,
  players: [player({ cards: declareUncoveredHand }), player({ id: 1, isHuman: false })],
};
const roundEndState: MarriageResponse = { ...drawPhaseState, phase: 2 };
const gameEndState: MarriageResponse = {
  ...drawPhaseState,
  phase: 3,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};
const gameEndByFlagState: MarriageResponse = {
  ...drawPhaseState,
  phase: 0,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};
const cpuTurnState: MarriageResponse = { ...drawPhaseState, currentPlayerIdx: 1 };
const noDiscardState: MarriageResponse = { ...drawPhaseState, discardTop: null };
const roundEndCpuCardsState: MarriageResponse = {
  ...drawPhaseState,
  phase: 2,
  players: [
    drawPhaseState.players[0],
    player({
      id: 1,
      isHuman: false,
      cardCount: 3,
      cards: [
        { design: 'DIAMOND', value: 5 },
        { design: 'CLOVER', value: 6 },
        { design: 'HEART', value: 8 },
      ],
      roundScore: 3,
      cumulativeScore: 10,
      deadwood: 19,
      hasPureSequence: true,
    }),
  ],
};

beforeEach(() => {
  mockExec.mockResolvedValue(drawPhaseState);
});

describe('MarriagePage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<MarriagePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with default config', async () => {
    renderWithProviders(<MarriagePage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        playerCount: 5,
        cpuDifficulty: 1,
        targetRounds: 3,
      }),
    );
  });

  it('renders draw phase with human cards', async () => {
    renderWithProviders(<MarriagePage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
      expect(screen.getByAltText('♥ J')).toBeInTheDocument();
    });
  });

  it('marks wild-rank cards in the human hand with a WILD badge', async () => {
    const wildInHandState: MarriageResponse = {
      ...drawPhaseState,
      players: [
        player({
          cards: [
            { design: 'SPADE', value: 1 },
            { design: 'DIAMOND', value: 5 }, // matches wildJoker rank 5 -> wild
            { design: 'JOKER', value: 0 }, // printed joker -> wild
          ],
        }),
        drawPhaseState.players[1],
      ],
    };
    mockExec.mockResolvedValue(wildInHandState);
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    // Two of the three hand cards are wild (the rank-5 card and the printed joker).
    expect(screen.getAllByTestId('marriage-wild-badge')).toHaveLength(2);
    // The non-wild ace carries no wild annotation in its label.
    expect(screen.getByRole('button', { name: '♠ A' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /♦ 5 .*ワイルド/ })).toBeInTheDocument();
  });

  it('marks wild-rank cards in revealed CPU hands with a WILD badge', async () => {
    mockExec.mockResolvedValue(roundEndCpuCardsState);
    renderWithProviders(<MarriagePage />);
    // CPU hand has a rank-5 card (DIAMOND 5) matching the wild joker.
    await waitFor(() => expect(screen.getAllByTestId('marriage-wild-badge').length).toBeGreaterThanOrEqual(1));
  });

  it('renders draw stock and draw discard buttons on human draw turn', async () => {
    renderWithProviders(<MarriagePage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '捨て札から引く' })).toBeInTheDocument();
    });
  });

  it('draw discard button disabled when no discard top', async () => {
    mockExec.mockResolvedValue(noDiscardState);
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札から引く' })).toBeDisabled());
  });

  it('calls drawstock when draw stock button clicked', async () => {
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('calls drawdiscard when draw discard button clicked', async () => {
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札から引く' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '捨て札から引く' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawdiscard'));
  });

  it('renders discard and declare buttons on human discard turn', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<MarriagePage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '捨てる' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '宣言' })).toBeInTheDocument();
    });
  });

  it('discard and declare disabled when not exactly 1 card selected', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨てる' })).toBeDisabled());
    expect(screen.getByRole('button', { name: '宣言' })).toBeDisabled();
  });

  it('discard button enabled when 1 card selected', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    expect(screen.getByRole('button', { name: '捨てる' })).not.toBeDisabled();
  });

  it('calls discard command when discard button clicked', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '捨てる' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', 0));
  });

  it('calls declare command when declare button clicked', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(roundEndState);
    fireEvent.click(screen.getByRole('button', { name: '宣言' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declare', 0));
  });

  // -- Declaration preview (client-side) --
  it('does not show the declare preview until a finish card is selected', async () => {
    mockExec.mockResolvedValue(declareValidState);
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '宣言' })).toBeInTheDocument());
    expect(screen.queryByTestId('marriage-declare-preview')).not.toBeInTheDocument();
  });

  it('shows a valid declare preview when the remaining 21 cards form a valid declaration', async () => {
    mockExec.mockResolvedValue(declareValidState);
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByAltText('♣ 2')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♣ 2').closest('button') as HTMLButtonElement);
    const preview = await screen.findByTestId('marriage-declare-preview');
    expect(preview).toBeInTheDocument();
    expect(screen.getByTestId('marriage-declare-preview-valid')).toBeInTheDocument();
    expect(screen.queryByTestId('marriage-declare-preview-invalid')).not.toBeInTheDocument();
    // Declare button is never blocked by the preview.
    expect(screen.getByRole('button', { name: '宣言' })).not.toBeDisabled();
  });

  it('warns about a missing pure sequence and the penalty for an invalid declaration', async () => {
    mockExec.mockResolvedValue(declareNoPureState);
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByAltText('♥ 8')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♥ 8').closest('button') as HTMLButtonElement);
    await screen.findByTestId('marriage-declare-preview-invalid');
    expect(screen.getByText('純シーケンス未成立')).toBeInTheDocument();
    expect(screen.getByText('このまま宣言すると +80 点')).toBeInTheDocument();
    // Player can still force the declaration through.
    expect(screen.getByRole('button', { name: '宣言' })).not.toBeDisabled();
  });

  it('warns about unmelded cards when a pure sequence exists but cards are uncovered', async () => {
    mockExec.mockResolvedValue(declareUncoveredState);
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByAltText('♣ 9')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♣ 9').closest('button') as HTMLButtonElement);
    await screen.findByTestId('marriage-declare-preview-invalid');
    expect(screen.queryByText('純シーケンス未成立')).not.toBeInTheDocument();
    expect(screen.getByText(/未メルド .*枚（115 点）/)).toBeInTheDocument();
  });

  it('does not show draw buttons when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '山札から引く' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '捨て札から引く' })).not.toBeInTheDocument();
  });

  it('shows next round button on round end and calls nextround', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のラウンド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('shows game end with action log button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<MarriagePage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('shows game end via gameEndFlag with non-3 phase', async () => {
    mockExec.mockResolvedValue(gameEndByFlagState);
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());
  });

  it('shows error alert', async () => {
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('shows CPU player area', async () => {
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByText(/CPU 1.*21枚/)).toBeInTheDocument());
  });

  it('score table shows all players', async () => {
    renderWithProviders(<MarriagePage />);
    await waitFor(() => {
      expect(screen.getByText('スコア')).toBeInTheDocument();
      expect(screen.getByText('あなた')).toBeInTheDocument();
      expect(screen.getByText('CPU 1')).toBeInTheDocument();
    });
  });

  it('shows maal only when it is positive', async () => {
    mockExec.mockResolvedValue(drawPhaseState);
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByTestId('marriage-maal')).not.toBeInTheDocument();

    mockExec.mockResolvedValue({
      ...drawPhaseState,
      phase: 1,
      players: [player({ maal: 4 }), player({ id: 1, isHuman: false, maal: 2 })],
    });
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(screen.getAllByTestId('marriage-maal')).toHaveLength(3));
  });

  it('score table headers have scope="col"', async () => {
    const { container } = renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByText('あなた')).toBeInTheDocument());
    for (const th of container.querySelectorAll('th')) {
      expect(th).toHaveAttribute('scope', 'col');
    }
  });

  it('shows wild joker indicator', async () => {
    renderWithProviders(<MarriagePage />);
    await waitFor(() => {
      expect(screen.getByTestId('marriage-wild-joker')).toBeInTheDocument();
      expect(screen.getByText('ワイルドジョーカー')).toBeInTheDocument();
    });
  });

  it('shows discard top card', async () => {
    renderWithProviders(<MarriagePage />);
    await waitFor(() => {
      expect(screen.getByText('捨て札')).toBeInTheDocument();
      expect(screen.getByAltText('♥ 7')).toBeInTheDocument();
    });
  });

  it('does not show discard top when null', async () => {
    mockExec.mockResolvedValue(noDiscardState);
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByText('捨て札')).not.toBeInTheDocument();
  });

  it('reveals CPU cards, deadwood, and pure-sequence badge on round end', async () => {
    mockExec.mockResolvedValue(roundEndCpuCardsState);
    renderWithProviders(<MarriagePage />);
    await waitFor(() => {
      expect(screen.getByAltText('♦ 5')).toBeInTheDocument();
      expect(screen.getByText(/デッドウッド 19/)).toBeInTheDocument();
      expect(screen.getByText('純シーケンス有')).toBeInTheDocument();
    });
  });

  it('does not reveal CPU cards during draw phase', async () => {
    const drawWithCpuCards: MarriageResponse = {
      ...drawPhaseState,
      players: [
        drawPhaseState.players[0],
        player({ id: 1, isHuman: false, cardCount: 3, cards: [{ design: 'DIAMOND', value: 5 }] }),
      ],
    };
    mockExec.mockResolvedValue(drawWithCpuCards);
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByAltText('♦ 5')).not.toBeInTheDocument();
  });

  it('card selection toggles aria-pressed', async () => {
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');
    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('card buttons have aria-label with card name', async () => {
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    expect(screen.getByAltText('♠ A').closest('button')).toHaveAttribute('aria-label', '♠ A');
  });

  it('reset button calls exec with confirm', async () => {
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        playerCount: 5,
        cpuDifficulty: 1,
        targetRounds: 3,
      }),
    );
  });

  it('shows and dismisses confirm dialog', async () => {
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('settings panel changes playerCount', async () => {
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    fireEvent.click(screen.getByText('設定'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[0], { target: { value: '3' } });
    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        playerCount: 3,
        cpuDifficulty: 1,
        targetRounds: 3,
      }),
    );
  });

  it('settings panel changes targetRounds', async () => {
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    fireEvent.click(screen.getByText('設定'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[2], { target: { value: '5' } });
    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        playerCount: 5,
        cpuDifficulty: 1,
        targetRounds: 5,
      }),
    );
  });

  it('round info displayed', async () => {
    renderWithProviders(<MarriagePage />);
    await waitFor(() => {
      expect(screen.getByText('ラウンド 1/3')).toBeInTheDocument();
      expect(screen.getByText('山札: 40枚')).toBeInTheDocument();
    });
  });

  it('shows loading state and disables reset while loading', async () => {
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
    let resolve!: (value: MarriageResponse) => void;
    const slow = new Promise<MarriageResponse>((r) => {
      resolve = r;
    });
    mockExec.mockReturnValueOnce(slow);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    expect(screen.getByRole('button', { name: 'リセット' })).toBeDisabled();
    resolve(drawPhaseState);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
  });

  it('sets aria-busy on container', async () => {
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
    const container = screen.getByRole('button', { name: 'リセット' }).closest('[aria-busy]') as HTMLElement;
    expect(container).toHaveAttribute('aria-busy', 'false');
  });

  // -- PhaseIndicator --
  it('phase indicator shows your turn on human draw turn', async () => {
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('あなたのターン'));
  });

  it('phase indicator shows waiting on cpu turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('待機中'));
  });

  // -- Keyboard navigation --
  it('number key toggles a card in discard phase', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('Enter key triggers discard in discard phase', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    fireEvent.keyDown(document, { key: '1' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.keyDown(document, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', 0));
  });

  it('Escape key clears selection', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');
    fireEvent.keyDown(document, { key: 'Escape' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('keyboard nav disabled in draw phase', async () => {
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('keyboard nav disabled when not human turn', async () => {
    mockExec.mockResolvedValue({ ...discardPhaseState, currentPlayerIdx: 1 });
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByText('CPU 1')).toBeInTheDocument());
    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  // -- Action log --
  it('handles action log visibility and API fetch', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());
    vi.mocked(actionLogApi.marriage).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));
    await waitFor(() => expect(actionLogApi.marriage).toHaveBeenCalledTimes(1));
    expect(screen.getByText('棋譜')).toBeInTheDocument();
    fireEvent.click(screen.getByText('閉じる'));
    await waitFor(() => expect(screen.queryByText(/^棋譜$/)).not.toBeInTheDocument());
  });

  it('does not show action log button when not game end', async () => {
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByText('棋譜を見る')).not.toBeInTheDocument();
  });

  // -- Tutorial --
  it('renders tutorial button and starts/skips tutorial', async () => {
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  // **CUI は DISCARD の手番で毎ターン出している。**Web は「1 枚選んで残り 21 枚が
  // 成立するとき」しか出しておらず、思案中に確認できなかった (#4824)。
  it('always shows the deadwood and pure-sequence status during the discard phase', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<MarriagePage />);

    const status = await screen.findByTestId('marriage-hand-status');
    expect(status).toHaveTextContent('デッドウッド');
    // カードを 1 枚も選んでいない状態で出ている。
    expect(screen.queryByTestId('marriage-declare-preview')).not.toBeInTheDocument();
  });

  it('does not show the hand status outside the discard phase', async () => {
    mockExec.mockResolvedValue(drawPhaseState);
    renderWithProviders(<MarriagePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('marriage-hand-status')).not.toBeInTheDocument();
  });
});

// #5501: 表示されるのは合計の未メルド点数だけで、なぜその数字になるのかを
// 個々のカードから逆算する手掛かりが無かった。
describe('MarriagePage points legend', () => {
  it('shows the card-point legend next to the deadwood readout', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<MarriagePage />);
    const legend = await screen.findByTestId('marriage-points-legend');
    // **A が 10 点であることが読み取れること。** ここが標準のジンラミー系と違う。
    expect(legend.textContent).toMatch(/A/);
    expect(legend.textContent).toMatch(/10/);
    // ワイルドが 0 点であることも同じ行で言う (tutorial.wildJoker と矛盾しない)。
    expect(legend.textContent).toMatch(/0/);
  });
});
