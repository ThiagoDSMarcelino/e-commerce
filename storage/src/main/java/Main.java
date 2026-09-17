import com.google.gson.Gson;
import com.google.gson.JsonSyntaxException;
import com.rabbitmq.client.amqp.Connection;
import com.rabbitmq.client.amqp.Consumer;
import com.rabbitmq.client.amqp.Environment;
import com.rabbitmq.client.amqp.Management;
import com.rabbitmq.client.amqp.Message;
import com.rabbitmq.client.amqp.Publisher;
import com.rabbitmq.client.amqp.impl.AmqpEnvironmentBuilder;

import java.nio.charset.StandardCharsets;
import java.nio.file.Path;
import java.util.concurrent.CountDownLatch;

public class Main {

    private static final String ORDER_CREATED = "pedido.criado";
    private static final String ORDER_DELETED = "pedido.excluido";

    private static final String QUEUE_NAME = "fila.estoque";

    private static final String[] bindingKeys = { ORDER_CREATED, ORDER_DELETED };

    private static final Gson GSON = new Gson();

    public static void main(String[] args) throws Exception {
        Settings settings = Settings.load();
        Storage storage = new Storage(Path.of(settings.storageFile()));
        Signer signer = new Signer(Path.of(settings.keysDir()), settings.serviceName());

        Environment environment = new AmqpEnvironmentBuilder()
                .connectionSettings()
                .uri(settings.brokerUri())
                .environmentBuilder()
                .build();

        Connection connection = environment.connectionBuilder().build();
        Management management = connection.management();

        management.exchange()
                .name(settings.exchangeName())
                .type(Management.ExchangeType.DIRECT)
                .declare();

        String queueName = management.queue()
                .name(QUEUE_NAME)
                .declare()
                .name();

        for (String bindingKey : bindingKeys) {
            management.binding()
                    .sourceExchange(settings.exchangeName())
                    .destinationQueue(queueName)
                    .key(bindingKey)
                    .bind();
        }

        Publisher publisher = connection.publisherBuilder().build();
        OrderHandler handler = new OrderHandler(storage, publisher, signer, settings);

        System.out.println(" [*] Waiting for messages. To exit press CTRL+C");

        Consumer consumer = connection.consumerBuilder()
                .queue(queueName)
                .initialCredits(1)
                .messageHandler((context, message) -> {
                    String body = new String(message.body(), StandardCharsets.UTF_8);

                    Order order;
                    try {
                        order = GSON.fromJson(body, Order.class);
                    } catch (JsonSyntaxException e) {
                        System.out.println(" [!] Pedido invalido, descartando: '" + body + "'");
                        context.discard();
                        return;
                    }

                    if (order == null) {
                        System.out.println(" [!] Pedido vazio, descartando");
                        context.discard();
                        return;
                    }

                    try {
                        switch (routingKey(message)) {
                            case ORDER_CREATED -> handler.handleOrderCreated(order);
                            case ORDER_DELETED -> handler.handleOrderDeleted(order);
                            default -> {
                                System.out.println(" [!] Routing key desconhecida: " + routingKey(message));
                                context.discard();
                                return;
                            }
                        }
                        context.accept();
                    } catch (RuntimeException e) {
                        System.out.println(" [!] Falha ao tratar pedido, devolvendo para a fila: " + e);
                        context.requeue();
                    }
                })
                .build();

        Runtime.getRuntime().addShutdownHook(new Thread(() -> {
            consumer.close();
            publisher.close();
            connection.close();
            environment.close();
        }));

        new CountDownLatch(1).await();
    }

    private static String routingKey(Message message) {
        Object rk = message.annotation("x-routing-key");
        if (rk != null) {
            return rk.toString();
        }

        String subject = message.subject();
        return subject != null ? subject : "";
    }
}
