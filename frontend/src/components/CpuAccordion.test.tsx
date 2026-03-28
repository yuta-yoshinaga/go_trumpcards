import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { CpuAccordion } from './CpuAccordion';

const mockT = vi.fn((key: string) => key);
vi.mock('react-i18next', async () => ({
  ...(await vi.importActual('react-i18next')),
  useTranslation: () => ({ t: mockT }),
}));

vi.mock('../hooks/useCardDimensions', () => ({
  useIsMobile: vi.fn(() => false),
}));

// eslint-disable-next-line @typescript-eslint/consistent-type-imports
const { useIsMobile } = await import('../hooks/useCardDimensions');
const mockUseIsMobile = vi.mocked(useIsMobile);

describe('CpuAccordion', () => {
  beforeEach(() => {
    mockT.mockClear();
  });

  it('renders a details element with summary showing player count', () => {
    render(
      <CpuAccordion playerCount={3}>
        <div data-testid="child">CPU cards</div>
      </CpuAccordion>,
    );
    expect(screen.getByTestId('cpu-accordion')).toBeInTheDocument();
    expect(mockT).toHaveBeenCalledWith('label.cpuOpponents', { count: 3 });
    expect(screen.getByTestId('child')).toBeInTheDocument();
  });

  it('is open by default on desktop', () => {
    mockUseIsMobile.mockReturnValue(false);
    render(
      <CpuAccordion playerCount={2}>
        <div>content</div>
      </CpuAccordion>,
    );
    const details = screen.getByTestId('cpu-accordion');
    expect(details).toHaveAttribute('open');
  });

  it('is closed by default on mobile', () => {
    mockUseIsMobile.mockReturnValue(true);
    render(
      <CpuAccordion playerCount={3}>
        <div>content</div>
      </CpuAccordion>,
    );
    const details = screen.getByTestId('cpu-accordion');
    expect(details).not.toHaveAttribute('open');
  });

  it('forwards data-tutorial attribute', () => {
    render(
      <CpuAccordion playerCount={3} dataTutorial="oh-cpu-area">
        <div>content</div>
      </CpuAccordion>,
    );
    const details = screen.getByTestId('cpu-accordion');
    expect(details).toHaveAttribute('data-tutorial', 'oh-cpu-area');
  });
});
