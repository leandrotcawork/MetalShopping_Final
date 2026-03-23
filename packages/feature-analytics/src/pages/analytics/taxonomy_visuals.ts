const EMOJI_MAP: Array<{ emoji: string; keywords: string[] }> = [
  { emoji: "🧱", keywords: ["revest", "porcel", "ceram", "pisos", "cobert"] },
  { emoji: "🚿", keywords: ["hidraul", "tub", "tigre", "registro", "torne", "banh", "chuve"] },
  { emoji: "🔌", keywords: ["eletric", "lamp", "tomada", "fio", "cabo", "disjunt"] },
  { emoji: "🧰", keywords: ["ferrag", "paraf", "broca", "fixa", "ferrament"] },
  { emoji: "🔩", keywords: ["metais", "metal", "inox", "aco", "lat", "cobre"] },
  { emoji: "🎨", keywords: ["tinta", "tintas", "pint", "suvin", "cor"] },
  { emoji: "🚽", keywords: ["louca", "loucas", "sanit", "vaso", "assento"] },
  { emoji: "🧪", keywords: ["argam", "rejunt", "cimento", "massa", "quartz"] },
  { emoji: "🪵", keywords: ["madeir", "mdf", "lamin", "rodape"] },
  { emoji: "🧊", keywords: ["clima", "ar-cond", "ventil", "exaus"] },
];

export function taxonomyEmojiForNodeName(name: string): string {
  const token = String(name || "")
    .trim()
    .toLocaleLowerCase("pt-BR");

  if (!token) return "🗂️";

  for (const entry of EMOJI_MAP) {
    if (entry.keywords.some((kw) => token.includes(kw))) return entry.emoji;
  }

  return "🗂️";
}

