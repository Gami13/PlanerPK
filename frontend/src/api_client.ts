import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { ENV } from "./env";
import { PKPlanerService } from "./gen/v1/pkplaner_pb.js";

const transport = createConnectTransport({
	// The base URL should point to the server root (no /api suffix).
	// ConnectRPC routes are under /pkplaner.v1.PKPlanerService/*.
	baseUrl: ENV.apiBaseUrl.replace(/\/api\/?$/, ""),
});

/**
 * Type-safe RPC client for the PKPlaner service.
 *
 * Usage:
 * ```ts
 * const res = await rpc.getAuthURL({});
 * console.log(res.url);
 * ```
 */
export const rpc = createClient(PKPlanerService, transport);
