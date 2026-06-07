import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { ChipBetInput } from './ChipBetInput';

describe('ChipBetInput', () => {
  it('renders label associated with input via htmlFor', () => {
    render(<ChipBetInput id="bet" label="Bet" value={50} onChange={() => {}} max={500} />);
    const input = screen.getByLabelText('Bet') as HTMLInputElement;
    expect(input.value).toBe('50');
    expect(input.min).toBe('10');
    expect(input.step).toBe('10');
    expect(input.max).toBe('500');
  });

  it('calls onChange with parsed numeric value', () => {
    const onChange = vi.fn();
    render(<ChipBetInput id="bet" label="Bet" value={10} onChange={onChange} max={500} />);
    fireEvent.change(screen.getByLabelText('Bet'), { target: { value: '120' } });
    expect(onChange).toHaveBeenCalledWith(120);
  });

  it('clamps values above max to max', () => {
    const onChange = vi.fn();
    render(<ChipBetInput id="bet" label="Bet" value={10} onChange={onChange} max={500} />);
    fireEvent.change(screen.getByLabelText('Bet'), { target: { value: '99999' } });
    expect(onChange).toHaveBeenCalledWith(500);
  });

  it('clamps values below min to min', () => {
    const onChange = vi.fn();
    render(<ChipBetInput id="bet" label="Bet" value={50} onChange={onChange} max={500} min={20} />);
    fireEvent.change(screen.getByLabelText('Bet'), { target: { value: '5' } });
    expect(onChange).toHaveBeenCalledWith(20);
  });

  it('treats empty input as min (Number("") === 0 is clamped up)', () => {
    const onChange = vi.fn();
    render(<ChipBetInput id="bet" label="Bet" value={50} onChange={onChange} max={500} />);
    fireEvent.change(screen.getByLabelText('Bet'), { target: { value: '' } });
    expect(onChange).toHaveBeenCalledWith(10);
  });

  it('respects custom min, step, and disabled props', () => {
    render(<ChipBetInput id="bet" label="Bet" value={5} onChange={() => {}} max={100} min={5} step={5} disabled />);
    const input = screen.getByLabelText('Bet') as HTMLInputElement;
    expect(input.min).toBe('5');
    expect(input.step).toBe('5');
    expect(input.disabled).toBe(true);
  });

  it('input meets the 44px minimum touch target (WCAG 2.5.5)', () => {
    // Issue #1510: chip bet number inputs are tapped during fast bet entry
    // (Baccarat / Three Card / Pai Gow); below 44px they're easy to miss.
    render(<ChipBetInput id="bet" label="Bet" value={50} onChange={() => {}} max={500} />);
    const input = screen.getByLabelText('Bet');
    expect(input.className).toContain('min-h-[44px]');
  });

  it('passes raw numeric values through without clamping when autoClamp is false', () => {
    const onChange = vi.fn();
    render(<ChipBetInput id="bet" label="Bet" value={20} onChange={onChange} max={50} autoClamp={false} />);
    fireEvent.change(screen.getByLabelText('Bet'), { target: { value: '80' } });
    expect(onChange).toHaveBeenCalledWith(80);
  });

  it('passes empty input through as 0 under autoClamp=false (parent decides validity)', () => {
    const onChange = vi.fn();
    render(<ChipBetInput id="bet" label="Bet" value={20} onChange={onChange} max={50} autoClamp={false} />);
    fireEvent.change(screen.getByLabelText('Bet'), { target: { value: '' } });
    expect(onChange).toHaveBeenCalledWith(0);
  });

  it('applies error styling and aria-invalid when invalid is true', () => {
    render(<ChipBetInput id="bet" label="Bet" value={5} onChange={() => {}} max={50} invalid />);
    const input = screen.getByLabelText('Bet');
    expect(input.className).toContain('bg-ds-surface');
    expect(input.className).toContain('border-ds-error');
    // Foreground is text-ds-text-primary (10.1:1 AAA on surface). Pairing
    // text-ds-error with bg-ds-surface only hits ~2.7:1 — fails AA — so the
    // error semantic comes from the coloured border, not the text colour.
    expect(input.className).toContain('text-ds-text-primary');
    expect(input).toHaveAttribute('aria-invalid', 'true');
  });

  it('wires aria-describedby through to the input element', () => {
    render(<ChipBetInput id="bet" label="Bet" value={5} onChange={() => {}} max={50} describedBy="bet-help" />);
    const input = screen.getByLabelText('Bet');
    expect(input).toHaveAttribute('aria-describedby', 'bet-help');
  });

  it('clamps to min when max is omitted (no-limit mode with autoClamp=true)', () => {
    const onChange = vi.fn();
    render(<ChipBetInput id="bet" label="Bet" value={20} onChange={onChange} />);
    fireEvent.change(screen.getByLabelText('Bet'), { target: { value: '5' } });
    expect(onChange).toHaveBeenCalledWith(10);
  });

  it('passes large values through unbounded when max is omitted', () => {
    const onChange = vi.fn();
    render(<ChipBetInput id="bet" label="Bet" value={20} onChange={onChange} />);
    fireEvent.change(screen.getByLabelText('Bet'), { target: { value: '999999999' } });
    expect(onChange).toHaveBeenCalledWith(999999999);
  });

  it('does not invoke onChange when raw input has no digits under autoClamp=false', () => {
    // The text/inputMode=numeric input drops all-non-digit input before
    // calling onChange to preserve the type=number "reject NaN" behavior.
    const onChange = vi.fn();
    const { container } = render(
      <ChipBetInput id="bet" label="Bet" value={20} onChange={onChange} max={50} autoClamp={false} />,
    );
    const input = container.querySelector<HTMLInputElement>('#bet');
    expect(input).not.toBeNull();
    if (!input) return;
    Object.defineProperty(input, 'value', { get: () => 'abc', configurable: true });
    fireEvent.change(input);
    expect(onChange).not.toHaveBeenCalled();
  });

  it('uses inputMode=numeric to surface the digit keyboard on mobile', () => {
    render(<ChipBetInput id="bet" label="Bet" value={50} onChange={() => {}} max={500} />);
    const input = screen.getByLabelText('Bet');
    expect(input).toHaveAttribute('type', 'text');
    expect(input).toHaveAttribute('inputmode', 'numeric');
    expect(input).toHaveAttribute('pattern', '[0-9]*');
  });

  it('strips non-digit characters from typed input before parsing', () => {
    const onChange = vi.fn();
    render(<ChipBetInput id="bet" label="Bet" value={50} onChange={onChange} max={500} />);
    fireEvent.change(screen.getByLabelText('Bet'), { target: { value: '1.2e5' } });
    // '1.2e5' → '125' after stripping; clamped under max=500 → 125.
    expect(onChange).toHaveBeenLastCalledWith(125);
  });

  it('blurs the input on wheel events to prevent accidental scroll changes', () => {
    render(<ChipBetInput id="bet" label="Bet" value={50} onChange={() => {}} max={500} />);
    const input = screen.getByLabelText('Bet') as HTMLInputElement;
    input.focus();
    expect(document.activeElement).toBe(input);
    fireEvent.wheel(input, { deltaY: 100 });
    expect(document.activeElement).not.toBe(input);
  });

  describe('steppers', () => {
    it('does not render steppers by default', () => {
      render(<ChipBetInput id="bet" label="Bet" value={50} onChange={() => {}} max={500} />);
      expect(screen.queryByRole('button', { name: /Bet \+/ })).not.toBeInTheDocument();
    });

    it('increments and decrements by step, clamped to [min, max]', () => {
      const onChange = vi.fn();
      render(
        <ChipBetInput id="bet" label="Bet" value={50} onChange={onChange} min={10} max={500} step={10} showSteppers />,
      );
      fireEvent.click(screen.getByRole('button', { name: 'Bet +10' }));
      expect(onChange).toHaveBeenLastCalledWith(60);
      fireEvent.click(screen.getByRole('button', { name: 'Bet −10' }));
      expect(onChange).toHaveBeenLastCalledWith(40);
    });

    it('disables the minus button at min and the plus button at max', () => {
      const { rerender } = render(
        <ChipBetInput id="bet" label="Bet" value={10} onChange={() => {}} min={10} max={100} step={10} showSteppers />,
      );
      expect(screen.getByRole('button', { name: 'Bet −10' })).toBeDisabled();
      expect(screen.getByRole('button', { name: 'Bet +10' })).toBeEnabled();
      rerender(
        <ChipBetInput id="bet" label="Bet" value={100} onChange={() => {}} min={10} max={100} step={10} showSteppers />,
      );
      expect(screen.getByRole('button', { name: 'Bet +10' })).toBeDisabled();
    });
  });
});
