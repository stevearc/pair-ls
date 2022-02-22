import * as React from "react";

const { useEffect, useRef } = React;

export default function usePrevious<T>(
  value: T,
  update: boolean = true
): T | null {
  const ref = useRef<T | null>(null);
  useEffect(() => {
    if (update) {
      ref.current = value;
    }
  }, [value]);
  return ref.current;
}
