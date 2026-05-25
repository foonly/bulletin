import { createRouter, createWebHistory } from "vue-router";
import { useAuthStore } from "../stores/auth";
import Login from "../views/Login.vue";
import Register from "../views/Register.vue";
import Home from "../views/Home.vue";
import CircleView from "../views/CircleView.vue";
import UserSettings from "../views/UserSettings.vue";
import RequestReset from "../views/RequestReset.vue";
import ResetPassword from "../views/ResetPassword.vue";
import VerifyEmail from "../views/VerifyEmail.vue";
import CirclePosts from "../views/circle/CirclePosts.vue";
import CircleChat from "../views/circle/CircleChat.vue";
import CircleThread from "../views/circle/CircleThread.vue";
import CircleNewThread from "../views/circle/CircleNewThread.vue";
import CircleSettings from "../views/circle/CircleSettings.vue";
import CircleSearch from "../views/circle/CircleSearch.vue";

const routes = [
	{ path: "/login", component: Login },
	{ path: "/register", component: Register },
	{ path: "/request-reset", component: RequestReset },
	{ path: "/reset-password", component: ResetPassword },
	{ path: "/verify-email", component: VerifyEmail },
	{
		path: "/",
		component: Home,
		meta: { requiresAuth: true },
		children: [
			{
				path: "circle/:id",
				component: CircleView,
				props: true,
				children: [
					{
						path: "",
						redirect: (to) => ({
							name: "circle-posts",
							params: { id: to.params.id },
						}),
					},
					{
						path: "posts",
						name: "circle-posts",
						component: CirclePosts,
						props: true,
					},
					{
						path: "posts/:threadId",
						name: "circle-thread",
						component: CircleThread,
						props: true,
					},
					{
						path: "new-thread",
						name: "circle-new-thread",
						component: CircleNewThread,
						props: true,
					},
					{
						path: "chat",
						name: "circle-chat",
						component: CircleChat,
						props: true,
					},
					{
						path: "search",
						name: "circle-search",
						component: CircleSearch,
						props: true,
					},
					{
						path: "settings",
						name: "circle-settings",
						component: CircleSettings,
						props: true,
					},
				],
			},
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
	const publicPages = [
		"/login",
		"/register",
		"/request-reset",
		"/reset-password",
		"/verify-email",
	];
	const authRequired = to.meta.requiresAuth;

	if (!auth.user && !publicPages.includes(to.path)) {
		await auth.fetchMe();
	}

	if (authRequired && !auth.user) {
		next("/login");
	} else {
		next();
	}
});

export default router;
