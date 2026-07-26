import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { SoundProvider, useSound } from '../providers/SoundProvider';
import { BettingControls } from './BettingControls';

// Track plays per sound file so the chip-sound tests can assert WHICH sound
// fired (the global setup.ts mock can't distinguish Howl instances).
const { playCalls } = vi.hoisted(() => ({ playCalls: [] as string[] }));
vi.mock('howler', () => ({
  Howl: class MockHowl {
    private src: string;
    constructor(opts: { src: string[] }) {
      this.src = opts.src[0];
    }
    play() {
      playCalls.push(this.src);
      return 1;
    }
    volume() {}
    rate() {}
  },
  Howler: { ctx: { state: 'running' } },
}));

function makeProps(overrides: Partial<Parameters<typeof BettingControls>[0]> = {}) {
  return {
    inputId: 'bet-input',
    betAmount: 20,
    onBetAmountChange: vi.fn(),
    minRaise: 10,
    hasOutstandingBet: false,
    loading: false,
    onCall: vi.fn(),
    onRaise: vi.fn(),
    onBet: vi.fn(),
    onCheck: vi.fn(),
    onFold: vi.fn(),
    onAllIn: vi.fn(),
    ...overrides,
  };
}

describe('BettingControls', () => {
  it('renders bet/check buttons when no outstanding bet', () => {
    render(<BettingControls {...makeProps()} />);
    expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'コール' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'レイズ' })).not.toBeInTheDocument();
  });

  it('renders call/raise buttons when there is an outstanding bet', () => {
    render(<BettingControls {...makeProps({ hasOutstandingBet: true })} />);
    expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'レイズ' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'チェック' })).not.toBeInTheDocument();
  });

  it('renders the call/raise key-hint line and per-button aria-keyshortcuts (outstanding bet)', () => {
    render(<BettingControls {...makeProps({ hasOutstandingBet: true })} />);
    const hints = screen.getByTestId('betting-key-hints');
    expect(hints).toHaveTextContent('C: コール');
    expect(hints).not.toHaveTextContent('K: チェック'); // check isn't available now
    expect(screen.getByRole('button', { name: 'コール' })).toHaveAttribute('aria-keyshortcuts', 'c');
    expect(screen.getByRole('button', { name: 'レイズ' })).toHaveAttribute('aria-keyshortcuts', 'r');
    expect(screen.getByRole('button', { name: 'フォールド' })).toHaveAttribute('aria-keyshortcuts', 'f');
    expect(screen.getByRole('button', { name: 'オールイン' })).toHaveAttribute('aria-keyshortcuts', 'a');
    // Desktop: each button also carries a visible <kbd> key chip (aria-hidden, so the name is unchanged).
    expect(screen.getByRole('button', { name: 'コール' }).querySelector('kbd')).toHaveTextContent('C');
    expect(screen.getByRole('button', { name: 'フォールド' }).querySelector('kbd')).toHaveTextContent('F');
  });

  it('renders the check/bet key-hint line and aria-keyshortcuts when there is no outstanding bet', () => {
    render(<BettingControls {...makeProps()} />);
    const hints = screen.getByTestId('betting-key-hints');
    expect(hints).toHaveTextContent('K: チェック');
    expect(hints).not.toHaveTextContent('C: コール'); // call isn't available now
    expect(screen.getByRole('button', { name: 'ベット' })).toHaveAttribute('aria-keyshortcuts', 'r');
    expect(screen.getByRole('button', { name: 'チェック' })).toHaveAttribute('aria-keyshortcuts', 'k');
  });

  it('always renders fold and all-in buttons', () => {
    render(<BettingControls {...makeProps()} />);
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument();
  });

  it('applies poker-themed styles to action buttons', () => {
    render(<BettingControls {...makeProps({ hasOutstandingBet: true })} />);
    expect(screen.getByRole('button', { name: 'コール' }).className).toContain('bg-poker-call');
    expect(screen.getByRole('button', { name: 'レイズ' }).className).toContain('bg-poker-raise');
    expect(screen.getByRole('button', { name: 'フォールド' }).className).toContain('bg-poker-fold');
    expect(screen.getByRole('button', { name: 'オールイン' }).className).toContain('bg-poker-allin');
  });

  it('applies poker-themed styles to bet/check buttons', () => {
    render(<BettingControls {...makeProps()} />);
    expect(screen.getByRole('button', { name: 'ベット' }).className).toContain('bg-poker-raise');
    expect(screen.getByRole('button', { name: 'チェック' }).className).toContain('bg-poker-call');
  });

  it('disables buttons when loading', () => {
    render(<BettingControls {...makeProps({ loading: true })} />);
    for (const btn of screen.getAllByRole('button')) {
      expect(btn).toBeDisabled();
    }
  });

  it('calls onBetAmountChange when input changes', () => {
    const onBetAmountChange = vi.fn();
    render(<BettingControls {...makeProps({ onBetAmountChange })} />);
    const input = screen.getByLabelText('ベット額:');
    fireEvent.change(input, { target: { value: '50' } });
    expect(onBetAmountChange).toHaveBeenCalledWith(50);
  });

  it('calls onBet when bet button clicked', () => {
    const onBet = vi.fn();
    render(<BettingControls {...makeProps({ onBet })} />);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    expect(onBet).toHaveBeenCalled();
  });

  it('calls onCheck when check button clicked', () => {
    const onCheck = vi.fn();
    render(<BettingControls {...makeProps({ onCheck })} />);
    fireEvent.click(screen.getByRole('button', { name: 'チェック' }));
    expect(onCheck).toHaveBeenCalled();
  });

  it('calls onCall when call button clicked', () => {
    const onCall = vi.fn();
    render(<BettingControls {...makeProps({ hasOutstandingBet: true, onCall })} />);
    fireEvent.click(screen.getByRole('button', { name: 'コール' }));
    expect(onCall).toHaveBeenCalled();
  });

  it('calls onRaise when raise button clicked', () => {
    const onRaise = vi.fn();
    render(<BettingControls {...makeProps({ hasOutstandingBet: true, onRaise })} />);
    fireEvent.click(screen.getByRole('button', { name: 'レイズ' }));
    expect(onRaise).toHaveBeenCalled();
  });

  it('calls onFold when fold button clicked', () => {
    const onFold = vi.fn();
    render(<BettingControls {...makeProps({ onFold })} />);
    fireEvent.click(screen.getByRole('button', { name: 'フォールド' }));
    expect(onFold).toHaveBeenCalled();
  });

  it('calls onAllIn when all-in button clicked', () => {
    const onAllIn = vi.fn();
    render(<BettingControls {...makeProps({ onAllIn })} />);
    fireEvent.click(screen.getByRole('button', { name: 'オールイン' }));
    expect(onAllIn).toHaveBeenCalled();
  });

  it('renders input with correct min attribute', () => {
    render(<BettingControls {...makeProps({ minRaise: 25 })} />);
    const input = screen.getByLabelText('ベット額:');
    expect(input).toHaveAttribute('min', '25');
  });

  it('sets max attribute on input when maxBetAmount is positive', () => {
    render(<BettingControls {...makeProps({ maxBetAmount: 100 })} />);
    const input = screen.getByLabelText('ベット額:');
    expect(input).toHaveAttribute('max', '100');
  });

  it('does not set max attribute when maxBetAmount is 0', () => {
    render(<BettingControls {...makeProps({ maxBetAmount: 0 })} />);
    const input = screen.getByLabelText('ベット額:');
    expect(input).not.toHaveAttribute('max');
  });

  it('does not set max attribute when maxBetAmount is undefined', () => {
    render(<BettingControls {...makeProps()} />);
    const input = screen.getByLabelText('ベット額:');
    expect(input).not.toHaveAttribute('max');
  });

  it('passes through value exceeding maxBetAmount without clamping', () => {
    const onBetAmountChange = vi.fn();
    render(<BettingControls {...makeProps({ maxBetAmount: 50, onBetAmountChange })} />);
    const input = screen.getByLabelText('ベット額:');
    fireEvent.change(input, { target: { value: '80' } });
    expect(onBetAmountChange).toHaveBeenCalledWith(80);
  });

  it('shows range hint when value exceeds maxBetAmount', () => {
    render(<BettingControls {...makeProps({ betAmount: 80, maxBetAmount: 50 })} />);
    expect(screen.getByRole('alert')).toBeInTheDocument();
    expect(screen.getByText(/10 〜 50/)).toBeInTheDocument();
  });

  it('shows range hint when value is below minRaise', () => {
    render(<BettingControls {...makeProps({ betAmount: 5, minRaise: 10, maxBetAmount: 100 })} />);
    expect(screen.getByRole('alert')).toBeInTheDocument();
    expect(screen.getByText(/10 〜 100/)).toBeInTheDocument();
  });

  it('does not show range hint when value is within range', () => {
    render(<BettingControls {...makeProps({ betAmount: 30, minRaise: 10, maxBetAmount: 100 })} />);
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('sets aria-invalid on input when out of range', () => {
    render(<BettingControls {...makeProps({ betAmount: 80, maxBetAmount: 50 })} />);
    const input = screen.getByLabelText('ベット額:');
    expect(input).toHaveAttribute('aria-invalid', 'true');
  });

  it('applies error styling to input when out of range', () => {
    render(<BettingControls {...makeProps({ betAmount: 80, maxBetAmount: 50 })} />);
    const input = screen.getByLabelText('ベット額:');
    // Background stays on surface; text stays on text-ds-text-primary (10.1:1 AAA).
    // Pairing text-ds-error with bg-ds-surface only hits ~2.7:1 — fails AA — so
    // the error semantic comes from the coloured border, not the text colour.
    // See `fixup(a11y): keep error/info badge text on text-ds-text-primary for AAA`.
    expect(input.className).toContain('bg-ds-surface');
    expect(input.className).toContain('border-ds-error');
    expect(input.className).toContain('text-ds-text-primary');
  });

  it('does not hardcode bg-white or text-ds-text-inverse on the input when in range', () => {
    render(<BettingControls {...makeProps({ betAmount: 30, minRaise: 10, maxBetAmount: 100 })} />);
    const input = screen.getByLabelText('ベット額:');
    expect(input.className).not.toContain('bg-white');
    expect(input.className).not.toContain('text-ds-text-inverse');
  });

  it('disables bet/raise button when value is out of range', () => {
    render(<BettingControls {...makeProps({ betAmount: 5 })} />);
    expect(screen.getByRole('button', { name: 'ベット' })).toBeDisabled();
  });

  it('disables raise button when value is out of range with outstanding bet', () => {
    render(<BettingControls {...makeProps({ betAmount: 5, hasOutstandingBet: true })} />);
    expect(screen.getByRole('button', { name: 'レイズ' })).toBeDisabled();
  });

  it('treats NaN betAmount as out of range', () => {
    render(<BettingControls {...makeProps({ betAmount: NaN })} />);
    expect(screen.getByRole('button', { name: 'ベット' })).toBeDisabled();
    expect(screen.getByLabelText('ベット額:')).toHaveAttribute('aria-invalid', 'true');
  });

  it('does not disable check/call/fold/all-in when out of range', () => {
    render(<BettingControls {...makeProps({ betAmount: 5, hasOutstandingBet: true })} />);
    expect(screen.getByRole('button', { name: 'コール' })).not.toBeDisabled();
    expect(screen.getByRole('button', { name: 'フォールド' })).not.toBeDisabled();
    expect(screen.getByRole('button', { name: 'オールイン' })).not.toBeDisabled();
  });

  describe('preset buttons', () => {
    it('does not render preset buttons when potSize is 0', () => {
      render(<BettingControls {...makeProps({ potSize: 0 })} />);
      expect(screen.queryByRole('button', { name: '1/2 Pot' })).not.toBeInTheDocument();
      expect(screen.queryByRole('button', { name: 'Pot' })).not.toBeInTheDocument();
      expect(screen.queryByRole('button', { name: 'Max' })).not.toBeInTheDocument();
    });

    it('does not render preset buttons when potSize is undefined', () => {
      render(<BettingControls {...makeProps()} />);
      expect(screen.queryByRole('button', { name: '1/2 Pot' })).not.toBeInTheDocument();
      expect(screen.queryByRole('button', { name: 'Pot' })).not.toBeInTheDocument();
      expect(screen.queryByRole('button', { name: 'Max' })).not.toBeInTheDocument();
    });

    it('renders 1/2 Pot and Pot preset buttons when potSize is positive', () => {
      render(<BettingControls {...makeProps({ potSize: 100 })} />);
      expect(screen.getByRole('button', { name: '1/2 Pot' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'Pot' })).toBeInTheDocument();
    });

    it('renders Max preset button when maxBetAmount is positive', () => {
      render(<BettingControls {...makeProps({ potSize: 100, maxBetAmount: 500 })} />);
      expect(screen.getByRole('button', { name: 'Max' })).toBeInTheDocument();
    });

    it('does not render Max preset when maxBetAmount is missing', () => {
      render(<BettingControls {...makeProps({ potSize: 100 })} />);
      expect(screen.queryByRole('button', { name: 'Max' })).not.toBeInTheDocument();
    });

    it('sets bet amount to half of pot when 1/2 Pot clicked', () => {
      const onBetAmountChange = vi.fn();
      render(<BettingControls {...makeProps({ potSize: 200, minRaise: 10, maxBetAmount: 1000, onBetAmountChange })} />);
      fireEvent.click(screen.getByRole('button', { name: '1/2 Pot' }));
      expect(onBetAmountChange).toHaveBeenCalledWith(100);
    });

    it('sets bet amount to pot when Pot clicked', () => {
      const onBetAmountChange = vi.fn();
      render(<BettingControls {...makeProps({ potSize: 200, minRaise: 10, maxBetAmount: 1000, onBetAmountChange })} />);
      fireEvent.click(screen.getByRole('button', { name: 'Pot' }));
      expect(onBetAmountChange).toHaveBeenCalledWith(200);
    });

    it('sets bet amount to maxBetAmount when Max clicked', () => {
      const onBetAmountChange = vi.fn();
      render(<BettingControls {...makeProps({ potSize: 200, minRaise: 10, maxBetAmount: 350, onBetAmountChange })} />);
      fireEvent.click(screen.getByRole('button', { name: 'Max' }));
      expect(onBetAmountChange).toHaveBeenCalledWith(350);
    });

    it('clamps 1/2 Pot preset to minRaise when half-pot is below minRaise', () => {
      const onBetAmountChange = vi.fn();
      render(<BettingControls {...makeProps({ potSize: 10, minRaise: 20, maxBetAmount: 1000, onBetAmountChange })} />);
      fireEvent.click(screen.getByRole('button', { name: '1/2 Pot' }));
      expect(onBetAmountChange).toHaveBeenCalledWith(20);
    });

    it('clamps Pot preset to maxBetAmount when pot exceeds max', () => {
      const onBetAmountChange = vi.fn();
      render(<BettingControls {...makeProps({ potSize: 1000, minRaise: 10, maxBetAmount: 300, onBetAmountChange })} />);
      fireEvent.click(screen.getByRole('button', { name: 'Pot' }));
      expect(onBetAmountChange).toHaveBeenCalledWith(300);
    });

    it('falls back to minRaise when preset computation yields NaN', () => {
      const onBetAmountChange = vi.fn();
      render(
        <BettingControls
          {...makeProps({ potSize: Number.NaN, minRaise: 10, maxBetAmount: 1000, onBetAmountChange })}
        />,
      );
      // potSize NaN should not render preset buttons (showPresets = pot > 0 is false for NaN)
      expect(screen.queryByRole('button', { name: '1/2 Pot' })).not.toBeInTheDocument();
    });

    it('floors Max preset when maxBetAmount has a fractional value', () => {
      const onBetAmountChange = vi.fn();
      render(
        <BettingControls {...makeProps({ potSize: 100, minRaise: 10, maxBetAmount: 350.7, onBetAmountChange })} />,
      );
      fireEvent.click(screen.getByRole('button', { name: 'Max' }));
      expect(onBetAmountChange).toHaveBeenCalledWith(350);
    });

    it('disables preset buttons when loading', () => {
      render(<BettingControls {...makeProps({ potSize: 100, maxBetAmount: 500, loading: true })} />);
      expect(screen.getByRole('button', { name: '1/2 Pot' })).toBeDisabled();
      expect(screen.getByRole('button', { name: 'Pot' })).toBeDisabled();
      expect(screen.getByRole('button', { name: 'Max' })).toBeDisabled();
    });
  });

  describe('central chip sound + exec claim', () => {
    /** Captures the live sound context so tests can inspect the claim token. */
    function Probe({ onReady }: { onReady: (ctx: ReturnType<typeof useSound>) => void }) {
      onReady(useSound());
      return null;
    }

    function renderWithSound(props: Parameters<typeof BettingControls>[0]) {
      let ctx!: ReturnType<typeof useSound>;
      render(
        <SoundProvider>
          <Probe
            onReady={(c) => {
              ctx = c;
            }}
          />
          <BettingControls {...props} />
        </SoundProvider>,
      );
      return () => ctx;
    }

    it.each([
      ['ベット', {}],
      ['コール', { hasOutstandingBet: true }],
      ['レイズ', { hasOutstandingBet: true }],
      ['オールイン', {}],
    ])('plays chipClick and claims the exec sound on %s', (label, overrides) => {
      playCalls.length = 0;
      const getCtx = renderWithSound(makeProps(overrides));
      fireEvent.click(screen.getByRole('button', { name: label }));
      expect(playCalls).toContain('/sounds/chip-click.ogg');
      expect(getCtx().consumeExecClaim()).toBe(true);
    });

    it.each([
      ['チェック', {}],
      ['フォールド', {}],
    ])('does not play chipClick or claim on %s (no chips move)', (label, overrides) => {
      playCalls.length = 0;
      const getCtx = renderWithSound(makeProps(overrides));
      fireEvent.click(screen.getByRole('button', { name: label }));
      expect(playCalls).not.toContain('/sounds/chip-click.ogg');
      expect(getCtx().consumeExecClaim()).toBe(false);
    });

    it('still invokes the page callback after the chip sound', () => {
      const onBet = vi.fn();
      renderWithSound(makeProps({ onBet }));
      fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
      expect(onBet).toHaveBeenCalledTimes(1);
    });

    it('renders and fires callbacks without a SoundProvider (providerless pages)', () => {
      const onBet = vi.fn();
      render(<BettingControls {...makeProps({ onBet })} />);
      fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
      expect(onBet).toHaveBeenCalledTimes(1);
    });
  });
});
