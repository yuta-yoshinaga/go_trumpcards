import { motion } from 'framer-motion';
import { useRef } from 'react';
import { useReducedMotion } from '../../hooks/useReducedMotion';
import { useOptionalSound } from '../../providers/SoundProvider';
import { flipSpring } from '../../styles/motionPresets';
import { CardBack } from '../CardImage';

/** Props for {@link AnimatedCardBack}. */
export interface AnimatedCardBackProps extends React.ComponentProps<typeof CardBack> {
  /** Stagger delay in seconds for deal animation. */
  dealDelay?: number;
  /** Shared layout animation ID. */
  layoutId?: string;
  /** Suppress the default 'cardFlip' SFX. */
  silent?: boolean;
  /**
   * Optional callback fired after the flip-in animation completes,
   * in addition to the default SFX. Pass only when extra side-effects
   * need to fire alongside the default sound.
   */
  onFlipComplete?: () => void;
}

/** Renders an animated face-down card back with deal and optional flip animation. */
export function AnimatedCardBack({
  dealDelay = 0,
  layoutId,
  silent = false,
  onFlipComplete,
  ...rest
}: AnimatedCardBackProps) {
  const reduced = useReducedMotion();
  const sound = useOptionalSound();
  // Guard: onAnimationComplete fires for any animation, not just the initial flip.
  // Use a ref so the flip callback fires exactly once per component instance.
  const flipCalledRef = useRef(false);

  if (reduced) {
    return <CardBack {...rest} />;
  }

  return (
    <motion.div
      layoutId={layoutId}
      initial={{ opacity: 0, rotateY: 90 }}
      animate={{ opacity: 1, rotateY: 0 }}
      transition={{
        ...flipSpring,
        delay: dealDelay,
      }}
      onAnimationComplete={() => {
        if (!flipCalledRef.current) {
          flipCalledRef.current = true;
          if (!silent) sound?.playSound('cardFlip');
          onFlipComplete?.();
        }
      }}
      style={{ display: 'inline-block', perspective: 1000 }}
      data-testid="animated-card-back"
    >
      <CardBack {...rest} />
    </motion.div>
  );
}
