import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { TutorialProvider } from '../providers/TutorialProvider';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, VideoPokerResponse } from '../types/card';
import type { TutorialConfig } from '../types/tutorial';
import { VideoPokerGameContent } from './VideoPokerGameContent';

vi.mock('../api/gameApi', () => ({
  videopokerApi: { exec: vi.fn() },
  actionLogApi: { videopoker: vi.fn().mockResolvedValue([]) },
}));

const mockExec = vi.fn();

const card = (design: CardDesign, value: number): Card => ({ design, value });

const betPhaseState: VideoPokerResponse = {
  hand: [],
  phase: 1,
  chips: 1000,
  betAmount: 0,
  result: 0,
  payout: 0,
  handRank: 0,
  handName: '',
  heldIndices: [false, false, false, false, false],
  variantName: 'jacksorbetter',
  message: '',
};

const drawPhaseState: VideoPokerResponse = {
  hand: [card('SPADE', 1), card('HEART', 11), card('CLOVER', 5), card('DIAMOND', 8), card('SPADE', 13)],
  phase: 2,
  chips: 997,
  betAmount: 3,
  result: 0,
  payout: 0,
  handRank: 0,
  handName: '',
  heldIndices: [false, false, false, false, false],
  variantName: 'jacksorbetter',
  message: '',
};

const resultPhaseWin: VideoPokerResponse = {
  hand: [card('SPADE', 11), card('CLOVER', 11), card('HEART', 3), card('DIAMOND', 5), card('SPADE', 9)],
  phase: 3,
  chips: 1001,
  betAmount: 1,
  result: 1,
  payout: 5,
  handRank: 1,
  handName: 'Jacks or Better',
  heldIndices: [true, true, false, false, false],
  variantName: 'jacksorbetter',
  message: 'Jacks or Better! You win!',
  messageCode: 'videopoker.result.win',
  messageParams: { handName: 'Jacks or Better', payout: '5' },
};

const resultPhaseLose: VideoPokerResponse = {
  hand: [card('SPADE', 2), card('CLOVER', 5), card('HEART', 7), card('DIAMOND', 9), card('SPADE', 11)],
  phase: 3,
  chips: 999,
  betAmount: 1,
  result: -1,
  payout: 0,
  handRank: 0,
  handName: '',
  heldIndices: [false, false, false, false, false],
  variantName: 'jacksorbetter',
  message: 'No winning hand.',
  messageCode: 'videopoker.result.lose',
};

const payoutRows = ['royalFlush5', 'royalFlush'];

const tutorialConfig: TutorialConfig = {
  gameName: 'videopoker',
  steps: [],
};

function renderContent() {
  return renderWithProviders(
    <TutorialProvider config={tutorialConfig} translateMessage={(k: string) => k}>
      <VideoPokerGameContent
        gameName="videopoker"
        i18nNamespace="videopoker"
        apiExec={mockExec}
        payoutTableRows={payoutRows}
      />
    </TutorialProvider>,
  );
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe('VideoPokerGameContent', () => {
  it('calls reset on mount and renders bet phase', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderContent();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByText(/チップ.*1000/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /ディール/ })).toBeInTheDocument();
  });

  it('renders draw phase with hold toggles', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(drawPhaseState);
    renderContent();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.click(screen.getByRole('button', { name: /ディール/ }));
    await waitFor(() => expect(screen.getByRole('button', { name: /ドロー/ })).toBeInTheDocument());

    // Toggle hold on first card
    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    expect(cardButtons.length).toBe(5);
    fireEvent.click(cardButtons[0]);
    expect(cardButtons[0]).toHaveAttribute('aria-pressed', 'true');
  });

  it('sends hold indices on draw', async () => {
    mockExec
      .mockResolvedValueOnce(betPhaseState)
      .mockResolvedValueOnce(drawPhaseState)
      .mockResolvedValueOnce(resultPhaseWin);
    renderContent();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.click(screen.getByRole('button', { name: /ディール/ }));
    await waitFor(() => expect(screen.getByRole('button', { name: /ドロー/ })).toBeInTheDocument());

    // Hold first two cards
    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    fireEvent.click(cardButtons[0]);
    fireEvent.click(cardButtons[1]);

    fireEvent.click(screen.getByRole('button', { name: /ドロー/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hold', undefined, [0, 1]));
  });

  it('renders result phase with payout on win', async () => {
    mockExec.mockResolvedValue(resultPhaseWin);
    renderContent();
    await waitFor(() => expect(screen.getByText(/配当.*5/)).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /次のハンド/ })).toBeInTheDocument();
  });

  it('renders result phase on lose', async () => {
    mockExec.mockResolvedValue(resultPhaseLose);
    renderContent();
    await waitFor(() => expect(screen.getByRole('button', { name: /次のハンド/ })).toBeInTheDocument());
  });

  it('shows reset confirmation dialog', async () => {
    mockExec.mockResolvedValue(resultPhaseWin);
    renderContent();
    await waitFor(() => expect(screen.getByRole('button', { name: /次のハンド/ })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: /次のハンド/ }));
    await waitFor(() => expect(screen.getByText(/リセットしますか/)).toBeInTheDocument());
  });

  it('changes bet amount with selector', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderContent();
    await waitFor(() => expect(screen.getByRole('button', { name: /ディール/ })).toBeInTheDocument());
    const select = screen.getByLabelText(/コイン数/);
    fireEvent.change(select, { target: { value: '5' } });
    fireEvent.click(screen.getByRole('button', { name: /ディール/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 5));
  });

  it('renders payout table collapsed by default in draw phase', async () => {
    mockExec.mockResolvedValue(drawPhaseState);
    renderContent();
    await waitFor(() => expect(screen.getByText(/配当表/)).toBeInTheDocument());
    // Payout table should be collapsed (details element should not have open attribute)
    const details = screen.getByText(/配当表/).closest('details');
    expect(details).not.toHaveAttribute('open');
  });

  it('shows payout table during bet phase for reference', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderContent();
    await waitFor(() => expect(screen.getByRole('button', { name: /ディール/ })).toBeInTheDocument());
    expect(screen.getByText(/配当表/)).toBeInTheDocument();
  });

  it('clicking card in result phase does not toggle hold', async () => {
    mockExec.mockResolvedValue(resultPhaseWin);
    renderContent();
    await waitFor(() => expect(screen.getByRole('button', { name: /次のハンド/ })).toBeInTheDocument());
    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    expect(cardButtons.length).toBe(5);
    // heldIndices[0] is true in resultPhaseWin, so aria-pressed starts as "true"
    expect(cardButtons[0]).toHaveAttribute('aria-pressed', 'true');
    fireEvent.click(cardButtons[0]);
    // Should remain held because toggleHold returns early when not in draw phase
    expect(cardButtons[0]).toHaveAttribute('aria-pressed', 'true');
  });

  it('shows action log panel when view log button is clicked', async () => {
    mockExec.mockResolvedValue(resultPhaseWin);
    renderContent();
    await waitFor(() => expect(screen.getByRole('button', { name: /棋譜を見る/ })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: /棋譜を見る/ }));
    // After clicking, the action log panel should appear (even if empty)
    await waitFor(() => expect(screen.getByText(/棋譜/)).toBeInTheDocument());
  });

  it('renders result phase with undefined heldIndices (falls back to empty)', async () => {
    const resultNoHeld = {
      ...resultPhaseLose,
      heldIndices: undefined,
    } as unknown as VideoPokerResponse;
    mockExec.mockResolvedValue(resultNoHeld);
    renderContent();
    await waitFor(() => expect(screen.getByRole('button', { name: /次のハンド/ })).toBeInTheDocument());
    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    expect(cardButtons.every((b) => b.getAttribute('aria-pressed') === 'false')).toBe(true);
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderContent();
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });
});
