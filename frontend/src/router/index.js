import { createRouter, createWebHistory } from "vue-router";
import { useAuthStore } from "../stores/auth";
import Login from "../views/Login.vue";
import Register from "../views/Register.vue";
import Home from "../views/Home.vue";
import CircleView from "../views/CircleView.vue";
import UserSettings from "../views/UserSettings.vue";

const routes = [
	{ path: "/login", component: Login },
	{ path: "/register", component: Register },
	{
		path: "/",
		component: Home,
		meta: { requiresAuth: true },
		children: [
			{ path: "circle/:id", component: CircleView, props: true },
			{ path: "settings", component: UserSettings },
		],
	},
];

const router = createRouter({
	history: createWebHistory(),
	routes,
});

router.beforeEach(async (to, from, next) => {
	const auth = useAuthStore();
	if (!auth.user && to.path !== "/login" && to.path !== "/register") {
		await auth.fetchMe();
	}

	if (to.meta.requiresAuth && !auth.user) {
		next("/login");
	} else {
		next();
	}
});

export default router;
