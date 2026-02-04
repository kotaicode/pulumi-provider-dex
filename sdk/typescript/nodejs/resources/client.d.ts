import * as pulumi from "@pulumi/pulumi";
/**
 * Manages an OAuth2 client in Dex. OAuth2 clients are applications that can authenticate users through Dex.
 */
export declare class Client extends pulumi.CustomResource {
    /**
     * Get an existing Client resource's state with the given name, ID, and optional extra
     * properties used to qualify the lookup.
     *
     * @param name The _unique_ name of the resulting resource.
     * @param id The _unique_ provider ID of the resource to lookup.
     * @param opts Optional settings to control the behavior of the CustomResource.
     */
    static get(name: string, id: pulumi.Input<pulumi.ID>, opts?: pulumi.CustomResourceOptions): Client;
    /**
     * Returns true if the given object is an instance of Client.  This is designed to work even
     * when multiple copies of the Pulumi SDK have been loaded into the same process.
     */
    static isInstance(obj: any): obj is Client;
    /**
     * Unique identifier for the OAuth2 client. This is used as the client_id in OAuth2 flows.
     */
    readonly clientId: pulumi.Output<string>;
    /**
     * Timestamp when the client was created (RFC3339 format).
     */
    readonly createdAt: pulumi.Output<string | undefined>;
    /**
     * URL to a logo image for the OAuth2 client. Used in consent screens.
     */
    readonly logoUrl: pulumi.Output<string | undefined>;
    /**
     * Human-readable name for the OAuth2 client.
     */
    readonly name: pulumi.Output<string>;
    /**
     * If true, this client is a public client (e.g., mobile app) and does not require a client secret.
     */
    readonly public: pulumi.Output<boolean | undefined>;
    /**
     * List of allowed redirect URIs for OAuth2 authorization flows. Must be valid HTTP/HTTPS URLs.
     */
    readonly redirectUris: pulumi.Output<string[]>;
    /**
     * Client secret for the OAuth2 client. If not provided, a secure random secret will be generated automatically.
     */
    readonly secret: pulumi.Output<string | undefined>;
    /**
     * List of trusted peer client IDs that can exchange tokens with this client.
     */
    readonly trustedPeers: pulumi.Output<string[] | undefined>;
    /**
     * Create a Client resource with the given unique name, arguments, and options.
     *
     * @param name The _unique_ name of the resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param opts A bag of options that control this resource's behavior.
     */
    constructor(name: string, args: ClientArgs, opts?: pulumi.CustomResourceOptions);
}
/**
 * The set of arguments for constructing a Client resource.
 */
export interface ClientArgs {
    /**
     * Unique identifier for the OAuth2 client. This is used as the client_id in OAuth2 flows.
     */
    clientId: pulumi.Input<string>;
    /**
     * URL to a logo image for the OAuth2 client. Used in consent screens.
     */
    logoUrl?: pulumi.Input<string>;
    /**
     * Human-readable name for the OAuth2 client.
     */
    name: pulumi.Input<string>;
    /**
     * If true, this client is a public client (e.g., mobile app) and does not require a client secret.
     */
    public?: pulumi.Input<boolean>;
    /**
     * List of allowed redirect URIs for OAuth2 authorization flows. Must be valid HTTP/HTTPS URLs.
     */
    redirectUris: pulumi.Input<pulumi.Input<string>[]>;
    /**
     * Client secret for the OAuth2 client. If not provided, a secure random secret will be generated automatically.
     */
    secret?: pulumi.Input<string>;
    /**
     * List of trusted peer client IDs that can exchange tokens with this client.
     */
    trustedPeers?: pulumi.Input<pulumi.Input<string>[]>;
}
