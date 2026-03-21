import { motion } from 'framer-motion';
import { useReducedMotion } from '../../hooks/useReducedMotion';
import { CardImage } from '../CardImage';

interface AnimatedCardProps extends React.ComponentProps<typeof CardImage> {
  /** Stagger delay in seconds for deal animation. */
  dealDelay?: number;
  /** Whether this card is selected (lift + glow). */
  isSelected?: boolean;
  /** Shared layout animation ID. */
  layoutId?: string;
}

/** Renders an animated face-up playing card with deal and select animations. */
export function AnimatedCard({ dealDelay = 0, isSelected = false, layoutId, ...rest }: AnimatedCardProps) {
  const reduced = useReducedMotion();

  if (reduced) {
    return <CardImage {...rest} />;
  }

  return (
    <motion.div
      layoutId={layoutId}
      initial={{ opacity: 0, y: 30 }}
      animate={{
        opacity: 1,
        y: isSelected ? -8 : 0,
        scale: isSelected ? 1.02 : 1,
      }}
      transition={{
        type: 'spring',
        stiffness: 300,
        damping: 25,
        delay: dealDelay,
      }}
      style={{ display: 'inline-block' }}
      data-testid="animated-card"
    >
      <CardImage {...rest} />
    </motion.div>
  );
}
