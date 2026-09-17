import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.security.GeneralSecurityException;
import java.security.KeyFactory;
import java.security.PrivateKey;
import java.security.PublicKey;
import java.security.Signature;
import java.security.spec.PKCS8EncodedKeySpec;
import java.security.spec.X509EncodedKeySpec;
import java.util.Base64;
import java.util.HashMap;
import java.util.Map;

public class Signer {

    private static final String ALGORITHM = "SHA256withRSA";

    private final Path keysDir;
    private final PrivateKey privateKey;
    private final Map<String, PublicKey> publicKeys = new HashMap<>();

    public Signer(Path keysDir, String serviceName) throws IOException, GeneralSecurityException {
        this.keysDir = keysDir;
        this.privateKey = loadPrivateKey(keysDir.resolve(serviceName + ".pem"));
    }

    public String sign(byte[] payload) throws GeneralSecurityException {
        Signature signature = Signature.getInstance(ALGORITHM);
        signature.initSign(privateKey);
        signature.update(payload);

        return Base64.getEncoder().encodeToString(signature.sign());
    }

    public boolean verify(String from, String signatureBase64, byte[] payload) {
        try {
            Signature signature = Signature.getInstance(ALGORITHM);
            signature.initVerify(publicKey(from));
            signature.update(payload);

            return signature.verify(Base64.getDecoder().decode(signatureBase64));
        } catch (Exception e) {
            System.out.println("     [!] nao foi possivel verificar a assinatura de " + from + ": " + e.getMessage());
            return false;
        }
    }

    private synchronized PublicKey publicKey(String service) throws IOException, GeneralSecurityException {
        PublicKey cached = publicKeys.get(service);
        if (cached != null) {
            return cached;
        }

        PublicKey key = loadPublicKey(keysDir.resolve(service + ".pub"));
        publicKeys.put(service, key);

        return key;
    }

    private static PrivateKey loadPrivateKey(Path path) throws IOException, GeneralSecurityException {
        return KeyFactory.getInstance("RSA")
                .generatePrivate(new PKCS8EncodedKeySpec(decodePem(Files.readString(path))));
    }

    private static PublicKey loadPublicKey(Path path) throws IOException, GeneralSecurityException {
        return KeyFactory.getInstance("RSA")
                .generatePublic(new X509EncodedKeySpec(decodePem(Files.readString(path))));
    }

    private static byte[] decodePem(String pem) {
        String base64 = pem.replaceAll("-----(BEGIN|END) (PRIVATE|PUBLIC) KEY-----", "").replaceAll("\\s", "");

        return Base64.getDecoder().decode(base64);
    }
}
