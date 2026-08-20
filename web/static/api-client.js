(function registerAPIClient(global) {
  async function execute(fetcher, endpoint, body) {
    const response = await fetcher(endpoint, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body
    });
    const rawBody = await response.text();
    let payload = {};
    if (rawBody) {
      try {
        payload = JSON.parse(rawBody);
      } catch {
        payload = { message: rawBody };
      }
    }
    return {
      ok: response.ok,
      status: response.status,
      statusText: response.statusText,
      payload
    };
  }

  global.StayforlongAPI = Object.freeze({ execute });
})(window);
