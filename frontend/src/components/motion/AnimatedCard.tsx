import { motion } from 'framer-motion';
import { useReducedMotion } from '../../hooks/useReducedMotion';
import type { Card } from '../../types/card';
import { CardImage } from '../CardImage';

interface AnimatedCardProps {
  card: Card;
  width?: number;
  style?: React.CSSProperties;
  className?: string;
  draggable?: boolean;
  onDragStart?: (e: React.DragEvent) => void;
  onDragOver?: (e: React.DragEvent) => void;
  onDrop?: (e: React.DragEvent) => void;
  /** Stagger delay in seconds for deal animation. */
  dealDelay?: number;
  /** Whether this card is selected (lift + glow). */
  isSelected?: boolean;
  /** Shared layout animation ID. */
  layoutId?: string;
}

/** Renders an animated face-up playing card with deal and select animations. */
export function AnimatedCard({
  card,
  width,
  style,
  className,
  draggable,
  onDragStart,
  onDragOver,
  onDrop,
  dealDelay = 0,
  isSelected = false,
  layoutId,
}: AnimatedCardProps) {
  const reduced = useReducedMotion();

  if (reduced) {
    return (
      <CardImage
        card={card}
        width={width}
        style={style}
        className={className}
        draggable={draggable}
        onDragStart={onDragStart}
        onDragOver={onDragOver}
        onDrop={onDrop}
      />
    );
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
      <CardImage
        card={card}
        width={width}
        style={style}
        className={className}
        draggable={draggable}
        onDragStart={onDragStart}
        onDragOver={onDragOver}
        onDrop={onDrop}
      />
    </motion.div>
  );
}
