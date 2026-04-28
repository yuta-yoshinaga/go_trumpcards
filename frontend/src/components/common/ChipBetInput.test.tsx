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
    expect(input.className).toContain('bg-ds-error/40');
    expect(input).toHaveAttribute('aria-invalid', 'true');
  });

  it('wires aria-describedby through to the input element', () => {
    render(<ChipBetInput id="bet" label="Bet" value={5} onChange={() => {}} max={50} describedBy="bet-help" />);
    const input = screen.getByLabelText('Bet');
    expect(input).toHaveAttribute('aria-describedby', 'bet-help');
  });
});
