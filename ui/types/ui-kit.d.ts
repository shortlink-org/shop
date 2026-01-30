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
}
