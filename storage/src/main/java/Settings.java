public record Settings(String brokerUri, String exchangeName, String serviceName, String storageFile,
        String keysDir, int port, int processingDelaySeconds) {

    public static Settings load() {
        return new Settings(
                requireEnv("BROKER_URI"),
                requireEnv("EXCHANGE_NAME"),
                requireEnv("SERVICE_NAME"),
                requireEnv("STORAGE_FILE"),
                requireEnv("KEYS_DIR"),
                Integer.parseInt(requireEnv("PORT")),
                optionalNonNegativeInt("PROCESSING_DELAY_SECONDS", 0));
    }

    private static int optionalNonNegativeInt(String name, int defaultValue) {
        String value = System.getenv(name);
        if (value == null || value.isBlank()) {
            return defaultValue;
        }

        int parsed = Integer.parseInt(value.trim());
        if (parsed < 0) {
            throw new IllegalStateException(name + " must be a non-negative integer");
        }
        return parsed;
    }

    private static String requireEnv(String name) {
        String value = System.getenv(name);
        if (value == null || value.isBlank()) {
            throw new IllegalStateException(name + " is not set or is empty");
        }
        return value;
    }
}
