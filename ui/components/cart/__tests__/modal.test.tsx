import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import CartModal from '../modal';

const { createCartAndSetCookieMock, redirectToCheckoutMock } = vi.hoisted(() => ({
  redirectToCheckoutMock: vi.fn(() => new Promise<void>(() => {})),
  createCartAndSetCookieMock: vi.fn()
}));

vi.mock('@/lib/ui-kit', () => ({
  Basket: ({ open, onCheckout }: { open: boolean; onCheckout: () => void }) =>
    open ? (
      <div data-testid="basket">
        <button type="button" onClick={onCheckout}>
          Proceed to checkout
        </button>
      </div>
    ) : null,
  Button: ({ children, onClick }: { children: React.ReactNode; onClick?: () => void }) => (
    <button type="button" onClick={onClick}>
      {children}
    </button>
  ),
  Drawer: ({ open, children }: { open: boolean; children: React.ReactNode }) =>
    open ? <div data-testid="drawer">{children}</div> : null,
  FeedbackPanel: ({ title }: { title: string }) => <div>{title}</div>
}));

vi.mock('../actions', () => ({
  createCartAndSetCookie: createCartAndSetCookieMock,
  redirectToCheckout: redirectToCheckoutMock
}));

vi.mock('../cart-context', () => ({
  useCart: () => ({
    cart: {
      id: 'cart-1',
      checkoutUrl: '',
      totalQuantity: 1,
      lines: [],
      cost: {
        subtotalAmount: { amount: 10, currencyCode: 'USD' },
        totalAmount: { amount: 10, currencyCode: 'USD' },
        totalTaxAmount: { amount: 0, currencyCode: 'USD' }
      }
    },
    cartUnavailable: false
  })
}));

vi.mock('../use-cart-basket', () => ({
  useCartBasket: () => ({
    items: [
      {
        id: 'good-1',
        name: 'Test item',
        price: '$10.00',
        quantity: 1
      }
    ],
    subtotal: '$10.00',
    handleRemoveItem: vi.fn(),
    handleQuantityChange: vi.fn()
  })
}));

vi.mock('../open-cart', () => ({
  default: ({ quantity }: { quantity?: number }) => <span>Open cart ({quantity ?? 0})</span>
}));

describe('CartModal', () => {
  it('closes the basket before redirecting to checkout', async () => {
    render(<CartModal />);

    fireEvent.click(screen.getByRole('button', { name: 'Open cart' }));
    expect(screen.getByTestId('basket')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Proceed to checkout' }));

    await waitFor(() => {
      expect(screen.queryByTestId('basket')).not.toBeInTheDocument();
    });
    expect(redirectToCheckoutMock).toHaveBeenCalledTimes(1);
  });
});
