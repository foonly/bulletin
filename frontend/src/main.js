import { createApp } from "vue";
import { createPinia } from "pinia";
import axios from "axios";
import App from "./App.vue";
import router from "./router";
import "./styles/index.css";

// Configure axios for Wails desktop mode
if (window.runtime) {
	axios.defaults.baseURL = "https://uplink.fi";
	axios.defaults.withCredentials = true;
}

const app = createApp(App);
app.use(createPinia());
app.use(router);

app.directive("focus", {
	mounted: (el) => el.focus(),
});

app.mount("#app");
