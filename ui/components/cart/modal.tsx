'use client';

import { Basket, Button, Drawer, FeedbackPanel } from '@shortlink-org/ui-kit';
import { useEffect, useRef, useState } from 'react';
import { createCartAndSetCookie, redirectToCheckout } from './actions';
import { useCart } from './cart-context';
import OpenCart from './open-cart';
import { useCartBasket } from './use-cart-basket';

function CartStateDrawer({
  open,
  onClose,
  cartUnavailable
}: {
  open: boolean;
  onClose: (open: boolean) => void;
  cartUnavailable: boolean;
}) {
  return (
    <Drawer
      open={open}
      onClose={onClose}
      position="right"
      size="md"
      title={
        <div>
          <h2 className="text-lg font-semibold tracking-tight text-[var(--color-foreground)] sm:text-xl">
            Cart
          </h2>
          <p className="mt-1 text-sm text-[var(--color-muted-foreground)]">Empty</p>
        </div>
      }
      titleClassName="!text-inherit"
      panelClassName="sm:!rounded-[1.75rem]"
      backdropClassName="bg-[var(--color-background)]/40 backdrop-blur-[2px]"
      contentClassName="!px-4 !pb-6 !pt-4 sm:!px-6"
    >
      <div className="flex h-full min-h-0 items-center justify-center">
        <div className="w-full max-w-sm rounded-[1.5rem] border border-[var(--color-border)] bg-[var(--color-surface)] p-6 shadow-[0_24px_60px_-40px_rgba(15,23,42,0.28)]">
          <FeedbackPanel
            variant={cartUnavailable ? 'error' : 'empty'}
            eyebrow="Cart status"
            title={cartUnavailable ? "We couldn't load your cart" : 'Your cart is empty'}
            message={
              cartUnavailable
                ? "We'll show it when it's available again. You can keep browsing in the meantime."
                : 'Add a few products and come back when you are ready to check out.'
            }
            size="sm"
            action={
              <Button variant="secondary" size="sm" onClick={() => onClose(false)}>
                Continue shopping
              </Button>
            }
          />
        </div>
      </div>
    </Drawer>
  );
}

export default function CartModal() {
  const { cart, cartUnavailable } = useCart();
  const { items, subtotal, handleRemoveItem, handleQuantityChange } = useCartBasket();
  const [isOpen, setIsOpen] = useState(false);
  const quantityRef = useRef(cart?.totalQuantity);
  const openCart = () => setIsOpen(true);
  const closeCart = () => setIsOpen(false);

  useEffect(() => {
    if (!cart && !cartUnavailable) {
      createCartAndSetCookie();
    }
  }, [cart, cartUnavailable]);

  useEffect(() => {
    if (
      cart?.totalQuantity &&
      cart?.totalQuantity !== quantityRef.current &&
      cart?.totalQuantity > 0
    ) {
      quantityRef.current = cart?.totalQuantity;
      if (!isOpen) {
        queueMicrotask(() => setIsOpen(true));
      }
    }
  }, [isOpen, cart?.totalQuantity, quantityRef]);

  return (
    <>
      <button
        type="button"
        aria-label="Open cart"
        onClick={openCart}
        className="focus-ring relative inline-flex"
      >
        <OpenCart quantity={cart?.totalQuantity ?? 0} />
      </button>
      {items.length === 0 ? (
        <CartStateDrawer open={isOpen} onClose={setIsOpen} cartUnavailable={cartUnavailable} />
      ) : (
        <Basket
          open={isOpen}
          onClose={closeCart}
          position="right"
          size="md"
          items={items}
          subtotal={subtotal}
          shippingNote="Shipping and taxes are confirmed during checkout."
          checkoutText="Proceed to checkout"
          continueShoppingText="Keep browsing"
          panelClassName="sm:!rounded-[1.75rem]"
          backdropClassName="bg-[var(--color-background)]/40 backdrop-blur-[2px]"
          onCheckout={() => {
            closeCart();
            void redirectToCheckout();
          }}
          onContinueShopping={closeCart}
          onRemoveItem={handleRemoveItem}
          onQuantityChange={handleQuantityChange}
        />
      )}
    </>
  );
}
