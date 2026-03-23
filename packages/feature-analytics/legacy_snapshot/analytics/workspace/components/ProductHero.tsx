import type { AnalyticsProductWorkspaceV1Dto } from "@metalshopping/feature-analytics";
import { HeroMetrics } from "./HeroMetrics";
import styles from "../product_workspace.module.css";

type ProductHeroProps = {
  model: AnalyticsProductWorkspaceV1Dto["model"];
};

export function ProductHero({ model }: ProductHeroProps) {
  return (
    <section className={styles.productHero}>
      <HeroMetrics items={model.heroMetrics} />
    </section>
  );
}
