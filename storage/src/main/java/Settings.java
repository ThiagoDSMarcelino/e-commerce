public record Settings(String brokerUri, String exchangeName, String serviceName, String storageFile,
        String keysDir, int port) {

    public static Settings load() {
        return new Settings(
                requireEnv("BROKER_URI"),
                requireEnv("EXCHANGE_NAME"),
                requireEnv("SERVICE_NAME"),
                requireEnv("STORAGE_FILE"),
                requireEnv("KEYS_DIR"),
                Integer.parseInt(requireEnv("PORT")));
    }

    private static String requireEnv(String name) {
        String value = System.getenv(name);
        if (value == null || value.isBlank()) {
            throw new IllegalStateException(name + " is not set or is empty");
        }
        return value;
    }
}
