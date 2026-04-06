import { motion } from 'framer-motion';
import { useRef } from 'react';
import { useReducedMotion } from '../../hooks/useReducedMotion';
import { dealSpring, hoverLift, selectLift } from '../../styles/motionPresets';
import { CardImage } from '../CardImage';

interface AnimatedCardProps extends React.ComponentProps<typeof CardImage> {
  /** Stagger delay in seconds for deal animation. */
  dealDelay?: number;
  /** Whether this card is selected (lift + glow). */
  isSelected?: boolean;
  /** Shared layout animation ID. */
  layoutId?: string;
  /** Additional class name for the motion wrapper div. */
  wrapperClassName?: string;
  /** Callback fired when the deal-in animation completes. */
  onDealComplete?: () => void;
}

/** Renders an animated face-up playing card with deal and select animations. */
export function AnimatedCard({
  dealDelay = 0,
  isSelected = false,
  layoutId,
  wrapperClassName,
  onDealComplete,
  ...rest
}: AnimatedCardProps) {
  const reduced = useReducedMotion();
  // Guard: onAnimationComplete fires for any animation (hover, selection), not just the initial deal.
  // Use a ref so the deal callback fires exactly once per component instance.
  const dealCalledRef = useRef(false);

  if (reduced) {
    return <CardImage {...rest} />;
  }

  return (
    <motion.div
      layoutId={layoutId}
      className={wrapperClassName}
      initial={{ opacity: 0, y: 30 }}
      animate={{
        opacity: 1,
        y: isSelected ? selectLift.y : 0,
        scale: isSelected ? selectLift.scale : 1,
      }}
      whileHover={hoverLift}
      transition={{
        ...dealSpring,
        delay: dealDelay,
      }}
      onAnimationComplete={() => {
        if (!dealCalledRef.current) {
          dealCalledRef.current = true;
          onDealComplete?.();
        }
      }}
      style={wrapperClassName ? undefined : { display: 'inline-block' }}
      data-testid="animated-card"
    >
      <CardImage {...rest} />
    </motion.div>
  );
}
