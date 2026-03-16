'use client';

import { useRouter, useSearchParams } from 'next/navigation';
import React, { createContext, useCallback, useContext, useMemo, useOptimistic, Suspense } from 'react';

type GoodState = {
  [key: string]: string;
} & {
  image?: string;
};

type GoodContextType = {
  state: GoodState;
  updateOption: (name: string, value: string) => GoodState;
  updateImage: (index: string) => GoodState;
};

const GoodContext = createContext<GoodContextType | undefined>(undefined);

function GoodProviderContent({ children }: { children: React.ReactNode }) {
  const searchParams = useSearchParams();

  const getInitialState = () => {
    const params: GoodState = {};
    for (const [key, value] of searchParams.entries()) {
      params[key] = value;
    }
    return params;
  };

  const [state, setOptimisticState] = useOptimistic(
    getInitialState(),
    (prevState: GoodState, update: GoodState) => ({
      ...prevState,
      ...update
    })
  );

  const updateOption = useCallback(
    (name: string, value: string) => {
      const newState = { [name]: value };
      setOptimisticState(newState);
      return { ...state, ...newState };
    },
    [state, setOptimisticState]
  );

  const updateImage = useCallback(
    (index: string) => {
      const newState = { image: index };
      setOptimisticState(newState);
      return { ...state, ...newState };
    },
    [state, setOptimisticState]
  );

  const value = useMemo(
    () => ({
      state,
      updateOption,
      updateImage
    }),
    [state, updateOption, updateImage]
  );

  return <GoodContext.Provider value={value}>{children}</GoodContext.Provider>;
}

export function GoodProvider({ children }: { children: React.ReactNode }) {
  return (
    <Suspense fallback={null}>
      <GoodProviderContent>{children}</GoodProviderContent>
    </Suspense>
  );
}

export function useGood() {
  const context = useContext(GoodContext);
  if (context === undefined) {
    throw new Error('useGood must be used within a GoodProvider');
  }
  return context;
}

export function useUpdateURL() {
  const router = useRouter();

  return (state: GoodState) => {
    const newParams = new URLSearchParams(window.location.search);
    Object.entries(state).forEach(([key, value]) => {
      newParams.set(key, value);
    });
    router.push(`?${newParams.toString()}`, { scroll: false });
  };
}
