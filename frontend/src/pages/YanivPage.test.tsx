import { fireEvent, screen, waitFor } from '@testing-library/react';
import i18n from 'i18next';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { yanivApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, YanivResponse } from '../types/card';
import { YanivPhase } from '../types/phases';
import { YanivPage } from './YanivPage';

vi.mock('../api/gameApi', () => ({
  yanivApi: { exec: vi.fn() },
  actionLogApi: { yaniv: vi.fn() },
}));

const mockExec = vi.mocked(yanivApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function player(id: number, isHuman: boolean, cards: Card[], over: Partial<YanivResponse['players'][number]> = {}) {
  return { id, isHuman, cardCount: cards.length, cards, score: 0, handTotal: 0, isEliminated: false, ...over };
}

function makeState(overrides: Partial<YanivResponse> = {}): YanivResponse {
  return {
    players: [
      player(0, true, [card('SPADE', 1), card('HEART', 2)], { handTotal: 3 }),
      player(1, false, []),
      player(2, false, []),
      player(3, false, []),
    ],
    phase: YanivPhase.DISCARD,
    roundNumber: 1,
    currentPlayerIdx: 0,
    pickupCards: [card('DIAMOND', 4)],
    drawPileCount: 39,
    gameEndFlag: false,
    winnerIdx: -1,
    callerIdx: -1,
    asafWinnerIdx: -1,
    isAsaf: false,
    roundScores: [],
    message: '',
    config: { cpuDifficulty: 1, scoreLimit: 200 },
    ...overrides,
  };
}

beforeEach(() => {
  localStorage.clear();
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('YanivPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<YanivPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('selects and discards cards in the discard phase', async () => {
    renderWithProviders(<YanivPage />);
    const discardBtn = await screen.findByTestId('discard-button');
    expect(discardBtn).toBeDisabled();
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('discard-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', { cardIndices: [0] }));
  });

  it('toggles card selection on and off', async () => {
    renderWithProviders(<YanivPage />);
    const card0 = await screen.findByTestId('hand-card-0');
    fireEvent.click(card0);
    expect(screen.getByTestId('discard-button')).toBeEnabled();
    fireEvent.click(card0);
    expect(screen.getByTestId('discard-button')).toBeDisabled();
  });

  it('does not warn for a valid single-card selection', async () => {
    renderWithProviders(<YanivPage />);
    const card0 = await screen.findByTestId('hand-card-0');
    fireEvent.click(card0);
    expect(screen.queryByTestId('discard-warning')).not.toBeInTheDocument();
  });

  it('warns when the selected combination is not a legal discard', async () => {
    renderWithProviders(<YanivPage />);
    // Default hand is [SPADE 1, HEART 2] — selecting both is a non-pair, non-run.
    fireEvent.click(await screen.findByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('hand-card-1'));
    const warning = await screen.findByTestId('discard-warning');
    expect(warning.textContent).toMatch(/同じ数字/);
    expect(screen.getByTestId('discard-button')).toHaveAttribute('title');
  });

  it('highlights the hand-total badge in success color and pulses the Yaniv button when total <= 5', async () => {
    renderWithProviders(<YanivPage />);
    const badge = await screen.findByTestId('hand-total-badge');
    expect(badge.className).toContain('bg-ds-success');
    const yanivBtn = screen.getByTestId('yaniv-button');
    expect(yanivBtn.className).toContain('ring-ds-success');
    expect(yanivBtn.className).toContain('animate-pulse');
  });

  it('explains the Yaniv threshold on the hand-total badge via a tooltip', async () => {
    renderWithProviders(<YanivPage />);
    const badge = await screen.findByTestId('hand-total-badge');
    expect(badge).toHaveAttribute('title');
    expect(badge.getAttribute('title')).toMatch(/5/);
  });

  it('soft-highlights the hand-total badge in warning color when total is 6-10', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          player(0, true, [card('SPADE', 8)], { handTotal: 8 }),
          player(1, false, []),
          player(2, false, []),
          player(3, false, []),
        ],
      }),
    );
    renderWithProviders(<YanivPage />);
    const badge = await screen.findByTestId('hand-total-badge');
    expect(badge.className).toContain('bg-ds-warning');
    expect(badge.className).not.toContain('bg-ds-success');
    expect(screen.getByTestId('yaniv-button').className).not.toContain('ring-ds-success');
  });

  it('leaves the hand-total badge unstyled when total exceeds 10', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          player(0, true, [card('SPADE', 13)], { handTotal: 15 }),
          player(1, false, []),
          player(2, false, []),
          player(3, false, []),
        ],
      }),
    );
    renderWithProviders(<YanivPage />);
    const badge = await screen.findByTestId('hand-total-badge');
    expect(badge.className).not.toContain('bg-ds-success');
    expect(badge.className).not.toContain('bg-ds-warning');
  });

  it('declares Yaniv when the hand total is low enough', async () => {
    renderWithProviders(<YanivPage />);
    const yanivBtn = await screen.findByTestId('yaniv-button');
    expect(yanivBtn).toBeEnabled();
    fireEvent.click(yanivBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('yaniv'));
  });

  it('disables Yaniv when the hand total is too high', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          player(0, true, [card('SPADE', 10)], { handTotal: 10 }),
          player(1, false, []),
          player(2, false, []),
          player(3, false, []),
        ],
      }),
    );
    renderWithProviders(<YanivPage />);
    expect(await screen.findByTestId('yaniv-button')).toBeDisabled();
  });

  it('draws from stock and takes a pickup card in the draw phase', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: YanivPhase.DRAW, pickupCards: [card('SPADE', 4), card('SPADE', 6)] }),
    );
    renderWithProviders(<YanivPage />);
    fireEvent.click(await screen.findByTestId('draw-stock-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
    fireEvent.click(screen.getByTestId('pickup-card-0'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawpickup', { end: 0 }));
  });

  it('marks only the two ends of a discarded run as pickup-able in the draw phase', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: YanivPhase.DRAW,
        pickupCards: [card('SPADE', 4), card('SPADE', 5), card('SPADE', 6)],
      }),
    );
    renderWithProviders(<YanivPage />);
    await screen.findByTestId('pickup-card-0');
    // Guide text is shown for a multi-card discard.
    expect(screen.getByTestId('pickup-hint')).toBeInTheDocument();
    // Both ends carry a "take" badge; the middle card does not.
    expect(screen.getByTestId('pickup-badge-0')).toBeInTheDocument();
    expect(screen.getByTestId('pickup-badge-2')).toBeInTheDocument();
    expect(screen.queryByTestId('pickup-badge-1')).not.toBeInTheDocument();
    // The middle card is not clickable and is marked as such.
    const middle = screen.getByTestId('pickup-card-1');
    expect(middle).toBeDisabled();
    expect(middle).toHaveAttribute('aria-disabled', 'true');
    // Ends are actionable and take the correct discard end.
    fireEvent.click(screen.getByTestId('pickup-card-2'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawpickup', { end: 1 }));
  });

  it('marks the single card of a single-card discard as pickup-able', async () => {
    mockExec.mockResolvedValue(makeState({ phase: YanivPhase.DRAW, pickupCards: [card('DIAMOND', 4)] }));
    renderWithProviders(<YanivPage />);
    await screen.findByTestId('pickup-card-0');
    expect(screen.getByTestId('pickup-badge-0')).toBeInTheDocument();
    // No middle-card guide for a lone card (no "ends only" ambiguity).
    expect(screen.queryByTestId('pickup-hint')).not.toBeInTheDocument();
    fireEvent.click(screen.getByTestId('pickup-card-0'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawpickup', { end: 0 }));
  });

  it('does not mark pickup cards outside the human draw turn', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: YanivPhase.DISCARD, pickupCards: [card('SPADE', 4), card('SPADE', 6)] }),
    );
    renderWithProviders(<YanivPage />);
    await screen.findByTestId('pickup-card-0');
    expect(screen.queryByTestId('pickup-badge-0')).not.toBeInTheDocument();
    expect(screen.queryByTestId('pickup-hint')).not.toBeInTheDocument();
    expect(screen.getByTestId('pickup-card-0')).toBeDisabled();
  });

  it('shows the next-round button at round end', async () => {
    mockExec.mockResolvedValue(makeState({ phase: YanivPhase.ROUND_END, callerIdx: 0 }));
    renderWithProviders(<YanivPage />);
    fireEvent.click(await screen.findByTestId('next-round-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('renders an eliminated indicator', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          player(0, true, [card('SPADE', 3)], { handTotal: 3 }),
          player(1, false, [], { isEliminated: true, score: 210 }),
          player(2, false, []),
          player(3, false, []),
        ],
      }),
    );
    renderWithProviders(<YanivPage />);
    await screen.findByTestId('discard-button');
    expect(screen.getByText('💀')).toBeInTheDocument();
  });

  it('renders the CLI terminal when CLI mode is enabled', async () => {
    localStorage.setItem('cli-mode-yaniv', 'true');
    renderWithProviders(<YanivPage />);
    expect(await screen.findByPlaceholderText(/コマンド/)).toBeInTheDocument();
    expect(screen.queryByTestId('discard-button')).not.toBeInTheDocument();
  });

  it('shows a retry button when an action fails', async () => {
    mockExec.mockResolvedValue(makeState({ phase: YanivPhase.DRAW }));
    renderWithProviders(<YanivPage />);
    const drawBtn = await screen.findByTestId('draw-stock-button');
    mockExec.mockRejectedValueOnce(new Error('boom'));
    fireEvent.click(drawBtn);
    const retry = await screen.findByText(NETWORK_ERROR_MESSAGE());
    mockExec.mockResolvedValue(makeState({ phase: YanivPhase.DRAW }));
    fireEvent.click(retry);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('renders CPU difficulty labels from the i18n bundle rather than literals', async () => {
    // A sentinel proves the label is *looked up*: a hardcoded 'Easy' would ignore it.
    const original = i18n.getResourceBundle('ja', 'yaniv') as Record<string, unknown>;
    i18n.addResourceBundle('ja', 'yaniv', { settings: { easy: '__EASY__' } }, true, true);
    try {
      renderWithProviders(<YanivPage />);
      await waitFor(() => expect(screen.getByRole('option', { name: '__EASY__' })).toBeInTheDocument());
    } finally {
      i18n.removeResourceBundle('ja', 'yaniv');
      i18n.addResourceBundle('ja', 'yaniv', original, true, true);
    }
  });
});

// #5629: CUI は上限の 80% を超えたプレイヤーに「脱落間近」を出しているのに、
// Web は同じ値 (state.config.scoreLimit) をレスポンスで受け取りながら
// 一度も読んでいなかった。
describe('YanivPage near-elimination warning', () => {
  it('warns the players who are close to the limit', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          player(0, true, [card('SPADE', 3)], { score: 170 }), // 200 の 85%
          player(1, false, [], { score: 100 }),
          player(2, false, [], { score: 190 }),
          player(3, false, [], { score: 20 }),
        ],
        config: { cpuDifficulty: 1, scoreLimit: 200 },
      }),
    );
    renderWithProviders(<YanivPage />);

    await waitFor(() => expect(screen.getByTestId('yv-near-out-0')).toBeInTheDocument());
    expect(screen.getByTestId('yv-near-out-2')).toBeInTheDocument();
    // 100 点はまだ半分なので出さない。
    expect(screen.queryByTestId('yv-near-out-1')).not.toBeInTheDocument();
  });

  // 脱落済みには出さない。すでに終わった話を警告しても意味がない。
  it('says nothing for a player already out', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          player(0, true, [card('SPADE', 3)], { score: 10 }),
          player(1, false, [], { score: 210, isEliminated: true }),
          player(2, false, [], { score: 20 }),
          player(3, false, [], { score: 20 }),
        ],
        config: { cpuDifficulty: 1, scoreLimit: 200 },
      }),
    );
    renderWithProviders(<YanivPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    expect(screen.queryByTestId('yv-near-out-1')).not.toBeInTheDocument();
  });
});
