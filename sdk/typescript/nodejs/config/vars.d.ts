/**
 * PEM-encoded CA certificate for validating Dex's TLS certificate.
 */
export declare const caCert: string | undefined;
/**
 * PEM-encoded client certificate for mTLS to Dex.
 */
export declare const clientCert: string | undefined;
/**
 * PEM-encoded private key for the client certificate.
 */
export declare const clientKey: string | undefined;
/**
 * Dex gRPC host:port, e.g. dex.internal.kotaicode:5557.
 */
export declare const host: string | undefined;
/**
 * If true, disables TLS verification (development only).
 */
export declare const insecureSkipVerify: boolean | undefined;
/**
 * Per-RPC timeout in seconds when talking to Dex.
 */
export declare const timeoutSeconds: number | undefined;
