/** Props for ChipBetInput. */
export interface ChipBetInputProps {
  /** DOM id used to associate the label with the input. */
  id: string;
  /** Visible label text. */
  label: string;
  /** Current bet value. */
  value: number;
  /** Called with the new numeric value when the user edits the field. */
  onChange: (value: number) => void;
  /** Maximum allowed bet (typically the player's chip balance). */
  max: number;
  /** Minimum allowed bet. Defaults to 10. */
  min?: number;
  /** Step size. Defaults to 10. */
  step?: number;
  /** Disable the input. */
  disabled?: boolean;
  /** Tailwind width class for the input. Defaults to "w-24". */
  widthClass?: string;
}

/**
 * Reusable label + numeric chip-bet input used across casino games
 * (Baccarat, Caribbean Stud, Three Card Poker, Pai Gow, Let It Ride).
 */
export function ChipBetInput({
  id,
  label,
  value,
  onChange,
  max,
  min = 10,
  step = 10,
  disabled,
  widthClass = 'w-24',
}: ChipBetInputProps) {
  return (
    <div className="flex items-center gap-2">
      <label htmlFor={id} className="text-white text-sm">
        {label}
      </label>
      <input
        id={id}
        type="number"
        min={min}
        max={max}
        step={step}
        value={value}
        onChange={(e) => {
          const parsed = Number(e.target.value);
          if (Number.isNaN(parsed)) return;
          onChange(Math.max(min, Math.min(parsed, max)));
        }}
        disabled={disabled}
        className={`${widthClass} px-2 py-1 rounded text-sm`}
      />
    </div>
  );
}
