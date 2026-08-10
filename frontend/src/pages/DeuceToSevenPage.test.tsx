import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { deuceToSevenApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { DeuceToSevenResponse } from '../types/card';
import { DeuceToSevenPhase } from '../types/phases';
import { DeuceToSevenPage } from './DeuceToSevenPage';

vi.mock('../api/gameApi', () => ({
  deuceToSevenApi: { exec: vi.fn() },
  actionLogApi: { deucetoseven: vi.fn() },
}));

const mockExec = vi.mocked(deuceToSevenApi.exec);

const humanPlayer = (overrides: Partial<import('../types/card').DeuceToSevenPlayerData> = {}) => ({
  id: 0,
  isHuman: true,
  // Made pat low 7-5-4-3-2 (no pair / straight / flush).
  cards: [
    { design: 'SPADE' as const, value: 7 },
    { design: 'HEART' as const, value: 5 },
    { design: 'DIAMOND' as const, value: 4 },
    { design: 'CLOVER' as const, value: 3 },
    { design: 'SPADE' as const, value: 2 },
  ],
  chips: 990,
  currentBet: 0,
  folded: false,
  allIn: false,
  handRank: 0,
  handName: '',
  drawCount: 0,
  totalDraws: 0,
  playStyleName: '',
  ...overrides,
});

const cpuPlayer = (id: number): import('../types/card').DeuceToSevenPlayerData => ({
  id,
  isHuman: false,
  cards: [],
  chips: 980,
  currentBet: 0,
  folded: false,
  allIn: false,
  handRank: 0,
  handName: '',
  drawCount: 0,
  totalDraws: 0,
  playStyleName: `CPU ${id}`,
});

const baseState = (overrides: Partial<DeuceToSevenResponse> = {}): DeuceToSevenResponse => ({
  players: [humanPlayer(), cpuPlayer(1), cpuPlayer(2), cpuPlayer(3)],
  pot: 40,
  sidePots: [],
  dealerIdx: 0,
  currentTurn: 0,
  phase: DeuceToSevenPhase.DEAL,
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

describe('DeuceToSevenPage', () => {
  beforeEach(() => {
    mockExec.mockReset();
  });

  it('calls reset on mount', async () => {
    mockExec.mockResolvedValue(baseState());
    renderWithProviders(<DeuceToSevenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders the pot and dealer label', async () => {
    mockExec.mockResolvedValue(baseState({ pot: 120, dealerIdx: 2 }));
    renderWithProviders(<DeuceToSevenPage />);
    await waitFor(() => expect(screen.getByText('120')).toBeInTheDocument());
    // Dealer renders via playerName (CPU 2), not the raw index.
    expect(screen.getAllByText('CPU 2').length).toBeGreaterThan(0);
    expect(screen.queryByText(/Player 2|プレイヤー 2/)).not.toBeInTheDocument();
  });

  it('renders the pre-draw badge on the initial deal', async () => {
    mockExec.mockResolvedValue(baseState({ phase: DeuceToSevenPhase.DEAL, drawIndex: 0 }));
    renderWithProviders(<DeuceToSevenPage />);
    await waitFor(() => expect(screen.getByText('プリドロー')).toBeInTheDocument());
  });

  it('renders the draw counter badge with the current draw and the max cap', async () => {
    mockExec.mockResolvedValue(baseState({ phase: DeuceToSevenPhase.DRAW, drawIndex: 2 }));
    renderWithProviders(<DeuceToSevenPage />);
    // The badge shows current/max (2/3), so the player sees the 3-draw limit.
    await waitFor(() => expect(screen.getAllByText('ドロー 2/3').length).toBeGreaterThan(0));
  });

  it('shows the max cap in the badge on the first draw', async () => {
    mockExec.mockResolvedValue(baseState({ phase: DeuceToSevenPhase.DRAW, drawIndex: 1 }));
    renderWithProviders(<DeuceToSevenPage />);
    await waitFor(() => expect(screen.getAllByText('ドロー 1/3').length).toBeGreaterThan(0));
  });

  it('shows the end message at showdown', async () => {
    mockExec.mockResolvedValue(
      baseState({
        phase: DeuceToSevenPhase.END,
        gameEndFlag: true,
        message: 'あなたの勝ちです。',
        messageCode: 'deucetoseven.result.win',
        roundResults: [{ playerIdx: 0, handRank: 0, handName: 'High Card', wonAmount: 40 }],
      }),
    );
    renderWithProviders(<DeuceToSevenPage />);
    await waitFor(() => expect(screen.getByText('あなたの勝ちです。')).toBeInTheDocument());
  });

  it('renders the human hand name translated for the current locale', async () => {
    mockExec.mockResolvedValue(
      baseState({
        phase: DeuceToSevenPhase.END,
        gameEndFlag: true,
        // Backend sends the English category; the badge should show the ja label.
        players: [
          humanPlayer({ handRank: 1, handName: 'One Pair', folded: false }),
          cpuPlayer(1),
          cpuPlayer(2),
          cpuPlayer(3),
        ],
      }),
    );
    renderWithProviders(<DeuceToSevenPage />);
    await waitFor(() => expect(screen.getByText('ワンペア')).toBeInTheDocument());
    expect(screen.queryByText('One Pair')).not.toBeInTheDocument();
  });

  it('falls back to the server hand name when the rank is out of range', async () => {
    mockExec.mockResolvedValue(
      baseState({
        phase: DeuceToSevenPhase.END,
        gameEndFlag: true,
        players: [
          humanPlayer({ handRank: 99, handName: 'Mystery Hand', folded: false }),
          cpuPlayer(1),
          cpuPlayer(2),
          cpuPlayer(3),
        ],
      }),
    );
    renderWithProviders(<DeuceToSevenPage />);
    await waitFor(() => expect(screen.getByText('Mystery Hand')).toBeInTheDocument());
  });

  it('wires betting buttons during the human turn', async () => {
    mockExec.mockResolvedValue(baseState({ phase: DeuceToSevenPhase.DEAL, currentTurn: 0 }));
    renderWithProviders(<DeuceToSevenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /チェック/ })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /チェック/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('check', undefined, undefined, undefined, 0));
  });

  it('wires fold and allin', async () => {
    mockExec.mockResolvedValue(baseState({ phase: DeuceToSevenPhase.BET, currentTurn: 0, lastBet: 20 }));
    renderWithProviders(<DeuceToSevenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /フォールド/ })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /フォールド/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('fold', undefined, undefined, undefined, 0));

    fireEvent.click(screen.getByRole('button', { name: /オールイン/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('allin', undefined, undefined, undefined, 0));
  });

  it('exposes exchange and stand during draw phase', async () => {
    mockExec.mockResolvedValue(baseState({ phase: DeuceToSevenPhase.DRAW, drawIndex: 1, currentTurn: 0 }));
    renderWithProviders(<DeuceToSevenPage />);

    await waitFor(() => expect(screen.getByRole('button', { name: '交換' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: '交換' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('exchange', []));

    fireEvent.click(screen.getByRole('button', { name: /スタンド/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('stand'));
  });

  it('hides betting controls when the human has folded', async () => {
    mockExec.mockResolvedValue(
      baseState({
        phase: DeuceToSevenPhase.BET,
        players: [humanPlayer({ folded: true }), cpuPlayer(1), cpuPlayer(2), cpuPlayer(3)],
      }),
    );
    renderWithProviders(<DeuceToSevenPage />);
    await waitFor(() => expect(screen.getByText(/フォールド/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: /チェック/ })).toBeNull();
  });

  it('shows next game button at end phase', async () => {
    mockExec.mockResolvedValue(baseState({ phase: DeuceToSevenPhase.END, gameEndFlag: true }));
    renderWithProviders(<DeuceToSevenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /次のゲーム/ })).toBeInTheDocument());
  });

  it('toggles betting limit and reissues reset with the new config', async () => {
    mockExec.mockResolvedValue(baseState({ phase: DeuceToSevenPhase.END, gameEndFlag: true }));
    renderWithProviders(<DeuceToSevenPage />);
    const summary = await screen.findByText('設定');
    fireEvent.click(summary);

    const limitSelect = await screen.findByLabelText(/リミット/);
    fireEvent.change(limitSelect, { target: { value: '2' } });
    fireEvent.click(screen.getByRole('button', { name: /次のゲーム/ }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenLastCalledWith('reset', undefined, undefined, { bettingLimit: 2, cpuMetaAI: false }),
    );
  });

  it('shows the made-low banner and pulses the stand button when the hand is already an 8-or-better low', async () => {
    // The default humanPlayer() holds 7-5-4-3-2 → a made pat low.
    mockExec.mockResolvedValue(baseState({ phase: DeuceToSevenPhase.DRAW, drawIndex: 1, currentTurn: 0 }));
    renderWithProviders(<DeuceToSevenPage />);
    await waitFor(() => expect(screen.getByTestId('d7-stand-btn')).toBeInTheDocument());

    expect(screen.getByTestId('d7-made-low-banner')).toBeInTheDocument();
    expect(screen.getByTestId('d7-stand-btn').className).toContain('animate-pulse');
  });

  it('hides the banner and skips the pulse when the hand is not a made low (pair)', async () => {
    mockExec.mockResolvedValue(
      baseState({
        phase: DeuceToSevenPhase.DRAW,
        drawIndex: 1,
        currentTurn: 0,
        players: [
          humanPlayer({
            cards: [
              { design: 'SPADE', value: 2 },
              { design: 'HEART', value: 2 },
              { design: 'DIAMOND', value: 4 },
              { design: 'CLOVER', value: 6 },
              { design: 'SPADE', value: 8 },
            ],
          }),
          cpuPlayer(1),
          cpuPlayer(2),
          cpuPlayer(3),
        ],
      }),
    );
    renderWithProviders(<DeuceToSevenPage />);
    await waitFor(() => expect(screen.getByTestId('d7-stand-btn')).toBeInTheDocument());

    expect(screen.queryByTestId('d7-made-low-banner')).not.toBeInTheDocument();
    expect(screen.getByTestId('d7-stand-btn').className).not.toContain('animate-pulse');
  });

  it('lifts the best-low core cards and dims the draw candidates', async () => {
    // The lift/dim assist follows the hint setting, which defaults to off.
    localStorage.setItem('hint_enabled_deucetoseven', 'true');
    // Hand 2-2-4-6-8: the duplicate 2 is a draw candidate; the rest are kept.
    mockExec.mockResolvedValue(
      baseState({
        phase: DeuceToSevenPhase.DRAW,
        drawIndex: 1,
        currentTurn: 0,
        players: [
          humanPlayer({
            cards: [
              { design: 'SPADE', value: 2 },
              { design: 'HEART', value: 2 },
              { design: 'DIAMOND', value: 4 },
              { design: 'CLOVER', value: 6 },
              { design: 'SPADE', value: 8 },
            ],
          }),
          cpuPlayer(1),
          cpuPlayer(2),
          cpuPlayer(3),
        ],
      }),
    );
    renderWithProviders(<DeuceToSevenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '交換' })).toBeInTheDocument());

    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    expect(cardButtons[0].className).toContain('-translate-y-1'); // kept 2
    expect(cardButtons[1].className).toContain('opacity-50'); // duplicate 2 → draw candidate
  });

  it('marks a selected card as pressed in draw phase', async () => {
    mockExec.mockResolvedValue(baseState({ phase: DeuceToSevenPhase.DRAW, drawIndex: 1, currentTurn: 0 }));
    renderWithProviders(<DeuceToSevenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '交換' })).toBeInTheDocument());

    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    expect(cardButtons.length).toBe(5);
    fireEvent.click(cardButtons[0]);
    await waitFor(() => expect(cardButtons[0]).toHaveAttribute('aria-pressed', 'true'));
  });

  it('renders the draw without a seated human', async () => {
    // Spectator-shaped state: no human seat, so there is no hand to advise on.
    localStorage.clear();
    localStorage.setItem('hint_enabled_deucetoseven', 'true');
    mockExec.mockReset();
    const spectator = baseState({ phase: DeuceToSevenPhase.DRAW, drawIndex: 1 });
    mockExec.mockResolvedValue({ ...spectator, players: spectator.players.filter((p) => !p.isHuman) });
    renderWithProviders(<DeuceToSevenPage />);
    await waitFor(() => expect(screen.getAllByRole('button').length).toBeGreaterThan(0));
    expect(document.querySelectorAll('.-translate-y-1, .opacity-50').length).toBe(0);
  });

  it('keeps the exchange lift/dim off outside the exchange window even with hints on', async () => {
    // Outside the draw the player cannot act on "keep these" advice, so the
    // assist stays off however the hint setting is configured.
    localStorage.clear();
    localStorage.setItem('hint_enabled_deucetoseven', 'true');
    mockExec.mockReset();
    mockExec.mockResolvedValue(baseState({ phase: DeuceToSevenPhase.SHOWDOWN }));
    renderWithProviders(<DeuceToSevenPage />);
    await waitFor(() => expect(screen.getAllByRole('button').length).toBeGreaterThan(0));
    expect(document.querySelectorAll('.-translate-y-1, .opacity-50').length).toBe(0);
  });

  it('keeps the exchange lift/dim off until hints are switched on', async () => {
    // The assist is a hint in all but name, so it follows the hint setting —
    // which defaults to off (#4701/#4702).
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue(baseState({ phase: DeuceToSevenPhase.DRAW, drawIndex: 1 }));
    const { unmount } = renderWithProviders(<DeuceToSevenPage />);
    await waitFor(() => expect(screen.getAllByRole('button').length).toBeGreaterThan(0));
    expect(document.querySelectorAll('.-translate-y-1, .opacity-50').length).toBe(0);
    unmount();

    localStorage.setItem('hint_enabled_deucetoseven', 'true');
    mockExec.mockResolvedValue(baseState({ phase: DeuceToSevenPhase.DRAW, drawIndex: 1 }));
    renderWithProviders(<DeuceToSevenPage />);
    await waitFor(() => expect(document.querySelectorAll('.-translate-y-1, .opacity-50').length).toBeGreaterThan(0));
  });
});
