const isDebug = import.meta.env.DEV;

let ENV = "${API_BASE_URL}";
if (isDebug) {
	ENV = "http://localhost:8080/api";
}

export { ENV };

