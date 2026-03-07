declare module '@shortlink-org/ui-kit' {
  import { ButtonHTMLAttributes, ElementType, FC, ReactNode } from 'react';

  export type ButtonVariant =
    | 'primary'
    | 'secondary'
    | 'outline'
    | 'ghost'
    | 'link'
    | 'destructive';
  export type ButtonSize = 'sm' | 'md' | 'lg' | 'icon';
  export type IconPosition = 'left' | 'right' | 'only';

  export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
    variant?: ButtonVariant;
    size?: ButtonSize;
    icon?: ReactNode;
    iconPosition?: IconPosition;
    loading?: boolean;
    className?: string;
    children?: ReactNode;
    as?: ElementType;
    asProps?: Record<string, unknown>;
  }

  export const Button: FC<ButtonProps>;

  export interface ToggleDarkModeProps {
    className?: string;
  }

  export const ToggleDarkMode: FC<ToggleDarkModeProps>;

  export interface FooterLink {
    id?: string;
    label: string;
    href: string;
    target?: '_blank' | '_self';
    render?: (props: {
      href: string;
      children: ReactNode;
      className?: string;
    }) => ReactNode;
  }

  export interface SocialLink {
    name: string;
    href: string;
    iconPath: string;
    viewBox?: string;
  }

  export interface FooterProps {
    className?: string;
    links?: FooterLink[];
    socialLinks?: SocialLink[];
    copyright?: ReactNode;
    logoSlot?: ReactNode;
    description?: ReactNode;
    LinkComponent?: FC<{
      href: string;
      children: ReactNode;
      className?: string;
      target?: string;
    }>;
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

  export type FeedbackVariant = 'loading' | 'error' | 'empty';

  export interface FeedbackPanelProps {
    variant: FeedbackVariant;
    title?: string;
    eyebrow?: string;
    message?: string;
    icon?: ReactNode;
    children?: ReactNode;
    className?: string;
    size?: 'sm' | 'md' | 'lg';
    action?: ReactNode;
  }

  export const FeedbackPanel: FC<FeedbackPanelProps>;

  export interface BasketItemData {
    id: number | string;
    name: string;
    href: string;
    color?: string;
    price: string;
    quantity: number;
    imageSrc: string;
    imageAlt: string;
  }

  export interface BasketItemProps {
    item: BasketItemData;
    onRemove?: (itemId: number | string) => void;
    onQuantityChange?: (itemId: number | string, quantity: number) => void;
    confirmRemove?: boolean;
    className?: string;
  }

  export const BasketItem: FC<BasketItemProps>;

  export interface BasketProps {
    open: boolean;
    onClose: (open: boolean) => void;
    items: BasketItemData[];
    subtotal: string;
    shippingNote?: string;
    checkoutText?: string;
    checkoutHref?: string;
    onCheckout?: () => void;
    onContinueShopping?: () => void;
    continueShoppingText?: string;
    onRemoveItem?: (itemId: number | string) => void;
    onClearBasket?: () => void;
    onQuantityChange?: (itemId: number | string, quantity: number) => void;
    clearBasketText?: string;
    emptyMessage?: ReactNode;
    itemsClassName?: string;
    position?: 'left' | 'right' | 'bottom';
    size?: 'sm' | 'md' | 'lg' | 'xl' | 'full';
    className?: string;
    panelClassName?: string;
    backdropClassName?: string;
    titleClassName?: string;
    contentClassName?: string;
  }

  export const Basket: FC<BasketProps>;
  export type BasketItem = BasketItemData;

  export interface AppHeaderBrand {
    name: string;
    logo?: ReactNode;
    href?: string;
    render?: (props: {
      href: string;
      children: ReactNode;
      className?: string;
    }) => ReactNode;
  }

  export interface AppHeaderNavigationItem {
    name: string;
    href: string;
    current?: boolean;
    badge?: string;
  }

  export interface AppHeaderStatusBadge {
    label: string;
    tone?: 'neutral' | 'accent' | 'success' | 'warning';
  }

  export interface AppHeaderMenuItem {
    name: string;
    description?: string;
    href: string;
    icon?: ReactNode;
  }

  export interface AppHeaderNotification {
    count?: number;
    items?: Array<{
      id: string;
      title: string;
      message: string;
      time: string;
      avatar?: string;
      onClick?: () => void;
    }>;
    seeAllHref?: string;
    render?: (props: {
      count?: number;
      items?: AppHeaderNotification['items'];
    }) => ReactNode;
  }

  export interface AppHeaderProfile {
    avatar?: string;
    name?: string;
    email?: string;
    menuItems?: Array<{
      name: string;
      href?: string;
      icon?: string | ReactNode;
      onClick?: () => void | Promise<void>;
      confirmDialog?: {
        title: string;
        description?: string;
        confirmText?: string;
        cancelText?: string;
        variant?: 'default' | 'danger' | 'success';
      };
    }>;
    render?: (props: {
      avatar?: string;
      name?: string;
      email?: string;
    }) => ReactNode;
  }

  export interface AppHeaderProps {
    className?: string;
    brand?: AppHeaderBrand;
    workspaceLabel?: string;
    statusBadge?: AppHeaderStatusBadge;
    showMenuButton?: boolean;
    onMenuClick?: () => void;
    menuButtonDisabled?: boolean;
    navigation?: AppHeaderNavigationItem[];
    currentPath?: string;
    showThemeToggle?: boolean;
    showSecondMenu?: boolean;
    secondMenuItems?: AppHeaderMenuItem[];
    secondMenuLabel?: string;
    showSearch?: boolean;
    searchProps?: {
      placeholder?: string;
      onSearch?: (query: string) => void;
      defaultQuery?: string;
    };
    showNotifications?: boolean;
    notifications?: AppHeaderNotification;
    showProfile?: boolean;
    profile?: AppHeaderProfile;
    showLogin?: boolean;
    loginButton?: {
      href?: string;
      label?: string;
      onClick?: () => void;
    };
    LinkComponent?: FC<{
      href: string;
      children: ReactNode;
      className?: string;
    }>;
    themeToggleComponent?: ReactNode;
    sticky?: boolean;
    fullWidth?: boolean;
  }

  export const AppHeader: FC<AppHeaderProps>;
}
