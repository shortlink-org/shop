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
}
