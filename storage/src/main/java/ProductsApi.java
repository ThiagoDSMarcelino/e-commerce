import com.google.gson.Gson;
import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpServer;

import java.io.IOException;
import java.io.OutputStream;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;

public class ProductsApi {

    private static final Gson GSON = new Gson();

    private final Storage storage;
    private final HttpServer server;

    public ProductsApi(Storage storage, int port) throws IOException {
        this.storage = storage;
        this.server = HttpServer.create(new InetSocketAddress(port), 0);

        server.createContext("/products", this::listAvailable);
    }

    public void start() {
        server.start();
        System.out.println(" [*] API de produtos ouvindo na porta " + server.getAddress().getPort());
    }

    public void stop() {
        server.stop(0);
    }

    private void listAvailable(HttpExchange exchange) throws IOException {
        try (exchange) {
            if (!exchange.getRequestURI().getPath().equals("/products")) {
                exchange.sendResponseHeaders(404, -1);
                return;
            }

            if (!exchange.getRequestMethod().equals("GET")) {
                exchange.getResponseHeaders().set("Allow", "GET");
                exchange.sendResponseHeaders(405, -1);
                return;
            }

            byte[] body = GSON.toJson(storage.available()).getBytes(StandardCharsets.UTF_8);

            exchange.getResponseHeaders().set("Content-Type", "application/json; charset=utf-8");
            exchange.sendResponseHeaders(200, body.length);

            try (OutputStream out = exchange.getResponseBody()) {
                out.write(body);
            }
        }
    }
}
