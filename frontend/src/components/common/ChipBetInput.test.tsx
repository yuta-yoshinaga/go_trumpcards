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

  it('respects custom min, step, and disabled props', () => {
    render(<ChipBetInput id="bet" label="Bet" value={5} onChange={() => {}} max={100} min={5} step={5} disabled />);
    const input = screen.getByLabelText('Bet') as HTMLInputElement;
    expect(input.min).toBe('5');
    expect(input.step).toBe('5');
    expect(input.disabled).toBe(true);
  });
});
