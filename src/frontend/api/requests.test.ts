import { fetchRequests, submitRequest } from "./requests";
import type { MediaItem } from "../types/plugin";

const mockFetch = jest.fn();

describe("requests api", () => {
  beforeEach(() => {
    mockFetch.mockReset();
    global.fetch = mockFetch as unknown as typeof fetch;
  });

  it("submits the full item payload for request creation", async () => {
    const item: MediaItem = {
      id: "jellyseerr:42",
      external_id: "42",
      plugin_id: "jellyseerr",
      type: "movie",
      title: "Dune",
      tmdb_id: "438631",
    };

    mockFetch.mockResolvedValue({
      ok: true,
      status: 201,
      json: async () => ({
        id: "jellyseerr:99",
        plugin_id: "jellyseerr",
        item,
        requester_id: "current-user",
        status: "pending",
        submitted_at: "2026-05-09T00:00:00Z",
        updated_at: "2026-05-09T00:00:00Z",
      }),
    });

    await submitRequest(item);

    expect(mockFetch).toHaveBeenCalledWith(
      "http://localhost:3000/api/v1/requests",
      expect.objectContaining({
        method: "POST",
        credentials: "include",
        body: JSON.stringify({
          item_id: "jellyseerr:42",
          media_id: "42",
          plugin_id: "jellyseerr",
          type: "movie",
          title: "Dune",
          tmdb_id: "438631",
          imdb_id: undefined,
          tvdb_id: undefined,
          isbn: undefined,
          asin: undefined,
        }),
      }),
    );
  });

  it("builds request filters into the list query string", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ items: [], review_urls: {} }),
    });

    await fetchRequests({ requesterId: "alice", status: "pending" });

    expect(mockFetch).toHaveBeenCalledWith(
      "http://localhost:3000/api/v1/requests?requester_id=alice&status=pending",
      expect.objectContaining({
        method: "GET",
        credentials: "include",
      }),
    );
  });
});
