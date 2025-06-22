import { test, expect } from "@playwright/test";

type Response<T> = {
  status: number;
  message: string;
  timestamp: string;
  data: T;
};

test.describe("GET /healthz", () => {
  test.skip(({ browserName }) => browserName !== "chromium", "Chromium only!");

  test("should return expected response structure and values", async ({
    request,
  }) => {
    const response = await request.get("http://localhost:3000/healthz");
    expect(response.status()).toBe(200);

    const body = (await response.json()) as Response<Record<string, string>>;

    // Validate response structure and values
    expect(new Date(body.timestamp).getTime()).toBeLessThanOrEqual(Date.now());
    expect(body).toMatchObject({
      status: 200,
      message: "Your request was successfully completed.",
      data: { database: "available" },
    });
  });
});
