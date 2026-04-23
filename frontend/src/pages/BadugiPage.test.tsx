import { fireEvent, screen, waitFor } from '@testing-library/react';
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

  it('wires betting buttons during the human turn', async () => {
    mockExec.mockResolvedValue(baseState({ phase: BadugiPhase.DEAL, currentTurn: 0 }));
    renderWithProviders(<BadugiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /チェック/ })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /チェック/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('check', undefined, undefined, undefined, 0));
  });

  it('wires fold and allin', async () => {
    mockExec.mockResolvedValue(baseState({ phase: BadugiPhase.BET, currentTurn: 0, lastBet: 20 }));
    renderWithProviders(<BadugiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /フォールド/ })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /フォールド/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('fold', undefined, undefined, undefined, 0));

    fireEvent.click(screen.getByRole('button', { name: /オールイン/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('allin', undefined, undefined, undefined, 0));
  });

  it('exposes exchange and stand during draw phase', async () => {
    mockExec.mockResolvedValue(baseState({ phase: BadugiPhase.DRAW, drawIndex: 1, currentTurn: 0 }));
    renderWithProviders(<BadugiPage />);

    await waitFor(() => expect(screen.getByRole('button', { name: '交換' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: '交換' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('exchange', []));

    fireEvent.click(screen.getByRole('button', { name: /スタンド/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('stand'));
  });

  it('hides betting controls when the human has folded', async () => {
    mockExec.mockResolvedValue(
      baseState({
        phase: BadugiPhase.BET,
        players: [humanPlayer({ folded: true }), cpuPlayer(1), cpuPlayer(2), cpuPlayer(3)],
      }),
    );
    renderWithProviders(<BadugiPage />);
    await waitFor(() => expect(screen.getByText(/フォールド/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: /チェック/ })).toBeNull();
  });

  it('shows next game button at end phase', async () => {
    mockExec.mockResolvedValue(baseState({ phase: BadugiPhase.END, gameEndFlag: true }));
    renderWithProviders(<BadugiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /次のゲーム/ })).toBeInTheDocument());
  });

  it('toggles betting limit and reissues reset with the new config', async () => {
    mockExec.mockResolvedValue(baseState({ phase: BadugiPhase.END, gameEndFlag: true }));
    renderWithProviders(<BadugiPage />);
    // SettingsPanel wraps its form in <details>; open it via the summary.
    const summary = await screen.findByText('設定');
    fireEvent.click(summary);

    const limitSelect = await screen.findByLabelText(/リミット/);
    fireEvent.change(limitSelect, { target: { value: '2' } });
    fireEvent.click(screen.getByRole('button', { name: /次のゲーム/ }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenLastCalledWith('reset', undefined, undefined, { bettingLimit: 2, cpuMetaAI: false }),
    );
  });

  it('marks a selected card as pressed in draw phase', async () => {
    mockExec.mockResolvedValue(baseState({ phase: BadugiPhase.DRAW, drawIndex: 1, currentTurn: 0 }));
    renderWithProviders(<BadugiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '交換' })).toBeInTheDocument());

    // Card buttons come first in the footer; pick the first non-action button
    // by role=button and aria-pressed attribute presence.
    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    expect(cardButtons.length).toBeGreaterThan(0);
    fireEvent.click(cardButtons[0]);
    await waitFor(() => expect(cardButtons[0]).toHaveAttribute('aria-pressed', 'true'));
  });
});
