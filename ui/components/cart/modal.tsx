'use client';

import { Basket, Button, FeedbackPanel } from '@shortlink-org/ui-kit';
import { useEffect, useRef, useState } from 'react';
import { createCartAndSetCookie, redirectToCheckout } from './actions';
import { useCart } from './cart-context';
import OpenCart from './open-cart';
import { useCartBasket } from './use-cart-basket';

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
      if (!isOpen) {
        setIsOpen(true);
      }
      quantityRef.current = cart?.totalQuantity;
    }
  }, [isOpen, cart?.totalQuantity, quantityRef]);

  return (
    <>
      <button aria-label="Open cart" onClick={openCart}>
        <OpenCart quantity={cart?.totalQuantity ?? 0} />
      </button>
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
        onCheckout={() => {
          void redirectToCheckout();
        }}
        onContinueShopping={closeCart}
        onRemoveItem={handleRemoveItem}
        onQuantityChange={handleQuantityChange}
        emptyMessage={
          cartUnavailable ? (
            <FeedbackPanel
              variant="error"
              eyebrow="Cart status"
              title="We couldn't load your cart"
              message="We'll show it when it's available again. You can keep browsing in the meantime."
              size="sm"
            />
          ) : (
            <FeedbackPanel
              variant="empty"
              eyebrow="Cart status"
              title="Your cart is empty"
              message="Add a few products and come back when you're ready to check out."
              size="sm"
              action={
                <Button variant="secondary" size="sm" onClick={closeCart}>
                  Continue shopping
                </Button>
              }
            />
          )
        }
      />
    </>
  );
}
