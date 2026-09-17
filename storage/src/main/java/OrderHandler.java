import com.google.gson.Gson;
import com.rabbitmq.client.amqp.Message;
import com.rabbitmq.client.amqp.Publisher;

import java.nio.charset.StandardCharsets;
import java.security.GeneralSecurityException;

public class OrderHandler {

    private static final String STOCK_OK = "pedido.estoque_ok";
    private static final String STOCK_NOK = "estoque.indisponivel";

    private static final Gson GSON = new Gson();

    private final Storage storage;
    private final Publisher publisher;
    private final Signer signer;
    private final Settings settings;

    public OrderHandler(Storage storage, Publisher publisher, Signer signer, Settings settings) {
        this.storage = storage;
        this.publisher = publisher;
        this.signer = signer;
        this.settings = settings;
    }

    public void handleOrderCreated(Order order) {
        System.out.println(" [x] Pedido criado: '" + order.id() + "'");

        if (storage.tryDebit(order)) {
            System.out.println("     pedido aprovado, estoque reservado");
            publish(STOCK_OK, order);
        } else {
            System.out.println("     pedido reprovado por falta de estoque");
            publish(STOCK_NOK, order);
        }
    }

    public void handleOrderDeleted(Order order) {
        System.out.println(" [x] Pedido excluido: '" + order.id() + "'");
        storage.credit(order);
    }

    private void publish(String routingKey, Order order) {
        byte[] payload = GSON.toJson(order).getBytes(StandardCharsets.UTF_8);

        String signature;
        try {
            signature = signer.sign(payload);
        } catch (GeneralSecurityException e) {
            System.out.println("     [!] falha ao assinar " + routingKey + ": " + e.getMessage());
            return;
        }

        Message event = publisher.message(payload)
                .property("from", settings.serviceName())
                .property("signature", signature)
                .toAddress()
                .exchange(settings.exchangeName())
                .key(routingKey)
                .message();

        publisher.publish(event, context -> {
            if (context.status() == Publisher.Status.ACCEPTED) {
                System.out.println("     publicado " + routingKey + " para o pedido '" + order.id() + "'");
            } else {
                System.out.println("     [!] falha ao publicar " + routingKey + ": " + context.status());
            }
        });
    }
}
