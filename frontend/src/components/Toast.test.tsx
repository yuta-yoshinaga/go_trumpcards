import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { Toast } from './Toast';

describe('Toast', () => {
  it('renders children in a polite status region by default', () => {
    render(<Toast testId="t">hello</Toast>);
    const region = screen.getByTestId('t');
    expect(region).toHaveAttribute('role', 'status');
    expect(region).toHaveAttribute('aria-live', 'polite');
    expect(region).toHaveTextContent('hello');
  });

  it('supports an assertive live region', () => {
    render(
      <Toast testId="t" live="assertive">
        urgent
      </Toast>,
    );
    expect(screen.getByTestId('t')).toHaveAttribute('aria-live', 'assertive');
  });

  it('uses an opaque surface background, not an alpha-suffixed color', () => {
    render(<Toast testId="t">x</Toast>);
    const cls = screen.getByTestId('t').className;
    expect(cls).toContain('bg-ds-surface-elevated');
    expect(cls).not.toContain('bg-black/');
  });

  it('omits the close button when onDismiss is not provided', () => {
    render(<Toast testId="t">x</Toast>);
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
  });

  it('renders a 44x44px close button that calls onDismiss', () => {
    const onDismiss = vi.fn();
    render(
      <Toast testId="t" onDismiss={onDismiss}>
        x
      </Toast>,
    );
    const btn = screen.getByRole('button');
    expect(btn.className).toContain('min-h-[44px]');
    expect(btn.className).toContain('min-w-[44px]');
    fireEvent.click(btn);
    expect(onDismiss).toHaveBeenCalledTimes(1);
  });
});
