import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, prsiApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { PrsiResponse } from '../types/card';
import { PrsiPage } from './PrsiPage';

vi.mock('../api/gameApi', () => ({
  prsiApi: { exec: vi.fn() },
  actionLogApi: { prsi: vi.fn() },
}));

const mockPlaySound = vi.fn();
const mockSoundValue = {
  playSound: mockPlaySound,
  muted: false,
  toggleMute: vi.fn(),
  claimExecSound: vi.fn(),
  consumeExecClaim: () => false,
};
vi.mock('../providers/SoundProvider', () => ({
  SoundProvider: ({ children }: { children: React.ReactNode }) => children,
  useSound: () => mockSoundValue,
  // AnimatedCard AND the central taps (useGameApi / GamePageShell / ErrorAlert)
  // consume useOptionalSound; route it to the same spy and assert on specific
  // sound names so per-card deal sounds don't interfere.
  useOptionalSound: () => mockSoundValue,
}));

const mockExec = vi.mocked(prsiApi.exec);

const playPhaseState: PrsiResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 2,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
      ],
    },
    { id: 1, isHuman: false, cardCount: 5, cards: [] },
    { id: 2, isHuman: false, cardCount: 5, cards: [] },
    { id: 3, isHuman: false, cardCount: 5, cards: [] },
  ],
  phase: 0,
  currentPlayerIdx: 0,
  discardTop: { design: 'HEART', value: 7 },
  drawPileCount: 30,
  penaltyDrawCount: 0,
  pendingSkips: 0,
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
  config: { cpuDifficulty: 1 },
};

const gameEndState: PrsiResponse = {
  ...playPhaseState,
  phase: 1,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};

const gameEndByFlagState: PrsiResponse = {
  ...playPhaseState,
  phase: 0,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};

const cpuTurnState: PrsiResponse = {
  ...playPhaseState,
  currentPlayerIdx: 1,
};

const penaltyState: PrsiResponse = {
  ...playPhaseState,
  penaltyDrawCount: 2,
};

const noDiscardState: PrsiResponse = {
  ...playPhaseState,
  discardTop: null,
};

beforeEach(() => {
  mockExec.mockResolvedValue(playPhaseState);
  mockPlaySound.mockReset();
});

describe('PrsiPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<PrsiPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<PrsiPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1 }));
  });

  it('renders play phase with human cards', async () => {
    renderWithProviders(<PrsiPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
      expect(screen.getByAltText('♥ J')).toBeInTheDocument();
    });
  });

  it('highlights legal cards and dims illegal ones on the human turn', async () => {
    // Discard top ♥7; hand has ♥J (legal — same suit) and ♠A (illegal).
    renderWithProviders(<PrsiPage />);
    const legal = await screen.findByLabelText('♥ J');
    const illegal = screen.getByLabelText('♠ A');
    expect(legal).toHaveAttribute('data-legal', 'true');
    expect(legal.className).toContain('ring-ds-success');
    expect(illegal).toHaveAttribute('data-legal', 'false');
    expect(illegal.className).toContain('opacity-50');
    expect(illegal).toHaveAttribute('title', 'このカードは出せません（スート・数字のいずれも不一致）');
  });

  it('does not mark card legality on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<PrsiPage />);
    const card = await screen.findByLabelText('♥ J');
    expect(card).not.toHaveAttribute('data-legal');
  });

  it('renders play and draw buttons when human turn', async () => {
    renderWithProviders(<PrsiPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '出す' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '引く' })).toBeInTheDocument();
    });
  });

  it('play button disabled when not 1 card selected', async () => {
    renderWithProviders(<PrsiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeDisabled());
  });

  it('calls play command when play button is clicked', async () => {
    renderWithProviders(<PrsiPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0));
  });

  it('calls draw command when draw button is clicked', async () => {
    renderWithProviders(<PrsiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '引く' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '引く' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('does not show play/draw buttons when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<PrsiPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '引く' })).not.toBeInTheDocument();
  });

  it('shows the penalty indicator when penaltyDrawCount > 0', async () => {
    mockExec.mockResolvedValue(penaltyState);
    renderWithProviders(<PrsiPage />);
    await waitFor(() => expect(screen.getByTestId('penalty-indicator')).toBeInTheDocument());
  });

  it('does not show the penalty indicator when penaltyDrawCount is 0', async () => {
    renderWithProviders(<PrsiPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());
    expect(screen.queryByTestId('penalty-indicator')).not.toBeInTheDocument();
  });

  // **スキップも重ねられる (#4772)。**7 の累積ペナルティは「+N」バッジと警告
  // バナーで目立たせているのに、pendingSkips は一度も読まれていなかった。
  it('shows the skip indicator when pendingSkips > 0', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, pendingSkips: 2 });
    renderWithProviders(<PrsiPage />);
    const banner = await screen.findByTestId('skip-indicator');
    expect(banner).toHaveTextContent('2');
    // **読み上げにも届かせる。**色だけの警告は SR に伝わらない。
    expect(banner).toHaveAttribute('role', 'status');
  });

  it('does not show the skip indicator when pendingSkips is 0', async () => {
    renderWithProviders(<PrsiPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());
    expect(screen.queryByTestId('skip-indicator')).not.toBeInTheDocument();
  });

  it('shows discard top card', async () => {
    renderWithProviders(<PrsiPage />);
    await waitFor(() => {
      expect(screen.getByText('捨て札')).toBeInTheDocument();
      expect(screen.getByAltText('♥ 7')).toBeInTheDocument();
    });
  });

  it('does not show discard top when null', async () => {
    mockExec.mockResolvedValue(noDiscardState);
    renderWithProviders(<PrsiPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.queryByText('捨て札')).not.toBeInTheDocument();
  });

  it('shows CPU player areas', async () => {
    renderWithProviders(<PrsiPage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1.*5枚/)).toBeInTheDocument();
      expect(screen.getByText(/CPU 2.*5枚/)).toBeInTheDocument();
      expect(screen.getByText(/CPU 3.*5枚/)).toBeInTheDocument();
    });
  });

  it('shows game end with action log button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<PrsiPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('shows game end via gameEndFlag with non-1 phase', async () => {
    mockExec.mockResolvedValue(gameEndByFlagState);
    renderWithProviders(<PrsiPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('shows error alert', async () => {
    renderWithProviders(<PrsiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('card selection toggle via aria-pressed', async () => {
    renderWithProviders(<PrsiPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('reset button calls exec with confirm', async () => {
    renderWithProviders(<PrsiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1 }));
  });

  it('shows confirm dialog when reset button is clicked', async () => {
    renderWithProviders(<PrsiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('settings panel changes cpuDifficulty', async () => {
    renderWithProviders(<PrsiPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());

    fireEvent.click(screen.getByText('設定'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[0], { target: { value: '2' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 2 }));
  });

  it('draw pile info displayed', async () => {
    renderWithProviders(<PrsiPage />);
    await waitFor(() => expect(screen.getByText('山札: 30枚')).toBeInTheDocument());
  });

  it('renders the clickable stock pile with the remaining count', async () => {
    renderWithProviders(<PrsiPage />);
    const stock = await screen.findByTestId('prsi-stock');
    expect(stock).toHaveTextContent('30');
    expect(stock).not.toBeDisabled();
  });

  it('drawing via a stock click dispatches the draw action on the human turn', async () => {
    renderWithProviders(<PrsiPage />);
    const stock = await screen.findByTestId('prsi-stock');

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(stock);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('disables the stock pile when it is not the human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<PrsiPage />);
    const stock = await screen.findByTestId('prsi-stock');
    expect(stock).toBeDisabled();
  });

  it('shows the penalty badge on the stock pile when penaltyDrawCount > 0', async () => {
    mockExec.mockResolvedValue(penaltyState);
    renderWithProviders(<PrsiPage />);
    const badge = await screen.findByTestId('prsi-stock-penalty');
    expect(badge).toHaveTextContent('+2');
  });

  it('phase indicator shows your turn when human play turn', async () => {
    renderWithProviders(<PrsiPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('あなたのターン'));
  });

  it('phase indicator shows waiting when cpu turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<PrsiPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('待機中'));
  });

  it('pressing number key toggles card in play phase', async () => {
    renderWithProviders(<PrsiPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');
  });

  it('Enter key triggers play in play phase', async () => {
    renderWithProviders(<PrsiPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.keyDown(document, { key: '1' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);

    fireEvent.keyDown(document, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0));
  });

  it('handles action log visibility and API fetch', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<PrsiPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    vi.mocked(actionLogApi.prsi).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));

    await waitFor(() => expect(actionLogApi.prsi).toHaveBeenCalledTimes(1));
    expect(screen.getByText('棋譜')).toBeInTheDocument();
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<PrsiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    renderWithProviders(<PrsiPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('plays the card-place sound when a card is played', async () => {
    renderWithProviders(<PrsiPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    mockPlaySound.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '出す' }));

    // The central tap plays after the exec resolves, so await it.
    await waitFor(() => expect(mockPlaySound).toHaveBeenCalledWith('cardPlace'));
  });

  it('plays the shuffle sound when a card is drawn', async () => {
    renderWithProviders(<PrsiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '引く' })).toBeInTheDocument());

    mockPlaySound.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '引く' }));

    expect(mockPlaySound).toHaveBeenCalledWith('shuffle');
  });

  it('plays the shuffle sound when drawing via the stock pile', async () => {
    renderWithProviders(<PrsiPage />);
    const stock = await screen.findByTestId('prsi-stock');

    mockPlaySound.mockClear();
    fireEvent.click(stock);

    expect(mockPlaySound).toHaveBeenCalledWith('shuffle');
  });

  it('does not play the card-place sound when no single card is selected', async () => {
    renderWithProviders(<PrsiPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    // No card selected: keyboard Enter triggers play, which must stay silent.
    mockPlaySound.mockClear();
    fireEvent.keyDown(document, { key: 'Enter' });

    expect(mockPlaySound).not.toHaveBeenCalledWith('cardPlace');
  });

  it('plays the error buzz when an action fails', async () => {
    renderWithProviders(<PrsiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    mockPlaySound.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockPlaySound).toHaveBeenCalledWith('errorBuzz'));
  });
});
