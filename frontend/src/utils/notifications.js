/**
 * Shows a browser notification if permissions are granted and the tab is not active.
 * @param {string} title - The notification title.
 * @param {object} options - Notification options (body, icon, etc.)
 */
export function showBrowserNotification(title, options = {}) {
	if (!("Notification" in window)) return;

	if (
		Notification.permission === "granted" &&
		document.visibilityState !== "visible"
	) {
		const notification = new Notification(title, {
			icon: "/favicon.svg",
			...options,
		});

		notification.onclick = () => {
			if (window.runtime && window.go && window.go.main && window.go.main.App) {
				window.go.main.App.ShowWindow();
			} else {
				window.focus();
			}
			notification.close();
		};
	}
}
