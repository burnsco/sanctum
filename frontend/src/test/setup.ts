import "@testing-library/jest-dom";

const localStorageBackingStore = new Map<string, string>();

Object.defineProperty(globalThis, "localStorage", {
  configurable: true,
  writable: true,
  value: {
    clear: () => localStorageBackingStore.clear(),
    getItem: (key: string) => localStorageBackingStore.get(key) ?? null,
    key: (index: number) => Array.from(localStorageBackingStore.keys())[index] ?? null,
    removeItem: (key: string) => {
      localStorageBackingStore.delete(key);
    },
    setItem: (key: string, value: string) => {
      localStorageBackingStore.set(key, String(value));
    },
    get length() {
      return localStorageBackingStore.size;
    },
  },
});

// Mock matchMedia
Object.defineProperty(window, "matchMedia", {
  writable: true,
  value: (query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: () => {},
    removeListener: () => {},
    addEventListener: () => {},
    removeEventListener: () => {},
    dispatchEvent: () => {},
  }),
});

// Mock ResizeObserver
globalThis.ResizeObserver = class ResizeObserver {
  constructor(cb: ResizeObserverCallback) {
    this.cb = cb;
  }
  cb: ResizeObserverCallback;
  observe() {}
  unobserve() {}
  disconnect() {}
};

// Mock IntersectionObserver
globalThis.IntersectionObserver = class IntersectionObserver {
  readonly root: Element | Document | null = null;
  readonly rootMargin: string = "";
  readonly thresholds: ReadonlyArray<number> = [];

  observe() {}
  unobserve() {}
  disconnect() {}
  takeRecords() {
    return [];
  }
  // Mocking intersection observer for tests
} as unknown as typeof globalThis.IntersectionObserver;
