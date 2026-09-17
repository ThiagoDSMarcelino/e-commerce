import com.google.gson.Gson;
import com.google.gson.GsonBuilder;

import java.io.IOException;
import java.io.Reader;
import java.io.Writer;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

public class Storage {

    private static final Gson GSON = new GsonBuilder().setPrettyPrinting().create();

    private final Path file;
    private final Map<String, Product> products = new LinkedHashMap<>();

    public Storage(Path file) throws IOException {
        this.file = file;

        try (Reader reader = Files.newBufferedReader(file)) {
            Catalog catalog = GSON.fromJson(reader, Catalog.class);
            for (Product product : catalog.products()) {
                products.put(product.id(), product);
            }
        }

        System.out.println(" [*] Estoque carregado de " + file + " (" + products.size() + " produtos)");
    }

    public synchronized boolean tryDebit(Order order) {
        if (order.products() == null || order.products().isEmpty()) {
            System.out.println("     [!] pedido sem produtos, reprovado");
            return false;
        }

        for (Product item : order.products()) {
            Product inStock = products.get(item.id());

            if (inStock == null) {
                System.out.println("     [!] produto fora do catalogo: " + item.id());
                return false;
            }

            if (item.amount() <= 0) {
                System.out.println("     [!] quantidade invalida de " + inStock.name() + ": " + item.amount());
                return false;
            }

            if (inStock.amount() < item.amount()) {
                System.out.println("     [!] estoque insuficiente de " + inStock.name()
                        + ": tem " + inStock.amount() + ", pedido " + item.amount());
                return false;
            }
        }

        apply(order, -1);
        return true;
    }

    public synchronized void credit(Order order) {
        apply(order, +1);
    }

    private void apply(Order order, int sign) {
        if (order.products() == null || order.products().isEmpty()) {
            System.out.println("     pedido sem produtos, estoque inalterado");
            return;
        }

        for (Product item : order.products()) {
            Product inStock = products.get(item.id());

            if (inStock == null) {
                System.out.println("     [!] produto fora do catalogo: " + item.id());
                continue;
            }

            int amount = inStock.amount() + sign * item.amount();
            products.put(inStock.id(), new Product(inStock.id(), inStock.name(), amount));

            System.out.println("     " + inStock.name() + ": " + inStock.amount() + " -> " + amount);
        }

        save();
    }

    private void save() {
        try (Writer writer = Files.newBufferedWriter(file)) {
            GSON.toJson(new Catalog(new ArrayList<>(products.values())), writer);
        } catch (IOException e) {
            throw new IllegalStateException("Falha ao gravar " + file, e);
        }
    }

    private record Catalog(List<Product> products) {
    }
}
