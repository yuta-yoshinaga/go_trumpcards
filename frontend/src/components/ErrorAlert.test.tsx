import { describe, expect, it } from 'bun:test';
import { render, screen } from '@testing-library/react';
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
});
