import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { StalemateEscapeButton } from './StalemateEscapeButton';

describe('StalemateEscapeButton', () => {
  it('renders nothing when undoToEscape is 0', () => {
    const { container } = render(<StalemateEscapeButton undoToEscape={0} onEscape={vi.fn()} />);
    expect(container.innerHTML).toBe('');
  });

  it('renders button with undo count when undoToEscape > 0', () => {
    render(<StalemateEscapeButton undoToEscape={5} onEscape={vi.fn()} />);
    expect(screen.getByTestId('stalemate-escape-button')).toHaveTextContent('脱出する (5手戻る)');
  });

  it('calls onEscape with undoToEscape count on click', () => {
    const onEscape = vi.fn();
    render(<StalemateEscapeButton undoToEscape={3} onEscape={onEscape} />);
    fireEvent.click(screen.getByTestId('stalemate-escape-button'));
    expect(onEscape).toHaveBeenCalledWith(3);
  });

  it('button is disabled when disabled prop is true', () => {
    render(<StalemateEscapeButton undoToEscape={5} onEscape={vi.fn()} disabled />);
    expect(screen.getByTestId('stalemate-escape-button')).toBeDisabled();
  });

  it('button has animate-pulse class', () => {
    render(<StalemateEscapeButton undoToEscape={5} onEscape={vi.fn()} />);
    expect(screen.getByTestId('stalemate-escape-button')).toHaveClass('animate-pulse');
  });
});
