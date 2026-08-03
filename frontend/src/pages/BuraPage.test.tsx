import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { buraApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BuraPlayer, BuraResponse, Card, CardDesign } from '../types/card';
import { BuraPage } from './BuraPage';

vi.mock('../api/gameApi', () => ({
  buraApi: { exec: vi.fn() },
  actionLogApi: { bura: vi.fn() },
}));

const mockExec = vi.mocked(buraApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function human(overrides?: Partial<BuraPlayer>): BuraPlayer {
  return {
    id: 0,
    isHuman: true,
    cardCount: 3,
    cards: [card('SPADE', 1), card('SPADE', 10), card('CLOVER', 7)],
    points: 0,
    hidden: false,
    ...overrides,
  };
}

function cpu(overrides?: Partial<BuraPlayer>): BuraPlayer {
  // A hidden seat arrives with a count and NO cards. Building the fixture this
  // way means a page that reads `cards` for the opponent renders nothing,
  // rather than quietly working off data the server never sends.
  return { id: 1, isHuman: false, cardCount: 3, cards: [], points: 0, hidden: true, ...overrides };
}

function makeState(overrides?: Partial<BuraResponse>): BuraResponse {
  return {
    players: [human(), cpu()],
    phase: 0,
    trickNumber: 0,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    currentLead: [],
    trumpSuit: 2,
    trumpCard: card('HEART', 7),
    stockRemaining: 29,
    winThreshold: 31,
    gameEndFlag: false,
    winnerIdx: -1,
    isDraw: false,
    message: '',
    ...overrides,
  };
}

describe('BuraPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<BuraPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('never renders the opponent hand as cards', async () => {
    renderWithProviders(<BuraPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    // The human's three cards are buttons; the opponent's are backs, which are
    // not. If the page ever renders the opponent from `cards`, this count moves.
    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    expect(cardButtons).toHaveLength(3);
  });

  it('plays the selected cards and clears the selection', async () => {
    renderWithProviders(<BuraPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    fireEvent.click(cardButtons[0]);
    fireEvent.click(cardButtons[1]);
    expect(cardButtons[0]).toHaveAttribute('aria-pressed', 'true');

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', [0, 1]));
  });

  it('refuses a mixed-suit lead', async () => {
    renderWithProviders(<BuraPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    // index 0 is a spade, index 2 a club -- illegal together as a lead.
    fireEvent.click(cardButtons[0]);
    fireEvent.click(cardButtons[2]);

    expect(screen.getByRole('button', { name: '出す' })).toBeDisabled();
  });

  it('requires a response to match the lead count exactly', async () => {
    mockExec.mockResolvedValue(
      makeState({ currentLead: [card('DIAMOND', 13), card('DIAMOND', 12)], leadPlayerIdx: 1, currentPlayerIdx: 0 }),
    );
    renderWithProviders(<BuraPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    const respond = screen.getByRole('button', { name: '2 枚で受ける' });

    fireEvent.click(cardButtons[0]);
    expect(respond).toBeDisabled();

    // Two cards of DIFFERENT suits are a legal response -- only a lead has to
    // be one suit. This is the case a "same suit" check would wrongly block.
    fireEvent.click(cardButtons[2]);
    expect(respond).toBeEnabled();
  });

  it('claims and declares through their own commands', async () => {
    renderWithProviders(<BuraPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '31点を宣言' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('claim'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '役を宣言' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declare'));
  });

  it('reports a draw distinctly from a loss', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: 1, gameEndFlag: true, isDraw: true, winnerIdx: -1, messageCode: 'bura.draw' }),
    );
    renderWithProviders(<BuraPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.getByText('流局 — 誰も宣言しませんでした')).toBeInTheDocument();
  });

  it('shows the trump suit once the indicator has been drawn', async () => {
    mockExec.mockResolvedValue(makeState({ trumpCard: undefined, trumpSuit: 2 }));
    renderWithProviders(<BuraPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.getByText('切札 ハート（指示カードは引かれた）')).toBeInTheDocument();
  });
});
