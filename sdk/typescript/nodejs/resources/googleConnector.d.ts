import * as pulumi from "@pulumi/pulumi";
/**
 * Manages a Google connector in Dex. This connector allows users to authenticate using their Google accounts and supports domain and group-based access control.
 */
export declare class GoogleConnector extends pulumi.CustomResource {
    /**
     * Get an existing GoogleConnector resource's state with the given name, ID, and optional extra
     * properties used to qualify the lookup.
     *
     * @param name The _unique_ name of the resulting resource.
     * @param id The _unique_ provider ID of the resource to lookup.
     * @param opts Optional settings to control the behavior of the CustomResource.
     */
    static get(name: string, id: pulumi.Input<pulumi.ID>, opts?: pulumi.CustomResourceOptions): GoogleConnector;
    /**
     * Returns true if the given object is an instance of GoogleConnector.  This is designed to work even
     * when multiple copies of the Pulumi SDK have been loaded into the same process.
     */
    static isInstance(obj: any): obj is GoogleConnector;
    /**
     * Google OAuth client ID.
     */
    readonly clientId: pulumi.Output<string>;
    /**
     * Google OAuth client secret.
     */
    readonly clientSecret: pulumi.Output<string>;
    /**
     * Unique identifier for the Google connector.
     */
    readonly connectorId: pulumi.Output<string>;
    /**
     * Map of domain names to admin email addresses. Used for group lookups in Google Workspace.
     */
    readonly domainToAdminEmail: pulumi.Output<{
        [key: string]: string;
    } | undefined>;
    /**
     * List of Google Groups. Only users in these groups will be allowed to authenticate.
     */
    readonly groups: pulumi.Output<string[] | undefined>;
    /**
     * List of Google Workspace domains. Only users with email addresses in these domains will be allowed to authenticate.
     */
    readonly hostedDomains: pulumi.Output<string[] | undefined>;
    /**
     * Human-readable name for the connector, displayed to users during login.
     */
    readonly name: pulumi.Output<string>;
    /**
     * OAuth prompt type. Valid values: 'consent' (default) or 'select_account'.
     */
    readonly promptType: pulumi.Output<string | undefined>;
    /**
     * Redirect URI registered in Google OAuth app. Must match Dex's callback URL.
     */
    readonly redirectUri: pulumi.Output<string>;
    /**
     * Path to Google service account JSON file. Required for group-based access control.
     */
    readonly serviceAccountFilePath: pulumi.Output<string | undefined>;
    /**
     * Create a GoogleConnector resource with the given unique name, arguments, and options.
     *
     * @param name The _unique_ name of the resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param opts A bag of options that control this resource's behavior.
     */
    constructor(name: string, args: GoogleConnectorArgs, opts?: pulumi.CustomResourceOptions);
}
/**
 * The set of arguments for constructing a GoogleConnector resource.
 */
export interface GoogleConnectorArgs {
    /**
     * Google OAuth client ID.
     */
    clientId: pulumi.Input<string>;
    /**
     * Google OAuth client secret.
     */
    clientSecret: pulumi.Input<string>;
    /**
     * Unique identifier for the Google connector.
     */
    connectorId: pulumi.Input<string>;
    /**
     * Map of domain names to admin email addresses. Used for group lookups in Google Workspace.
     */
    domainToAdminEmail?: pulumi.Input<{
        [key: string]: pulumi.Input<string>;
    }>;
    /**
     * List of Google Groups. Only users in these groups will be allowed to authenticate.
     */
    groups?: pulumi.Input<pulumi.Input<string>[]>;
    /**
     * List of Google Workspace domains. Only users with email addresses in these domains will be allowed to authenticate.
     */
    hostedDomains?: pulumi.Input<pulumi.Input<string>[]>;
    /**
     * Human-readable name for the connector, displayed to users during login.
     */
    name: pulumi.Input<string>;
    /**
     * OAuth prompt type. Valid values: 'consent' (default) or 'select_account'.
     */
    promptType?: pulumi.Input<string>;
    /**
     * Redirect URI registered in Google OAuth app. Must match Dex's callback URL.
     */
    redirectUri: pulumi.Input<string>;
    /**
     * Path to Google service account JSON file. Required for group-based access control.
     */
    serviceAccountFilePath?: pulumi.Input<string>;
}
