const assert = require("node:assert/strict");
const test = require("node:test");

global.window = {};
require("../static/api-client.js");

test("execute sends the playground request to the selected API endpoint", async () => {
  let captured;
  const fetcher = async (endpoint, options) => {
    captured = { endpoint, options };
    return response(true, 200, "OK", '{"avg_night":10}');
  };

  const result = await window.StayforlongAPI.execute(fetcher, "/data", "[]");

  assert.equal(captured.endpoint, "/data");
  assert.deepEqual(captured.options, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: "[]"
  });
  assert.deepEqual(result.payload, { avg_night: 10 });
  assert.equal(result.ok, true);
});

test("execute preserves a human-friendly API failure for frontend feedback", async () => {
  const fetcher = async () => response(false, 400, "Bad Request", '{"message":"nights must be between 1 and 365"}');

  const result = await window.StayforlongAPI.execute(fetcher, "/revenue", "[]");

  assert.equal(result.ok, false);
  assert.equal(result.status, 400);
  assert.equal(result.payload.message, "nights must be between 1 and 365");
});

test("execute converts a non-JSON server response into readable feedback", async () => {
  const fetcher = async () => response(false, 502, "Bad Gateway", "Upstream unavailable");

  const result = await window.StayforlongAPI.execute(fetcher, "/data", "[]");

  assert.deepEqual(result.payload, { message: "Upstream unavailable" });
});

function response(ok, status, statusText, body) {
  return { ok, status, statusText, text: async () => body };
}
