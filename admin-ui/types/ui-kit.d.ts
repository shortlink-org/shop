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
  export type StatCardTone = 'neutral' | 'accent' | 'success' | 'warning' | 'danger';
  export type FeedbackVariant = 'loading' | 'error' | 'empty';

  export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
    variant?: ButtonVariant;
    size?: ButtonSize;
    icon?: ReactNode;
    iconPosition?: 'left' | 'right' | 'only';
    loading?: boolean;
    className?: string;
    children?: ReactNode;
    as?: ElementType;
    asProps?: Record<string, unknown>;
  }

  export const Button: FC<ButtonProps>;

  export interface StatCardProps {
    label: ReactNode;
    value: ReactNode;
    change?: ReactNode;
    tone?: StatCardTone;
    className?: string;
    labelClassName?: string;
    valueClassName?: string;
    changeClassName?: string;
  }

  export const StatCard: FC<StatCardProps>;

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

  export interface AppHeaderBrand {
    name: string;
    logo?: ReactNode;
    href?: string;
    render?: (props: { href: string; children: ReactNode; className?: string }) => ReactNode;
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
    render?: (props: { avatar?: string; name?: string; email?: string }) => ReactNode;
  }

  export interface AppHeaderProps {
    className?: string;
    brand?: AppHeaderBrand;
    workspaceLabel?: string;
    statusBadge?: { label: string; tone?: 'neutral' | 'accent' | 'success' | 'warning' };
    showMenuButton?: boolean;
    onMenuClick?: () => void;
    menuButtonDisabled?: boolean;
    navigation?: Array<{ name: string; href: string; current?: boolean; badge?: string }>;
    currentPath?: string;
    showThemeToggle?: boolean;
    showSecondMenu?: boolean;
    secondMenuItems?: Array<{ name: string; description?: string; href: string; icon?: ReactNode }>;
    secondMenuLabel?: string;
    showSearch?: boolean;
    searchProps?: { placeholder?: string; onSearch?: (query: string) => void; defaultQuery?: string };
    showNotifications?: boolean;
    notifications?: {
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
      render?: (props: { count?: number; items?: AppHeaderProps['notifications']['items'] }) => ReactNode;
    };
    showProfile?: boolean;
    profile?: AppHeaderProfile;
    showLogin?: boolean;
    loginButton?: { href?: string; label?: string; onClick?: () => void };
    LinkComponent?: FC<{ href: string; children: ReactNode; className?: string }>;
    themeToggleComponent?: ReactNode;
    sticky?: boolean;
    fullWidth?: boolean;
  }

  export const AppHeader: FC<AppHeaderProps>;
}
