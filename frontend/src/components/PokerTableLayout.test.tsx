import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { PokerTableLayout } from './PokerTableLayout';

// Mock useIsLargeDesktop hook
const mockUseIsLargeDesktop = vi.fn<() => boolean>();
vi.mock('../hooks/useCardDimensions', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../hooks/useCardDimensions')>();
  return {
    ...actual,
    useIsLargeDesktop: () => mockUseIsLargeDesktop(),
  };
});

describe('PokerTableLayout', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  const communityCards = <div data-testid="community-cards">Community</div>;
  const cpuPlayers = (
    <>
      <div data-testid="cpu-1">CPU 1</div>
      <div data-testid="cpu-2">CPU 2</div>
      <div data-testid="cpu-3">CPU 3</div>
    </>
  );

  it('renders community cards before CPU players on mobile', () => {
    mockUseIsLargeDesktop.mockReturnValue(false);
    const { container } = render(<PokerTableLayout communityCards={communityCards} cpuPlayers={cpuPlayers} />);
    const allText = container.textContent ?? '';
    expect(allText.indexOf('Community')).toBeLessThan(allText.indexOf('CPU 1'));
  });

  it('renders CPU players in grid before community cards on large desktop', () => {
    mockUseIsLargeDesktop.mockReturnValue(true);
    const { container } = render(<PokerTableLayout communityCards={communityCards} cpuPlayers={cpuPlayers} />);
    const cpuGrid = screen.getByTestId('cpu-1').parentElement;
    expect(cpuGrid?.className).toContain('grid');
    expect(cpuGrid?.className).toContain('grid-cols-3');
    const allText = container.textContent ?? '';
    expect(allText.indexOf('CPU 1')).toBeLessThan(allText.indexOf('Community'));
  });

  it.each([
    { isLargeDesktop: false, device: 'mobile' },
    { isLargeDesktop: true, device: 'desktop' },
  ])('applies data-tutorial attribute to CPU area on $device', ({ isLargeDesktop }) => {
    mockUseIsLargeDesktop.mockReturnValue(isLargeDesktop);
    render(<PokerTableLayout communityCards={communityCards} cpuPlayers={cpuPlayers} cpuAreaTutorial="he-cpu-area" />);
    const cpuArea = screen.getByTestId('cpu-1').closest('[data-tutorial]');
    expect(cpuArea?.getAttribute('data-tutorial')).toBe('he-cpu-area');
  });

  it.each([
    { isLargeDesktop: false, device: 'mobile' },
    { isLargeDesktop: true, device: 'desktop' },
  ])('applies data-tutorial attribute to community cards on $device', ({ isLargeDesktop }) => {
    mockUseIsLargeDesktop.mockReturnValue(isLargeDesktop);
    render(
      <PokerTableLayout
        communityCards={communityCards}
        cpuPlayers={cpuPlayers}
        communityCardsTutorial="he-community-cards"
      />,
    );
    const wrapper = screen.getByTestId('community-cards').closest('[data-tutorial]');
    expect(wrapper?.getAttribute('data-tutorial')).toBe('he-community-cards');
  });

  it.each([
    { isLargeDesktop: false, device: 'mobile' },
    { isLargeDesktop: true, device: 'desktop' },
  ])('renders all CPU players on $device', ({ isLargeDesktop }) => {
    mockUseIsLargeDesktop.mockReturnValue(isLargeDesktop);
    render(<PokerTableLayout communityCards={communityCards} cpuPlayers={cpuPlayers} />);
    expect(screen.getByTestId('cpu-1')).toBeInTheDocument();
    expect(screen.getByTestId('cpu-2')).toBeInTheDocument();
    expect(screen.getByTestId('cpu-3')).toBeInTheDocument();
  });
});
