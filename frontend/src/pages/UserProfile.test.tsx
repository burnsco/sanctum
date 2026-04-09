import { QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import UserProfile from "@/pages/UserProfile";
import { buildUser, createTestQueryClient } from "@/test/test-utils";

vi.mock("@/hooks/useUsers", () => ({
  useUserProfile: vi.fn(),
}));

vi.mock("@/hooks/usePosts", () => ({
  usePosts: vi.fn(),
}));

import { usePosts } from "@/hooks/usePosts";
import { useUserProfile } from "@/hooks/useUsers";

function renderUserProfile() {
  const queryClient = createTestQueryClient();
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={["/users/1"]}>
        <Routes>
          <Route path="/users/:id" element={<UserProfile />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("UserProfile", () => {
  it("hides a redacted email instead of rendering a blank row", () => {
    vi.mocked(useUserProfile).mockReturnValue({
      data: buildUser({
        id: 1,
        username: "alice",
        email: "",
        bio: "Painter",
      }),
      isLoading: false,
      error: null,
    } as never);
    vi.mocked(usePosts).mockReturnValue({
      data: [],
    } as never);

    renderUserProfile();

    expect(screen.getByRole("heading", { name: "alice" })).toBeInTheDocument();
    expect(screen.queryByText("test@example.com")).not.toBeInTheDocument();
    expect(screen.getByText(/Joined /)).toBeInTheDocument();
    expect(screen.getByText("Painter")).toBeInTheDocument();
  });
});
