import { motion } from 'framer-motion';
import { useReducedMotion } from '../../hooks/useReducedMotion';
import { flipSpring } from '../../styles/motionPresets';
import { CardBack } from '../CardImage';

interface AnimatedCardBackProps extends React.ComponentProps<typeof CardBack> {
  /** Stagger delay in seconds for deal animation. */
  dealDelay?: number;
  /** Shared layout animation ID. */
  layoutId?: string;
  /** Callback fired when the flip-in animation completes. */
  onFlipComplete?: () => void;
}

/** Renders an animated face-down card back with deal and optional flip animation. */
export function AnimatedCardBack({ dealDelay = 0, layoutId, onFlipComplete, ...rest }: AnimatedCardBackProps) {
  const reduced = useReducedMotion();

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
      onAnimationComplete={onFlipComplete}
      style={{ display: 'inline-block', perspective: 1000 }}
      data-testid="animated-card-back"
    >
      <CardBack {...rest} />
    </motion.div>
  );
}
