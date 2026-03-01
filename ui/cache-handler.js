/**
 * Next.js cache handler backed by Redis (ISR / route handler cache).
 * Uses REDIS_URI or REDIS_URL. Falls back to in-memory Map if Redis is unavailable.
 * @see https://nextjs.org/docs/app/api-reference/config/next-config-js/incrementalCacheHandlerPath
 */

const { createClient } = require('redis');

const KEY_PREFIX = 'nextjs:';
const TAG_PREFIX = 'nextjs:tag:';

function getRedisUrl() {
  return process.env.REDIS_URI || process.env.REDIS_URL || 'redis://localhost:6379';
}

function serialize(entry) {
  const payload = { ...entry };
  if (Buffer.isBuffer(payload.value)) {
    payload.value = { __b64: payload.value.toString('base64') };
  }
  return JSON.stringify(payload);
}

function deserialize(str) {
  if (str === null || str === undefined) return null;
  const entry = JSON.parse(str);
  if (entry === null) return null;
  if (entry.value && entry.value.__b64) {
    entry.value = Buffer.from(entry.value.__b64, 'base64');
  }
  return entry;
}

let clientPromise = null;
let fallbackMap = null;

async function getClient() {
  if (clientPromise === null) {
    const url = getRedisUrl();
    clientPromise = (async () => {
      try {
        const c = createClient({ url });
        c.on('error', (err) => {
          if (process.env.NEXT_PRIVATE_DEBUG_CACHE) {
            console.warn('[cache-handler] Redis error', err);
          }
        });
        await c.connect();
        return c;
      } catch (err) {
        if (process.env.NEXT_PRIVATE_DEBUG_CACHE) {
          console.warn('[cache-handler] Redis connect failed, using in-memory fallback', err);
        }
        return null;
      }
    })();
  }
  return clientPromise;
}

function getFallback() {
  if (fallbackMap === null) fallbackMap = new Map();
  return fallbackMap;
}

module.exports = class CacheHandler {
  constructor(options) {
    this.options = options;
  }

  async get(key) {
    const redis = await getClient();
    const fullKey = KEY_PREFIX + key;
    if (redis?.isReady) {
      try {
        const raw = await redis.get(fullKey);
        if (raw === null) return null;
        return deserialize(raw);
      } catch (err) {
        if (process.env.NEXT_PRIVATE_DEBUG_CACHE) {
          console.warn('[cache-handler] Redis get error', err);
        }
      }
    }
    const entry = getFallback().get(fullKey);
    return entry ?? null;
  }

  async set(key, data, ctx) {
    const fullKey = KEY_PREFIX + key;
    const tags = Array.isArray(ctx?.tags) ? ctx.tags : [];
    const entry = {
      value: data,
      lastModified: Date.now(),
      tags,
    };
    const raw = serialize(entry);

    const redis = await getClient();
    if (redis?.isReady) {
      try {
        await redis.set(fullKey, raw);
        for (const tag of tags) {
          await redis.sAdd(TAG_PREFIX + tag, fullKey);
        }
        return;
      } catch (err) {
        if (process.env.NEXT_PRIVATE_DEBUG_CACHE) {
          console.warn('[cache-handler] Redis set error', err);
        }
      }
    }
    getFallback().set(fullKey, entry);
    for (const tag of tags) {
      const setKey = TAG_PREFIX + tag;
      if (!getFallback().has(setKey)) getFallback().set(setKey, new Set());
      getFallback().get(setKey).add(fullKey);
    }
  }

  async revalidateTag(tags) {
    const tagList = Array.isArray(tags) ? tags : [tags];

    const redis = await getClient();
    if (redis?.isReady) {
      try {
        for (const tag of tagList) {
          const setKey = TAG_PREFIX + tag;
          const keys = await redis.sMembers(setKey);
          if (keys.length) {
            await redis.del(keys);
          }
          await redis.del(setKey);
        }
        return;
      } catch (err) {
        if (process.env.NEXT_PRIVATE_DEBUG_CACHE) {
          console.warn('[cache-handler] Redis revalidateTag error', err);
        }
      }
    }
    const cache = getFallback();
    for (const tag of tagList) {
      const setKey = TAG_PREFIX + tag;
      const keySet = cache.get(setKey);
      if (keySet) {
        for (const k of keySet) cache.delete(k);
        cache.delete(setKey);
      }
    }
  }

  resetRequestCache() {}
};
