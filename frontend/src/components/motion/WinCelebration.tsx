import { motion } from 'framer-motion';
import { useMemo } from 'react';
import { useReducedMotion } from '../../hooks/useReducedMotion';

interface WinCelebrationProps {
  /** Whether to show the celebration animation. */
  show: boolean;
}

const PARTICLE_COUNT = 15;
const COLORS = ['#FFD700', '#FF6B6B', '#4ECDC4', '#45B7D1', '#96CEB4', '#FFEAA7'];

interface Particle {
  x: number;
  y: number;
  color: string;
  size: number;
  delay: number;
}

function generateParticles(): Particle[] {
  return Array.from({ length: PARTICLE_COUNT }, () => ({
    x: (Math.random() - 0.5) * 300,
    y: -(Math.random() * 200 + 50),
    color: COLORS[Math.floor(Math.random() * COLORS.length)],
    size: Math.random() * 8 + 4,
    delay: Math.random() * 0.3,
  }));
}

/** Renders a lightweight particle burst celebration animation on win. */
export function WinCelebration({ show }: WinCelebrationProps) {
  const reduced = useReducedMotion();
  const particles = useMemo(() => generateParticles(), []);

  if (!show || reduced) return null;

  return (
    <div
      className="pointer-events-none fixed inset-0 flex items-center justify-center overflow-hidden z-50"
      aria-hidden="true"
      data-testid="win-celebration"
    >
      {particles.map((p, i) => (
        <motion.div
          key={i}
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
  );
}
