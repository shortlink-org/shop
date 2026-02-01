export type Maybe<T> = T | null;

export type Connection<T> = {
  edges: Array<Edge<T>>;
};

export type Edge<T> = {
  node: T;
};

export type Cart = Omit<ShopifyCart, 'lines'> & {
  lines: CartItem[];
};

export type CartProduct = {
  id: number;
  handle: string;
  title: string;
  featuredImage?: Image;
};

export type CartItem = {
  id: string | undefined;
  quantity: number;
  cost: {
    totalAmount: Money;
  };
  merchandise: {
    id: string;
    title: string;
    selectedOptions: {
      name: string;
      value: string;
    }[];
    product: CartProduct;
  };
};

export type Collection = ShopifyCollection & {
  path: string;
};

export type Image = {
  url: string;
  altText: string;
  width: number;
  height: number;
};

export type Menu = {
  title: string;
  path: string;
};

export type Money = {
  amount: number;
  currencyCode: string;
};

export type Page = {
  id: string;
  title: string;
  handle: string;
  body: string;
  bodySummary: string;
  seo?: SEO;
  createdAt: string;
  updatedAt: string;
};

/**
 * Good type represents a product/item available for purchase.
 */
export type Good = Omit<ShopifyProduct, 'price' | 'description' | 'createdAt' | 'updatedAt'> & {
  name: string;
  price: number;
  description: string;
  createdAt: string;
  updatedAt: string;
};

export type GoodOption = {
  id: string;
  name: string;
  values: string[];
};

export type GoodVariant = {
  id: string;
  title: string;
  availableForSale: boolean;
  selectedOptions: {
    name: string;
    value: string;
  }[];
  price: Money;
};

// Legacy aliases for backward compatibility (can be removed later)
export type Product = Good;
export type ProductOption = GoodOption;
export type ProductVariant = GoodVariant;

export type SEO = {
  title: string;
  description: string;
};

export type ShopifyCart = {
  id: string | undefined;
  checkoutUrl: string;
  cost: {
    subtotalAmount: Money;
    totalAmount: Money;
    totalTaxAmount: Money;
  };
  lines: Connection<CartItem>;
  totalQuantity: number;
};

export type ShopifyCollection = {
  handle: string;
  title: string;
  description: string;
  seo: SEO;
  updatedAt: string;
};

export type ShopifyProduct = {
  id: number;
  name: string;
  price: number;
  description: string;
  createdAt?: string;
  updatedAt?: string;
  created_at?: string;
  updated_at?: string;
};

export type ShopifyCartOperation = {
  data: {
    getCart: {
      state: {
        cartId?: string | null;
        customerId?: string | null;
        items?: Array<{
          goodId?: string | null;
          quantity?: number | null;
        }> | null;
      };
    };
  };
  variables: {
    customerId?: string;
  };
};

export type ShopifyCreateCartOperation = {
  data: {
    addItem: {
      _?: boolean | null;
    } | null;
  };
};

export type ShopifyAddToCartOperation = {
  data: {
    addItem: {
      _?: boolean | null;
    } | null;
  };
  variables: {
    addRequest: {
      customerId: string;
      items: {
        goodId: string;
        quantity: number;
      }[];
    };
  };
};

export type ShopifyRemoveFromCartOperation = {
  data: {
    removeItem: {
      _?: boolean | null;
    } | null;
  };
  variables: {
    removeRequest: {
      customerId: string;
      items: {
        goodId: string;
        quantity: number;
      }[];
    };
  };
};

export type ShopifyUpdateCartOperation = {
  data: {
    addItem: {
      _?: boolean | null;
    } | null;
  };
  variables: {
    addRequest: {
      customerId: string;
      items: {
        goodId: string;
        quantity: number;
      }[];
    };
  };
};

export type ShopifyCollectionOperation = {
  data: {
    collection: ShopifyCollection;
  };
  variables: {
    id: number;
  };
};

export type ShopifyCollectionProductsOperation = {
  data: {
    goods_goods_list: {
      count: number;
      next: string | null;
      previous: string | null;
      results: ShopifyProduct[];
    };
  };
};

export type ShopifyCollectionsOperation = {
  data: {
    collections: Connection<ShopifyCollection>;
  };
};

export type ShopifyMenuOperation = {
  data: {
    menu?: {
      items: {
        title: string;
        url: string;
      }[];
    };
  };
  variables: {
    id: number;
  };
};

export type ShopifyPageOperation = {
  data: { pageByHandle: Page };
  variables: { id: number };
};

export type ShopifyPagesOperation = {
  data: {
    pages: Connection<Page>;
  };
};

export type ShopifyProductOperation = {
  data: { 
    good: ShopifyProduct;
  };
  variables: {
    id: number;
  };
};

export type ShopifyProductRecommendationsOperation = {
  data: {
    goods: {
      count: number;
      next: string | null;
      previous: string | null;
      results: ShopifyProduct[];
    };
  };
  variables: {
    page?: number;
  };
};

export type ShopifyProductsOperation = {
  data: {
    goods: {
      count: number;
      next: string | null;
      previous: string | null;
      results: ShopifyProduct[];
    };
  };
  variables: {
    query?: string;
    reverse?: boolean;
    sortKey?: string;
    page?: number;
  };
};
