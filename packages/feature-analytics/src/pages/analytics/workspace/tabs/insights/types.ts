export type InsightDomain = "COMPETITIVIDADE" | "ESTOQUE" | "RENTABILIDADE" | "RISCO" | "DADOS" | "MERCADO";

export type InsightSeverity = "CRITICAL" | "WARN" | "INFO" | "GOOD";

export type InsightAction = {
  label: string;
  goTo?: "overview" | "history" | "simulator";
  kind?: "primary" | "secondary";
};

export type InsightEvidence = {
  label: string;
  value: string;
};

export type Insight = {
  id: string;
  title: string;
  summary: string;
  domain: InsightDomain;
  severity: InsightSeverity;
  category: "preco" | "campanha" | "estoque" | "risco" | "portfolio";
  confidence?: number;
  updatedAt?: string;
  evidence?: InsightEvidence[];
  tags?: string[];
  alerts?: string[];
  actionTags?: string[];
  actions?: InsightAction[];
};

export type InsightSection = {
  title: string;
  subtitle?: string;
  items: Insight[];
};
