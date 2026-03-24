# Design Tokens — MetalShopping

Source: `apps/web/src/app/global.css` (`:root` CSS variables).
Never hardcode values — always use these.

## Color palette

### Wine — primary brand, actions, interactive
```css
--ms-wine-900: #5f1227   /* pressed state, deepest hover */
--ms-wine-700: #8a1735   /* primary action color, active nav, table header */
--ms-wine-500: #c23b54   /* gradient endpoint, hover on primary */
```

### Ink — text
```css
--ms-ink-900: #251b22    /* h1, main body text, strong values */
--ms-ink-700: #4d3e47    /* secondary text, td content */
--ms-ink-500: #73606a    /* hints, labels, subtitles, captions */
```

### Line — borders
```css
--ms-line-200: #eadfe4   /* card borders (SurfaceCard, MetricCard) */
--ms-line-100: #f3ebee   /* subtle separators */
```

### Surface — backgrounds
```css
--ms-surface-0:   #ffffff   /* pure white — card bg, input bg */
--ms-surface-50:  #fdfafc   /* slight warm tint — page bg, AppFrame hero */
--ms-surface-100: #f8f2f5   /* section bg, table thead gradient endpoint */
```

### Semantic — use only in the appropriate context
```css
/* Success */   color: #0b7c47;  background: #e6f6ee;
/* Warning */   color: #a97900;  background: #fff8e8;
/* Error */     color: #b91c1c;  background: #fef2f2;  border: #fecaca;
/* Info */      color: #1b3a8a;  background: #eff6ff;
```

---

## Typography

### Font families
```css
font-family: "Inter", -apple-system, system-ui, "Segoe UI", sans-serif;  /* all UI */
font-family: "JetBrains Mono", ui-monospace, SFMono-Regular, monospace;  /* code, IDs, mono */
```

### Size scale
| Token | rem | px equiv | Use |
|-------|-----|----------|-----|
| eyebrow/kicker | .68rem | ~11px | uppercase label above title |
| field label | .72rem | ~12px | form field label, table th |
| hint/meta | .75–.78rem | ~12–12.5px | SurfaceCard subtitle, cell meta |
| small/caption | .82rem | ~13px | list item text, cell content |
| body | .88rem | ~14px | paragraph, card text |
| button | .94rem | ~15px | Button component |
| subtitle | 1rem | 16px | AppFrame subtitle |
| card title | 1.12rem | ~18px | SurfaceCard h2 |
| metric value | 1.25rem | 20px | MetricCard value |
| chip value | 1.4rem | ~22px | MetricChip strong |
| hero | 2rem | 32px | AppFrame h1 |

### Font weights
```
600  table cell light, list secondary
700  labels, body text, buttons secondary
800  card titles, metric labels, uppercase headers, cell strong
900  hero title, brand name, metric chip value
```

---

## Border radius
```
6px    checkbox, small badge
10px   button-sm, input, status pill, filter pill
11px   user avatar
12px   button default, input, MetricCard, filter chip, dropdown trigger
14px   MetricChip, SurfaceCard small
16px   SurfaceCard default
20px   AppFrame hero, wizard panel
999px  status pill (fully rounded), filter badge
```

---

## Shadows
```css
/* AppFrame hero card */
box-shadow: 0 12px 30px rgba(74, 39, 50, 0.06);

/* Dropdown menu */
box-shadow: 0 10px 38px rgba(2, 6, 23, 0.08);

/* Button primary */
box-shadow: 0 4px 10px rgba(145, 19, 42, 0.22);

/* Focus ring (input, checkbox, dropdown) */
box-shadow: 0 0 0 3px rgba(184, 56, 79, 0.14);
box-shadow: 0 0 0 4px rgba(199, 63, 103, 0.14);  /* checkbox variant */
```

---

## Spacing / gap reference
```
4px   chip internal gap, hero main gap small
6px   AppFrame heroMain gap, field label gap
8px   list items gap, filter chips gap
10px  filter grid gap, selection actions gap
12px  SurfaceCard internal gap, metrics grid gap, two-col gap
14px  page stack gap (between SurfaceCards)
16px  stack gap variant (HomePage uses 16px)
20px  wizard/shopping container gap
24px  AppShell content padding (set by layout, don't repeat)
32px  table empty state padding
```

---

## Transitions
```css
transition: all 0.2s ease;          /* generic hover states */
transition: all 180ms ease;         /* Button component */
transition: all 0.16s ease;         /* list item hover */
transition: width 0.5s ease;        /* progress bar fill */
transition: opacity 0.28s ease, max-width 0.32s ease;  /* sidebar labels */
```
