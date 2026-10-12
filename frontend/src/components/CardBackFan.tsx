import { AnimatedCardBack } from './motion/AnimatedCardBack';

/** Render a compact horizontal fan of CPU card backs. */
export function CardBackFan({
  count,
  width,
  step = Math.round(width / 4),
}: {
  count: number;
  width: number;
  step?: number;
}) {
  return (
    <div className="flex justify-center" data-testid="card-back-fan" aria-hidden="true">
      {Array.from({ length: count }, (_, index) => (
        <div key={index} style={{ marginLeft: index === 0 ? 0 : step - width }}>
          <AnimatedCardBack width={width} />
        </div>
      ))}
    </div>
  );
}
