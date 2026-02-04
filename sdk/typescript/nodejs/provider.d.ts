import * as pulumi from "@pulumi/pulumi";
export declare class Provider extends pulumi.ProviderResource {
    /**
     * Returns true if the given object is an instance of Provider.  This is designed to work even
     * when multiple copies of the Pulumi SDK have been loaded into the same process.
     */
    static isInstance(obj: any): obj is Provider;
    /**
     * PEM-encoded CA certificate for validating Dex's TLS certificate.
     */
    readonly caCert: pulumi.Output<string | undefined>;
    /**
     * PEM-encoded client certificate for mTLS to Dex.
     */
    readonly clientCert: pulumi.Output<string | undefined>;
    /**
     * PEM-encoded private key for the client certificate.
     */
    readonly clientKey: pulumi.Output<string | undefined>;
    /**
     * Dex gRPC host:port, e.g. dex.internal.kotaicode:5557.
     */
    readonly host: pulumi.Output<string>;
    /**
     * Create a Provider resource with the given unique name, arguments, and options.
     *
     * @param name The _unique_ name of the resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param opts A bag of options that control this resource's behavior.
     */
    constructor(name: string, args: ProviderArgs, opts?: pulumi.ResourceOptions);
}
/**
 * The set of arguments for constructing a Provider resource.
 */
export interface ProviderArgs {
    /**
     * PEM-encoded CA certificate for validating Dex's TLS certificate.
     */
    caCert?: pulumi.Input<string>;
    /**
     * PEM-encoded client certificate for mTLS to Dex.
     */
    clientCert?: pulumi.Input<string>;
    /**
     * PEM-encoded private key for the client certificate.
     */
    clientKey?: pulumi.Input<string>;
    /**
     * Dex gRPC host:port, e.g. dex.internal.kotaicode:5557.
     */
    host: pulumi.Input<string>;
    /**
     * If true, disables TLS verification (development only).
     */
    insecureSkipVerify?: pulumi.Input<boolean>;
    /**
     * Per-RPC timeout in seconds when talking to Dex.
     */
    timeoutSeconds?: pulumi.Input<number>;
}
