import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, highcardflushApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, HighCardFlushResponse } from '../types/card';
import { HighCardFlushPage } from './HighCardFlushPage';

vi.mock('../hooks/useGameHint');

const cliModeDisabled = {
  cliEnabled: false,
  toggleCli: vi.fn(),
  logEntries: [],
  addInput: vi.fn(),
  addOutput: vi.fn(),
  addError: vi.fn(),
  clearLog: vi.fn(),
};
vi.mock('../hooks/useCliMode', () => ({ useCliMode: vi.fn() }));

vi.mock('../api/gameApi', () => ({
  highcardflushApi: { exec: vi.fn() },
  actionLogApi: { highcardflush: vi.fn() },
}));

const mockExec = vi.mocked(highcardflushApi.exec);
const card = (design: CardDesign, value: number): Card => ({ design, value });

const betPhaseState: HighCardFlushResponse = {
  playerHand: [],
  dealerHand: [],
  phase: 1,
  chips: 1000,
  anteBet: 0,
  flushBonusBet: 0,
  straightFlushBet: 0,
  raiseBet: 0,
  result: 0,
  antePayout: 0,
  raisePayout: 0,
  flushBonusPayout: 0,
  straightFlushPayout: 0,
  totalPayout: 0,
  dealerQualified: false,
  playerFlushLen: 0,
  dealerFlushLen: 0,
  playerStraightFlushLen: 0,
  maxRaiseMultiplier: 1,
  message: '',
};

const actionPhase5Flush: HighCardFlushResponse = {
  ...betPhaseState,
  phase: 2,
  playerHand: [
    card('SPADE', 5),
    card('SPADE', 6),
    card('SPADE', 7),
    card('SPADE', 11),
    card('SPADE', 13),
    card('HEART', 9),
    card('CLOVER', 4),
  ],
  anteBet: 100,
  chips: 900,
  playerFlushLen: 5,
  maxRaiseMultiplier: 2,
};

const endPhasePlayerWins: HighCardFlushResponse = {
  ...actionPhase5Flush,
  phase: 3,
  dealerHand: [
    card('CLOVER', 4),
    card('CLOVER', 6),
    card('CLOVER', 9),
    card('HEART', 8),
    card('DIAMOND', 11),
    card('SPADE', 2),
    card('DIAMOND', 13),
  ],
  raiseBet: 100,
  result: 1,
  antePayout: 200,
  raisePayout: 200,
  totalPayout: 400,
  dealerQualified: true,
  dealerFlushLen: 3,
  message: 'Player wins!',
  messageCode: 'highcardflush.result.playerWins',
};

const endPhaseFold: HighCardFlushResponse = {
  ...actionPhase5Flush,
  phase: 3,
  raiseBet: 0,
  result: -1,
  totalPayout: 0,
  dealerHand: [],
  message: 'Folded',
  messageCode: 'highcardflush.result.fold',
};

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  vi.mocked(useCliMode).mockReturnValue(cliModeDisabled);
});

describe('HighCardFlushPage', () => {
  it('renders skeleton before state loads', () => {
    mockExec.mockReturnValue(new Promise(() => {}));
    renderWithProviders(<HighCardFlushPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('renders bet phase with bet button on mount', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<HighCardFlushPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument();
  });

  it('labels player flush vs non-flush cards with text, not just glow/opacity', async () => {
    mockExec.mockResolvedValue(actionPhase5Flush);
    renderWithProviders(<HighCardFlushPage />);
    // ♠ cards form the flush; the ♥ is not part of it.
    await waitFor(() => expect(screen.getByRole('img', { name: '♠ 5 フラッシュ対象' })).toBeInTheDocument());
    expect(screen.getByRole('img', { name: '♥ 9 フラッシュ対象外' })).toBeInTheDocument();
  });

  it('labels the dealer flush cards at showdown', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<HighCardFlushPage />);
    // Dealer's ♣ cards form its 3-card flush; the ♥ is not in it.
    await waitFor(() => expect(screen.getByRole('img', { name: '♣ 4 フラッシュ対象' })).toBeInTheDocument());
    expect(screen.getByRole('img', { name: '♥ 8 フラッシュ対象外' })).toBeInTheDocument();
  });

  it('shows action phase with raise/fold buttons matching multiplier cap', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(actionPhase5Flush);
    renderWithProviders(<HighCardFlushPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('button', { name: /レイズ x2/ })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /レイズ x1/ })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /レイズ x3/ })).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
  });

  it('previews the resulting bet amount (anteBet × multiplier) on each raise button', async () => {
    // actionPhase5Flush has anteBet=100, maxRaiseMultiplier=2 → x1 = 100, x2 = 200.
    mockExec.mockResolvedValue(actionPhase5Flush);
    renderWithProviders(<HighCardFlushPage />);
    await waitFor(() => expect(screen.getByTestId('raise-1x')).toBeInTheDocument());
    expect(screen.getByTestId('raise-1x')).toHaveTextContent('レイズ x1（100）');
    expect(screen.getByTestId('raise-2x')).toHaveTextContent('レイズ x2（200）');
    expect(screen.queryByTestId('raise-3x')).not.toBeInTheDocument();
  });

  it('scales the previewed bet amount with a larger ante', async () => {
    // anteBet=300, maxRaiseMultiplier=3 → x1 = 300, x2 = 600, x3 = 900.
    mockExec.mockResolvedValue({
      ...actionPhase5Flush,
      anteBet: 300,
      playerFlushLen: 6,
      maxRaiseMultiplier: 3,
    });
    renderWithProviders(<HighCardFlushPage />);
    await waitFor(() => expect(screen.getByTestId('raise-3x')).toBeInTheDocument());
    expect(screen.getByTestId('raise-1x')).toHaveTextContent('レイズ x1（300）');
    expect(screen.getByTestId('raise-2x')).toHaveTextContent('レイズ x2（600）');
    expect(screen.getByTestId('raise-3x')).toHaveTextContent('レイズ x3（900）');
  });

  it('calls raise API on raise button click', async () => {
    mockExec.mockResolvedValueOnce(actionPhase5Flush).mockResolvedValueOnce(endPhasePlayerWins);
    renderWithProviders(<HighCardFlushPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /レイズ x2/ })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /レイズ x2/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('raise', undefined, undefined, undefined, 2));
  });

  it('shows end phase with player wins and payout breakdown', async () => {
    mockExec.mockResolvedValueOnce(actionPhase5Flush).mockResolvedValueOnce(endPhasePlayerWins);
    renderWithProviders(<HighCardFlushPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /レイズ x1/ })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /レイズ x1/ }));
    await waitFor(() => expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument());
    expect(screen.getByText(/合計: 400/)).toBeInTheDocument();
  });

  it('shows end phase with fold', async () => {
    mockExec.mockResolvedValueOnce(actionPhase5Flush).mockResolvedValueOnce(endPhaseFold);
    renderWithProviders(<HighCardFlushPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'フォールド' }));
    await waitFor(() => expect(mockExec).toHaveBeenLastCalledWith('fold'));
  });

  it('keyboard shortcut "1" triggers raise 1x during action phase', async () => {
    mockExec.mockResolvedValue(actionPhase5Flush);
    renderWithProviders(<HighCardFlushPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /レイズ x1/ })).toBeInTheDocument());

    fireEvent.keyDown(document, { key: '1' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('raise', undefined, undefined, undefined, 1));
  });

  it('keyboard shortcut "f" triggers fold during action phase', async () => {
    mockExec.mockResolvedValue(actionPhase5Flush);
    renderWithProviders(<HighCardFlushPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument());

    fireEvent.keyDown(document, { key: 'f' });
    await waitFor(() => expect(mockExec).toHaveBeenLastCalledWith('fold'));
  });

  it('keyboard shortcut "r" triggers reset during end phase', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<HighCardFlushPage />);
    await waitFor(() => expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument());

    fireEvent.keyDown(document, { key: 'r' });
    await waitFor(() => expect(mockExec).toHaveBeenLastCalledWith('reset'));
  });

  it('renders payout reference panel during bet phase', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<HighCardFlushPage />);
    await waitFor(() => expect(screen.getByText('配当表')).toBeInTheDocument());
    expect(screen.getByText(/レイズルール/)).toBeInTheDocument();
    expect(screen.getByText(/ディーラークオリファイ/)).toBeInTheDocument();
  });

  it('shows hint tooltip when hints are enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'raise2x', reason: 'hintReason.raise2x', confidence: 'strong' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    mockExec.mockResolvedValue(actionPhase5Flush);
    renderWithProviders(<HighCardFlushPage />);
    await waitFor(() => expect(screen.getByText(/5枚フラッシュ — 2倍レイズ推奨/)).toBeInTheDocument());
  });

  it('updates ante input value', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<HighCardFlushPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    const anteInput = screen.getByLabelText('アンテ') as HTMLInputElement;
    fireEvent.change(anteInput, { target: { value: '200' } });
    expect(anteInput.value).toBe('200');
  });

  it('opens action log panel from end phase', async () => {
    mockExec.mockResolvedValueOnce(endPhasePlayerWins);
    // Mock the log fetch
    vi.mocked(actionLogApi.highcardflush).mockResolvedValue({ entries: [] });
    renderWithProviders(<HighCardFlushPage />);
    await waitFor(() => expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '棋譜を見る' }));
    await waitFor(() => expect(actionLogApi.highcardflush).toHaveBeenCalled());
  });

  it('renders dealer hand at end phase with qualified label', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<HighCardFlushPage />);
    await waitFor(() => expect(screen.getByText(/クオリファイ/)).toBeInTheDocument());
  });

  it('renders 3x raise button when maxRaiseMultiplier is 3', async () => {
    mockExec.mockResolvedValue({
      ...actionPhase5Flush,
      playerFlushLen: 6,
      maxRaiseMultiplier: 3,
    });
    renderWithProviders(<HighCardFlushPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /レイズ x3/ })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /レイズ x2/ })).toBeInTheDocument();
  });

  it('lifts and glows the cards in the longest flush, dims off-suit cards', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(actionPhase5Flush);
    renderWithProviders(<HighCardFlushPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('button', { name: /レイズ x2/ })).toBeInTheDocument());
    // actionPhase5Flush has 5 SPADE + 1 HEART + 1 CLOVER → SPADE is the longest flush.
    const flushCards = document.querySelectorAll('[data-flush-card="true"]');
    const offCards = document.querySelectorAll('[data-flush-card="false"]');
    expect(flushCards.length).toBe(5);
    expect(offCards.length).toBe(2);
    expect((flushCards[0] as HTMLElement).className).toContain('drop-shadow-[0_0_8px_var(--color-ds-warning)]');
    expect((offCards[0] as HTMLElement).className).toContain('opacity-50');
  });

  it('does not dim or highlight the dealer hand until end phase (no spoilers)', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce({
      ...actionPhase5Flush,
      dealerHand: [card('SPADE', 2)],
    });
    renderWithProviders(<HighCardFlushPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('button', { name: /レイズ x2/ })).toBeInTheDocument());
    // Scope by data-card-section so we don't rely on alt-text matching (which
    // would silently start testing the wrong element if AnimatedCard's alt
    // format ever changes).
    const dealerWrappers = document.querySelectorAll('[data-card-section="dealer"]');
    expect(dealerWrappers.length).toBe(1);
    const dealerWrapper = dealerWrappers[0] as HTMLElement;
    expect(dealerWrapper.className).not.toContain('drop-shadow');
    expect(dealerWrapper.className).not.toContain('opacity-50');
  });

  it('renders the CLI terminal and hides the betting UI when CLI mode is enabled', async () => {
    vi.mocked(useCliMode).mockReturnValue({ ...cliModeDisabled, cliEnabled: true });
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<HighCardFlushPage />);
    await waitFor(() => expect(screen.getByRole('log')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
  });
});
