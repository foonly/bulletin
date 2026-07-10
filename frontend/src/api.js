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

// Attach Authorization header if token exists in localStorage
api.interceptors.request.use((config) => {
	const token = localStorage.getItem("session_token");
	if (token) {
		config.headers.Authorization = `Bearer ${token}`;
	}
	return config;
});

// Save token to localStorage when received in login response
api.interceptors.response.use((response) => {
	if (response.data && response.data.token) {
		localStorage.setItem("session_token", response.data.token);
	}
	return response;
});

export default api;
