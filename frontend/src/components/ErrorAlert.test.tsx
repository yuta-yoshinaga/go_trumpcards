import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { ErrorAlert } from './ErrorAlert';

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
    expect(screen.queryByRole('button', { name: 'Retry' })).not.toBeInTheDocument();
  });

  it('renders retry button and calls onRetry when clicked', () => {
    const onRetry = vi.fn();
    render(<ErrorAlert message="error" onRetry={onRetry} />);
    const btn = screen.getByRole('button', { name: 'Retry' });
    expect(btn).toBeInTheDocument();
    fireEvent.click(btn);
    expect(onRetry).toHaveBeenCalledTimes(1);
  });
});
