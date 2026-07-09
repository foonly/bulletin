import { defineStore } from "pinia";
import { useCircleStore } from "./circles";
import { useAuthStore } from "./auth";
import { showBrowserNotification } from "../utils/notifications";

export const useSocketStore = defineStore("socket", {
	state: () => ({
		socket: null,
		onlineUserIds: new Set(),
		isConnected: false,
		messageSound: new Audio("/new-message.mp3"),
	}),
	actions: {
		connect() {
			if (this.socket) return;

			const auth = useAuthStore();
			if (!auth.user) return;

			const isDesktop = !!window.runtime;
			const protocol = isDesktop
				? "wss:"
				: window.location.protocol === "https:"
					? "wss:"
					: "ws:";
			const host = isDesktop ? "uplink.fi" : window.location.host;
			const wsUrl = `${protocol}//${host}/api/chat/ws`;

			this.socket = new WebSocket(wsUrl);

			this.socket.onopen = () => {
				this.isConnected = true;
				console.log("WebSocket connected");
			};

			this.socket.onmessage = (event) => {
				const msg = JSON.parse(event.data);
				const circleStore = useCircleStore();

				if (msg.type === "presence") {
					this.onlineUserIds = new Set(msg.online_ids);
				} else if (msg.type === "join") {
					this.onlineUserIds.add(msg.user_id);
					this.onlineUserIds = new Set(this.onlineUserIds);
				} else if (msg.type === "leave") {
					this.onlineUserIds.delete(msg.user_id);
					this.onlineUserIds = new Set(this.onlineUserIds);
				} else if (msg.type === "chat") {
					// Play sound for all incoming messages from others
					if (msg.user_id !== auth.user?.id) {
						this.messageSound.play().catch((err) => {
							console.log("Audio playback blocked:", err);
						});
					}

					// Add message to store if it's the active circle
					if (circleStore.activeCircle?.id === msg.circle_id) {
						circleStore.addChatMessage(msg);
						window.dispatchEvent(
							new CustomEvent("chat-message-received", { detail: msg }),
						);
					}

					// Increment unread count if it's not the active chat or not the current user
					const isNotActiveChat =
						window.location.pathname !== `/circles/${msg.circle_id}/chat`;
					const isNotCurrentUser = msg.user_id !== auth.user?.id;

					if (isNotActiveChat && isNotCurrentUser) {
						circleStore.incrementUnreadChat(msg.circle_id);

						if (auth.notificationsEnabled) {
							const circle = circleStore.circles.find(
								(c) => c.id === msg.circle_id,
							);
							showBrowserNotification(
								`New message in ${circle?.name || "a circle"}`,
								{ body: `${msg.username}: ${msg.content}` },
							);
						}
					}
				} else if (msg.type === "unread_update") {
					circleStore.refreshUnreadCounts();
					if (circleStore.activeCircle) {
						circleStore.refreshActiveTags();
					}
				}
			};

			this.socket.onclose = () => {
				this.isConnected = false;
				this.socket = null;
				console.log("WebSocket disconnected");
				// Reconnect after a delay
				setTimeout(() => this.connect(), 3000);
			};

			this.socket.onerror = (err) => {
				console.error("WebSocket error:", err);
				this.socket.close();
			};
		},
		disconnect() {
			if (this.socket) {
				this.socket.close();
				this.socket = null;
			}
		},
		sendChatMessage(circleId, content) {
			if (this.socket && this.isConnected) {
				this.socket.send(
					JSON.stringify({
						circle_id: circleId,
						content: content,
					}),
				);
			}
		},
	},
});
