const ART_PALETTES = [
  {
    background: ['#082f49', '#0f766e'],
    accent: '#38bdf8',
    accentSoft: 'rgba(56, 189, 248, 0.28)',
    secondary: '#f59e0b',
    surface: '#f8fafc',
    ink: '#e0f2fe'
  },
  {
    background: ['#172554', '#312e81'],
    accent: '#818cf8',
    accentSoft: 'rgba(129, 140, 248, 0.28)',
    secondary: '#f472b6',
    surface: '#eef2ff',
    ink: '#e0e7ff'
  },
  {
    background: ['#3f1d2e', '#7c2d12'],
    accent: '#fb7185',
    accentSoft: 'rgba(251, 113, 133, 0.28)',
    secondary: '#fdba74',
    surface: '#fff7ed',
    ink: '#ffe4e6'
  },
  {
    background: ['#052e16', '#166534'],
    accent: '#34d399',
    accentSoft: 'rgba(52, 211, 153, 0.28)',
    secondary: '#fde68a',
    surface: '#f0fdf4',
    ink: '#dcfce7'
  },
  {
    background: ['#111827', '#1f2937'],
    accent: '#22d3ee',
    accentSoft: 'rgba(34, 211, 238, 0.24)',
    secondary: '#f9a8d4',
    surface: '#f8fafc',
    ink: '#e5e7eb'
  }
] as const;

function hashSeed(seed: string): number {
  let hash = 0;

  for (let index = 0; index < seed.length; index += 1) {
    hash = (hash << 5) - hash + seed.charCodeAt(index);
    hash |= 0;
  }

  return Math.abs(hash);
}

function escapeXml(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&apos;');
}

function splitTitle(title: string, maxLines = 3, maxLength = 15): string[] {
  const words = title.trim().split(/\s+/).filter(Boolean);
  const lines: string[] = [];
  let current = '';

  for (const word of words) {
    const next = current ? `${current} ${word}` : word;
    if (next.length <= maxLength || !current) {
      current = next;
      continue;
    }

    lines.push(current);
    current = word;

    if (lines.length === maxLines - 1) {
      break;
    }
  }

  const consumedWords = lines.join(' ').split(/\s+/).filter(Boolean).length;
  const remainingWords = words.slice(consumedWords);
  const tail = [current, ...remainingWords].filter(Boolean).join(' ').trim();

  if (tail) {
    lines.push(
      tail.length > maxLength * 1.3 ? `${tail.slice(0, maxLength * 1.3).trim()}...` : tail
    );
  }

  return lines.slice(0, maxLines);
}

export function getStorefrontCategory(title: string): string {
  const normalized = title.toLowerCase();

  if (/hoodie|shirt|cap|tee|sweatshirt/.test(normalized)) return 'wearables';
  if (/mug|bottle|cup|glass/.test(normalized)) return 'desk essentials';
  if (/sticker|pack|poster|print/.test(normalized)) return 'small-format drop';
  if (/bag|tote|case/.test(normalized)) return 'carry collection';

  return 'signature goods';
}

function getMotifMarkup(
  variant: number,
  width: number,
  height: number,
  accent: string,
  accentSoft: string,
  secondary: string
): string {
  switch (variant % 4) {
    case 0:
      return `
        <circle cx="${width * 0.76}" cy="${height * 0.26}" r="${Math.min(width, height) * 0.18}" fill="${accentSoft}" />
        <circle cx="${width * 0.32}" cy="${height * 0.78}" r="${Math.min(width, height) * 0.25}" fill="none" stroke="${secondary}" stroke-width="3" opacity="0.7" />
        <path d="M ${width * 0.08} ${height * 0.62} C ${width * 0.24} ${height * 0.42}, ${width * 0.54} ${height * 0.52}, ${width * 0.92} ${height * 0.24}" stroke="${accent}" stroke-width="8" stroke-linecap="round" opacity="0.95" />
      `;
    case 1:
      return `
        <rect x="${width * 0.12}" y="${height * 0.14}" width="${width * 0.58}" height="${height * 0.3}" rx="32" fill="${accentSoft}" />
        <rect x="${width * 0.42}" y="${height * 0.5}" width="${width * 0.38}" height="${height * 0.16}" rx="24" fill="${secondary}" opacity="0.9" />
        <path d="M ${width * 0.08} ${height * 0.86} L ${width * 0.92} ${height * 0.18}" stroke="${accent}" stroke-width="4" stroke-dasharray="10 14" opacity="0.55" />
      `;
    case 2:
      return `
        <circle cx="${width * 0.25}" cy="${height * 0.26}" r="${Math.min(width, height) * 0.12}" fill="${secondary}" opacity="0.85" />
        <circle cx="${width * 0.74}" cy="${height * 0.72}" r="${Math.min(width, height) * 0.2}" fill="${accentSoft}" />
        <path d="M ${width * 0.18} ${height * 0.58} C ${width * 0.28} ${height * 0.4}, ${width * 0.48} ${height * 0.36}, ${width * 0.62} ${height * 0.52} S ${width * 0.88} ${height * 0.78}, ${width * 0.9} ${height * 0.56}" stroke="${accent}" stroke-width="7" fill="none" stroke-linecap="round" opacity="0.95" />
      `;
    default:
      return `
        <rect x="${width * 0.12}" y="${height * 0.18}" width="${width * 0.2}" height="${height * 0.2}" rx="28" fill="${secondary}" opacity="0.9" />
        <rect x="${width * 0.58}" y="${height * 0.16}" width="${width * 0.18}" height="${height * 0.44}" rx="28" fill="${accentSoft}" />
        <rect x="${width * 0.26}" y="${height * 0.54}" width="${width * 0.52}" height="${height * 0.18}" rx="36" fill="none" stroke="${accent}" stroke-width="4" opacity="0.8" />
        <path d="M ${width * 0.16} ${height * 0.78} H ${width * 0.82}" stroke="${accent}" stroke-width="10" stroke-linecap="round" opacity="0.88" />
      `;
  }
}

export function getStorefrontArtwork(
  title: string,
  seed: string,
  options?: {
    width?: number;
    height?: number;
    variant?: number;
    eyebrow?: string;
    subtitle?: string;
  }
): string {
  const width = options?.width ?? 720;
  const height = options?.height ?? 900;
  const seedHash = hashSeed(`${seed}:${title}:${options?.variant ?? 0}`);
  const palette = ART_PALETTES[seedHash % ART_PALETTES.length] ?? ART_PALETTES[0];
  const titleLines = splitTitle(title);
  const eyebrow = (options?.eyebrow ?? 'shortlink shop').toUpperCase();
  const subtitle = options?.subtitle ?? getStorefrontCategory(title);
  const motif = getMotifMarkup(
    seedHash,
    width,
    height,
    palette.accent,
    palette.accentSoft,
    palette.secondary
  );

  const titleMarkup = titleLines
    .map(
      (line, index) =>
        `<text x="72" y="${height - 180 + index * 48}" fill="${palette.surface}" font-size="36" font-weight="700" font-family="Inter, Arial, sans-serif" letter-spacing="-0.03em">${escapeXml(line)}</text>`
    )
    .join('');

  const svg = `
    <svg xmlns="http://www.w3.org/2000/svg" width="${width}" height="${height}" viewBox="0 0 ${width} ${height}" fill="none">
      <defs>
        <linearGradient id="bg" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stop-color="${palette.background[0]}" />
          <stop offset="100%" stop-color="${palette.background[1]}" />
        </linearGradient>
        <radialGradient id="glow" cx="0" cy="0" r="1" gradientUnits="userSpaceOnUse" gradientTransform="translate(${width * 0.72} ${height * 0.22}) rotate(118) scale(${width * 0.7} ${height * 0.54})">
          <stop stop-color="${palette.accentSoft}" />
          <stop offset="1" stop-color="transparent" />
        </radialGradient>
        <pattern id="grid" width="40" height="40" patternUnits="userSpaceOnUse">
          <path d="M 40 0 L 0 0 0 40" fill="none" stroke="rgba(255,255,255,0.08)" stroke-width="1" />
        </pattern>
      </defs>
      <rect width="${width}" height="${height}" rx="48" fill="url(#bg)" />
      <rect width="${width}" height="${height}" rx="48" fill="url(#glow)" />
      <rect x="32" y="32" width="${width - 64}" height="${height - 64}" rx="34" fill="url(#grid)" opacity="0.55" />
      <rect x="48" y="48" width="${width - 96}" height="${height - 96}" rx="36" stroke="rgba(255,255,255,0.12)" />
      ${motif}
      <rect x="56" y="${height - 252}" width="${width - 112}" height="196" rx="30" fill="rgba(6, 10, 22, 0.42)" stroke="rgba(255,255,255,0.12)" />
      <text x="72" y="${height - 214}" fill="${palette.ink}" font-size="18" font-weight="700" font-family="Inter, Arial, sans-serif" letter-spacing="0.26em">${escapeXml(eyebrow)}</text>
      ${titleMarkup}
      <text x="72" y="${height - 70}" fill="${palette.ink}" font-size="20" font-weight="500" font-family="Inter, Arial, sans-serif" opacity="0.92">${escapeXml(subtitle)}</text>
    </svg>
  `;

  return `data:image/svg+xml;charset=UTF-8,${encodeURIComponent(svg)}`;
}
