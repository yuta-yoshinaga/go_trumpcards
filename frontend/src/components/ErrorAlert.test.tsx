import { fireEvent, render, screen } from '@testing-library/react';
import i18n from 'i18next';
import { describe, expect, it, vi } from 'vitest';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { ErrorAlert } from './ErrorAlert';

const retryLabel = i18n.t('button.retry', { ns: 'common' });

describe('ErrorAlert', () => {
  it('renders nothing when message is null', () => {
    const { container } = render(<ErrorAlert message={null} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders the message when provided', () => {
    render(<ErrorAlert message={NETWORK_ERROR_MESSAGE()} />);
    expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument();
    expect(screen.getByRole('alert')).toBeInTheDocument();
  });

  it('does not render retry button when onRetry is not provided', () => {
    render(<ErrorAlert message="error" />);
    expect(screen.queryByRole('button', { name: retryLabel })).not.toBeInTheDocument();
  });

  it('renders retry button and calls onRetry when clicked', () => {
    const onRetry = vi.fn();
    render(<ErrorAlert message="error" onRetry={onRetry} />);
    const btn = screen.getByRole('button', { name: retryLabel });
    expect(btn).toBeInTheDocument();
    fireEvent.click(btn);
    expect(onRetry).toHaveBeenCalledTimes(1);
  });

  it('applies 44x44px touch target to retry button', () => {
    render(<ErrorAlert message="error" onRetry={() => {}} />);
    const btn = screen.getByRole('button', { name: retryLabel });
    expect(btn.className).toContain('min-h-[44px]');
    expect(btn.className).toContain('min-w-[44px]');
  });

  it('uses fully opaque error background (no alpha) for WCAG AA contrast', () => {
    render(<ErrorAlert message="error" />);
    const alert = screen.getByRole('alert');
    expect(alert.className).toContain('bg-ds-error');
    expect(alert.className).not.toContain('bg-ds-error/90');
  });

  it('renders the retry button on an opaque surface (no alpha) per DESIGN.md Opacity rule', () => {
    render(<ErrorAlert message="error" onRetry={() => {}} />);
    const btn = screen.getByRole('button', { name: retryLabel });
    expect(btn.className).toContain('bg-white');
    expect(btn.className).toContain('text-ds-error');
    expect(btn.className).not.toContain('bg-white/20');
    expect(btn.className).not.toContain('bg-white/30');
    // Hover deepens the red text so the cream hover surface stays WCAG AA (~5.4:1).
    expect(btn.className).toContain('hover:text-ds-error-hover');
  });
});
