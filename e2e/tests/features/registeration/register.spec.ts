import { test, expect } from "@playwright/test";
import { faker } from "@faker-js/faker";

type UserModel = {
  id: string;
  username: string;
  email: string;
  createdAt: string;
  createdBy: string;
  updatedAt: string | null;
  updatedBy: string | null;
};

type Response<T> = {
  status: number;
  message: string;
  timestamp: string;
  data: T;
};

test.describe("Registration Endpoint", () => {
  test.skip(({ browserName }) => browserName !== "chromium", "Chromium only!");

  test("should successfully register a user with valid registration data", async ({
    request,
  }) => {
    const userData = {
      username: faker.internet.username(),
      email: faker.internet.email(),
      password: faker.internet.password(),
    };

    const response =
      await test.step("Register a new user with valid data", () => {
        // Send POST request to /auth/register
        return request.post("http://localhost:3000/auth/register", {
          data: userData,
        });
      });

    await test.step("Verify registration response", async () => {
      expect(response.status(), "Response status should be 201").toBe(201);

      const json = (await response.json()) as Response<UserModel>;

      expect(json.status, "JSON status should be 201").toBe(201);
      expect(
        json.message,
        "JSON message should indicate successful creation",
      ).toBe("The resource has been successfully created on the server.");
      expect(json.data.id, "User ID should be present").toBeTruthy();
      expect(json.data.username, "Username should match the input").toBe(
        userData.username,
      );
      expect(json.data.email, "Email should match the input").toBe(
        userData.email,
      );
      expect(json.data.createdAt, "createdAt should be present").toBeTruthy();
      expect(json.data.createdBy, "createdBy should be 'system'").toBe(
        "system",
      );
      expect(json.data.updatedAt, "updatedAt should be null").toBeNull();
      expect(json.data.updatedBy, "updatedBy should be null").toBeNull();
    });
  });

  test("should return validation errors with invalid registration data", async ({
    request,
  }) => {
    const userData = {
      username: "",
      email: "",
      password: "",
    };

    const response = await test.step("Send invalid registration data", () =>
      request.post("http://localhost:3000/auth/register", {
        data: userData,
      }),
    );

    await test.step("Verify validation error response", async () => {
      expect(response.status(), "Response status should be 400").toBe(400);

      const json = (await response.json()) as Response<UserModel>;

      expect(json.status, "JSON status should be 400").toBe(400);
      expect(
        json.message,
        "JSON message should indicate invalid parameters",
      ).toBe("The request contains invalid parameters or is malformed.");
      expect(json.timestamp, "Timestamp should not be null").not.toBeNull();
      expect(
        json.data,
        "Validation errors should match expected fields",
      ).toMatchObject({
        "RegisterFormData.Username": "Username is a required field",
        "RegisterFormData.Password": "Password is a required field",
        "RegisterFormData.Email": "Email is a required field",
      });
    });
  });

  test("should return createdBy = tester when registering a user as another user", async ({
    request,
  }) => {
    // Register the initial user "tester"
    const testerData = {
      username: faker.internet.username(),
      email: faker.internet.email(),
      password: faker.internet.password(),
    };

    const registerTesterResponse = await test.step(
      "Register the tester user",
      () =>
        request.post("http://localhost:3000/auth/register", {
          data: testerData,
        }),
    );
    await test.step("Verify tester registration response", async () => {
      expect(
        registerTesterResponse.status(),
        "Register tester response status should be 201",
      ).toBe(201);
    });

    // Login as "tester" to get session cookie
    const loginResponse = await test.step("Login as tester", () =>
      request.post("http://localhost:3000/auth/login", {
        data: {
          username: testerData.username,
          password: testerData.password,
        },
      }),
    );
    await test.step("Verify login response", async () => {
      expect(loginResponse.status(), "Login response status should be 200").toBe(
        200,
      );
    });

    await test.step("Save tester session state", async () => {
      await request.storageState();
    });

    // Register a new user as "tester"
    const newUserData = {
      username: faker.internet.username(),
      email: faker.internet.email(),
      password: faker.internet.password(),
    };

    const registerResponse = await test.step(
      "Register a new user as tester",
      () =>
        request.post("http://localhost:3000/auth/register", {
          data: newUserData,
        }),
    );

    await test.step("Verify new user registration response", async () => {
      expect(
        registerResponse.status(),
        "Register response status should be 201",
      ).toBe(201);

      const json = (await registerResponse.json()) as Response<UserModel>;

      expect(json.status, "Status must be 201").toBe(201);
      expect(json.message, "Message should indicate successful creation").toBe(
        "The resource has been successfully created on the server.",
      );
      expect(
        new Date(json.timestamp).getTime(),
        "Timestamp should not be in the future",
      ).toBeLessThanOrEqual(Date.now());
      expect(json.data.id, "User ID should not be null").not.toBeNull();
      expect(json.data.username, "Username should match new user data").toBe(
        newUserData.username,
      );
      expect(json.data.email, "Email should match new user data").toBe(
        newUserData.email,
      );
      expect(
        new Date(json.data.createdAt).getTime(),
        "createdAt should not be in the future",
      ).toBeLessThanOrEqual(Date.now());
      expect(json.data.createdBy, "createdBy should match tester's user ID").toBe(
        (await registerTesterResponse.json()).data.id,
      );
      expect(json.data.updatedAt, "updatedAt should be null").toBeNull();
      expect(json.data.updatedBy, "updatedBy should be null").toBeNull();
    });
  });
});
