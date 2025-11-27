'use client';

import { useRouter, useSearchParams } from 'next/navigation';
import React, { createContext, useContext, useMemo, useOptimistic } from 'react';

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

export function GoodProvider({ children }: { children: React.ReactNode }) {
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

  const updateOption = (name: string, value: string) => {
    const newState = { [name]: value };
    setOptimisticState(newState);
    return { ...state, ...newState };
  };

  const updateImage = (index: string) => {
    const newState = { image: index };
    setOptimisticState(newState);
    return { ...state, ...newState };
  };

  const value = useMemo(
    () => ({
      state,
      updateOption,
      updateImage
    }),
    [state]
  );

  return <GoodContext.Provider value={value}>{children}</GoodContext.Provider>;
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

