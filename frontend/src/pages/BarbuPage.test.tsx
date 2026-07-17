import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { barbuApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BarbuResponse, Card } from '../types/card';
import { BarbuPage } from './BarbuPage';

vi.mock('../api/gameApi', () => ({
  barbuApi: { exec: vi.fn() },
  actionLogApi: { barbu: vi.fn() },
}));

const mockExec = vi.mocked(barbuApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<BarbuResponse> = {}): BarbuResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 2,
        cards: [card('SPADE', 3), card('HEART', 5)],
        trickCount: 0,
        dominoRank: 0,
        totalScore: 0,
      },
      { id: 1, isHuman: false, cardCount: 2, cards: [], trickCount: 0, dominoRank: 0, totalScore: 0 },
      { id: 2, isHuman: false, cardCount: 2, cards: [], trickCount: 0, dominoRank: 0, totalScore: 0 },
      { id: 3, isHuman: false, cardCount: 2, cards: [], trickCount: 0, dominoRank: 0, totalScore: 0 },
    ],
    phase: 'selectContract',
    dealNumber: 0,
    totalDeals: 28,
    dealerIdx: 0,
    currentTurn: 0,
    currentContract: -1,
    trumpSuit: -1,
    trickNumber: 0,
    currentTrick: [],
    lastTrick: [],
    lastTrickWinner: -1,
    tablePlaced: [0, 0, 0, 0, 0],
    dominoPlayable: [],
    usedContracts: [false, false, false, false, false, false, false],
    gameEndFlag: false,
    config: { cpuDifficulty: 1 },
    roundWinners: [],
    lastDealDetail: null,
    message: '',
    ...overrides,
  };
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('BarbuPage', () => {
  it('calls reset on mount with the short "r" command', async () => {
    renderWithProviders(<BarbuPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('r'));
  });

  it('renders CPU difficulty options with localized labels', async () => {
    renderWithProviders(<BarbuPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('r'));
    // Difficulty options are localized (ja), not the hardcoded Easy/Normal/Hard.
    expect(screen.getByRole('option', { name: 'かんたん' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'ふつう' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'むずかしい' })).toBeInTheDocument();
  });

  it('renders contract buttons in the select phase', async () => {
    renderWithProviders(<BarbuPage />);
    await waitFor(() => expect(screen.getByTestId('contract-0')).toBeInTheDocument());
    expect(screen.getByTestId('contract-6')).toBeInTheDocument();
  });

  it('shows a contract description panel that updates on focus', async () => {
    renderWithProviders(<BarbuPage />);
    const panel = await screen.findByTestId('contract-desc');
    // Defaults to the first available contract's description.
    const initial = panel.textContent;
    expect(initial).toBeTruthy();
    // Focusing a different contract updates the panel...
    fireEvent.focus(screen.getByTestId('contract-3'));
    expect(panel.textContent).not.toBe(initial);
    // ...and blurring it restores the original description.
    fireEvent.blur(screen.getByTestId('contract-3'));
    expect(panel.textContent).toBe(initial);
  });

  it('selects a non-trump contract directly', async () => {
    renderWithProviders(<BarbuPage />);
    await waitFor(() => expect(screen.getByTestId('contract-0')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('contract-0'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('c', { contract: 0, trumpSuit: -1 }));
  });

  it('opens the trump picker for the Trumps contract', async () => {
    renderWithProviders(<BarbuPage />);
    await waitFor(() => expect(screen.getByTestId('contract-5')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('contract-5'));
    await waitFor(() => expect(screen.getByTestId('trump-1')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('trump-3'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('c', { contract: 5, trumpSuit: 3 }));
  });

  it('disables used contracts', async () => {
    mockExec.mockResolvedValue(makeState({ usedContracts: [true, false, false, false, false, false, false] }));
    renderWithProviders(<BarbuPage />);
    await waitFor(() => expect(screen.getByTestId('contract-0')).toBeDisabled());
  });

  it('plays a trick card after selecting it', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 'play', currentContract: 0, currentTurn: 0 }));
    renderWithProviders(<BarbuPage />);
    await waitFor(() => expect(screen.getByTestId('play-button')).toBeInTheDocument());
    expect(screen.getByTestId('play-button')).toBeDisabled();

    fireEvent.click(screen.getByTestId('hand-card-0'));
    await waitFor(() => expect(screen.getByTestId('play-button')).not.toBeDisabled());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('play-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('p', { handIndex: 0 }));
  });

  it('labels each trick card with the player who played it and marks the lead', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 'play',
        currentContract: 0,
        currentTurn: 2,
        currentTrick: [
          { playerIdx: 1, card: card('HEART', 10) },
          { playerIdx: 0, card: card('HEART', 12) },
        ],
      }),
    );
    renderWithProviders(<BarbuPage />);
    const trickCards = await screen.findAllByTestId('bb-trick-card');
    expect(trickCards).toHaveLength(2);
    // Lead card (index 0) was played by CPU 1 and carries the lead marker (▸).
    expect(trickCards[0]).toHaveTextContent('CPU 1');
    expect(trickCards[0]).toHaveTextContent('▸');
    // Follow card was played by the human; no lead marker on it.
    expect(trickCards[1]).toHaveTextContent('あなた');
    expect(trickCards[1]).not.toHaveTextContent('▸');
  });

  it('shows a pass button in Dominoes and passes when no card is playable', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 'play', currentContract: 6, currentTurn: 0, dominoPlayable: [] }));
    renderWithProviders(<BarbuPage />);
    await waitFor(() => expect(screen.getByTestId('pass-button')).toBeInTheDocument());
    expect(screen.getByTestId('pass-button')).not.toBeDisabled();
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('pass-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('p', { handIndex: -1 }));
  });

  it('shows the next-deal button at deal end', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 'dealEnd', currentContract: 0 }));
    renderWithProviders(<BarbuPage />);
    await waitFor(() => expect(screen.getByTestId('next-deal-button')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('next-deal-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('n'));
  });

  it('shows loading state when there are no players', async () => {
    mockExec.mockResolvedValue(makeState({ players: [] }));
    renderWithProviders(<BarbuPage />);
    await waitFor(() => expect(screen.queryByTestId('contract-0')).not.toBeInTheDocument());
  });

  it('renders CLI terminal when CLI mode is enabled via localStorage', async () => {
    localStorage.setItem('cli-mode-barbu', 'true');
    renderWithProviders(<BarbuPage />);
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
    localStorage.removeItem('cli-mode-barbu');
  });
});
