import { screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { badugiApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BadugiResponse } from '../types/card';
import { BadugiPhase } from '../types/phases';
import { BadugiPage } from './BadugiPage';

vi.mock('../api/gameApi', () => ({
  badugiApi: { exec: vi.fn() },
  actionLogApi: { badugi: vi.fn() },
}));

const mockExec = vi.mocked(badugiApi.exec);

const humanPlayer = (overrides: Partial<import('../types/card').BadugiPlayerData> = {}) => ({
  id: 0,
  isHuman: true,
  cards: [
    { design: 'SPADE' as const, value: 1 },
    { design: 'HEART' as const, value: 2 },
    { design: 'DIAMOND' as const, value: 3 },
    { design: 'CLOVER' as const, value: 4 },
  ],
  chips: 990,
  currentBet: 0,
  folded: false,
  allIn: false,
  handSize: 0,
  handName: '',
  drawCount: 0,
  totalDraws: 0,
  playStyleName: '',
  ...overrides,
});

const cpuPlayer = (id: number): import('../types/card').BadugiPlayerData => ({
  id,
  isHuman: false,
  cards: [],
  chips: 980,
  currentBet: 0,
  folded: false,
  allIn: false,
  handSize: 0,
  handName: '',
  drawCount: 0,
  totalDraws: 0,
  playStyleName: `CPU ${id}`,
});

const baseState = (overrides: Partial<BadugiResponse> = {}): BadugiResponse => ({
  players: [humanPlayer(), cpuPlayer(1), cpuPlayer(2), cpuPlayer(3)],
  pot: 40,
  sidePots: [],
  dealerIdx: 0,
  currentTurn: 0,
  phase: BadugiPhase.DEAL,
  drawIndex: 0,
  gameEndFlag: false,
  lastBet: 0,
  minRaise: 10,
  ante: 10,
  bettingLimit: 0,
  raiseCount: 0,
  maxBetAmount: 0,
  roundResults: [],
  cpuActions: [],
  cpuExchanges: [],
  message: '',
  ...overrides,
});

describe('BadugiPage', () => {
  beforeEach(() => {
    mockExec.mockReset();
  });

  it('calls reset on mount', async () => {
    mockExec.mockResolvedValue(baseState());
    renderWithProviders(<BadugiPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders the pot and dealer label', async () => {
    mockExec.mockResolvedValue(baseState({ pot: 120, dealerIdx: 2 }));
    renderWithProviders(<BadugiPage />);
    await waitFor(() => expect(screen.getByText('120')).toBeInTheDocument());
    expect(screen.getByText(/Player 2/)).toBeInTheDocument();
  });

  it('renders the pre-draw badge on the initial deal', async () => {
    mockExec.mockResolvedValue(baseState({ phase: BadugiPhase.DEAL, drawIndex: 0 }));
    renderWithProviders(<BadugiPage />);
    // Japanese default locale renders "プリドロー" for the pre-draw badge.
    await waitFor(() => expect(screen.getByText('プリドロー')).toBeInTheDocument());
  });

  it('renders the draw counter badge during draw phases', async () => {
    mockExec.mockResolvedValue(baseState({ phase: BadugiPhase.DRAW, drawIndex: 2 }));
    renderWithProviders(<BadugiPage />);
    // The counter appears both in the phase name and in the info badge; just
    // assert presence of at least one occurrence.
    await waitFor(() => expect(screen.getAllByText('ドロー 2/3').length).toBeGreaterThan(0));
  });

  it('shows the end message at showdown', async () => {
    mockExec.mockResolvedValue(
      baseState({
        phase: BadugiPhase.END,
        gameEndFlag: true,
        message: 'あなたの勝ちです。',
        messageCode: 'badugi.result.win',
        roundResults: [{ playerIdx: 0, handSize: 4, handName: 'Badugi', wonAmount: 40 }],
      }),
    );
    renderWithProviders(<BadugiPage />);
    await waitFor(() => expect(screen.getByText('あなたの勝ちです。')).toBeInTheDocument());
  });
});
