import { beforeEach, describe, expect, it, vi } from "vitest";

const STORAGE_KEY = "auth-session-storage";

async function loadStore() {
  vi.resetModules();
  const mod = await import("./useAuthSessionStore");
  return mod.useAuthSessionStore;
}

describe("useAuthSessionStore", () => {
  beforeEach(() => {
    const backingStore = new Map<string, string>();
    Object.defineProperty(window, "localStorage", {
      configurable: true,
      value: {
        clear: () => backingStore.clear(),
        getItem: (key: string) => backingStore.get(key) ?? null,
        removeItem: (key: string) => {
          backingStore.delete(key);
        },
        setItem: (key: string, value: string) => {
          backingStore.set(key, value);
        },
      },
    });
  });

  it("does not hydrate accessToken from legacy persisted storage", async () => {
    window.localStorage.setItem(
      STORAGE_KEY,
      JSON.stringify({
        state: { accessToken: "token-123" },
        version: 0,
      }),
    );

    const store = await loadStore();

    expect(store.getState()._hasHydrated).toBe(true);
    expect(store.getState().accessToken).toBeNull();
  });

  it("ignores corrupt legacy persisted payloads", async () => {
    window.localStorage.setItem(STORAGE_KEY, "{not-json");

    const store = await loadStore();

    expect(store.getState()._hasHydrated).toBe(true);
    expect(store.getState().accessToken).toBeNull();
  });

  it("keeps _hasHydrated true when setAccessToken and clear are called", async () => {
    const store = await loadStore();

    store.setState({ _hasHydrated: false, accessToken: null });
    store.getState().setAccessToken("token-456");

    expect(store.getState()._hasHydrated).toBe(true);
    expect(store.getState().accessToken).toBe("token-456");

    store.getState().clear();

    expect(store.getState()._hasHydrated).toBe(true);
    expect(store.getState().accessToken).toBeNull();
  });
});
