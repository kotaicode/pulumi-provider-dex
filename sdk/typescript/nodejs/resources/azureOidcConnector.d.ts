import * as pulumi from "@pulumi/pulumi";
/**
 * Manages an Azure AD/Entra ID connector in Dex using the generic OIDC connector (type: oidc). This connector allows users to authenticate using their Azure AD/Entra ID credentials.
 */
export declare class AzureOidcConnector extends pulumi.CustomResource {
    /**
     * Get an existing AzureOidcConnector resource's state with the given name, ID, and optional extra
     * properties used to qualify the lookup.
     *
     * @param name The _unique_ name of the resulting resource.
     * @param id The _unique_ provider ID of the resource to lookup.
     * @param opts Optional settings to control the behavior of the CustomResource.
     */
    static get(name: string, id: pulumi.Input<pulumi.ID>, opts?: pulumi.CustomResourceOptions): AzureOidcConnector;
    /**
     * Returns true if the given object is an instance of AzureOidcConnector.  This is designed to work even
     * when multiple copies of the Pulumi SDK have been loaded into the same process.
     */
    static isInstance(obj: any): obj is AzureOidcConnector;
    /**
     * Azure AD application (client) ID.
     */
    readonly clientId: pulumi.Output<string>;
    /**
     * Azure AD application client secret.
     */
    readonly clientSecret: pulumi.Output<string>;
    /**
     * Unique identifier for the Azure connector.
     */
    readonly connectorId: pulumi.Output<string>;
    /**
     * Additional OIDC configuration fields as key-value pairs for advanced scenarios.
     */
    readonly extraOidc: pulumi.Output<{
        [key: string]: any;
    } | undefined>;
    /**
     * Human-readable name for the connector, displayed to users during login.
     */
    readonly name: pulumi.Output<string>;
    /**
     * Redirect URI registered in Azure AD. Must match Dex's callback URL (typically 'https://dex.example.com/callback').
     */
    readonly redirectUri: pulumi.Output<string>;
    /**
     * OIDC scopes to request from Azure AD. Defaults to ['openid', 'profile', 'email', 'offline_access'] if not specified.
     */
    readonly scopes: pulumi.Output<string[] | undefined>;
    /**
     * Azure AD tenant ID (UUID format). This identifies your Azure AD organization.
     */
    readonly tenantId: pulumi.Output<string>;
    /**
     * Source for the username claim. Valid values: 'preferred_username' (default), 'upn' (User Principal Name), or 'email'.
     */
    readonly userNameSource: pulumi.Output<string | undefined>;
    /**
     * Create a AzureOidcConnector resource with the given unique name, arguments, and options.
     *
     * @param name The _unique_ name of the resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param opts A bag of options that control this resource's behavior.
     */
    constructor(name: string, args: AzureOidcConnectorArgs, opts?: pulumi.CustomResourceOptions);
}
/**
 * The set of arguments for constructing a AzureOidcConnector resource.
 */
export interface AzureOidcConnectorArgs {
    /**
     * Azure AD application (client) ID.
     */
    clientId: pulumi.Input<string>;
    /**
     * Azure AD application client secret.
     */
    clientSecret: pulumi.Input<string>;
    /**
     * Unique identifier for the Azure connector.
     */
    connectorId: pulumi.Input<string>;
    /**
     * Additional OIDC configuration fields as key-value pairs for advanced scenarios.
     */
    extraOidc?: pulumi.Input<{
        [key: string]: any;
    }>;
    /**
     * Human-readable name for the connector, displayed to users during login.
     */
    name: pulumi.Input<string>;
    /**
     * Redirect URI registered in Azure AD. Must match Dex's callback URL (typically 'https://dex.example.com/callback').
     */
    redirectUri: pulumi.Input<string>;
    /**
     * OIDC scopes to request from Azure AD. Defaults to ['openid', 'profile', 'email', 'offline_access'] if not specified.
     */
    scopes?: pulumi.Input<pulumi.Input<string>[]>;
    /**
     * Azure AD tenant ID (UUID format). This identifies your Azure AD organization.
     */
    tenantId: pulumi.Input<string>;
    /**
     * Source for the username claim. Valid values: 'preferred_username' (default), 'upn' (User Principal Name), or 'email'.
     */
    userNameSource?: pulumi.Input<string>;
}
