const isDebug = import.meta.env.DEV;




let ENV = {
	apiBaseUrl: "${API_BASE_URL}"
}

if (isDebug) {
	ENV.apiBaseUrl = "http://localhost:8000"
}

export { ENV };

