import axios from "axios";

const isWails =
	!!window.runtime ||
	window.location.port === "34115" ||
	window.location.port === "5173";
const baseURL = isWails ? "https://uplink.fi" : "";

console.log("API Config:", { isWails, baseURL, port: window.location.port });

const api = axios.create({
	baseURL,
	withCredentials: true,
});

// Intercept requests in Wails to use the Go proxy and bypass CORS
if (window.runtime) {
	api.interceptors.request.use(async (config) => {
		// Only proxy requests to the remote API
		if (
			config.url.startsWith("https://uplink.fi") ||
			config.url.startsWith("/api") ||
			config.url.startsWith("api/")
		) {
			const fullUrl = config.url.startsWith("http")
				? config.url
				: config.url.startsWith("/")
					? `https://uplink.fi${config.url}`
					: `https://uplink.fi/${config.url}`;

			console.log("Proxying request:", config.method.toUpperCase(), fullUrl);

			try {
				const headers = {};
				for (const [key, value] of Object.entries(config.headers)) {
					if (typeof value === "string") {
						headers[key] = value;
					}
				}

				const result = await window.go.main.App.Fetch(
					config.method.toUpperCase(),
					fullUrl,
					config.data ? JSON.stringify(config.data) : "",
					headers,
				);

				console.log("Proxy Result:", {
					status: result.status,
					body: result.body,
					url: fullUrl,
				});

				// Return a fake response to axios
				config.adapter = () => {
					let data = {};
					try {
						data = result.body ? JSON.parse(result.body) : {};
					} catch (e) {
						data = result.body; // Fallback to raw string
					}

					const response = {
						data: data,
						status: result.status,
						statusText:
							result.status >= 200 && result.status < 300 ? "OK" : "Error",
						headers: result.headers,
						config,
						request: {},
					};

					if (result.status >= 200 && result.status < 300) {
						return Promise.resolve(response);
					} else {
						// Axios expect an error object for non-2xx statuses
						const error = new Error(
							"Request failed with status code " + result.status,
						);
						error.response = response;
						error.config = config;
						return Promise.reject(error);
					}
				};
			} catch (err) {
				console.error("Wails Proxy Error:", err);
			}
		}
		return config;
	});
}

export default api;
