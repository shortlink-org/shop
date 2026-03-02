import { ProductGrid } from '@shortlink-org/ui-kit';

export default function SearchLoading() {
  return (
    <ProductGrid
      className="shop-productgrid"
      products={[]}
      loading
      skeletonCount={6}
    />
  );
}
