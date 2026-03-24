import { browser } from "$app/environment";
import { client } from "$lib/api/client.gen";

let isConfigured = false;

export function configureApiClient(): void {
	if (isConfigured) {
		return;
	}

	client.setConfig({
		baseUrl: browser ? "" : "http://localhost:8080"
	});
	isConfigured = true;
}
