import { AnimatePresence, motion } from 'framer-motion';
import type { ReactNode } from 'react';
import { useReducedMotion } from '../../hooks/useReducedMotion';

interface PhaseTransitionProps {
  /** Unique key identifying the current phase. */
  phaseKey: string | number;
  children: ReactNode;
}

/** Wraps children in an AnimatePresence with fade/slide transition between phases. */
export function PhaseTransition({ phaseKey, children }: PhaseTransitionProps) {
  const reduced = useReducedMotion();

  if (reduced) {
    return <>{children}</>;
  }

  return (
    <AnimatePresence mode="wait">
      <motion.div
        key={phaseKey}
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        exit={{ opacity: 0, y: -10 }}
        transition={{ duration: 0.2 }}
        data-testid="phase-transition"
      >
        {children}
      </motion.div>
    </AnimatePresence>
  );
}
