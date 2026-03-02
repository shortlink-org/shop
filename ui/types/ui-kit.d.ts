declare module '@shortlink-org/ui-kit' {
  import { FC, ReactNode } from 'react';

  export interface ToggleDarkModeProps {
    className?: string;
  }

  export const ToggleDarkMode: FC<ToggleDarkModeProps>;

  export interface FooterProps {
    children?: ReactNode;
    className?: string;
  }

  export const Footer: FC<FooterProps>;

  export interface BreadcrumbItem {
    id: string;
    name: string;
    href?: string;
  }

  export interface BreadcrumbsProps {
    breadcrumbs: BreadcrumbItem[];
    className?: string;
  }

  export const Breadcrumbs: FC<BreadcrumbsProps>;

  export interface ProductDescriptionProps {
    description?: string;
    highlights?: string[];
    details?: string;
    className?: string;
  }

  export const ProductDescription: FC<ProductDescriptionProps>;

  export interface AddToCartButtonProps {
    text?: string;
    onAddToCart?: () => void | Promise<void>;
    className?: string;
    ariaLabel?: string;
    scale?: number;
    reveal?: boolean;
  }

  export const AddToCartButton: FC<AddToCartButtonProps>;

  /** Product shape for ProductGrid (shop-productgrid) */
  export interface ProductGridProduct {
    id: string;
    name: string;
    href: string;
    imageSrc: string;
    imageAlt: string;
    price: string | { current: number; original?: number; currency?: string; locale?: string; discount?: number };
    description?: string;
    badges?: Array<{ label: string; tone?: 'info' | 'success' | 'warning' | 'error' | 'neutral'; icon?: ReactNode }>;
    badge?: string;
    onSale?: boolean;
    inventory?: { status?: 'out_of_stock' | 'low_stock' | 'preorder' };
    onAddToCart?: () => void | Promise<void>;
    cta?: {
      onFavorite?: (isFavorite: boolean) => void;
      onQuickView?: () => void;
      rating?: number;
      reviewCount?: number;
    };
    isFavorite?: boolean;
  }

  export interface ProductGridProps {
    products: ProductGridProduct[];
    title?: string;
    className?: string;
    gridClassName?: string;
    productClassName?: string;
    columns?: { sm?: number; md?: number; lg?: number; xl?: number };
    spacingY?: 'none' | 'sm' | 'md' | 'lg' | 'xl';
    spacingX?: 'none' | 'sm' | 'md' | 'lg' | 'xl';
    loading?: boolean;
    skeletonCount?: number;
    onProductClick?: (product: ProductGridProduct) => void;
    aspectMobile?: string;
    aspectDesktop?: string;
  }

  export const ProductGrid: FC<ProductGridProps>;

  export interface ProductPageImage {
    src: string;
    alt: string;
  }

  export interface ProductPageProps {
    name: string;
    price: string;
    href?: string;
    breadcrumbs: BreadcrumbItem[];
    images: ProductPageImage[];
    colors?: Array<{ id: string; name: string; hex?: string }>;
    sizes?: Array<{ id: string; name: string; inStock?: boolean }>;
    description?: string;
    highlights?: string[];
    details?: string;
    reviews?: { average?: number; totalCount?: number; href?: string };
    selectedColorId?: string;
    selectedSizeId?: string;
    onColorChange?: (id: string) => void;
    onSizeChange?: (id: string) => void;
    onAddToCart?: () => void | Promise<void>;
    onBuyNow?: () => void | Promise<void>;
    showBuyNow?: boolean;
    sizeGuideHref?: string;
    headerSlot?: ReactNode;
    actionSlot?: ReactNode;
    gallerySlot?: ReactNode;
    galleryVariant?: string;
    enableZoom?: boolean;
    className?: string;
  }

  export const ProductPage: FC<ProductPageProps>;

  /** Product shape for ProductQuickView drawer */
  export interface ProductQuickViewProduct {
    name: string;
    imageSrc: string;
    imageAlt: string;
    price: string;
    rating?: number;
    reviewCount?: number;
    reviewsHref?: string;
    colors?: Array<{ id: string; name: string; hex?: string }>;
    sizes?: Array<{ id: string; name: string; inStock?: boolean }>;
  }

  export interface ProductQuickViewProps {
    open: boolean;
    onClose: (open: false) => void;
    product: ProductQuickViewProduct;
    selectedColorId?: string;
    selectedSizeId?: string;
    onColorChange?: (id: string) => void;
    onSizeChange?: (id: string) => void;
    onAddToCart?: () => void | Promise<void>;
    sizeGuideHref?: string;
    className?: string;
    position?: 'left' | 'right' | 'top' | 'bottom';
    size?: string;
  }

  export const ProductQuickView: FC<ProductQuickViewProps>;
}
