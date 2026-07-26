import { motion } from 'framer-motion';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useReducedMotion } from '../../hooks/useReducedMotion';

/** Props for {@link WinCelebration}. */
export interface WinCelebrationProps {
  /** Whether to show the celebration animation. */
  show: boolean;
  /** Delay in ms before particles appear. Default: 400 (the "tension beat"). */
  delayMs?: number;
  /** Callback fired when particles start animating (use for sound sync). */
  onCelebrate?: () => void;
}

const PARTICLE_COUNT = 15;

// Design system tokens: accent gold (60%), success green (20%), warm ivory (20%)
// Fallbacks match the current CSS variable values so particles render correctly
// even if getComputedStyle returns empty (e.g., in tests or SSR).
const FALLBACK_ACCENT = '#D4A853';
const FALLBACK_SUCCESS = '#4CAF7D';
const FALLBACK_IVORY = '#E8E0D4';

/** Reads design-system CSS variables at runtime so theme changes propagate to particles. */
function getTokenColors(): string[] {
  const style = typeof document !== 'undefined' ? getComputedStyle(document.documentElement) : null;
  const accent = style?.getPropertyValue('--color-ds-accent').trim() || FALLBACK_ACCENT;
  const success = style?.getPropertyValue('--color-ds-success').trim() || FALLBACK_SUCCESS;
  const ivory = style?.getPropertyValue('--color-ds-text-primary').trim() || FALLBACK_IVORY;
  // 60% accent, 20% success, 20% ivory
  return [
    accent,
    accent,
    accent,
    accent,
    accent,
    accent,
    accent,
    accent,
    accent,
    success,
    success,
    success,
    ivory,
    ivory,
    ivory,
  ];
}

interface Particle {
  id: number;
  x: number;
  y: number;
  color: string;
  size: number;
  delay: number;
}

function generateParticles(): Particle[] {
  const colors = getTokenColors();
  return Array.from({ length: PARTICLE_COUNT }, (_, i) => ({
    id: i,
    x: (Math.random() - 0.5) * 300,
    y: -(Math.random() * 200 + 50),
    color: colors[i % colors.length],
    size: Math.random() * 8 + 4,
    delay: Math.random() * 0.3,
  }));
}

/** Renders a celebration animation on win with configurable delay and sound callback. */
export function WinCelebration({ show, delayMs = 400, onCelebrate }: WinCelebrationProps) {
  const reduced = useReducedMotion();
  const [particles, setParticles] = useState<Particle[]>([]);
  const [active, setActive] = useState(false);
  const { t } = useTranslation('common');

  useEffect(() => {
    if (!show) {
      setActive(false);
      setParticles([]);
      return;
    }

    if (reduced) {
      setActive(true);
      onCelebrate?.();
      return;
    }

    const timer = setTimeout(() => {
      setParticles(generateParticles());
      setActive(true);
      onCelebrate?.();
    }, delayMs);

    return () => clearTimeout(timer);
  }, [show, delayMs, reduced, onCelebrate]);

  if (!show) return null;

  // Reduced motion: show text banner instead of particles
  if (reduced) {
    return (
      <div
        role="status"
        aria-live="polite"
        className="fixed inset-x-0 top-0 z-50 flex justify-center pointer-events-none"
        data-testid="win-celebration"
      >
        <div className="mt-4 px-4 py-2 rounded-md bg-ds-accent text-ds-text-on-accent text-sm font-medium">
          {t('result.win', 'You won!')}
        </div>
      </div>
    );
  }

  if (!active) return null;

  return (
    <>
      <div
        className="pointer-events-none fixed inset-0 flex items-center justify-center overflow-hidden z-50"
        aria-hidden="true"
        data-testid="win-celebration"
      >
        {particles.map((p) => (
          <motion.div
            key={p.id}
            initial={{ opacity: 1, x: 0, y: 0, scale: 0 }}
            animate={{ opacity: 0, x: p.x, y: p.y, scale: 1 }}
            transition={{ duration: 1, delay: p.delay, ease: 'easeOut' }}
            style={{
              position: 'absolute',
              width: p.size,
              height: p.size,
              borderRadius: '50%',
              backgroundColor: p.color,
            }}
          />
        ))}
      </div>
      <div role="status" aria-live="polite" className="sr-only">
        {t('result.win', 'You won!')}
      </div>
    </>
  );
}
