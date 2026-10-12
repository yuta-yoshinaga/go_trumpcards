/** A rendered text node that falls below its WCAG AA contrast threshold. */
export interface TextContrastViolation {
  ratio: number;
  text: string;
  kind: 'muted' | 'colored-button' | 'other';
}

/** Counts and findings produced by the rendered text contrast scan. */
export interface TextContrastResult {
  violations: TextContrastViolation[];
  skipped: number;
  checked: number;
}

/** Measure visible text contrast against its composited DOM background. */
export function measureTextContrast(): TextContrastResult {
  const canvas = document.createElement('canvas');
  canvas.width = canvas.height = 1;
  const context = canvas.getContext('2d', { willReadFrequently: true });
  if (!context) return { violations: [], skipped: 0, checked: 0 };

  const parseColor = (value: string) => {
    if (!value || value === 'transparent') return null;
    context.clearRect(0, 0, 1, 1);
    context.fillStyle = '#000';
    context.fillStyle = value;
    context.fillRect(0, 0, 1, 1);
    const pixel = context.getImageData(0, 0, 1, 1).data;
    if (pixel[3] === 0) return null;
    return { r: pixel[0], g: pixel[1], b: pixel[2], a: pixel[3] / 255 };
  };

  const blend = (top: NonNullable<ReturnType<typeof parseColor>>, bottom: { r: number; g: number; b: number }) => ({
    r: top.r * top.a + bottom.r * (1 - top.a),
    g: top.g * top.a + bottom.g * (1 - top.a),
    b: top.b * top.a + bottom.b * (1 - top.a),
  });
  const luminance = (color: { r: number; g: number; b: number }) => {
    const channel = (value: number) => {
      const normalized = value / 255;
      return normalized <= 0.03928 ? normalized / 12.92 : ((normalized + 0.055) / 1.055) ** 2.4;
    };
    return 0.2126 * channel(color.r) + 0.7152 * channel(color.g) + 0.0722 * channel(color.b);
  };

  const backgroundOf = (element: Element) => {
    const layers: NonNullable<ReturnType<typeof parseColor>>[] = [];
    let current: Element | null = element;
    while (current) {
      const style = getComputedStyle(current);
      if (style.backgroundImage !== 'none') return null;
      const color = parseColor(style.backgroundColor);
      if (color) {
        layers.push(color);
        if (color.a >= 1) break;
      }
      current = current.parentElement;
    }
    // The nearest opaque layer is the base; translucent layers above it are blended on top.
    let background = { r: 255, g: 255, b: 255 };
    const base = layers.at(-1);
    if (base && base.a >= 1) {
      background = { r: base.r, g: base.g, b: base.b };
      layers.pop();
    }
    for (let index = layers.length - 1; index >= 0; index--) {
      background = blend(layers[index], background);
    }
    return background;
  };

  const muted = parseColor(
    getComputedStyle(document.documentElement).getPropertyValue('--ds-text-muted-surface').trim(),
  );
  const matchesColor = (
    left: NonNullable<ReturnType<typeof parseColor>>,
    right: NonNullable<ReturnType<typeof parseColor>>,
  ) => left.r === right.r && left.g === right.g && left.b === right.b;
  const violations: TextContrastViolation[] = [];
  let skipped = 0;
  let checked = 0;
  const seen = new Set<Element>();
  const walker = document.createTreeWalker(document.body, NodeFilter.SHOW_TEXT);
  let node: Node | null;

  while (true) {
    node = walker.nextNode();
    if (!node) break;
    const text = node.textContent?.trim() ?? '';
    const element = node.parentElement;
    if (text.length < 2 || !element || seen.has(element)) continue;
    seen.add(element);
    const bounds = element.getBoundingClientRect();
    const style = getComputedStyle(element);
    if (bounds.width === 0 || bounds.height === 0 || style.visibility === 'hidden') continue;
    if (element.closest('.sr-only,[aria-hidden="true"],button:disabled,[aria-disabled="true"]')) continue;

    let opacity = 1;
    let ancestor: Element | null = element;
    while (ancestor) {
      opacity *= Number.parseFloat(getComputedStyle(ancestor).opacity);
      ancestor = ancestor.parentElement;
    }
    if (opacity < 0.1) continue;

    let hasBackgroundImage = false;
    let backgroundAncestor: Element | null = element;
    while (backgroundAncestor) {
      if (getComputedStyle(backgroundAncestor).backgroundImage !== 'none') {
        hasBackgroundImage = true;
        break;
      }
      backgroundAncestor = backgroundAncestor.parentElement;
    }
    if (hasBackgroundImage) {
      skipped++;
      continue;
    }
    const foreground = parseColor(style.color);
    const background = backgroundOf(element);
    if (!foreground || !background) continue;
    checked++;
    foreground.a *= opacity;
    const renderedForeground = foreground.a < 1 ? blend(foreground, background) : foreground;
    const foregroundLuminance = luminance(renderedForeground);
    const backgroundLuminance = luminance(background);
    const ratio =
      (Math.max(foregroundLuminance, backgroundLuminance) + 0.05) /
      (Math.min(foregroundLuminance, backgroundLuminance) + 0.05);
    const fontSize = Number.parseFloat(style.fontSize);
    const large = fontSize >= 24 || (Number.parseInt(style.fontWeight, 10) >= 700 && fontSize >= 18.66);
    if (ratio >= (large ? 3 : 4.5)) continue;

    const color = parseColor(style.color);
    const kind = element.closest('button')?.matches('.text-ds-text-on-accent')
      ? 'colored-button'
      : color && muted && matchesColor(color, muted)
        ? 'muted'
        : 'other';
    violations.push({ ratio, text: text.slice(0, 80), kind });
  }
  return { violations, skipped, checked };
}
