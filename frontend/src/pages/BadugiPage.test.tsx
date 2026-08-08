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
    // Dealer renders via playerName (CPU 2), not the raw index.
    expect(screen.getAllByText('CPU 2').length).toBeGreaterThan(0);
    expect(screen.queryByText(/Player 2|プレイヤー 2/)).not.toBeInTheDocument();
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

  it('annotates card aria-labels with best-subset / exchange-candidate hints during the draw phase', async () => {
    // The lift/dim assist follows the hint setting, which defaults to off.
    localStorage.setItem('hint_enabled_badugi', 'true');
    // S1,H2,D3 form the best 3-card subset (lowest sum); the duplicate-suit S5
    // is the exchange candidate.
    mockExec.mockResolvedValue(
      baseState({
        phase: BadugiPhase.DRAW,
        currentTurn: 0,
        players: [
          humanPlayer({
            cards: [
              { design: 'SPADE', value: 1 },
              { design: 'HEART', value: 2 },
              { design: 'DIAMOND', value: 3 },
              { design: 'SPADE', value: 5 },
            ],
          }),
          cpuPlayer(1),
          cpuPlayer(2),
          cpuPlayer(3),
        ],
      }),
    );
    renderWithProviders(<BadugiPage />);
    await waitFor(() => expect(screen.getAllByRole('button', { name: /ベストハンドの一部/ }).length).toBe(3));
    expect(screen.getAllByRole('button', { name: /交換候補/ })).toHaveLength(1);
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

  // **サーバの handName は英語の生値。**バドゥーギの役名はポーカー役表と対応
  // しないので共通の訳表が無く、Web だけ英語のまま残っていた (#4987 の続き)。
  it('shows the localized hand name at showdown, not the raw English one', async () => {
    mockExec.mockResolvedValue(
      baseState({
        phase: BadugiPhase.END,
        gameEndFlag: true,
        players: [humanPlayer({ handSize: 4, handName: 'Badugi' }), cpuPlayer(1), cpuPlayer(2)],
      }),
    );
    renderWithProviders(<BadugiPage />);
    await waitFor(() => expect(screen.getByTestId('bg-hand-name')).toHaveTextContent('バドゥーギ'));
    // 生の英語名がどこにも漏れていないこと。
    expect(screen.queryByText('Badugi')).not.toBeInTheDocument();
  });

  // 未評価 (handSize 0) の席には役名を出さない。
  it('shows no hand name before the hand has been evaluated', async () => {
    mockExec.mockResolvedValue(baseState({ phase: BadugiPhase.END, gameEndFlag: true }));
    renderWithProviders(<BadugiPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
    expect(screen.queryByTestId('bg-hand-name')).not.toBeInTheDocument();
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

  it('exposes next-hand application notes for the betting-limit and meta-AI settings', async () => {
    mockExec.mockResolvedValue(baseState({ phase: BadugiPhase.END, gameEndFlag: true }));
    renderWithProviders(<BadugiPage />);
    const summary = await screen.findByText('設定');
    fireEvent.click(summary);

    // Each setting exposes a (?) help button that reveals a note explaining the
    // change only takes effect from the next reset (new hand).
    const helpButtons = await screen.findAllByRole('button', { name: '説明を表示' });
    expect(helpButtons.length).toBeGreaterThanOrEqual(2);

    fireEvent.click(helpButtons[0]);
    expect(screen.getByText(/リミットの変更は現在のハンドには反映されず/)).toBeInTheDocument();

    fireEvent.click(helpButtons[1]);
    expect(screen.getByText(/メタAIの変更は現在のハンドには反映されず/)).toBeInTheDocument();
  });

  it('shows the complete-Badugi banner and pulses the stand button when the hand is already a 4-card Badugi', async () => {
    // The default humanPlayer() has 4 distinct ranks AND 4 distinct suits → complete Badugi.
    mockExec.mockResolvedValue(baseState({ phase: BadugiPhase.DRAW, drawIndex: 1, currentTurn: 0 }));
    renderWithProviders(<BadugiPage />);
    await waitFor(() => expect(screen.getByTestId('bg-stand-btn')).toBeInTheDocument());

    expect(screen.getByTestId('bg-complete-badugi-banner')).toBeInTheDocument();
    expect(screen.getByTestId('bg-stand-btn').className).toContain('animate-pulse');
  });

  it('hides the banner and skips the pulse when the hand is incomplete (suit dupe)', async () => {
    mockExec.mockResolvedValue(
      baseState({
        phase: BadugiPhase.DRAW,
        drawIndex: 1,
        currentTurn: 0,
        players: [
          humanPlayer({
            cards: [
              { design: 'SPADE', value: 1 },
              { design: 'SPADE', value: 2 },
              { design: 'DIAMOND', value: 3 },
              { design: 'CLOVER', value: 4 },
            ],
          }),
          cpuPlayer(1),
          cpuPlayer(2),
          cpuPlayer(3),
        ],
      }),
    );
    renderWithProviders(<BadugiPage />);
    await waitFor(() => expect(screen.getByTestId('bg-stand-btn')).toBeInTheDocument());

    expect(screen.queryByTestId('bg-complete-badugi-banner')).not.toBeInTheDocument();
    expect(screen.getByTestId('bg-stand-btn').className).not.toContain('animate-pulse');
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

  it('marks every card with data-badugi-subset during the draw phase when the hand is a perfect Badugi', async () => {
    // The lift/dim assist follows the hint setting, which defaults to off.
    localStorage.setItem('hint_enabled_badugi', 'true');
    mockExec.mockResolvedValue(baseState({ phase: BadugiPhase.DRAW, drawIndex: 1, currentTurn: 0 }));
    renderWithProviders(<BadugiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '交換' })).toBeInTheDocument());

    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    expect(cardButtons.length).toBe(4);
    for (const btn of cardButtons) {
      expect(btn).toHaveAttribute('data-badugi-subset', 'true');
      expect(btn.className).toContain('-translate-y-1');
      expect(btn.className).not.toContain('opacity-50');
    }
  });

  it('dims duplicate-suit cards in the draw phase and omits data-badugi-subset on them', async () => {
    // The lift/dim assist follows the hint setting, which defaults to off.
    localStorage.setItem('hint_enabled_badugi', 'true');
    // Force a duplicate suit so one card falls out of the best subset (lowball: keep the lower ♠).
    mockExec.mockResolvedValue(
      baseState({
        phase: BadugiPhase.DRAW,
        drawIndex: 1,
        currentTurn: 0,
        players: [
          humanPlayer({
            cards: [
              { design: 'SPADE', value: 1 },
              { design: 'SPADE', value: 9 },
              { design: 'HEART', value: 3 },
              { design: 'DIAMOND', value: 6 },
            ],
          }),
          cpuPlayer(1),
          cpuPlayer(2),
          cpuPlayer(3),
        ],
      }),
    );
    renderWithProviders(<BadugiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '交換' })).toBeInTheDocument());

    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    // Subset = {SPADE 1, HEART 3, DIAMOND 6} = indices [0, 2, 3]; duplicate SPADE 9 = index 1.
    expect(cardButtons[0]).toHaveAttribute('data-badugi-subset', 'true');
    expect(cardButtons[1]).not.toHaveAttribute('data-badugi-subset');
    expect(cardButtons[1].className).toContain('opacity-50');
    expect(cardButtons[2]).toHaveAttribute('data-badugi-subset', 'true');
    expect(cardButtons[3]).toHaveAttribute('data-badugi-subset', 'true');
  });

  it('does not annotate subset cards outside the draw phase', async () => {
    // BET phase: subset hint should not appear (player cannot act on the dim/lift cue here).
    mockExec.mockResolvedValue(baseState({ phase: BadugiPhase.BET, currentTurn: 0, lastBet: 20 }));
    renderWithProviders(<BadugiPage />);
    await waitFor(() => expect(screen.getAllByText(/ベット/).length).toBeGreaterThan(0));
    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    for (const btn of cardButtons) {
      expect(btn).not.toHaveAttribute('data-badugi-subset');
      expect(btn.className).not.toContain('opacity-50');
      expect(btn.className).not.toContain('-translate-y-1');
    }
  });

  it('renders the draw without a seated human', async () => {
    // Spectator-shaped state: no human seat, so there is no hand to advise on.
    localStorage.clear();
    localStorage.setItem('hint_enabled_badugi', 'true');
    mockExec.mockReset();
    const spectator = baseState({ phase: BadugiPhase.DRAW, drawIndex: 1 });
    mockExec.mockResolvedValue({ ...spectator, players: spectator.players.filter((p) => !p.isHuman) });
    renderWithProviders(<BadugiPage />);
    await waitFor(() => expect(screen.getAllByRole('button').length).toBeGreaterThan(0));
    expect(document.querySelectorAll('.-translate-y-1, .opacity-50').length).toBe(0);
  });

  it('keeps the exchange lift/dim off outside the exchange window even with hints on', async () => {
    // Outside the draw the player cannot act on "keep these" advice, so the
    // assist stays off however the hint setting is configured.
    localStorage.clear();
    localStorage.setItem('hint_enabled_badugi', 'true');
    mockExec.mockReset();
    mockExec.mockResolvedValue(baseState({ phase: BadugiPhase.SHOWDOWN, currentTurn: 0 }));
    renderWithProviders(<BadugiPage />);
    await waitFor(() => expect(screen.getAllByRole('button').length).toBeGreaterThan(0));
    expect(document.querySelectorAll('.-translate-y-1, .opacity-50').length).toBe(0);
  });

  it('keeps the exchange lift/dim off until hints are switched on', async () => {
    // The assist is a hint in all but name, so it follows the hint setting —
    // which defaults to off (#4701/#4702).
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue(baseState({ phase: BadugiPhase.DRAW, drawIndex: 1, currentTurn: 0 }));
    const { unmount } = renderWithProviders(<BadugiPage />);
    await waitFor(() => expect(screen.getAllByRole('button').length).toBeGreaterThan(0));
    expect(document.querySelectorAll('.-translate-y-1, .opacity-50').length).toBe(0);
    unmount();

    localStorage.setItem('hint_enabled_badugi', 'true');
    mockExec.mockResolvedValue(baseState({ phase: BadugiPhase.DRAW, drawIndex: 1, currentTurn: 0 }));
    renderWithProviders(<BadugiPage />);
    await waitFor(() => expect(document.querySelectorAll('.-translate-y-1, .opacity-50').length).toBeGreaterThan(0));
  });
});
