// /api is proxied to the go server by vite in dev and preview (see vite.config.ts).
// set VITE_API_URL to the api's full url for a production build served some other way
export const API_URL: string = import.meta.env.VITE_API_URL ?? "/api";
