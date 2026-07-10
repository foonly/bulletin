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

export default api;
