import { Configuration, FrontendApi } from '@ory/client';

const ORY_SDK_URL = process.env.NEXT_PUBLIC_ORY_SDK_URL || 'http://localhost:4433';

/**
 * Ory Kratos Frontend API client
 * Used for session management and authentication
 */
const ory = new FrontendApi(
  new Configuration({
    basePath: ORY_SDK_URL,
    baseOptions: {
      withCredentials: true,
    },
  }),
);

export default ory;
