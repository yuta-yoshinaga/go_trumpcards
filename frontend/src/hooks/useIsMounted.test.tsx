import { render, screen } from '@testing-library/react';
import { StrictMode, useEffect, useState } from 'react';
import { describe, expect, it } from 'vitest';
import { useIsMounted } from './useIsMounted';

/** Renders the flag's value, re-read on every render. */
function Harness({ onRead }: { onRead: (isMounted: () => boolean) => void }) {
  const isMounted = useIsMounted();
  const [, force] = useState(0);
  useEffect(() => {
    // Re-render once after mount so the assertion reads the flag after all effects
    // (including StrictMode's simulated cleanup/remount) have settled.
    force(1);
  }, []);
  onRead(isMounted);
  return <div data-testid="mounted">{isMounted() ? 'yes' : 'no'}</div>;
}

describe('useIsMounted', () => {
  it('reports mounted while mounted', () => {
    render(<Harness onRead={() => {}} />);
    expect(screen.getByTestId('mounted')).toHaveTextContent('yes');
  });

  // The regression that matters: StrictMode runs mount → cleanup → remount in dev, so
  // a cleanup-only effect would latch the flag false on a component that is genuinely
  // mounted and silently skip every guarded state write from then on. See #4446.
  it('stays mounted through a StrictMode mount/cleanup/remount cycle', () => {
    let latest: (() => boolean) | undefined;
    render(
      <StrictMode>
        <Harness
          onRead={(isMounted) => {
            latest = isMounted;
          }}
        />
      </StrictMode>,
    );
    expect(latest?.()).toBe(true);
    expect(screen.getByTestId('mounted')).toHaveTextContent('yes');
  });

  it('reports unmounted after unmount', () => {
    let latest: (() => boolean) | undefined;
    const { unmount } = render(
      <Harness
        onRead={(isMounted) => {
          latest = isMounted;
        }}
      />,
    );
    expect(latest?.()).toBe(true);
    unmount();
    expect(latest?.()).toBe(false);
  });
});
