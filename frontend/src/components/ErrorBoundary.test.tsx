import { describe, expect, it, vi } from 'bun:test';
import { fireEvent, render, screen } from '@testing-library/react';
import type { ReactNode } from 'react';
import jaCommon from '../i18n/locales/ja/common.json';
import { ErrorBoundary } from './ErrorBoundary';

const errorTitle = jaCommon.label.errorBoundaryTitle;
const retryLabel = jaCommon.label.errorBoundaryRetry;

function ThrowingChild(): ReactNode {
  throw new Error('test error');
}

describe('ErrorBoundary', () => {
  it('renders children when no error occurs', () => {
    render(
      <ErrorBoundary>
        <p>child content</p>
      </ErrorBoundary>,
    );
    expect(screen.getByText('child content')).toBeInTheDocument();
  });

  it('shows fallback UI when a child throws during render', () => {
    vi.spyOn(console, 'error').mockImplementation(() => {});
    render(
      <ErrorBoundary>
        <ThrowingChild />
      </ErrorBoundary>,
    );
    expect(screen.getByText(errorTitle)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: retryLabel })).toBeInTheDocument();
    vi.restoreAllMocks();
  });

  it('resets and re-renders children when retry button is clicked', () => {
    let shouldThrow = true;

    function MaybeThrow(): ReactNode {
      if (shouldThrow) {
        throw new Error('test error');
      }
      return <p>recovered</p>;
    }

    vi.spyOn(console, 'error').mockImplementation(() => {});
    render(
      <ErrorBoundary>
        <MaybeThrow />
      </ErrorBoundary>,
    );
    expect(screen.getByText(errorTitle)).toBeInTheDocument();

    shouldThrow = false;
    fireEvent.click(screen.getByRole('button', { name: retryLabel }));

    expect(screen.getByText('recovered')).toBeInTheDocument();
    expect(screen.queryByText(errorTitle)).not.toBeInTheDocument();
    vi.restoreAllMocks();
  });
});
