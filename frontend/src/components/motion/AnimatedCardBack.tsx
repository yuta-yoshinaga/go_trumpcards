import { motion } from 'framer-motion';
import { useReducedMotion } from '../../hooks/useReducedMotion';
import { CardBack } from '../CardImage';

interface AnimatedCardBackProps {
  width?: number;
  style?: React.CSSProperties;
  className?: string;
  onClick?: () => void;
  ariaLabel?: string;
  /** Stagger delay in seconds for deal animation. */
  dealDelay?: number;
  /** Shared layout animation ID. */
  layoutId?: string;
}

/** Renders an animated face-down card back with deal and optional flip animation. */
export function AnimatedCardBack({
  width,
  style,
  className,
  onClick,
  ariaLabel,
  dealDelay = 0,
  layoutId,
}: AnimatedCardBackProps) {
  const reduced = useReducedMotion();

  if (reduced) {
    return <CardBack width={width} style={style} className={className} onClick={onClick} ariaLabel={ariaLabel} />;
  }

  return (
    <motion.div
      layoutId={layoutId}
      initial={{ opacity: 0, rotateY: 90 }}
      animate={{ opacity: 1, rotateY: 0 }}
      transition={{
        type: 'spring',
        stiffness: 300,
        damping: 25,
        delay: dealDelay,
      }}
      style={{ display: 'inline-block' }}
      data-testid="animated-card-back"
    >
      <CardBack width={width} style={style} className={className} onClick={onClick} ariaLabel={ariaLabel} />
    </motion.div>
  );
}
