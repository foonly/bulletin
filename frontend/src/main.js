import { createApp } from "vue";
import { createPinia } from "pinia";
import App from "./App.vue";
import router from "./router";
import "./styles/index.css";

// Configure axios for Wails desktop mode
const isWails =
	!!window.runtime ||
	window.location.port === "34115" ||
	window.location.port === "5173";
if (isWails) {
	// We don't set baseURL here anymore because the proxy in api.js handles it,
	// but we should still ensure the stores use the right api instance.
}

const app = createApp(App);
app.use(createPinia());
app.use(router);

app.directive("focus", {
	mounted: (el) => el.focus(),
});

app.mount("#app");
