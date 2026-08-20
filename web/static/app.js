const examples = {
  data: [
    { request_id: "550e8400-e29b-41d4-a716-446655440001", check_in: "2026-09-01", nights: 2, selling_rate: 240, margin: 20 },
    { request_id: "550e8400-e29b-41d4-a716-446655440002", check_in: "2026-09-05", nights: 3, selling_rate: 450, margin: 18 },
    { request_id: "550e8400-e29b-41d4-a716-446655440003", check_in: "2026-09-12", nights: 1, selling_rate: 160, margin: 25 }
  ],
  revenue: [
    { request_id: "550e8400-e29b-41d4-a716-446655440001", check_in: "2026-09-01", nights: 4, selling_rate: 480, margin: 20 },
    { request_id: "550e8400-e29b-41d4-a716-446655440002", check_in: "2026-09-05", nights: 3, selling_rate: 390, margin: 18 },
    { request_id: "550e8400-e29b-41d4-a716-446655440003", check_in: "2026-09-03", nights: 6, selling_rate: 780, margin: 15 },
    { request_id: "550e8400-e29b-41d4-a716-446655440004", check_in: "2026-09-09", nights: 2, selling_rate: 280, margin: 25 }
  ]
};

const endpointCopy = {
  data: "Calculate average, minimum, and maximum profit per night for every submitted booking.",
  revenue: "Select a non-conflicting schedule that maximises occupied nights and report its profit."
};

const apiEndpoints = Object.freeze({
  data: "/data",
  revenue: "/revenue"
});

let activeEndpoint = "data";
const requestBody = document.querySelector("#request-body");
const endpointPath = document.querySelector("#endpoint-path");
const endpointDescription = document.querySelector("#endpoint-description");
const curlOutput = document.querySelector("#curl-output");
const responseOutput = document.querySelector("#response-output");
const responseStatus = document.querySelector("#response-status");
const responseTime = document.querySelector("#response-time");
const requestError = document.querySelector("#request-error");
const executeButton = document.querySelector("#execute");
const metricCards = document.querySelector("#metric-cards");
const executionFeedback = document.querySelector("#execution-feedback");
const feedbackIcon = document.querySelector("#feedback-icon");
const feedbackTitle = document.querySelector("#feedback-title");
const feedbackMessage = document.querySelector("#feedback-message");

function prettyExample(endpoint) {
  return JSON.stringify(examples[endpoint], null, 2);
}

function setEndpoint(endpoint) {
  activeEndpoint = endpoint;
  document.querySelectorAll(".endpoint-item").forEach((button) => button.classList.toggle("is-active", button.dataset.endpoint === endpoint));
  endpointPath.textContent = apiEndpoints[endpoint];
  endpointDescription.textContent = endpointCopy[endpoint];
  requestBody.value = prettyExample(endpoint);
  clearResponse();
  updateCurl();
}

function updateCurl() {
  const body = requestBody.value || "[]";
  const escaped = body.replaceAll("'", "'\\''");
  curlOutput.textContent = `curl --request POST '${window.location.origin}${apiEndpoints[activeEndpoint]}' \\
  --header 'Content-Type: application/json' \\
  --data-raw '${escaped}'`;
}

function showFeedback(kind, title, message) {
  const icons = { loading: "…", success: "✓", error: "!" };
  executionFeedback.hidden = false;
  executionFeedback.className = `execution-feedback ${kind}`;
  feedbackIcon.textContent = icons[kind];
  feedbackTitle.textContent = title;
  feedbackMessage.textContent = message;
}

function clearResponse() {
  responseOutput.textContent = "Run the request to see the response.";
  responseStatus.textContent = "Not executed";
  responseStatus.className = "status-pill";
  responseTime.textContent = "";
  metricCards.hidden = true;
  metricCards.innerHTML = "";
  requestError.textContent = "";
  executionFeedback.hidden = true;
}

function validateJSON() {
  try {
    const parsed = JSON.parse(requestBody.value);
    if (!Array.isArray(parsed)) throw new Error("The request body must be a JSON array.");
    requestError.textContent = "";
    return true;
  } catch (error) {
    requestError.textContent = error.message;
    showFeedback("error", "Request not sent", error.message);
    responseStatus.textContent = "Client validation failed";
    responseStatus.className = "status-pill error";
    responseTime.textContent = "";
    metricCards.hidden = true;
    return false;
  }
}

function renderMetrics(payload) {
  const metrics = [
    ["Average / night", payload.avg_night],
    ["Minimum / night", payload.min_night],
    ["Maximum / night", payload.max_night]
  ];
  if (Object.hasOwn(payload, "total_profit")) metrics.unshift(["Total profit", payload.total_profit]);
  metricCards.innerHTML = metrics.map(([label, value]) => `<div class="metric"><span>${label}</span><strong>${Number(value).toFixed(2)} €</strong></div>`).join("");
  metricCards.hidden = false;
}

async function executeRequest() {
  if (!validateJSON()) return;
  executeButton.disabled = true;
  executeButton.querySelector("span").textContent = "Executing…";
  showFeedback("loading", "Request in progress", `Sending POST ${apiEndpoints[activeEndpoint]} to the booking API.`);
  responseStatus.textContent = "Waiting for API";
  responseStatus.className = "status-pill";
  responseTime.textContent = "";
  metricCards.hidden = true;
  const startedAt = performance.now();
  try {
    const response = await window.StayforlongAPI.execute(fetch, apiEndpoints[activeEndpoint], requestBody.value);
    const payload = response.payload;
    responseOutput.textContent = JSON.stringify(payload, null, 2);
    responseStatus.textContent = `${response.status} ${response.statusText}`;
    responseStatus.className = `status-pill ${response.ok ? "success" : "error"}`;
    responseTime.textContent = `${Math.round(performance.now() - startedAt)} ms`;
    if (response.ok) {
      renderMetrics(payload);
      showFeedback("success", "Request completed successfully", `POST ${apiEndpoints[activeEndpoint]} returned ${response.status}. The values below came from this API.`);
    } else {
      metricCards.hidden = true;
      const reason = payload.message || `The API returned ${response.status} ${response.statusText || "without an error message"}.`;
      showFeedback("error", `Request failed with HTTP ${response.status}`, reason);
    }
  } catch (error) {
    responseOutput.textContent = JSON.stringify({ message: `Could not reach the API: ${error.message}` }, null, 2);
    responseStatus.textContent = "Network error";
    responseStatus.className = "status-pill error";
    responseTime.textContent = "";
    metricCards.hidden = true;
    showFeedback("error", "The booking API could not be reached", `POST ${apiEndpoints[activeEndpoint]} was not completed. ${error.message}`);
  } finally {
    executeButton.disabled = false;
    executeButton.querySelector("span").textContent = "Execute request";
  }
}

function showView(view) {
  document.querySelectorAll(".view").forEach((section) => section.classList.toggle("is-active", section.id === view));
  document.querySelectorAll(".tab").forEach((tab) => tab.classList.toggle("is-active", tab.dataset.tab === view));
  history.replaceState(null, "", view === "faq" ? "/faq" : "/");
  window.scrollTo({ top: 0, behavior: "smooth" });
}

document.querySelectorAll(".endpoint-item").forEach((button) => button.addEventListener("click", () => setEndpoint(button.dataset.endpoint)));
document.querySelectorAll(".tab").forEach((button) => button.addEventListener("click", () => showView(button.dataset.tab)));
document.querySelector("#reset-example").addEventListener("click", () => setEndpoint(activeEndpoint));
document.querySelector("#execute").addEventListener("click", executeRequest);
document.querySelectorAll(".copy-button").forEach((button) => button.addEventListener("click", async () => {
  await navigator.clipboard.writeText(document.querySelector(`#${button.dataset.copy}`).textContent);
  const previous = button.textContent;
  button.textContent = "Copied";
  setTimeout(() => { button.textContent = previous; }, 1200);
}));
requestBody.addEventListener("input", updateCurl);

setEndpoint("data");
if (window.location.pathname === "/faq" || window.location.hash === "#faq") showView("faq");
