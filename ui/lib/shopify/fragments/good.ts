import imageFragment from './image';
import seoFragment from './seo';

/**
 * GraphQL fragment for Good.
 */
const goodFragment = /* GraphQL */ `
  fragment good on Product {
    id
    name
    price
    description
    created_at
    updated_at
  }
  // ${imageFragment}
  // ${seoFragment}
`;

export default goodFragment;

