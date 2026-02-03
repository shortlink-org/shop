// Sentinels and result types (used by Server/Client Components)
export {
  CART_UNAVAILABLE,
  GOODS_UNAVAILABLE,
  type CartLoadResult,
  type GoodsLoadResult
} from './sentinels';

// Low-level fetch (for advanced use or revalidate)
export { shopifyFetch } from './fetch';

// Cart API
export {
  addToCart,
  checkout,
  createCart,
  getCart,
  removeFromCart,
  updateCart
} from './api/cart';

// Good API
export { getGood, getGoodRecommendations, getGoods } from './api/good';

// Collection API
export {
  getCollection,
  getCollectionProducts,
  getCollections
} from './api/collection';

// Menu API
export { getMenu } from './api/menu';

// Page API
export { getPage, getPages } from './api/page';
